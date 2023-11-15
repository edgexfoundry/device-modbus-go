//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"fmt"
	"sync"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"
)

var (
	pwc *provisionWatcherCache
)

type ProvisionWatcherCache interface {
	ForName(name string) (models.ProvisionWatcher, bool)
	All() []models.ProvisionWatcher
	Add(device models.ProvisionWatcher) errors.EdgeX
	Update(device models.ProvisionWatcher) errors.EdgeX
	RemoveByName(name string) errors.EdgeX
	UpdateAdminState(name string, state models.AdminState) errors.EdgeX
}

type provisionWatcherCache struct {
	pwMap map[string]*models.ProvisionWatcher // key is ProvisionWatcher name
	mutex sync.RWMutex
}

func newProvisionWatcherCache(pws []models.ProvisionWatcher) ProvisionWatcherCache {
	defaultSize := len(pws)
	pwMap := make(map[string]*models.ProvisionWatcher, defaultSize)
	for i, pw := range pws {
		pwMap[pw.Name] = &pws[i]
	}

	pwc = &provisionWatcherCache{pwMap: pwMap}
	return pwc
}

// ForName returns a provision watcher with the given name.
func (p *provisionWatcherCache) ForName(name string) (models.ProvisionWatcher, bool) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	watcher, ok := p.pwMap[name]
	if !ok {
		return models.ProvisionWatcher{}, false
	}
	return *watcher, ok
}

// All returns the current list of provision watchers in the cache.
func (p *provisionWatcherCache) All() []models.ProvisionWatcher {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	i := 0
	watchers := make([]models.ProvisionWatcher, len(p.pwMap))
	for _, watcher := range p.pwMap {
		watchers[i] = *watcher
		i++
	}
	return watchers
}

// Add adds a new provision watcher to the cache.
func (p *provisionWatcherCache) Add(watcher models.ProvisionWatcher) errors.EdgeX {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.add(watcher)
}

func (p *provisionWatcherCache) add(watcher models.ProvisionWatcher) errors.EdgeX {
	if _, ok := p.pwMap[watcher.Name]; ok {
		errMsg := fmt.Sprintf("ProvisionWatcher %s has already existed in cache", watcher.Name)
		return errors.NewCommonEdgeX(errors.KindDuplicateName, errMsg, nil)
	}

	p.pwMap[watcher.Name] = &watcher
	return nil
}

// Update updates the provision watcher in the cache
func (p *provisionWatcherCache) Update(watcher models.ProvisionWatcher) errors.EdgeX {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if err := p.removeByName(watcher.Name); err != nil {
		return err
	}
	return p.add(watcher)
}

// RemoveByName removes the specified provision watcher by name from the cache.
func (p *provisionWatcherCache) RemoveByName(name string) errors.EdgeX {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.removeByName(name)
}

func (p *provisionWatcherCache) removeByName(name string) errors.EdgeX {
	_, ok := p.pwMap[name]
	if !ok {
		errMsg := fmt.Sprintf("failed to find ProvisionWatcher %s in cache", name)
		return errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, errMsg, nil)
	}

	delete(p.pwMap, name)
	return nil
}

// UpdateAdminState updates the ProvisionWatcher admin state in cache by name.
func (p *provisionWatcherCache) UpdateAdminState(name string, state models.AdminState) errors.EdgeX {
	if state != models.Locked && state != models.Unlocked {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "invalid AdminState", nil)
	}

	p.mutex.Lock()
	defer p.mutex.Unlock()

	_, ok := p.pwMap[name]
	if !ok {
		errMsg := fmt.Sprintf("failed to find ProvisionWatcher %s in cache", name)
		return errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, errMsg, nil)
	}

	p.pwMap[name].AdminState = state
	return nil
}

func ProvisionWatchers() ProvisionWatcherCache {
	return pwc
}
