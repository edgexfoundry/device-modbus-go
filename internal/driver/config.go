// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2019-2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"fmt"
	"strconv"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"

	"github.com/spf13/cast"
)

// ConnectionInfo is device connection info
type ConnectionInfo struct {
	Protocol                string
	Address                 string
	Port                    int
	BaudRate                int
	DataBits                int
	StopBits                int
	Parity                  string
	UnitID                  uint8
	Timeout                 time.Duration // Connect & Read / Write timeout
	IdleTimeout             time.Duration // Idle timeout to close the connection (use 0 to dial for each request and negative value to never close)
	LinkRecoveryTimeout     time.Duration // Recovery timeout if tcp communication misbehaves (TCP only)
	ProtocolRecoveryTimeout time.Duration // Recovery timeout if the protocol is malformed, e.g. wrong transaction ID (TCP only)
	ConnectDelay            time.Duration // Silent period after successful connection (TCP only)
}

func (info *ConnectionInfo) String() string {
	if info.Protocol == ProtocolTCP {
		return fmt.Sprintf("%s:%s:%d:%d", info.Protocol, info.Address, info.Port, info.UnitID)
	}
	return fmt.Sprintf("%s:%s:%d:%d:%d:%d:%d:%s", info.Protocol, info.Address, info.Port, info.UnitID, info.BaudRate, info.DataBits, info.StopBits, info.Parity)
}

func createConnectionInfo(protocols map[string]models.ProtocolProperties) (info *ConnectionInfo, err error) {
	protocolRTU, rtuExist := protocols[ProtocolRTU]
	protocolASCII, asciiExist := protocols[ProtocolASCII]
	protocolTCP, tcpExist := protocols[ProtocolTCP]

	if rtuExist && tcpExist || rtuExist && asciiExist || tcpExist && asciiExist || rtuExist && tcpExist && asciiExist {
		driver.Logger.Errorf("unsupported multiple protocols,protocols: %+v", protocols)
		return info, fmt.Errorf("unsupported multiple protocols, protocols:%+v", protocols)
	} else if !rtuExist && !tcpExist && !asciiExist {
		driver.Logger.Errorf("unable to create connection info,protocols: %+v", protocols)
		return info, fmt.Errorf("unable to create connection info, protocol config '%s' or %s not exist", ProtocolRTU, ProtocolTCP)
	}

	if rtuExist {
		info, err = createRTUConnectionInfo(protocolRTU)
		if err != nil {
			driver.Logger.Errorf("protocolRTU %+v create RTUConnectionInfo failed", protocolRTU)
			return nil, err
		}
	} else if asciiExist {
		info, err = createASCIIConnectionInfo(protocolASCII)
		if err != nil {
			driver.Logger.Errorf("protocolASCII %+v create ASCIIConnectionInfo failed", protocolASCII)
			return nil, err
		}
	} else if tcpExist {
		info, err = createTcpConnectionInfo(protocolTCP)
		if err != nil {
			driver.Logger.Errorf("protocolTCP %+v create TCPConnectionInfo failed", protocolTCP)
			return nil, err
		}
	}

	return info, nil
}

func parseIntValue(properties map[string]any, key string) (int, error) {
	value, ok := properties[key]
	if !ok {
		return 0, fmt.Errorf("protocol config '%s' not exist", key)
	}

	res, err := cast.ToIntE(value)
	if err != nil {
		return 0, fmt.Errorf("cannot transfrom protocol config %s value %v to int, %v", key, value, err)
	}

	return res, nil
}

func parseFloatValue(properties map[string]any, key string) (float64, error) {
	value, ok := properties[key]
	if !ok {
		return 0, fmt.Errorf("protocol config '%s' not exist", key)
	}

	res, err := cast.ToFloat64E(value)
	if err != nil {
		return 0, fmt.Errorf("cannot transfrom protocol config %s value %v to float64, %v", key, value, err)
	}

	return res, nil
}

func parseDurationValue(properties map[string]any, key string) (time.Duration, error) {
	value, ok := properties[key]
	if !ok {
		return 0, fmt.Errorf("protocol config '%s' not exist", key)
	}
	// for straight number value, treat it as seconds
	if v, err := parseFloatValue(properties, key); err == nil {
		return time.Duration(v) * time.Second, nil
	}

	res, err := cast.ToDurationE(value)
	if err != nil {
		return 0, fmt.Errorf("cannot transfrom protocol config %s value %v to time.Duration, %v", key, value, err)
	}

	return res, nil
}

