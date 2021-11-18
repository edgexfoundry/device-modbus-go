// -*- mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package clients

import (
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/interfaces"
)

type EdgeXClients struct {
	DeviceClient           interfaces.DeviceClient
	DeviceServiceClient    interfaces.DeviceServiceClient
	DeviceProfileClient    interfaces.DeviceProfileClient
	ProvisionWatcherClient interfaces.ProvisionWatcherClient
	EventClient            interfaces.EventClient
}
