// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package driver

import "testing"

// Test byte, word swap
func TestByteSwap32BitDataBytes(t *testing.T) {
	var isByteSwap = true
	var isWordSwap = false
	var dataBytes = []byte{6, 9, 12, 15}

	var expectResult = swap32BitDataBytes(dataBytes, isByteSwap, isWordSwap)

	if expectResult[0] != 9 || expectResult[1] != 6 || expectResult[2] != 15 || expectResult[3] != 12 {
		t.Fatalf("Swap32BitDataBytes 32 bits failed! ")
	}
}

func TestWordSwap32BitDataBytes(t *testing.T) {
	var isByteSwap = false
	var isWordSwap = true
	var dataBytes = []byte{6, 9, 12, 15}

	var expectResult = swap32BitDataBytes(dataBytes, isByteSwap, isWordSwap)

	if expectResult[0] != 12 || expectResult[1] != 15 || expectResult[2] != 6 || expectResult[3] != 9 {
		t.Fatalf("Swap32BitDataBytes 32 bits failed! ")
	}
}
