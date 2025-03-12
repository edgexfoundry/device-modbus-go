// -*- mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018-2025 IOTech Ltd
// Copyright (c) 2021 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"github.com/edgexfoundry/go-mod-bootstrap/v4/config"
)

// WritableInfo is a struct which contains configuration settings that can be changed in the Registry .
type WritableInfo struct {
	// Level is the logging level of writing log message
	LogLevel        string
	InsecureSecrets config.InsecureSecrets
	Reading         Reading
	Telemetry       config.TelemetryInfo
}

// Reading is a struct which contains reading configuration settings.
type Reading struct {
	// ReadingUnits specifies whether or not to indicate the units of measure for the value in the reading
	ReadingUnits bool
}

// DeviceInfo is a struct which contains device specific configuration settings.
type DeviceInfo struct {
	// DataTransform specifies whether or not the DS perform transformations
	// specified by value descriptor on a actuation or query command.
	DataTransform bool
	// MaxCmdOps defines the maximum number of resource operations that
	// can be sent to a Driver in a single command.
	MaxCmdOps int
	// MaxCmdValueLen is the maximum string length of a command parameter or
	// result (including the value descriptor name) that can be returned
	// by a Driver.
	MaxCmdValueLen int
	// ProfilesDir specifies a directory which contains device profiles
	// files which should be imported on startup.
	ProfilesDir string
	// DevicesDir specifies a directory contains devices files which should be imported on startup.
	DevicesDir string
	// ProvisionWatchersDir specifies a directory contains provision watcher files which should be imported on startup.
	ProvisionWatchersDir string
	Discovery            DiscoveryInfo
	// AsyncBufferSize defines the size of asynchronous channel
	AsyncBufferSize int
	// EnableAsyncReadings to determine whether the Device Service would deal with the asynchronous readings
	EnableAsyncReadings bool
	// Labels are properties applied to the device service to help with searching
	Labels []string
	// AllowedFails specifies the number of failed requests allowed before a device is marked as down.
	AllowedFails uint
	// DeviceDownTimeout specifies the duration in seconds that the Device Service will try to contact a device if it is marked as down.
	DeviceDownTimeout uint
}

// DiscoveryInfo is a struct which contains configuration of device auto discovery.
type DiscoveryInfo struct {
	// Enabled controls whether or not device discovery is enabled.
	Enabled bool
	// Interval indicates how often the discovery process will be triggered.
	// It represents as a duration string.
	Interval string
}

// Telemetry provides metrics (on a given device service) to system management.
type Telemetry struct {
	Alloc,
	TotalAlloc,
	Sys,
	Mallocs,
	Frees,
	LiveObjects uint64
}
