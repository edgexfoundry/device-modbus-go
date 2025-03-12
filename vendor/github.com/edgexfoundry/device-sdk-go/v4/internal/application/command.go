// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020-2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/http/utils"

	"github.com/edgexfoundry/device-sdk-go/v4/internal/cache"
	sdkCommon "github.com/edgexfoundry/device-sdk-go/v4/internal/common"
	"github.com/edgexfoundry/device-sdk-go/v4/internal/container"
	"github.com/edgexfoundry/device-sdk-go/v4/internal/transformer"
	sdkModels "github.com/edgexfoundry/device-sdk-go/v4/pkg/models"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
)

func GetCommand(ctx context.Context, deviceName string, commandName string, queryParams string, regexCmd bool, dic *di.Container) (res *dtos.Event, err errors.EdgeX) {
	if deviceName == "" {
		return nil, errors.NewCommonEdgeX(errors.KindContractInvalid, "device name is empty", nil)
	}
	if commandName == "" {
		return nil, errors.NewCommonEdgeX(errors.KindContractInvalid, "command is empty", nil)
	}
	var device models.Device
	defer func() {
		if err != nil {
			DeviceRequestFailed(deviceName, dic)
		} else {
			DeviceRequestSucceeded(device, dic)
		}
	}()

	device, err = validateServiceAndDeviceState(deviceName, dic)
	if err != nil {
		return nil, errors.NewCommonEdgeXWrapper(err)
	}

	_, cmdExist := cache.Profiles().DeviceCommand(device.ProfileName, commandName)
	if cmdExist {
		res, err = readDeviceCommand(device, commandName, queryParams, dic)
	} else if regexCmd {
		res, err = readDeviceResourcesRegex(device, commandName, queryParams, dic)
	} else {
		res, err = readDeviceResource(device, commandName, queryParams, dic)
	}

	if err != nil {
		return nil, errors.NewCommonEdgeXWrapper(err)
	}

	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	lc.Debugf("GET Device Command successfully. Device: %s, Source: %s, %s: %s", deviceName, commandName, common.CorrelationHeader, utils.FromContext(ctx, common.CorrelationHeader))

	cache.Devices().SetLastConnectedByName(deviceName)
	return res, nil
}

func SetCommand(ctx context.Context, deviceName string, commandName string, queryParams string, requests map[string]any, dic *di.Container) (event *dtos.Event, err errors.EdgeX) {
	if deviceName == "" {
		return nil, errors.NewCommonEdgeX(errors.KindContractInvalid, "device name is empty", nil)
	}
	if commandName == "" {
		return nil, errors.NewCommonEdgeX(errors.KindContractInvalid, "command is empty", nil)
	}
	var device models.Device
	defer func() {
		if err != nil {
			DeviceRequestFailed(deviceName, dic)
		} else {
			DeviceRequestSucceeded(device, dic)
		}
	}()

	device, err = validateServiceAndDeviceState(deviceName, dic)
	if err != nil {
		return nil, errors.NewCommonEdgeXWrapper(err)
	}

	_, cmdExist := cache.Profiles().DeviceCommand(device.ProfileName, commandName)
	if cmdExist {
		event, err = writeDeviceCommand(device, commandName, queryParams, requests, dic)
	} else {
		event, err = writeDeviceResource(device, commandName, queryParams, requests, dic)
	}

	if err != nil {
		return nil, errors.NewCommonEdgeXWrapper(err)
	}

	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	lc.Debugf("SET Device Command successfully. Device: %s, Source: %s, %s: %s", deviceName, commandName, common.CorrelationHeader, utils.FromContext(ctx, common.CorrelationHeader))

	cache.Devices().SetLastConnectedByName(deviceName)
	return event, nil
}

