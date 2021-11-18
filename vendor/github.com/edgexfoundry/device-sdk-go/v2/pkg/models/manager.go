//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package models

type AutoEventManager interface {
	// StartAutoEvents starts all the AutoEvents of the device service
	StartAutoEvents()
	// RestartForDevice restarts all the AutoEvents of the specific device
	RestartForDevice(name string)
	// StopForDevice stops all the AutoEvents of the specific device
	StopForDevice(name string)
}
