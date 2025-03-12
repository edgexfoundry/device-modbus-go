// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020-2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"fmt"
	"strings"
	"sync"
	"time"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"

	gometrics "github.com/rcrowley/go-metrics"
)

const (
	deviceNameText      = "{DeviceName}"
	lastConnectedPrefix = "LastConnected-" + deviceNameText
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
	SetLastConnectedByName(name string)
	GetLastConnectedByName(name string) int64
}

type deviceCache struct {
	deviceMap     map[string]*models.Device // key is Device name
	mutex         sync.RWMutex
	dic           *di.Container
	lastConnected map[string]gometrics.Gauge
}

func newDeviceCache(devices []models.Device, dic *di.Container) DeviceCache {
	defaultSize := len(devices)
	dMap := make(map[string]*models.Device, defaultSize)
	dc = &deviceCache{deviceMap: dMap, dic: dic}
	lastConnectedMetrics := make(map[string]gometrics.Gauge)
	for _, d := range devices {
		dMap[d.Name] = &d
		deviceMetric := gometrics.NewGauge()
		registerMetric(d.Name, deviceMetric, dic)
		lastConnectedMetrics[d.Name] = deviceMetric
	}
	dc.lastConnected = lastConnectedMetrics

	return dc
}

func registerMetric(deviceName string, metric interface{}, dic *di.Container) {
	metricsManager := bootstrapContainer.MetricsManagerFrom(dic.Get)
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	registeredName := strings.Replace(lastConnectedPrefix, deviceNameText, deviceName, 1)

	err := metricsManager.Register(registeredName, metric, map[string]string{"device": deviceName})
	if err != nil {
		lc.Warnf("Unable to register %s metric. Metric will not be reported : %s", registeredName, err.Error())
	} else {
		lc.Infof("%s metric has been registered and will be reported (if enabled)", registeredName)
	}
}

func unregisterMetric(deviceName string, dic *di.Container) {
	metricsManager := bootstrapContainer.MetricsManagerFrom(dic.Get)
	registeredName := strings.Replace(lastConnectedPrefix, deviceNameText, deviceName, 1)

	metricsManager.Unregister(registeredName)
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

	// register the lastConnected metric for the new added device
	deviceMetric := gometrics.NewGauge()
	registerMetric(device.Name, deviceMetric, d.dic)
	d.lastConnected[device.Name] = deviceMetric
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

	// unregister the lastConnected metric for the removed device
	unregisterMetric(name, d.dic)
	delete(d.lastConnected, name)

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

// currentTimestamp returns the current timestamp in nanoseconds
var currentTimestamp = func() int64 {
	return time.Now().UnixNano()
}

func (d *deviceCache) SetLastConnectedByName(name string) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	g := d.lastConnected[name]
	g.Update(currentTimestamp())
}

func (d *deviceCache) GetLastConnectedByName(name string) int64 {
	g := d.lastConnected[name]
	return g.Value()
}
