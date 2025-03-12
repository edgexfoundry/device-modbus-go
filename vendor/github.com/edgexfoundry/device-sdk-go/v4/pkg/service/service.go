// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018-2025 IOTech Ltd
// Copyright (C) 2019,2023 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

// Package service provides a basic EdgeX Foundry device service implementation
// meant to be embedded in an application, similar in approach to the builtin
// net/http package.
package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/controller"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/handlers"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/startup"
	"github.com/panjf2000/ants/v2"

	"github.com/edgexfoundry/device-sdk-go/v4/internal/autodiscovery"
	"github.com/edgexfoundry/device-sdk-go/v4/internal/autoevent"
	sdkCommon "github.com/edgexfoundry/device-sdk-go/v4/internal/common"
	"github.com/edgexfoundry/device-sdk-go/v4/internal/config"
	"github.com/edgexfoundry/device-sdk-go/v4/internal/container"
	restController "github.com/edgexfoundry/device-sdk-go/v4/internal/controller/http"
	sdkUtils "github.com/edgexfoundry/device-sdk-go/v4/internal/utils"
	"github.com/edgexfoundry/device-sdk-go/v4/pkg/interfaces"
	sdkModels "github.com/edgexfoundry/device-sdk-go/v4/pkg/models"

	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/config"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/flags"
	bootstrapInterfaces "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/interfaces"
	bootstrapTypes "github.com/edgexfoundry/go-mod-bootstrap/v4/config"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/requests"
	edgexErr "github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const EnvInstanceName = "EDGEX_INSTANCE_NAME"

type deviceService struct {
	serviceKey         string
	baseServiceName    string
	lc                 logger.LoggingClient
	driver             interfaces.ProtocolDriver
	extdriver          interfaces.ExtendedProtocolDriver
	autoEventManager   interfaces.AutoEventManager
	commonController   *controller.CommonController
	controller         *restController.RestController
	asyncCh            chan *sdkModels.AsyncValues
	deviceCh           chan []sdkModels.DiscoveredDevice
	flags              *flags.Default
	deviceServiceModel *models.DeviceService
	config             *config.ConfigurationStruct
	configProcessor    *bootstrapConfig.Processor
	wg                 *sync.WaitGroup
	ctx                context.Context
	dic                *di.Container
	pool               *ants.Pool
}

// NewDeviceService returns an implementation of interfaces.DeviceServiceSDKExt for the specified key, version, and driver.
func NewDeviceService(serviceKey string, serviceVersion string, driver interfaces.ProtocolDriver) (interfaces.DeviceServiceSDK, error) {
	var service deviceService
	if serviceKey == "" {
		return nil, errors.New("please specify device service name")
	}
	service.serviceKey = serviceKey

	if serviceVersion == "" {
		return nil, errors.New("please specify device service version")
	}
	sdkCommon.ServiceVersion = serviceVersion

	service.driver = driver

	if extdriver, ok := driver.(interfaces.ExtendedProtocolDriver); ok {
		service.extdriver = extdriver
	} else {
		service.extdriver = nil
	}

	service.config = &config.ConfigurationStruct{}
	return interfaces.DeviceServiceSDK(&service), nil
}

func (s *deviceService) Run() error {
	var instanceName string
	startupTimer := startup.NewStartUpTimer(s.serviceKey)

	additionalUsage :=
		"    -i, --instance                  Provides a service name suffix which allows unique instance to be created\n" +
			"                                    If the option is provided, service name will be replaced with \"<name>_<instance>\"\n"
	s.flags = flags.NewWithUsage(additionalUsage)
	s.flags.FlagSet.StringVar(&instanceName, "instance", "", "")
	s.flags.FlagSet.StringVar(&instanceName, "i", "", "")
	s.flags.Parse(os.Args[1:])
	s.setServiceName(instanceName)

	s.config = &config.ConfigurationStruct{}
	s.deviceServiceModel = &models.DeviceService{Name: s.serviceKey}

	s.dic = di.NewContainer(di.ServiceConstructorMap{
		container.ConfigurationName: func(get di.Get) any {
			return s.config
		},
		container.DeviceServiceName: func(get di.Get) any {
			return s.deviceServiceModel
		},
		container.ProtocolDriverName: func(get di.Get) any {
			return s.driver
		},
		container.ExtendedProtocolDriverName: func(get di.Get) any {
			return s.extdriver
		},
	})

	// set poolSize to config.Device.AsyncBufferSize
	config := container.ConfigurationFrom(s.dic.Get)
	pool, err := ants.NewPool(config.Device.AsyncBufferSize)
	if err != nil {
		return err
	}
	s.pool = pool

	router := echo.New()
	httpServer := handlers.NewHttpServer(router, true, s.serviceKey)

	ctx, cancel := context.WithCancel(context.Background())
	wg, deferred, successful := bootstrap.RunAndReturnWaitGroup(
		ctx,
		cancel,
		s.flags,
		s.serviceKey,
		common.ConfigStemDevice,
		s.config,
		nil,
		startupTimer,
		s.dic,
		true,
		bootstrapTypes.ServiceTypeDevice,
		[]bootstrapInterfaces.BootstrapHandler{
			httpServer.BootstrapHandler,
			newMessageBusBootstrap(s.baseServiceName).messageBusBootstrapHandler,
			handlers.NewServiceMetrics(s.serviceKey).BootstrapHandler, // Must be after Messaging
			handlers.NewClientsBootstrap().BootstrapHandler,
			autoevent.NewBootstrap(s.pool).BootstrapHandler,
			NewBootstrap(s, router).BootstrapHandler,
			autodiscovery.BootstrapHandler,
			handlers.NewStartMessage(s.serviceKey, sdkCommon.ServiceVersion).BootstrapHandler,
		})

	defer func() {
		deferred()
		s.Stop(false)
		s.pool.Release()
	}()

	if !successful {
		cancel()
		return errors.New("bootstrapping failed")
	}

	err = s.driver.Start()
	if err != nil {
		cancel()
		return fmt.Errorf("failed to Start ProtocolDriver: %v", err)
	}

	wg.Wait()
	return nil
}

