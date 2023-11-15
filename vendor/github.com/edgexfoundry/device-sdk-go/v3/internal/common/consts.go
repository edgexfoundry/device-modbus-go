// -*- mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018-2022 IOTech Ltd
// Copyright (c) 2019 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package common

const (
	URLRawQuery       = "urlRawQuery"
	SDKReservedPrefix = "ds-"
)

// SDKVersion indicates the version of the SDK - will be overwritten by build
var SDKVersion string = "0.0.0"

// ServiceVersion indicates the version of the device service itself, not the SDK - will be overwritten by build
var ServiceVersion string = "0.0.0"
