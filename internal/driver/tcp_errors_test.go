// Copyright (C) 2026 encentive GmbH
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"encoding/binary"
	"errors"
	"fmt"
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
	param, err := sdkModel.NewCommandValue("holding", common.ValueTypeUint16, value)
	if err != nil {
		t.Fatal(err)
	}
	return d.HandleWriteCommands(fmt.Sprintf("unit-%d", unit), tcpProtocols(srv, unit, true), []sdkModel.CommandRequest{req}, []*sdkModel.CommandValue{param})
}

func TestSharedTCPResponseErrors(t *testing.T) {
	cases := []struct {
		name      string
		exception bool
		mutate    func([]byte) []byte
	}{
		{"exception", true, func(response []byte) []byte {
			response = response[:9]
			binary.BigEndian.PutUint16(response[4:6], 3)
			response[7] |= 0x80
			response[8] = MODBUS.ExceptionCodeIllegalDataAddress
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
		for _, write := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/write=%t", tc.name, write), func(t *testing.T) {
				d, srv := startShareTest(t)
				readUnit(t, d, srv, 1, true)
				readUnit(t, d, srv, 2, true)
				first := d.clientMap["modbus-tcp:"+srv.host+"1"]
				second := d.clientMap["modbus-tcp:"+srv.host+"2"]
				srv.mu.Lock()
				srv.respond = tc.mutate
				srv.mu.Unlock()
				var err error
				if write {
					err = writeUnitErr(t, d, srv, 1, registerRequest(), uint16(42))
				} else {
					err = readUnitErr(d, srv, 1, true)
				}
				if err == nil {
					t.Fatal("invalid response succeeded")
				}
				var exception *MODBUS.Error
				if errors.As(err, &exception) != tc.exception {
					t.Fatalf("exception identity was lost: %v", err)
				}
				srv.mu.Lock()
				srv.respond = nil
				srv.mu.Unlock()
				readUnit(t, d, srv, 2, true)
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
