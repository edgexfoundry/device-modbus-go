// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018-2025 IOTech Ltd
// Copyright (C) 2026 encentive GmbH
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"sync/atomic"

	MODBUS "github.com/grid-x/modbus"
)

// ModbusClient is used for connecting the device and read/write value
type ModbusClient struct {
	// ModbusType is used for holding the connection type
	ModbusType string
	// TCPClientHandler is used for holding device TCP connection
	TCPClientHandler MODBUS.TCPClientHandler
	// sharedTCP uses a per-UnitID packager and a socket owned by the driver pool.
	sharedTCP    bool
	tcpPackager  *MODBUS.TCPClientHandler
	tcpTransport *tcpTransportSlot
	closeOnce    sync.Once
	closeErr     error
	closed       atomic.Bool
	// RTUClientHandler is used for holding device RTU connection
	RTUClientHandler MODBUS.RTUClientHandler
	// ASCIIClientHandler is used for holding device ASCII connection
	ASCIIClientHandler MODBUS.ASCIIClientHandler
	// Context is used for holding the context of the client
	ctx context.Context

	client MODBUS.Client
}

func (c *ModbusClient) OpenConnection() error {
	if c.closed.Load() {
		return fmt.Errorf("Modbus client is closed")
	}
	if err := c.ctx.Err(); err != nil {
		return err
	}
	var newClient MODBUS.Client
	switch c.ModbusType {
	case ProtocolTCP:
		if c.sharedTCP {
			// The pool already dialed the shared transport. Connecting the packager would open a second socket.
			c.client = MODBUS.NewClient2(c.tcpPackager, c.tcpTransport)
			driver.Logger.Info("Modbus client uses shared TCP transport.")
			return nil
		}
		err := c.TCPClientHandler.Connect(c.ctx)
		if err != nil {
			driver.Logger.Errorf("client %+v failed to connect to Modbus device: %v", &c, err)
			return err
		}
		newClient = MODBUS.NewClient(&c.TCPClientHandler)
		driver.Logger.Info("Modbus client create TCP connection.")
	case ProtocolRTU:
		err := c.RTUClientHandler.Connect(c.ctx)
		if err != nil {
			driver.Logger.Errorf("client %+v failed to connect to Modbus device: %v", &c, err)
			return err
		}
		newClient = MODBUS.NewClient(&c.RTUClientHandler)
		driver.Logger.Info("Modbus client create RTU connection.")
	case ProtocolASCII:
		err := c.ASCIIClientHandler.Connect(c.ctx)
		if err != nil {
			driver.Logger.Errorf("client %+v failed to connect to Modbus device: %v", &c, err)
			return err
		}
		newClient = MODBUS.NewClient(&c.ASCIIClientHandler)
		driver.Logger.Info("Modbus client create ASCII connection.")
	default:
		driver.Logger.Errorf("modbus connection type don't support!")
		return fmt.Errorf("modbus connection type %v don't support", c.ModbusType)
	}
	c.client = newClient
	return nil
}

func (c *ModbusClient) CloseConnection() error {
	c.closeOnce.Do(func() {
		c.closed.Store(true)
		c.closeErr = c.closeConnection()
	})
	return c.closeErr
}

func (c *ModbusClient) closeConnection() error {
	if c.sharedTCP {
		// Each client owns one lease, released exactly once by CloseConnection.
		return c.tcpTransport.release()
	}
	var err error
	switch c.ModbusType {
	case ProtocolTCP:
		err = c.TCPClientHandler.Close()
	case ProtocolRTU:
		err = c.RTUClientHandler.Close()
	case ProtocolASCII:
		err = c.ASCIIClientHandler.Close()
	default:
		driver.Logger.Errorf("modbus connection type don't support!")
		return fmt.Errorf("modbus connection type %v don't support", c.ModbusType)
	}
	return err
}

