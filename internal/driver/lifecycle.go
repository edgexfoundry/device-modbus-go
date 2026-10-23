// Copyright (C) 2026 encentive GmbH
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"context"
	"fmt"
	"sync"
)

// The registry mutex only protects lookup. Never wait for an endpoint while
// holding it. Endpoint -> address -> client -> transport is the lock order.
type deviceEndpoint struct {
	mu      sync.RWMutex
	info    ConnectionInfo
	known   bool
	removed bool
}

// Caller holds lifecycleMutex. Cancellation is reserved for forced shutdown
// (or completion of graceful shutdown), so admitted commands can drain normally.
func (d *Driver) initOperationContext() {
	if d.operationContext == nil {
		d.operationContext, d.cancelOperations = context.WithCancel(context.Background())
	}
}

func (d *Driver) commandContext() context.Context {
	d.lifecycleMutex.Lock()
	defer d.lifecycleMutex.Unlock()
	d.initOperationContext()
	return d.operationContext
}

func (d *Driver) beginOperation() error {
	d.lifecycleMutex.Lock()
	defer d.lifecycleMutex.Unlock()
	if d.stopped {
		return fmt.Errorf("service is stopping and cannot accept new work")
	}
	d.initOperationContext()
	// Stop changes stopped under this same lock before calling Wait.
	d.operations.Add(1)
	return nil
}

func (d *Driver) deviceEndpoint(name string) *deviceEndpoint {
	d.endpointMutex.Lock()
	defer d.endpointMutex.Unlock()
	if d.deviceEndpoints == nil {
		d.deviceEndpoints = make(map[string]*deviceEndpoint)
	}
	endpoint := d.deviceEndpoints[name]
	if endpoint == nil {
		endpoint = &deviceEndpoint{}
		d.deviceEndpoints[name] = endpoint
	}
	return endpoint
}

func (d *Driver) lockDeviceCommand(name string, info *ConnectionInfo) (func(), error) {
	endpoint := d.deviceEndpoint(name)
	endpoint.mu.RLock()
	if !endpoint.known {
		endpoint.mu.RUnlock()
		endpoint.mu.Lock()
		if !endpoint.known {
			endpoint.info, endpoint.known = *info, true
		}
		endpoint.mu.Unlock()
		endpoint.mu.RLock()
	}
	if err := d.commandContext().Err(); err != nil {
		endpoint.mu.RUnlock()
		return nil, err
	}
	if endpoint.removed || endpoint.info != *info {
		endpoint.mu.RUnlock()
		return nil, fmt.Errorf("device %s was removed or its connection configuration changed", name)
	}
	return endpoint.mu.RUnlock, nil
}
