// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"context"
	"sync"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/gorilla/mux"

	"github.com/edgexfoundry/device-sdk-go/v2/internal/cache"
	"github.com/edgexfoundry/device-sdk-go/v2/internal/provision"
	"github.com/edgexfoundry/device-sdk-go/v2/pkg/models"
)

// Bootstrap contains references to dependencies required by the BootstrapHandler.
type Bootstrap struct {
	router *mux.Router
}

// NewBootstrap is a factory method that returns an initialized Bootstrap receiver struct.
func NewBootstrap(router *mux.Router) *Bootstrap {
	return &Bootstrap{
		router: router,
	}
}

func (b *Bootstrap) BootstrapHandler(ctx context.Context, wg *sync.WaitGroup, startupTimer startup.Timer, dic *di.Container) (success bool) {
	ds.UpdateFromContainer(b.router, dic)
	ds.ctx = ctx
	ds.wg = wg
	ds.controller.InitRestRoutes()

	err := cache.InitCache(ds.ServiceName, dic)
	if err != nil {
		ds.LoggingClient.Errorf("Failed to init cache: %v", err)
		return false
	}

	if ds.AsyncReadings() {
		ds.asyncCh = make(chan *models.AsyncValues, ds.config.Device.AsyncBufferSize)
		go ds.processAsyncResults(ctx, wg, dic)
	}
	if ds.DeviceDiscovery() {
		ds.deviceCh = make(chan []models.DiscoveredDevice, 1)
		go ds.processAsyncFilterAndAdd(ctx, wg)
	}

	e := ds.driver.Initialize(ds.LoggingClient, ds.asyncCh, ds.deviceCh)
	if e != nil {
		ds.LoggingClient.Errorf("Failed to init ProtocolDriver: %v", e)
		return false
	}
	ds.initialized = true

	err = ds.selfRegister()
	if err != nil {
		ds.LoggingClient.Errorf("Failed to register service on Metadata: %v", err)
		return false
	}

	err = provision.LoadProfiles(ds.config.Device.ProfilesDir, dic)
	if err != nil {
		ds.LoggingClient.Errorf("Failed to create the pre-defined device profiles: %v", err)
		return false
	}

	err = provision.LoadDevices(ds.config.Device.DevicesDir, dic)
	if err != nil {
		ds.LoggingClient.Errorf("Failed to create the pre-defined devices: %v", err)
		return false
	}

	ds.manager.StartAutoEvents()

	return true
}
