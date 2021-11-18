// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"net/http"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"

	"github.com/edgexfoundry/device-sdk-go/v2/internal/autodiscovery"
	"github.com/edgexfoundry/device-sdk-go/v2/internal/container"
)

func (c *RestController) Discovery(writer http.ResponseWriter, request *http.Request) {
	ds := container.DeviceServiceFrom(c.dic.Get)
	if ds.AdminState == models.Locked {
		err := errors.NewCommonEdgeX(errors.KindServiceLocked, "service locked", nil)
		c.sendEdgexError(writer, request, err, common.ApiDiscoveryRoute)
		return
	}

	configuration := container.ConfigurationFrom(c.dic.Get)
	if !configuration.Device.Discovery.Enabled {
		err := errors.NewCommonEdgeX(errors.KindServiceUnavailable, "device discovery disabled", nil)
		c.sendEdgexError(writer, request, err, common.ApiDiscoveryRoute)
		return
	}

	discovery := container.ProtocolDiscoveryFrom(c.dic.Get)
	if discovery == nil {
		err := errors.NewCommonEdgeX(errors.KindNotImplemented, "protocolDiscovery not implemented", nil)
		c.sendEdgexError(writer, request, err, common.ApiDiscoveryRoute)
		return
	}

	go autodiscovery.DiscoveryWrapper(discovery, c.lc)
	c.sendResponse(writer, request, common.ApiDiscoveryRoute, nil, http.StatusAccepted)
}
