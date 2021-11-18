// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018-2021 IOTech Ltd
// Copyright (c) 2019 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

// This package provides a basic EdgeX Foundry device service implementation
// meant to be embedded in an application, similar in approach to the builtin
// net/http package.
package service

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"

	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/config"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/flags"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/interfaces"
	bootstrapTypes "github.com/edgexfoundry/go-mod-bootstrap/v2/config"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
	"github.com/edgexfoundry/go-mod-registry/v2/registry"
	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/edgexfoundry/device-sdk-go/v2/internal/clients"
	sdkCommon "github.com/edgexfoundry/device-sdk-go/v2/internal/common"
	"github.com/edgexfoundry/device-sdk-go/v2/internal/config"
	"github.com/edgexfoundry/device-sdk-go/v2/internal/container"
	restController "github.com/edgexfoundry/device-sdk-go/v2/internal/controller/http"
	sdkModels "github.com/edgexfoundry/device-sdk-go/v2/pkg/models"
)

var (
	ds *DeviceService
)

// UpdatableConfig interface allows services to have custom configuration populated from configuration stored
// in the Configuration Provider (aka Consul). Services using custom configuration must implement this interface
// on their custom configuration, even if they do not use Configuration Provider. If they do not use the
// Configuration Provider they can have a dummy implementation of this interface.
// This wraps the actual interface from go-mod-bootstrap so device service code doesn't have to have the additional
// direct import of go-mod-bootstrap.
type UpdatableConfig interface {
	interfaces.UpdatableConfig
}

type DeviceService struct {
	ServiceName     string
	LoggingClient   logger.LoggingClient
	RegistryClient  registry.Client
	SecretProvider  interfaces.SecretProvider
	edgexClients    clients.EdgeXClients
	controller      *restController.RestController
	config          *config.ConfigurationStruct
	deviceService   *models.DeviceService
	driver          sdkModels.ProtocolDriver
	discovery       sdkModels.ProtocolDiscovery
	manager         sdkModels.AutoEventManager
	asyncCh         chan *sdkModels.AsyncValues
	deviceCh        chan []sdkModels.DiscoveredDevice
	initialized     bool
	dic             *di.Container
	flags           flags.Common
	configProcessor *bootstrapConfig.Processor
	ctx             context.Context
	wg              *sync.WaitGroup
}

func (s *DeviceService) Initialize(serviceName, serviceVersion string, proto interface{}) {
	if serviceName == "" {
		_, _ = fmt.Fprintf(os.Stderr, "Please specify device service name")
		os.Exit(1)
	}
	s.ServiceName = serviceName

	if serviceVersion == "" {
		_, _ = fmt.Fprintf(os.Stderr, "Please specify device service version")
		os.Exit(1)
	}
	sdkCommon.ServiceVersion = serviceVersion

	if driver, ok := proto.(sdkModels.ProtocolDriver); ok {
		s.driver = driver
	} else {
		_, _ = fmt.Fprintf(os.Stderr, "Please implement and specify the protocoldriver")
		os.Exit(1)
	}

	if discovery, ok := proto.(sdkModels.ProtocolDiscovery); ok {
		s.discovery = discovery
	} else {
		s.discovery = nil
	}

	s.deviceService = &models.DeviceService{}
	s.config = &config.ConfigurationStruct{}
}

func (s *DeviceService) UpdateFromContainer(r *mux.Router, dic *di.Container) {
	s.LoggingClient = bootstrapContainer.LoggingClientFrom(dic.Get)
	s.RegistryClient = bootstrapContainer.RegistryFrom(dic.Get)
	s.SecretProvider = bootstrapContainer.SecretProviderFrom(dic.Get)
	s.edgexClients.DeviceClient = bootstrapContainer.MetadataDeviceClientFrom(dic.Get)
	s.edgexClients.DeviceServiceClient = bootstrapContainer.MetadataDeviceServiceClientFrom(dic.Get)
	s.edgexClients.DeviceProfileClient = bootstrapContainer.MetadataDeviceProfileClientFrom(dic.Get)
	s.edgexClients.ProvisionWatcherClient = bootstrapContainer.MetadataProvisionWatcherClientFrom(dic.Get)
	s.edgexClients.EventClient = bootstrapContainer.DataEventClientFrom(dic.Get)
	s.config = container.ConfigurationFrom(dic.Get)
	s.manager = container.ManagerFrom(dic.Get)
	s.controller = restController.NewRestController(r, dic)
}