func readDeviceResource(device models.Device, resourceName string, attributes string, dic *di.Container) (*dtos.Event, errors.EdgeX) {
	dr, ok := cache.Profiles().DeviceResource(device.ProfileName, resourceName)
	if !ok {
		errMsg := fmt.Sprintf("DeviceResource %s not found", resourceName)
		return nil, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, errMsg, nil)
	}
	// check deviceResource is not write-only
	if dr.Properties.ReadWrite == common.ReadWrite_W {
		errMsg := fmt.Sprintf("DeviceResource %s is marked as write-only", dr.Name)
		return nil, errors.NewCommonEdgeX(errors.KindNotAllowed, errMsg, nil)
	}

	var req sdkModels.CommandRequest
	var reqs []sdkModels.CommandRequest

	// prepare CommandRequest
	req.DeviceResourceName = dr.Name
	req.Attributes = dr.Attributes
	if attributes != "" {
		if len(req.Attributes) <= 0 {
			req.Attributes = make(map[string]interface{})
		}
		req.Attributes[sdkCommon.URLRawQuery] = attributes
	}
	req.Type = dr.Properties.ValueType
	reqs = append(reqs, req)

	// execute protocol-specific read operation
	driver := container.ProtocolDriverFrom(dic.Get)
	results, err := driver.HandleReadCommands(device.Name, device.Protocols, reqs)
	if err != nil {
		errMsg := fmt.Sprintf("error reading DeviceResource %s for %s", dr.Name, device.Name)
		return nil, errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
	}

	// convert CommandValue to Event
	configuration := container.ConfigurationFrom(dic.Get)
	event, err := transformer.CommandValuesToEventDTO(results, device.Name, dr.Name, configuration.Device.DataTransform, dic)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, "failed to convert CommandValue to Event", err)
	}

	return event, nil
}

func readDeviceResourcesRegex(device models.Device, regexResourceName string, attributes string, dic *di.Container) (*dtos.Event, errors.EdgeX) {
	regex, err := regexp.CompilePOSIX(regexResourceName)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindContractInvalid, "failed to CompilePOSIX resource name", err)
	}

	deviceResources, ok := cache.Profiles().DeviceResourcesByRegex(device.ProfileName, regex)
	if !ok || len(deviceResources) == 0 {
		errMsg := fmt.Sprintf("Regex DeviceResource %s not found", regexResourceName)
		return nil, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, errMsg, nil)
	}

	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	reqs := make([]sdkModels.CommandRequest, 0)
	for _, dr := range deviceResources {
		// check deviceResource is not write-only
		if dr.Properties.ReadWrite == common.ReadWrite_W {
			lc.Debugf("DeviceResource %s is marked as write-only, skipping adding to RegEx Read list", dr.Name)
			continue
		}

		// prepare CommandRequest
		var req sdkModels.CommandRequest
		req.DeviceResourceName = dr.Name
		req.Attributes = dr.Attributes
		if attributes != "" {
			if len(req.Attributes) <= 0 {
				req.Attributes = make(map[string]any)
			}
			req.Attributes[sdkCommon.URLRawQuery] = attributes
		}
		req.Type = dr.Properties.ValueType

		reqs = append(reqs, req)
	}

	if len(reqs) == 0 {
		errMsg := fmt.Sprintf("no readable resources matched with %s", regexResourceName)
		return nil, errors.NewCommonEdgeX(errors.KindNotAllowed, errMsg, nil)
	}

	// execute protocol-specific read operation
	driver := container.ProtocolDriverFrom(dic.Get)
	results, err := driver.HandleReadCommands(device.Name, device.Protocols, reqs)
	if err != nil {
		errMsg := fmt.Sprintf("error reading Regex DeviceResource(s) %s for %s", regexResourceName, device.Name)
		return nil, errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
	}

	// convert CommandValue to Event
	configuration := container.ConfigurationFrom(dic.Get)
	event, err := transformer.CommandValuesToEventDTO(results, device.Name, regexResourceName, configuration.Device.DataTransform, dic)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, "failed to convert CommandValue to Event", err)
	}

	return event, nil
}

