//
// Copyright (C) 2022-2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package messaging

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-messaging/v4/pkg/types"

	"github.com/edgexfoundry/device-sdk-go/v4/internal/application"
	sdkCommon "github.com/edgexfoundry/device-sdk-go/v4/internal/common"
	"github.com/edgexfoundry/device-sdk-go/v4/internal/container"
)

func SubscribeCommands(ctx context.Context, dic *di.Container) errors.EdgeX {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	configuration := container.ConfigurationFrom(dic.Get)
	messageBusInfo := container.ConfigurationFrom(dic.Get).MessageBus
	deviceService := container.DeviceServiceFrom(dic.Get)

	requestSubscribeTopic := common.NewPathBuilder().EnableNameFieldEscape(configuration.Service.EnableNameFieldEscape).
		SetPath(messageBusInfo.GetBaseTopicPrefix()).SetPath(common.CommandRequestSubscribeTopic).SetNameFieldPath(deviceService.Name).SetPath("#").BuildPath()
	lc.Infof("Subscribing to command requests on topic: %s", requestSubscribeTopic)

	responsePublishTopicPrefix := common.NewPathBuilder().EnableNameFieldEscape(configuration.Service.EnableNameFieldEscape).
		SetPath(messageBusInfo.GetBaseTopicPrefix()).SetPath(common.ResponseTopic).SetNameFieldPath(deviceService.Name).BuildPath()
	lc.Infof("Responses to command requests will be published on topic: %s/<requestId>", responsePublishTopicPrefix)

	messages := make(chan types.MessageEnvelope, 1)
	messageErrors := make(chan error, 1)
	topics := []types.TopicChannel{
		{
			Topic:    requestSubscribeTopic,
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
				lc.Infof("Exiting waiting for MessageBus '%s' topic messages", requestSubscribeTopic)
				return
			case err = <-messageErrors:
				lc.Error(err.Error())
			case msgEnvelope := <-messages:
				lc.Debugf("Command request received on message queue. Topic: %s, Correlation-id: %s", msgEnvelope.ReceivedTopic, msgEnvelope.CorrelationID)

				// expected command request topic scheme: #/<service-name>/<device-name>/<command-name>/<method>
				topicLevels := strings.Split(msgEnvelope.ReceivedTopic, "/")
				length := len(topicLevels)
				if length < 4 {
					lc.Error("Failed to parse and construct command response topic scheme, expected request topic scheme: '#/<service-name>/<device-name>/<command-name>/<method>'")
					continue
				}

				// expected command response topic scheme: #/<service-name>/<device-name>/<command-name>/<method>
				deviceName, err := url.PathUnescape(topicLevels[length-3])
				if err != nil {
					lc.Errorf("Failed to unescape device name: %s", err.Error())
					continue
				}
				commandName, err := url.PathUnescape(topicLevels[length-2])
				if err != nil {
					lc.Errorf("Failed to unescape command name: %s", err.Error())
					continue
				}
				method := topicLevels[length-1]

				responsePublishTopic := common.BuildTopic(responsePublishTopicPrefix, msgEnvelope.RequestID)

				switch strings.ToUpper(method) {
				case "GET":
					getCommand(ctx, msgEnvelope, responsePublishTopic, deviceName, commandName, dic)
				case "SET":
					setCommand(ctx, msgEnvelope, responsePublishTopic, deviceName, commandName, dic)
				default:
					lc.Errorf("unknown command method '%s', only 'get' or 'set' is allowed", method)
					continue
				}

				lc.Debugf("Command response published on message queue. Topic: %s, Correlation-id: %s", responsePublishTopic, msgEnvelope.CorrelationID)
			}
		}
	}()

	return nil
}

