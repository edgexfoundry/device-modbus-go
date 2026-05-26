// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020-2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/edgexfoundry/device-sdk-go/v4/internal/application"
	"github.com/edgexfoundry/device-sdk-go/v4/internal/autodiscovery"
	"github.com/edgexfoundry/device-sdk-go/v4/internal/cache"
	"github.com/edgexfoundry/device-sdk-go/v4/internal/container"
	"github.com/edgexfoundry/device-sdk-go/v4/internal/controller/http/correlation"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
)

func (c *RestController) Discovery(e echo.Context) error {
	request := e.Request()
	writer := e.Response()
	ctx := request.Context()
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

	// Use correlation id as request id since there is no request body
	requestId := correlation.IdFromContext(ctx)
	go func() {
		c.lc.Infof("Discovery triggered. Correlation Id: %s", requestId)
		autodiscovery.DiscoveryWrapper(driver, ctx, c.dic)
		c.lc.Infof("Discovery end. Correlation Id: %s", requestId)
	}()

	response := commonDTO.NewBaseResponse(requestId, "Device Discovery is triggered.", http.StatusAccepted)
	return c.sendResponse(writer, request, common.ApiDiscoveryRoute, response, http.StatusAccepted)
}

func (c *RestController) ProfileScan(e echo.Context) error {
	request := e.Request()
	writer := e.Response()
	ctx := request.Context()
	if request.Body != nil {
		defer func() { _ = request.Body.Close() }()
	}

	body, err := io.ReadAll(request.Body)
	if err != nil {
		edgexErr := errors.NewCommonEdgeX(errors.KindServerError, "Failed to read request body", err)
		return c.sendEdgexError(writer, request, edgexErr, common.ApiProfileScanRoute)
	}

	extdriver := container.ExtendedProtocolDriverFrom(c.dic.Get)
	if extdriver == nil {
		edgexErr := errors.NewCommonEdgeX(errors.KindNotImplemented, "Profile scan is not implemented", nil)
		return c.sendEdgexError(writer, request, edgexErr, common.ApiProfileScanRoute)
	}

	req, edgexErr := profileScanValidation(body, ctx, c.dic)
	if edgexErr != nil {
		return c.sendEdgexError(writer, request, edgexErr, common.ApiProfileScanRoute)
	}

	busy := make(chan bool)
	go func() {
		c.lc.Infof("Profile scanning is triggered. Correlation Id: %s", req.RequestId)
		application.ProfileScanWrapper(busy, extdriver, req, ctx, c.dic)
		c.lc.Infof("Profile scanning is end. Correlation Id: %s", req.RequestId)
	}()
	b := <-busy
	if b {
		edgexErr := errors.NewCommonEdgeX(errors.KindStatusConflict, fmt.Sprintf("Another profile scan process for %s is currently running", req.DeviceName), nil)
		return c.sendEdgexError(writer, request, edgexErr, common.ApiProfileScanRoute)
	}

	response := commonDTO.NewBaseResponse(req.RequestId, "Device ProfileScan is triggered.", http.StatusAccepted)
	return c.sendResponse(writer, request, common.ApiProfileScanRoute, response, http.StatusAccepted)
}

func profileScanValidation(request []byte, ctx context.Context, dic *di.Container) (requests.ProfileScanRequest, errors.EdgeX) {
	var req requests.ProfileScanRequest
	// check device service AdminState
	ds := container.DeviceServiceFrom(dic.Get)
	if ds.AdminState == models.Locked {
		return req, errors.NewCommonEdgeX(errors.KindServiceLocked, "service locked", nil)
	}

	// parse request payload
	err := req.UnmarshalJSON(request)
	if err != nil {
		return req, errors.NewCommonEdgeX(errors.KindContractInvalid, "failed to parse request body", err)
	}

	// check requested device exists
	_, exist := cache.Devices().ForName(req.DeviceName)
	if !exist {
		return req, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("device %s not found", req.DeviceName), nil)
	}

	// check profile should not exist
	if len(req.ProfileName) > 0 {
		if _, exist := cache.Profiles().ForName(req.ProfileName); exist {
			return req, errors.NewCommonEdgeX(errors.KindStatusConflict, fmt.Sprintf("profile name %s is duplicated", req.ProfileName), nil)
		}
	} else {
		req.ProfileName = fmt.Sprintf("%s_profile_%d", req.DeviceName, time.Now().UnixMilli())
	}

	requestId := req.RequestId
	if len(requestId) == 0 {
		// Use correlation id as request id if request id is not provided
		requestId = correlation.IdFromContext(ctx)
	}

	req = requests.ProfileScanRequest{
		BaseRequest: commonDTO.BaseRequest{
			Versionable: commonDTO.NewVersionable(),
			RequestId:   requestId,
		},
		DeviceName:  req.DeviceName,
		ProfileName: req.ProfileName,
		Options:     req.Options,
	}

	return req, nil
}

func (c *RestController) StopDeviceDiscovery(e echo.Context) error {
	request := e.Request()
	writer := e.Response()

	// URL parameters
	requestId := e.Param(common.RequestId)
	queryParams, err := url.ParseQuery(request.URL.RawQuery)
	if err != nil {
		edgexErr := errors.NewCommonEdgeX(errors.KindServerError, "failed to parse query parameter", err)
		return c.sendEdgexErrorWithRequestId(writer, request, edgexErr, common.ApiDiscoveryByIdRoute, requestId)
	}
	options := make(map[string]any)
	for key, vals := range queryParams {
		options[key] = vals
	}
	edgexErr := autodiscovery.StopDeviceDiscovery(c.dic, requestId, options)
	if edgexErr != nil {
		return c.sendEdgexErrorWithRequestId(writer, request, edgexErr, common.ApiDiscoveryByIdRoute, requestId)
	}

	res := commonDTO.NewBaseResponse(requestId, "", http.StatusOK)
	return c.sendResponse(writer, request, common.ApiDiscoveryByIdRoute, res, http.StatusOK)
}

func (c *RestController) StopProfileScan(e echo.Context) error {
	request := e.Request()
	writer := e.Response()

	// URL parameters
	deviceName := e.Param(common.Name)
	queryParams, err := url.ParseQuery(request.URL.RawQuery)
	if err != nil {
		edgexErr := errors.NewCommonEdgeX(errors.KindServerError, "failed to parse query parameter", err)
		return c.sendEdgexError(writer, request, edgexErr, common.ApiProfileScanByDeviceNameRoute)
	}
	options := make(map[string]any)
	for key, vals := range queryParams {
		options[key] = vals
	}
	edgexErr := application.StopProfileScan(c.dic, deviceName, options)
	if edgexErr != nil {
		return c.sendEdgexError(writer, request, edgexErr, common.ApiProfileScanByDeviceNameRoute)
	}

	res := commonDTO.NewBaseResponse("", "", http.StatusOK)
	return c.sendResponse(writer, request, common.ApiProfileScanByDeviceNameRoute, res, http.StatusOK)
}
