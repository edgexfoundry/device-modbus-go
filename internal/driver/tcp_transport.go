// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2026 encentive GmbH
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"

	MODBUS "github.com/grid-x/modbus"
)

// tcpTransportSlot is one shared TCP handler for an address and port.
// The mutex covers that endpoint only, so a dial to one server does not block another.
type tcpTransportSlot struct {
	mu      sync.Mutex
	handler *MODBUS.TCPClientHandler
	refs    int
	closed  bool
	ctx     context.Context
}

var errTransportClosed = errors.New("shared TCP transport is closed")

// Send is the only path from a unit client to the shared handler. Close on the
// library handler alone is insufficient: its next Send would reconnect.
func (slot *tcpTransportSlot) Send(ctx context.Context, request []byte) ([]byte, error) {
	slot.mu.Lock()
	defer slot.mu.Unlock()
	if slot.closed || slot.ctx.Err() != nil || slot.handler == nil {
		return nil, errTransportClosed
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	// Classify the completed request in noteCommandError before resetting the
	// socket. Recovering here as well would close it twice for the same error.
	return slot.handler.Send(ctx, request)
}

func (slot *tcpTransportSlot) release() error {
	slot.mu.Lock()
	defer slot.mu.Unlock()
	if slot.refs > 0 {
		slot.refs--
	}
	if slot.refs == 0 && slot.handler != nil {
		err := slot.handler.Close()
		slot.handler = nil
		return err
	}
	return nil
}

// tcpTransportKey identifies a TCP endpoint without the UnitID.
func tcpTransportKey(info *ConnectionInfo) string {
	return fmt.Sprintf("%s:%s:%d", info.Protocol, info.Address, info.Port)
}

func (d *Driver) newTCPHandler(info *ConnectionInfo) *MODBUS.TCPClientHandler {
	options := []MODBUS.TCPClientHandlerOption{}
	if d.tcpDial != nil {
		options = append(options, MODBUS.WithDialer(d.tcpDial))
	}
	handler := MODBUS.NewTCPClientHandler(fmt.Sprintf("%s:%d", info.Address, info.Port), options...)
	handler.Timeout = info.Timeout
	handler.IdleTimeout = info.IdleTimeout
	handler.ProtocolRecoveryTimeout = info.ProtocolRecoveryTimeout
	handler.LinkRecoveryTimeout = info.LinkRecoveryTimeout
	handler.ConnectDelay = info.ConnectDelay
	handler.Logger = log.New(os.Stdout, "", log.LstdFlags)
	return handler
}

// getSharedTCPTransport returns the shared slot for this endpoint and counts one more unit client.
// The first caller dials. Later callers receive the same pointer.
func (d *Driver) getSharedTCPTransport(info *ConnectionInfo) (*tcpTransportSlot, error) {
	key := tcpTransportKey(info)

	d.transportMutex.Lock()
	if d.transportClosed {
		d.transportMutex.Unlock()
		return nil, errTransportClosed
	}
	if d.transportContext == nil {
		d.transportContext, d.cancelTransports = context.WithCancel(d.commandContext())
	}
	if d.tcpTransports == nil {
		d.tcpTransports = make(map[string]*tcpTransportSlot)
	}
	slot, ok := d.tcpTransports[key]
	if !ok {
		slot = &tcpTransportSlot{ctx: d.transportContext}
		d.tcpTransports[key] = slot
	}
	d.transportMutex.Unlock()

	slot.mu.Lock()
	defer slot.mu.Unlock()
	if slot.closed || slot.ctx.Err() != nil {
		return nil, errTransportClosed
	}

	if slot.handler != nil {
		d.warnIfTCPTimingsDiffer(key, slot.handler, info)
		slot.refs++
		return slot, nil
	}

	handler := d.newTCPHandler(info)
	if err := handler.Connect(slot.ctx); err != nil {
		_ = handler.Close()
		return nil, err
	}
	// Shutdown can begin while Connect is blocked, even with a custom dialer
	// which does not honor cancellation. Never publish its late connection.
	if slot.ctx.Err() != nil {
		_ = handler.Close()
		return nil, errTransportClosed
	}
	slot.handler = handler
	slot.refs++
	return slot, nil
}

func (d *Driver) warnIfTCPTimingsDiffer(key string, handler *MODBUS.TCPClientHandler, info *ConnectionInfo) {
	if handler.Timeout == info.Timeout &&
		handler.IdleTimeout == info.IdleTimeout &&
		handler.ProtocolRecoveryTimeout == info.ProtocolRecoveryTimeout &&
		handler.LinkRecoveryTimeout == info.LinkRecoveryTimeout &&
		handler.ConnectDelay == info.ConnectDelay {
		return
	}
	if d.Logger == nil {
		return
	}
	d.Logger.Warnf("shared TCP transport %s keeps the timings of the first device; UnitID %d requested different Timeout, IdleTimeout, recovery timeouts, or ConnectDelay", key, info.UnitID)
}

// dropSharedConnection closes the socket and leaves the pool entry and unit clients in place.
// The next Send dials once.
func (d *Driver) dropSharedConnection(info *ConnectionInfo) {
	key := tcpTransportKey(info)
	d.transportMutex.Lock()
	slot := d.tcpTransports[key]
	d.transportMutex.Unlock()
	if slot == nil {
		return
	}
	slot.mu.Lock()
	defer slot.mu.Unlock()
	if slot.handler == nil {
		return
	}
	if err := slot.handler.Close(); err != nil && d.Logger != nil {
		d.Logger.Errorf("shared TCP transport reconnect close error, key = %s, err = %v", key, err)
	}
}

func (d *Driver) closeSharedTransports() {
	d.transportMutex.Lock()
	d.transportClosed = true
	if d.cancelTransports != nil {
		d.cancelTransports()
	}
	// Snapshot while locked; no slot can be added after transportClosed is set.
	slots := make(map[string]*tcpTransportSlot, len(d.tcpTransports))
	for key, slot := range d.tcpTransports {
		slots[key] = slot
	}
	d.transportMutex.Unlock()
	for key, slot := range slots {
		slot.mu.Lock()
		slot.closed = true
		if slot.handler != nil {
			if err := slot.handler.Close(); err != nil && d.Logger != nil {
				d.Logger.Errorf("shared TCP transport shutdown close error, key = %s, err = %v", key, err)
			}
			slot.handler = nil
		}
		slot.refs = 0
		slot.mu.Unlock()
	}
}

// noteCommandError keeps a shared socket across Modbus exceptions.
// It owns command-error recovery after response validation and exception decoding.
// The caller still holds the address lock, so another unit cannot send between
// the failed request and its socket reset. Unit clients and leases stay cached.
func (d *Driver) noteCommandError(info *ConnectionInfo, err error) {
	if info != nil && info.Protocol == ProtocolTCP && info.SharedConnection {
		var requestErr *modbusRequestError
		if errors.As(err, &requestErr) && isTransportError(requestErr.err) {
			d.dropSharedConnection(info)
		}
		return
	}
	d.removeDeviceClient(info)
}

func isTransportError(err error) bool {
	if err == nil {
		return false
	}
	var exception *MODBUS.Error
	if errors.As(err, &exception) && exception.FunctionCode&0x80 != 0 && exception.ExceptionCode != 0 {
		return false
	}
	// Only called for errors from a library request, never local transformations.
	// Unrecognized protocol errors also require a fresh TCP stream.
	return true
}
