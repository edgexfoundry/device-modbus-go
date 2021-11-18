// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2019-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package transformer

import (
	"fmt"
	"math"
	"strconv"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"

	sdkCommon "github.com/edgexfoundry/device-sdk-go/v2/internal/common"
	sdkModels "github.com/edgexfoundry/device-sdk-go/v2/pkg/models"
)

const (
	defaultBase   string = "0"
	defaultScale  string = "1.0"
	defaultOffset string = "0.0"
	defaultMask   string = "0"
	defaultShift  string = "0"

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
	newValue := value

	if pv.Mask != "" && pv.Mask != defaultMask &&
		(cv.Type == common.ValueTypeUint8 || cv.Type == common.ValueTypeUint16 || cv.Type == common.ValueTypeUint32 || cv.Type == common.ValueTypeUint64) {
		newValue, err = transformReadMask(newValue, pv.Mask)
		if err != nil {
			return errors.NewCommonEdgeXWrapper(err)
		}
	}
	if pv.Shift != "" && pv.Shift != defaultShift &&
		(cv.Type == common.ValueTypeUint8 || cv.Type == common.ValueTypeUint16 || cv.Type == common.ValueTypeUint32 || cv.Type == common.ValueTypeUint64) {
		newValue, err = transformReadShift(newValue, pv.Shift)
		if err != nil {
			return errors.NewCommonEdgeXWrapper(err)
		}
	}
	if pv.Base != "" && pv.Base != defaultBase {
		newValue, err = transformBase(newValue, pv.Base, true)
		if err != nil {
			return errors.NewCommonEdgeXWrapper(err)
		}
	}
	if pv.Scale != "" && pv.Scale != defaultScale {
		newValue, err = transformScale(newValue, pv.Scale, true)
		if err != nil {
			return errors.NewCommonEdgeXWrapper(err)
		}
	}
	if pv.Offset != "" && pv.Offset != defaultOffset {
		newValue, err = transformOffset(newValue, pv.Offset, true)
		if err != nil {
			return errors.NewCommonEdgeXWrapper(err)
		}
	}

	if value != newValue {
		cv.Value = newValue
	}
	return nil
}

func transformBase(value interface{}, base string, read bool) (interface{}, errors.EdgeX) {
	b, err := strconv.ParseFloat(base, 64)
	if err != nil {
		errMsg := fmt.Sprintf("the base value %s of PropertyValue cannot be parsed to float64", base)
		return value, errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
	}

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
		valueFloat64 = math.Pow(b, valueFloat64)
	} else {
		valueFloat64 = math.Log(valueFloat64) / math.Log(b)
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

func transformScale(value interface{}, scale string, read bool) (interface{}, errors.EdgeX) {
	s, err := strconv.ParseFloat(scale, 64)
	if err != nil {
		errMsg := fmt.Sprintf("the scale value %s of PropertyValue cannot be parsed to float64", scale)
		return value, errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
	}

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
		valueFloat64 = valueFloat64 * s
	} else {
		valueFloat64 = valueFloat64 / s
	}
	inRange := checkTransformedValueInRange(value, valueFloat64)
	if !inRange {
		errMsg := fmt.Sprintf("transformed value out of its original type (%T) range", value)
		return 0, errors.NewCommonEdgeX(errors.KindOverflowError, errMsg, nil)
	}

	switch v := value.(type) {
	case uint8:
		if read {
			value = v * uint8(s)
		} else {
			value = v / uint8(s)
		}
	case uint16:
		if read {
			value = v * uint16(s)
		} else {
			value = v / uint16(s)
		}
	case uint32:
		if read {
			value = v * uint32(s)
		} else {
			value = v / uint32(s)
		}
	case uint64:
		if read {
			value = v * uint64(s)
		} else {
			value = v / uint64(s)
		}
	case int8:
		if read {
			value = v * int8(s)
		} else {
			value = v / int8(s)
		}
	case int16:
		if read {
			value = v * int16(s)
		} else {
			value = v / int16(s)
		}
	case int32:
		if read {
			value = v * int32(s)
		} else {
			value = v / int32(s)
		}
	case int64:
		if read {
			value = v * int64(s)
		} else {
			value = v / int64(s)
		}
	case float32:
		if read {
			value = v * float32(s)
		} else {
			value = v / float32(s)
		}
	case float64:
		if read {
			value = v * s
		} else {
			value = v / s
		}
	}
	return value, nil
}

