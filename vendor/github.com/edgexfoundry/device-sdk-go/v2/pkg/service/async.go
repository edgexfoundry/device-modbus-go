// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 Canonical Ltd
// Copyright (C) 2018-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"context"
	"regexp"
	"sync"

	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"

	"github.com/edgexfoundry/device-sdk-go/v2/internal/cache"
	"github.com/edgexfoundry/device-sdk-go/v2/internal/common"
	"github.com/edgexfoundry/device-sdk-go/v2/internal/transformer"
	sdkModels "github.com/edgexfoundry/device-sdk-go/v2/pkg/models"
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
func (s *DeviceService) processAsyncResults(ctx context.Context, wg *sync.WaitGroup, dic *di.Container) {
	wg.Add(1)
	defer func() {
		wg.Done()
	}()

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
func (s *DeviceService) sendAsyncValues(acv *sdkModels.AsyncValues, working chan bool, dic *di.Container) {
	working <- true
	defer func() {
		<-working
	}()

	if len(acv.CommandValues) == 0 {
		s.LoggingClient.Error("Skip sending AsyncValues because the CommandValues is empty.")
		return
	}
	if len(acv.CommandValues) > 1 && acv.SourceName == "" {
		s.LoggingClient.Error("Skip sending AsyncValues because the SourceName is empty.")
		return
	}
	// We can use the first reading's DeviceResourceName as the SourceName
	// when the CommandValues contains only one reading and the AsyncValues's SourceName is empty.
	if len(acv.CommandValues) == 1 && acv.SourceName == "" {
		acv.SourceName = acv.CommandValues[0].DeviceResourceName
	}
	event, err := transformer.CommandValuesToEventDTO(acv.CommandValues, acv.DeviceName, acv.SourceName, dic)
	if err != nil {
		s.LoggingClient.Errorf("failed to transform CommandValues to Event: %v", err)
		return
	}

	common.SendEvent(event, "", dic)
}

// processAsyncFilterAndAdd filter and add devices discovered by
// device service protocol discovery.
func (s *DeviceService) processAsyncFilterAndAdd(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	defer func() {
		wg.Done()
	}()
	for {
		select {
		case <-ctx.Done():
			return
		case devices := <-s.deviceCh:
			ctx := context.Background()
			pws := cache.ProvisionWatchers().All()
			for _, d := range devices {
				for _, pw := range pws {
					if checkAllowList(d, pw, s.LoggingClient) && checkBlockList(d, pw, s.LoggingClient) {
						if _, ok := cache.Devices().ForName(d.Name); ok {
							s.LoggingClient.Debugf("Candidate discovered device %s already existed", d.Name)
							break
						}

						s.LoggingClient.Infof("Adding discovered device %s to Metadata", d.Name)
						device := models.Device{
							Name:           d.Name,
							Description:    d.Description,
							ProfileName:    pw.ProfileName,
							Protocols:      d.Protocols,
							Labels:         d.Labels,
							ServiceName:    pw.ServiceName,
							AdminState:     pw.AdminState,
							OperatingState: models.Up,
							AutoEvents:     pw.AutoEvents,
						}

						req := requests.NewAddDeviceRequest(dtos.FromDeviceModelToDTO(device))
						_, err := s.edgexClients.DeviceClient.Add(ctx, []requests.AddDeviceRequest{req})
						if err != nil {
							s.LoggingClient.Errorf("failed to create discovered device %s: %v", device.Name, err)
						} else {
							break
						}
					}
				}
			}
			s.LoggingClient.Debug("Filtered device addition finished")
		}
	}
}

func checkAllowList(d sdkModels.DiscoveredDevice, pw models.ProvisionWatcher, lc logger.LoggingClient) bool {
	// ignore the device protocol properties name
	for _, protocol := range d.Protocols {
		matchedCount := 0
		for name, regex := range pw.Identifiers {
			if value, ok := protocol[name]; ok {
				matched, err := regexp.MatchString(regex, value)
				if !matched || err != nil {
					lc.Debugf("Device %s's %s value %s did not match PW identifier: %s", d.Name, name, value, regex)
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

func checkBlockList(d sdkModels.DiscoveredDevice, pw models.ProvisionWatcher, lc logger.LoggingClient) bool {
	// a candidate should match none of the blocking identifiers
	for name, blacklist := range pw.BlockingIdentifiers {
		// ignore the device protocol properties name
		for _, protocol := range d.Protocols {
			if value, ok := protocol[name]; ok {
				for _, v := range blacklist {
					if value == v {
						lc.Debugf("Discovered Device %s's %s should not be %s", d.Name, name, value)
						return false
					}
				}
			}
		}
	}
	return true
}
