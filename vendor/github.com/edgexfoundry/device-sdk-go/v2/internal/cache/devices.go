// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"fmt"
	"sync"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
)

var (
	dc *deviceCache
)

type DeviceCache interface {
	ForName(name string) (models.Device, bool)
	All() []models.Device
	Add(device models.Device) errors.EdgeX
	Update(device models.Device) errors.EdgeX
	RemoveByName(name string) errors.EdgeX
	UpdateAdminState(name string, state models.AdminState) errors.EdgeX
}

type deviceCache struct {
	deviceMap map[string]*models.Device // key is Device name
	mutex     sync.RWMutex
}

func newDeviceCache(devices []models.Device) DeviceCache {
	defaultSize := len(devices)
	dMap := make(map[string]*models.Device, defaultSize)
	for i, d := range devices {
		dMap[d.Name] = &devices[i]
	}

	dc = &deviceCache{deviceMap: dMap}
	return dc
}

// ForName returns a Device with the given device name.
func (d *deviceCache) ForName(name string) (models.Device, bool) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	device, ok := d.deviceMap[name]
	if !ok {
		return models.Device{}, false
	}
	return *device, ok
}

// All returns the current list of devices in the cache.
func (d *deviceCache) All() []models.Device {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	i := 0
	devices := make([]models.Device, len(d.deviceMap))
	for _, device := range d.deviceMap {
		devices[i] = *device
		i++
	}
	return devices
}

// Add adds a new device to the cache. This method is used to populate the
// device cache with pre-existing or recently-added devices from Core Metadata.
func (d *deviceCache) Add(device models.Device) errors.EdgeX {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	return d.add(device)
}

func (d *deviceCache) add(device models.Device) errors.EdgeX {
	if _, ok := d.deviceMap[device.Name]; ok {
		errMsg := fmt.Sprintf("Device %s has already existed in cache", device.Name)
		return errors.NewCommonEdgeX(errors.KindDuplicateName, errMsg, nil)
	}

	d.deviceMap[device.Name] = &device
	return nil
}

// Update updates the device in the cache
func (d *deviceCache) Update(device models.Device) errors.EdgeX {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if err := d.removeByName(device.Name); err != nil {
		return err
	}
	return d.add(device)
}

// RemoveByName removes the specified device by name from the cache.
func (d *deviceCache) RemoveByName(name string) errors.EdgeX {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	return d.removeByName(name)
}

func (d *deviceCache) removeByName(name string) errors.EdgeX {
	_, ok := d.deviceMap[name]
	if !ok {
		errMsg := fmt.Sprintf("failed to find Device %s in cache", name)
		return errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, errMsg, nil)
	}

	delete(d.deviceMap, name)
	return nil
}

// UpdateAdminState updates the device admin state in cache by name. This method
// is used by the UpdateHandler to trigger update device admin state that's been
// updated directly to Core Metadata.
func (d *deviceCache) UpdateAdminState(name string, state models.AdminState) errors.EdgeX {
	if state != models.Locked && state != models.Unlocked {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "invalid AdminState", nil)
	}

	d.mutex.Lock()
	defer d.mutex.Unlock()

	_, ok := d.deviceMap[name]
	if !ok {
		errMsg := fmt.Sprintf("failed to find Device %s in cache", name)
		return errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, errMsg, nil)
	}

	d.deviceMap[name].AdminState = state
	return nil
}

func CheckProfileNotUsed(profileName string) bool {
	for _, device := range dc.deviceMap {
		if device.ProfileName == profileName {
			return false
		}
	}

	return true
}

func Devices() DeviceCache {
	return dc
}
