// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"
	"time"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"

	"github.com/edgexfoundry/device-sdk-go/v4/internal/cache"
	sdkCommon "github.com/edgexfoundry/device-sdk-go/v4/internal/common"
	"github.com/edgexfoundry/device-sdk-go/v4/internal/container"
)

func deviceReturn(deviceName string, dic *di.Container) {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	dc := bootstrapContainer.DeviceClientFrom(dic.Get)
	config := container.ConfigurationFrom(dic.Get)

	for {
	LOOP:
		time.Sleep(time.Duration(config.Device.DeviceDownTimeout) * time.Second)
		lc.Infof("Checking operational state for device: %s", deviceName)

		d, found := cache.Devices().ForName(deviceName)
		if !found {
			lc.Warnf("Device %s not found. Exiting retry loop.", deviceName)
			return
		}

		if d.OperatingState == models.Up {
			lc.Infof("Device %s is already operational. Exiting retry loop.", deviceName)
			return
		}

		p, found := cache.Profiles().ForName(d.ProfileName)
		if !found {
			lc.Warnf("Device %s has no profile. Cannot set operational state automatically.", deviceName)
			return
		}

		for _, dr := range p.DeviceResources {
			if dr.Properties.ReadWrite == common.ReadWrite_R ||
				dr.Properties.ReadWrite == common.ReadWrite_RW ||
				dr.Properties.ReadWrite == common.ReadWrite_WR {
				_, err := GetCommand(context.Background(), deviceName, dr.Name, "", true, dic)
				if err == nil {
					lc.Infof("Device %s responsive: setting operational state to up.", deviceName)
					sdkCommon.UpdateOperatingState(deviceName, models.Up, lc, dc)
					return
				} else {
					lc.Errorf("Device %s unresponsive: retrying in %v seconds.", deviceName, config.Device.DeviceDownTimeout)
					goto LOOP
				}
			}
		}
		lc.Infof("Device %s has no readable resources. Setting operational state to up without checking.", deviceName)
		sdkCommon.UpdateOperatingState(deviceName, models.Up, lc, dc)
		return
	}
}

func DeviceRequestFailed(deviceName string, dic *di.Container) {
	config := container.ConfigurationFrom(dic.Get)
	if config.Device.AllowedFails > 0 {
		lc := bootstrapContainer.LoggingClientFrom(dic.Get)
		dc := bootstrapContainer.DeviceClientFrom(dic.Get)
		reqFailsTracker := container.AllowedRequestFailuresTrackerFrom(dic.Get)

		if reqFailsTracker.Decrease(deviceName) == 0 {
			d, ok := cache.Devices().ForName(deviceName)
			if !ok {
				return
			}
			if d.OperatingState != models.Down {
				lc.Infof("Marking device %s non-operational", deviceName)
				sdkCommon.UpdateOperatingState(deviceName, models.Down, lc, dc)
			}
			if config.Device.DeviceDownTimeout > 0 {
				lc.Warnf("Will retry device %s in %v seconds", deviceName, config.Device.DeviceDownTimeout)
				go deviceReturn(deviceName, dic)
			}
			return
		}
	}
}

func DeviceRequestSucceeded(d models.Device, dic *di.Container) {
	config := container.ConfigurationFrom(dic.Get)
	reqFailsTracker := container.AllowedRequestFailuresTrackerFrom(dic.Get)
	if config.Device.AllowedFails > 0 && reqFailsTracker.Value(d.Name) < int(config.Device.AllowedFails) {
		reqFailsTracker.Set(d.Name, int(config.Device.AllowedFails))
		if d.OperatingState == models.Down {
			lc := bootstrapContainer.LoggingClientFrom(dic.Get)
			dc := bootstrapContainer.DeviceClientFrom(dic.Get)
			sdkCommon.UpdateOperatingState(d.Name, models.Up, lc, dc)
		}
	}
}
