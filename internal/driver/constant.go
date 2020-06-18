// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018-2019 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package driver

import sdkModel "github.com/edgexfoundry/device-sdk-go/pkg/models"

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

var ValueTypeBitCountMap = map[sdkModel.ValueType]uint16{
	sdkModel.Int16: 16,
	sdkModel.Int32: 32,
	sdkModel.Int64: 64,

	sdkModel.Uint16: 16,
	sdkModel.Uint32: 32,
	sdkModel.Uint64: 64,

	sdkModel.Float32: 32,
	sdkModel.Float64: 64,

	sdkModel.Bool: 1,
}
