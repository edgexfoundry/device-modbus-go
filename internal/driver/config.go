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
	//flag to indicate 0 or 1 based addressing
	ZeroBase bool
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

	var zeroBase = false
	zero, hasKey := rtuProtocol[ZeroBase]

	if hasKey {
		value, err := strconv.ParseBool(zero)
		if err != nil {
			return nil, fmt.Errorf("zeroBase value should be true or false. Error: %v", err)
		}
		zeroBase = value

	}
	return &ConnectionInfo{
		Protocol: ProtocolRTU,
		Address:  address,
		BaudRate: baudRate,
		DataBits: dataBits,
		StopBits: stopBits,
		Parity:   parity,
		UnitID:   byte(unitID),
		ZeroBase: zeroBase,
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

	var zeroBase = false
	zero, hasKey := tcpProtocol[ZeroBase]

	if hasKey {
		value, err := strconv.ParseBool(zero)
		if err != nil {
      return nil, fmt.Errorf("zeroBase value should be true or false. Error: %v", err)
		}
		zeroBase = value

	}
	return &ConnectionInfo{
		Protocol: ProtocolTCP,
		Address:  address,
		Port:     int(port),
		UnitID:   byte(unitID),
		ZeroBase: zeroBase,
	}, nil
}
