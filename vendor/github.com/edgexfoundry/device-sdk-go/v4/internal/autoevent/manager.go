// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2019-2024 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package autoevent

import (
	"context"
	"sync"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
	"github.com/panjf2000/ants/v2"

	"github.com/edgexfoundry/device-sdk-go/v4/internal/cache"
	"github.com/edgexfoundry/device-sdk-go/v4/internal/container"
)

type manager struct {
	executorMap     map[string][]*Executor
	ctx             context.Context
	wg              *sync.WaitGroup
	mutex           sync.Mutex
	autoeventBuffer chan bool
	dic             *di.Container
	pool            *ants.Pool
}

type Bootstrap struct {
	pool *ants.Pool
}

func NewBootstrap(p *ants.Pool) *Bootstrap {
	return &Bootstrap{
		pool: p,
	}
}

func (b *Bootstrap) BootstrapHandler(
	ctx context.Context,
	wg *sync.WaitGroup,
	_ startup.Timer,
	dic *di.Container) bool {
	config := container.ConfigurationFrom(dic.Get)
	m := &manager{
		ctx:             ctx,
		wg:              wg,
		executorMap:     make(map[string][]*Executor),
		dic:             dic,
		autoeventBuffer: make(chan bool, config.Device.AsyncBufferSize),
		pool:            b.pool,
	}

	dic.Update(di.ServiceConstructorMap{
		container.AutoEventManagerName: func(get di.Get) interface{} {
			return m
		},
	})

	return true
}

func (m *manager) StartAutoEvents() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, d := range cache.Devices().All() {
		if len(d.ProfileName) == 0 || d.AdminState == models.Locked {
			// don't run the auto event if the device doesn't define the profile, or it is locked
			continue
		}
		if _, ok := m.executorMap[d.Name]; !ok {
			executors := m.triggerExecutors(d.Name, d.AutoEvents, m.dic)
			m.executorMap[d.Name] = executors
		}
	}
}

func (m *manager) triggerExecutors(deviceName string, autoEvents []models.AutoEvent, dic *di.Container) []*Executor {
	var executors []*Executor
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

	for _, autoEvent := range autoEvents {
		executor, err := NewExecutor(deviceName, autoEvent, m.pool)
		if err != nil {
			lc.Errorf("failed to create executor of AutoEvent %s for Device %s: %v", autoEvent.SourceName, deviceName, err)
			// skip this AutoEvent if it causes error during creation
			continue
		}
		executors = append(executors, executor)
		go executor.Run(m.ctx, m.wg, m.autoeventBuffer, dic)
	}
	return executors
}

func (m *manager) RestartForDevice(deviceName string) {
	lc := bootstrapContainer.LoggingClientFrom(m.dic.Get)

	m.StopForDevice(deviceName)
	d, ok := cache.Devices().ForName(deviceName)
	if !ok {
		lc.Errorf("failed to find device %s in cache to start AutoEvent", deviceName)
	}

	if len(d.ProfileName) == 0 || d.AdminState == models.Locked {
		// don't run the auto event if the device doesn't define the profile, or it is locked
		return
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()
	executors := m.triggerExecutors(deviceName, d.AutoEvents, m.dic)
	m.executorMap[deviceName] = executors
}

func (m *manager) StopForDevice(deviceName string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	executors, ok := m.executorMap[deviceName]
	if ok {
		for _, executor := range executors {
			executor.Stop()
		}
		delete(m.executorMap, deviceName)
	}
}
