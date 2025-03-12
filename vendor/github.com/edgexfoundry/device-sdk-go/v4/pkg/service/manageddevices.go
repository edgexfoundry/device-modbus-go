// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018-2023 IOTech Ltd
// Copyright (C) 2023 Intel
//
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"context"
	"fmt"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
	"github.com/google/uuid"

	"github.com/edgexfoundry/device-sdk-go/v4/internal/cache"
)

// AddDevice adds a new Device to the Device Service and Core Metadata
// Returns new Device id or non-nil error.
func (s *deviceService) AddDevice(device models.Device) (string, error) {
	if d, ok := cache.Devices().ForName(device.Name); ok {
		return d.Id, errors.NewCommonEdgeX(errors.KindDuplicateName, fmt.Sprintf("name conflicted, Device %s exists", device.Name), nil)
	}

	device.ServiceName = s.serviceKey

	s.lc.Debugf("Adding managed Device %s", device.Name)
	req := requests.NewAddDeviceRequest(dtos.FromDeviceModelToDTO(device))
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.NewString()) // nolint:staticcheck
	res, err := container.DeviceClientFrom(s.dic.Get).Add(ctx, []requests.AddDeviceRequest{req})
	if err != nil {
		s.lc.Errorf("failed to add Device %s to Core Metadata: %v", device.Name, err)
		return "", err
	}

	return res[0].Id, nil
}

// Devices return all managed Devices from cache
func (s *deviceService) Devices() []models.Device {
	return cache.Devices().All()
}

// GetDeviceByName returns the Device by its name if it exists in the cache, or returns an error.
func (s *deviceService) GetDeviceByName(name string) (models.Device, error) {
	device, ok := cache.Devices().ForName(name)
	if !ok {
		msg := fmt.Sprintf("failed to find Device %s in cache", name)
		s.lc.Error(msg)
		return models.Device{}, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, msg, nil)
	}
	return device, nil
}

// DeviceExistsForName returns true if a device exists in cache with the specified name, otherwise it returns false.
func (s *deviceService) DeviceExistsForName(name string) bool {
	_, ok := cache.Devices().ForName(name)
	return ok
}

// RemoveDeviceByName removes the specified Device by name from the cache and ensures that the
// instance in Core Metadata is also removed.
func (s *deviceService) RemoveDeviceByName(name string) error {
	if _, err := s.GetDeviceByName(name); err != nil {
		return err
	}

	s.lc.Debugf("Removing managed Device %s", name)
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.NewString()) // nolint:staticcheck
	_, err := container.DeviceClientFrom(s.dic.Get).DeleteDeviceByName(ctx, name)
	if err != nil {
		s.lc.Errorf("failed to delete Device %s in Core Metadata", name)
	}

	return err
}

// UpdateDevice updates the Device in Core Metadata
func (s *deviceService) UpdateDevice(device models.Device) error {
	return s.PatchDevice(dtos.FromDeviceModelToUpdateDTO(device))
}

// UpdateDeviceOperatingState updates the OperatingState for the Device with given name
// in Core Metadata
func (s *deviceService) UpdateDeviceOperatingState(name string, state models.OperatingState) error {
	stateString := string(state)
	return s.PatchDevice(dtos.UpdateDevice{
		Name:           &name,
		OperatingState: &stateString,
	})
}

// PatchDevice patches the specified device properties in Core Metadata. Device name is required
// to be provided in the UpdateDevice. Note that all properties of UpdateDevice are pointers
// and anything that is nil will not modify the device. In the case of Arrays and Maps, the whole new value
// must be sent, as it is applied as an overwrite operation.
func (s *deviceService) PatchDevice(updateDevice dtos.UpdateDevice) error {
	if updateDevice.Name == nil {
		msg := "missing device name for patch device call"
		s.lc.Error(msg)
		return errors.NewCommonEdgeX(errors.KindContractInvalid, msg, nil)
	}

	if _, err := s.GetDeviceByName(*updateDevice.Name); err != nil {
		return err
	}

	s.lc.Debugf("Patching managed Device %s", *updateDevice.Name)
	req := requests.UpdateDeviceRequest{
		BaseRequest: commonDTO.NewBaseRequest(),
		Device:      updateDevice,
	}
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.NewString()) // nolint:staticcheck
	_, err := container.DeviceClientFrom(s.dic.Get).Update(ctx, []requests.UpdateDeviceRequest{req})
	if err != nil {
		s.lc.Errorf("failed to update Device %s in Core Metadata: %v", *updateDevice.Name, err)
	}

	return err
}
