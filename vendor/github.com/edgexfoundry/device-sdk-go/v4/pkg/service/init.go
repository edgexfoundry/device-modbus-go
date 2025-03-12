// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020-2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"context"
	"net/http"
	"sync"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/controller"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/handlers"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"

	"github.com/edgexfoundry/device-sdk-go/v4/internal/cache"
	sdkCommon "github.com/edgexfoundry/device-sdk-go/v4/internal/common"
	"github.com/edgexfoundry/device-sdk-go/v4/internal/container"
	restController "github.com/edgexfoundry/device-sdk-go/v4/internal/controller/http"
	"github.com/edgexfoundry/device-sdk-go/v4/internal/controller/messaging"
	"github.com/edgexfoundry/device-sdk-go/v4/internal/provision"
	"github.com/edgexfoundry/device-sdk-go/v4/pkg/models"

	"github.com/labstack/echo/v4"
)

// Bootstrap contains references to dependencies required by the BootstrapHandler.
type Bootstrap struct {
	deviceService *deviceService
	router        *echo.Echo
}

// NewBootstrap is a factory method that returns an initialized Bootstrap receiver struct.
func NewBootstrap(ds *deviceService, router *echo.Echo) *Bootstrap {
	return &Bootstrap{
		deviceService: ds,
		router:        router,
	}
}

func (b *Bootstrap) BootstrapHandler(ctx context.Context, wg *sync.WaitGroup, startupTimer startup.Timer, dic *di.Container) (success bool) {
	s := b.deviceService
	s.wg = wg
	s.ctx = ctx
	s.lc = bootstrapContainer.LoggingClientFrom(dic.Get)
	s.autoEventManager = container.AutoEventManagerFrom(dic.Get)
	s.commonController = controller.NewCommonController(dic, b.router, s.serviceKey, sdkCommon.ServiceVersion)
	s.commonController.SetSDKVersion(sdkCommon.SDKVersion)
	s.controller = restController.NewRestController(b.router, dic, s.serviceKey)
	s.controller.InitRestRoutes(dic)

	if !b.checkDependencyServiceAvailable(common.CoreMetaDataServiceKey, startupTimer) {
		return false
	}

	edgexErr := cache.InitCache(s.serviceKey, s.baseServiceName, dic)
	if edgexErr != nil {
		s.lc.Errorf("Failed to init cache: %s", edgexErr.Error())
		return false
	}

	devices := cache.Devices().All()
	config := container.ConfigurationFrom(dic.Get)
	reqFailsTracker := container.NewAllowedFailuresTracker()
	for _, d := range devices {
		reqFailsTracker.Set(d.Name, int(config.Device.AllowedFails))
	}
	dic.Update(di.ServiceConstructorMap{
		container.AllowedRequestFailuresTrackerName: func(get di.Get) any {
			return reqFailsTracker
		},
	})

	if s.AsyncReadingsEnabled() {
		s.asyncCh = make(chan *models.AsyncValues, s.config.Device.AsyncBufferSize)
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.processAsyncResults(ctx, dic)
		}()
	}

	if s.DeviceDiscoveryEnabled() {
		s.deviceCh = make(chan []models.DiscoveredDevice, 1)
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.processAsyncFilterAndAdd(ctx)
		}()
	}

	err := s.driver.Initialize(s)
	if err != nil {
		s.lc.Errorf("ProtocolDriver init failed: %s", err.Error())
		return false
	}

	edgexErr = s.selfRegister()
	if edgexErr != nil {
		s.lc.Errorf("Failed to register %s on Metadata: %s", s.serviceKey, edgexErr.Error())
		return false
	}

	edgexErr = provision.LoadProfiles(s.config.Device.ProfilesDir, dic)
	if edgexErr != nil {
		s.lc.Errorf("Failed to load device profiles: %s", edgexErr.Error())
		return false
	}

	edgexErr = provision.LoadDevices(s.config.Device.DevicesDir, dic)
	if edgexErr != nil {
		s.lc.Errorf("Failed to load devices: %s", edgexErr.Error())
		return false
	}

	edgexErr = provision.LoadProvisionWatchers(s.config.Device.ProvisionWatchersDir, dic)
	if edgexErr != nil {
		s.lc.Errorf("Failed to load provision watchers: %s", edgexErr.Error())
		return false
	}

	s.autoEventManager.StartAutoEvents()

	// Very important that this bootstrap handler is called after the NewServiceMetrics handler so
	// MetricsManager dependency has been created.
	sdkCommon.InitializeSentMetrics(s.lc, dic)
	return true
}

func (b *Bootstrap) checkDependencyServiceAvailable(serviceKey string, startupTimer startup.Timer) bool {
	lc := b.deviceService.lc
	registry := bootstrapContainer.RegistryFrom(b.deviceService.dic.Get)
	mode := bootstrapContainer.DevRemoteModeFrom(b.deviceService.dic.Get)
	clients := bootstrapContainer.ConfigurationFrom(b.deviceService.dic.Get).GetBootstrap().Clients
	clientInfo, ok := (*clients)[serviceKey]
	if !ok {
		lc.Errorf("Client configuration for '%s' not found, missing common config? Use -cp or -cc flags for common config.", serviceKey)
		return false
	}
	pingUrl := clientInfo.Url() + common.ApiPingRoute

	var err error
	for startupTimer.HasNotElapsed() {
		if registry == nil || mode.InDevMode || mode.InRemoteMode {
			lc.Debugf("Check service '%s' availability by Ping", serviceKey)
			client := &http.Client{}
			_, err = client.Get(pingUrl)
		} else {
			lc.Debugf("Check service '%s' availability via Registry", serviceKey)
			_, err = registry.IsServiceAvailable(serviceKey)
		}
		if err == nil {
			break
		}
		lc.Warnf("Check service '%s' availability failed: %s. retrying...", serviceKey, err.Error())
		startupTimer.SleepForInterval()
	}

	if err != nil {
		lc.Errorf("Check service '%s' availability time out: %s", serviceKey, err.Error())
		return false
	}

	lc.Infof("Check service '%s' availability succeeded", serviceKey)

	return true
}

func newMessageBusBootstrap(baseServiceName string) *messageBusBootstrap {
	return &messageBusBootstrap{
		baseServiceName: baseServiceName,
	}
}

type messageBusBootstrap struct {
	baseServiceName string
}

func (h *messageBusBootstrap) messageBusBootstrapHandler(ctx context.Context, wg *sync.WaitGroup, startupTimer startup.Timer, dic *di.Container) bool {
	if !handlers.MessagingBootstrapHandler(ctx, wg, startupTimer, dic) {
		return false
	}

	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	err := messaging.SubscribeCommands(ctx, dic)
	if err != nil {
		lc.Errorf("Failed to subscribe internal command request: %v", err)
		return false
	}

	err = messaging.MetadataSystemEventsCallback(ctx, h.baseServiceName, dic)
	if err != nil {
		lc.Errorf("Failed to subscribe Metadata system events: %v", err)
		return false
	}

	err = messaging.SubscribeDeviceValidation(ctx, dic)
	if err != nil {
		lc.Errorf("Failed to subscribe device validation request: %v", err)
		return false
	}

	return true
}
