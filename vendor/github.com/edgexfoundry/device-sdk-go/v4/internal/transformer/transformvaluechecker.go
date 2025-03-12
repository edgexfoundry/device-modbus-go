// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2019-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package transformer

import (
	"math"
)

func checkTransformedValueInRange(origin interface{}, transformed float64) bool {
	inRange := false
	switch origin.(type) {
	case uint8:
		if transformed >= 0 && transformed <= math.MaxUint8 && math.Trunc(transformed) == transformed {
			inRange = true
		}
	case uint16:
		if transformed >= 0 && transformed <= math.MaxUint16 && math.Trunc(transformed) == transformed {
			inRange = true
		}
	case uint32:
		if transformed >= 0 && transformed <= math.MaxUint32 && math.Trunc(transformed) == transformed {
			inRange = true
		}
	case uint64:
		// if the variable isn't casted to float64, this statement will cause error on 32bit system
		maxiMum := float64(math.MaxUint64)
		if transformed >= 0 && transformed <= maxiMum && math.Trunc(transformed) == transformed {
			inRange = true
		}
	case int8:
		if transformed >= math.MinInt8 && transformed <= math.MaxInt8 && math.Trunc(transformed) == transformed {
			inRange = true
		}
	case int16:
		if transformed >= math.MinInt16 && transformed <= math.MaxInt16 && math.Trunc(transformed) == transformed {
			inRange = true
		}
	case int32:
		if transformed >= math.MinInt32 && transformed <= math.MaxInt32 && math.Trunc(transformed) == transformed {
			inRange = true
		}
	case int64:
		if transformed >= math.MinInt64 && transformed <= math.MaxInt64 && math.Trunc(transformed) == transformed {
			inRange = true
		}
	case float32:
		if !math.IsNaN(float64(transformed)) && math.Abs(transformed) <= math.MaxFloat32 {
			inRange = true
		}
	case float64:
		if !math.IsNaN(transformed) && !math.IsInf(transformed, 0) {
			inRange = true
		}
	default:
		return false
	}

	return inRange
}
