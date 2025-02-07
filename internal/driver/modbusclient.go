// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018-2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	MODBUS "github.com/goburrow/modbus"
)

// ModbusClient is used for connecting the device and read/write value
type ModbusClient struct {
	// ModbusType is used for holding the connection type
	ModbusType string
	// TCPClientHandler is used for holding device TCP connection
	TCPClientHandler MODBUS.TCPClientHandler
	// TCPClientHandler is used for holding device RTU connection
	RTUClientHandler MODBUS.RTUClientHandler
	// ASCIIClientHandler is used for holding device ASCII connection
	ASCIIClientHandler MODBUS.ASCIIClientHandler

	client MODBUS.Client
}

func (c *ModbusClient) OpenConnection() error {
	var newClient MODBUS.Client
	switch c.ModbusType {
	case ProtocolTCP:
		err := c.TCPClientHandler.Connect()
		if err != nil {
			driver.Logger.Errorf("client %+v failed to connect to Modbus device: %v", &c, err)
			return err
		}
		newClient = MODBUS.NewClient(&c.TCPClientHandler)
		driver.Logger.Info("Modbus client create TCP connection.")
	case ProtocolRTU:
		err := c.RTUClientHandler.Connect()
		if err != nil {
			driver.Logger.Errorf("client %+v failed to connect to Modbus device: %v", &c, err)
			return err
		}
		newClient = MODBUS.NewClient(&c.RTUClientHandler)
		driver.Logger.Info("Modbus client create RTU connection.")
	case ProtocolASCII:
		err := c.ASCIIClientHandler.Connect()
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
	var modbusCommandInfo = commandInfo.(*CommandInfo)

	// Reading value from device
	var response []byte
	var err error

	switch modbusCommandInfo.PrimaryTable {
	case DISCRETES_INPUT, DISCRETE_INPUTS:
		response, err = c.client.ReadDiscreteInputs(modbusCommandInfo.StartingAddress, modbusCommandInfo.Length)
	case COILS:
		response, err = c.client.ReadCoils(modbusCommandInfo.StartingAddress, modbusCommandInfo.Length)

	case INPUT_REGISTERS:
		response, err = c.client.ReadInputRegisters(modbusCommandInfo.StartingAddress, modbusCommandInfo.Length)
	case HOLDING_REGISTERS:
		response, err = c.client.ReadHoldingRegisters(modbusCommandInfo.StartingAddress, modbusCommandInfo.Length)
	default:
		driver.Logger.Error("None supported primary table! ")
	}

	if err != nil {
		return response, err
	}

	driver.Logger.Tracef("Modbus client GetValue's results %v", response)

	return response, nil
}

func (c *ModbusClient) SetValue(commandInfo interface{}, value []byte) error {
	var modbusCommandInfo = commandInfo.(*CommandInfo)

	// Write value to device
	var result []byte
	var err error

	switch modbusCommandInfo.PrimaryTable {
	case DISCRETES_INPUT, DISCRETE_INPUTS:
		err = fmt.Errorf("Error: DISCRETES_INPUT is Read-Only..!!")

	case COILS:
		result, err = c.client.WriteMultipleCoils(uint16(modbusCommandInfo.StartingAddress), modbusCommandInfo.Length, value)

	case INPUT_REGISTERS:
		err = fmt.Errorf("Error: INPUT_REGISTERS is Read-Only..!!")

	case HOLDING_REGISTERS:
		if modbusCommandInfo.Length == 1 {
			result, err = c.client.WriteSingleRegister(uint16(modbusCommandInfo.StartingAddress), binary.BigEndian.Uint16(value))
		} else {
			result, err = c.client.WriteMultipleRegisters(uint16(modbusCommandInfo.StartingAddress), modbusCommandInfo.Length, value)
		}
	default:
	}

	if err != nil {
		return err
	}
	driver.Logger.Tracef("Modbus client SetValue successful, results: %v", result)

	return nil
}

func NewDeviceClient(connectionInfo *ConnectionInfo) (*ModbusClient, error) {
	client := new(ModbusClient)
	client.ModbusType = connectionInfo.Protocol
	switch client.ModbusType {
	case ProtocolTCP:
		client.TCPClientHandler.Address = fmt.Sprintf("%s:%d", connectionInfo.Address, connectionInfo.Port)
		client.TCPClientHandler.SlaveId = byte(connectionInfo.UnitID)
		client.TCPClientHandler.Timeout = time.Duration(connectionInfo.Timeout) * time.Second
		client.TCPClientHandler.IdleTimeout = time.Duration(connectionInfo.IdleTimeout) * time.Second
		client.TCPClientHandler.Logger = log.New(os.Stdout, "", log.LstdFlags)
	case ProtocolRTU:
		serialParams := strings.Split(connectionInfo.Address, ",")
		client.RTUClientHandler.Address = serialParams[0]
		client.RTUClientHandler.SlaveId = byte(connectionInfo.UnitID)
		client.RTUClientHandler.Timeout = time.Duration(connectionInfo.Timeout) * time.Second
		client.RTUClientHandler.IdleTimeout = time.Duration(connectionInfo.IdleTimeout) * time.Second
		client.RTUClientHandler.BaudRate = connectionInfo.BaudRate
		client.RTUClientHandler.DataBits = connectionInfo.DataBits
		client.RTUClientHandler.StopBits = connectionInfo.StopBits
		client.RTUClientHandler.Parity = connectionInfo.Parity
		client.RTUClientHandler.Logger = log.New(os.Stdout, "", log.LstdFlags)
	case ProtocolASCII:
		serialParams := strings.Split(connectionInfo.Address, ",")
		client.ASCIIClientHandler.Address = serialParams[0]
		client.ASCIIClientHandler.SlaveId = byte(connectionInfo.UnitID)
		client.ASCIIClientHandler.Timeout = time.Duration(connectionInfo.Timeout) * time.Second
		client.ASCIIClientHandler.IdleTimeout = time.Duration(connectionInfo.IdleTimeout) * time.Second
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
		return nil, err
	}
	return client, nil
}
