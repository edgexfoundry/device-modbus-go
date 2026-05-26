// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018-2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package startup

import (
	"fmt"
	"os"

	"github.com/edgexfoundry/device-sdk-go/v4/pkg/interfaces"
	"github.com/edgexfoundry/device-sdk-go/v4/pkg/service"
)

func Bootstrap(serviceKey string, serviceVersion string, driver interfaces.ProtocolDriver) {
	deviceService, err := service.NewDeviceService(serviceKey, serviceVersion, driver)
	if err != nil {
		_, _ = fmt.Fprint(os.Stderr, err.Error())
		os.Exit(-1)
	}

	err = deviceService.Run()
	if err != nil {
		deviceService.LoggingClient().Errorf("Device Service %s", err.Error())
		os.Exit(-1)
	}

	os.Exit(0)
}
