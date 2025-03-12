// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020-2023 IOTech Ltd
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

// AddProvisionWatcher adds a new Watcher to the cache and Core Metadata
// Returns new Watcher id or non-nil error.
func (s *deviceService) AddProvisionWatcher(watcher models.ProvisionWatcher) (string, error) {
	if pw, ok := cache.ProvisionWatchers().ForName(watcher.Name); ok {
		return pw.Id,
			errors.NewCommonEdgeX(errors.KindDuplicateName, fmt.Sprintf("name conflicted, ProvisionWatcher %s exists", watcher.Name), nil)
	}

	// baseServiceName is the service name w/o the instance portion added when -i/--instance flag is used.
	// The ServiceName in the ProvisionWatcher must use the base name since ProvisionWatchers are used by all instances
	// of the device service.
	watcher.ServiceName = s.baseServiceName

	s.lc.Debugf("Adding managed ProvisionWatcher %s", watcher.Name)
	req := requests.NewAddProvisionWatcherRequest(dtos.FromProvisionWatcherModelToDTO(watcher))
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.NewString()) // nolint:staticcheck
	res, err := container.ProvisionWatcherClientFrom(s.dic.Get).Add(ctx, []requests.AddProvisionWatcherRequest{req})
	if err != nil {
		s.lc.Errorf("failed to add ProvisionWatcher to Core Metadata: %v", watcher.Name, err)
		return "", err
	}

	return res[0].Id, nil
}

// ProvisionWatchers return all managed Watchers from cache
func (s *deviceService) ProvisionWatchers() []models.ProvisionWatcher {
	return cache.ProvisionWatchers().All()
}

// GetProvisionWatcherByName returns the Watcher by its name if it exists in the cache, or returns an error.
func (s *deviceService) GetProvisionWatcherByName(name string) (models.ProvisionWatcher, error) {
	pw, ok := cache.ProvisionWatchers().ForName(name)
	if !ok {
		msg := fmt.Sprintf("failed to find ProvisionWatcher %s in cache", name)
		s.lc.Error(msg)
		return models.ProvisionWatcher{}, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, msg, nil)
	}
	return pw, nil
}

// RemoveProvisionWatcher removes the specified Watcher by name from the cache and ensures that the
// instance in Core Metadata is also removed.
func (s *deviceService) RemoveProvisionWatcher(name string) error {
	pw, ok := cache.ProvisionWatchers().ForName(name)
	if !ok {
		msg := fmt.Sprintf("failed to find ProvisionWatcher %s in cache", name)
		s.lc.Error(msg)
		return errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, msg, nil)
	}

	s.lc.Debugf("Removing managed ProvisionWatcher: %s", pw.Name)
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.NewString()) // nolint:staticcheck
	_, err := container.ProvisionWatcherClientFrom(s.dic.Get).DeleteProvisionWatcherByName(ctx, name)
	if err != nil {
		s.lc.Errorf("failed to delete ProvisionWatcher %s in Core Metadata", name)
		return err
	}

	return nil
}

// UpdateProvisionWatcher updates the Watcher in the cache and ensures that the
// copy in Core Metadata is also updated.
func (s *deviceService) UpdateProvisionWatcher(watcher models.ProvisionWatcher) error {
	_, ok := cache.ProvisionWatchers().ForName(watcher.Name)
	if !ok {
		msg := fmt.Sprintf("failed to find ProvisionWatcher %s in cache", watcher.Name)
		s.lc.Error(msg)
		return errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, msg, nil)
	}

	s.lc.Debugf("Updating managed ProvisionWatcher: %s", watcher.Name)
	req := requests.NewUpdateProvisionWatcherRequest(dtos.FromProvisionWatcherModelToUpdateDTO(watcher))
	req.ProvisionWatcher.Id = nil
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.NewString()) // nolint:staticcheck
	_, err := container.ProvisionWatcherClientFrom(s.dic.Get).Update(ctx, []requests.UpdateProvisionWatcherRequest{req})
	if err != nil {
		s.lc.Errorf("failed to update ProvisionWatcher %s in Core Metadata: %v", watcher.Name, err)
		return err
	}

	return nil
}
