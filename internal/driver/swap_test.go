// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
)

func init() {
	driver = new(Driver)
	driver.Logger = logger.NewMockClient()
}

func TestByteSwap32BitDataBytes(t *testing.T) {
	var isByteSwap = true
	var isWordSwap = false
	var dataBytes = []byte{6, 9, 12, 15}

	var expectResult = swap32BitDataBytes(dataBytes, isByteSwap, isWordSwap)

	if expectResult[0] != 9 || expectResult[1] != 6 || expectResult[2] != 15 || expectResult[3] != 12 {
		t.Fatalf("Failed to swap bytes for 32 bits! ")
	}
}

func TestWordSwap32BitDataBytes(t *testing.T) {
	var isByteSwap = false
	var isWordSwap = true
	var dataBytes = []byte{6, 9, 12, 15}

	var expectResult = swap32BitDataBytes(dataBytes, isByteSwap, isWordSwap)

	if expectResult[0] != 12 || expectResult[1] != 15 || expectResult[2] != 6 || expectResult[3] != 9 {
		t.Fatalf("Failed to swap words for 32 bits! ")
	}
}

func TestWordAndBytesSwap32BitDataBytes(t *testing.T) {
	var isByteSwap = true
	var isWordSwap = true
	var dataBytes = []byte{6, 9, 12, 15}

	var expectResult = swap32BitDataBytes(dataBytes, isByteSwap, isWordSwap)

	if expectResult[0] != 15 || expectResult[1] != 12 || expectResult[2] != 9 || expectResult[3] != 6 {
		t.Fatalf("Failed to swap bytes and words for 32 bits!")
	}
}
