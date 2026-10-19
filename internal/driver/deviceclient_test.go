// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018-2021 IOTech Ltd
// Copyright (C) 2026 encentive GmbH
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"testing"
	"time"

	"github.com/edgexfoundry/device-sdk-go/v4/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	driver = new(Driver)
	driver.Logger = logger.NewMockClient()
}

func TestTransformDataBytesToResult_INT16(t *testing.T) {
	req := models.CommandRequest{
		DeviceResourceName: "light",
		Type:               common.ValueTypeInt16,
		Attributes: map[string]interface{}{
			PRIMARY_TABLE:    INPUT_REGISTERS,
			STARTING_ADDRESS: 10,
		},
	}
	commandInfo, err := createCommandInfo(&req)
	require.NoError(t, err)
	dataBytes := []byte{255, 231} // => big-endian [231,255] => -25
	expected := int16(-25)

	commandValue, err := TransformDataBytesToResult(&req, dataBytes, commandInfo)
	require.NoError(t, err)
	result, err := commandValue.Int16Value()
	require.NoError(t, err)

	assert.Equal(t, expected, result)
}

func TestTransformDataBytesToResult_INT32(t *testing.T) {
	req := models.CommandRequest{
		DeviceResourceName: "light",
		Type:               common.ValueTypeInt32,
		Attributes: map[string]interface{}{
			PRIMARY_TABLE:    INPUT_REGISTERS,
			STARTING_ADDRESS: 10,
		},
	}
	commandInfo, err := createCommandInfo(&req)
	require.NoError(t, err)
	dataBytes := []byte{0, 0, 1, 11} // big-endian [11,1,0,0] => 11+2^8=267
	expected := int32(267)

	commandValue, err := TransformDataBytesToResult(&req, dataBytes, commandInfo)
	require.NoError(t, err)
	result, err := commandValue.Int32Value()
	require.NoError(t, err)

	assert.Equal(t, expected, result)
}

func TestTransformDataBytesToResult_INT64(t *testing.T) {
	req := models.CommandRequest{
		DeviceResourceName: "light",
		Type:               common.ValueTypeInt64,
		Attributes: map[string]interface{}{
			PRIMARY_TABLE:    INPUT_REGISTERS,
			STARTING_ADDRESS: 10,
		},
	}
	commandInfo, err := createCommandInfo(&req)
	require.NoError(t, err)
	dataBytes := []byte{0, 1, 0, 0, 0, 2, 1, 1} // big-endian [1,1,2,0,0,0,1,0] => 1+2^8+2^17+2^48=281474976841985
	expected := int64(281474976841985)

	commandValue, err := TransformDataBytesToResult(&req, dataBytes, commandInfo)
	require.NoError(t, err)
	result, err := commandValue.Int64Value()
	require.NoError(t, err)

	assert.Equal(t, expected, result)
}

func TestTransformDataBytesToResult_UINT16(t *testing.T) {
	req := models.CommandRequest{
		DeviceResourceName: "light",
		Type:               common.ValueTypeUint16,
		Attributes: map[string]interface{}{
			PRIMARY_TABLE:    INPUT_REGISTERS,
			STARTING_ADDRESS: 10,
		},
	}
	commandInfo, err := createCommandInfo(&req)
	require.NoError(t, err)
	dataBytes := []byte{0, 11} // => big-endian [11,0] => 11
	expected := uint16(11)

	commandValue, err := TransformDataBytesToResult(&req, dataBytes, commandInfo)
	require.NoError(t, err)
	result, err := commandValue.Uint16Value()
	require.NoError(t, err)

	assert.Equal(t, expected, result)
}

func TestTransformDataBytesToResult_UINT32(t *testing.T) {
	req := models.CommandRequest{
		DeviceResourceName: "light",
		Type:               common.ValueTypeUint32,
		Attributes: map[string]interface{}{
			PRIMARY_TABLE:    HOLDING_REGISTERS,
			STARTING_ADDRESS: 10,
		},
	}
	commandInfo, err := createCommandInfo(&req)
	require.NoError(t, err)
	dataBytes := []byte{0, 0, 1, 11} // big-endian [11,1,0,0] => 11+2^8=267
	expected := uint32(267)

	commandValue, err := TransformDataBytesToResult(&req, dataBytes, commandInfo)
	require.NoError(t, err)
	result, err := commandValue.Uint32Value()
	require.NoError(t, err)

	assert.Equal(t, expected, result)
}

func TestTransformDataBytesToResult_UINT64(t *testing.T) {
	req := models.CommandRequest{
		DeviceResourceName: "light",
		Type:               common.ValueTypeUint64,
		Attributes: map[string]interface{}{
			PRIMARY_TABLE:    HOLDING_REGISTERS,
			STARTING_ADDRESS: 10,
		},
	}
	commandInfo, err := createCommandInfo(&req)
	require.NoError(t, err)
	dataBytes := []byte{0, 1, 0, 0, 0, 2, 1, 1} // big-endian [1,1,2,0,0,0,1,0] => 1+2^8+2^17+2^48=281474976841985

	expected := uint64(281474976841985)

	commandValue, err := TransformDataBytesToResult(&req, dataBytes, commandInfo)
	require.NoError(t, err)
	result, err := commandValue.Uint64Value()
	require.NoError(t, err)

	assert.Equal(t, expected, result)
}

func TestTransformDataBytesToResult_FLOAT32(t *testing.T) {
	req := models.CommandRequest{
		DeviceResourceName: "light",
		Type:               common.ValueTypeFloat32,
		Attributes: map[string]interface{}{
			PRIMARY_TABLE:    HOLDING_REGISTERS,
			STARTING_ADDRESS: 10,
		},
	}
	commandInfo, err := createCommandInfo(&req)
	require.NoError(t, err)
	dataBytes := []byte{63, 143, 92, 41} // big-endian [41,92,143,63] => 1.12
	expected := float32(1.12)

	commandValue, err := TransformDataBytesToResult(&req, dataBytes, commandInfo)
	require.NoError(t, err)
	result, err := commandValue.Float32Value()
	require.NoError(t, err)

	assert.Equal(t, expected, result)
}

