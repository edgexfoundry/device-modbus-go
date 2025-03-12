//
// Copyright (C) 2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"
	"fmt"
	"sync"

	"github.com/edgexfoundry/device-sdk-go/v4/internal/cache"
	"github.com/edgexfoundry/device-sdk-go/v4/internal/container"
	"github.com/edgexfoundry/device-sdk-go/v4/internal/controller/http/correlation"
	"github.com/edgexfoundry/device-sdk-go/v4/internal/utils"
	"github.com/edgexfoundry/device-sdk-go/v4/pkg/interfaces"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
)

type profileScanLocker struct {
	mux     sync.Mutex
	busyMap map[string]bool
}

var locker = profileScanLocker{busyMap: make(map[string]bool)}

func ProfileScanWrapper(busy chan bool, extdriver interfaces.ExtendedProtocolDriver, req requests.ProfileScanRequest, ctx context.Context, dic *di.Container) {
	locker.mux.Lock()
	b := locker.busyMap[req.DeviceName]
	busy <- b
	if b {
		locker.mux.Unlock()
		return
	}
	locker.busyMap[req.DeviceName] = true
	locker.mux.Unlock()

	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	dpc := bootstrapContainer.DeviceProfileClientFrom(dic.Get)
	dc := bootstrapContainer.DeviceClientFrom(dic.Get)
	if correlation.IdFromContext(ctx) != req.RequestId {
		// ensure the correlation id matches the request id.
		ctx = context.WithValue(ctx, common.CorrelationHeader, req.RequestId) //nolint: staticcheck
	}

	utils.PublishProfileScanProgressSystemEvent(req.RequestId, 0, "", ctx, dic)
	lc.Debugf("profile scan triggered with device name '%s' and profile name '%s', Correlation Id: %s", req.DeviceName, req.ProfileName, req.RequestId)
	profile, err := extdriver.ProfileScan(req)
	if err != nil {
		errMsg := fmt.Sprintf("failed to trigger profile scan: %v, Correlation Id: %s", err.Error(), req.RequestId)
		utils.PublishProfileScanProgressSystemEvent(req.RequestId, -1, errMsg, ctx, dic)
		lc.Error(errMsg)
		releaseLock(req.DeviceName)
		return
	}
	// Add profile to metadata
	profileReq := requests.NewDeviceProfileRequest(dtos.FromDeviceProfileModelToDTO(profile))
	_, err = dpc.Add(ctx, []requests.DeviceProfileRequest{profileReq})
	if err != nil {
		errMsg := fmt.Sprintf("failed to add device profile '%s': %v, Correlation Id: %s", profile.Name, err, req.RequestId)
		utils.PublishProfileScanProgressSystemEvent(req.RequestId, -1, errMsg, ctx, dic)
		lc.Error(errMsg)
		releaseLock(req.DeviceName)
		return
	}
	// Update device
	deviceReq := requests.NewUpdateDeviceRequest(dtos.UpdateDevice{Name: &req.DeviceName, ProfileName: &profile.Name})
	_, err = dc.Update(ctx, []requests.UpdateDeviceRequest{deviceReq})
	if err != nil {
		errMsg := fmt.Sprintf("failed to update device '%s' with profile '%s': %v, Correlation Id: %s", req.DeviceName, profile.Name, err, req.RequestId)
		utils.PublishProfileScanProgressSystemEvent(req.RequestId, -1, errMsg, ctx, dic)
		lc.Error(errMsg)
	} else {
		utils.PublishProfileScanProgressSystemEvent(req.RequestId, 100, "", ctx, dic)
	}

	// ReleaseLock
	releaseLock(req.DeviceName)
}

func releaseLock(deviceName string) {
	locker.mux.Lock()
	locker.busyMap[deviceName] = false
	locker.mux.Unlock()
}

func StopProfileScan(dic *di.Container, deviceName string, options map[string]any) errors.EdgeX {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	extdriver := container.ExtendedProtocolDriverFrom(dic.Get)
	if extdriver == nil {
		return errors.NewCommonEdgeX(errors.KindNotImplemented, "Stop profile scan is not implemented", nil)
	}

	// check device service AdminState
	ds := container.DeviceServiceFrom(dic.Get)
	if ds.AdminState == models.Locked {
		return errors.NewCommonEdgeX(errors.KindServiceLocked, "service locked", nil)
	}

	// check requested device exists
	_, exist := cache.Devices().ForName(deviceName)
	if !exist {
		return errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("device %s not found", deviceName), nil)
	}

	locker.mux.Lock()
	busy := locker.busyMap[deviceName]
	locker.mux.Unlock()
	if !busy {
		lc.Debugf("no active profile scan process was found")
		return nil
	}

	lc.Debugf("Stopping profile scan for device – %s", deviceName)
	extdriver.StopProfileScan(deviceName, options)
	lc.Debugf("Profile scan for device – %s stop signal is sent", deviceName)

	return nil
}
