// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"fmt"
	"sync"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
)

var (
	pc *profileCache
)

type ProfileCache interface {
	ForName(name string) (models.DeviceProfile, bool)
	All() []models.DeviceProfile
	Add(profile models.DeviceProfile) errors.EdgeX
	Update(profile models.DeviceProfile) errors.EdgeX
	RemoveByName(name string) errors.EdgeX
	DeviceResource(profileName string, resourceName string) (models.DeviceResource, bool)
	DeviceCommand(profileName string, commandName string) (models.DeviceCommand, bool)
	ResourceOperation(profileName string, deviceResource string) (models.ResourceOperation, errors.EdgeX)
}

type profileCache struct {
	deviceProfileMap  map[string]*models.DeviceProfile // key is DeviceProfile name
	deviceResourceMap map[string]map[string]models.DeviceResource
	deviceCommandMap  map[string]map[string]models.DeviceCommand
	mutex             sync.RWMutex
}

func newProfileCache(profiles []models.DeviceProfile) ProfileCache {
	defaultSize := len(profiles)
	dpMap := make(map[string]*models.DeviceProfile, defaultSize)
	drMap := make(map[string]map[string]models.DeviceResource, defaultSize)
	dcMap := make(map[string]map[string]models.DeviceCommand, defaultSize)
	for i, dp := range profiles {
		dpMap[dp.Name] = &profiles[i]
		drMap[dp.Name] = deviceResourceSliceToMap(dp.DeviceResources)
		dcMap[dp.Name] = deviceCommandSliceToMap(dp.DeviceCommands)
	}

	pc = &profileCache{
		deviceProfileMap:  dpMap,
		deviceResourceMap: drMap,
		deviceCommandMap:  dcMap,
	}
	return pc
}

// ForName returns a profile with the given profile name.
func (p *profileCache) ForName(name string) (models.DeviceProfile, bool) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	profile, ok := p.deviceProfileMap[name]
	if !ok {
		return models.DeviceProfile{}, false
	}
	return *profile, ok
}

// All returns the current list of profiles in the cache.
func (p *profileCache) All() []models.DeviceProfile {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	i := 0
	ps := make([]models.DeviceProfile, len(p.deviceProfileMap))
	for _, profile := range p.deviceProfileMap {
		ps[i] = *profile
		i += 1
	}
	return ps
}

// Add adds a new profile to the cache. This method is used to populate the
// profile cache with pre-existing or recently-added profiles from Core Metadata.
func (p *profileCache) Add(profile models.DeviceProfile) errors.EdgeX {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.add(profile)
}

func (p *profileCache) add(profile models.DeviceProfile) errors.EdgeX {
	if _, ok := p.deviceProfileMap[profile.Name]; ok {
		errMsg := fmt.Sprintf("Profile %s has already existed in cache", profile.Name)
		return errors.NewCommonEdgeX(errors.KindDuplicateName, errMsg, nil)
	}

	p.deviceProfileMap[profile.Name] = &profile
	p.deviceResourceMap[profile.Name] = deviceResourceSliceToMap(profile.DeviceResources)
	p.deviceCommandMap[profile.Name] = deviceCommandSliceToMap(profile.DeviceCommands)
	return nil
}

func deviceResourceSliceToMap(deviceResources []models.DeviceResource) map[string]models.DeviceResource {
	result := make(map[string]models.DeviceResource, len(deviceResources))
	for _, dr := range deviceResources {
		result[dr.Name] = dr
	}

	return result
}

func deviceCommandSliceToMap(deviceCommands []models.DeviceCommand) map[string]models.DeviceCommand {
	result := make(map[string]models.DeviceCommand, len(deviceCommands))
	for _, dc := range deviceCommands {
		result[dc.Name] = dc
	}

	return result
}

// Update updates the profile in the cache
func (p *profileCache) Update(profile models.DeviceProfile) errors.EdgeX {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if err := p.removeByName(profile.Name); err != nil {
		return err
	}

	return p.add(profile)
}

// RemoveByName removes the specified profile by name from the cache.
func (p *profileCache) RemoveByName(name string) errors.EdgeX {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.removeByName(name)
}

func (p *profileCache) removeByName(name string) errors.EdgeX {
	_, ok := p.deviceProfileMap[name]
	if !ok {
		errMsg := fmt.Sprintf("failed to find Profile %s in cache", name)
		return errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, errMsg, nil)
	}

	delete(p.deviceProfileMap, name)
	delete(p.deviceResourceMap, name)
	delete(p.deviceCommandMap, name)
	return nil
}

// DeviceResource returns the DeviceResource with given profileName and resourceName
func (p *profileCache) DeviceResource(profileName string, resourceName string) (models.DeviceResource, bool) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	drs, ok := p.deviceResourceMap[profileName]
	if !ok {
		return models.DeviceResource{}, false
	}

	dr, ok := drs[resourceName]
	return dr, ok
}

// DeviceCommand returns the DeviceCommand with given profileName and commandName
func (p *profileCache) DeviceCommand(profileName string, commandName string) (models.DeviceCommand, bool) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	dcs, ok := p.deviceCommandMap[profileName]
	if !ok {
		return models.DeviceCommand{}, false
	}

	dc, ok := dcs[commandName]
	return dc, ok
}

// ResourceOperation returns the first matched ResourceOperation with given resourceName
func (p *profileCache) ResourceOperation(profileName string, resourceName string) (models.ResourceOperation, errors.EdgeX) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	if err := p.verifyProfileExists(profileName); err != nil {
		return models.ResourceOperation{}, err
	}

	deviceCommandMap := p.deviceCommandMap[profileName]
	for _, dc := range deviceCommandMap {
		for _, ro := range dc.ResourceOperations {
			if ro.DeviceResource == resourceName {
				return ro, nil
			}
		}
	}

	errMsg := fmt.Sprintf("failed to find ResourceOpertaion with DeviceResource %s in Profile %s", resourceName, profileName)
	return models.ResourceOperation{}, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, errMsg, nil)
}

func (p *profileCache) verifyProfileExists(profileName string) errors.EdgeX {
	if _, ok := p.deviceProfileMap[profileName]; !ok {
		errMsg := fmt.Sprintf("failed to find Profile %s in cache", profileName)
		return errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, errMsg, nil)
	}
	return nil
}

func Profiles() ProfileCache {
	return pc
}
