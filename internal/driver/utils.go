// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2021 IOTech Ltd
// Copyright (C) 2022 Schneider Electric
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"fmt"
	"strings"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/spf13/cast"
)

func castStartingAddress(i interface{}) (uint16, errors.EdgeX) {
	res, err := cast.ToUint16E(i)
	if err != nil {
		return 0, errors.NewCommonEdgeX(errors.KindContractInvalid, "startingAddress should be castable to an integer value", err)
	}

	return res, nil
}

func normalizeRawType(rawType string) (normalized string, err errors.EdgeX) {
	switch strings.ToUpper(rawType) {
	case UINT16:
		normalized = common.ValueTypeUint16
	case INT16:
		normalized = common.ValueTypeInt16
	case INT32:
		normalized = common.ValueTypeInt32
	default:
		return "", errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("the raw type %s is not supported", rawType), nil)
	}
	return normalized, err
}

func castSwapAttribute(i interface{}) (bool, errors.EdgeX) {
	res, err := cast.ToBoolE(i)
	if err != nil {
		return res, errors.NewCommonEdgeX(errors.KindContractInvalid, "swap attribute should be castable to a boolean value", err)
	}

	return res, nil
}
