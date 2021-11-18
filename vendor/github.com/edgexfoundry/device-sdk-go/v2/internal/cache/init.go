// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"context"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
)

// InitCache Init basic state for cache
func InitCache(name string, dic *di.Container) errors.EdgeX {
	dc := bootstrapContainer.MetadataDeviceClientFrom(dic.Get)
	dpc := bootstrapContainer.MetadataDeviceProfileClientFrom(dic.Get)
	pwc := bootstrapContainer.MetadataProvisionWatcherClientFrom(dic.Get)

	// init device cache
	deviceRes, err := dc.DevicesByServiceName(context.Background(), name, 0, -1)
	if err != nil {
		return err
	}
	devices := make([]models.Device, len(deviceRes.Devices))
	for i := range deviceRes.Devices {
		devices[i] = dtos.ToDeviceModel(deviceRes.Devices[i])
	}
	newDeviceCache(devices)

	// init profile cache
	profiles := make([]models.DeviceProfile, len(devices))
	for i, d := range devices {
		res, err := dpc.DeviceProfileByName(context.Background(), d.ProfileName)
		if err != nil {
			return err
		}
		profiles[i] = dtos.ToDeviceProfileModel(res.Profile)
	}
	newProfileCache(profiles)

	// init provision watcher cache
	pwRes, err := pwc.ProvisionWatchersByServiceName(context.Background(), name, 0, -1)
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