// Name returns the name of this Device Service
func (s *deviceService) Name() string {
	return s.serviceKey
}

// Version returns the version number of this Device Service
func (s *deviceService) Version() string {
	return sdkCommon.ServiceVersion
}

// SecretProvider returns the SecretProvider
func (s *deviceService) SecretProvider() bootstrapInterfaces.SecretProvider {
	return bootstrapContainer.SecretProviderFrom(s.dic.Get)
}

// MetricsManager returns the Metrics Manager used to register counter, gauge, gaugeFloat64 or timer metric types from
// github.com/rcrowley/go-metrics
func (s *deviceService) MetricsManager() bootstrapInterfaces.MetricsManager {
	return bootstrapContainer.MetricsManagerFrom(s.dic.Get)
}

// LoggingClient returns the logger.LoggingClient
func (s *deviceService) LoggingClient() logger.LoggingClient {
	if s.lc == nil {
		s.lc = bootstrapContainer.LoggingClientFrom(s.dic.Get)
	}
	return s.lc
}

// AsyncReadingsEnabled returns a bool value to indicate whether the asynchronous reading is enabled.
func (s *deviceService) AsyncReadingsEnabled() bool {
	return s.config.Device.EnableAsyncReadings
}

func (s *deviceService) AsyncValuesChannel() chan *sdkModels.AsyncValues {
	return s.asyncCh
}

// DeviceDiscoveryEnabled returns a bool value to indicate whether the device discovery is enabled.
func (s *deviceService) DeviceDiscoveryEnabled() bool {
	return s.config.Device.Discovery.Enabled
}

func (s *deviceService) DiscoveredDeviceChannel() chan []sdkModels.DiscoveredDevice {
	return s.deviceCh
}

// AddCustomRoute allows leveraging the existing internal web server to add routes specific to Device Service.
func (s *deviceService) AddCustomRoute(route string, authentication interfaces.Authentication, handler func(e echo.Context) error, methods ...string) error {
	if authentication == interfaces.Authenticated {
		authenticationHook := handlers.AutoConfigAuthenticationFunc(s.dic)

		return s.controller.AddRoute(route, handler, methods, authenticationHook)
	}
	return s.controller.AddRoute(route, handler, methods)
}

// LoadCustomConfig uses the Config Processor from go-mod-bootstrap to attempt to load service's
// custom configuration. It uses the same command line flags to process the custom config in the same manner
// as the standard configuration.
func (s *deviceService) LoadCustomConfig(customConfig interfaces.UpdatableConfig, sectionName string) error {
	if s.configProcessor == nil {
		s.configProcessor = bootstrapConfig.NewProcessorForCustomConfig(s.flags, s.ctx, s.wg, s.dic)
	}

	if err := s.configProcessor.LoadCustomConfigSection(customConfig, sectionName); err != nil {
		return err
	}

	s.controller.SetCustomConfigInfo(customConfig)
	s.commonController.SetCustomConfigInfo(customConfig)

	return nil
}