// Name returns the name of this Device Service
func (s *DeviceService) Name() string {
	return s.ServiceName
}

// Version returns the version number of this Device Service
func (s *DeviceService) Version() string {
	return sdkCommon.ServiceVersion
}

// AsyncReadings returns a bool value to indicate whether the asynchronous reading is enabled.
func (s *DeviceService) AsyncReadings() bool {
	return s.config.Device.EnableAsyncReadings
}

func (s *DeviceService) DeviceDiscovery() bool {
	return s.config.Device.Discovery.Enabled
}

// AddRoute allows leveraging the existing internal web server to add routes specific to Device Service.
func (s *DeviceService) AddRoute(route string, handler func(http.ResponseWriter, *http.Request), methods ...string) error {
	return s.controller.AddRoute(route, handler, methods...)
}

// Stop shuts down the Service
func (s *DeviceService) Stop(force bool) {
	if s.initialized {
		err := s.driver.Stop(force)
		if err != nil {
			s.LoggingClient.Error(err.Error())
		}
	}
}

// LoadCustomConfig uses the Config Processor from go-mod-bootstrap to attempt to load service's
// custom configuration. It uses the same command line flags to process the custom config in the same manner
// as the standard configuration.
func (s *DeviceService) LoadCustomConfig(customConfig UpdatableConfig, sectionName string) error {
	if s.configProcessor == nil {
		s.configProcessor = bootstrapConfig.NewProcessorForCustomConfig(s.flags, s.ctx, s.wg, s.dic)
	}
	return s.configProcessor.LoadCustomConfigSection(customConfig, sectionName)
}

// ListenForCustomConfigChanges uses the Config Processor from go-mod-bootstrap to attempt to listen for
// changes to the specified custom configuration section. LoadCustomConfig must be called previously so that
// the instance of svc.configProcessor has already been set.
func (s *DeviceService) ListenForCustomConfigChanges(
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
func (s *DeviceService) selfRegister() errors.EdgeX {
	localDeviceService := models.DeviceService{
		Name:        s.ServiceName,
		Labels:      s.config.Device.Labels,
		BaseAddress: bootstrapTypes.DefaultHttpProtocol + "://" + s.config.Service.Host + ":" + strconv.FormatInt(int64(s.config.Service.Port), 10),
		AdminState:  models.Unlocked,
	}
	*s.deviceService = localDeviceService
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.NewString())

	s.LoggingClient.Debugf("trying to find device service %s", localDeviceService.Name)
	res, err := s.edgexClients.DeviceServiceClient.DeviceServiceByName(ctx, localDeviceService.Name)
	if err != nil {
		if errors.Kind(err) == errors.KindEntityDoesNotExist {
			s.LoggingClient.Infof("device service %s doesn't exist, creating a new one", localDeviceService.Name)
			req := requests.NewAddDeviceServiceRequest(dtos.FromDeviceServiceModelToDTO(localDeviceService))
			idRes, err := s.edgexClients.DeviceServiceClient.Add(ctx, []requests.AddDeviceServiceRequest{req})
			if err != nil {
				s.LoggingClient.Errorf("failed to add device service %s: %v", localDeviceService.Name, err)
				return err
			}
			s.deviceService.Id = idRes[0].Id
			s.LoggingClient.Debugf("new device service id: %s", localDeviceService.Id)
		} else {
			s.LoggingClient.Errorf("failed to find device service %s", localDeviceService.Name)
			return err
		}
	} else {
		s.LoggingClient.Infof("device service %s exists, updating it", s.ServiceName)
		req := requests.NewUpdateDeviceServiceRequest(dtos.FromDeviceServiceModelToUpdateDTO(localDeviceService))
		req.Service.Id = nil
		_, err = s.edgexClients.DeviceServiceClient.Update(ctx, []requests.UpdateDeviceServiceRequest{req})
		if err != nil {
			s.LoggingClient.Errorf("failed to update device service %s with local config: %v", localDeviceService.Name, err)
			oldDeviceService := dtos.ToDeviceServiceModel(res.Service)
			*s.deviceService = oldDeviceService
		}
	}

	return nil
}

// RunningService returns the Service instance which is running
func RunningService() *DeviceService {
	return ds
}

// DriverConfigs retrieves the driver specific configuration
func DriverConfigs() map[string]string {
	return ds.config.Driver
}
