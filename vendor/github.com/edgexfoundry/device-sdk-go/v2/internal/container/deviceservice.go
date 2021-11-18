// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"

	sdkModels "github.com/edgexfoundry/device-sdk-go/v2/pkg/models"
)

// DeviceServiceName contains the name of device service struct in the DIC.
var DeviceServiceName = di.TypeInstanceToName(models.DeviceService{})

// ProtocolDiscoveryName contains the name of protocol discovery implementation in the DIC.
var ProtocolDiscoveryName = di.TypeInstanceToName((*sdkModels.ProtocolDiscovery)(nil))

// ProtocolDriverName contains the name of protocol driver implementation in the DIC.
var ProtocolDriverName = di.TypeInstanceToName((*sdkModels.ProtocolDriver)(nil))

// ManagerName contains the name of autoevent manager implementation in the DIC
var ManagerName = di.TypeInstanceToName((*sdkModels.AutoEventManager)(nil))

// DeviceServiceFrom helper function queries the DIC and returns device service struct.
func DeviceServiceFrom(get di.Get) *models.DeviceService {
	return get(DeviceServiceName).(*models.DeviceService)
}

// ProtocolDiscoveryFrom helper function queries the DIC and returns protocol discovery implementation.
func ProtocolDiscoveryFrom(get di.Get) sdkModels.ProtocolDiscovery {
	casted, ok := get(ProtocolDiscoveryName).(sdkModels.ProtocolDiscovery)
	if ok {
		return casted
	}
	return nil
}

// ProtocolDriverFrom helper function queries the DIC and returns protocol driver implementation.
func ProtocolDriverFrom(get di.Get) sdkModels.ProtocolDriver {
	return get(ProtocolDriverName).(sdkModels.ProtocolDriver)
}

// ManagerFrom helper function queries the DIC and returns autoevent manager implementation
func ManagerFrom(get di.Get) sdkModels.AutoEventManager {
	return get(ManagerName).(sdkModels.AutoEventManager)
}