func readDeviceCommand(device models.Device, commandName string, attributes string, dic *di.Container) (*dtos.Event, errors.EdgeX) {
	dc, ok := cache.Profiles().DeviceCommand(device.ProfileName, commandName)
	if !ok {
		errMsg := fmt.Sprintf("DeviceCommand %s not found", commandName)
		return nil, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, errMsg, nil)
	}
	// check deviceCommand is not write-only
	if dc.ReadWrite == common.ReadWrite_W {
		errMsg := fmt.Sprintf("DeviceCommand %s is marked as write-only", dc.Name)
		return nil, errors.NewCommonEdgeX(errors.KindNotAllowed, errMsg, nil)
	}
	// check ResourceOperation count does not exceed MaxCmdOps defined in configuration
	configuration := container.ConfigurationFrom(dic.Get)
	if len(dc.ResourceOperations) > configuration.Device.MaxCmdOps {
		errMsg := fmt.Sprintf("GET command %s exceed device %s MaxCmdOps (%d)", dc.Name, device.Name, configuration.Device.MaxCmdOps)
		return nil, errors.NewCommonEdgeX(errors.KindServerError, errMsg, nil)
	}

	// prepare CommandRequests
	reqs := make([]sdkModels.CommandRequest, len(dc.ResourceOperations))
	for i, op := range dc.ResourceOperations {
		drName := op.DeviceResource
		// check the deviceResource in ResourceOperation actually exist
		dr, ok := cache.Profiles().DeviceResource(device.ProfileName, drName)
		if !ok {
			errMsg := fmt.Sprintf("DeviceResource %s in GET commnd %s for %s not defined", drName, dc.Name, device.Name)
			return nil, errors.NewCommonEdgeX(errors.KindServerError, errMsg, nil)
		}

		reqs[i].DeviceResourceName = dr.Name
		reqs[i].Attributes = dr.Attributes
		if attributes != "" {
			if len(reqs[i].Attributes) <= 0 {
				reqs[i].Attributes = make(map[string]interface{})
			}
			reqs[i].Attributes[sdkCommon.URLRawQuery] = attributes
		}
		reqs[i].Type = dr.Properties.ValueType
	}

	// execute protocol-specific read operation
	driver := container.ProtocolDriverFrom(dic.Get)
	results, err := driver.HandleReadCommands(device.Name, device.Protocols, reqs)
	if err != nil {
		errMsg := fmt.Sprintf("error reading DeviceCommand %s for %s", dc.Name, device.Name)
		return nil, errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
	}

	// convert CommandValue to Event
	event, err := transformer.CommandValuesToEventDTO(results, device.Name, dc.Name, configuration.Device.DataTransform, dic)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, "failed to transform CommandValue to Event", err)
	}

	return event, nil
}

