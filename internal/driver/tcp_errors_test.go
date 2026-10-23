// Copyright (C) 2026 encentive GmbH
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"sync/atomic"
	"testing"

	sdkModel "github.com/edgexfoundry/device-sdk-go/v4/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	MODBUS "github.com/grid-x/modbus"
)

func registerRequest() sdkModel.CommandRequest {
	return sdkModel.CommandRequest{
		DeviceResourceName: "holding",
		Type:               common.ValueTypeUint16,
		Attributes:         map[string]interface{}{PRIMARY_TABLE: HOLDING_REGISTERS, STARTING_ADDRESS: 0},
	}
}

func writeUnitErr(t *testing.T, d *Driver, srv *shareServer, unit byte, req sdkModel.CommandRequest, value interface{}) error {
	t.Helper()
	param, err := sdkModel.NewCommandValue("holding", req.Type, value)
	if err != nil {
		t.Fatal(err)
	}
	return d.HandleWriteCommands(fmt.Sprintf("unit-%d", unit), tcpProtocols(srv, unit, true), []sdkModel.CommandRequest{req}, []*sdkModel.CommandValue{param})
}

func TestSharedTCPResponseErrors(t *testing.T) {
	request := func(table, valueType string) sdkModel.CommandRequest {
		req := registerRequest()
		req.Type = valueType
		req.Attributes[PRIMARY_TABLE] = table
		return req
	}
	operations := []struct {
		name     string
		function byte
		request  sdkModel.CommandRequest
		write    bool
		value    interface{}
	}{
		{"read-holding", MODBUS.FuncCodeReadHoldingRegisters, registerRequest(), false, nil},
		{"read-input", MODBUS.FuncCodeReadInputRegisters, request(INPUT_REGISTERS, common.ValueTypeUint16), false, nil},
		{"read-coils", MODBUS.FuncCodeReadCoils, request(COILS, common.ValueTypeBool), false, nil},
		{"read-discretes", MODBUS.FuncCodeReadDiscreteInputs, request(DISCRETE_INPUTS, common.ValueTypeBool), false, nil},
		{"write-single", MODBUS.FuncCodeWriteSingleRegister, registerRequest(), true, uint16(42)},
		{"write-multiple", MODBUS.FuncCodeWriteMultipleRegisters, request(HOLDING_REGISTERS, common.ValueTypeUint32), true, uint32(42)},
		{"write-coils", MODBUS.FuncCodeWriteMultipleCoils, request(COILS, common.ValueTypeBool), true, true},
	}
	exceptionResponse := func(response []byte) []byte {
		response = response[:9]
		binary.BigEndian.PutUint16(response[4:6], 3)
		response[7] |= 0x80
		response[8] = MODBUS.ExceptionCodeIllegalDataAddress
		return response
	}
	cases := []struct {
		name      string
		exception bool
		mutate    func([]byte) []byte
	}{
		{"exception", true, exceptionResponse},
		{"missing-exception-code", false, func(response []byte) []byte {
			response = exceptionResponse(response)[:8]
			binary.BigEndian.PutUint16(response[4:6], 2)
			return response
		}},
		{"extra-exception-data", false, func(response []byte) []byte {
			response = append(exceptionResponse(response), 0)
			binary.BigEndian.PutUint16(response[4:6], 4)
			return response
		}},
		{"zero-exception-code", false, func(response []byte) []byte {
			response = exceptionResponse(response)
			response[8] = 0
			return response
		}},
		{"transaction", false, func(response []byte) []byte { response[1] ^= 0xff; return response }},
		{"unit", false, func(response []byte) []byte { response[6]++; return response }},
		{"protocol", false, func(response []byte) []byte { response[3] = 1; return response }},
		{"function", false, func(response []byte) []byte { response[7] ^= 1; return response }},
		{"wrong-exception-function", false, func(response []byte) []byte { response[7] = (response[7] | 0x80) ^ 1; return response }},
		{"header-length", false, func(response []byte) []byte { binary.BigEndian.PutUint16(response[4:6], 0); return response }},
		{"payload", false, func(response []byte) []byte { response[8] ^= 0xff; return response }},
	}
	for _, tc := range cases {
		for _, operation := range operations {
			t.Run(tc.name+"/"+operation.name, func(t *testing.T) {
				d, srv := startShareTest(t)
				readUnit(t, d, srv, 1, true)
				readUnit(t, d, srv, 2, true)
				first := d.clientMap["modbus-tcp:"+srv.host+"1"]
				second := d.clientMap["modbus-tcp:"+srv.host+"2"]
				srv.mu.Lock()
				srv.respond = tc.mutate
				srv.mu.Unlock()
				var err error
				if operation.write {
					err = writeUnitErr(t, d, srv, 1, operation.request, operation.value)
				} else {
					_, err = d.HandleReadCommands("unit-1", tcpProtocols(srv, 1, true), []sdkModel.CommandRequest{operation.request})
				}
				if err == nil {
					t.Fatal("invalid response succeeded")
				}
				var exception *MODBUS.Error
				if errors.As(err, &exception) != tc.exception {
					t.Fatalf("exception identity was lost: %v", err)
				}
				if tc.exception && (exception.FunctionCode != operation.function|0x80 || exception.ExceptionCode != MODBUS.ExceptionCodeIllegalDataAddress) {
					t.Fatalf("exception does not match request function %d: %+v", operation.function, exception)
				}
				srv.mu.Lock()
				srv.respond = nil
				srv.mu.Unlock()
				readUnit(t, d, srv, 2, true)
				readUnit(t, d, srv, 1, true)
				wantAccepts := int32(2)
				if tc.exception {
					wantAccepts = 1
				}
				if srv.accepts.Load() != wantAccepts {
					t.Fatalf("accepts=%d, want %d", srv.accepts.Load(), wantAccepts)
				}
				if d.clientMap["modbus-tcp:"+srv.host+"1"] != first || d.clientMap["modbus-tcp:"+srv.host+"2"] != second {
					t.Fatal("response error evicted shared unit clients")
				}
				if first.(*ModbusClient).tcpTransport.refs != 2 {
					t.Fatal("response error changed the unit leases")
				}
			})
		}
	}
}

