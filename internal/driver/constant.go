// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package driver

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
)

var PrimaryTableBitCountMap = map[string]uint16{
	DISCRETES_INPUT:   1,
	COILS:             1,
	INPUT_REGISTERS:   16,
	HOLDING_REGISTERS: 16,
}

var ValueTypeBitCountMap = map[string]uint16{
	INT16: 16,
	INT32: 32,
	INT64: 64,

	UINT16: 16,
	UINT32: 32,
	UINT64: 64,

	FLOAT32: 32,
	FLOAT64: 64,

	BOOL: 1,
}