func writeDeviceResource(device models.Device, resourceName string, attributes string, requests map[string]any, dic *di.Container) (*dtos.Event, errors.EdgeX) {
	dr, ok := cache.Profiles().DeviceResource(device.ProfileName, resourceName)
	if !ok {
		errMsg := fmt.Sprintf("DeviceResource %s not found", resourceName)
		return nil, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, errMsg, nil)
	}
	// check deviceResource is not read-only
	if dr.Properties.ReadWrite == common.ReadWrite_R {
		errMsg := fmt.Sprintf("DeviceResource %s is marked as read-only", dr.Name)
		return nil, errors.NewCommonEdgeX(errors.KindNotAllowed, errMsg, nil)
	}

	// check set parameters contains provided deviceResource
	v, ok := requests[dr.Name]
	if !ok {
		if dr.Properties.DefaultValue != "" {
			v = dr.Properties.DefaultValue
		} else {
			errMsg := fmt.Sprintf("DeviceResource %s not found in request body and no default value defined", dr.Name)
			return nil, errors.NewCommonEdgeX(errors.KindServerError, errMsg, nil)
		}
	}

	// create CommandValue
	cv, edgexErr := createCommandValueFromDeviceResource(dr, v)
	if edgexErr != nil {
		return nil, errors.NewCommonEdgeX(errors.KindContractInvalid, "failed to create CommandValue", edgexErr)
	}

	// prepare CommandRequest
	reqs := make([]sdkModels.CommandRequest, 1)
	reqs[0].DeviceResourceName = cv.DeviceResourceName
	reqs[0].Attributes = dr.Attributes
	if attributes != "" {
		if len(reqs[0].Attributes) <= 0 {
			reqs[0].Attributes = make(map[string]any)
		}
		reqs[0].Attributes[sdkCommon.URLRawQuery] = attributes
	}
	reqs[0].Type = cv.Type

	// transform write value
	configuration := container.ConfigurationFrom(dic.Get)
	if configuration.Device.DataTransform {
		edgexErr = transformer.TransformWriteParameter(cv, dr.Properties)
		if edgexErr != nil {
			return nil, errors.NewCommonEdgeX(errors.KindContractInvalid, "failed to transform set parameter", edgexErr)
		}
	}

	// execute protocol-specific write operation
	driver := container.ProtocolDriverFrom(dic.Get)
	err := driver.HandleWriteCommands(device.Name, device.Protocols, reqs, []*sdkModels.CommandValue{cv})
	if err != nil {
		errMsg := fmt.Sprintf("error writing DeviceResource %s for %s", dr.Name, device.Name)
		return nil, errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
	}

	// Updated resource value will be published to MessageBus as long as it's not write-only
	if dr.Properties.ReadWrite != common.ReadWrite_W {
		return transformer.CommandValuesToEventDTO([]*sdkModels.CommandValue{cv}, device.Name, resourceName, configuration.Device.DataTransform, dic)
	}

	return nil, nil
}

func writeDeviceCommand(device models.Device, commandName string, attributes string, requests map[string]any, dic *di.Container) (*dtos.Event, errors.EdgeX) {
	dc, ok := cache.Profiles().DeviceCommand(device.ProfileName, commandName)
	if !ok {
		errMsg := fmt.Sprintf("DeviceCommand %s not found", commandName)
		return nil, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, errMsg, nil)
	}
	// check deviceCommand is not read-only
	if dc.ReadWrite == common.ReadWrite_R {
		errMsg := fmt.Sprintf("DeviceCommand %s is marked as read-only", dc.Name)
		return nil, errors.NewCommonEdgeX(errors.KindNotAllowed, errMsg, nil)
	}
	// check ResourceOperation count does not exceed MaxCmdOps defined in configuration
	configuration := container.ConfigurationFrom(dic.Get)
	if len(dc.ResourceOperations) > configuration.Device.MaxCmdOps {
		errMsg := fmt.Sprintf("SET command %s exceed device %s MaxCmdOps (%d)", dc.Name, device.Name, configuration.Device.MaxCmdOps)
		return nil, errors.NewCommonEdgeX(errors.KindServerError, errMsg, nil)
	}

	// create CommandValues
	cvs := make([]*sdkModels.CommandValue, 0, len(requests))
	for _, ro := range dc.ResourceOperations {
		drName := ro.DeviceResource
		// check the deviceResource in ResourceOperation actually exist
		dr, ok := cache.Profiles().DeviceResource(device.ProfileName, drName)
		if !ok {
			errMsg := fmt.Sprintf("DeviceResource %s in SET commnd %s for %s not defined", drName, dc.Name, device.Name)
			return nil, errors.NewCommonEdgeX(errors.KindServerError, errMsg, nil)
		}

		// check request body contains the deviceResource
		value, ok := requests[ro.DeviceResource]
		if !ok {
			if ro.DefaultValue != "" {
				value = ro.DefaultValue
			} else if dr.Properties.DefaultValue != "" {
				value = dr.Properties.DefaultValue
			} else {
				errMsg := fmt.Sprintf("DeviceResource %s not found in request body and no default value defined", dr.Name)
				return nil, errors.NewCommonEdgeX(errors.KindServerError, errMsg, nil)
			}
		}

		// ResourceOperation mapping, notice that the order is opposite to get command mapping
		// i.e. the mapping value is actually the key for set command.
		if len(ro.Mappings) > 0 {
			for k, v := range ro.Mappings {
				if v == value {
					value = k
					break
				}
			}
		}

		// create CommandValue
		cv, err := createCommandValueFromDeviceResource(dr, value)
		if err == nil {
			cvs = append(cvs, cv)
		} else {
			return nil, errors.NewCommonEdgeX(errors.KindContractInvalid, "failed to create CommandValue", err)
		}
	}

	// prepare CommandRequests
	reqs := make([]sdkModels.CommandRequest, len(cvs))
	for i, cv := range cvs {
		dr, _ := cache.Profiles().DeviceResource(device.ProfileName, cv.DeviceResourceName)

		reqs[i].DeviceResourceName = cv.DeviceResourceName
		reqs[i].Attributes = dr.Attributes
		if attributes != "" {
			if len(reqs[i].Attributes) <= 0 {
				reqs[i].Attributes = make(map[string]interface{})
			}
			reqs[i].Attributes[sdkCommon.URLRawQuery] = attributes
		}
		reqs[i].Type = cv.Type

		// transform write value
		if configuration.Device.DataTransform {
			err := transformer.TransformWriteParameter(cv, dr.Properties)
			if err != nil {
				return nil, errors.NewCommonEdgeX(errors.KindContractInvalid, "failed to transform set parameter", err)
			}
		}
	}

	// execute protocol-specific write operation
	driver := container.ProtocolDriverFrom(dic.Get)
	err := driver.HandleWriteCommands(device.Name, device.Protocols, reqs, cvs)
	if err != nil {
		errMsg := fmt.Sprintf("error writing DeviceCommand %s for %s", dc.Name, device.Name)
		return nil, errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
	}

	// Updated resource(s) value will be published to MessageBus as long as they're not write-only
	if dc.ReadWrite != common.ReadWrite_W {
		return transformer.CommandValuesToEventDTO(cvs, device.Name, commandName, configuration.Device.DataTransform, dic)
	}

	return nil, nil
}

