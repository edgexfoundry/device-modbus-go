// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 Canonical Ltd
// Copyright (C) 2018-2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

// Package interfaces defines interfaces and structs used to build an EdgeX Foundry Device
// Service.  The interfaces provide an abstraction layer for the device
// or protocol specific logic of a Device Service, and the structs represents request
// and response data format used by the protocol driver.
package interfaces

import (
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"

	sdkModels "github.com/edgexfoundry/device-sdk-go/v4/pkg/models"
)

// ProtocolDriver is a low-level device-specific interface used by
// other components of an EdgeX Device Service to interact with
// a specific class of devices.
type ProtocolDriver interface {
	// Initialize performs protocol-specific initialization for the device service.
	// The given *AsyncValues channel can be used to push asynchronous events and
	// readings to Core Data. The given []DiscoveredDevice channel is used to send
	// discovered devices that will be filtered and added to Core Metadata asynchronously.
	Initialize(sdk DeviceServiceSDK) error

	// HandleReadCommands passes a slice of CommandRequest struct each representing
	// a ResourceOperation for a specific device resource.
	HandleReadCommands(deviceName string, protocols map[string]models.ProtocolProperties, reqs []sdkModels.CommandRequest) ([]*sdkModels.CommandValue, error)

	// HandleWriteCommands passes a slice of CommandRequest struct each representing
	// a ResourceOperation for a specific device resource.
	// Since the commands are actuation commands, params provide parameters for the individual
	// command.
	HandleWriteCommands(deviceName string, protocols map[string]models.ProtocolProperties, reqs []sdkModels.CommandRequest, params []*sdkModels.CommandValue) error

	// Stop instructs the protocol-specific DS code to shutdown gracefully, or
	// if the force parameter is 'true', immediately. The driver is responsible
	// for closing any in-use channels, including the channel used to send async
	// readings (if supported).
	Stop(force bool) error

	// Start runs Device Service startup tasks after the SDK has been completely initialized.
	// This allows Device Service to safely use DeviceServiceSDK interface features in this function call.
	Start() error

	// AddDevice is a callback function that is invoked
	// when a new Device associated with this Device Service is added
	AddDevice(deviceName string, protocols map[string]models.ProtocolProperties, adminState models.AdminState) error

	// UpdateDevice is a callback function that is invoked
	// when a Device associated with this Device Service is updated
	UpdateDevice(deviceName string, protocols map[string]models.ProtocolProperties, adminState models.AdminState) error

	// RemoveDevice is a callback function that is invoked
	// when a Device associated with this Device Service is removed
	RemoveDevice(deviceName string, protocols map[string]models.ProtocolProperties) error

	// Discover triggers protocol specific device discovery, asynchronously
	// writes the results to the channel which is passed to the implementation
	// via ProtocolDriver.Initialize(). The results may be added to the device service
	// based on a set of acceptance criteria (i.e. Provision Watchers).
	Discover() error

	// ValidateDevice triggers device's protocol properties validation, returns error
	// if validation failed and the incoming device will not be added into EdgeX.
	ValidateDevice(device models.Device) error
}