type closeCountingConn struct {
	net.Conn
	closes *atomic.Int32
}

func (c *closeCountingConn) Close() error {
	c.closes.Add(1)
	return c.Conn.Close()
}

func TestSharedTCPRecoveryAfterClassification(t *testing.T) {
	for _, exception := range []bool{false, true} {
		t.Run(fmt.Sprintf("exception=%t", exception), func(t *testing.T) {
			d, srv := startShareTest(t)
			var closes atomic.Int32
			dial := d.tcpDial
			d.tcpDial = func(ctx context.Context, network, address string) (net.Conn, error) {
				conn, err := dial(ctx, network, address)
				if err != nil {
					return nil, err
				}
				return &closeCountingConn{Conn: conn, closes: &closes}, nil
			}
			readUnit(t, d, srv, 1, true)
			readUnit(t, d, srv, 2, true)
			srv.mu.Lock()
			srv.respond = func(response []byte) []byte {
				if exception {
					response = response[:9]
					binary.BigEndian.PutUint16(response[4:6], 3)
					response[7] |= 0x80
					response[8] = MODBUS.ExceptionCodeIllegalDataAddress
				} else {
					response[1] ^= 0xff
				}
				return response
			}
			srv.mu.Unlock()
			client := d.clientMap["modbus-tcp:"+srv.host+"1"]
			_, err := client.GetValue(&CommandInfo{PrimaryTable: HOLDING_REGISTERS, Length: 1})
			if err == nil {
				t.Fatal("response error was ignored")
			}
			if closes.Load() != 0 {
				t.Fatal("transport reset the socket before command error classification")
			}
			info, infoErr := createConnectionInfo(tcpProtocols(srv, 1, true))
			if infoErr != nil {
				t.Fatal(infoErr)
			}
			d.noteCommandError(info, err)
			wantCloses, wantAccepts := int32(1), int32(2)
			if exception {
				wantCloses, wantAccepts = 0, 1
			}
			if closes.Load() != wantCloses {
				t.Fatalf("socket closes=%d, want %d", closes.Load(), wantCloses)
			}
			srv.mu.Lock()
			srv.respond = nil
			srv.mu.Unlock()
			readUnit(t, d, srv, 2, true)
			readUnit(t, d, srv, 1, true)
			if srv.accepts.Load() != wantAccepts {
				t.Fatalf("TCP sessions=%d, want %d", srv.accepts.Load(), wantAccepts)
			}
		})
	}
}

