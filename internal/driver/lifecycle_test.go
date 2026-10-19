// Copyright (C) 2026 encentive GmbH
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	MODBUS "github.com/grid-x/modbus"

	"github.com/edgexfoundry/device-sdk-go/v4/pkg/interfaces"
	sdkModel "github.com/edgexfoundry/device-sdk-go/v4/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
)

func awaitResult[T any](t *testing.T, result <-chan T) T {
	t.Helper()
	select {
	case value := <-result:
		return value
	case <-time.After(5 * time.Second):
		t.Fatal("operation did not finish")
		var zero T
		return zero
	}
}

// Observe state transitions, rather than assuming a goroutine has reached a
// particular instruction after an arbitrary sleep.
func awaitState(t *testing.T, ready func() bool) {
	t.Helper()
	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()
	deadline := time.After(5 * time.Second)
	for !ready() {
		select {
		case <-ticker.C:
		case <-deadline:
			t.Fatal("state transition did not occur")
		}
	}
}

func TestSharedTCPShutdownDuringDial(t *testing.T) {
	d, srv := startShareTest(t)
	entered, release := make(chan struct{}), make(chan struct{})
	var releaseOnce sync.Once
	unblock := func() { releaseOnce.Do(func() { close(release) }) }
	t.Cleanup(unblock)
	clientConn, peer := net.Pipe()
	t.Cleanup(func() { _ = peer.Close() })
	d.tcpDial = func(context.Context, string, string) (net.Conn, error) {
		close(entered)
		<-release // Simulate a dialer returning a connection after cancellation.
		return clientConn, nil
	}
	info, err := createConnectionInfo(tcpProtocols(srv, 1, true))
	if err != nil {
		t.Fatal(err)
	}
	acquired := make(chan error, 1)
	go func() {
		_, err := d.getSharedTCPTransport(info)
		acquired <- err
	}()
	awaitResult(t, entered)
	d.transportMutex.Lock()
	poolContext := d.transportContext
	d.transportMutex.Unlock()
	stopped := make(chan struct{})
	go func() { d.closeSharedTransports(); close(stopped) }()
	awaitResult(t, poolContext.Done())
	unblock()
	if err := awaitResult(t, acquired); err == nil {
		t.Fatal("shutdown published a late connection")
	}
	awaitResult(t, stopped)
	_ = peer.SetReadDeadline(time.Now().Add(time.Second))
	if _, err := peer.Read(make([]byte, 1)); !errors.Is(err, io.EOF) {
		t.Fatalf("late connection was not closed: %v", err)
	}
	if _, err := d.getSharedTCPTransport(info); !errors.Is(err, errTransportClosed) {
		t.Fatalf("closed pool accepted an acquisition: %v", err)
	}
	for _, slot := range d.tcpTransports {
		if !slot.closed || slot.handler != nil || slot.refs != 0 {
			t.Fatal("shutdown left a live transport")
		}
	}
}

func TestSharedTCPClosedSlotCannotReconnect(t *testing.T) {
	d, srv := startShareTest(t)
	readUnit(t, d, srv, 1, true)
	client := d.clientMap["modbus-tcp:"+srv.host+"1"].(*ModbusClient)
	request, err := client.tcpPackager.Encode(&MODBUS.ProtocolDataUnit{FunctionCode: MODBUS.FuncCodeReadHoldingRegisters, Data: []byte{0, 0, 0, 1}})
	if err != nil {
		t.Fatal(err)
	}
	d.closeSharedTransports()
	if _, err := client.tcpTransport.Send(context.Background(), request); !errors.Is(err, errTransportClosed) {
		t.Fatalf("closed slot allowed Send: %v", err)
	}
	if err := readUnitErr(d, srv, 2, true); err == nil {
		t.Fatal("closed pool allowed a new unit client")
	}
	if srv.accepts.Load() != 1 {
		t.Fatal("closed slot reconnected")
	}
}