func createSerialConnectionInfo(protocol map[string]any) (info *ConnectionInfo, err error) {
	errorMessage := "unable to create RTU connection info, protocol config '%s' not exist"
	value, ok := protocol[Address]
	if !ok {
		return nil, fmt.Errorf(errorMessage, Address)
	}
	address, ok := value.(string)
	if !ok {
		return nil, fmt.Errorf("cannot transform '%s' value %v to string", Address, value)
	}

	us, ok := protocol[UnitID]
	if !ok {
		return nil, fmt.Errorf(errorMessage, UnitID)
	}
	unitID, err := strconv.ParseUint(fmt.Sprintf("%v", us), 0, 8)
	if err != nil {
		return nil, fmt.Errorf("uintID value out of range(0–255). Error: %v", err)
	}

	baudRate, err := parseIntValue(protocol, BaudRate)
	if err != nil {
		return nil, err
	}

	dataBits, err := parseIntValue(protocol, DataBits)
	if err != nil {
		return nil, err
	}

	stopBits, err := parseIntValue(protocol, StopBits)
	if err != nil {
		return nil, err
	}

	value, ok = protocol[Parity]
	if !ok {
		return nil, fmt.Errorf(errorMessage, Parity)
	}
	parity, ok := value.(string)
	if !ok {
		return nil, fmt.Errorf("cannot transform '%s' value %v to string", Parity, parity)
	}
	if parity != "N" && parity != "O" && parity != "E" {
		return nil, fmt.Errorf("invalid parity value, it should be N(None) or O(Odd) or E(Even)")
	}

	timeout, err := parseDurationValue(protocol, Timeout)
	if err != nil {
		return nil, err
	}

	idleTimeout, err := parseDurationValue(protocol, IdleTimeout)
	if err != nil {
		return nil, err
	}

	return &ConnectionInfo{
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

func createRTUConnectionInfo(rtuProtocol map[string]any) (info *ConnectionInfo, err error) {
	rtuInfo, err := createSerialConnectionInfo(rtuProtocol)
	if err != nil {
		return nil, err
	}
	return &ConnectionInfo{
		Protocol:    ProtocolRTU,
		Address:     rtuInfo.Address,
		BaudRate:    rtuInfo.BaudRate,
		DataBits:    rtuInfo.DataBits,
		StopBits:    rtuInfo.StopBits,
		Parity:      rtuInfo.Parity,
		UnitID:      rtuInfo.UnitID,
		Timeout:     rtuInfo.Timeout,
		IdleTimeout: rtuInfo.IdleTimeout,
	}, nil
}

func createASCIIConnectionInfo(asciiProtocol map[string]any) (info *ConnectionInfo, err error) {
	asciiInfo, err := createSerialConnectionInfo(asciiProtocol)
	if err != nil {
		return nil, err
	}
	return &ConnectionInfo{
		Protocol:    ProtocolASCII,
		Address:     asciiInfo.Address,
		BaudRate:    asciiInfo.BaudRate,
		DataBits:    asciiInfo.DataBits,
		StopBits:    asciiInfo.StopBits,
		Parity:      asciiInfo.Parity,
		UnitID:      asciiInfo.UnitID,
		Timeout:     asciiInfo.Timeout,
		IdleTimeout: asciiInfo.IdleTimeout,
	}, nil
}

func createTcpConnectionInfo(tcpProtocol map[string]any) (info *ConnectionInfo, err error) {
	errorMessage := "unable to create TCP connection info, protocol config '%s' not exist"
	value, ok := tcpProtocol[Address]
	if !ok {
		return nil, fmt.Errorf(errorMessage, Address)
	}
	address, ok := value.(string)
	if !ok {
		return nil, fmt.Errorf("cannot transform '%s' value %v to string", Address, value)
	}

	value, ok = tcpProtocol[Port]
	if !ok {
		return nil, fmt.Errorf(errorMessage, Port)
	}
	port, err := strconv.ParseUint(fmt.Sprintf("%v", value), 0, 16)
	if err != nil {
		return nil, fmt.Errorf("port value out of range(0–65535). Error: %v", err)
	}

	value, ok = tcpProtocol[UnitID]
	if !ok {
		return nil, fmt.Errorf(errorMessage, UnitID)
	}
	unitID, err := strconv.ParseUint(fmt.Sprintf("%v", value), 0, 8)
	if err != nil {
		return nil, fmt.Errorf("uintID value out of range(0–255). Error: %v", err)
	}

	timeout, err := parseDurationValue(tcpProtocol, Timeout)
	if err != nil {
		return nil, err
	}

	idleTimeout, err := parseDurationValue(tcpProtocol, IdleTimeout)
	if err != nil {
		idleTimeout = 0 // default = dial for each request
	}

	protocolRecoveryTimeout, err := parseDurationValue(tcpProtocol, ProtocolRecoveryTimeout)
	if err != nil {
		protocolRecoveryTimeout = 0 // default = no protocol recovery
	}

	linkRecoveryTimeout, err := parseDurationValue(tcpProtocol, LinkRecoveryTimeout)
	if err != nil {
		linkRecoveryTimeout = 0 // default = no link recovery
	}

	connectDelay, err := parseDurationValue(tcpProtocol, ConnectDelay)
	if err != nil {
		connectDelay = 0 // default no connect delay
	}

	return &ConnectionInfo{
		Protocol:                ProtocolTCP,
		Address:                 address,
		Port:                    int(port),
		UnitID:                  byte(unitID),
		Timeout:                 timeout,
		IdleTimeout:             idleTimeout,
		ProtocolRecoveryTimeout: protocolRecoveryTimeout,
		LinkRecoveryTimeout:     linkRecoveryTimeout,
		ConnectDelay:            connectDelay,
	}, nil
}