func TestTransformDataBytesToResult_FLOAT64(t *testing.T) {
	req := models.CommandRequest{
		DeviceResourceName: "light",
		Type:               common.ValueTypeFloat64,
		Attributes: map[string]interface{}{
			PRIMARY_TABLE:    HOLDING_REGISTERS,
			STARTING_ADDRESS: 10,
		},
	}
	commandInfo, err := createCommandInfo(&req)
	require.NoError(t, err)
	dataBytes := []byte{63, 241, 235, 133, 30, 184, 81, 236} // => 1.12
	expected := float64(1.12)

	commandValue, err := TransformDataBytesToResult(&req, dataBytes, commandInfo)
	require.NoError(t, err)
	result, err := commandValue.Float64Value()
	require.NoError(t, err)

	assert.Equal(t, expected, result)
}

func TestTransformDataBytesToResult_BOOL(t *testing.T) {
	req := models.CommandRequest{
		DeviceResourceName: "light",
		Type:               common.ValueTypeBool,
		Attributes: map[string]interface{}{
			PRIMARY_TABLE:    DISCRETES_INPUT,
			STARTING_ADDRESS: 10,
		},
	}

	tests := []struct {
		name         string
		primaryTable string
		data         []byte
		expected     bool
	}{
		{"transform true value from DISCRETES_INPUT", DISCRETES_INPUT, []byte{1}, true},
		{"transform false value from DISCRETES_INPUT", DISCRETES_INPUT, []byte{0}, false},
		{"transform true value from DISCRETE_INPUTS", DISCRETE_INPUTS, []byte{1}, true},
		{"transform false value from DISCRETE_INPUTS", DISCRETE_INPUTS, []byte{0}, false},
		{"transform true value from COILS", COILS, []byte{1}, true},
		{"transform false value from COILS", COILS, []byte{0}, false},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req.Attributes[PRIMARY_TABLE] = testCase.primaryTable
			commandInfo, err := createCommandInfo(&req)
			require.NoError(t, err)
			commandValue, err := TransformDataBytesToResult(&req, testCase.data, commandInfo)
			require.NoError(t, err)
			result, err := commandValue.BoolValue()
			require.NoError(t, err)

			assert.Equal(t, testCase.expected, result)
		})
	}
}

func TestTransformDataBytesToResult_BOOL_REGISTER_Bit_Decomposition(t *testing.T) {
	req := models.CommandRequest{
		DeviceResourceName: "light",
		Type:               common.ValueTypeBool,
		Attributes: map[string]interface{}{
			PRIMARY_TABLE:    HOLDING_REGISTERS,
			STARTING_ADDRESS: 10.01,
		},
	}

	tests := []struct {
		name         string
		primaryTable string
		data         []byte
		expected     bool
	}{
		{"transform true value from HOLDING_REGISTERS", HOLDING_REGISTERS, []byte{0, 2}, true},
		{"transform false value fromHOLDING_REGISTERS", HOLDING_REGISTERS, []byte{0, 0}, false},
		{"transform true value from INPUT_REGISTERS", INPUT_REGISTERS, []byte{0, 2}, true},
		{"transform false value from INPUT_REGISTERS", INPUT_REGISTERS, []byte{0, 0}, false},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			req.Attributes[PRIMARY_TABLE] = testCase.primaryTable
			commandInfo, err := createCommandInfo(&req)
			require.NoError(t, err)
			commandValue, err := TransformDataBytesToResult(&req, testCase.data, commandInfo)
			require.NoError(t, err)
			result, err := commandValue.BoolValue()
			require.NoError(t, err)

			assert.Equal(t, testCase.expected, result)
		})
	}
}

func TestTransformDataBytesToResult_RawType_INT16_ValueType_FLOAT32(t *testing.T) {
	req := models.CommandRequest{
		DeviceResourceName: "light",
		Type:               common.ValueTypeFloat32,
		Attributes: map[string]interface{}{
			PRIMARY_TABLE:    INPUT_REGISTERS,
			STARTING_ADDRESS: 10,
			RAW_TYPE:         INT16,
		},
	}
	commandInfo, err := createCommandInfo(&req)
	require.NoError(t, err)
	dataBytes := []byte{255, 231} // => big-endian [231,255] => -25
	expected := float32(-25)

	commandValue, err := TransformDataBytesToResult(&req, dataBytes, commandInfo)
	require.NoError(t, err)
	result, err := commandValue.Float32Value()
	require.NoError(t, err)

	assert.Equal(t, expected, result)
}

func TestTransformDataBytesToResult_RawType_UINT16_ValueType_FLOAT32(t *testing.T) {
	req := models.CommandRequest{
		DeviceResourceName: "light",
		Type:               common.ValueTypeFloat32,
		Attributes: map[string]interface{}{
			PRIMARY_TABLE:    INPUT_REGISTERS,
			STARTING_ADDRESS: 10,
			RAW_TYPE:         UINT16,
		},
	}

	commandInfo, err := createCommandInfo(&req)
	require.NoError(t, err)
	dataBytes := []byte{0, 11} // => big-endian [11,0] => 11
	expected := float32(11)

	commandValue, err := TransformDataBytesToResult(&req, dataBytes, commandInfo)
	require.NoError(t, err)
	result, err := commandValue.Float32Value()
	require.NoError(t, err)

	assert.Equal(t, expected, result)
}

func TestTransformDataBytesToResult_RawType_INT16_ValueType_FLOAT64(t *testing.T) {
	req := models.CommandRequest{
		DeviceResourceName: "light",
		Type:               common.ValueTypeFloat64,
		Attributes: map[string]interface{}{
			PRIMARY_TABLE:    INPUT_REGISTERS,
			STARTING_ADDRESS: 10,
			RAW_TYPE:         INT16,
		},
	}
	commandInfo, err := createCommandInfo(&req)
	require.NoError(t, err)
	dataBytes := []byte{255, 231} // => big-endian [231,255] => -25
	expected := float64(-25)

	commandValue, err := TransformDataBytesToResult(&req, dataBytes, commandInfo)
	require.NoError(t, err)
	result, err := commandValue.Float64Value()
	require.NoError(t, err)

	assert.Equal(t, expected, result)
}

