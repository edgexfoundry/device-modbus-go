// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 Canonical Ltd
// Copyright (C) 2018-2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"context"
	"fmt"
	"regexp"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"

	"github.com/edgexfoundry/device-sdk-go/v4/internal/cache"
	"github.com/edgexfoundry/device-sdk-go/v4/internal/common"
	"github.com/edgexfoundry/device-sdk-go/v4/internal/container"
	"github.com/edgexfoundry/device-sdk-go/v4/internal/transformer"
	sdkModels "github.com/edgexfoundry/device-sdk-go/v4/pkg/models"
)

// processAsyncResults processes readings that are pushed from
// a DS implementation. Each is reading is optionally transformed
// before being pushed to Core Data.
// In this function, AsyncBufferSize is used to create a buffer for
// processing AsyncValues concurrently, so that events may arrive
// out-of-order in core-data / app service when AsyncBufferSize value
// is greater than or equal to two. Alternatively, we can process
// AsyncValues one by one in the same order by changing the AsyncBufferSize
// value to one.
func (s *deviceService) processAsyncResults(ctx context.Context, dic *di.Container) {
	working := make(chan bool, s.config.Device.AsyncBufferSize)
	for {
		select {
		case <-ctx.Done():
			return
		case acv := <-s.asyncCh:
			go s.sendAsyncValues(acv, working, dic)
		}
	}
}

// sendAsyncValues convert AsyncValues to event and send the event to CoreData
func (s *deviceService) sendAsyncValues(acv *sdkModels.AsyncValues, working chan bool, dic *di.Container) {
	working <- true
	defer func() {
		<-working
	}()

	// Update the LastConnected metric in deviceCache
	cache.Devices().SetLastConnectedByName(acv.DeviceName)

	if len(acv.CommandValues) == 0 {
		s.lc.Error("Skip sending AsyncValues because the CommandValues is empty.")
		return
	}
	if len(acv.CommandValues) > 1 && acv.SourceName == "" {
		s.lc.Error("Skip sending AsyncValues because the SourceName is empty.")
		return
	}
	// We can use the first reading's DeviceResourceName as the SourceName
	// when the CommandValues contains only one reading and the AsyncValues's SourceName is empty.
	if len(acv.CommandValues) == 1 && acv.SourceName == "" {
		acv.SourceName = acv.CommandValues[0].DeviceResourceName
	}

	configuration := container.ConfigurationFrom(dic.Get)
	event, err := transformer.CommandValuesToEventDTO(acv.CommandValues, acv.DeviceName, acv.SourceName, configuration.Device.DataTransform, dic)
	if err != nil {
		s.lc.Errorf("failed to transform CommandValues to Event: %v", err)
		return
	}

	common.SendEvent(event, "", dic)
}

// processAsyncFilterAndAdd filter and add devices discovered by
// device service protocol discovery.
func (s *deviceService) processAsyncFilterAndAdd(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case devices := <-s.deviceCh:
			ctx := context.Background()
			pws := cache.ProvisionWatchers().All()
			for _, d := range devices {
				for _, pw := range pws {
					if pw.AdminState == models.Locked {
						s.lc.Debugf("Skip th locked provision watcher %v", pw.Name)
						continue
					}
					if s.checkAllowList(d, pw) && s.checkBlockList(d, pw) {
						if _, ok := cache.Devices().ForName(d.Name); ok {
							s.lc.Debugf("Candidate discovered device %s already existed", d.Name)
							break
						}

						s.lc.Infof("Adding discovered device %s to Metadata", d.Name)
						device := models.Device{
							Name:           d.Name,
							Description:    d.Description,
							ProfileName:    pw.DiscoveredDevice.ProfileName,
							Protocols:      d.Protocols,
							Labels:         d.Labels,
							ServiceName:    s.serviceKey,
							AdminState:     pw.DiscoveredDevice.AdminState,
							OperatingState: models.Up,
							AutoEvents:     pw.DiscoveredDevice.AutoEvents,
							Properties:     pw.DiscoveredDevice.Properties,
						}

						req := requests.NewAddDeviceRequest(dtos.FromDeviceModelToDTO(device))
						_, err := bootstrapContainer.DeviceClientFrom(s.dic.Get).Add(ctx, []requests.AddDeviceRequest{req})
						if err != nil {
							s.lc.Errorf("failed to create discovered device %s: %v", device.Name, err)
						} else {
							break
						}
					}
				}
			}
			s.lc.Debug("Filtered device addition finished")
		}
	}
}

func (s *deviceService) checkAllowList(d sdkModels.DiscoveredDevice, pw models.ProvisionWatcher) bool {
	// ignore the device protocol properties name
	for _, protocol := range d.Protocols {
		matchedCount := 0
		for name, regex := range pw.Identifiers {
			if value, ok := protocol[name]; ok {
				valueString := fmt.Sprintf("%v", value)
				if valueString == "" {
					s.lc.Debugf("Skipping identifier %s, cannot transform %s value '%v' to string type for discovered device %s", name, name, value, d.Name)
					continue
				}
				matched, err := regexp.MatchString(regex, valueString)
				if !matched || err != nil {
					s.lc.Debugf("Discovered Device %s %s value '%v' did not match PW identifier: %s", d.Name, name, value, regex)
					break
				}
				matchedCount += 1
			}
		}
		// match succeed on all identifiers
		if matchedCount == len(pw.Identifiers) {
			return true
		}
	}
	return false
}

func (s *deviceService) checkBlockList(d sdkModels.DiscoveredDevice, pw models.ProvisionWatcher) bool {
	// a candidate should match none of the blocking identifiers
	for name, blacklist := range pw.BlockingIdentifiers {
		// ignore the device protocol properties name
		for _, protocol := range d.Protocols {
			if value, ok := protocol[name]; ok {
				valueString := fmt.Sprintf("%v", value)
				if valueString == "" {
					s.lc.Debugf("Skipping identifier %s, cannot transform %s value '%v' to string type for discovered device %s", name, name, value, d.Name)
					continue
				}
				for _, v := range blacklist {
					if valueString == v {
						s.lc.Debugf("Discovered Device %s %s value cannot be %v", d.Name, name, value)
						return false
					}
				}
			}
		}
	}
	return true
}
