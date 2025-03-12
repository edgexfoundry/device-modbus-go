//
// Copyright (C) 2023-2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package messaging

import (
	"context"
	"fmt"
	"strings"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-messaging/v4/pkg/types"

	"github.com/edgexfoundry/device-sdk-go/v4/internal/application"
	"github.com/edgexfoundry/device-sdk-go/v4/internal/container"
)

func MetadataSystemEventsCallback(ctx context.Context, serviceBaseName string, dic *di.Container) errors.EdgeX {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	configuration := container.ConfigurationFrom(dic.Get)
	messageBusInfo := container.ConfigurationFrom(dic.Get).MessageBus
	deviceService := container.DeviceServiceFrom(dic.Get)
	metadataSystemEventTopic := common.NewPathBuilder().EnableNameFieldEscape(configuration.Service.EnableNameFieldEscape).
		SetPath(messageBusInfo.GetBaseTopicPrefix()).SetPath(common.MetadataSystemEventSubscribeTopic).SetNameFieldPath(deviceService.Name).SetPath("#").BuildPath()
	profileDeleteSystemEventTopic := common.NewPathBuilder().EnableNameFieldEscape(configuration.Service.EnableNameFieldEscape).
		SetPath(messageBusInfo.GetBaseTopicPrefix()).
		SetPath(strings.Replace(common.MetadataSystemEventSubscribeTopic, "+/+", common.DeviceProfileSystemEventType+"/"+common.SystemEventActionDelete, 1)).
		SetPath("#").BuildPath()
	lc.Infof("Subscribing to System Events on topics: %s and %s", metadataSystemEventTopic, profileDeleteSystemEventTopic)

	messages := make(chan types.MessageEnvelope, 1)
	messageErrors := make(chan error, 1)
	topics := []types.TopicChannel{
		{
			Topic:    metadataSystemEventTopic,
			Messages: messages,
		},
		{
			Topic:    profileDeleteSystemEventTopic,
			Messages: messages,
		},
	}

	// Must subscribe to provision watcher System Events separately when the service has an instance name. i.e. -i flag was used.
	// This is because Provision Watchers apply to all instances of the device service, i.e. the service base name (without the instance portion).
	// The above subscribe topic is for the specific instance name which is appropriate for Devices, Profiles and Devices service System Events
	// and when the -i flag is not used.
	if serviceBaseName != deviceService.Name {
		// Must replace the first wildcard with the type for Provision Watchers
		baseSubscribeTopic := strings.Replace(common.MetadataSystemEventSubscribeTopic, "+", common.ProvisionWatcherSystemEventType, 1)
		provisionWatcherSystemEventSubscribeTopic := common.NewPathBuilder().EnableNameFieldEscape(configuration.Service.EnableNameFieldEscape).
			SetPath(messageBusInfo.GetBaseTopicPrefix()).SetPath(baseSubscribeTopic).SetNameFieldPath(serviceBaseName).SetPath("#").BuildPath()

		topics = append(topics, types.TopicChannel{
			Topic:    provisionWatcherSystemEventSubscribeTopic,
			Messages: messages,
		})

		lc.Infof("Subscribing additionally to Provision Watcher System Events on topic: %s", provisionWatcherSystemEventSubscribeTopic)
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
				lc.Infof("Exiting waiting for MessageBus '%s' topic messages", metadataSystemEventTopic)
				return
			case err = <-messageErrors:
				lc.Error(err.Error())
			case msgEnvelope := <-messages:
				lc.Debugf("System event received on message queue. Topic: %s, Correlation-id: %s", msgEnvelope.ReceivedTopic, msgEnvelope.CorrelationID)

				var systemEvent dtos.SystemEvent
				systemEvent, err := types.GetMsgPayload[dtos.SystemEvent](msgEnvelope)
				if err != nil {
					lc.Errorf("failed to JSON decoding system event: %s", err.Error())
					continue
				}

				serviceName := container.DeviceServiceFrom(dic.Get).Name
				if systemEvent.Owner == common.CoreMetaDataServiceKey {
					if systemEvent.Type != common.DeviceProfileSystemEventType && systemEvent.Action != common.SystemEventActionDelete {
						lc.Errorf("only support device profile delete system events from owner %s", systemEvent.Owner)
					}
				} else if systemEvent.Owner != serviceName && systemEvent.Owner != serviceBaseName {
					lc.Errorf("unmatched system event owner %s with service name %s", systemEvent.Owner, serviceName)
					continue
				}

				switch systemEvent.Type {
				case common.DeviceSystemEventType:
					err = deviceSystemEventAction(systemEvent, dic)
					if err != nil {
						lc.Error(err.Error(), common.CorrelationHeader, msgEnvelope.CorrelationID)
					}
				case common.DeviceProfileSystemEventType:
					err = deviceProfileSystemEventAction(systemEvent, dic)
					if err != nil {
						lc.Error(err.Error(), common.CorrelationHeader, msgEnvelope.CorrelationID)
					}
				case common.ProvisionWatcherSystemEventType:
					err = provisionWatcherSystemEventAction(systemEvent, dic)
					if err != nil {
						lc.Error(err.Error(), common.CorrelationHeader, msgEnvelope.CorrelationID)
					}
				case common.DeviceServiceSystemEventType:
					err = deviceServiceSystemEventAction(systemEvent, dic)
					if err != nil {
						lc.Error(err.Error(), common.CorrelationHeader, msgEnvelope.CorrelationID)
					}
				default:
					lc.Errorf("unknown system event type %s", systemEvent.Type)
					continue
				}
			}
		}
	}()

	return nil
}