func TestTransformDataBytesToResult_RawType_UINT16_ValueType_FLOAT64(t *testing.T) {
	req := models.CommandRequest{
		DeviceResourceName: "light",
		Type:               common.ValueTypeFloat64,
		Attributes: map[string]interface{}{
			PRIMARY_TABLE:    INPUT_REGISTERS,
			STARTING_ADDRESS: 10,
			RAW_TYPE:         UINT16,
		},
	}

	commandInfo, err := createCommandInfo(&req)
	require.NoError(t, err)
	dataBytes := []byte{0, 11} // => big-endian [11,0] => 11
	expected := float64(11)

	commandValue, err := TransformDataBytesToResult(&req, dataBytes, commandInfo)
	require.NoError(t, err)
	result, err := commandValue.Float64Value()
	require.NoError(t, err)

	assert.Equal(t, expected, result)
}

func TestTransformDataBytesToResult_RawType_INT32_ValueType_FLOAT64(t *testing.T) {

	req := models.CommandRequest{
		DeviceResourceName: "light",
		Type:               common.ValueTypeFloat64,
		Attributes: map[string]interface{}{
			PRIMARY_TABLE:    INPUT_REGISTERS,
			STARTING_ADDRESS: 10,
			RAW_TYPE:         INT32,
		},
	}

	// Create a command info struct with the appropriate settings
	commandInfo, err := createCommandInfo(&req)
	require.NoError(t, err)
	// Create some test data bytes
	dataBytes := []byte{0x01, 0x01, 0x01, 0x01}
	expected := float64(16843009)

	// Call the function being tested
	commandValue, err := TransformDataBytesToResult(&req, dataBytes, commandInfo)
	require.NoError(t, err)
	result, err := commandValue.Float64Value()
	require.NoError(t, err)

	assert.Equal(t, expected, result)
}

