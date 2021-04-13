// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"strings"
	"testing"

	"github.com/edgexfoundry/device-sdk-go/v2/pkg/models"
)

func TestCheckAttributes(t *testing.T) {
	requests := []models.CommandRequest{
		{
			Attributes: map[string]string{
				PRIMARY_TABLE:    HOLDING_REGISTERS,
				STARTING_ADDRESS: "1001",
				IS_BYTE_SWAP:     "true",
			},
		},
		{
			Attributes: map[string]string{
				PRIMARY_TABLE:    HOLDING_REGISTERS,
				STARTING_ADDRESS: "1002",
				IS_BYTE_SWAP:     "true",
			},
		},
	}

	err := checkAttributes(requests)
	if err != nil {
		t.Fatalf("Test check attributes failed! Error: %v", err)
	}
}

func TestCheckAttributes_fail(t *testing.T) {
	requests := []models.CommandRequest{
		{
			Attributes: map[string]string{
				PRIMARY_TABLE:    HOLDING_REGISTERS,
				STARTING_ADDRESS: "1001",
				IS_BYTE_SWAP:     "true",
			},
		},
		{
			Attributes: map[string]string{
				PRIMARY_TABLE:    HOLDING_REGISTERS,
				STARTING_ADDRESS: "test-1002",
				IS_BYTE_SWAP:     "true",
			},
		},
	}

	err := checkAttributes(requests)
	if err == nil || !strings.Contains(err.Error(), "startingAddress fail to convert interface inoto unit16") {
		t.Fatalf("Test should be failed!")
	}
}