func validateServiceAndDeviceState(deviceName string, dic *di.Container) (models.Device, errors.EdgeX) {
	// check device service AdminState
	ds := container.DeviceServiceFrom(dic.Get)
	if ds.AdminState == models.Locked {
		return models.Device{}, errors.NewCommonEdgeX(errors.KindServiceLocked, "service locked", nil)
	}

	// check requested device exists
	device, ok := cache.Devices().ForName(deviceName)
	if !ok {
		return models.Device{}, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("device %s not found", deviceName), nil)
	}

	// check device's AdminState
	if device.AdminState == models.Locked {
		return models.Device{}, errors.NewCommonEdgeX(errors.KindServiceLocked, fmt.Sprintf("device %s locked", device.Name), nil)
	}
	// check device's OperatingState
	// if it's a device return attempt, operating state is allowed to be DOWN
	if device.OperatingState == models.Down {
		err := errors.NewCommonEdgeX(errors.KindServiceLocked, fmt.Sprintf("device %s OperatingState is DOWN", device.Name), nil)
		config := container.ConfigurationFrom(dic.Get)
		if config.Device.AllowedFails == 0 || config.Device.DeviceDownTimeout == 0 {
			return models.Device{}, err
		}
		reqFailsTracker := container.AllowedRequestFailuresTrackerFrom(dic.Get)
		if reqFailsTracker.Value(deviceName) > 0 {
			return models.Device{}, err
		}
	}

	// check device's ProfileName
	if device.ProfileName == "" {
		return models.Device{}, errors.NewCommonEdgeX(errors.KindServiceLocked, "no associated device profile", nil)
	}

	return device, nil
}