func (c *ModbusClient) GetValue(commandInfo interface{}) ([]byte, error) {
	if c.closed.Load() {
		return nil, fmt.Errorf("Modbus client is closed")
	}
	if err := c.ctx.Err(); err != nil {
		return nil, err
	}
	var modbusCommandInfo = commandInfo.(*CommandInfo)

	// Reading value from device
	var response []byte
	var err error
	requestClient, transport := c.requestClient()

	switch modbusCommandInfo.PrimaryTable {
	case DISCRETES_INPUT, DISCRETE_INPUTS:
		response, err = requestClient.ReadDiscreteInputs(c.ctx, modbusCommandInfo.StartingAddress, modbusCommandInfo.Length)
	case COILS:
		response, err = requestClient.ReadCoils(c.ctx, modbusCommandInfo.StartingAddress, modbusCommandInfo.Length)

	case INPUT_REGISTERS:
		response, err = requestClient.ReadInputRegisters(c.ctx, modbusCommandInfo.StartingAddress, modbusCommandInfo.Length)
	case HOLDING_REGISTERS:
		response, err = requestClient.ReadHoldingRegisters(c.ctx, modbusCommandInfo.StartingAddress, modbusCommandInfo.Length)
	default:
		driver.Logger.Error("None supported primary table! ")
	}

	if err != nil {
		if transport != nil && !transport.sent {
			return response, err
		}
		return response, &modbusRequestError{err}
	}

	driver.Logger.Tracef("Modbus client GetValue's results %v", response)

	return response, nil
}

func (c *ModbusClient) SetValue(commandInfo interface{}, value []byte) error {
	if c.closed.Load() {
		return fmt.Errorf("Modbus client is closed")
	}
	if err := c.ctx.Err(); err != nil {
		return err
	}
	var modbusCommandInfo = commandInfo.(*CommandInfo)

	// Write value to device
	var result []byte
	var err error
	requestClient, transport := c.requestClient()

	switch modbusCommandInfo.PrimaryTable {
	case DISCRETES_INPUT, DISCRETE_INPUTS:
		return fmt.Errorf("error: DISCRETES_INPUT is Read-Only")

	case COILS:
		result, err = requestClient.WriteMultipleCoils(c.ctx, uint16(modbusCommandInfo.StartingAddress), modbusCommandInfo.Length, value)

	case INPUT_REGISTERS:
		return fmt.Errorf("error: INPUT_REGISTERS is Read-Only")

	case HOLDING_REGISTERS:
		if modbusCommandInfo.Length == 1 {
			result, err = requestClient.WriteSingleRegister(c.ctx, uint16(modbusCommandInfo.StartingAddress), binary.BigEndian.Uint16(value))
		} else {
			result, err = requestClient.WriteMultipleRegisters(c.ctx, uint16(modbusCommandInfo.StartingAddress), modbusCommandInfo.Length, value)
		}
	default:
	}

	if err != nil {
		if transport != nil && !transport.sent {
			return err
		}
		return &modbusRequestError{err}
	}
	driver.Logger.Tracef("Modbus client SetValue successful, results: %v", result)

	return nil
}

func NewDeviceClient(connectionInfo *ConnectionInfo) (*ModbusClient, error) {
	if err := driver.beginOperation(); err != nil {
		return nil, err
	}
	defer driver.operations.Done()
	return driver.newDeviceClient(connectionInfo)
}

