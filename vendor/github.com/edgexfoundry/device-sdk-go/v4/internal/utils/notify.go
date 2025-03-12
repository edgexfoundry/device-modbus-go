//
// Copyright (C) 2024-2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"context"

	"github.com/edgexfoundry/device-sdk-go/v4/internal/container"
	sdkModels "github.com/edgexfoundry/device-sdk-go/v4/pkg/models"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	"github.com/edgexfoundry/go-mod-messaging/v4/pkg/types"
)

func PublishDeviceDiscoveryProgressSystemEvent(id string, progress, discoveredDeviceCount int, message string, ctx context.Context, dic *di.Container) {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	lc.Debugf("Publishing device discovery progress system event. Correlation Id: %s", id)
	details := sdkModels.DeviceDiscoveryProgress{Progress: sdkModels.Progress{RequestId: id, Progress: progress, Message: message}, DiscoveredDeviceCount: discoveredDeviceCount}
	PublishGenericSystemEvent(common.DeviceSystemEventType, common.SystemEventActionDiscovery, details, ctx, dic)
}

func PublishProfileScanProgressSystemEvent(id string, progress int, message string, ctx context.Context, dic *di.Container) {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	lc.Debugf("Publishing device profile scan progress system event. Correlation Id: %s", id)
	details := sdkModels.Progress{RequestId: id, Progress: progress, Message: message}
	PublishGenericSystemEvent(common.DeviceSystemEventType, common.SystemEventActionProfileScan, details, ctx, dic)
}

func PublishGenericSystemEvent(eventType, action string, details any, ctx context.Context, dic *di.Container) {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	config := container.ConfigurationFrom(dic.Get)
	serviceName := container.DeviceServiceFrom(dic.Get).Name
	messagingClient := bootstrapContainer.MessagingClientFrom(dic.Get)
	if messagingClient == nil {
		lc.Errorf("unable to publish '%s' '%s' System Event: MessageBus Client not available. Please update MessageBus configuration to enable sending System Events via the EdgeX MessageBus", eventType, action)
		return
	}

	systemEvent := dtos.NewSystemEvent(eventType, action, serviceName, serviceName, nil, details)
	topicPathBuilder := common.NewPathBuilder().EnableNameFieldEscape(config.Service.EnableNameFieldEscape)
	publishTopic := topicPathBuilder.SetPath(config.MessageBus.GetBaseTopicPrefix()).SetPath(common.SystemEventPublishTopic).
		SetPath(systemEvent.Source).SetPath(systemEvent.Type).SetPath(systemEvent.Action).SetNameFieldPath(systemEvent.Owner).BuildPath()
	envelope := types.NewMessageEnvelope(systemEvent, ctx)
	// Correlation ID and Content type are set by the above factory function from the context of the request that
	// triggered this System Event. We'll keep that Correlation ID, but need to make sure the Content Type is set appropriate
	// for how the payload was encoded above.
	envelope.ContentType = common.ContentTypeJSON
	if err := messagingClient.Publish(envelope, publishTopic); err != nil {
		lc.Errorf("unable to publish '%s' '%s' System Event for owner: %s and source: %s to topic '%s': %v", eventType, action, serviceName, serviceName, publishTopic, err)
		return
	}

	lc.Debugf("Published the '%s' '%s' System Event for owner: %s and source: %s to topic '%s'", eventType, action, serviceName, serviceName, publishTopic)
}
