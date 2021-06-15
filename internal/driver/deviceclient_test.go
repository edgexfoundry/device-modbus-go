// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"testing"
	"time"

	"github.com/edgexfoundry/device-sdk-go/v2/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"

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
	commandInfo, err := createCommandInfo(&req)
	require.NoError(t, err)
	dataBytes := []byte{1} // => 00000001
	expected := true

	commandValue, err := TransformDataBytesToResult(&req, dataBytes, commandInfo)
	require.NoError(t, err)
	result, err := commandValue.BoolValue()
	require.NoError(t, err)

	assert.Equal(t, expected, result)
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
