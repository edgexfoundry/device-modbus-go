// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"strings"
	"testing"

	sdkModel "github.com/edgexfoundry/device-sdk-go/pkg/models"
)

func TestCheckAttributes(t *testing.T) {
	requests := []sdkModel.CommandRequest{
		{
			Attributes: map[string]string{
				"primaryTable":    "HOLDING_REGISTERS",
				"startingAddress": "1001",
				"isByteSwap":      "true",
			},
		},
		{
			Attributes: map[string]string{
				"primaryTable":    "HOLDING_REGISTERS",
				"startingAddress": "1002",
				"isByteSwap":      "true",
			},
		},
	}

	err := checkAttributes(requests)
	if err != nil {
		t.Fatalf("Test check attributes failed! Error: %v", err)
	}
}

func TestCheckAttributes_fail(t *testing.T) {
	requests := []sdkModel.CommandRequest{
		{
			Attributes: map[string]string{
				"primaryTable":    "HOLDING_REGISTERS",
				"startingAddress": "1001",
				"isByteSwap":      "true",
			},
		},
		{
			Attributes: map[string]string{
				"primaryTable":    "HOLDING_REGISTERS",
				"startingAddress": "test-1002",
				"isByteSwap":      "true",
			},
		},
	}

	err := checkAttributes(requests)
	if err == nil || !strings.Contains(err.Error(), "startingAddress fail to convert interface inoto unit16") {
		t.Fatalf("Test should be failed!")
	}
}
