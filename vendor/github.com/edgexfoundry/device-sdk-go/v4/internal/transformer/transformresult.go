// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2019-2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package transformer

import (
	"fmt"
	"math"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"

	sdkCommon "github.com/edgexfoundry/device-sdk-go/v4/internal/common"
	sdkModels "github.com/edgexfoundry/device-sdk-go/v4/pkg/models"
)

const (
	defaultBase   float64 = 0.0
	defaultScale  float64 = 1.0
	defaultOffset float64 = 0.0
	defaultMask   uint64  = 0
	defaultShift  int64   = 0

	Overflow = "overflow"
	NaN      = "NaN"
)

func TransformReadResult(cv *sdkModels.CommandValue, pv models.ResourceProperties) errors.EdgeX {
	if !isNumericValueType(cv) {
		return nil
	}
	res, err := isNaN(cv)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	} else if res {
		errMSg := fmt.Sprintf("NaN error for DeviceResource %s", cv.DeviceResourceName)
		return errors.NewCommonEdgeX(errors.KindNaNError, errMSg, nil)
	}

	value, err := commandValueForTransform(cv)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	newValue := value

	if pv.Mask != nil && *pv.Mask != defaultMask &&
		(cv.Type == common.ValueTypeUint8 || cv.Type == common.ValueTypeUint16 || cv.Type == common.ValueTypeUint32 || cv.Type == common.ValueTypeUint64) {
		newValue, err = transformReadMask(newValue, *pv.Mask)
		if err != nil {
			return errors.NewCommonEdgeXWrapper(err)
		}
	}
	if pv.Shift != nil && *pv.Shift != defaultShift &&
		(cv.Type == common.ValueTypeUint8 || cv.Type == common.ValueTypeUint16 || cv.Type == common.ValueTypeUint32 || cv.Type == common.ValueTypeUint64) {
		newValue, err = transformReadShift(newValue, *pv.Shift)
		if err != nil {
			return errors.NewCommonEdgeXWrapper(err)
		}
	}
	if pv.Base != nil && *pv.Base != defaultBase {
		newValue, err = transformBase(newValue, *pv.Base, true)
		if err != nil {
			return errors.NewCommonEdgeXWrapper(err)
		}
	}
	if pv.Scale != nil && *pv.Scale != defaultScale {
		newValue, err = transformScale(newValue, *pv.Scale, true)
		if err != nil {
			return errors.NewCommonEdgeXWrapper(err)
		}
	}
	if pv.Offset != nil && *pv.Offset != defaultOffset {
		newValue, err = transformOffset(newValue, *pv.Offset, true)
		if err != nil {
			return errors.NewCommonEdgeXWrapper(err)
		}
	}

	if value != newValue {
		cv.Value = newValue
	}
	return nil
}

func transformBase(value any, base float64, read bool) (any, errors.EdgeX) {
	var valueFloat64 float64
	switch v := value.(type) {
	case uint8:
		valueFloat64 = float64(v)
	case uint16:
		valueFloat64 = float64(v)
	case uint32:
		valueFloat64 = float64(v)
	case uint64:
		valueFloat64 = float64(v)
	case int8:
		valueFloat64 = float64(v)
	case int16:
		valueFloat64 = float64(v)
	case int32:
		valueFloat64 = float64(v)
	case int64:
		valueFloat64 = float64(v)
	case float32:
		valueFloat64 = float64(v)
	case float64:
		valueFloat64 = v
	}

	if read {
		valueFloat64 = math.Pow(base, valueFloat64)
	} else {
		valueFloat64 = math.Log(valueFloat64) / math.Log(base)
	}
	inRange := checkTransformedValueInRange(value, valueFloat64)
	if !inRange {
		errMsg := fmt.Sprintf("transformed value out of its original type (%T) range", value)
		return 0, errors.NewCommonEdgeX(errors.KindOverflowError, errMsg, nil)
	}

	switch value.(type) {
	case uint8:
		value = uint8(valueFloat64)
	case uint16:
		value = uint16(valueFloat64)
	case uint32:
		value = uint32(valueFloat64)
	case uint64:
		value = uint64(valueFloat64)
	case int8:
		value = int8(valueFloat64)
	case int16:
		value = int16(valueFloat64)
	case int32:
		value = int32(valueFloat64)
	case int64:
		value = int64(valueFloat64)
	case float32:
		value = float32(valueFloat64)
	case float64:
		value = valueFloat64
	}
	return value, nil
}

