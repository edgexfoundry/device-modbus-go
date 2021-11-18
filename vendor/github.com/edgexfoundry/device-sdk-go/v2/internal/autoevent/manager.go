// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2019-2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package autoevent

import (
	"context"
	"sync"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"

	"github.com/edgexfoundry/device-sdk-go/v2/internal/cache"
	"github.com/edgexfoundry/device-sdk-go/v2/internal/container"
)

type manager struct {
	executorMap     map[string][]*Executor
	ctx             context.Context
	wg              *sync.WaitGroup
	mutex           sync.Mutex
	autoeventBuffer chan bool
	dic             *di.Container
}

func BootstrapHandler(
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
	}

	dic.Update(di.ServiceConstructorMap{
		container.ManagerName: func(get di.Get) interface{} {
			return m
		},
	})

	return true
}

func (m *manager) StartAutoEvents() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, d := range cache.Devices().All() {
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
		executor, err := NewExecutor(deviceName, autoEvent)
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