func (d *Driver) newDeviceClient(connectionInfo *ConnectionInfo) (*ModbusClient, error) {
	client := new(ModbusClient)
	client.ctx = d.commandContext()
	if err := client.ctx.Err(); err != nil {
		return nil, err
	}
	client.ModbusType = connectionInfo.Protocol
	switch client.ModbusType {
	case ProtocolTCP:
		if connectionInfo.SharedConnection {
			transport, err := d.getSharedTCPTransport(connectionInfo)
			if err != nil {
				driver.Logger.Errorf("shared TCP transport failed, err:%v", err)
				return nil, err
			}
			packager := MODBUS.NewTCPClientHandler(fmt.Sprintf("%s:%d", connectionInfo.Address, connectionInfo.Port))
			packager.SlaveID = byte(connectionInfo.UnitID)
			client.sharedTCP = true
			client.tcpPackager = packager
			client.tcpTransport = transport
			break
		}
		client.TCPClientHandler = *d.newTCPHandler(connectionInfo)
		client.TCPClientHandler.SlaveID = byte(connectionInfo.UnitID)
	case ProtocolRTU:
		serialParams := strings.Split(connectionInfo.Address, ",")
		client.RTUClientHandler = *MODBUS.NewRTUClientHandler(serialParams[0])
		client.RTUClientHandler.SlaveID = byte(connectionInfo.UnitID)
		client.RTUClientHandler.Timeout = connectionInfo.Timeout
		client.RTUClientHandler.IdleTimeout = connectionInfo.IdleTimeout
		client.RTUClientHandler.BaudRate = connectionInfo.BaudRate
		client.RTUClientHandler.DataBits = connectionInfo.DataBits
		client.RTUClientHandler.StopBits = connectionInfo.StopBits
		client.RTUClientHandler.Parity = connectionInfo.Parity
		client.RTUClientHandler.Logger = log.New(os.Stdout, "", log.LstdFlags)
	case ProtocolASCII:
		serialParams := strings.Split(connectionInfo.Address, ",")
		client.ASCIIClientHandler = *MODBUS.NewASCIIClientHandler(serialParams[0])
		client.ASCIIClientHandler.SlaveID = byte(connectionInfo.UnitID)
		client.ASCIIClientHandler.Timeout = connectionInfo.Timeout
		client.ASCIIClientHandler.IdleTimeout = connectionInfo.IdleTimeout
		client.ASCIIClientHandler.BaudRate = connectionInfo.BaudRate
		client.ASCIIClientHandler.DataBits = connectionInfo.DataBits
		client.ASCIIClientHandler.StopBits = connectionInfo.StopBits
		client.ASCIIClientHandler.Parity = connectionInfo.Parity
		client.ASCIIClientHandler.Logger = log.New(os.Stdout, "", log.LstdFlags)
	default:
		driver.Logger.Errorf("modbus connection type don't support!")
		return nil, fmt.Errorf("modbus connection type %v don't support", client.ModbusType)
	}
	err := client.OpenConnection()
	if err != nil {
		driver.Logger.Errorf("client %+v open connection failed,err:%v", client, err)
		_ = client.CloseConnection()
		return nil, err
	}
	return client, nil
}

// Preserve the cause while distinguishing device requests from local validation
// and transformation errors, which must not reset another unit's connection.
type modbusRequestError struct {
	err error
}

func (e *modbusRequestError) Error() string { return e.err.Error() }
func (e *modbusRequestError) Unwrap() error { return e.err }

// A request-local tracker distinguishes library validation (before Send) from
// protocol errors after Send without matching error strings or duplicating the
// library's validation rules. Unit clients keep their own packager/counter.
type requestTransport struct {
	MODBUS.Transporter
	sent bool
}

func (t *requestTransport) Send(ctx context.Context, request []byte) ([]byte, error) {
	t.sent = true
	response, err := t.Transporter.Send(ctx, request)
	// The library represents any differing function code as *MODBUS.Error.
	// Only the requested function with its exception bit set is an exception.
	if err == nil && len(request) > 7 && len(response) > 7 && response[7] != request[7] && response[7] != request[7]|0x80 {
		return nil, fmt.Errorf("modbus: response function '%d' does not match request '%d'", response[7], request[7])
	}
	return response, err
}

func (c *ModbusClient) requestClient() (MODBUS.Client, *requestTransport) {
	if !c.sharedTCP {
		return c.client, nil
	}
	transport := &requestTransport{Transporter: c.tcpTransport}
	return MODBUS.NewClient2(c.tcpPackager, transport), transport
}