func transformScale(value any, scale float64, read bool) (any, errors.EdgeX) {
	var valueFloat64 float64
	switch v := value.(type) {
	case uint8:
		valueFloat64 = float64(v)
	case uint16:
		valueFloat64 = float64(v)
	case uint32:
		valueFloat64 = float64(v)
	case uint64:
		valueFloat64 = float64(v)
	case int8:
		valueFloat64 = float64(v)
	case int16:
		valueFloat64 = float64(v)
	case int32:
		valueFloat64 = float64(v)
	case int64:
		valueFloat64 = float64(v)
	case float32:
		valueFloat64 = float64(v)
	case float64:
		valueFloat64 = v
	}

	if read {
		valueFloat64 = valueFloat64 * scale
	} else {
		valueFloat64 = valueFloat64 / scale
	}
	inRange := checkTransformedValueInRange(value, valueFloat64)
	if !inRange {
		errMsg := fmt.Sprintf("transformed value out of its original type (%T) range", value)
		return 0, errors.NewCommonEdgeX(errors.KindOverflowError, errMsg, nil)
	}

	switch v := value.(type) {
	case uint8:
		if read {
			value = v * uint8(scale)
		} else {
			value = v / uint8(scale)
		}
	case uint16:
		if read {
			value = v * uint16(scale)
		} else {
			value = v / uint16(scale)
		}
	case uint32:
		if read {
			value = v * uint32(scale)
		} else {
			value = v / uint32(scale)
		}
	case uint64:
		if read {
			value = v * uint64(scale)
		} else {
			value = v / uint64(scale)
		}
	case int8:
		if read {
			value = v * int8(scale)
		} else {
			value = v / int8(scale)
		}
	case int16:
		if read {
			value = v * int16(scale)
		} else {
			value = v / int16(scale)
		}
	case int32:
		if read {
			value = v * int32(scale)
		} else {
			value = v / int32(scale)
		}
	case int64:
		if read {
			value = v * int64(scale)
		} else {
			value = v / int64(scale)
		}
	case float32:
		if read {
			value = v * float32(scale)
		} else {
			value = v / float32(scale)
		}
	case float64:
		if read {
			value = v * scale
		} else {
			value = v / scale
		}
	}
	return value, nil
}

func transformOffset(value any, offset float64, read bool) (any, errors.EdgeX) {
	var valueFloat64 float64
	switch v := value.(type) {
	case uint8:
		valueFloat64 = float64(v)
	case uint16:
		valueFloat64 = float64(v)
	case uint32:
		valueFloat64 = float64(v)
	case uint64:
		valueFloat64 = float64(v)
	case int8:
		valueFloat64 = float64(v)
	case int16:
		valueFloat64 = float64(v)
	case int32:
		valueFloat64 = float64(v)
	case int64:
		valueFloat64 = float64(v)
	case float32:
		valueFloat64 = float64(v)
	case float64:
		valueFloat64 = v
	}

	if read {
		valueFloat64 = valueFloat64 + offset
	} else {
		valueFloat64 = valueFloat64 - offset
	}
	inRange := checkTransformedValueInRange(value, valueFloat64)
	if !inRange {
		errMsg := fmt.Sprintf("transformed value out of its original type (%T) range", value)
		return 0, errors.NewCommonEdgeX(errors.KindOverflowError, errMsg, nil)
	}

	switch v := value.(type) {
	case uint8:
		if read {
			value = v + uint8(offset)
		} else {
			value = v - uint8(offset)
		}
	case uint16:
		if read {
			value = v + uint16(offset)
		} else {
			value = v - uint16(offset)
		}
	case uint32:
		if read {
			value = v + uint32(offset)
		} else {
			value = v - uint32(offset)
		}
	case uint64:
		if read {
			value = v + uint64(offset)
		} else {
			value = v - uint64(offset)
		}
	case int8:
		if read {
			value = v + int8(offset)
		} else {
			value = v - int8(offset)
		}
	case int16:
		if read {
			value = v + int16(offset)
		} else {
			value = v - int16(offset)
		}
	case int32:
		if read {
			value = v + int32(offset)
		} else {
			value = v - int32(offset)
		}
	case int64:
		if read {
			value = v + int64(offset)
		} else {
			value = v - int64(offset)
		}
	case float32:
		if read {
			value = v + float32(offset)
		} else {
			value = v - float32(offset)
		}
	case float64:
		if read {
			value = v + offset
		} else {
			value = v - offset
		}
	}
	return value, nil
}

