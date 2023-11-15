// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020-2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package autodiscovery

import (
	"sync"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"

	"github.com/edgexfoundry/device-sdk-go/v3/pkg/interfaces"
)

type discoveryLocker struct {
	busy bool
	mux  sync.Mutex
}

var locker discoveryLocker

func DiscoveryWrapper(driver interfaces.ProtocolDriver, lc logger.LoggingClient) {
	locker.mux.Lock()
	if locker.busy {
		lc.Info("another device discovery process is currently running")
		locker.mux.Unlock()
		return
	}
	locker.busy = true
	locker.mux.Unlock()

	lc.Debug("protocol discovery triggered")
	err := driver.Discover()
	if err != nil {
		lc.Error("failed to trigger protocol discovery", err.Error())
	}

	// ReleaseLock
	locker.mux.Lock()
	locker.busy = false
	locker.mux.Unlock()
}
