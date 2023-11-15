// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020-2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"net/http"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"

	"github.com/edgexfoundry/device-sdk-go/v3/internal/autodiscovery"
	"github.com/edgexfoundry/device-sdk-go/v3/internal/container"

	"github.com/labstack/echo/v4"
)

func (c *RestController) Discovery(e echo.Context) error {
	request := e.Request()
	writer := e.Response()
	ds := container.DeviceServiceFrom(c.dic.Get)
	if ds.AdminState == models.Locked {
		err := errors.NewCommonEdgeX(errors.KindServiceLocked, "service locked", nil)
		return c.sendEdgexError(writer, request, err, common.ApiDiscoveryRoute)
	}

	configuration := container.ConfigurationFrom(c.dic.Get)
	if !configuration.Device.Discovery.Enabled {
		err := errors.NewCommonEdgeX(errors.KindServiceUnavailable, "device discovery disabled", nil)
		return c.sendEdgexError(writer, request, err, common.ApiDiscoveryRoute)
	}

	driver := container.ProtocolDriverFrom(c.dic.Get)

	go autodiscovery.DiscoveryWrapper(driver, c.lc)
	return c.sendResponse(writer, request, common.ApiDiscoveryRoute, nil, http.StatusAccepted)
}