func TestSharedTCPClientReleasesOnlyItsOwnLease(t *testing.T) {
	d, srv := startShareTest(t)
	readUnit(t, d, srv, 1, true)
	readUnit(t, d, srv, 2, true)
	client := d.clientMap["modbus-tcp:"+srv.host+"1"].(*ModbusClient)
	for range 2 {
		if err := client.CloseConnection(); err != nil {
			t.Fatal(err)
		}
	}
	if client.tcpTransport.refs != 1 {
		t.Fatal("repeated CloseConnection released another client's lease")
	}
	readUnit(t, d, srv, 2, true)
	if srv.accepts.Load() != 1 {
		t.Fatal("closing one client closed another unit's connection")
	}
}

func TestUpdateDeviceChangesSharedConnection(t *testing.T) {
	for _, initial := range []bool{false, true} {
		t.Run(fmt.Sprintf("initial=%t", initial), func(t *testing.T) {
			d, srv := startShareTest(t)
			readUnit(t, d, srv, 2, true)
			readUnit(t, d, srv, 1, initial)
			old := d.clientMap["modbus-tcp:"+srv.host+"1"].(*ModbusClient)
			if err := d.UpdateDevice("unit-1", tcpProtocols(srv, 1, !initial), models.Unlocked); err != nil {
				t.Fatal(err)
			}
			if !old.closed.Load() {
				t.Fatal("update did not close the previous client")
			}
			if err := readUnitErr(d, srv, 1, initial); err == nil {
				t.Fatal("stale command recreated the previous configuration")
			}
			readUnit(t, d, srv, 1, !initial)
			current := d.clientMap["modbus-tcp:"+srv.host+"1"].(*ModbusClient)
			if current == old || current.sharedTCP == initial {
				t.Fatal("update reused the old connection mode")
			}
			readUnit(t, d, srv, 2, true)
			if srv.accepts.Load() != 2 {
				t.Fatalf("expected one shared and one separate session over the update, got %d", srv.accepts.Load())
			}
		})
	}
}

func TestDeviceChangeWaitsForRunningCommand(t *testing.T) {
	for _, remove := range []bool{false, true} {
		t.Run(fmt.Sprintf("remove=%t", remove), func(t *testing.T) {
			d, srv := startShareTest(t)
			entered, release := make(chan struct{}), make(chan struct{})
			var once sync.Once
			unblock := func() { once.Do(func() { close(release) }) }
			t.Cleanup(unblock)
			srv.mu.Lock()
			srv.respond = func(response []byte) []byte { close(entered); <-release; return response }
			srv.mu.Unlock()
			read := make(chan error, 1)
			go func() { read <- readUnitErr(d, srv, 1, true) }()
			awaitResult(t, entered)
			changed := make(chan error, 1)
			go func() {
				if remove {
					changed <- d.RemoveDevice("unit-1", tcpProtocols(srv, 1, true))
				} else {
					changed <- d.UpdateDevice("unit-1", tcpProtocols(srv, 1, false), models.Unlocked)
				}
			}()
			endpoint := d.deviceEndpoint("unit-1")
			awaitState(t, func() bool {
				if endpoint.mu.TryRLock() {
					endpoint.mu.RUnlock()
					return false
				}
				return true // Writer is waiting behind the running command.
			})
			// Another device's lifecycle must not wait on this device's I/O.
			if err := d.AddDevice("other", tcpProtocols(srv, 3, true), models.Unlocked); err != nil {
				t.Fatal(err)
			}
			select {
			case err := <-changed:
				t.Fatalf("device changed before its command completed: %v", err)
			default:
			}
			unblock()
			if err := awaitResult(t, read); err != nil {
				t.Fatal(err)
			}
			if err := awaitResult(t, changed); err != nil {
				t.Fatal(err)
			}
			if err := readUnitErr(d, srv, 1, true); err == nil {
				t.Fatal("stale command survived the device change")
			}
		})
	}
}

