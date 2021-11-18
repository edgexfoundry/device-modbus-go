// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package transformer

import (
	"math"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"

	"github.com/edgexfoundry/device-sdk-go/v2/pkg/models"
)

func isNaN(cv *models.CommandValue) (bool, errors.EdgeX) {
	switch cv.Type {
	case common.ValueTypeFloat32:
		v, err := cv.Float32Value()
		if err != nil {
			return false, errors.NewCommonEdgeXWrapper(err)
		}
		if math.IsNaN(float64(v)) {
			return true, nil
		}
	case common.ValueTypeFloat64:
		v, err := cv.Float64Value()
		if err != nil {
			return false, errors.NewCommonEdgeXWrapper(err)
		}
		if math.IsNaN(v) {
			return true, nil
		}
	}
	return false, nil
}
