// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"context"
	"fmt"
	"time"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"
	"github.com/edgexfoundry/go-mod-messaging/v2/pkg/types"

	"github.com/edgexfoundry/device-sdk-go/v2/internal/container"
)

func UpdateLastConnected(name string, lc logger.LoggingClient, dc interfaces.DeviceClient) {
	t := time.Now().UnixNano() / int64(time.Millisecond)
	device := dtos.UpdateDevice{
		Name:          &name,
		LastConnected: &t,
	}

	req := requests.NewUpdateDeviceRequest(device)
	_, err := dc.Update(context.Background(), []requests.UpdateDeviceRequest{req})
	if err != nil {
		lc.Errorf("failed to update LastConnected for Device %s in Core Metadata: %v", name, err)
	}
}

func UpdateOperatingState(name string, state string, lc logger.LoggingClient, dc interfaces.DeviceClient) {
	device := dtos.UpdateDevice{
		Name:           &name,
		OperatingState: &state,
	}

	req := requests.NewUpdateDeviceRequest(device)
	_, err := dc.Update(context.Background(), []requests.UpdateDeviceRequest{req})
	if err != nil {
		lc.Errorf("failed to update OperatingState for Device %s in Core Metadata: %v", name, err)
	}
}

func SendEvent(event *dtos.Event, correlationID string, dic *di.Container) {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	configuration := container.ConfigurationFrom(dic.Get)
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, correlationID)
	req := requests.NewAddEventRequest(*event)

	if configuration.Device.UseMessageBus {
		mc := container.MessagingClientFrom(dic.Get)
		bytes, encoding, err := req.Encode()
		if err != nil {
			lc.Error(err.Error())
		}
		ctx = context.WithValue(ctx, common.ContentType, encoding)
		envelope := types.NewMessageEnvelope(bytes, ctx)
		publishTopic := fmt.Sprintf("%s/%s/%s/%s", configuration.MessageQueue.PublishTopicPrefix, event.ProfileName, event.DeviceName, event.SourceName)
		err = mc.Publish(envelope, publishTopic)
		if err != nil {
			lc.Errorf("Failed to publish event to MessageBus: %s", err)
		}
		lc.Debugf("Event(profileName: %s, deviceName: %s, sourceName: %s, id: %s) published to MessageBus", event.ProfileName, event.DeviceName, event.SourceName, event.Id)
	} else {
		ec := bootstrapContainer.DataEventClientFrom(dic.Get)
		_, err := ec.Add(ctx, req)
		if err != nil {
			lc.Errorf("Failed to push event to Coredata: %s", err)
		} else {
			lc.Debugf("Event(profileName: %s, deviceName: %s, sourceName: %s, id: %s) pushed to Coredata", event.ProfileName, event.DeviceName, event.SourceName, event.Id)
		}
	}
}
