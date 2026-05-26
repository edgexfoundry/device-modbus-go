// -*- mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018-2022 IOTech Ltd
// Copyright (c) 2019 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package common

import "github.com/edgexfoundry/go-mod-core-contracts/v4/common"

const (
	URLRawQuery       = "urlRawQuery"
	SDKReservedPrefix = "ds-"
)

// SDKVersion indicates the version of the SDK - will be overwritten by build
var SDKVersion string = "0.0.0"

// ServiceVersion indicates the version of the device service itself, not the SDK - will be overwritten by build
var ServiceVersion string = "0.0.0"

// contextKey is a custom type used to represent the context header key, rather than using a plain string directly.
// We define contextKey type here to avoid SA1029 lint error as detailed in https://staticcheck.dev/docs/checks#SA1029
type contextKey string

// CorrelationHeaderKey is a constant key used to represent the correlation header.
// It is used instead of CorrelationHeader (from go-mod-core-contracts) to avoid the SA1029 lint warning.
// For example, the following code would trigger an SA1029 lint error:
// ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.NewString())
// To prevent this, use CorrelationHeaderKey instead:
// ctx := context.WithValue(context.Background(), common.CorrelationHeaderKey, uuid.NewString())
const CorrelationHeaderKey = contextKey(common.CorrelationHeader)
