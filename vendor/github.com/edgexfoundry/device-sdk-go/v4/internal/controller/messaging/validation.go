//
// Copyright (C) 2023-2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package messaging

import (
	"context"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-messaging/v4/pkg/types"

	"github.com/edgexfoundry/device-sdk-go/v4/internal/container"
)

func SubscribeDeviceValidation(ctx context.Context, dic *di.Container) errors.EdgeX {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	configuration := container.ConfigurationFrom(dic.Get)
	messageBusInfo := container.ConfigurationFrom(dic.Get).MessageBus
	serviceName := container.DeviceServiceFrom(dic.Get).Name

	requestTopic := common.NewPathBuilder().EnableNameFieldEscape(configuration.Service.EnableNameFieldEscape).
		SetPath(messageBusInfo.GetBaseTopicPrefix()).SetNameFieldPath(serviceName).SetPath(common.ValidateDeviceSubscribeTopic).BuildPath()
	lc.Infof("Subscribing to device validation requests on topic: %s", requestTopic)

	responseTopicPrefix := common.NewPathBuilder().EnableNameFieldEscape(configuration.Service.EnableNameFieldEscape).
		SetPath(messageBusInfo.GetBaseTopicPrefix()).SetPath(common.ResponseTopic).SetNameFieldPath(serviceName).BuildPath()
	lc.Infof("Responses to device validation requests will be published on topic: %s/<requestId>", responseTopicPrefix)

	messages := make(chan types.MessageEnvelope, 1)
	messageErrors := make(chan error, 1)
	topics := []types.TopicChannel{
		{
			Topic:    requestTopic,
			Messages: messages,
		},
	}

	messageBus := bootstrapContainer.MessagingClientFrom(dic.Get)
	err := messageBus.Subscribe(topics, messageErrors)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				lc.Infof("Exiting waiting for MessageBus '%s' topic messages", requestTopic)
				return
			case err = <-messageErrors:
				lc.Error(err.Error())
			case msgEnvelope := <-messages:
				lc.Debugf("Device validation request received on message queue. Topic: %s, Correlation-id: %s", msgEnvelope.ReceivedTopic, msgEnvelope.CorrelationID)

				responseTopic := common.BuildTopic(responseTopicPrefix, msgEnvelope.RequestID)

				driver := container.ProtocolDriverFrom(dic.Get)

				var deviceRequest requests.AddDeviceRequest
				deviceRequest, err = types.GetMsgPayload[requests.AddDeviceRequest](msgEnvelope)
				if err != nil {
					lc.Errorf("Failed to JSON decoding AddDeviceRequest: %s", err.Error())
					res := types.NewMessageEnvelopeWithError(msgEnvelope.RequestID, err.Error())
					err = messageBus.Publish(res, responseTopic)
					if err != nil {
						lc.Errorf("Failed to publish device validation error response: %s", err.Error())
					}
					continue
				}

				err = driver.ValidateDevice(dtos.ToDeviceModel(deviceRequest.Device))
				if err != nil {
					lc.Errorf("Device validation failed: %s", err.Error())
					res := types.NewMessageEnvelopeWithError(msgEnvelope.RequestID, err.Error())
					err = messageBus.Publish(res, responseTopic)
					if err != nil {
						lc.Errorf("Failed to publish device validation error response: %s", err.Error())
					}
					continue
				}

				res, err := types.NewMessageEnvelopeForResponse(nil, msgEnvelope.RequestID, msgEnvelope.CorrelationID, common.ContentTypeJSON)
				if err != nil {
					lc.Errorf("Failed to create device validation response envelope: %s", err.Error())
					continue
				}

				err = messageBus.Publish(res, responseTopic)
				if err != nil {
					lc.Errorf("Failed to publish device validation response: %s", err.Error())
					continue
				}

				lc.Debugf("Device validation response published on message queue. Topic: %s, Correlation-id: %s", responseTopic, msgEnvelope.CorrelationID)
			}
		}
	}()

	return nil
}
