// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020-2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"context"
	"sync"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/controller"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/handlers"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"

	"github.com/edgexfoundry/device-sdk-go/v3/internal/cache"
	"github.com/edgexfoundry/device-sdk-go/v3/internal/common"
	"github.com/edgexfoundry/device-sdk-go/v3/internal/container"
	"github.com/edgexfoundry/device-sdk-go/v3/internal/controller/http"
	"github.com/edgexfoundry/device-sdk-go/v3/internal/controller/messaging"
	"github.com/edgexfoundry/device-sdk-go/v3/internal/provision"
	"github.com/edgexfoundry/device-sdk-go/v3/pkg/models"

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

func (b *Bootstrap) BootstrapHandler(ctx context.Context, wg *sync.WaitGroup, _ startup.Timer, dic *di.Container) (success bool) {
	s := b.deviceService
	s.wg = wg
	s.ctx = ctx
	s.lc = bootstrapContainer.LoggingClientFrom(dic.Get)
	s.autoEventManager = container.AutoEventManagerFrom(dic.Get)
	s.commonController = controller.NewCommonController(dic, b.router, s.serviceKey, common.ServiceVersion)
	s.commonController.SetSDKVersion(common.SDKVersion)
	s.controller = http.NewRestController(b.router, dic, s.serviceKey)
	s.controller.InitRestRoutes()

	if bootstrapContainer.DeviceClientFrom(dic.Get) == nil {
		s.lc.Error("Client configuration for core-metadata not found, missing common config? Use -cp or -cc flags for common config.")
		return false
	}

	edgexErr := cache.InitCache(s.serviceKey, s.baseServiceName, dic)
	if edgexErr != nil {
		s.lc.Errorf("Failed to init cache: %s", edgexErr.Error())
		return false
	}

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
	common.InitializeSentMetrics(s.lc, dic)
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
