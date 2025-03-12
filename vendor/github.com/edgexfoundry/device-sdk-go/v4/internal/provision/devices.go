// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018-2023 IOTech Ltd
// Copyright (C) 2023 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package provision

import (
	"context"
	"encoding/json"
	"fmt"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/file"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
	"github.com/google/uuid"
	"github.com/hashicorp/go-multierror"
	"gopkg.in/yaml.v3"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/edgexfoundry/device-sdk-go/v4/internal/cache"
	"github.com/edgexfoundry/device-sdk-go/v4/internal/container"
)

func LoadDevices(path string, dic *di.Container) errors.EdgeX {
	var addDevicesReq []requests.AddDeviceRequest
	var edgexErr errors.EdgeX
	if path == "" {
		return nil
	}

	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	serviceName := container.DeviceServiceFrom(dic.Get).Name
	parsedUrl, err := url.Parse(path)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, "failed to parse Devices path as a URI", err)
	}
	if parsedUrl.Scheme == "http" || parsedUrl.Scheme == "https" {
		secretProvider := bootstrapContainer.SecretProviderFrom(dic.Get)
		addDevicesReq, edgexErr = loadDevicesFromURI(path, parsedUrl, serviceName, secretProvider, lc)
		if edgexErr != nil {
			return edgexErr
		}
	} else {
		addDevicesReq, edgexErr = loadDevicesFromFile(path, serviceName, lc)
		if edgexErr != nil {
			return edgexErr
		}
	}

	if len(addDevicesReq) == 0 {
		return nil
	}
	dc := bootstrapContainer.DeviceClientFrom(dic.Get)
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.NewString()) //nolint: staticcheck
	responses, edgexErr := dc.Add(ctx, addDevicesReq)
	if edgexErr != nil {
		return edgexErr
	}

	err = nil
	for _, response := range responses {
		if response.StatusCode != http.StatusCreated {
			if response.StatusCode == http.StatusConflict {
				lc.Warnf("%s. Device may be owned by other Device service instance.", response.Message)
				continue
			}

			err = multierror.Append(err, fmt.Errorf("add Device failed: %s", response.Message))
		}
	}

	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	return nil
}

func loadDevicesFromFile(path, serviceName string, lc logger.LoggingClient) ([]requests.AddDeviceRequest, errors.EdgeX) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, "failed to create absolute path for Devices", err)
	}

	files, err := os.ReadDir(absPath)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, "failed to read directory for Devices", err)
	}

	if len(files) == 0 {
		return nil, nil
	}

	lc.Infof("Loading pre-defined Devices from %s(%d files found)", absPath, len(files))
	var addDevicesReq, processedDevicesReq []requests.AddDeviceRequest
	for _, file := range files {
		fullPath := filepath.Join(absPath, file.Name())
		processedDevicesReq = processDevices(fullPath, fullPath, serviceName, nil, lc)
		if len(processedDevicesReq) > 0 {
			addDevicesReq = append(addDevicesReq, processedDevicesReq...)
		}
	}
	return addDevicesReq, nil
}

func loadDevicesFromURI(inputURI string, parsedURI *url.URL, serviceName string, secretProvider interfaces.SecretProvider, lc logger.LoggingClient) ([]requests.AddDeviceRequest, errors.EdgeX) {
	// the input URI contains the index file containing the Device list to be loaded
	bytes, err := file.Load(inputURI, secretProvider, lc)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to load Devices list from URI %s", parsedURI.Redacted()), err)
	}

	var files []string
	err = json.Unmarshal(bytes, &files)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, "could not unmarshal Devices list contents", err)
	}

	if len(files) == 0 {
		lc.Infof("Index file %s for Devices list is empty", parsedURI.Redacted())
		return nil, nil
	}

	lc.Infof("Loading pre-defined devices from %s(%d files found)", parsedURI.Redacted(), len(files))
	var addDevicesReq, processedDevicesReq []requests.AddDeviceRequest
	for _, file := range files {
		fullPath, redactedPath := GetFullAndRedactedURI(parsedURI, file, "Device", lc)
		processedDevicesReq = processDevices(fullPath, redactedPath, serviceName, secretProvider, lc)
		if len(processedDevicesReq) > 0 {
			addDevicesReq = append(addDevicesReq, processedDevicesReq...)
		}
	}
	return addDevicesReq, nil
}

func processDevices(fullPath, displayPath, serviceName string, secretProvider interfaces.SecretProvider, lc logger.LoggingClient) []requests.AddDeviceRequest {
	var devices []dtos.Device
	var addDevicesReq []requests.AddDeviceRequest

	fileType := GetFileType(fullPath)

	// if the file type is not yaml or json, it cannot be parsed - just return to not break the loop for other devices
	if fileType == OTHER {
		return nil
	}

	content, err := file.Load(fullPath, secretProvider, lc)
	if err != nil {
		lc.Errorf("Failed to read Devices from %s: %v", displayPath, err)
		return nil
	}

	switch fileType {
	case YAML:
		d := struct {
			DeviceList []dtos.Device `yaml:"deviceList"`
		}{}
		err = yaml.Unmarshal(content, &d)
		if err != nil {
			lc.Errorf("Failed to YAML decode Devices from %s: %v", displayPath, err)
			return nil
		}
		devices = d.DeviceList
	case JSON:
		err = json.Unmarshal(content, &devices)
		if err != nil {
			lc.Errorf("Failed to JSON decode Devices from %s: %v", displayPath, err)
			return nil
		}
	}

	for _, device := range devices {
		if _, ok := cache.Devices().ForName(device.Name); ok {
			lc.Infof("Device %s exists, using the existing one", device.Name)
		} else {
			lc.Infof("Device %s not found in Metadata, adding it ...", device.Name)
			device.ServiceName = serviceName
			device.AdminState = models.Unlocked
			device.OperatingState = models.Up
			req := requests.NewAddDeviceRequest(device)
			addDevicesReq = append(addDevicesReq, req)
		}
	}
	return addDevicesReq
}
