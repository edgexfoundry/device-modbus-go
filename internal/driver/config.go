// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2019 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"fmt"
	"strconv"

	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

// ConnectionInfo is device connection info
type ConnectionInfo struct {
	Protocol string
	Address  string
	Port     int
	BaudRate int
	DataBits int
	StopBits int
	Parity   string
	UnitID   uint8
}

func createConnectionInfo(protocols map[string]models.ProtocolProperties) (info *ConnectionInfo, err error) {
	protocolRTU, rtuExist := protocols[ProtocolRTU]
	protocolASCII, asciiExist := protocols[ProtocolASCII]
	protocolTCP, tcpExist := protocols[ProtocolTCP]

	if rtuExist && !asciiExist && !tcpExist {
		fmt.Printf("configs modbus rtu mode, now create connection")
		info, err = createRTUConnectionInfo(protocolRTU)
		if err != nil {
			return nil, err
		}
	} else if !rtuExist && asciiExist && !tcpExist {
		fmt.Printf("configs modbus ascii mode, now create connection")
		info, err = createASCIIConnectionInfo(protocolASCII)
		if err != nil {
			return nil, err
		}
	} else if !rtuExist && !asciiExist && tcpExist {
		fmt.Printf("configs modbus tcp mode, now create connection")
		info, err = createTcpConnectionInfo(protocolTCP)
		if err != nil {
			return nil, err
		}
	} else {
		return info, fmt.Errorf("unsupported modbus protocol, please config one of %s, %s or %s mode", ProtocolRTU, ProtocolASCII, ProtocolTCP)
	}

	return info, nil
}

func createRTUConnectionInfo(rtuProtocol map[string]string) (info *ConnectionInfo, err error) {
	errorMessage := "unable to create RTU connection info, protocol config '%s' not exist"
	address, ok := rtuProtocol[Address]
	if !ok {
		return nil, fmt.Errorf(errorMessage, Address)
	}

	us, ok := rtuProtocol[UnitID]
	if !ok {
		return nil, fmt.Errorf(errorMessage, UnitID)
	}
	unitID, err := strconv.ParseUint(us, 0, 8)
	if err != nil {
		return nil, fmt.Errorf("uintID value out of range(0–255). Error: %v", err)
	}

	bs, ok := rtuProtocol[BaudRate]
	if !ok {
		return nil, fmt.Errorf(errorMessage, BaudRate)
	}
	baudRate, err := strconv.Atoi(bs)
	if err != nil {
		return nil, err
	}

	ds, ok := rtuProtocol[DataBits]
	if !ok {
		return nil, fmt.Errorf(errorMessage, DataBits)
	}
	dataBits, err := strconv.Atoi(ds)
	if err != nil {
		return nil, err
	}

	ss, ok := rtuProtocol[StopBits]
	if !ok {
		return nil, fmt.Errorf(errorMessage, StopBits)
	}
	stopBits, err := strconv.Atoi(ss)
	if err != nil {
		return nil, err
	}

	parity, ok := rtuProtocol[Parity]
	if !ok {
		return nil, fmt.Errorf(errorMessage, Parity)
	}
	if parity != "N" && parity != "O" && parity != "E" {
		return nil, fmt.Errorf("invalid parity value, it should be N(None) or O(Odd) or E(Even)")
	}

	return &ConnectionInfo{
		Protocol: ProtocolRTU,
		Address:  address,
		BaudRate: baudRate,
		DataBits: dataBits,
		StopBits: stopBits,
		Parity:   parity,
		UnitID:   byte(unitID),
	}, nil
}

func createASCIIConnectionInfo(asciiProtocol map[string]string) (info *ConnectionInfo, err error) {
	errorMessage := "unable to create ASCII connection info, protocol config '%s' not exist"
	address, ok := asciiProtocol[Address]
	if !ok {
		return nil, fmt.Errorf(errorMessage, Address)
	}

	us, ok := asciiProtocol[UnitID]
	if !ok {
		return nil, fmt.Errorf(errorMessage, UnitID)
	}
	unitID, err := strconv.ParseUint(us, 0, 8)
	if err != nil {
		return nil, fmt.Errorf("unitID value out of range(0-255). Error: %v", err)
	}

	br, ok := asciiProtocol[BaudRate]
	if !ok {
		return nil, fmt.Errorf(errorMessage, BaudRate)
	}
	baudRate, err := strconv.Atoi(br)
	if err != nil {
		return nil, err
	}
	ds, ok := asciiProtocol[DataBits]
	if !ok {
		return nil, fmt.Errorf(errorMessage, DataBits)
	}
	dataBits, err := strconv.Atoi(ds)
	if err != nil {
		return nil, err
	}

	ss, ok := asciiProtocol[StopBits]
	if !ok {
		return nil, fmt.Errorf(errorMessage, StopBits)
	}
	stopBits, err := strconv.Atoi(ss)
	if err != nil {
		return nil, err
	}

	parity, ok := asciiProtocol[Parity]
	if !ok {
		return nil, fmt.Errorf(errorMessage, Parity)
	}
	if parity != "N" && parity != "O" && parity != "E" {
		return nil, fmt.Errorf("invalid parity value, it should be N(None) or O(Odd) or E(Even)")
	}

	return &ConnectionInfo{
		Protocol: ProtocolASCII,
		Address:  address,
		BaudRate: baudRate,
		DataBits: dataBits,
		StopBits: stopBits,
		Parity:   parity,
		UnitID:   byte(unitID),
	}, nil
}

func createTcpConnectionInfo(tcpProtocol map[string]string) (info *ConnectionInfo, err error) {
	errorMessage := "unable to create TCP connection info, protocol config '%s' not exist"
	address, ok := tcpProtocol[Address]
	if !ok {
		return nil, fmt.Errorf(errorMessage, Address)
	}

	portString, ok := tcpProtocol[Port]
	if !ok {
		return nil, fmt.Errorf(errorMessage, Port)
	}
	port, err := strconv.ParseUint(portString, 0, 16)
	if err != nil {
		return nil, fmt.Errorf("port value out of range(0–65535). Error: %v", err)
	}

	unitIDString, ok := tcpProtocol[UnitID]
	if !ok {
		return nil, fmt.Errorf(errorMessage, UnitID)
	}
	unitID, err := strconv.ParseUint(unitIDString, 0, 8)
	if err != nil {
		return nil, fmt.Errorf("uintID value out of range(0–255). Error: %v", err)
	}

	return &ConnectionInfo{
		Protocol: ProtocolTCP,
		Address:  address,
		Port:     int(port),
		UnitID:   byte(unitID),
	}, nil
}
