// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020-2024 IOTech Ltd
// Copyright (C) 2023 Intel Corp.
//
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"strings"

	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v4/config"
)

// ConfigurationStruct contains the configuration properties for the device service.
type ConfigurationStruct struct {
	// WritableInfo contains configuration settings that can be changed in the Registry .
	Writable WritableInfo
	// Clients is a collection of services used by a DS.
	Clients bootstrapConfig.ClientsCollection
	// Registry contains registry-specific settings.
	Registry bootstrapConfig.RegistryInfo
	// Service contains DeviceService-specific settings.
	Service bootstrapConfig.ServiceInfo
	// Device contains device-specific configuration settings.
	Device DeviceInfo
	// Driver is a string map contains customized configuration for the protocol driver implemented based on Device SDK
	Driver map[string]string
	// MessageBus contains information for connecting to MessageBus which provides alternative way to publish event
	MessageBus bootstrapConfig.MessageBusInfo
	// MaxEventSize is the maximum event size that can be sent to MessageBus or CoreData
	MaxEventSize int64
}

// UpdateFromRaw converts configuration received from the registry to a service-specific configuration struct which is
// then used to overwrite the service's existing configuration struct.
func (c *ConfigurationStruct) UpdateFromRaw(rawConfig interface{}) bool {
	configuration, ok := rawConfig.(*ConfigurationStruct)
	if ok {
		*c = *configuration
	}
	// Device labels are stored as a single string, so we need to convert it to a string slice.
	// The decision to do the conversion in sdk is based on the fact that we can only modify the logic inside the
	// decoding function of the Keeper client. Unless we opt to develop a new decoder for Consul ourselves.
	// See:
	// Keeper client https://github.com/edgexfoundry/go-mod-configuration/blob/2c3512b731558a2be3f8460c3f0fed9361b5dcd4/internal/pkg/keeper/client.go#L163
	// Consul client https://github.com/edgexfoundry/go-mod-configuration/blob/2c3512b731558a2be3f8460c3f0fed9361b5dcd4/internal/pkg/consul/client.go#L206
	if len(c.Device.Labels) == 1 && strings.HasPrefix(c.Device.Labels[0], "[") &&
		strings.HasSuffix(c.Device.Labels[0], "]") {
		c.Device.Labels = strings.Fields(strings.Trim(c.Device.Labels[0], "[]"))
	}
	return ok
}

// EmptyWritablePtr returns a pointer to a service-specific empty WritableInfo struct.  It is used by the bootstrap to
// provide the appropriate structure to registry.Client's WatchForChanges().
func (c *ConfigurationStruct) EmptyWritablePtr() interface{} {
	return &WritableInfo{}
}

// GetWritablePtr returns pointer to the writable section
func (c *ConfigurationStruct) GetWritablePtr() any {
	return &c.Writable
}

// UpdateWritableFromRaw converts configuration received from the registry to a service-specific WritableInfo struct
// which is then used to overwrite the service's existing configuration's WritableInfo struct.
func (c *ConfigurationStruct) UpdateWritableFromRaw(rawWritable interface{}) bool {
	writable, ok := rawWritable.(*WritableInfo)
	if ok {
		c.Writable = *writable
	}
	return ok
}

// GetBootstrap returns the configuration elements required by the bootstrap.  Currently, a copy of the configuration
// data is returned.  This is intended to be temporary -- since ConfigurationStruct drives the configuration.yaml's
// structure -- until we can make backwards-breaking configuration.yaml changes (which would consolidate these fields
// into an bootstrapConfig.BootstrapConfiguration struct contained within ConfigurationStruct).
func (c *ConfigurationStruct) GetBootstrap() bootstrapConfig.BootstrapConfiguration {
	return bootstrapConfig.BootstrapConfiguration{
		Clients:    &c.Clients,
		Service:    &c.Service,
		Registry:   &c.Registry,
		MessageBus: &c.MessageBus,
	}
}

// GetLogLevel returns the current ConfigurationStruct's log level.
func (c *ConfigurationStruct) GetLogLevel() string {
	return c.Writable.LogLevel
}

// GetRegistryInfo gets the config.RegistryInfo field from the ConfigurationStruct.
func (c *ConfigurationStruct) GetRegistryInfo() bootstrapConfig.RegistryInfo {
	return c.Registry
}

// GetInsecureSecrets returns the service's InsecureSecrets.
func (c *ConfigurationStruct) GetInsecureSecrets() bootstrapConfig.InsecureSecrets {
	return c.Writable.InsecureSecrets
}

// GetTelemetryInfo returns the service's Telemetry settings.
func (c *ConfigurationStruct) GetTelemetryInfo() *bootstrapConfig.TelemetryInfo {
	return &c.Writable.Telemetry
}
