//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package transformer

import (
	"fmt"
	"sync"
	"time"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"

	"github.com/edgexfoundry/device-sdk-go/v2/internal/cache"
	"github.com/edgexfoundry/device-sdk-go/v2/internal/container"
	"github.com/edgexfoundry/device-sdk-go/v2/pkg/models"
)

var (
	previousOrigin int64
	originMutex    sync.Mutex
)

func CommandValuesToEventDTO(cvs []*models.CommandValue, deviceName string, sourceName string, dic *di.Container) (*dtos.Event, errors.EdgeX) {
	// in some case device service driver implementation would generate no readings
	// in this case no event would be created. Based on the implementation there would be 2 scenarios:
	// 1. uninitialized *CommandValue slices, i.e. nil
	// 2. initialized *CommandValue slice with no value in it.
	if cvs == nil {
		return nil, nil
	}

	device, exist := cache.Devices().ForName(deviceName)
	if !exist {
		errMsg := fmt.Sprintf("failed to find device %s", deviceName)
		return nil, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, errMsg, nil)
	}

	var transformsOK = true
	origin := getUniqueOrigin()
	tags := make(map[string]interface{})
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	config := container.ConfigurationFrom(dic.Get)
	readings := make([]dtos.BaseReading, 0, config.Device.MaxCmdOps)
	for _, cv := range cvs {
		if cv == nil {
			continue
		}
		// double check the CommandValue return from ProtocolDriver match device command
		dr, ok := cache.Profiles().DeviceResource(device.ProfileName, cv.DeviceResourceName)
		if !ok {
			msg := fmt.Sprintf("failed to find DeviceResource %s in Device %s for CommandValue (%s)", cv.DeviceResourceName, deviceName, cv.String())
			lc.Error(msg)
			return nil, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, msg, nil)
		}

		// perform data transformation
		if config.Device.DataTransform {
			edgexErr := TransformReadResult(cv, dr.Properties)
			if edgexErr != nil {
				lc.Errorf("failed to transform CommandValue (%s): %v", cv.String(), edgexErr)

				var err error
				if errors.Kind(edgexErr) == errors.KindOverflowError {
					cv, err = models.NewCommandValue(cv.DeviceResourceName, common.ValueTypeString, Overflow)
					if err != nil {
						return nil, errors.NewCommonEdgeXWrapper(err)
					}
				} else if errors.Kind(edgexErr) == errors.KindNaNError {
					cv, err = models.NewCommandValue(cv.DeviceResourceName, common.ValueTypeString, NaN)
					if err != nil {
						return nil, errors.NewCommonEdgeXWrapper(err)
					}
				} else {
					transformsOK = false
				}
			}
		}

		// assertion
		dc := bootstrapContainer.MetadataDeviceClientFrom(dic.Get)
		err := checkAssertion(cv, dr.Properties.Assertion, device.Name, lc, dc)
		if err != nil {
			return nil, errors.NewCommonEdgeXWrapper(err)
		}

		// ResourceOperation mapping
		ro, err := cache.Profiles().ResourceOperation(device.ProfileName, cv.DeviceResourceName)
		if err != nil {
			// this allows SDK to directly read deviceResource without deviceCommands defined.
			lc.Debugf("failed to read ResourceOperation: %v", err)
		} else if len(ro.Mappings) > 0 {
			newCV, ok := mapCommandValue(cv, ro.Mappings)
			if ok {
				cv = newCV
			}
		}

		for key, value := range cv.Tags {
			tags[key] = value
		}

		reading, err := commandValueToReading(cv, device.Name, device.ProfileName, dr.Properties.MediaType, origin)
		if err != nil {
			return nil, errors.NewCommonEdgeXWrapper(err)
		}
		readings = append(readings, reading)

		if cv.Type == common.ValueTypeBinary {
			lc.Debugf("device: %s DeviceResource: %v reading: binary value", device.Name, cv.DeviceResourceName)
		} else {
			lc.Debugf("device: %s DeviceResource: %v reading: %+v", device.Name, cv.DeviceResourceName, reading)
		}
	}

	if !transformsOK {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to transform value for %s", deviceName), nil)
	}

	if len(readings) > 0 {
		eventDTO := dtos.NewEvent(device.ProfileName, device.Name, sourceName)
		eventDTO.Readings = readings
		eventDTO.Origin = origin
		eventDTO.Tags = tags

		return &eventDTO, nil
	} else {
		return nil, nil
	}
}

func commandValueToReading(cv *models.CommandValue, deviceName, profileName, mediaType string, eventOrigin int64) (dtos.BaseReading, errors.EdgeX) {
	var err error
	var reading dtos.BaseReading

	if cv.Type == common.ValueTypeBinary {
		var binary []byte
		binary, err = cv.BinaryValue()
		if err != nil {
			return reading, errors.NewCommonEdgeXWrapper(err)
		}
		reading = dtos.NewBinaryReading(profileName, deviceName, cv.DeviceResourceName, binary, mediaType)
	} else if cv.Type == common.ValueTypeObject {
		reading = dtos.NewObjectReading(profileName, deviceName, cv.DeviceResourceName, cv.Value)
	} else {
		reading, err = dtos.NewSimpleReading(profileName, deviceName, cv.DeviceResourceName, cv.Type, cv.Value)
		if err != nil {
			return reading, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to transform CommandValue (%s) to Reading", cv.String()), err)
		}
	}
	// use the Origin if it was already set by ProtocolDriver implementation
	// otherwise use the same Origin of the upstream Event
	if cv.Origin != 0 {
		reading.Origin = cv.Origin
	} else {
		reading.Origin = eventOrigin
	}

	return reading, nil
}

func getUniqueOrigin() int64 {
	originMutex.Lock()
	defer originMutex.Unlock()
	now := time.Now().UnixNano()
	if now <= previousOrigin {
		now = previousOrigin + 1
	}
	previousOrigin = now
	return now
}
