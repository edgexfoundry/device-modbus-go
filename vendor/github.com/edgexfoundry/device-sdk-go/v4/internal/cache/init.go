// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020-2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"context"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
)

// InitCache Init basic state for cache
func InitCache(instanceName string, baseServiceName string, dic *di.Container) errors.EdgeX {
	dc := bootstrapContainer.DeviceClientFrom(dic.Get)
	dpc := bootstrapContainer.DeviceProfileClientFrom(dic.Get)
	pwc := bootstrapContainer.ProvisionWatcherClientFrom(dic.Get)

	// init device cache
	deviceRes, err := dc.DevicesByServiceName(context.Background(), instanceName, 0, -1)
	if err != nil {
		return err
	}
	devices := make([]models.Device, len(deviceRes.Devices))
	for i := range deviceRes.Devices {
		devices[i] = dtos.ToDeviceModel(deviceRes.Devices[i])
	}
	newDeviceCache(devices, dic)

	// init profile cache
	profiles := make([]models.DeviceProfile, 0, len(devices))
	for _, d := range devices {
		if len(d.ProfileName) == 0 {
			continue
		}
		res, err := dpc.DeviceProfileByName(context.Background(), d.ProfileName)
		if err != nil {
			return err
		}
		profiles = append(profiles, dtos.ToDeviceProfileModel(res.Profile))
	}
	newProfileCache(profiles)

	// init provision watcher cache
	// baseServiceName is the service name w/o the instance portion added when -i/--instance flag is used.
	// Using baseServiceName here since ProvisionWatchers are used by all instances of the device service and thus have the ServiceName set to the baseServiceName.
	pwRes, err := pwc.ProvisionWatchersByServiceName(context.Background(), baseServiceName, 0, -1)
	if err != nil {
		return err
	}
	pws := make([]models.ProvisionWatcher, len(pwRes.ProvisionWatchers))
	for i := range pwRes.ProvisionWatchers {
		pws[i] = dtos.ToProvisionWatcherModel(pwRes.ProvisionWatchers[i])
	}
	newProvisionWatcherCache(pws)

	return nil
}