func deviceSystemEventAction(systemEvent dtos.SystemEvent, dic *di.Container) error {
	var device dtos.Device
	err := systemEvent.DecodeDetails(&device)
	if err != nil {
		return fmt.Errorf("failed to decode %s system event details: %s", systemEvent.Type, err.Error())
	}

	switch systemEvent.Action {
	case common.SystemEventActionAdd:
		err = application.AddDevice(requests.NewAddDeviceRequest(device), dic)
	case common.SystemEventActionUpdate:
		deviceModel := dtos.ToDeviceModel(device)
		updateDeviceDTO := dtos.FromDeviceModelToUpdateDTO(deviceModel)
		err = application.UpdateDevice(requests.NewUpdateDeviceRequest(updateDeviceDTO), dic)
	case common.SystemEventActionDelete:
		err = application.DeleteDevice(device.Name, dic)
	default:
		return fmt.Errorf("unknown %s system event action %s", systemEvent.Type, systemEvent.Action)
	}

	return err
}

func deviceProfileSystemEventAction(systemEvent dtos.SystemEvent, dic *di.Container) error {
	var deviceProfile dtos.DeviceProfile
	err := systemEvent.DecodeDetails(&deviceProfile)
	if err != nil {
		return fmt.Errorf("failed to decode %s system event details: %s", systemEvent.Type, err.Error())
	}

	switch systemEvent.Action {
	case common.SystemEventActionUpdate:
		err = application.UpdateProfile(requests.NewDeviceProfileRequest(deviceProfile), dic)
	case common.SystemEventActionDelete:
		err = application.DeleteProfile(deviceProfile.Name, dic)
	// there is no action needed for Device Profile Add in Device Service
	case common.SystemEventActionAdd:
		break
	default:
		return fmt.Errorf("unknown %s system event action %s", systemEvent.Type, systemEvent.Action)
	}

	return err
}

func provisionWatcherSystemEventAction(systemEvent dtos.SystemEvent, dic *di.Container) error {
	var pw dtos.ProvisionWatcher
	err := systemEvent.DecodeDetails(&pw)
	if err != nil {
		return fmt.Errorf("failed to decode %s system event details: %s", systemEvent.Type, err.Error())
	}

	switch systemEvent.Action {
	case common.SystemEventActionAdd:
		err = application.AddProvisionWatcher(requests.NewAddProvisionWatcherRequest(pw), dic)
	case common.SystemEventActionUpdate:
		pwModel := dtos.ToProvisionWatcherModel(pw)
		pwDTO := dtos.FromProvisionWatcherModelToUpdateDTO(pwModel)
		err = application.UpdateProvisionWatcher(requests.NewUpdateProvisionWatcherRequest(pwDTO), dic)
	case common.SystemEventActionDelete:
		err = application.DeleteProvisionWatcher(pw.Name, dic)
	default:
		return fmt.Errorf("unknown %s system event action %s", systemEvent.Type, systemEvent.Action)
	}

	return err
}

func deviceServiceSystemEventAction(systemEvent dtos.SystemEvent, dic *di.Container) error {
	var deviceService dtos.DeviceService
	err := systemEvent.DecodeDetails(&deviceService)
	if err != nil {
		return fmt.Errorf("failed to decode %s system event details: %s", systemEvent.Type, err.Error())
	}

	switch systemEvent.Action {
	case common.SystemEventActionUpdate:
		deviceServiceModel := dtos.ToDeviceServiceModel(deviceService)
		updateDeviceServiceDTO := dtos.FromDeviceServiceModelToUpdateDTO(deviceServiceModel)
		err = application.UpdateDeviceService(requests.NewUpdateDeviceServiceRequest(updateDeviceServiceDTO), dic)
	// there is no action needed for Device Service Add and Delete in Device Service
	case common.SystemEventActionAdd, common.SystemEventActionDelete:
		break
	default:
		return fmt.Errorf("unknown %s system event action %s", systemEvent.Type, systemEvent.Action)
	}

	return err
}
