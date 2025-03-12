// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018-2025 IOTech Ltd
// Copyright (c) 2019 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/handlers"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"

	"github.com/labstack/echo/v4"
)

type RestController struct {
	serviceName    string
	router         *echo.Echo
	reservedRoutes map[string]bool
	customConfig   interfaces.UpdatableConfig
	lc             logger.LoggingClient
	dic            *di.Container
}

func NewRestController(r *echo.Echo, dic *di.Container, serviceName string) *RestController {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	return &RestController{
		lc:             lc,
		router:         r,
		reservedRoutes: make(map[string]bool),
		dic:            dic,
		serviceName:    serviceName,
	}
}

// SetCustomConfigInfo sets the custom configuration, which is used to include the service's custom config in the /config endpoint response.
func (c *RestController) SetCustomConfigInfo(customConfig interfaces.UpdatableConfig) {
	c.customConfig = customConfig
}

func (c *RestController) InitRestRoutes(dic *di.Container) {
	c.lc.Info("Registering routes...")

	authenticationHook := handlers.AutoConfigAuthenticationFunc(dic)

	// discovery
	c.addReservedRoute(common.ApiDiscoveryRoute, c.Discovery, http.MethodPost, authenticationHook)
	c.addReservedRoute(common.ApiProfileScanRoute, c.ProfileScan, http.MethodPost, authenticationHook)
	c.addReservedRoute(common.ApiDiscoveryRoute, c.StopDeviceDiscovery, http.MethodDelete, authenticationHook)
	c.addReservedRoute(common.ApiDiscoveryByIdRoute, c.StopDeviceDiscovery, http.MethodDelete, authenticationHook)
	c.addReservedRoute(common.ApiProfileScanByDeviceNameRoute, c.StopProfileScan, http.MethodDelete, authenticationHook)
	// device command
	c.addReservedRoute(common.ApiDeviceNameCommandNameRoute, c.GetCommand, http.MethodGet, authenticationHook)
	c.addReservedRoute(common.ApiDeviceNameCommandNameRoute, c.SetCommand, http.MethodPut, authenticationHook)
}

func (c *RestController) addReservedRoute(route string, handler func(e echo.Context) error, method string,
	middlewareFunc ...echo.MiddlewareFunc) *echo.Route {
	c.reservedRoutes[route] = true
	return c.router.Add(method, route, handler, middlewareFunc...)
}

func (c *RestController) AddRoute(route string, handler func(e echo.Context) error, methods []string, middlewareFunc ...echo.MiddlewareFunc) errors.EdgeX {
	if c.reservedRoutes[route] {
		return errors.NewCommonEdgeX(errors.KindServerError, "route is reserved", nil)
	}

	c.router.Match(methods, route, handler, middlewareFunc...)
	c.lc.Debug("Route added", "route", route, "methods", fmt.Sprintf("%v", methods))

	return nil
}

func (c *RestController) Router() *echo.Echo {
	return c.router
}

// sendResponse puts together the response packet for the V2 API
func (c *RestController) sendResponse(
	writer *echo.Response,
	request *http.Request,
	api string,
	response interface{},
	statusCode int) error {

	correlationID := request.Header.Get(common.CorrelationHeader)

	writer.Header().Set(common.CorrelationHeader, correlationID)
	writer.Header().Set(common.ContentType, common.ContentTypeJSON)
	writer.WriteHeader(statusCode)

	if response != nil {
		data, err := json.Marshal(response)
		if err != nil {
			c.lc.Error(fmt.Sprintf("Unable to marshal %s response", api), "error", err.Error(), common.CorrelationHeader, correlationID)
			// set Response.Committed to false in order to rewrite the status code
			writer.Committed = false
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		_, err = writer.Write(data)
		if err != nil {
			c.lc.Error(fmt.Sprintf("Unable to write %s response", api), "error", err.Error(), common.CorrelationHeader, correlationID)
			// set Response.Committed to false in order to rewrite the status code
			writer.Committed = false
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	}
	return nil
}

// sendEventResponse puts together the EventResponse packet for the V2 API
func (c *RestController) sendEventResponse(
	writer *echo.Response,
	request *http.Request,
	response responses.EventResponse,
	statusCode int) error {

	correlationID := request.Header.Get(common.CorrelationHeader)
	data, encoding, err := response.Encode()
	if err != nil {
		c.lc.Errorf("Unable to marshal EventResponse: %s; %s: %s", err.Error(), common.CorrelationHeader, correlationID)
		// set Response.Committed to false in order to rewrite the status code
		writer.Committed = false
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	writer.Header().Set(common.CorrelationHeader, correlationID)
	writer.Header().Set(common.ContentType, encoding)
	writer.WriteHeader(statusCode)

	_, err = writer.Write(data)
	if err != nil {
		c.lc.Errorf("Unable to write DeviceCommand response: %s; %s: %s", err.Error(), common.CorrelationHeader, correlationID)
		// set Response.Committed to false in order to rewrite the status code
		writer.Committed = false
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return nil
}

func (c *RestController) sendEdgexError(
	writer *echo.Response,
	request *http.Request,
	err errors.EdgeX,
	api string) error {
	return c.sendEdgexErrorWithRequestId(writer, request, err, api, "")
}

func (c *RestController) sendEdgexErrorWithRequestId(
	writer *echo.Response,
	request *http.Request,
	err errors.EdgeX,
	api string,
	requestId string) error {
	correlationID := request.Header.Get(common.CorrelationHeader)
	c.lc.Error(err.Error(), common.CorrelationHeader, correlationID)
	c.lc.Debug(err.DebugMessages(), common.CorrelationHeader, correlationID)
	response := commonDTO.NewBaseResponse(requestId, err.Error(), err.Code())
	return c.sendResponse(writer, request, api, response, err.Code())
}