func transformOffset(value interface{}, offset string, read bool) (interface{}, errors.EdgeX) {
	o, err := strconv.ParseFloat(offset, 64)
	if err != nil {
		errMsg := fmt.Sprintf("the offset value %s of PropertyValue cannot be parsed to float64", offset)
		return value, errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
	}

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
		valueFloat64 = valueFloat64 + o
	} else {
		valueFloat64 = valueFloat64 - 0
	}
	inRange := checkTransformedValueInRange(value, valueFloat64)
	if !inRange {
		errMsg := fmt.Sprintf("transformed value out of its original type (%T) range", value)
		return 0, errors.NewCommonEdgeX(errors.KindOverflowError, errMsg, nil)
	}

	switch v := value.(type) {
	case uint8:
		if read {
			value = v + uint8(o)
		} else {
			value = v - uint8(o)
		}
	case uint16:
		if read {
			value = v + uint16(o)
		} else {
			value = v - uint16(o)
		}
	case uint32:
		if read {
			value = v + uint32(o)
		} else {
			value = v - uint32(o)
		}
	case uint64:
		if read {
			value = v + uint64(o)
		} else {
			value = v - uint64(o)
		}
	case int8:
		if read {
			value = v + int8(o)
		} else {
			value = v - int8(o)
		}
	case int16:
		if read {
			value = v + int16(o)
		} else {
			value = v - int16(o)
		}
	case int32:
		if read {
			value = v + int32(o)
		} else {
			value = v - int32(o)
		}
	case int64:
		if read {
			value = v + int64(o)
		} else {
			value = v - int64(o)
		}
	case float32:
		if read {
			value = v + float32(o)
		} else {
			value = v - float32(o)
		}
	case float64:
		if read {
			value = v + o
		} else {
			value = v - o
		}
	}
	return value, nil
}

func transformReadMask(value interface{}, mask string) (interface{}, errors.EdgeX) {
	switch v := value.(type) {
	case uint8:
		m, err := strconv.ParseUint(mask, 10, 8)
		if err != nil {
			errMsg := fmt.Sprintf("the mask value %s of PropertyValue cannot be parsed to %T", mask, v)
			return value, errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
		}
		value = v & uint8(m)
	case uint16:
		m, err := strconv.ParseUint(mask, 10, 16)
		if err != nil {
			errMsg := fmt.Sprintf("the mask value %s of PropertyValue cannot be parsed to %T", mask, v)
			return value, errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
		}
		value = v & uint16(m)
	case uint32:
		m, err := strconv.ParseUint(mask, 10, 32)
		if err != nil {
			errMsg := fmt.Sprintf("the mask value %s of PropertyValue cannot be parsed to %T", mask, v)
			return value, errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
		}
		value = v & uint32(m)
	case uint64:
		m, err := strconv.ParseUint(mask, 10, 64)
		if err != nil {
			errMsg := fmt.Sprintf("the mask value %s of PropertyValue cannot be parsed to %T", mask, v)
			return value, errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
		}
		value = v & m
	}

	return value, nil
}

func transformReadShift(value interface{}, shift string) (interface{}, errors.EdgeX) {
	switch v := value.(type) {
	case uint8:
		s, err := strconv.ParseInt(shift, 10, 8)
		if err != nil {
			errMsg := fmt.Sprintf("the shift value %s of PropertyValue cannot be parsed to %T", shift, v)
			return value, errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
		}
		if s > 0 {
			value = v << s
		} else {
			value = v >> (-s)
		}
	case uint16:
		s, err := strconv.ParseInt(shift, 10, 16)
		if err != nil {
			errMsg := fmt.Sprintf("the shift value %s of PropertyValue cannot be parsed to %T", shift, v)
			return value, errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
		}
		if s > 0 {
			value = v << s
		} else {
			value = v >> (-s)
		}
	case uint32:
		s, err := strconv.ParseInt(shift, 10, 32)
		if err != nil {
			errMsg := fmt.Sprintf("the shift value %s of PropertyValue cannot be parsed to %T", shift, v)
			return value, errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
		}
		if s > 0 {
			value = v << s
		} else {
			value = v >> (-s)
		}
	case uint64:
		s, err := strconv.ParseInt(shift, 10, 64)
		if err != nil {
			errMsg := fmt.Sprintf("the shift value %s of PropertyValue cannot be parsed to %T", shift, v)
			return value, errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
		}
		if s > 0 {
			value = v << s
		} else {
			value = v >> (-s)
		}
	}

	return value, nil
}

func commandValueForTransform(cv *sdkModels.CommandValue) (interface{}, errors.EdgeX) {
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
