// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020-2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package autodiscovery

import (
	"context"
	"sync"
	"time"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"

	"github.com/edgexfoundry/device-sdk-go/v4/internal/container"
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
	if err != nil {
		lc.Errorf("AutoDiscovery stopped: interval %s error in configuration: %v", configuration.Device.Discovery.Interval, err)
		runDiscovery = false
	} else if duration <= 0 {
		lc.Info("AutoDiscovery schedule is not started: interval <= 0")
		runDiscovery = false
	}

	if runDiscovery {
		wg.Add(1)
		go func() {
			defer wg.Done()

			lc.Infof("Starting auto-discovery with duration %v", duration)
			DiscoveryWrapper(driver, ctx, dic)
			for {
				select {
				case <-ctx.Done():
					return
				case <-time.After(duration):
					DiscoveryWrapper(driver, ctx, dic)
				}
			}
		}()
	}

	return true
}
