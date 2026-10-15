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
	"io"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"syscall"

	MODBUS "github.com/grid-x/modbus"
)

// tcpTransportSlot is one shared TCP handler for an address and port.
// The mutex covers that endpoint only, so a dial to one server does not block another.
type tcpTransportSlot struct {
	mu      sync.Mutex
	handler *MODBUS.TCPClientHandler
	refs    int
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

// getSharedTCPTransport returns the shared handler for this endpoint and counts one more unit client.
// The first caller dials. Later callers receive the same pointer.
func (d *Driver) getSharedTCPTransport(info *ConnectionInfo) (*MODBUS.TCPClientHandler, error) {
	key := tcpTransportKey(info)

	d.transportMutex.Lock()
	if d.tcpTransports == nil {
		d.tcpTransports = make(map[string]*tcpTransportSlot)
	}
	slot, ok := d.tcpTransports[key]
	if !ok {
		slot = &tcpTransportSlot{}
		d.tcpTransports[key] = slot
	}
	d.transportMutex.Unlock()

	slot.mu.Lock()
	defer slot.mu.Unlock()

	if slot.handler != nil {
		d.warnIfTCPTimingsDiffer(key, slot.handler, info)
		slot.refs++
		return slot.handler, nil
	}

	handler := d.newTCPHandler(info)
	if err := handler.Connect(context.Background()); err != nil {
		return nil, err
	}
	slot.handler = handler
	slot.refs++
	return handler, nil
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

// releaseSharedTransport drops one unit client. The socket closes when the last client is gone.
// The pool slot stays so a later UnitID reuses it and dials again, without racing a map delete.
func (d *Driver) releaseSharedTransport(key string) {
	if d.tcpTransports == nil {
		return
	}
	d.transportMutex.Lock()
	slot := d.tcpTransports[key]
	d.transportMutex.Unlock()
	if slot == nil {
		return
	}

	slot.mu.Lock()
	defer slot.mu.Unlock()
	if slot.refs > 0 {
		slot.refs--
	}
	if slot.refs == 0 && slot.handler != nil {
		if err := slot.handler.Close(); err != nil && d.Logger != nil {
			d.Logger.Errorf("shared TCP transport close error, key = %s, err = %v", key, err)
		}
		slot.handler = nil
	}
}

// dropSharedConnection closes the socket and leaves the pool entry and unit clients in place.
// The next Send dials once.
func (d *Driver) dropSharedConnection(info *ConnectionInfo) {
	if d.tcpTransports == nil {
		return
	}
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
	slots := d.tcpTransports
	d.tcpTransports = make(map[string]*tcpTransportSlot)
	d.transportMutex.Unlock()
	for key, slot := range slots {
		slot.mu.Lock()
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
// A broken TCP connection is closed once so the next command dials again, without dropping other UnitIDs.
func (d *Driver) noteCommandError(info *ConnectionInfo, err error) {
	if info != nil && info.Protocol == ProtocolTCP && info.SharedConnection {
		if isTransportError(err) {
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
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, syscall.ECONNRESET) || errors.Is(err, syscall.EPIPE) || errors.Is(err, syscall.ECONNABORTED) {
		return true
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "modbus:") {
		return false
	}
	for _, part := range []string{
		"eof",
		"timeout",
		"connection reset",
		"broken pipe",
		"connection refused",
		"forcibly closed",
		"use of closed network connection",
		"wsarecv",
		"wsasend",
	} {
		if strings.Contains(msg, part) {
			return true
		}
	}
	return false
}