func createCommandValueFromDeviceResource(dr models.DeviceResource, value interface{}) (*sdkModels.CommandValue, errors.EdgeX) {
	if value == nil {
		return &sdkModels.CommandValue{
			DeviceResourceName: dr.Name,
			Type:               dr.Properties.ValueType,
			Value:              value,
			Tags:               make(map[string]string)}, nil
	}

	var err error
	var result *sdkModels.CommandValue

	v := fmt.Sprint(value)

	if dr.Properties.ValueType != common.ValueTypeString && strings.TrimSpace(v) == "" {
		return nil, errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("empty string is invalid for %v value type", dr.Properties.ValueType), nil)
	}

	switch dr.Properties.ValueType {
	case common.ValueTypeString:
		result, err = sdkModels.NewCommandValue(dr.Name, common.ValueTypeString, v)
	case common.ValueTypeStringArray:
		var arr []string
		err = json.Unmarshal([]byte(v), &arr)
		if err != nil {
			errMsg := fmt.Sprintf("failed to convert set parameter %s to ValueType %s", v, dr.Properties.ValueType)
			return result, errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
		}
		result, err = sdkModels.NewCommandValue(dr.Name, common.ValueTypeStringArray, arr)
	case common.ValueTypeBool:
		var value bool
		value, err = strconv.ParseBool(v)
		if err != nil {
			errMsg := fmt.Sprintf("failed to convert set parameter %s to ValueType %s", v, dr.Properties.ValueType)
			return result, errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
		}
		result, err = sdkModels.NewCommandValue(dr.Name, common.ValueTypeBool, value)
	case common.ValueTypeBoolArray:
		var arr []bool
		err = json.Unmarshal([]byte(v), &arr)
		if err != nil {
			errMsg := fmt.Sprintf("failed to convert set parameter %s to ValueType %s", v, dr.Properties.ValueType)
			return result, errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
		}
		result, err = sdkModels.NewCommandValue(dr.Name, common.ValueTypeBoolArray, arr)
	case common.ValueTypeUint8:
		var n uint64
		n, err = strconv.ParseUint(v, 10, 8)
		if err != nil {
			errMsg := fmt.Sprintf("failed to convert set parameter %s to ValueType %s", v, dr.Properties.ValueType)
			return result, errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
		}
		result, err = sdkModels.NewCommandValue(dr.Name, common.ValueTypeUint8, uint8(n))
	case common.ValueTypeUint8Array:
		var arr []uint8
		strArr := strings.Split(strings.Trim(v, "[]"), ",")
		for _, u := range strArr {
			n, err := strconv.ParseUint(strings.Trim(u, " "), 10, 8)
			if err != nil {
				errMsg := fmt.Sprintf("failed to convert set parameter %s to ValueType %s", v, dr.Properties.ValueType)
				return result, errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
			}
			arr = append(arr, uint8(n))
		}
		result, err = sdkModels.NewCommandValue(dr.Name, common.ValueTypeUint8Array, arr)
	case common.ValueTypeUint16:
		var n uint64
		n, err = strconv.ParseUint(v, 10, 16)
		if err != nil {
			errMsg := fmt.Sprintf("failed to convert set parameter %s to ValueType %s", v, dr.Properties.ValueType)
			return result, errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
		}
		result, err = sdkModels.NewCommandValue(dr.Name, common.ValueTypeUint16, uint16(n))
	case common.ValueTypeUint16Array:
		var arr []uint16
		strArr := strings.Split(strings.Trim(v, "[]"), ",")
		for _, u := range strArr {
			n, err := strconv.ParseUint(strings.Trim(u, " "), 10, 16)
			if err != nil {
				errMsg := fmt.Sprintf("failed to convert set parameter %s to ValueType %s", v, dr.Properties.ValueType)
				return result, errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
			}
			arr = append(arr, uint16(n))
		}
		result, err = sdkModels.NewCommandValue(dr.Name, common.ValueTypeUint16Array, arr)
	case common.ValueTypeUint32:
		var n uint64
		n, err = strconv.ParseUint(v, 10, 32)
		if err != nil {
			errMsg := fmt.Sprintf("failed to convert set parameter %s to ValueType %s", v, dr.Properties.ValueType)
			return result, errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
		}
		result, err = sdkModels.NewCommandValue(dr.Name, common.ValueTypeUint32, uint32(n))
	case common.ValueTypeUint32Array:
		var arr []uint32
		strArr := strings.Split(strings.Trim(v, "[]"), ",")
		for _, u := range strArr {
			n, err := strconv.ParseUint(strings.Trim(u, " "), 10, 32)
			if err != nil {
				errMsg := fmt.Sprintf("failed to convert set parameter %s to ValueType %s", v, dr.Properties.ValueType)
				return result, errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
			}
			arr = append(arr, uint32(n))
		}
		result, err = sdkModels.NewCommandValue(dr.Name, common.ValueTypeUint32Array, arr)
	case common.ValueTypeUint64:
		var n uint64
		n, err = strconv.ParseUint(v, 10, 64)
		if err != nil {
			errMsg := fmt.Sprintf("failed to convert set parameter %s to ValueType %s", v, dr.Properties.ValueType)
			return result, errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
		}
		result, err = sdkModels.NewCommandValue(dr.Name, common.ValueTypeUint64, n)
	case common.ValueTypeUint64Array:
		var arr []uint64
		strArr := strings.Split(strings.Trim(v, "[]"), ",")
		for _, u := range strArr {
			n, err := strconv.ParseUint(strings.Trim(u, " "), 10, 64)
			if err != nil {
				errMsg := fmt.Sprintf("failed to convert set parameter %s to ValueType %s", v, dr.Properties.ValueType)
				return result, errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
			}
			arr = append(arr, n)
		}
		result, err = sdkModels.NewCommandValue(dr.Name, common.ValueTypeUint64Array, arr)
	case common.ValueTypeInt8:
		var n int64
		n, err = strconv.ParseInt(v, 10, 8)
		if err != nil {
			errMsg := fmt.Sprintf("failed to convert set parameter %s to ValueType %s", v, dr.Properties.ValueType)
			return result, errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
		}
		result, err = sdkModels.NewCommandValue(dr.Name, common.ValueTypeInt8, int8(n))
	case common.ValueTypeInt8Array:
		var arr []int8
		err = json.Unmarshal([]byte(v), &arr)
		if err != nil {
			errMsg := fmt.Sprintf("failed to convert set parameter %s to ValueType %s", v, dr.Properties.ValueType)
			return result, errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
		}
		result, err = sdkModels.NewCommandValue(dr.Name, common.ValueTypeInt8Array, arr)
	case common.ValueTypeInt16:
		var n int64
		n, err = strconv.ParseInt(v, 10, 16)
		if err != nil {
			errMsg := fmt.Sprintf("failed to convert set parameter %s to ValueType %s", v, dr.Properties.ValueType)
			return result, errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
		}
		result, err = sdkModels.NewCommandValue(dr.Name, common.ValueTypeInt16, int16(n))
	case common.ValueTypeInt16Array:
		var arr []int16
		err = json.Unmarshal([]byte(v), &arr)
		if err != nil {
			errMsg := fmt.Sprintf("failed to convert set parameter %s to ValueType %s", v, dr.Properties.ValueType)
			return result, errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
		}
		result, err = sdkModels.NewCommandValue(dr.Name, common.ValueTypeInt16Array, arr)
	case common.ValueTypeInt32:
		var n int64
		n, err = strconv.ParseInt(v, 10, 32)
		if err != nil {
			errMsg := fmt.Sprintf("failed to convert set parameter %s to ValueType %s", v, dr.Properties.ValueType)
			return result, errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
		}
		result, err = sdkModels.NewCommandValue(dr.Name, common.ValueTypeInt32, int32(n))
	case common.ValueTypeInt32Array:
		var arr []int32
		err = json.Unmarshal([]byte(v), &arr)
		if err != nil {
			errMsg := fmt.Sprintf("failed to convert set parameter %s to ValueType %s", v, dr.Properties.ValueType)
			return result, errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
		}
		result, err = sdkModels.NewCommandValue(dr.Name, common.ValueTypeInt32Array, arr)
	case common.ValueTypeInt64:
		var n int64
		n, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			errMsg := fmt.Sprintf("failed to convert set parameter %s to ValueType %s", v, dr.Properties.ValueType)
			return result, errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
		}
		result, err = sdkModels.NewCommandValue(dr.Name, common.ValueTypeInt64, n)
	case common.ValueTypeInt64Array:
		var arr []int64
		err = json.Unmarshal([]byte(v), &arr)
		if err != nil {
			errMsg := fmt.Sprintf("failed to convert set parameter %s to ValueType %s", v, dr.Properties.ValueType)
			return result, errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
		}
		result, err = sdkModels.NewCommandValue(dr.Name, common.ValueTypeInt64Array, arr)
	case common.ValueTypeFloat32:
		var val float64
		val, err = strconv.ParseFloat(v, 32)
		if err == nil {
			result, err = sdkModels.NewCommandValue(dr.Name, common.ValueTypeFloat32, float32(val))
			break
		}
		if numError, ok := err.(*strconv.NumError); ok {
			if numError.Err == strconv.ErrRange {
				err = errors.NewCommonEdgeX(errors.KindServerError, "NumError", err)
				break
			}
		}
		var decodedToBytes []byte
		decodedToBytes, err = base64.StdEncoding.DecodeString(v)
		if err == nil {
			var val float32
			val, err = float32FromBytes(decodedToBytes)
			if err != nil {
				break
			} else if math.IsNaN(float64(val)) {
				err = fmt.Errorf("fail to parse %v to float32, unexpected result %v", v, val)
			} else {
				result, err = sdkModels.NewCommandValue(dr.Name, common.ValueTypeFloat32, val)
			}
		}
	case common.ValueTypeFloat32Array:
		var arr []float32
		err = json.Unmarshal([]byte(v), &arr)
		if err != nil {
			errMsg := fmt.Sprintf("failed to convert set parameter %s to ValueType %s", v, dr.Properties.ValueType)
			return result, errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
		}
		result, err = sdkModels.NewCommandValue(dr.Name, common.ValueTypeFloat32Array, arr)
	case common.ValueTypeFloat64:
		var val float64
		val, err = strconv.ParseFloat(v, 64)
		if err == nil {
			result, err = sdkModels.NewCommandValue(dr.Name, common.ValueTypeFloat64, val)
			break
		}
		if numError, ok := err.(*strconv.NumError); ok {
			if numError.Err == strconv.ErrRange {
				err = errors.NewCommonEdgeX(errors.KindServerError, "NumError", err)
				break
			}
		}
		var decodedToBytes []byte
		decodedToBytes, err = base64.StdEncoding.DecodeString(v)
		if err == nil {
			val, err = float64FromBytes(decodedToBytes)
			if err != nil {
				break
			} else if math.IsNaN(val) {
				err = fmt.Errorf("fail to parse %v to float64, unexpected result %v", v, val)
			} else {
				result, err = sdkModels.NewCommandValue(dr.Name, common.ValueTypeFloat64, val)
			}
		}
	case common.ValueTypeFloat64Array:
		var arr []float64
		err = json.Unmarshal([]byte(v), &arr)
		if err != nil {
			errMsg := fmt.Sprintf("failed to convert set parameter %s to ValueType %s", v, dr.Properties.ValueType)
			return result, errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
		}
		result, err = sdkModels.NewCommandValue(dr.Name, common.ValueTypeFloat64Array, arr)
	case common.ValueTypeObject:
		result, err = sdkModels.NewCommandValue(dr.Name, common.ValueTypeObject, value)
	default:
		err = errors.NewCommonEdgeX(errors.KindServerError, "unrecognized value type", nil)
	}

	if err != nil {
		return nil, errors.NewCommonEdgeXWrapper(err)
	}
	return result, nil
}

func float32FromBytes(numericValue []byte) (res float32, err error) {
	reader := bytes.NewReader(numericValue)
	err = binary.Read(reader, binary.BigEndian, &res)
	return
}

func float64FromBytes(numericValue []byte) (res float64, err error) {
	reader := bytes.NewReader(numericValue)
	err = binary.Read(reader, binary.BigEndian, &res)
	return
}
