// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2019-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"fmt"
	"strconv"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/models"
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
	// Connect & Read timeout(seconds)
	Timeout int
	// Idle timeout(seconds) to close the connection
	IdleTimeout int
}

func createConnectionInfo(protocols map[string]models.ProtocolProperties) (info *ConnectionInfo, err error) {
	protocolRTU, rtuExist := protocols[ProtocolRTU]
	protocolTCP, tcpExist := protocols[ProtocolTCP]

	if rtuExist && tcpExist {
		return info, fmt.Errorf("unsupported multiple protocols, please choose %s or %s, not both", ProtocolRTU, ProtocolTCP)
	} else if !rtuExist && !tcpExist {
		return info, fmt.Errorf("unable to create connection info, protocol config '%s' or %s not exist", ProtocolRTU, ProtocolTCP)
	}

	if rtuExist {
		info, err = createRTUConnectionInfo(protocolRTU)
		if err != nil {
			return nil, err
		}
	} else if tcpExist {
		info, err = createTcpConnectionInfo(protocolTCP)
		if err != nil {
			return nil, err
		}
	}

	return info, nil
}

func parseIntValue(properties map[string]string, key string) (int, error) {
	str, ok := properties[key]
	if !ok {
		return 0, fmt.Errorf("protocol config '%s' not exist", key)
	}
	val, err := strconv.Atoi(str)
	if err != nil {
		return 0, fmt.Errorf("fail to parse protocol config '%s', %v", key, err)
	}
	return val, nil
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

	baudRate, err := parseIntValue(rtuProtocol, BaudRate)
	if err != nil {
		return nil, err
	}

	dataBits, err := parseIntValue(rtuProtocol, DataBits)
	if err != nil {
		return nil, err
	}

	stopBits, err := parseIntValue(rtuProtocol, StopBits)
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

	timeout, err := parseIntValue(rtuProtocol, Timeout)
	if err != nil {
		return nil, err
	}

	idleTimeout, err := parseIntValue(rtuProtocol, IdleTimeout)
	if err != nil {
		return nil, err
	}

	return &ConnectionInfo{
		Protocol:    ProtocolRTU,
		Address:     address,
		BaudRate:    baudRate,
		DataBits:    dataBits,
		StopBits:    stopBits,
		Parity:      parity,
		UnitID:      byte(unitID),
		Timeout:     timeout,
		IdleTimeout: idleTimeout,
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

	timeout, err := parseIntValue(tcpProtocol, Timeout)
	if err != nil {
		return nil, err
	}

	idleTimeout, err := parseIntValue(tcpProtocol, IdleTimeout)
	if err != nil {
		return nil, err
	}

	return &ConnectionInfo{
		Protocol:    ProtocolTCP,
		Address:     address,
		Port:        int(port),
		UnitID:      byte(unitID),
		Timeout:     timeout,
		IdleTimeout: idleTimeout,
	}, nil
}