func transformReadMask(value any, mask uint64) (any, errors.EdgeX) {
	switch v := value.(type) {
	case uint8:
		value = v & uint8(mask)
	case uint16:
		value = v & uint16(mask)
	case uint32:
		value = v & uint32(mask)
	case uint64:
		value = v & mask
	}

	return value, nil
}

func transformReadShift(value any, shift int64) (any, errors.EdgeX) {
	switch v := value.(type) {
	case uint8:
		if shift > 0 {
			value = v << int8(shift)
		} else {
			value = v >> int8(-shift)
		}
	case uint16:
		if shift > 0 {
			value = v << int16(shift)
		} else {
			value = v >> int16(-shift)
		}
	case uint32:
		if shift > 0 {
			value = v << int32(shift)
		} else {
			value = v >> int32(-shift)
		}
	case uint64:
		if shift > 0 {
			value = v << shift
		} else {
			value = v >> (-shift)
		}
	}

	return value, nil
}

func commandValueForTransform(cv *sdkModels.CommandValue) (interface{}, errors.EdgeX) {
	if cv.Value == nil {
		return nil, nil
	}
	var v interface{}
	var err error
	switch cv.Type {
	case common.ValueTypeUint8:
		v, err = cv.Uint8Value()
		if err != nil {
			return 0, errors.NewCommonEdgeXWrapper(err)
		}
	case common.ValueTypeUint16:
		v, err = cv.Uint16Value()
		if err != nil {
			return 0, errors.NewCommonEdgeXWrapper(err)
		}
	case common.ValueTypeUint32:
		v, err = cv.Uint32Value()
		if err != nil {
			return 0, errors.NewCommonEdgeXWrapper(err)
		}
	case common.ValueTypeUint64:
		v, err = cv.Uint64Value()
		if err != nil {
			return 0, errors.NewCommonEdgeXWrapper(err)
		}
	case common.ValueTypeInt8:
		v, err = cv.Int8Value()
		if err != nil {
			return 0, errors.NewCommonEdgeXWrapper(err)
		}
	case common.ValueTypeInt16:
		v, err = cv.Int16Value()
		if err != nil {
			return 0, errors.NewCommonEdgeXWrapper(err)
		}
	case common.ValueTypeInt32:
		v, err = cv.Int32Value()
		if err != nil {
			return 0, errors.NewCommonEdgeXWrapper(err)
		}
	case common.ValueTypeInt64:
		v, err = cv.Int64Value()
		if err != nil {
			return 0, errors.NewCommonEdgeXWrapper(err)
		}
	case common.ValueTypeFloat32:
		v, err = cv.Float32Value()
		if err != nil {
			return 0, errors.NewCommonEdgeXWrapper(err)
		}
	case common.ValueTypeFloat64:
		v, err = cv.Float64Value()
		if err != nil {
			return 0, errors.NewCommonEdgeXWrapper(err)
		}
	default:
		return nil, errors.NewCommonEdgeX(errors.KindServerError, "unsupported ValueType for transformation", nil)
	}
	return v, nil
}

func checkAssertion(
	cv *sdkModels.CommandValue,
	assertion string,
	deviceName string,
	lc logger.LoggingClient,
	dc interfaces.DeviceClient) errors.EdgeX {
	if assertion != "" && cv.ValueToString() != assertion {
		go sdkCommon.UpdateOperatingState(deviceName, models.Down, lc, dc)
		errMsg := fmt.Sprintf("Assertion failed for DeviceResource %s, with value %s", cv.DeviceResourceName, cv.ValueToString())
		return errors.NewCommonEdgeX(errors.KindServerError, errMsg, nil)
	}
	return nil
}

func mapCommandValue(value *sdkModels.CommandValue, mappings map[string]string) (*sdkModels.CommandValue, bool) {
	var err error
	var result *sdkModels.CommandValue

	newValue, ok := mappings[value.ValueToString()]
	if ok {
		result, err = sdkModels.NewCommandValue(value.DeviceResourceName, common.ValueTypeString, newValue)
		if err != nil {
			return nil, false
		}
	}
	return result, ok
}

func isNumericValueType(cv *sdkModels.CommandValue) bool {
	switch cv.Type {
	case common.ValueTypeUint8:
	case common.ValueTypeUint16:
	case common.ValueTypeUint32:
	case common.ValueTypeUint64:
	case common.ValueTypeInt8:
	case common.ValueTypeInt16:
	case common.ValueTypeInt32:
	case common.ValueTypeInt64:
	case common.ValueTypeFloat32:
	case common.ValueTypeFloat64:
	default:
		return false
	}
	return true
}
