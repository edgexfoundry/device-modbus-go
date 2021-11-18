// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018-2021 IOTech Ltd
// Copyright (c) 2019 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/gorilla/mux"

	"github.com/edgexfoundry/device-sdk-go/v2/internal/container"
	"github.com/edgexfoundry/device-sdk-go/v2/internal/controller/http/correlation"
)

type RestController struct {
	lc             logger.LoggingClient
	router         *mux.Router
	reservedRoutes map[string]bool
	dic            *di.Container
}

func NewRestController(r *mux.Router, dic *di.Container) *RestController {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	return &RestController{
		lc:             lc,
		router:         r,
		reservedRoutes: make(map[string]bool),
		dic:            dic,
	}
}

func (c *RestController) InitRestRoutes() {
	c.lc.Info("Registering v2 routes...")
	// common
	c.addReservedRoute(common.ApiPingRoute, c.Ping).Methods(http.MethodGet)
	c.addReservedRoute(common.ApiVersionRoute, c.Version).Methods(http.MethodGet)
	c.addReservedRoute(common.ApiConfigRoute, c.Config).Methods(http.MethodGet)
	c.addReservedRoute(common.ApiMetricsRoute, c.Metrics).Methods(http.MethodGet)
	// secret
	c.addReservedRoute(common.ApiSecretRoute, c.Secret).Methods(http.MethodPost)
	// discovery
	c.addReservedRoute(common.ApiDiscoveryRoute, c.Discovery).Methods(http.MethodPost)
	// device command
	c.addReservedRoute(common.ApiDeviceNameCommandNameRoute, c.Command).Methods(http.MethodPut, http.MethodGet)
	// callback
	c.addReservedRoute(common.ApiDeviceCallbackRoute, c.AddDevice).Methods(http.MethodPost)
	c.addReservedRoute(common.ApiDeviceCallbackRoute, c.UpdateDevice).Methods(http.MethodPut)
	c.addReservedRoute(common.ApiDeviceCallbackNameRoute, c.DeleteDevice).Methods(http.MethodDelete)
	c.addReservedRoute(common.ApiProfileCallbackRoute, c.UpdateProfile).Methods(http.MethodPut)
	c.addReservedRoute(common.ApiWatcherCallbackRoute, c.AddProvisionWatcher).Methods(http.MethodPost)
	c.addReservedRoute(common.ApiWatcherCallbackRoute, c.UpdateProvisionWatcher).Methods(http.MethodPut)
	c.addReservedRoute(common.ApiWatcherCallbackNameRoute, c.DeleteProvisionWatcher).Methods(http.MethodDelete)
	c.addReservedRoute(common.ApiServiceCallbackRoute, c.UpdateDeviceService).Methods(http.MethodPut)

	c.router.Use(correlation.RequestLimitMiddleware(container.ConfigurationFrom(c.dic.Get).Service.MaxRequestSize))
	c.router.Use(correlation.ManageHeader)
	c.router.Use(correlation.LoggingMiddleware(c.lc))
}

func (c *RestController) addReservedRoute(route string, handler func(http.ResponseWriter, *http.Request)) *mux.Route {
	c.reservedRoutes[route] = true
	return c.router.HandleFunc(route, handler)
}

func (c *RestController) AddRoute(route string, handler func(http.ResponseWriter, *http.Request), methods ...string) errors.EdgeX {
	if c.reservedRoutes[route] {
		return errors.NewCommonEdgeX(errors.KindServerError, "route is reserved", nil)
	}

	c.router.HandleFunc(route, handler).Methods(methods...)
	c.lc.Debug("Route added", "route", route, "methods", fmt.Sprintf("%v", methods))

	return nil
}

func (c *RestController) Router() *mux.Router {
	return c.router
}

// sendResponse puts together the response packet for the V2 API
func (c *RestController) sendResponse(
	writer http.ResponseWriter,
	request *http.Request,
	api string,
	response interface{},
	statusCode int) {

	correlationID := request.Header.Get(common.CorrelationHeader)

	writer.Header().Set(common.CorrelationHeader, correlationID)
	writer.Header().Set(common.ContentType, common.ContentTypeJSON)
	writer.WriteHeader(statusCode)

	if response != nil {
		data, err := json.Marshal(response)
		if err != nil {
			c.lc.Error(fmt.Sprintf("Unable to marshal %s response", api), "error", err.Error(), common.CorrelationHeader, correlationID)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = writer.Write(data)
		if err != nil {
			c.lc.Error(fmt.Sprintf("Unable to write %s response", api), "error", err.Error(), common.CorrelationHeader, correlationID)
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// sendEventResponse puts together the EventResponse packet for the V2 API
func (c *RestController) sendEventResponse(
	writer http.ResponseWriter,
	request *http.Request,
	response responses.EventResponse,
	statusCode int) {

	correlationID := request.Header.Get(common.CorrelationHeader)
	data, encoding, err := response.Encode()
	if err != nil {
		c.lc.Errorf("Unable to marshal EventResponse: %s; %s: %s", err.Error(), common.CorrelationHeader, correlationID)
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	writer.Header().Set(common.CorrelationHeader, correlationID)
	writer.Header().Set(common.ContentType, encoding)
	writer.WriteHeader(statusCode)

	_, err = writer.Write(data)
	if err != nil {
		c.lc.Errorf("Unable to write DeviceCommand response: %s; %s: %s", err.Error(), common.CorrelationHeader, correlationID)
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (c *RestController) sendEdgexError(
	writer http.ResponseWriter,
	request *http.Request,
	err errors.EdgeX,
	api string) {
	correlationID := request.Header.Get(common.CorrelationHeader)
	c.lc.Error(err.Error(), common.CorrelationHeader, correlationID)
	c.lc.Debug(err.DebugMessages(), common.CorrelationHeader, correlationID)
	response := commonDTO.NewBaseResponse("", err.Error(), err.Code())
	c.sendResponse(writer, request, api, response, err.Code())
}
