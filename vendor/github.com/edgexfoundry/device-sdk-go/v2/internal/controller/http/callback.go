// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"encoding/json"
	"net/http"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/gorilla/mux"

	"github.com/edgexfoundry/device-sdk-go/v2/internal/application"
)

func (c *RestController) DeleteDevice(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	name := vars[common.Name]

	err := application.DeleteDevice(name, c.dic)
	if err == nil {
		res := commonDTO.NewBaseResponse("", "", http.StatusOK)
		c.sendResponse(writer, request, common.ApiDeviceCallbackNameRoute, res, http.StatusOK)
	} else {
		c.sendEdgexError(writer, request, err, common.ApiDeviceCallbackNameRoute)
	}
}

func (c *RestController) AddDevice(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()

	var addDeviceRequest requests.AddDeviceRequest

	err := json.NewDecoder(request.Body).Decode(&addDeviceRequest)
	if err != nil {
		edgexErr := errors.NewCommonEdgeX(errors.KindServerError, "failed to decode JSON", err)
		c.sendEdgexError(writer, request, edgexErr, common.ApiDeviceCallbackRoute)
		return
	}

	edgexErr := application.AddDevice(addDeviceRequest, c.dic)
	if edgexErr == nil {
		res := commonDTO.NewBaseResponse(addDeviceRequest.RequestId, "", http.StatusOK)
		c.sendResponse(writer, request, common.ApiDeviceCallbackRoute, res, http.StatusOK)
	} else {
		c.sendEdgexError(writer, request, edgexErr, common.ApiDeviceCallbackRoute)
	}
}

func (c *RestController) UpdateDevice(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()

	var updateDeviceRequest requests.UpdateDeviceRequest

	err := json.NewDecoder(request.Body).Decode(&updateDeviceRequest)
	if err != nil {
		edgexErr := errors.NewCommonEdgeX(errors.KindServerError, "failed to decode JSON", err)
		c.sendEdgexError(writer, request, edgexErr, common.ApiDeviceCallbackRoute)
		return
	}

	edgexErr := application.UpdateDevice(updateDeviceRequest, c.dic)
	if edgexErr == nil {
		res := commonDTO.NewBaseResponse(updateDeviceRequest.RequestId, "", http.StatusOK)
		c.sendResponse(writer, request, common.ApiDeviceCallbackRoute, res, http.StatusOK)
	} else {
		c.sendEdgexError(writer, request, edgexErr, common.ApiDeviceCallbackRoute)
	}
}

func (c *RestController) UpdateProfile(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()

	var edgexErr errors.EdgeX
	var profileRequest requests.DeviceProfileRequest

	err := json.NewDecoder(request.Body).Decode(&profileRequest)
	if err != nil {
		edgexErr = errors.NewCommonEdgeX(errors.KindServerError, "failed to decode JSON", err)
		c.sendEdgexError(writer, request, edgexErr, common.ApiProfileCallbackRoute)
		return
	}

	edgexErr = application.UpdateProfile(profileRequest, c.lc)
	if edgexErr == nil {
		res := commonDTO.NewBaseResponse(profileRequest.RequestId, "", http.StatusOK)
		c.sendResponse(writer, request, common.ApiProfileCallbackRoute, res, http.StatusOK)
	} else {
		c.sendEdgexError(writer, request, edgexErr, common.ApiProfileCallbackRoute)
	}
}

func (c *RestController) DeleteProvisionWatcher(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	name := vars[common.Name]

	err := application.DeleteProvisionWatcher(name, c.lc)
	if err == nil {
		res := commonDTO.NewBaseResponse("", "", http.StatusOK)
		c.sendResponse(writer, request, common.ApiWatcherCallbackNameRoute, res, http.StatusOK)
	} else {
		c.sendEdgexError(writer, request, err, common.ApiWatcherCallbackNameRoute)
	}
}

func (c *RestController) AddProvisionWatcher(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()

	var addProvisionWatcherRequest requests.AddProvisionWatcherRequest

	err := json.NewDecoder(request.Body).Decode(&addProvisionWatcherRequest)
	if err != nil {
		edgexErr := errors.NewCommonEdgeX(errors.KindServerError, "failed to decode JSON", err)
		c.sendEdgexError(writer, request, edgexErr, common.ApiWatcherCallbackRoute)
		return
	}

	edgexErr := application.AddProvisionWatcher(addProvisionWatcherRequest, c.lc, c.dic)
	if edgexErr == nil {
		res := commonDTO.NewBaseResponse(addProvisionWatcherRequest.RequestId, "", http.StatusOK)
		c.sendResponse(writer, request, common.ApiWatcherCallbackRoute, res, http.StatusOK)
	} else {
		c.sendEdgexError(writer, request, edgexErr, common.ApiWatcherCallbackRoute)
	}
}

func (c *RestController) UpdateProvisionWatcher(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()

	var updateProvisionWatcherRequest requests.UpdateProvisionWatcherRequest

	err := json.NewDecoder(request.Body).Decode(&updateProvisionWatcherRequest)
	if err != nil {
		edgexErr := errors.NewCommonEdgeX(errors.KindServerError, "failed to decode JSON", err)
		c.sendEdgexError(writer, request, edgexErr, common.ApiWatcherCallbackRoute)
		return
	}

	edgexErr := application.UpdateProvisionWatcher(updateProvisionWatcherRequest, c.dic)
	if edgexErr == nil {
		res := commonDTO.NewBaseResponse(updateProvisionWatcherRequest.RequestId, "", http.StatusOK)
		c.sendResponse(writer, request, common.ApiWatcherCallbackRoute, res, http.StatusOK)
	} else {
		c.sendEdgexError(writer, request, edgexErr, common.ApiWatcherCallbackRoute)
	}
}

func (c *RestController) UpdateDeviceService(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()

	var updateDeviceServiceRequest requests.UpdateDeviceServiceRequest

	err := json.NewDecoder(request.Body).Decode(&updateDeviceServiceRequest)
	if err != nil {
		edgexErr := errors.NewCommonEdgeX(errors.KindContractInvalid, "failed to decode JSON", err)
		c.sendEdgexError(writer, request, edgexErr, common.ApiServiceCallbackRoute)
		return
	}

	edgexErr := application.UpdateDeviceService(updateDeviceServiceRequest, c.dic)
	if edgexErr == nil {
		res := commonDTO.NewBaseResponse(updateDeviceServiceRequest.RequestId, "", http.StatusOK)
		c.sendResponse(writer, request, common.ApiServiceCallbackRoute, res, http.StatusOK)
	} else {
		c.sendEdgexError(writer, request, edgexErr, common.ApiServiceCallbackRoute)
	}
}