type errorResponseTransport struct {
	err error
}

func (t errorResponseTransport) Send(context.Context, []byte) ([]byte, error) {
	return nil, t.err
}

func TestRequestTransportValidatesExceptionErrors(t *testing.T) {
	request := []byte{0, 1, 0, 0, 0, 6, 1, MODBUS.FuncCodeReadHoldingRegisters, 0, 0, 0, 1}
	cases := []struct {
		name      string
		err       error
		exception bool
	}{
		{"valid", &MODBUS.Error{FunctionCode: 0x83, ExceptionCode: 2}, true},
		{"wrapped-valid", fmt.Errorf("wrapped: %w", &MODBUS.Error{FunctionCode: 0x83, ExceptionCode: 2}), true},
		{"wrong-function", &MODBUS.Error{FunctionCode: 0x84, ExceptionCode: 2}, false},
		{"wrapped-wrong-function", fmt.Errorf("wrapped: %w", &MODBUS.Error{FunctionCode: 0x84, ExceptionCode: 2}), false},
		{"missing-exception-bit", &MODBUS.Error{FunctionCode: 0x03, ExceptionCode: 2}, false},
		{"zero-exception-code", &MODBUS.Error{FunctionCode: 0x83}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Also guard Transporter implementations that return decoded errors.
			transport := requestTransport{Transporter: errorResponseTransport{err: tc.err}}
			_, err := transport.Send(context.Background(), request)
			if err == nil {
				t.Fatal("exception was lost")
			}
			var exception *MODBUS.Error
			if errors.As(err, &exception) != tc.exception {
				t.Fatalf("invalid exception classification: %v", err)
			}
			if isTransportError(err) == tc.exception {
				t.Fatalf("invalid recovery decision for exception=%t: %v", tc.exception, err)
			}
		})
	}
}

func TestSharedTCPLocalWriteErrorsKeepConnection(t *testing.T) {
	d, srv := startShareTest(t)
	readUnit(t, d, srv, 1, true)
	readUnit(t, d, srv, 2, true)
	req := registerRequest()
	req.Attributes[PRIMARY_TABLE] = INPUT_REGISTERS
	if err := writeUnitErr(t, d, srv, 1, req, uint16(42)); err == nil {
		t.Fatal("write to input registers succeeded")
	}
	// A string cannot be encoded as the numeric register value requested here.
	param, err := sdkModel.NewCommandValue("holding", common.ValueTypeString, "invalid")
	if err != nil {
		t.Fatal(err)
	}
	if err := d.HandleWriteCommands("unit-1", tcpProtocols(srv, 1, true), []sdkModel.CommandRequest{registerRequest()}, []*sdkModel.CommandValue{param}); err == nil {
		t.Fatal("invalid local transformation succeeded")
	}
	readUnit(t, d, srv, 2, true)
	if srv.accepts.Load() != 1 {
		t.Fatal("local validation reset another unit's socket")
	}
}

func TestSharedTCPLibraryValidationKeepsConnection(t *testing.T) {
	d, srv := startShareTest(t)
	readUnit(t, d, srv, 1, true)
	readUnit(t, d, srv, 2, true)
	client := d.clientMap["modbus-tcp:"+srv.host+"1"].(*ModbusClient)
	info, err := createConnectionInfo(tcpProtocols(srv, 1, true))
	if err != nil {
		t.Fatal(err)
	}
	// The library rejects quantities outside the protocol limit before Send.
	_, err = client.GetValue(&CommandInfo{PrimaryTable: HOLDING_REGISTERS, Length: 126})
	if err == nil {
		t.Fatal("invalid register quantity succeeded")
	}
	d.noteCommandError(info, err)
	err = client.SetValue(&CommandInfo{PrimaryTable: HOLDING_REGISTERS, Length: 124}, make([]byte, 248))
	if err == nil {
		t.Fatal("invalid write quantity succeeded")
	}
	d.noteCommandError(info, err)
	readUnit(t, d, srv, 2, true)
	if srv.accepts.Load() != 1 {
		t.Fatal("library validation reset the shared socket")
	}
}
