// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020-2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/http/utils"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"

	"github.com/edgexfoundry/device-sdk-go/v4/internal/application"
	sdkCommon "github.com/edgexfoundry/device-sdk-go/v4/internal/common"

	"github.com/labstack/echo/v4"
)

func (c *RestController) GetCommand(e echo.Context) error {
	deviceName := e.Param(common.Name)
	commandName := e.Param(common.Command)
	r := e.Request()
	w := e.Response()
	ctx := r.Context()
	correlationId := utils.FromContext(ctx, common.CorrelationHeader)

	// parse query parameter
	queryParams, reserved, err := filterQueryParams(r.URL.RawQuery)
	if err != nil {
		return c.sendEdgexError(w, r, err, common.ApiDeviceNameCommandNameRoute)
	}

	regexCmd := true
	if useRegex := reserved.Get(common.RegexCommand); useRegex == common.ValueFalse {
		regexCmd = false
	}

	event, err := application.GetCommand(ctx, deviceName, commandName, queryParams, regexCmd, c.dic)
	if err != nil {
		return c.sendEdgexError(w, r, err, common.ApiDeviceNameCommandNameRoute)
	}

	// push event to CoreData if specified (default false)
	if pushEvent := reserved.Get(common.PushEvent); pushEvent == common.ValueTrue {
		go sdkCommon.SendEvent(event, correlationId, c.dic)
	}

	// return event in http response if specified (default true)
	if returnEvent := reserved.Get(common.ReturnEvent); returnEvent == "" || returnEvent == common.ValueTrue {
		res := responses.NewEventResponse("", "", http.StatusOK, *event)
		return c.sendEventResponse(w, r, res, http.StatusOK)
	}

	w.WriteHeader(http.StatusOK)
	return nil
}

func (c *RestController) SetCommand(e echo.Context) error {
	r := e.Request()
	w := e.Response()
	if r.Body != nil {
		defer func() { _ = r.Body.Close() }()
	}

	ctx := r.Context()
	deviceName := e.Param(common.Name)
	commandName := e.Param(common.Command)

	// parse query parameter
	queryParams, _, err := filterQueryParams(r.URL.RawQuery)
	if err != nil {
		return c.sendEdgexError(w, r, err, common.ApiDeviceNameCommandNameRoute)
	}

	requestParamsMap, err := parseRequestBody(r)
	if err != nil {
		return c.sendEdgexError(w, r, err, common.ApiDeviceNameCommandNameRoute)
	}

	event, err := application.SetCommand(ctx, deviceName, commandName, queryParams, requestParamsMap, c.dic)
	if err != nil {
		return c.sendEdgexError(w, r, err, common.ApiDeviceNameCommandNameRoute)
	}

	if event != nil {
		correlationId := utils.FromContext(ctx, common.CorrelationHeader)
		go sdkCommon.SendEvent(event, correlationId, c.dic)
	}

	res := commonDTO.NewBaseResponse("", "", http.StatusOK)
	return c.sendResponse(w, r, common.ApiDeviceNameCommandNameRoute, res, http.StatusOK)
}

func parseRequestBody(req *http.Request) (map[string]interface{}, errors.EdgeX) {
	defer req.Body.Close()
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, "failed to read request body", err)
	}

	paramMap := make(map[string]interface{})
	if len(body) == 0 {
		return paramMap, nil
	}

	err = json.Unmarshal(body, &paramMap)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindContractInvalid, "failed to parse request body", err)
	}

	return paramMap, nil
}

func filterQueryParams(rawQuery string) (string, url.Values, errors.EdgeX) {
	queryParams, err := url.ParseQuery(rawQuery)
	if err != nil {
		edgexErr := errors.NewCommonEdgeX(errors.KindServerError, "failed to parse query parameter", err)
		return "", nil, edgexErr
	}

	var reserved = make(url.Values)
	// Separate parameters with SDK reserved prefix
	for k := range queryParams {
		if strings.HasPrefix(k, sdkCommon.SDKReservedPrefix) {
			reserved.Set(k, queryParams.Get(k))
			delete(queryParams, k)
		}
	}

	return queryParams.Encode(), reserved, nil
}