func TestConcurrentDeviceLifecycle(t *testing.T) {
	d, srv := startShareTest(t)
	start := make(chan struct{})
	errs := make(chan error, 24)
	var wg sync.WaitGroup
	for i := range 24 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			name := fmt.Sprintf("device-%d", i)
			for j := range 10 {
				protocols := tcpProtocols(srv, byte(i+1), j%2 == 0)
				protocols[ProtocolTCP][Address] = fmt.Sprintf("192.0.2.%d", i+1)
				if err := d.AddDevice(name, protocols, models.Unlocked); err != nil {
					errs <- err
					return
				}
				if err := d.UpdateDevice(name, protocols, models.Unlocked); err != nil {
					errs <- err
					return
				}
				if err := d.RemoveDevice(name, protocols); err != nil {
					errs <- err
					return
				}
			}
		}(i)
	}
	close(start)
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Error(err)
	}
}

func TestGracefulStopDrainsCommand(t *testing.T) {
	d, srv := startShareTest(t)
	entered, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	unblock := func() { once.Do(func() { close(release) }) }
	t.Cleanup(unblock)
	srv.mu.Lock()
	srv.respond = func(response []byte) []byte { close(entered); <-release; return response }
	srv.mu.Unlock()
	read := make(chan error, 1)
	go func() { read <- readUnitErr(d, srv, 1, true) }()
	awaitResult(t, entered)
	stopped := make(chan error, 1)
	go func() { stopped <- d.Stop(false) }()
	awaitState(t, func() bool { d.lifecycleMutex.Lock(); defer d.lifecycleMutex.Unlock(); return d.stopped })
	if err := readUnitErr(d, srv, 2, true); err == nil {
		t.Fatal("shutdown accepted a new command")
	}
	if err := d.UpdateDevice("unit-1", tcpProtocols(srv, 1, false), models.Unlocked); err == nil {
		t.Fatal("shutdown accepted a device update")
	}
	select {
	case err := <-stopped:
		t.Fatalf("graceful stop returned before the command finished: %v", err)
	default:
	}
	unblock()
	if err := awaitResult(t, read); err != nil {
		t.Fatal(err)
	}
	if err := awaitResult(t, stopped); err != nil {
		t.Fatal(err)
	}
	if len(d.clientMap) != 0 {
		t.Fatal("shutdown retained cached clients")
	}
}

func TestForcedStopCancelsQueuedAddress(t *testing.T) {
	d, srv := startShareTest(t)
	address := fmt.Sprintf("%s:%d", srv.addr, srv.port)
	if err := d.lockAddress(address); err != nil {
		t.Fatal(err)
	}
	queued := make(chan error, 1)
	go func() { queued <- readUnitErr(d, srv, 1, true) }()
	awaitState(t, func() bool {
		d.addressMutex.Lock()
		defer d.addressMutex.Unlock()
		return d.workingAddressCount[address] == 2
	})
	if err := d.Stop(true); err != nil {
		t.Fatal(err)
	}
	if err := awaitResult(t, queued); !errors.Is(err, context.Canceled) {
		t.Fatalf("queued command was not canceled: %v", err)
	}
	d.unlockAddress(address) // Must remain safe after forced shutdown.
	if d.workingAddressCount[address] != 0 || srv.accepts.Load() != 0 {
		t.Fatal("queued command leaked or dialed after shutdown")
	}
	if err := d.Stop(false); err != nil {
		t.Fatal(err)
	}
}

func TestForcedStopCancelsDial(t *testing.T) {
	d, srv := startShareTest(t)
	entered := make(chan struct{})
	d.tcpDial = func(ctx context.Context, _, _ string) (net.Conn, error) {
		close(entered)
		<-ctx.Done()
		return nil, ctx.Err()
	}
	read := make(chan error, 1)
	go func() { read <- readUnitErr(d, srv, 1, true) }()
	awaitResult(t, entered)
	if err := d.Stop(true); err != nil {
		t.Fatal(err)
	}
	if err := awaitResult(t, read); !errors.Is(err, context.Canceled) {
		t.Fatalf("in-flight dial was not canceled: %v", err)
	}
}

type startupSDK struct {
	interfaces.DeviceServiceSDK
	devices []models.Device
	logging logger.LoggingClient
}

func (s startupSDK) Devices() []models.Device                       { return s.devices }
func (s startupSDK) LoggingClient() logger.LoggingClient            { return s.logging }
func (s startupSDK) AsyncValuesChannel() chan *sdkModel.AsyncValues { return nil }

