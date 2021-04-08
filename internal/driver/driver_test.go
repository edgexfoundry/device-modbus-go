// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"strings"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
)

func init() {
	driver = new(Driver)
	driver.Logger = logger.NewMockClient()
}

func TestLockAddressWithAddressCountLimit(t *testing.T) {
	address := "/dev/USB0tty"
	driver.addressMap = make(map[string]chan bool)
	driver.workingAddressCount = make(map[string]int)
	driver.workingAddressCount[address] = concurrentCommandLimit

	err := driver.lockAddress(address)

	if err == nil || !strings.Contains(err.Error(), "High-frequency command execution") {
		t.Errorf("Unexpect result, it should return high-frequency error, %v", err)
	}
}

func TestLockAddressWithAddressCountUnderLimit(t *testing.T) {
	address := "/dev/USB0tty"
	driver.addressMap = make(map[string]chan bool)
	driver.workingAddressCount = make(map[string]int)
	driver.workingAddressCount[address] = concurrentCommandLimit - 1

	err := driver.lockAddress(address)

	if err != nil {
		t.Errorf("Unexpect result, address should be lock successfully, %v", err)
	}
}
