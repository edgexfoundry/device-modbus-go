// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2019-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package driver

const (
	ProtocolTCP = "modbus-tcp"
	ProtocolRTU = "modbus-rtu"

	Address  = "Address"
	Port     = "Port"
	UnitID   = "UnitID"
	BaudRate = "BaudRate"
	DataBits = "DataBits"
	StopBits = "StopBits"
	// Parity: N - None, O - Odd, E - Even
	Parity = "Parity"

	Timeout     = "Timeout"
	IdleTimeout = "IdleTimeout"
)
