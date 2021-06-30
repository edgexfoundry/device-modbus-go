// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
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

	// STRING_REGISTER_SIZE  E.g. "abcd" need 4 bytes as is 2 registers(2 words), so STRING_REGISTER_SIZE=2
	STRING_REGISTER_SIZE   = "stringRegisterSize"
	SERVICE_STOP_WAIT_TIME = 1
)

var PrimaryTableBitCountMap = map[string]uint16{
	DISCRETES_INPUT:   1,
	COILS:             1,
	INPUT_REGISTERS:   16,
	HOLDING_REGISTERS: 16,
}

var ValueTypeBitCountMap = map[string]uint16{
	common.ValueTypeInt16: 16,
	common.ValueTypeInt32: 32,
	common.ValueTypeInt64: 64,

	common.ValueTypeUint16: 16,
	common.ValueTypeUint32: 32,
	common.ValueTypeUint64: 64,

	common.ValueTypeFloat32: 32,
	common.ValueTypeFloat64: 64,

	common.ValueTypeBool:   1,
	common.ValueTypeString: 16,
}
