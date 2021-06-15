// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"fmt"
	"strings"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
)

func castStartingAddress(i interface{}) (res uint16, err errors.EdgeX) {
	switch v := i.(type) {
	case float64:
		res = uint16(v)
	case float32:
		res = uint16(v)
	case int64:
		res = uint16(v)
	case int32:
		res = uint16(v)
	case int16:
		res = uint16(v)
	case int:
		res = uint16(v)
	case uint64:
		res = uint16(v)
	case uint32:
		res = uint16(v)
	case uint16:
		res = v
	default:
		return 0, errors.NewCommonEdgeX(errors.KindContractInvalid, "startingAddress should be integer value", nil)
	}
	return res, nil
}

func normalizeRawType(rawType string) (normalized string, err errors.EdgeX) {
	switch strings.ToUpper(rawType) {
	case UINT16:
		normalized = common.ValueTypeUint16
	case INT16:
		normalized = common.ValueTypeInt16
	default:
		return "", errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("the raw type %s is not supported", rawType), nil)
	}
	return normalized, err
}

func castSwapAttribute(i interface{}) (res bool, err errors.EdgeX) {
	switch v := i.(type) {
	case bool:
		res = v
	default:
		return res, errors.NewCommonEdgeX(errors.KindContractInvalid, "swap attribute should be boolean value", nil)
	}
	return res, nil
}
