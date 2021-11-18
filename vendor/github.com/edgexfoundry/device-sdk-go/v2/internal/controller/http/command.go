// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/gorilla/mux"

	"github.com/edgexfoundry/device-sdk-go/v2/internal/application"
	sdkCommon "github.com/edgexfoundry/device-sdk-go/v2/internal/common"
	"github.com/edgexfoundry/device-sdk-go/v2/internal/container"
)

func (c *RestController) Command(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()

	var requestParamsMap map[string]interface{}
	var queryParams string
	var err errors.EdgeX
	var reserved url.Values
	vars := mux.Vars(request)
	correlationID := request.Header.Get(common.CorrelationHeader)
	if correlationID == "" {
		correlationID, _ = request.Context().Value(common.CorrelationHeader).(string)
	}

	// read request body for SET command
	if request.Method == http.MethodPut {
		requestParamsMap, err = parseRequestBody(request, container.ConfigurationFrom(c.dic.Get).Service.MaxRequestSize)
		if err != nil {
			c.sendEdgexError(writer, request, err, common.ApiDeviceNameCommandNameRoute)
			return
		}
	}
	// parse query parameter
	queryParams, reserved, err = filterQueryParams(request.URL.RawQuery)
	if err != nil {
		c.sendEdgexError(writer, request, err, common.ApiDeviceNameCommandNameRoute)
		return
	}

	var sendEvent bool
	// push event to CoreData if specified (default no)
	if ok, exist := reserved[common.PushEvent]; exist && ok[0] == common.ValueYes {
		sendEvent = true
	}
	isRead := request.Method == http.MethodGet
	eventDTO, err := application.CommandHandler(isRead, sendEvent, correlationID, vars, requestParamsMap, queryParams, c.dic)
	if err != nil {
		c.sendEdgexError(writer, request, err, common.ApiDeviceNameCommandNameRoute)
		return
	}

	// return event in http response if specified (default yes)
	if ok, exist := reserved[common.ReturnEvent]; !exist || ok[0] == common.ValueYes {
		if eventDTO != nil {
			res := responses.NewEventResponse("", "", http.StatusOK, *eventDTO)
			c.sendEventResponse(writer, request, res, http.StatusOK)
		} else {
			res := commonDTO.NewBaseResponse("", "", http.StatusOK)
			c.sendResponse(writer, request, common.ApiDeviceNameCommandNameRoute, res, http.StatusOK)
		}
	}
}

func parseRequestBody(req *http.Request, maxRequestSize int64) (map[string]interface{}, errors.EdgeX) {
	defer req.Body.Close()
	body, err := io.ReadAll(req.Body)
	if err != nil {
		if err.Error() == "http: request body too large" {
			errMsg := fmt.Sprintf("request size exceed Service.MaxRequestSize(%d)", maxRequestSize)
			return nil, errors.NewCommonEdgeX(errors.KindLimitExceeded, errMsg, err)
		}
		return nil, errors.NewCommonEdgeX(errors.KindServerError, "failed to read request body", err)
	}

	var paramMap = make(map[string]interface{})
	if len(body) == 0 {
		return paramMap, nil
	}

	err = json.Unmarshal(body, &paramMap)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindContractInvalid, "failed to parse SET command parameters", err)
	}

	return paramMap, nil
}

func filterQueryParams(queryParams string) (string, url.Values, errors.EdgeX) {
	m, err := url.ParseQuery(queryParams)
	if err != nil {
		edgexErr := errors.NewCommonEdgeX(errors.KindServerError, "failed to parse query parameter", err)
		return "", nil, edgexErr
	}

	var reserved = make(url.Values)
	// Separate parameters with SDK reserved prefix
	for k := range m {
		if strings.HasPrefix(k, sdkCommon.SDKReservedPrefix) {
			reserved.Set(k, m.Get(k))
			delete(m, k)
		}
	}

	return m.Encode(), reserved, nil
}
