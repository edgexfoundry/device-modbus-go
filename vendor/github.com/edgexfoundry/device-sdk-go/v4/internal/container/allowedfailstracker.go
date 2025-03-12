// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package container

// AllowedFailuresTracker wraps a map of device names to atomic integers that track the number of allowed request
// failures for each device.
type AllowedFailuresTracker struct {
	data map[string]*AtomicInt
}

// NewAllowedFailuresTracker creates and initializes a new tracker.
func NewAllowedFailuresTracker() AllowedFailuresTracker {
	return AllowedFailuresTracker{
		data: make(map[string]*AtomicInt),
	}
}

// Get retrieves the AtomicInt for a given device name.
// Returns nil if the device does not exist.
func (aft *AllowedFailuresTracker) Get(deviceName string) *AtomicInt {
	return aft.data[deviceName]
}

// Set initializes or updates the AtomicInt for a given device.
func (aft *AllowedFailuresTracker) Set(deviceName string, value int) {
	if _, exists := aft.data[deviceName]; !exists {
		aft.data[deviceName] = &AtomicInt{}
	}
	aft.data[deviceName].Set(value)
}

// Decrease decreases the AtomicInt value for a given device by 1.
// Returns the updated value or -1 if the device does not exist.
func (aft *AllowedFailuresTracker) Decrease(deviceName string) int {
	if atomicInt, exists := aft.data[deviceName]; exists {
		if atomicInt.Value() >= 0 {
			return atomicInt.Decrease()
		}
	}
	return -1
}

// Value retrieves the current value of the AtomicInt for a device.
// Returns -1 if the device does not exist.
func (aft *AllowedFailuresTracker) Value(deviceName string) int {
	if atomicInt, exists := aft.data[deviceName]; exists {
		return atomicInt.Value()
	}
	return -1
}

// Remove deletes the entry for a given device name from the tracker.
func (aft *AllowedFailuresTracker) Remove(deviceName string) {
	delete(aft.data, deviceName)
}
