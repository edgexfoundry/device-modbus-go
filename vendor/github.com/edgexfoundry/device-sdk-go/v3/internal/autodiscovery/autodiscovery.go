// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020-2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package autodiscovery

import (
	"context"
	"sync"
	"time"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v3/di"

	"github.com/edgexfoundry/device-sdk-go/v3/internal/container"
)

func BootstrapHandler(
	ctx context.Context,
	wg *sync.WaitGroup,
	_ startup.Timer,
	dic *di.Container) bool {
	driver := container.ProtocolDriverFrom(dic.Get)
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	configuration := container.ConfigurationFrom(dic.Get)
	var runDiscovery bool = true

	if !configuration.Device.Discovery.Enabled {
		lc.Info("AutoDiscovery stopped: disabled by configuration")
		runDiscovery = false
	}
	duration, err := time.ParseDuration(configuration.Device.Discovery.Interval)
	if err != nil || duration <= 0 {
		lc.Info("AutoDiscovery stopped: interval error in configuration")
		runDiscovery = false
	}

	if runDiscovery {
		wg.Add(1)
		go func() {
			defer wg.Done()

			lc.Infof("Starting auto-discovery with duration %v", duration)
			DiscoveryWrapper(driver, lc)
			for {
				select {
				case <-ctx.Done():
					return
				case <-time.After(duration):
					DiscoveryWrapper(driver, lc)
				}
			}
		}()
	}

	return true
}
