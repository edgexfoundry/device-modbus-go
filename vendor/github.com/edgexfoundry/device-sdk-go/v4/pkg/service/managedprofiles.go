// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018-2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"context"
	"fmt"

	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
	"github.com/google/uuid"

	"github.com/edgexfoundry/device-sdk-go/v4/internal/cache"
)

// AddDeviceProfile adds a new DeviceProfile to the Device Service and Core Metadata
// Returns new DeviceProfile id or non-nil error.
func (s *deviceService) AddDeviceProfile(profile models.DeviceProfile) (string, error) {
	if p, ok := cache.Profiles().ForName(profile.Name); ok {
		return p.Id, errors.NewCommonEdgeX(errors.KindDuplicateName, fmt.Sprintf("name conflicted, Profile %s exists", profile.Name), nil)
	}

	s.lc.Debugf("Adding managed Profile %s", profile.Name)
	req := requests.NewDeviceProfileRequest(dtos.FromDeviceProfileModelToDTO(profile))
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.NewString()) // nolint:staticcheck
	res, err := container.DeviceProfileClientFrom(s.dic.Get).Add(ctx, []requests.DeviceProfileRequest{req})
	if err != nil {
		s.lc.Errorf("failed to add Profile %s to Core Metadata: %v", profile.Name, err)
		return "", err
	}
	err = cache.Profiles().Add(profile)
	if err != nil {
		s.lc.Errorf("failed to add Profile %s to cache", profile.Name)
		return "", err
	}

	return res[0].Id, nil
}

// DeviceProfiles return all managed DeviceProfiles from cache
func (s *deviceService) DeviceProfiles() []models.DeviceProfile {
	return cache.Profiles().All()
}

// GetProfileByName returns the Profile by its name if it exists in the cache, or returns an error.
func (s *deviceService) GetProfileByName(name string) (models.DeviceProfile, error) {
	profile, ok := cache.Profiles().ForName(name)
	if !ok {
		msg := fmt.Sprintf("failed to find Profile %s in cache", name)
		s.lc.Error(msg)
		return models.DeviceProfile{}, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, msg, nil)
	}
	return profile, nil
}

// RemoveDeviceProfileByName removes the specified DeviceProfile by name from the cache and ensures that the
// instance in Core Metadata is also removed.
func (s *deviceService) RemoveDeviceProfileByName(name string) error {
	profile, ok := cache.Profiles().ForName(name)
	if !ok {
		msg := fmt.Sprintf("failed to find Profile %s in cache", name)
		s.lc.Error(msg)
		return errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, msg, nil)
	}

	s.lc.Debugf("Removing managed Profile %s", profile.Name)
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.NewString()) // nolint:staticcheck
	_, err := container.DeviceProfileClientFrom(s.dic.Get).DeleteByName(ctx, name)
	if err != nil {
		s.lc.Errorf("failed to delete Profile %s in Core Metadata", name)
		return err
	}

	err = cache.Profiles().RemoveByName(name)
	return err
}

// UpdateDeviceProfile updates the DeviceProfile in the cache and ensures that the
// copy in Core Metadata is also updated.
func (s *deviceService) UpdateDeviceProfile(profile models.DeviceProfile) error {
	_, ok := cache.Profiles().ForName(profile.Name)
	if !ok {
		msg := fmt.Sprintf("failed to find Profile %s in cache", profile.Name)
		s.lc.Error(msg)
		return errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, msg, nil)
	}

	s.lc.Debugf("Updating managed Profile %s", profile.Name)
	req := requests.NewDeviceProfileRequest(dtos.FromDeviceProfileModelToDTO(profile))
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.NewString()) // nolint:staticcheck
	_, err := container.DeviceProfileClientFrom(s.dic.Get).Update(ctx, []requests.DeviceProfileRequest{req})
	if err != nil {
		s.lc.Errorf("failed to update Profile %s in Core Metadata: %v", profile.Name, err)
		return err
	}

	return err
}

// DeviceCommand retrieves the specific DeviceCommand instance from cache according to
// the Device name and Command name
func (s *deviceService) DeviceCommand(deviceName string, commandName string) (models.DeviceCommand, bool) {
	device, ok := cache.Devices().ForName(deviceName)
	if !ok {
		s.lc.Errorf("failed to find device %s in cache", deviceName)
		return models.DeviceCommand{}, false
	}

	dc, ok := cache.Profiles().DeviceCommand(device.ProfileName, commandName)
	if !ok {
		return dc, false
	}
	return dc, true
}

// DeviceResource retrieves the specific DeviceResource instance from cache according to
// the Device name and Device Resource name
func (s *deviceService) DeviceResource(deviceName string, deviceResource string) (models.DeviceResource, bool) {
	device, ok := cache.Devices().ForName(deviceName)
	if !ok {
		s.lc.Errorf("failed to find device %s in cache", deviceName)
		return models.DeviceResource{}, false
	}

	dr, ok := cache.Profiles().DeviceResource(device.ProfileName, deviceResource)
	if !ok {
		return dr, false
	}
	return dr, true
}
