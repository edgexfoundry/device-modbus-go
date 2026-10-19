// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2026 encentive GmbH
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"

	MODBUS "github.com/grid-x/modbus"

	sdkModel "github.com/edgexfoundry/device-sdk-go/v4/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
)

func TestIsTransportError(t *testing.T) {
	if isTransportError(nil) {
		t.Fatal("nil is not a transport error")
	}
	if isTransportError(fmt.Errorf("wrapped: %w", &MODBUS.Error{FunctionCode: 0x83, ExceptionCode: 2})) {
		t.Fatal("a Modbus exception must keep the shared socket")
	}
	if !isTransportError(io.EOF) {
		t.Fatal("EOF is a transport error")
	}
	if !isTransportError(errors.New("read tcp 127.0.0.1:1000: i/o timeout")) {
		t.Fatal("timeout text is a transport error")
	}
	if !isTransportError(MODBUS.ErrTCPHeaderLength(0)) {
		t.Fatal("a malformed Modbus header requires a new connection")
	}
}

func TestSharedTCPOneAcceptTwoUnitIDs(t *testing.T) {
	d, srv := startShareTest(t)
	readUnit(t, d, srv, 1, true)
	readUnit(t, d, srv, 2, true)

	if got := srv.accepts.Load(); got != 1 {
		t.Fatalf("shared endpoint opened %d TCP sessions, want 1", got)
	}
	if got := srv.units(); len(got) != 2 || got[0] != 1 || got[1] != 2 {
		t.Fatalf("MBAP unit ids = %v, want [1 2]", got)
	}

	c1 := d.clientMap["modbus-tcp:"+srv.host+"1"].(*ModbusClient)
	c2 := d.clientMap["modbus-tcp:"+srv.host+"2"].(*ModbusClient)
	if c1.tcpTransport == nil || c1.tcpTransport != c2.tcpTransport {
		t.Fatal("UnitIDs do not share one transport pointer")
	}
	if c1.tcpPackager == nil || c1.tcpPackager == c2.tcpPackager {
		t.Fatal("each UnitID needs its own packager")
	}
	if c1.tcpPackager.SlaveID != 1 || c2.tcpPackager.SlaveID != 2 {
		t.Fatalf("slave ids = %d %d, want 1 and 2", c1.tcpPackager.SlaveID, c2.tcpPackager.SlaveID)
	}
}

func TestSharedTCPParallelReadsStaySerialized(t *testing.T) {
	d, srv := startShareTest(t)
	errCh := make(chan error, 2)
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		errCh <- readUnitErr(d, srv, 1, true)
	}()
	go func() {
		defer wg.Done()
		errCh <- readUnitErr(d, srv, 2, true)
	}()
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			t.Fatal(err)
		}
	}
	if got := srv.accepts.Load(); got != 1 {
		t.Fatalf("parallel reads opened %d TCP sessions, want 1", got)
	}
	units := srv.units()
	if len(units) != 2 || !containsByte(units, 1) || !containsByte(units, 2) {
		t.Fatalf("MBAP unit ids = %v, want 1 and 2", units)
	}
}

func TestSharedTCPReconnectsOnceAfterDisconnect(t *testing.T) {
	d, srv := startShareTest(t)
	readUnit(t, d, srv, 1, true)
	readUnit(t, d, srv, 2, true)
	srv.closeClients()

	if err := readUnitErr(d, srv, 1, true); err == nil {
		t.Fatal("read after disconnect succeeded; the broken socket was still accepted")
	}
	readUnit(t, d, srv, 2, true)
	if got := srv.accepts.Load(); got != 2 {
		t.Fatalf("reconnect opened %d sessions, want the original plus one", got)
	}
}

func TestSeparateTCPSessionsWithoutSharedConnection(t *testing.T) {
	d, srv := startShareTest(t)
	readUnit(t, d, srv, 1, false)
	readUnit(t, d, srv, 2, false)
	if got := srv.accepts.Load(); got != 2 {
		t.Fatalf("SharedConnection false opened %d sessions, want 2", got)
	}
	if _, ok := d.tcpTransports[tcpTransportKey(&ConnectionInfo{Protocol: ProtocolTCP, Address: srv.addr, Port: srv.port})]; ok {
		t.Fatal("SharedConnection false must not use the transport pool")
	}
}

func TestRemoveOneUnitKeepsSharedSocket(t *testing.T) {
	d, srv := startShareTest(t)
	readUnit(t, d, srv, 1, true)
	readUnit(t, d, srv, 2, true)

	if err := d.RemoveDevice("unit-1", tcpProtocols(srv, 1, true)); err != nil {
		t.Fatal(err)
	}
	readUnit(t, d, srv, 2, true)
	if got := srv.accepts.Load(); got != 1 {
		t.Fatalf("removing one UnitID opened %d sessions, want the original socket", got)
	}
	slot := d.tcpTransports[tcpTransportKey(&ConnectionInfo{Protocol: ProtocolTCP, Address: srv.addr, Port: srv.port})]
	if slot == nil || slot.handler == nil || slot.refs != 1 {
		t.Fatalf("shared transport after one removal: %+v", slot)
	}

	if err := d.RemoveDevice("unit-2", tcpProtocols(srv, 2, true)); err != nil {
		t.Fatal(err)
	}
	if slot.handler != nil || slot.refs != 0 {
		t.Fatal("removing the last UnitID left the shared socket open")
	}
	if err := readUnitErr(d, srv, 2, true); err == nil {
		t.Fatal("a stale command recreated a removed device")
	}
	if err := d.AddDevice("unit-2", tcpProtocols(srv, 2, true), models.Unlocked); err != nil {
		t.Fatal(err)
	}
	readUnit(t, d, srv, 2, true)
	if got := srv.accepts.Load(); got != 2 {
		t.Fatalf("read after the last removal opened %d sessions, want one new dial", got)
	}
}

