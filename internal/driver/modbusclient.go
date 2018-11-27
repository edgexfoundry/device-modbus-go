// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	MODBUS "github.com/goburrow/modbus"
	"strconv"
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
	isModbusTcp := false
	var tcpClientHandler = new(MODBUS.TCPClientHandler)
	var rtuClientHandler = new(MODBUS.RTUClientHandler)
	if strings.Contains(connectionInfo.Protocol, "TCP") || strings.Contains(connectionInfo.Protocol, "HTTP") {
		isModbusTcp = true
	}
	if isModbusTcp {
		tcpClientHandler = MODBUS.NewTCPClientHandler(fmt.Sprintf("%s:%d", connectionInfo.Address, connectionInfo.Port))
		tcpClientHandler.SlaveId = byte(connectionInfo.UnitID)
		tcpClientHandler.Timeout = 10 * time.Second
		tcpClientHandler.Logger = log.New(os.Stdout, "", log.LstdFlags)
	} else {
		serialParams := strings.Split(connectionInfo.Address, ",")
		rtuClientHandler = MODBUS.NewRTUClientHandler(serialParams[0])
		rtuClientHandler.SlaveId = byte(connectionInfo.UnitID)
		rtuClientHandler.Timeout = 10 * time.Second
		rtuClientHandler.Logger = log.New(os.Stdout, "", log.LstdFlags)

		baudRate, err := strconv.Atoi(serialParams[1])
		if err != nil {
			return client, err
		}
		rtuClientHandler.BaudRate = baudRate

		dataBits, err := strconv.Atoi(serialParams[2])
		if err != nil {
			return client, err
		}
		rtuClientHandler.DataBits = dataBits

		stopBits, err := strconv.Atoi(serialParams[3])
		if err != nil {
			return client, err
		}
		rtuClientHandler.StopBits = stopBits

		// Parity: N - None(0), O - Odd(1), E - Even(2)  default E
		parity := "E"
		if serialParams[4] == "0" {
			parity = "N"
		} else if serialParams[4] == "1" {
			parity = "O"
		}
		rtuClientHandler.Parity = parity

	}

	client.IsModbusTcp = isModbusTcp
	client.TCPClientHandler = *tcpClientHandler
	client.RTUClientHandler = *rtuClientHandler
	return client, err
}
