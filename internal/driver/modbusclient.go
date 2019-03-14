// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018-2019 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"fmt"
	"log"
	"os"
	"strings"

	MODBUS "github.com/goburrow/modbus"
)

// ModbusClient is used for connecting the device and read/write value
type ModbusClient struct {
	// IsModbusTcp is a value indicating the connection type
	IsModbusTcp bool
	// TCPClientHandler is ued for holding device TCP connection
	TCPClientHandler MODBUS.TCPClientHandler
	// TCPClientHandler is ued for holding device RTU connection
	RTUClientHandler MODBUS.RTUClientHandler

	client MODBUS.Client
}

func (c *ModbusClient) OpenConnection() error {
	var err error
	var newClient MODBUS.Client
	if c.IsModbusTcp {
		err = c.TCPClientHandler.Connect()
		newClient = MODBUS.NewClient(&c.TCPClientHandler)
		driver.Logger.Info(fmt.Sprintf("Modbus client create TCP connection."))
	} else {
		err = c.RTUClientHandler.Connect()
		newClient = MODBUS.NewClient(&c.RTUClientHandler)
		driver.Logger.Info(fmt.Sprintf("Modbus client create RTU connection."))
	}
	c.client = newClient
	return err
}

func (c *ModbusClient) CloseConnection() error {
	var err error
	if c.IsModbusTcp {
		err = c.TCPClientHandler.Close()

	} else {
		err = c.RTUClientHandler.Close()
	}
	return err
}

func (c *ModbusClient) GetValue(commandInfo interface{}) ([]byte, error) {
	var modbusCommandInfo = commandInfo.(*CommandInfo)

	// Reading value from device
	var response []byte
	var err error

	switch modbusCommandInfo.PrimaryTable {
	case DISCRETES_INPUT:
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

	driver.Logger.Info(fmt.Sprintf("Modbus client GetValue's results %v", response))

	return response, nil
}

func (c *ModbusClient) SetValue(commandInfo interface{}, value []byte) error {
	var modbusCommandInfo = commandInfo.(*CommandInfo)

	// Write value to device
	var result []byte
	var err error

	switch modbusCommandInfo.PrimaryTable {
	case DISCRETES_INPUT:
		result, err = c.client.WriteMultipleCoils(uint16(modbusCommandInfo.StartingAddress), modbusCommandInfo.Length, value)

	case COILS:
		result, err = c.client.WriteMultipleCoils(uint16(modbusCommandInfo.StartingAddress), modbusCommandInfo.Length, value)

	case INPUT_REGISTERS:
		result, err = c.client.WriteMultipleRegisters(uint16(modbusCommandInfo.StartingAddress), modbusCommandInfo.Length, value)

	case HOLDING_REGISTERS:
		result, err = c.client.WriteMultipleRegisters(uint16(modbusCommandInfo.StartingAddress), modbusCommandInfo.Length, value)
	default:
	}

	if err != nil {
		return err
	}
	driver.Logger.Info(fmt.Sprintf("Modbus client SetValue successful, results: %v", result))

	return nil
}

func NewDeviceClient(connectionInfo *ConnectionInfo) (*ModbusClient, error) {
	client := new(ModbusClient)
	var err error
	var tcpClientHandler = new(MODBUS.TCPClientHandler)
	var rtuClientHandler = new(MODBUS.RTUClientHandler)
	if connectionInfo.Protocol == ProtocolTCP {
		client.IsModbusTcp = true
	}
	if client.IsModbusTcp {
		tcpClientHandler = MODBUS.NewTCPClientHandler(fmt.Sprintf("%s:%d", connectionInfo.Address, connectionInfo.Port))
		tcpClientHandler.SlaveId = byte(connectionInfo.UnitID)
		tcpClientHandler.IdleTimeout = 0
		tcpClientHandler.Logger = log.New(os.Stdout, "", log.LstdFlags)
	} else {
		serialParams := strings.Split(connectionInfo.Address, ",")
		rtuClientHandler = MODBUS.NewRTUClientHandler(serialParams[0])
		rtuClientHandler.SlaveId = byte(connectionInfo.UnitID)
		rtuClientHandler.IdleTimeout = 0
		rtuClientHandler.BaudRate = connectionInfo.BaudRate
		rtuClientHandler.DataBits = connectionInfo.DataBits
		rtuClientHandler.StopBits = connectionInfo.StopBits
		rtuClientHandler.Parity = connectionInfo.Parity
		rtuClientHandler.Logger = log.New(os.Stdout, "", log.LstdFlags)
	}

	client.TCPClientHandler = *tcpClientHandler
	client.RTUClientHandler = *rtuClientHandler
	return client, err
}
