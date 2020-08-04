// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package driver

func swap32BitDataBytes(dataBytes []byte, isByteSwap bool, isWordSwap bool) []byte {

	if !isByteSwap && !isWordSwap {
		return dataBytes
	}

	if len(dataBytes) < 4 {
		return dataBytes
	}

	var newDataBytes = make([]byte, len(dataBytes))

	if isByteSwap {
		newDataBytes[0] = dataBytes[1]
		newDataBytes[1] = dataBytes[0]
		newDataBytes[2] = dataBytes[3]
		newDataBytes[3] = dataBytes[2]
		// Copy new DataBytes to the original DataBytes which can combine the ByteSwap with the WordSwap operation.
		copy(dataBytes, newDataBytes)
	}
	if isWordSwap {
		newDataBytes[0] = dataBytes[2]
		newDataBytes[1] = dataBytes[3]
		newDataBytes[2] = dataBytes[0]
		newDataBytes[3] = dataBytes[1]
	}

	return newDataBytes
}