func getCommand(ctx context.Context, msgEnvelope types.MessageEnvelope, responseTopic string, deviceName string, commandName string, dic *di.Container) {
	var responseEnvelope types.MessageEnvelope

	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	messageBus := bootstrapContainer.MessagingClientFrom(dic.Get)
	rawQuery, reserved := filterQueryParams(msgEnvelope.QueryParams)

	// TODO: fix properly in EdgeX 3.0
	ctx = context.WithValue(ctx, common.CorrelationHeader, msgEnvelope.CorrelationID) // nolint: staticcheck
	event, edgexErr := application.GetCommand(ctx, deviceName, commandName, rawQuery, reserved[common.RegexCommand], dic)
	if edgexErr != nil {
		lc.Errorf("Failed to process get device command %s for device %s: %s", commandName, deviceName, edgexErr.Error())
		responseEnvelope = types.NewMessageEnvelopeWithError(msgEnvelope.RequestID, edgexErr.Error())
		err := messageBus.Publish(responseEnvelope, responseTopic)
		if err != nil {
			lc.Errorf("Failed to publish command error response: %s", err.Error())
		}
		return
	}

	var err error
	var encoding string
	var eventResponse any
	if reserved[common.ReturnEvent] {
		resp := responses.NewEventResponse(msgEnvelope.RequestID, "", http.StatusOK, *event)
		encoding = resp.GetEncodingContentType()
		eventResponse = resp
	} else {
		encoding = common.ContentTypeJSON
	}

	responseEnvelope, err = types.NewMessageEnvelopeForResponse(eventResponse, msgEnvelope.RequestID, msgEnvelope.CorrelationID, encoding)
	if err != nil {
		lc.Errorf("Failed to create response message envelope: %s", err.Error())
		responseEnvelope = types.NewMessageEnvelopeWithError(msgEnvelope.RequestID, err.Error())
		err = messageBus.Publish(responseEnvelope, responseTopic)
		if err != nil {
			lc.Errorf("Failed to publish command error response: %s", err.Error())
		}
		return
	}

	configuration := container.ConfigurationFrom(dic.Get)
	err = messageBus.PublishWithSizeLimit(responseEnvelope, responseTopic, configuration.MaxEventSize)
	if err != nil {
		lc.Errorf("Failed to publish command response: %s", err.Error())
		responseEnvelope = types.NewMessageEnvelopeWithError(msgEnvelope.RequestID, err.Error())
		err = messageBus.Publish(responseEnvelope, responseTopic)
		if err != nil {
			lc.Errorf("Failed to publish command error response: %s", err.Error())
		}
		return
	}

	if reserved[common.PushEvent] {
		go sdkCommon.SendEvent(event, msgEnvelope.CorrelationID, dic)
	}

}

func setCommand(ctx context.Context, msgEnvelope types.MessageEnvelope, responseTopic string, deviceName string, commandName string, dic *di.Container) {
	var responseEnvelope types.MessageEnvelope

	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	messageBus := bootstrapContainer.MessagingClientFrom(dic.Get)
	rawQuery, _ := filterQueryParams(msgEnvelope.QueryParams)
	requestPayload, err := types.GetMsgPayload[map[string]any](msgEnvelope)
	if err != nil {
		lc.Errorf("Failed to decode set command request payload: %s", err.Error())
		responseEnvelope = types.NewMessageEnvelopeWithError(msgEnvelope.RequestID, err.Error())
		err = messageBus.Publish(responseEnvelope, responseTopic)
		if err != nil {
			lc.Errorf("Failed to publish command response: %s", err.Error())
		}
		return
	}

	// TODO: fix properly in EdgeX 3.0
	ctx = context.WithValue(ctx, common.CorrelationHeader, msgEnvelope.CorrelationID) // nolint: staticcheck
	event, edgexErr := application.SetCommand(ctx, deviceName, commandName, rawQuery, requestPayload, dic)
	if edgexErr != nil {
		lc.Errorf("Failed to process set device command %s for device %s: %s", commandName, deviceName, edgexErr.Error())
		responseEnvelope = types.NewMessageEnvelopeWithError(msgEnvelope.RequestID, edgexErr.Error())
		err = messageBus.Publish(responseEnvelope, responseTopic)
		if err != nil {
			lc.Errorf("Failed to publish command response: %s", err.Error())
		}
		return
	}

	responseEnvelope, err = types.NewMessageEnvelopeForResponse(nil, msgEnvelope.RequestID, msgEnvelope.CorrelationID, common.ContentTypeJSON)
	if err != nil {
		lc.Errorf("Failed to create response message envelope: %s", err.Error())
		responseEnvelope = types.NewMessageEnvelopeWithError(msgEnvelope.RequestID, err.Error())
		err = messageBus.Publish(responseEnvelope, responseTopic)
		if err != nil {
			lc.Errorf("Failed to publish command response: %s", err.Error())
		}
		return
	}

	err = messageBus.Publish(responseEnvelope, responseTopic)
	if err != nil {
		lc.Errorf("Failed to publish command response: %s", err.Error())
		return
	}

	if event != nil {
		go sdkCommon.SendEvent(event, msgEnvelope.CorrelationID, dic)
	}
}

func filterQueryParams(queries map[string]string) (string, map[string]bool) {
	var rawQuery []string
	reserved := make(map[string]bool)
	reserved[common.PushEvent] = false
	reserved[common.ReturnEvent] = true
	reserved[common.RegexCommand] = true

	for k, v := range queries {
		if k == common.PushEvent && v == common.ValueTrue {
			reserved[common.PushEvent] = true
			continue
		}
		if k == common.ReturnEvent && v == common.ValueFalse {
			reserved[common.ReturnEvent] = false
			continue
		}
		if k == common.RegexCommand && v == common.ValueFalse {
			reserved[common.RegexCommand] = false
		}
		if strings.HasPrefix(k, sdkCommon.SDKReservedPrefix) {
			continue
		}
		rawQuery = append(rawQuery, fmt.Sprintf("%s=%s", k, v))
	}

	return strings.Join(rawQuery, "&"), reserved
}
