// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"context"
	"fmt"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	commonDTO "github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
	"github.com/google/uuid"

	"github.com/edgexfoundry/device-sdk-go/v2/internal/cache"
)

// AddDevice adds a new Device to the Device Service and Core Metadata
// Returns new Device id or non-nil error.
func (s *DeviceService) AddDevice(device models.Device) (string, error) {
	if d, ok := cache.Devices().ForName(device.Name); ok {
		return d.Id, errors.NewCommonEdgeX(errors.KindDuplicateName, fmt.Sprintf("name conflicted, Device %s exists", device.Name), nil)
	}

	_, cacheExist := cache.Profiles().ForName(device.ProfileName)
	if !cacheExist {
		res, err := s.edgexClients.DeviceProfileClient.DeviceProfileByName(context.Background(), device.ProfileName)
		if err != nil {
			errMsg := fmt.Sprintf("failed to find Profile %s for Device %s", device.ProfileName, device.Name)
			s.LoggingClient.Error(errMsg)
			return "", err
		}
		err = cache.Profiles().Add(dtos.ToDeviceProfileModel(res.Profile))
		if err != nil {
			return "", errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to cache the profile %s", res.Profile.Name), err)
		}
	}
	device.ServiceName = s.ServiceName

	s.LoggingClient.Debugf("Adding managed Device %s", device.Name)
	req := requests.NewAddDeviceRequest(dtos.FromDeviceModelToDTO(device))
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.NewString())
	res, err := s.edgexClients.DeviceClient.Add(ctx, []requests.AddDeviceRequest{req})
	if err != nil {
		s.LoggingClient.Errorf("failed to add Device %s to Core Metadata: %v", device.Name, err)
		return "", err
	}

	return res[0].Id, nil
}

// Devices return all managed Devices from cache
func (s *DeviceService) Devices() []models.Device {
	return cache.Devices().All()
}

// GetDeviceByName returns the Device by its name if it exists in the cache, or returns an error.
func (s *DeviceService) GetDeviceByName(name string) (models.Device, error) {
	device, ok := cache.Devices().ForName(name)
	if !ok {
		msg := fmt.Sprintf("failed to find Device %s in cache", name)
		s.LoggingClient.Error(msg)
		return models.Device{}, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, msg, nil)
	}
	return device, nil
}

// RemoveDeviceByName removes the specified Device by name from the cache and ensures that the
// instance in Core Metadata is also removed.
func (s *DeviceService) RemoveDeviceByName(name string) error {
	device, ok := cache.Devices().ForName(name)
	if !ok {
		msg := fmt.Sprintf("failed to find device %s in cache", name)
		s.LoggingClient.Error(msg)
		return errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, msg, nil)
	}

	s.LoggingClient.Debugf("Removing managed Device %s", device.Name)
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.NewString())
	_, err := s.edgexClients.DeviceClient.DeleteDeviceByName(ctx, name)
	if err != nil {
		s.LoggingClient.Errorf("failed to delete Device %s in Core Metadata", name)
	}

	return err
}

// UpdateDevice updates the Device in the cache and ensures that the
// copy in Core Metadata is also updated.
func (s *DeviceService) UpdateDevice(device models.Device) error {
	_, ok := cache.Devices().ForName(device.Name)
	if !ok {
		msg := fmt.Sprintf("failed to find Device %s in cache", device.Name)
		s.LoggingClient.Error(msg)
		return errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, msg, nil)
	}

	s.LoggingClient.Debugf("Updating managed Device %s", device.Name)
	req := requests.NewUpdateDeviceRequest(dtos.FromDeviceModelToUpdateDTO(device))
	req.Device.Id = nil
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.NewString())
	_, err := s.edgexClients.DeviceClient.Update(ctx, []requests.UpdateDeviceRequest{req})
	if err != nil {
		s.LoggingClient.Errorf("failed to update Device %s in Core Metadata: %v", device.Name, err)
	}

	return err
}

// UpdateDeviceOperatingState updates the Device's OperatingState with given name
// in Core Metadata and device service cache.
func (s *DeviceService) UpdateDeviceOperatingState(deviceName string, state string) error {
	d, ok := cache.Devices().ForName(deviceName)
	if !ok {
		msg := fmt.Sprintf("failed to find Device %s in cache", deviceName)
		s.LoggingClient.Error(msg)
		return errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, msg, nil)
	}

	s.LoggingClient.Debugf("Updating managed Device OperatingState %s", d.Name)
	req := requests.UpdateDeviceRequest{
		BaseRequest: commonDTO.NewBaseRequest(),
		Device: dtos.UpdateDevice{
			Name:           &deviceName,
			OperatingState: &state,
		},
	}
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.NewString())
	_, err := s.edgexClients.DeviceClient.Update(ctx, []requests.UpdateDeviceRequest{req})
	if err != nil {
		s.LoggingClient.Errorf("failed to update Device %s OperatingState in Core Metadata: %v", d.Name, err)
	}

	return err
}
