// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018-2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package transformer

import (
	"fmt"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"

	dsModels "github.com/edgexfoundry/device-sdk-go/v4/pkg/models"
)

func TransformWriteParameter(cv *dsModels.CommandValue, pv models.ResourceProperties) errors.EdgeX {
	if cv.Value == nil {
		return nil
	}
	if !isNumericValueType(cv) {
		return nil
	}

	value, err := commandValueForTransform(cv)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	newValue := value

	if pv.Maximum != nil {
		err = validateWriteMaximum(value, *pv.Maximum)
		if err != nil {
			return errors.NewCommonEdgeXWrapper(err)
		}
	}
	if pv.Minimum != nil {
		err = validateWriteMinimum(value, *pv.Minimum)
		if err != nil {
			return errors.NewCommonEdgeXWrapper(err)
		}
	}
	if pv.Offset != nil && *pv.Offset != defaultOffset {
		newValue, err = transformOffset(newValue, *pv.Offset, false)
		if err != nil {
			return errors.NewCommonEdgeXWrapper(err)
		}
	}
	if pv.Scale != nil && *pv.Scale != defaultScale {
		newValue, err = transformScale(newValue, *pv.Scale, false)
		if err != nil {
			return errors.NewCommonEdgeXWrapper(err)
		}
	}
	if pv.Base != nil && *pv.Base != defaultBase {
		newValue, err = transformBase(newValue, *pv.Base, false)
		if err != nil {
			return errors.NewCommonEdgeXWrapper(err)
		}
	}

	if value != newValue {
		cv.Value = newValue
	}
	return nil
}

func validateWriteMaximum(value any, maximum float64) errors.EdgeX {

	switch v := value.(type) {
	case uint8:
		if v > uint8(maximum) {
			errMsg := fmt.Sprintf("set command parameter out of maximum value %v", maximum)
			return errors.NewCommonEdgeX(errors.KindContractInvalid, errMsg, nil)
		}
	case uint16:
		if v > uint16(maximum) {
			errMsg := fmt.Sprintf("set command parameter out of maximum value %v", maximum)
			return errors.NewCommonEdgeX(errors.KindContractInvalid, errMsg, nil)
		}
	case uint32:
		if v > uint32(maximum) {
			errMsg := fmt.Sprintf("set command parameter out of maximum value %v", maximum)
			return errors.NewCommonEdgeX(errors.KindContractInvalid, errMsg, nil)
		}
	case uint64:
		if v > uint64(maximum) {
			errMsg := fmt.Sprintf("set command parameter out of maximum value %v", maximum)
			return errors.NewCommonEdgeX(errors.KindContractInvalid, errMsg, nil)
		}
	case int8:
		if v > int8(maximum) {
			errMsg := fmt.Sprintf("set command parameter out of maximum value %v", maximum)
			return errors.NewCommonEdgeX(errors.KindContractInvalid, errMsg, nil)
		}
	case int16:
		if v > int16(maximum) {
			errMsg := fmt.Sprintf("set command parameter out of maximum value %v", maximum)
			return errors.NewCommonEdgeX(errors.KindContractInvalid, errMsg, nil)
		}
	case int32:
		if v > int32(maximum) {
			errMsg := fmt.Sprintf("set command parameter out of maximum value %v", maximum)
			return errors.NewCommonEdgeX(errors.KindContractInvalid, errMsg, nil)
		}
	case int64:
		if v > int64(maximum) {
			errMsg := fmt.Sprintf("set command parameter out of maximum value %v", maximum)
			return errors.NewCommonEdgeX(errors.KindContractInvalid, errMsg, nil)
		}
	case float32:
		if v > float32(maximum) {
			errMsg := fmt.Sprintf("set command parameter out of maximum value %v", maximum)
			return errors.NewCommonEdgeX(errors.KindContractInvalid, errMsg, nil)
		}
	case float64:
		if v > maximum {
			errMsg := fmt.Sprintf("set command parameter out of maximum value %v", maximum)
			return errors.NewCommonEdgeX(errors.KindContractInvalid, errMsg, nil)
		}
	}
	return nil
}

func validateWriteMinimum(value any, minimum float64) errors.EdgeX {
	switch v := value.(type) {
	case uint8:
		if v < uint8(minimum) {
			errMsg := fmt.Sprintf("set command parameter out of minimum value %v", minimum)
			return errors.NewCommonEdgeX(errors.KindContractInvalid, errMsg, nil)
		}
	case uint16:
		if v < uint16(minimum) {
			errMsg := fmt.Sprintf("set command parameter out of minimum value %v", minimum)
			return errors.NewCommonEdgeX(errors.KindContractInvalid, errMsg, nil)
		}
	case uint32:
		if v < uint32(minimum) {
			errMsg := fmt.Sprintf("set command parameter out of minimum value %v", minimum)
			return errors.NewCommonEdgeX(errors.KindContractInvalid, errMsg, nil)
		}
	case uint64:
		if v < uint64(minimum) {
			errMsg := fmt.Sprintf("set command parameter out of minimum value %v", minimum)
			return errors.NewCommonEdgeX(errors.KindContractInvalid, errMsg, nil)
		}
	case int8:
		if v < int8(minimum) {
			errMsg := fmt.Sprintf("set command parameter out of minimum value %v", minimum)
			return errors.NewCommonEdgeX(errors.KindContractInvalid, errMsg, nil)
		}
	case int16:
		if v < int16(minimum) {
			errMsg := fmt.Sprintf("set command parameter out of minimum value %v", minimum)
			return errors.NewCommonEdgeX(errors.KindContractInvalid, errMsg, nil)
		}
	case int32:
		if v < int32(minimum) {
			errMsg := fmt.Sprintf("set command parameter out of minimum value %v", minimum)
			return errors.NewCommonEdgeX(errors.KindContractInvalid, errMsg, nil)
		}
	case int64:
		if v < int64(minimum) {
			errMsg := fmt.Sprintf("set command parameter out of minimum value %v", minimum)
			return errors.NewCommonEdgeX(errors.KindContractInvalid, errMsg, nil)
		}
	case float32:
		if v < float32(minimum) {
			errMsg := fmt.Sprintf("set command parameter out of minimum value %v", minimum)
			return errors.NewCommonEdgeX(errors.KindContractInvalid, errMsg, nil)
		}
	case float64:
		if v < minimum {
			errMsg := fmt.Sprintf("set command parameter out of minimum value %v", minimum)
			return errors.NewCommonEdgeX(errors.KindContractInvalid, errMsg, nil)
		}
	}
	return nil
}