func TestInitializeTracksExistingMetadataDevices(t *testing.T) {
	d, srv := startShareTest(t)
	protocols := tcpProtocols(srv, 1, false)
	sdk := startupSDK{logging: d.Logger, devices: []models.Device{{Name: "unit-1", Protocols: protocols}}}
	if err := d.Initialize(sdk); err != nil {
		t.Fatal(err)
	}
	if !d.deviceEndpoint("unit-1").known {
		t.Fatal("existing device was not tracked")
	}
	readUnit(t, d, srv, 1, false)
	if err := d.UpdateDevice("unit-1", tcpProtocols(srv, 1, true), models.Unlocked); err != nil {
		t.Fatal(err)
	}
	readUnit(t, d, srv, 1, true)
	if srv.accepts.Load() != 2 {
		t.Fatal("existing device retained its old session")
	}
}

func TestClientCacheChecksActualSharedMode(t *testing.T) {
	d, srv := startShareTest(t)
	readUnit(t, d, srv, 1, true)
	readUnit(t, d, srv, 2, true)
	for _, shared := range []bool{false, true} {
		previous := d.clientMap["modbus-tcp:"+srv.host+"1"].(*ModbusClient)
		info, err := createConnectionInfo(tcpProtocols(srv, 1, shared))
		if err != nil {
			t.Fatal(err)
		}
		current, err := d.createDeviceClient(info)
		if err != nil {
			t.Fatal(err)
		}
		if !previous.closed.Load() || current.(*ModbusClient).sharedTCP != shared {
			t.Fatal("cache reused a client with the wrong sharing mode")
		}
		if _, err := current.GetValue(&CommandInfo{PrimaryTable: HOLDING_REGISTERS, Length: 1}); err != nil {
			t.Fatal(err)
		}
	}
	readUnit(t, d, srv, 2, true)
	if srv.accepts.Load() != 2 {
		t.Fatal("cache replacement disrupted another unit's connection")
	}
}

func TestRemoveUsesTrackedEndpointAfterUpdate(t *testing.T) {
	d, srv := startShareTest(t)
	readUnit(t, d, srv, 1, true)
	readUnit(t, d, srv, 2, true)
	if err := d.UpdateDevice("unit-1", tcpProtocols(srv, 3, true), models.Unlocked); err != nil {
		t.Fatal(err)
	}
	if _, err := d.HandleReadCommands("unit-1", tcpProtocols(srv, 3, true), []sdkModel.CommandRequest{registerRequest()}); err != nil {
		t.Fatal(err)
	}
	// A delayed removal may still carry the old protocol snapshot.
	if err := d.RemoveDevice("unit-1", tcpProtocols(srv, 1, true)); err != nil {
		t.Fatal(err)
	}
	if d.clientMap["modbus-tcp:"+srv.host+"3"] != nil {
		t.Fatal("removal retained the updated client")
	}
	client := d.clientMap["modbus-tcp:"+srv.host+"2"].(*ModbusClient)
	if client.tcpTransport.refs != 1 {
		t.Fatal("removal released the wrong transport lease")
	}
	readUnit(t, d, srv, 2, true)
	if srv.accepts.Load() != 1 {
		t.Fatal("endpoint update disrupted the shared socket")
	}
}

func TestConcurrentUpdatesAndRemovalOfSameDevice(t *testing.T) {
	d, srv := startShareTest(t)
	start := make(chan struct{})
	errs := make(chan error, 8)
	var wg sync.WaitGroup
	for i := range 8 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			for j := range 10 {
				protocols := tcpProtocols(srv, 1, (i+j)%2 == 0)
				if err := d.UpdateDevice("unit-1", protocols, models.Unlocked); err != nil {
					errs <- err
					return
				}
				if err := d.RemoveDevice("unit-1", protocols); err != nil {
					errs <- err
					return
				}
			}
		}(i)
	}
	close(start)
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Error(err)
	}
	if err := d.RemoveDevice("unit-1", tcpProtocols(srv, 1, true)); err != nil {
		t.Fatal(err)
	}
	if err := readUnitErr(d, srv, 1, true); err == nil {
		t.Fatal("concurrent lifecycle operations lost the removal")
	}
}
