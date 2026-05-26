// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020-2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"

	"github.com/edgexfoundry/device-sdk-go/v4/pkg/interfaces"
)

// DeviceServiceName contains the name of device service struct in the DIC.
var DeviceServiceName = di.TypeInstanceToName(models.DeviceService{})

// ProtocolDriverName contains the name of protocol driver implementation in the DIC.
var ProtocolDriverName = di.TypeInstanceToName((*interfaces.ProtocolDriver)(nil))

// AutoEventManagerName contains the name of autoevent manager implementation in the DIC
var AutoEventManagerName = di.TypeInstanceToName((*interfaces.AutoEventManager)(nil))

// ExtendedProtocolDriverName contains the name of extended protocol driver implementation in the DIC.
var ExtendedProtocolDriverName = di.TypeInstanceToName((*interfaces.ExtendedProtocolDriver)(nil))

// DeviceServiceFrom helper function queries the DIC and returns device service struct.
func DeviceServiceFrom(get di.Get) *models.DeviceService {
	return get(DeviceServiceName).(*models.DeviceService)
}

// ProtocolDriverFrom helper function queries the DIC and returns protocol driver implementation.
func ProtocolDriverFrom(get di.Get) interfaces.ProtocolDriver {
	return get(ProtocolDriverName).(interfaces.ProtocolDriver)
}

// AutoEventManagerFrom helper function queries the DIC and returns autoevent manager implementation
func AutoEventManagerFrom(get di.Get) interfaces.AutoEventManager {
	return get(AutoEventManagerName).(interfaces.AutoEventManager)
}

// ExtendedProtocolDriverFrom helper function queries the DIC and returns extended protocol driver implementation.
func ExtendedProtocolDriverFrom(get di.Get) interfaces.ExtendedProtocolDriver {
	casted, ok := get(ExtendedProtocolDriverName).(interfaces.ExtendedProtocolDriver)
	if ok {
		return casted
	}
	return nil
}

// DiscoveryRequestIdName contains the name of discovery request id implementation in the DIC.
var DiscoveryRequestIdName = di.TypeInstanceToName(new(string))

// DiscoveryRequestIdFrom helper function queries the DIC and returns discovery request id.
func DiscoveryRequestIdFrom(get di.Get) string {
	id, ok := get(DiscoveryRequestIdName).(string)
	if !ok {
		return ""
	}
	return id
}

// AllowedRequestFailuresTrackerName contains the name of allowed request failures tracker in the DIC.
var AllowedRequestFailuresTrackerName = di.TypeInstanceToName(AllowedFailuresTracker{})

// AllowedRequestFailuresTrackerFrom helper function queries the DIC and returns a device request failures tracker.
func AllowedRequestFailuresTrackerFrom(get di.Get) AllowedFailuresTracker {
	return get(AllowedRequestFailuresTrackerName).(AllowedFailuresTracker)
}