type shareServer struct {
	addr    string
	port    int
	host    string
	ln      net.Listener
	accepts atomic.Int32
	mu      sync.Mutex
	seen    []byte
	conns   []net.Conn
	respond func([]byte) []byte
}

func startShareTest(t *testing.T) (*Driver, *shareServer) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	tcpAddr := ln.Addr().(*net.TCPAddr)
	srv := &shareServer{ln: ln, addr: tcpAddr.IP.String(), port: tcpAddr.Port}
	srv.host = srv.addr + ":" + strconv.Itoa(srv.port) + ":"

	prev := driver
	d := &Driver{
		Logger:              logger.NewMockClient(),
		addressMap:          make(map[string]chan bool),
		workingAddressCount: make(map[string]int),
		clientMap:           make(map[string]DeviceClient),
		tcpTransports:       make(map[string]*tcpTransportSlot),
		deviceEndpoints:     make(map[string]*deviceEndpoint),
	}
	d.tcpDial = func(ctx context.Context, network, address string) (net.Conn, error) {
		var dialer net.Dialer
		return dialer.DialContext(ctx, network, address)
	}
	driver = d

	go srv.acceptLoop()
	t.Cleanup(func() {
		_ = d.Stop(true)
		d.operations.Wait()
		_ = ln.Close()
		srv.closeClients()
		driver = prev
	})
	return d, srv
}

func (s *shareServer) acceptLoop() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			return
		}
		s.accepts.Add(1)
		s.mu.Lock()
		s.conns = append(s.conns, conn)
		s.mu.Unlock()
		go s.serve(conn)
	}
}

func (s *shareServer) serve(conn net.Conn) {
	defer conn.Close()
	for {
		header := make([]byte, 7)
		if _, err := io.ReadFull(conn, header); err != nil {
			return
		}
		length := int(binary.BigEndian.Uint16(header[4:6]))
		if length < 2 {
			return
		}
		rest := make([]byte, length-1)
		if _, err := io.ReadFull(conn, rest); err != nil {
			return
		}
		s.mu.Lock()
		s.seen = append(s.seen, header[6])
		s.mu.Unlock()

		pdu := []byte{rest[0], 2, 0x00, 0x2A}
		if rest[0] == MODBUS.FuncCodeWriteSingleRegister {
			pdu = rest
		} else if rest[0] == MODBUS.FuncCodeWriteMultipleRegisters || rest[0] == MODBUS.FuncCodeWriteMultipleCoils {
			pdu = rest[:5]
		} else if rest[0] == MODBUS.FuncCodeReadCoils || rest[0] == MODBUS.FuncCodeReadDiscreteInputs {
			pdu = []byte{rest[0], 1, 1}
		}
		resp := make([]byte, 7+len(pdu))
		copy(resp[:4], header[:4])
		binary.BigEndian.PutUint16(resp[4:], uint16(1+len(pdu)))
		resp[6] = header[6]
		copy(resp[7:], pdu)
		s.mu.Lock()
		respond := s.respond
		s.mu.Unlock()
		if respond != nil {
			resp = respond(resp)
		}
		if _, err := conn.Write(resp); err != nil {
			return
		}
	}
}

func (s *shareServer) units() []byte {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]byte, len(s.seen))
	copy(out, s.seen)
	return out
}

func (s *shareServer) closeClients() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, conn := range s.conns {
		_ = conn.Close()
	}
}

func tcpProtocols(srv *shareServer, unit byte, shared bool) map[string]models.ProtocolProperties {
	props := models.ProtocolProperties{
		Address:     srv.addr,
		Port:        srv.port,
		UnitID:      unit,
		Timeout:     "2s",
		IdleTimeout: "-1s",
	}
	if shared {
		props[SharedConnection] = true
	}
	return map[string]models.ProtocolProperties{ProtocolTCP: props}
}

func readUnit(t *testing.T, d *Driver, srv *shareServer, unit byte, shared bool) {
	t.Helper()
	if err := readUnitErr(d, srv, unit, shared); err != nil {
		t.Fatal(err)
	}
}

func readUnitErr(d *Driver, srv *shareServer, unit byte, shared bool) error {
	req := sdkModel.CommandRequest{
		DeviceResourceName: "holding",
		Type:               common.ValueTypeUint16,
		Attributes: map[string]interface{}{
			PRIMARY_TABLE:    HOLDING_REGISTERS,
			STARTING_ADDRESS: 0,
		},
	}
	_, err := d.HandleReadCommands(fmt.Sprintf("unit-%d", unit), tcpProtocols(srv, unit, shared), []sdkModel.CommandRequest{req})
	return err
}

func containsByte(values []byte, want byte) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