func TestTransformDataBytesToResult_RawType_UINT32_ValueType_FLOAT64(t *testing.T) {
	req := models.CommandRequest{
		DeviceResourceName: "light",
		Type:               common.ValueTypeFloat64,
		Attributes: map[string]interface{}{
			PRIMARY_TABLE:    INPUT_REGISTERS,
			STARTING_ADDRESS: 10,
			RAW_TYPE:         UINT32,
		},
	}
	commandInfo, err := createCommandInfo(&req)
	require.NoError(t, err)

	cases := []struct {
		name      string
		dataBytes []byte
		expected  float64
	}{
		{name: "sign bit stays positive", dataBytes: []byte{0x80, 0x00, 0x00, 0x00}, expected: 2147483648.0},
		{name: "above 2^31", dataBytes: []byte{0xB2, 0xD0, 0x5E, 0x00}, expected: 3000000000.0},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			commandValue, err := TransformDataBytesToResult(&req, tt.dataBytes, commandInfo)
			require.NoError(t, err)
			result, err := commandValue.Float64Value()
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTransformDataBytesToResult_RawType_UINT32_ValueType_FLOAT64_Swaps(t *testing.T) {
	cases := []struct {
		name       string
		attributes map[string]interface{}
		dataBytes  []byte
	}{
		{
			name: "byte swap",
			attributes: map[string]interface{}{
				PRIMARY_TABLE:    INPUT_REGISTERS,
				STARTING_ADDRESS: 10,
				RAW_TYPE:         UINT32,
				IS_BYTE_SWAP:     true,
			},
			dataBytes: []byte{0, 0, 11, 1},
		},
		{
			name: "word swap",
			attributes: map[string]interface{}{
				PRIMARY_TABLE:    INPUT_REGISTERS,
				STARTING_ADDRESS: 10,
				RAW_TYPE:         UINT32,
				IS_WORD_SWAP:     true,
			},
			dataBytes: []byte{1, 11, 0, 0},
		},
		{
			name: "byte and word swap",
			attributes: map[string]interface{}{
				PRIMARY_TABLE:    INPUT_REGISTERS,
				STARTING_ADDRESS: 10,
				RAW_TYPE:         UINT32,
				IS_BYTE_SWAP:     true,
				IS_WORD_SWAP:     true,
			},
			dataBytes: []byte{11, 1, 0, 0},
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			req := models.CommandRequest{
				DeviceResourceName: "light",
				Type:               common.ValueTypeFloat64,
				Attributes:         tt.attributes,
			}
			commandInfo, err := createCommandInfo(&req)
			require.NoError(t, err)
			commandValue, err := TransformDataBytesToResult(&req, tt.dataBytes, commandInfo)
			require.NoError(t, err)
			result, err := commandValue.Float64Value()
			require.NoError(t, err)
			assert.Equal(t, float64(267), result)
		})
	}
}

func TestTransformDataBytesToResult_RawType_UINT64_ValueType_FLOAT64(t *testing.T) {
	attributes := map[string]interface{}{
		PRIMARY_TABLE:    INPUT_REGISTERS,
		STARTING_ADDRESS: 10,
		RAW_TYPE:         UINT64,
		IS_BYTE_SWAP:     true,
		IS_WORD_SWAP:     true,
	}
	req := models.CommandRequest{
		DeviceResourceName: "light",
		Type:               common.ValueTypeFloat64,
		Attributes:         attributes,
	}
	commandInfo, err := createCommandInfo(&req)
	require.NoError(t, err)
	dataBytes := []byte{0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00}
	commandValue, err := TransformDataBytesToResult(&req, dataBytes, commandInfo)
	require.NoError(t, err)
	result, err := commandValue.Float64Value()
	require.NoError(t, err)
	assert.Equal(t, float64(4294967296), result)
}

func TestTransformCommandValueToDataBytes_INT16(t *testing.T) {
	req := models.CommandRequest{
		DeviceResourceName: "light",
		Type:               common.ValueTypeInt16,
		Attributes: map[string]interface{}{
			PRIMARY_TABLE:    HOLDING_REGISTERS,
			STARTING_ADDRESS: 10,
		},
	}
	commandInfo, err := createCommandInfo(&req)
	require.NoError(t, err)
	resTime := time.Now().UnixNano()
	val, err := models.NewCommandValue(req.DeviceResourceName, common.ValueTypeInt16, int16(-25))
	require.NoError(t, err)
	val.Origin = resTime
	expected := []byte{255, 231}

	dataBytes, err := TransformCommandValueToDataBytes(commandInfo, val)
	require.NoError(t, err)

	assert.ElementsMatch(t, expected, dataBytes)
}

func TestTransformCommandValueToDataBytes_INT32(t *testing.T) {
	req := models.CommandRequest{
		DeviceResourceName: "light",
		Type:               common.ValueTypeInt32,
		Attributes: map[string]interface{}{
			PRIMARY_TABLE:    HOLDING_REGISTERS,
			STARTING_ADDRESS: 10,
		},
	}
	commandInfo, err := createCommandInfo(&req)
	require.NoError(t, err)
	resTime := time.Now().UnixNano()
	val, err := models.NewCommandValue(req.DeviceResourceName, common.ValueTypeInt32, int32(267))
	require.NoError(t, err)
	val.Origin = resTime
	expected := []byte{0, 0, 1, 11}

	dataBytes, err := TransformCommandValueToDataBytes(commandInfo, val)
	require.NoError(t, err)

	assert.ElementsMatch(t, expected, dataBytes)
}

func TestTransformCommandValueToDataBytes_INT64(t *testing.T) {
	req := models.CommandRequest{
		DeviceResourceName: "light",
		Type:               common.ValueTypeInt64,
		Attributes: map[string]interface{}{
			PRIMARY_TABLE:    HOLDING_REGISTERS,
			STARTING_ADDRESS: 10,
		},
	}
	commandInfo, err := createCommandInfo(&req)
	require.NoError(t, err)
	resTime := time.Now().UnixNano()
	val, err := models.NewCommandValue(req.DeviceResourceName, common.ValueTypeInt64, int64(281474976841985))
	require.NoError(t, err)
	val.Origin = resTime
	expected := []byte{0, 1, 0, 0, 0, 2, 1, 1}

	dataBytes, err := TransformCommandValueToDataBytes(commandInfo, val)
	require.NoError(t, err)

	assert.ElementsMatch(t, expected, dataBytes)
}

func TestTransformCommandValueToDataBytes_UINT16(t *testing.T) {
	req := models.CommandRequest{
		DeviceResourceName: "light",
		Type:               common.ValueTypeUint16,
		Attributes: map[string]interface{}{
			PRIMARY_TABLE:    HOLDING_REGISTERS,
			STARTING_ADDRESS: 10,
		},
	}
	commandInfo, err := createCommandInfo(&req)
	require.NoError(t, err)
	resTime := time.Now().UnixNano()
	val, err := models.NewCommandValue(req.DeviceResourceName, common.ValueTypeUint16, uint16(11))
	require.NoError(t, err)
	val.Origin = resTime
	expected := []byte{0, 11}

	dataBytes, err := TransformCommandValueToDataBytes(commandInfo, val)
	require.NoError(t, err)

	assert.ElementsMatch(t, expected, dataBytes)
}

func TestTransformCommandValueToDataBytes_UINT32(t *testing.T) {
	req := models.CommandRequest{
		DeviceResourceName: "light",
		Type:               common.ValueTypeUint32,
		Attributes: map[string]interface{}{
			PRIMARY_TABLE:    HOLDING_REGISTERS,
			STARTING_ADDRESS: 10,
		},
	}
	commandInfo, err := createCommandInfo(&req)
	require.NoError(t, err)
	resTime := time.Now().UnixNano()
	val, err := models.NewCommandValue(req.DeviceResourceName, common.ValueTypeUint32, uint32(267))
	require.NoError(t, err)
	val.Origin = resTime
	expected := []byte{0, 0, 1, 11}

	dataBytes, err := TransformCommandValueToDataBytes(commandInfo, val)
	require.NoError(t, err)

	assert.ElementsMatch(t, expected, dataBytes)
}

func TestTransformCommandValueToDataBytes_UINT64(t *testing.T) {
	req := models.CommandRequest{
		DeviceResourceName: "light",
		Type:               common.ValueTypeUint64,
		Attributes: map[string]interface{}{
			PRIMARY_TABLE:    HOLDING_REGISTERS,
			STARTING_ADDRESS: 10,
		},
	}
	commandInfo, err := createCommandInfo(&req)
	require.NoError(t, err)
	resTime := time.Now().UnixNano()
	val, err := models.NewCommandValue(req.DeviceResourceName, common.ValueTypeUint64, uint64(281474976841985))
	require.NoError(t, err)
	val.Origin = resTime
	expected := []byte{0, 1, 0, 0, 0, 2, 1, 1}

	dataBytes, err := TransformCommandValueToDataBytes(commandInfo, val)
	require.NoError(t, err)

	assert.ElementsMatch(t, expected, dataBytes)
}

func TestTransformCommandValueToDataBytes_FLOAT32(t *testing.T) {
	req := models.CommandRequest{
		DeviceResourceName: "light",
		Type:               common.ValueTypeFloat32,
		Attributes: map[string]interface{}{
			PRIMARY_TABLE:    HOLDING_REGISTERS,
			STARTING_ADDRESS: 10,
		},
	}
	commandInfo, err := createCommandInfo(&req)
	require.NoError(t, err)
	resTime := time.Now().UnixNano()
	val, err := models.NewCommandValue(req.DeviceResourceName, common.ValueTypeFloat32, float32(1.12))
	require.NoError(t, err)
	val.Origin = resTime
	expected := []byte{63, 143, 92, 41}

	dataBytes, err := TransformCommandValueToDataBytes(commandInfo, val)
	require.NoError(t, err)

	assert.ElementsMatch(t, expected, dataBytes)
}

func TestTransformCommandValueToDataBytes_FLOAT64(t *testing.T) {
	req := models.CommandRequest{
		DeviceResourceName: "light",
		Type:               common.ValueTypeFloat64,
		Attributes: map[string]interface{}{
			PRIMARY_TABLE:    HOLDING_REGISTERS,
			STARTING_ADDRESS: 10,
		},
	}
	commandInfo, err := createCommandInfo(&req)
	require.NoError(t, err)
	resTime := time.Now().UnixNano()
	val, err := models.NewCommandValue(req.DeviceResourceName, common.ValueTypeFloat64, 1.12)
	require.NoError(t, err)
	val.Origin = resTime
	expected := []byte{63, 241, 235, 133, 30, 184, 81, 236}

	dataBytes, err := TransformCommandValueToDataBytes(commandInfo, val)
	require.NoError(t, err)

	assert.ElementsMatch(t, expected, dataBytes)
}

func TestTransformCommandValueToDataBytes_BOOL(t *testing.T) {
	req := models.CommandRequest{
		DeviceResourceName: "light",
		Type:               common.ValueTypeBool,
		Attributes: map[string]interface{}{
			PRIMARY_TABLE:    COILS,
			STARTING_ADDRESS: 10,
		},
	}
	commandInfo, err := createCommandInfo(&req)
	require.NoError(t, err)
	resTime := time.Now().UnixNano()
	val, err := models.NewCommandValue(req.DeviceResourceName, common.ValueTypeBool, true)
	require.NoError(t, err)
	val.Origin = resTime
	expected := []byte{1}

	dataBytes, err := TransformCommandValueToDataBytes(commandInfo, val)
	require.NoError(t, err)

	assert.ElementsMatch(t, expected, dataBytes)
}

func TestTransformCommandValueToDataBytes_ValueType_FLOAT32_RawType_INT16(t *testing.T) {
	req := models.CommandRequest{
		DeviceResourceName: "light",
		Type:               common.ValueTypeFloat32,
		Attributes: map[string]interface{}{
			PRIMARY_TABLE:    HOLDING_REGISTERS,
			STARTING_ADDRESS: 10,
			RAW_TYPE:         INT16,
		},
	}
	commandInfo, err := createCommandInfo(&req)
	require.NoError(t, err)
	resTime := time.Now().UnixNano()
	val, err := models.NewCommandValue(req.DeviceResourceName, common.ValueTypeFloat32, float32(-52.0))
	require.NoError(t, err)
	val.Origin = resTime
	expected := []byte{255, 204}

	dataBytes, err := TransformCommandValueToDataBytes(commandInfo, val)
	require.NoError(t, err)

	assert.ElementsMatch(t, expected, dataBytes)
}

func TestTransformCommandValueToDataBytes_ValueType_FLOAT32_RawType_UINT16(t *testing.T) {
	req := models.CommandRequest{
		DeviceResourceName: "light",
		Type:               common.ValueTypeFloat32,
		Attributes: map[string]interface{}{
			PRIMARY_TABLE:    HOLDING_REGISTERS,
			STARTING_ADDRESS: 10,
			RAW_TYPE:         UINT16,
		},
	}
	commandInfo, err := createCommandInfo(&req)
	require.NoError(t, err)
	resTime := time.Now().UnixNano()
	val, err := models.NewCommandValue(req.DeviceResourceName, common.ValueTypeFloat32, float32(112.1))
	require.NoError(t, err)
	val.Origin = resTime
	expected := []byte{0, 112}

	dataBytes, err := TransformCommandValueToDataBytes(commandInfo, val)
	require.NoError(t, err)

	assert.ElementsMatch(t, expected, dataBytes)
}

func TestTransformCommandValueToDataBytes_ValueType_FLOAT64_RawType_INT16(t *testing.T) {
	req := models.CommandRequest{
		DeviceResourceName: "light",
		Type:               common.ValueTypeFloat64,
		Attributes: map[string]interface{}{
			PRIMARY_TABLE:    HOLDING_REGISTERS,
			STARTING_ADDRESS: 10,
			RAW_TYPE:         INT16,
		},
	}
	commandInfo, err := createCommandInfo(&req)
	require.NoError(t, err)
	resTime := time.Now().UnixNano()
	val, err := models.NewCommandValue(req.DeviceResourceName, common.ValueTypeFloat64, -52.0)
	require.NoError(t, err)
	val.Origin = resTime
	expected := []byte{255, 204}

	dataBytes, err := TransformCommandValueToDataBytes(commandInfo, val)
	require.NoError(t, err)

	assert.ElementsMatch(t, expected, dataBytes)
}

func TestTransformCommandValueToDataBytes_ValueType_FLOAT64_RawType_UINT16(t *testing.T) {
	req := models.CommandRequest{
		DeviceResourceName: "light",
		Type:               common.ValueTypeFloat64,
		Attributes: map[string]interface{}{
			PRIMARY_TABLE:    HOLDING_REGISTERS,
			STARTING_ADDRESS: 10,
			RAW_TYPE:         UINT16,
		},
	}
	commandInfo, err := createCommandInfo(&req)
	require.NoError(t, err)
	resTime := time.Now().UnixNano()
	val, err := models.NewCommandValue(req.DeviceResourceName, common.ValueTypeFloat64, 112.1)
	require.NoError(t, err)
	val.Origin = resTime
	expected := []byte{0, 112}

	dataBytes, err := TransformCommandValueToDataBytes(commandInfo, val)
	require.NoError(t, err)

	assert.ElementsMatch(t, expected, dataBytes)
}

func TestTransformCommandValueToDataBytes_ValueType_FLOAT64_RawType_UINT32_WordSwap(t *testing.T) {
	req := models.CommandRequest{
		DeviceResourceName: "light",
		Type:               common.ValueTypeFloat64,
		Attributes: map[string]interface{}{
			PRIMARY_TABLE:    HOLDING_REGISTERS,
			STARTING_ADDRESS: 10,
			RAW_TYPE:         UINT32,
			IS_WORD_SWAP:     true,
		},
	}
	commandInfo, err := createCommandInfo(&req)
	require.NoError(t, err)
	val, err := models.NewCommandValue(req.DeviceResourceName, common.ValueTypeFloat64, float64(267))
	require.NoError(t, err)
	dataBytes, err := TransformCommandValueToDataBytes(commandInfo, val)
	require.NoError(t, err)
	assert.Equal(t, []byte{1, 11, 0, 0}, dataBytes)

	commandValue, err := TransformDataBytesToResult(&req, dataBytes, commandInfo)
	require.NoError(t, err)
	result, err := commandValue.Float64Value()
	require.NoError(t, err)
	assert.Equal(t, float64(267), result)
}

func TestTransformCommandValueToDataBytes_ValueType_FLOAT64_RawType_UINT64_IgnoresSwap(t *testing.T) {
	req := models.CommandRequest{
		DeviceResourceName: "light",
		Type:               common.ValueTypeFloat64,
		Attributes: map[string]interface{}{
			PRIMARY_TABLE:    HOLDING_REGISTERS,
			STARTING_ADDRESS: 10,
			RAW_TYPE:         UINT64,
			IS_BYTE_SWAP:     true,
			IS_WORD_SWAP:     true,
		},
	}
	commandInfo, err := createCommandInfo(&req)
	require.NoError(t, err)
	val, err := models.NewCommandValue(req.DeviceResourceName, common.ValueTypeFloat64, float64(4294967296))
	require.NoError(t, err)
	dataBytes, err := TransformCommandValueToDataBytes(commandInfo, val)
	require.NoError(t, err)
	assert.Equal(t, []byte{0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00}, dataBytes)

	commandValue, err := TransformDataBytesToResult(&req, dataBytes, commandInfo)
	require.NoError(t, err)
	result, err := commandValue.Float64Value()
	require.NoError(t, err)
	assert.Equal(t, float64(4294967296), result)
}

// Test swap operation for read command
func TestTransformDataBytesToResultWithByteSwap_INT32(t *testing.T) {
	req := models.CommandRequest{
		DeviceResourceName: "light",
		Type:               common.ValueTypeInt32,
		Attributes: map[string]interface{}{
			PRIMARY_TABLE:    INPUT_REGISTERS,
			STARTING_ADDRESS: 10,
			IS_BYTE_SWAP:     true,
		},
	}
	commandInfo, err := createCommandInfo(&req)
	require.NoError(t, err)
	dataBytes := []byte{0, 0, 11, 1} // bytes swap & big-endian => [11,1,0,0] => 11+2^8=267
	expected := int32(267)

	commandValue, err := TransformDataBytesToResult(&req, dataBytes, commandInfo)
	require.NoError(t, err)
	result, err := commandValue.Int32Value()
	require.NoError(t, err)

	require.Equal(t, expected, result)
}

func TestTransformDataBytesToResultWithWordSwap_INT32(t *testing.T) {
	req := models.CommandRequest{
		DeviceResourceName: "light",
		Type:               common.ValueTypeInt32,
		Attributes: map[string]interface{}{
			PRIMARY_TABLE:    INPUT_REGISTERS,
			STARTING_ADDRESS: 10,
			IS_WORD_SWAP:     true,
		},
	}
	commandInfo, err := createCommandInfo(&req)
	require.NoError(t, err)
	dataBytes := []byte{1, 11, 0, 0} // words swap & big-endian => [11,1,0,0] => 11+2^8=267
	expected := int32(267)

	commandValue, err := TransformDataBytesToResult(&req, dataBytes, commandInfo)
	require.NoError(t, err)
	result, err := commandValue.Int32Value()
	require.NoError(t, err)

	require.Equal(t, expected, result)
}

func TestTransformDataBytesToResultWithByteAndWordSwap_INT32(t *testing.T) {
	req := models.CommandRequest{
		DeviceResourceName: "light",
		Type:               common.ValueTypeInt32,
		Attributes: map[string]interface{}{
			PRIMARY_TABLE:    INPUT_REGISTERS,
			STARTING_ADDRESS: 10,
			IS_BYTE_SWAP:     true,
			IS_WORD_SWAP:     true,
		},
	}
	commandInfo, err := createCommandInfo(&req)
	require.NoError(t, err)
	dataBytes := []byte{11, 1, 0, 0} // bytes and words swap & big-endian => [11,1,0,0] => 11+2^8=267
	expected := int32(267)

	commandValue, err := TransformDataBytesToResult(&req, dataBytes, commandInfo)
	require.NoError(t, err)
	result, err := commandValue.Int32Value()
	require.NoError(t, err)

	require.Equal(t, expected, result)
}

func TestTransformDataBytesToResultWithByteSwap_UINT32(t *testing.T) {
	req := models.CommandRequest{
		DeviceResourceName: "light",
		Type:               common.ValueTypeUint32,
		Attributes: map[string]interface{}{
			PRIMARY_TABLE:    HOLDING_REGISTERS,
			STARTING_ADDRESS: 10,
			IS_BYTE_SWAP:     true,
		},
	}
	commandInfo, err := createCommandInfo(&req)
	require.NoError(t, err)
	dataBytes := []byte{0, 0, 11, 1} // bytes swap & big-endian => [11,1,0,0] => 11+2^8=267
	expected := uint32(267)

	commandValue, err := TransformDataBytesToResult(&req, dataBytes, commandInfo)
	require.NoError(t, err)
	result, err := commandValue.Uint32Value()
	require.NoError(t, err)

	require.Equal(t, expected, result)
}

func TestTransformDataBytesToResultWithWordSwap_UINT32(t *testing.T) {
	req := models.CommandRequest{
		DeviceResourceName: "light",
		Type:               common.ValueTypeUint32,
		Attributes: map[string]interface{}{
			PRIMARY_TABLE:    HOLDING_REGISTERS,
			STARTING_ADDRESS: 10,
			IS_WORD_SWAP:     true,
		},
	}
	commandInfo, err := createCommandInfo(&req)
	require.NoError(t, err)
	dataBytes := []byte{1, 11, 0, 0} // words swap & big-endian => [11,1,0,0] => 11+2^8=267
	expected := uint32(267)

	commandValue, err := TransformDataBytesToResult(&req, dataBytes, commandInfo)
	require.NoError(t, err)
	result, err := commandValue.Uint32Value()
	require.NoError(t, err)

	require.Equal(t, expected, result)
}

func TestTransformDataBytesToResultWithByteAndWordSwap_UINT32(t *testing.T) {
	req := models.CommandRequest{
		DeviceResourceName: "light",
		Type:               common.ValueTypeUint32,
		Attributes: map[string]interface{}{
			PRIMARY_TABLE:    HOLDING_REGISTERS,
			STARTING_ADDRESS: 10,
			IS_BYTE_SWAP:     true,
			IS_WORD_SWAP:     true,
		},
	}
	commandInfo, err := createCommandInfo(&req)
	require.NoError(t, err)
	dataBytes := []byte{11, 1, 0, 0} // bytes and words swap & big-endian => [11,1,0,0] => 11+2^8=267
	expected := uint32(267)

	commandValue, err := TransformDataBytesToResult(&req, dataBytes, commandInfo)
	require.NoError(t, err)
	result, err := commandValue.Uint32Value()
	require.NoError(t, err)

	require.Equal(t, expected, result)
}

func TestTransformDataBytesToResultWithByteSwap_FLOAT32(t *testing.T) {
	req := models.CommandRequest{
		DeviceResourceName: "light",
		Type:               common.ValueTypeFloat32,
		Attributes: map[string]interface{}{
			PRIMARY_TABLE:    HOLDING_REGISTERS,
			STARTING_ADDRESS: 10,
			IS_BYTE_SWAP:     true,
		},
	}
	commandInfo, err := createCommandInfo(&req)
	require.NoError(t, err)
	dataBytes := []byte{143, 63, 41, 92} // bytes swap & big-endian => [41,92,143,63] => 1.12
	expected := float32(1.12)

	commandValue, err := TransformDataBytesToResult(&req, dataBytes, commandInfo)
	require.NoError(t, err)
	result, err := commandValue.Float32Value()
	require.NoError(t, err)

	require.Equal(t, expected, result)
}

func TestTransformDataBytesToResultWithWordSwap_FLOAT32(t *testing.T) {
	req := models.CommandRequest{
		DeviceResourceName: "light",
		Type:               common.ValueTypeFloat32,
		Attributes: map[string]interface{}{
			PRIMARY_TABLE:    HOLDING_REGISTERS,
			STARTING_ADDRESS: 10,
			IS_WORD_SWAP:     true,
		},
	}
	commandInfo, err := createCommandInfo(&req)
	require.NoError(t, err)
	dataBytes := []byte{92, 41, 63, 143} // words swap & big-endian => [41,92,143,63] => 1.12
	expected := float32(1.12)

	commandValue, err := TransformDataBytesToResult(&req, dataBytes, commandInfo)
	require.NoError(t, err)
	result, err := commandValue.Float32Value()
	require.NoError(t, err)

	require.Equal(t, expected, result)
}

func TestTransformDataBytesToResultWithByteAndWordSwap_FLOAT32(t *testing.T) {
	req := models.CommandRequest{
		DeviceResourceName: "light",
		Type:               common.ValueTypeFloat32,
		Attributes: map[string]interface{}{
			PRIMARY_TABLE:    HOLDING_REGISTERS,
			STARTING_ADDRESS: 10,
			IS_BYTE_SWAP:     true,
			IS_WORD_SWAP:     true,
		},
	}
	commandInfo, err := createCommandInfo(&req)
	require.NoError(t, err)
	dataBytes := []byte{41, 92, 143, 63} // bytes and words swap & big-endian => [41,92,143,63] => 1.12
	expected := float32(1.12)

	commandValue, err := TransformDataBytesToResult(&req, dataBytes, commandInfo)
	require.NoError(t, err)
	result, err := commandValue.Float32Value()
	require.NoError(t, err)

	require.Equal(t, expected, result)
}

// Test swap operation for write command
func TestTransformCommandValueToDataBytesWithByteSwap_INT32(t *testing.T) {
	req := models.CommandRequest{
		DeviceResourceName: "light",
		Type:               common.ValueTypeInt32,
		Attributes: map[string]interface{}{
			PRIMARY_TABLE:    HOLDING_REGISTERS,
			STARTING_ADDRESS: 10,
			IS_BYTE_SWAP:     true,
		},
	}
	commandInfo, err := createCommandInfo(&req)
	require.NoError(t, err)
	resTime := time.Now().UnixNano()
	val, err := models.NewCommandValue(req.DeviceResourceName, common.ValueTypeInt32, int32(267))
	require.NoError(t, err)
	val.Origin = resTime
	expected := []byte{0, 0, 11, 1}

	dataBytes, err := TransformCommandValueToDataBytes(commandInfo, val)
	require.NoError(t, err)

	assert.ElementsMatch(t, expected, dataBytes)
}

func TestTransformCommandValueToDataBytesWithWordSwap_INT32(t *testing.T) {
	req := models.CommandRequest{
		DeviceResourceName: "light",
		Type:               common.ValueTypeInt32,
		Attributes: map[string]interface{}{
			PRIMARY_TABLE:    HOLDING_REGISTERS,
			STARTING_ADDRESS: 10,
			IS_WORD_SWAP:     true,
		},
	}
	commandInfo, err := createCommandInfo(&req)
	require.NoError(t, err)
	resTime := time.Now().UnixNano()
	val, err := models.NewCommandValue(req.DeviceResourceName, common.ValueTypeInt32, int32(267))
	require.NoError(t, err)
	val.Origin = resTime
	expected := []byte{1, 11, 0, 0}

	dataBytes, err := TransformCommandValueToDataBytes(commandInfo, val)
	require.NoError(t, err)

	assert.ElementsMatch(t, expected, dataBytes)
}

func TestTransformCommandValueToDataBytesWithByteAndWordSwap_INT32(t *testing.T) {
	req := models.CommandRequest{
		DeviceResourceName: "light",
		Type:               common.ValueTypeInt32,
		Attributes: map[string]interface{}{
			PRIMARY_TABLE:    HOLDING_REGISTERS,
			STARTING_ADDRESS: 10,
			IS_BYTE_SWAP:     true,
			IS_WORD_SWAP:     true,
		},
	}
	commandInfo, err := createCommandInfo(&req)
	require.NoError(t, err)
	resTime := time.Now().UnixNano()
	val, err := models.NewCommandValue(req.DeviceResourceName, common.ValueTypeInt32, int32(267))
	require.NoError(t, err)
	val.Origin = resTime
	expected := []byte{11, 1, 0, 0}

	dataBytes, err := TransformCommandValueToDataBytes(commandInfo, val)
	require.NoError(t, err)

	assert.ElementsMatch(t, expected, dataBytes)
}

func TestTransformCommandValueToDataBytesWithByteSwap_UINT32(t *testing.T) {
	req := models.CommandRequest{
		DeviceResourceName: "light",
		Type:               common.ValueTypeUint32,
		Attributes: map[string]interface{}{
			PRIMARY_TABLE:    HOLDING_REGISTERS,
			STARTING_ADDRESS: 10,
			IS_BYTE_SWAP:     true,
		},
	}
	commandInfo, err := createCommandInfo(&req)
	require.NoError(t, err)
	resTime := time.Now().UnixNano()
	val, err := models.NewCommandValue(req.DeviceResourceName, common.ValueTypeUint32, uint32(267))
	require.NoError(t, err)
	val.Origin = resTime
	expected := []byte{0, 0, 11, 1}

	dataBytes, err := TransformCommandValueToDataBytes(commandInfo, val)
	require.NoError(t, err)

	assert.ElementsMatch(t, expected, dataBytes)
}

func TestTransformCommandValueToDataBytesWithWordSwap_UINT32(t *testing.T) {
	req := models.CommandRequest{
		DeviceResourceName: "light",
		Type:               common.ValueTypeUint32,
		Attributes: map[string]interface{}{
			PRIMARY_TABLE:    HOLDING_REGISTERS,
			STARTING_ADDRESS: 10,
			IS_WORD_SWAP:     true,
		},
	}
	commandInfo, err := createCommandInfo(&req)
	require.NoError(t, err)
	resTime := time.Now().UnixNano()
	val, err := models.NewCommandValue(req.DeviceResourceName, common.ValueTypeUint32, uint32(267))
	require.NoError(t, err)
	val.Origin = resTime
	expected := []byte{1, 11, 0, 0}

	dataBytes, err := TransformCommandValueToDataBytes(commandInfo, val)
	require.NoError(t, err)

	assert.ElementsMatch(t, expected, dataBytes)
}

func TestTransformCommandValueToDataBytesWithByteAndWordSwap_UINT32(t *testing.T) {
	req := models.CommandRequest{
		DeviceResourceName: "light",
		Type:               common.ValueTypeUint32,
		Attributes: map[string]interface{}{
			PRIMARY_TABLE:    HOLDING_REGISTERS,
			STARTING_ADDRESS: 10,
			IS_BYTE_SWAP:     true,
			IS_WORD_SWAP:     true,
		},
	}
	commandInfo, err := createCommandInfo(&req)
	require.NoError(t, err)
	resTime := time.Now().UnixNano()
	val, err := models.NewCommandValue(req.DeviceResourceName, common.ValueTypeUint32, uint32(267))
	require.NoError(t, err)
	val.Origin = resTime
	expected := []byte{11, 1, 0, 0}

	dataBytes, err := TransformCommandValueToDataBytes(commandInfo, val)
	require.NoError(t, err)

	assert.ElementsMatch(t, expected, dataBytes)
}

func TestTransformCommandValueToDataBytesWithByteSwap_FLOAT32(t *testing.T) {
	req := models.CommandRequest{
		DeviceResourceName: "light",
		Type:               common.ValueTypeFloat32,
		Attributes: map[string]interface{}{
			PRIMARY_TABLE:    HOLDING_REGISTERS,
			STARTING_ADDRESS: 10,
			IS_BYTE_SWAP:     true,
		},
	}
	commandInfo, err := createCommandInfo(&req)
	require.NoError(t, err)
	resTime := time.Now().UnixNano()
	val, err := models.NewCommandValue(req.DeviceResourceName, common.ValueTypeFloat32, float32(1.12))
	require.NoError(t, err)
	val.Origin = resTime
	expected := []byte{143, 63, 41, 92}

	dataBytes, err := TransformCommandValueToDataBytes(commandInfo, val)
	require.NoError(t, err)

	assert.ElementsMatch(t, expected, dataBytes)
}

func TestTransformCommandValueToDataBytesWithWordSwap_FLOAT32(t *testing.T) {
	req := models.CommandRequest{
		DeviceResourceName: "light",
		Type:               common.ValueTypeFloat32,
		Attributes: map[string]interface{}{
			PRIMARY_TABLE:    HOLDING_REGISTERS,
			STARTING_ADDRESS: 10,
			IS_WORD_SWAP:     true,
		},
	}
	commandInfo, err := createCommandInfo(&req)
	require.NoError(t, err)
	resTime := time.Now().UnixNano()
	val, err := models.NewCommandValue(req.DeviceResourceName, common.ValueTypeFloat32, float32(1.12))
	require.NoError(t, err)
	val.Origin = resTime
	expected := []byte{92, 41, 63, 143}

	dataBytes, err := TransformCommandValueToDataBytes(commandInfo, val)
	require.NoError(t, err)

	assert.ElementsMatch(t, expected, dataBytes)
}

func TestTransformCommandValueToDataBytesWithByteAndWordSwap_FLOAT32(t *testing.T) {
	req := models.CommandRequest{
		DeviceResourceName: "light",
		Type:               common.ValueTypeFloat32,
		Attributes: map[string]interface{}{
			PRIMARY_TABLE:    HOLDING_REGISTERS,
			STARTING_ADDRESS: 10,
			IS_BYTE_SWAP:     true,
			IS_WORD_SWAP:     true,
		},
	}
	commandInfo, err := createCommandInfo(&req)
	require.NoError(t, err)
	resTime := time.Now().UnixNano()
	val, err := models.NewCommandValue(req.DeviceResourceName, common.ValueTypeFloat32, float32(1.12))
	require.NoError(t, err)
	val.Origin = resTime
	expected := []byte{41, 92, 143, 63}

	dataBytes, err := TransformCommandValueToDataBytes(commandInfo, val)
	require.NoError(t, err)

	assert.ElementsMatch(t, expected, dataBytes)
}
