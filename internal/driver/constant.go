// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2"
)

const (
	BOOL = "BOOL"

	INT16 = "INT16"
	INT32 = "INT32"
	INT64 = "INT64"

	UINT16 = "UINT16"
	UINT32 = "UINT32"
	UINT64 = "UINT64"

	FLOAT32 = "FLOAT32"
	FLOAT64 = "FLOAT64"

	DISCRETES_INPUT   = "DISCRETES_INPUT"
	COILS             = "COILS"
	INPUT_REGISTERS   = "INPUT_REGISTERS"
	HOLDING_REGISTERS = "HOLDING_REGISTERS"

	PRIMARY_TABLE    = "primaryTable"
	STARTING_ADDRESS = "startingAddress"
	IS_BYTE_SWAP     = "isByteSwap"
	IS_WORD_SWAP     = "isWordSwap"
	// RAW_TYPE define binary data type which read from Modbus device
	RAW_TYPE = "rawType"

	SERVICE_STOP_WAIT_TIME = 1
)

var PrimaryTableBitCountMap = map[string]uint16{
	DISCRETES_INPUT:   1,
	COILS:             1,
	INPUT_REGISTERS:   16,
	HOLDING_REGISTERS: 16,
}

var ValueTypeBitCountMap = map[string]uint16{
	v2.ValueTypeInt16: 16,
	v2.ValueTypeInt32: 32,
	v2.ValueTypeInt64: 64,

	v2.ValueTypeUint16: 16,
	v2.ValueTypeUint32: 32,
	v2.ValueTypeUint64: 64,

	v2.ValueTypeFloat32: 32,
	v2.ValueTypeFloat64: 64,

	v2.ValueTypeBool: 1,
}