// ListenForCustomConfigChanges uses the Config Processor from go-mod-bootstrap to attempt to listen for
// changes to the specified custom configuration section. LoadCustomConfig must be called previously so that
// the instance of svc.configProcessor has already been set.
func (s *deviceService) ListenForCustomConfigChanges(
	configToWatch interface{},
	sectionName string,
	changedCallback func(interface{})) error {
	if s.configProcessor == nil {
		return fmt.Errorf(
			"custom configuration must be loaded before '%s' section can be watched for changes",
			sectionName)
	}

	s.configProcessor.ListenForCustomConfigChanges(configToWatch, sectionName, changedCallback)
	return nil
}

// selfRegister register device service itself onto metadata.
func (s *deviceService) selfRegister() edgexErr.EdgeX {
	localDeviceService := models.DeviceService{
		Name:        s.serviceKey,
		Labels:      s.config.Device.Labels,
		BaseAddress: bootstrapTypes.DefaultHttpProtocol + "://" + s.config.Service.Host + ":" + strconv.FormatInt(int64(s.config.Service.Port), 10),
		AdminState:  models.Unlocked,
		Properties:  make(map[string]any),
	}
	*s.deviceServiceModel = localDeviceService
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.NewString()) // nolint:staticcheck
	dsc := bootstrapContainer.DeviceServiceClientFrom(s.dic.Get)

	s.lc.Debugf("trying to find device service %s", localDeviceService.Name)
	res, err := dsc.DeviceServiceByName(ctx, localDeviceService.Name)
	if err != nil {
		if edgexErr.Kind(err) == edgexErr.KindEntityDoesNotExist {
			s.lc.Infof("device service %s doesn't exist, creating a new one", localDeviceService.Name)
			req := requests.NewAddDeviceServiceRequest(dtos.FromDeviceServiceModelToDTO(localDeviceService))
			idRes, err := dsc.Add(ctx, []requests.AddDeviceServiceRequest{req})
			if err != nil {
				s.lc.Errorf("failed to add device service %s: %v", localDeviceService.Name, err)
				return err
			}
			s.deviceServiceModel.Id = idRes[0].Id
			s.lc.Debugf("new device service id: %s", localDeviceService.Id)
		} else {
			s.lc.Errorf("failed to find device service %s", localDeviceService.Name)
			return err
		}
	} else {
		s.lc.Infof("device service %s exists, updating it", s.serviceKey)
		req := requests.NewUpdateDeviceServiceRequest(dtos.FromDeviceServiceModelToUpdateDTO(localDeviceService))
		req.Service.Id = nil
		_, err = dsc.Update(ctx, []requests.UpdateDeviceServiceRequest{req})
		if err != nil {
			s.lc.Errorf("failed to update device service %s with local config: %v", localDeviceService.Name, err)
			oldDeviceService := dtos.ToDeviceServiceModel(res.Service)
			*s.deviceServiceModel = oldDeviceService
		}
	}

	return nil
}

// DriverConfigs retrieves the driver specific configuration
func (s *deviceService) DriverConfigs() map[string]string {
	return s.config.Driver
}

// Stop shuts down the Service
func (s *deviceService) Stop(force bool) {
	err := s.driver.Stop(force)
	if err != nil {
		s.lc.Errorf(err.Error())
	}
}

func (s *deviceService) setServiceName(instanceName string) {
	envValue := os.Getenv(EnvInstanceName)
	if len(envValue) > 0 {
		instanceName = envValue
	}

	// Need to capture the base service name to use when loading Provision Watchers so that all instances find the defined provision watchers.
	s.baseServiceName = s.serviceKey

	if len(instanceName) > 0 {
		s.serviceKey = s.serviceKey + "_" + instanceName
	}
}

// PublishDeviceDiscoveryProgressSystemEvent sends a system event to the EdgeX Message Bus
// indicating the progress of the device discovery process.
func (s *deviceService) PublishDeviceDiscoveryProgressSystemEvent(progress, discoveredDeviceCount int, message string) {
	reqId := container.DiscoveryRequestIdFrom(s.dic.Get)
	sdkUtils.PublishDeviceDiscoveryProgressSystemEvent(reqId, progress, discoveredDeviceCount, message, s.ctx, s.dic)
}

// PublishProfileScanProgressSystemEvent sends a system event to the EdgeX Message Bus
// indicating the progress of the profile scan process.
func (s *deviceService) PublishProfileScanProgressSystemEvent(reqId string, progress int, message string) {
	sdkUtils.PublishProfileScanProgressSystemEvent(reqId, progress, message, s.ctx, s.dic)
}

// PublishGenericSystemEvent sends a system event to the EdgeX Message Bus
func (s *deviceService) PublishGenericSystemEvent(eventType, action string, details any) {
	sdkUtils.PublishGenericSystemEvent(eventType, action, details, s.ctx, s.dic)
}
