// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package provision

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
	"github.com/google/uuid"
	"github.com/pelletier/go-toml"

	"github.com/edgexfoundry/device-sdk-go/v2/internal/cache"
	"github.com/edgexfoundry/device-sdk-go/v2/internal/container"
)

func LoadDevices(path string, dic *di.Container) errors.EdgeX {
	if path == "" {
		return nil
	}
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

	absPath, err := filepath.Abs(path)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, "failed to create absolute path", err)
	}

	fileInfo, err := os.ReadDir(absPath)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, "failed to read directory", err)
	}

	var addDevicesReq []requests.AddDeviceRequest
	serviceName := container.DeviceServiceFrom(dic.Get).Name
	lc.Infof("Loading pre-defined devices from %s", absPath)
	for _, file := range fileInfo {
		var devices []dtos.Device
		fullPath := filepath.Join(absPath, file.Name())
		if strings.HasSuffix(fullPath, ".toml") {
			content, err := os.ReadFile(fullPath)
			if err != nil {
				lc.Errorf("Failed to read %s: %v", fullPath, err)
				continue
			}
			d := struct {
				DeviceList []dtos.Device
			}{}
			err = toml.Unmarshal(content, &d)
			if err != nil {
				lc.Errorf("Failed to decode %s: %v", fullPath, err)
				continue
			}
			devices = d.DeviceList
		} else if strings.HasSuffix(fullPath, ".json") {
			content, err := os.ReadFile(fullPath)
			if err != nil {
				lc.Errorf("Failed to read %s: %v", fullPath, err)
				continue
			}
			err = json.Unmarshal(content, &devices)
			if err != nil {
				lc.Errorf("Failed to decode %s: %v", fullPath, err)
				continue
			}
		} else {
			continue
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
	}

	if len(addDevicesReq) == 0 {
		return nil
	}
	dc := bootstrapContainer.MetadataDeviceClientFrom(dic.Get)
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.NewString())
	_, edgexErr := dc.Add(ctx, addDevicesReq)
	return edgexErr
}
