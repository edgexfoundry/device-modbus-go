// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018-2019 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"bytes"
	"testing"
	"time"

	sdkModel "github.com/edgexfoundry/device-sdk-go/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

func init() {
	driver = new(Driver)
	driver.Logger = logger.NewClient("test", false, "", "DEBUG")
}

// Test byte, word swap
func TestByteSwap32BitDataBytes(t *testing.T) {
	var isByteSwap = true
	var isWordSwap = false
	var dataBytes = []byte{6, 9, 12, 15}

	var expectResult = swap32BitDataBytes(dataBytes, isByteSwap, isWordSwap)

	if expectResult[0] != 9 || expectResult[1] != 6 || expectResult[2] != 15 || expectResult[3] != 12 {
		t.Fatalf("Swap32BitDataBytes 32 bits failed! ")
	}
}

func TestWordSwap32BitDataBytes(t *testing.T) {
	var isByteSwap = false
	var isWordSwap = true
	var dataBytes = []byte{6, 9, 12, 15}

	var expectResult = swap32BitDataBytes(dataBytes, isByteSwap, isWordSwap)

	if expectResult[0] != 12 || expectResult[1] != 15 || expectResult[2] != 6 || expectResult[3] != 9 {
		t.Fatalf("Swap32BitDataBytes 32 bits failed! ")
	}
}

func TestTransformDateBytesToResult_INT16(t *testing.T) {
	deviceObject := models.DeviceResource{
		Attributes: map[string]interface{}{
			"primaryTable":    INPUT_REGISTERS,
			"startingAddress": "10",
		},
		Properties: models.ProfileProperty{
			Value: models.PropertyValue{
				Type: INT16,
			},
		},
	}
	commandInfo := createCommandInfo(deviceObject)
	ro := models.ResourceOperation{}
	dataBytes := []byte{255, 231} // => big-endian [231,255] => -25

	expected := int16(-25)

	commandValue, err := TransformDateBytesToResult(&ro, dataBytes, commandInfo)

	if err != nil {
		t.Fatalf("Fail to tramsform data bytes to result. Error: %v", err)
	}
	result, err := commandValue.Int16Value()
	if err != nil || expected != result {
		t.Fatalf("Unexpected result. Error: %v", err)
	}
}

func TestTransformDateBytesToResult_INT32(t *testing.T) {
	deviceObject := models.DeviceResource{
		Attributes: map[string]interface{}{
			"primaryTable":    INPUT_REGISTERS,
			"startingAddress": "10",
		},
		Properties: models.ProfileProperty{
			Value: models.PropertyValue{
				Type: INT32,
			},
		},
	}
	commandInfo := createCommandInfo(deviceObject)
	ro := models.ResourceOperation{}
	dataBytes := []byte{0, 0, 1, 11} // big-endian [11,1,0,0] => 11+2^8=267

	expected := int32(267)

	commandValue, err := TransformDateBytesToResult(&ro, dataBytes, commandInfo)
	if err != nil {
		t.Fatalf("Fail to tramsform data bytes to result. Error: %v", err)
	}
	result, err := commandValue.Int32Value()
	if err != nil || expected != result {
		t.Fatalf("Unexpected result. Error: %v", err)
	}
}

func TestTransformDateBytesToResult_INT64(t *testing.T) {
	deviceObject := models.DeviceResource{
		Attributes: map[string]interface{}{
			"primaryTable":    HOLDING_REGISTERS,
			"startingAddress": "10",
		},
		Properties: models.ProfileProperty{
			Value: models.PropertyValue{
				Type: INT64,
			},
		},
	}
	commandInfo := createCommandInfo(deviceObject)
	ro := models.ResourceOperation{}
	dataBytes := []byte{0, 1, 0, 0, 0, 2, 1, 1} // big-endian [1,1,2,0,0,0,1,0] => 1+2^8+2^17+2^48=281474976841985

	expected := int64(281474976841985)

	commandValue, err := TransformDateBytesToResult(&ro, dataBytes, commandInfo)
	if err != nil {
		t.Fatalf("Fail to tramsform data bytes to result. Error: %v", err)
	}
	result, err := commandValue.Int64Value()
	if err != nil || expected != result {
		t.Fatalf("Unexpected result. Error: %v", err)
	}
}

func TestTransformDateBytesToResult_UINT16(t *testing.T) {
	deviceObject := models.DeviceResource{
		Attributes: map[string]interface{}{
			"primaryTable":    HOLDING_REGISTERS,
			"startingAddress": "10",
		},
		Properties: models.ProfileProperty{
			Value: models.PropertyValue{
				Type: UINT16,
			},
		},
	}
	commandInfo := createCommandInfo(deviceObject)
	ro := models.ResourceOperation{}
	dataBytes := []byte{0, 11} // => big-endian [11,0] => 11

	expected := uint16(11)

	commandValue, err := TransformDateBytesToResult(&ro, dataBytes, commandInfo)

	if err != nil {
		t.Fatalf("Fail to tramsform data bytes to result. Error: %v", err)
	}
	result, err := commandValue.Uint16Value()
	if err != nil || expected != result {
		t.Fatalf("Unexpected result. Error: %v", err)
	}
}

func TestTransformDateBytesToResult_UINT32(t *testing.T) {
	deviceObject := models.DeviceResource{
		Attributes: map[string]interface{}{
			"primaryTable":    HOLDING_REGISTERS,
			"startingAddress": "10",
		},
		Properties: models.ProfileProperty{
			Value: models.PropertyValue{
				Type: UINT32,
			},
		},
	}
	commandInfo := createCommandInfo(deviceObject)
	ro := models.ResourceOperation{}
	dataBytes := []byte{0, 0, 1, 11} // big-endian [11,1,0,0] => 11+2^8=267

	expected := uint32(267)

	commandValue, err := TransformDateBytesToResult(&ro, dataBytes, commandInfo)
	if err != nil {
		t.Fatalf("Fail to tramsform data bytes to result. Error: %v", err)
	}
	result, err := commandValue.Uint32Value()
	if err != nil || expected != result {
		t.Fatalf("Unexpected result. Error: %v", err)
	}
}

func TestTransformDateBytesToResult_UINT64(t *testing.T) {
	deviceObject := models.DeviceResource{
		Attributes: map[string]interface{}{
			"primaryTable":    HOLDING_REGISTERS,
			"startingAddress": "10",
		},
		Properties: models.ProfileProperty{
			Value: models.PropertyValue{
				Type: UINT64,
			},
		},
	}
	commandInfo := createCommandInfo(deviceObject)
	ro := models.ResourceOperation{}
	dataBytes := []byte{0, 1, 0, 0, 0, 2, 1, 1} // big-endian [1,1,2,0,0,0,1,0] => 1+2^8+2^17+2^48=281474976841985

	expected := uint64(281474976841985)

	commandValue, err := TransformDateBytesToResult(&ro, dataBytes, commandInfo)
	if err != nil {
		t.Fatalf("Fail to tramsform data bytes to result. Error: %v", err)
	}
	result, err := commandValue.Uint64Value()
	if err != nil || expected != result {
		t.Fatalf("Unexpected result. Error: %v", err)
	}
}

func TestTransformDateBytesToResult_FLOAT32(t *testing.T) {
	deviceObject := models.DeviceResource{
		Attributes: map[string]interface{}{
			"primaryTable":    HOLDING_REGISTERS,
			"startingAddress": "10",
		},
		Properties: models.ProfileProperty{
			Value: models.PropertyValue{
				Type: FLOAT32,
			},
		},
	}
	commandInfo := createCommandInfo(deviceObject)
	ro := models.ResourceOperation{}
	dataBytes := []byte{63, 143, 92, 41} // big-endian [41,92,143,63] => 1.12

	expected := float32(1.12)

	commandValue, err := TransformDateBytesToResult(&ro, dataBytes, commandInfo)
	if err != nil {
		t.Fatalf("Fail to tramsform data bytes to result. Error: %v", err)
	}
	result, err := commandValue.Float32Value()
	if err != nil || expected != result {
		t.Fatalf("Unexpected result. Error: %v", err)
	}
}

func TestTransformDateBytesToResult_FLOAT64(t *testing.T) {
	deviceObject := models.DeviceResource{
		Attributes: map[string]interface{}{
			"primaryTable":    HOLDING_REGISTERS,
			"startingAddress": "10",
		},
		Properties: models.ProfileProperty{
			Value: models.PropertyValue{
				Type: FLOAT64,
			},
		},
	}
	commandInfo := createCommandInfo(deviceObject)
	ro := models.ResourceOperation{}
	dataBytes := []byte{63, 241, 235, 133, 30, 184, 81, 236} // => 1.12

	expected := float64(1.12)

	commandValue, err := TransformDateBytesToResult(&ro, dataBytes, commandInfo)
	if err != nil {
		t.Fatalf("Fail to tramsform data bytes to result. Error: %v", err)
	}
	result, err := commandValue.Float64Value()
	if err != nil || expected != result {
		t.Fatalf("Unexpected result. Error: %v", err)
	}
}

func TestTransformDateBytesToResult_BOOL(t *testing.T) {
	deviceObject := models.DeviceResource{
		Attributes: map[string]interface{}{
			"primaryTable":    DISCRETES_INPUT,
			"startingAddress": "10",
		},
		Properties: models.ProfileProperty{
			Value: models.PropertyValue{
				Type: BOOL,
			},
		},
	}
	commandInfo := createCommandInfo(deviceObject)
	ro := models.ResourceOperation{}
	dataBytes := []byte{1} // => 00000001

	expected := true

	commandValue, err := TransformDateBytesToResult(&ro, dataBytes, commandInfo)
	if err != nil {
		t.Fatalf("Fail to tramsform data bytes to result. Error: %v", err)
	}
	result, err := commandValue.BoolValue()
	if err != nil || expected != result {
		t.Fatalf("Unexpected result. Error: %v", err)
	}
}

func TestTransformCommandValueToDataBytes_INT16(t *testing.T) {
	deviceObject := models.DeviceResource{
		Attributes: map[string]interface{}{
			"primaryTable":    HOLDING_REGISTERS,
			"startingAddress": "10",
		},
		Properties: models.ProfileProperty{
			Value: models.PropertyValue{
				Type: INT16,
			},
		},
	}
	commandInfo := createCommandInfo(deviceObject)

	ro := models.ResourceOperation{}
	resTime := time.Now().UnixNano() / int64(time.Millisecond)
	val, _ := sdkModel.NewInt16Value(&ro, resTime, -25)

	expected := []byte{255, 231}

	dataBytes, err := TransformCommandValueToDataBytes(commandInfo, val)

	if err != nil || !bytes.Equal(dataBytes, expected) {
		t.Fatalf("Fail to tramsform command value to data bytes. Error: %v", err)
	}
}

func TestTransformCommandValueToDataBytes_INT32(t *testing.T) {
	deviceObject := models.DeviceResource{
		Attributes: map[string]interface{}{
			"primaryTable":    HOLDING_REGISTERS,
			"startingAddress": "10",
		},
		Properties: models.ProfileProperty{
			Value: models.PropertyValue{
				Type: INT32,
			},
		},
	}
	commandInfo := createCommandInfo(deviceObject)

	ro := models.ResourceOperation{}
	resTime := time.Now().UnixNano() / int64(time.Millisecond)
	val, _ := sdkModel.NewInt32Value(&ro, resTime, 267)

	expected := []byte{0, 0, 1, 11}

	dataBytes, err := TransformCommandValueToDataBytes(commandInfo, val)

	if err != nil || !bytes.Equal(dataBytes, expected) {
		t.Fatalf("Fail to tramsform command value to data bytes. Error: %v", err)
	}
}

func TestTransformCommandValueToDataBytes_INT64(t *testing.T) {
	deviceObject := models.DeviceResource{
		Attributes: map[string]interface{}{
			"primaryTable":    HOLDING_REGISTERS,
			"startingAddress": "10",
		},
		Properties: models.ProfileProperty{
			Value: models.PropertyValue{
				Type: INT64,
			},
		},
	}
	commandInfo := createCommandInfo(deviceObject)

	ro := models.ResourceOperation{}
	resTime := time.Now().UnixNano() / int64(time.Millisecond)
	val, _ := sdkModel.NewInt64Value(&ro, resTime, 281474976841985)

	expected := []byte{0, 1, 0, 0, 0, 2, 1, 1}

	dataBytes, err := TransformCommandValueToDataBytes(commandInfo, val)

	if err != nil || !bytes.Equal(dataBytes, expected) {
		t.Fatalf("Fail to tramsform command value to data bytes. Error: %v", err)
	}
}

func TestTransformCommandValueToDataBytes_UINT16(t *testing.T) {
	deviceObject := models.DeviceResource{
		Attributes: map[string]interface{}{
			"primaryTable":    HOLDING_REGISTERS,
			"startingAddress": "10",
		},
		Properties: models.ProfileProperty{
			Value: models.PropertyValue{
				Type: UINT16,
			},
		},
	}
	commandInfo := createCommandInfo(deviceObject)

	ro := models.ResourceOperation{}
	resTime := time.Now().UnixNano() / int64(time.Millisecond)
	val, _ := sdkModel.NewUint16Value(&ro, resTime, 11)

	expected := []byte{0, 11}

	dataBytes, err := TransformCommandValueToDataBytes(commandInfo, val)

	if err != nil || !bytes.Equal(dataBytes, expected) {
		t.Fatalf("Fail to tramsform command value to data bytes. Error: %v", err)
	}
}

func TestTransformCommandValueToDataBytes_UINT32(t *testing.T) {
	deviceObject := models.DeviceResource{
		Attributes: map[string]interface{}{
			"primaryTable":    HOLDING_REGISTERS,
			"startingAddress": "10",
		},
		Properties: models.ProfileProperty{
			Value: models.PropertyValue{
				Type: UINT32,
			},
		},
	}
	commandInfo := createCommandInfo(deviceObject)

	ro := models.ResourceOperation{}
	resTime := time.Now().UnixNano() / int64(time.Millisecond)
	val, _ := sdkModel.NewUint32Value(&ro, resTime, 267)

	expected := []byte{0, 0, 1, 11}

	dataBytes, err := TransformCommandValueToDataBytes(commandInfo, val)

	if err != nil || !bytes.Equal(dataBytes, expected) {
		t.Fatalf("Fail to tramsform command value to data bytes. Error: %v", err)
	}
}

func TestTransformCommandValueToDataBytes_UINT64(t *testing.T) {
	deviceObject := models.DeviceResource{
		Attributes: map[string]interface{}{
			"primaryTable":    HOLDING_REGISTERS,
			"startingAddress": "10",
		},
		Properties: models.ProfileProperty{
			Value: models.PropertyValue{
				Type: UINT64,
			},
		},
	}
	commandInfo := createCommandInfo(deviceObject)

	ro := models.ResourceOperation{}
	resTime := time.Now().UnixNano() / int64(time.Millisecond)
	val, _ := sdkModel.NewUint64Value(&ro, resTime, 281474976841985)

	expected := []byte{0, 1, 0, 0, 0, 2, 1, 1}

	dataBytes, err := TransformCommandValueToDataBytes(commandInfo, val)

	if err != nil || !bytes.Equal(dataBytes, expected) {
		t.Fatalf("Fail to tramsform command value to data bytes. Error: %v", err)
	}
}

func TestTransformCommandValueToDataBytes_FLOAT32(t *testing.T) {
	deviceObject := models.DeviceResource{
		Attributes: map[string]interface{}{
			"primaryTable":    HOLDING_REGISTERS,
			"startingAddress": "10",
		},
		Properties: models.ProfileProperty{
			Value: models.PropertyValue{
				Type: FLOAT32,
			},
		},
	}
	commandInfo := createCommandInfo(deviceObject)

	ro := models.ResourceOperation{}
	resTime := time.Now().UnixNano() / int64(time.Millisecond)
	val, _ := sdkModel.NewFloat32Value(&ro, resTime, 1.12)

	expected := []byte{63, 143, 92, 41}

	dataBytes, err := TransformCommandValueToDataBytes(commandInfo, val)

	if err != nil || !bytes.Equal(dataBytes, expected) {
		t.Fatalf("Fail to tramsform command value to data bytes. Error: %v", err)
	}
}

func TestTransformCommandValueToDataBytes_FLOAT64(t *testing.T) {
	deviceObject := models.DeviceResource{
		Attributes: map[string]interface{}{
			"primaryTable":    HOLDING_REGISTERS,
			"startingAddress": "10",
		},
		Properties: models.ProfileProperty{
			Value: models.PropertyValue{
				Type: FLOAT64,
			},
		},
	}
	commandInfo := createCommandInfo(deviceObject)

	ro := models.ResourceOperation{}
	resTime := time.Now().UnixNano() / int64(time.Millisecond)
	val, _ := sdkModel.NewFloat64Value(&ro, resTime, 1.12)

	expected := []byte{63, 241, 235, 133, 30, 184, 81, 236}

	dataBytes, err := TransformCommandValueToDataBytes(commandInfo, val)

	if err != nil || !bytes.Equal(dataBytes, expected) {
		t.Fatalf("Fail to tramsform command value to data bytes. Error: %v", err)
	}
}

func TestTransformCommandValueToDataBytes_BOOL(t *testing.T) {
	deviceObject := models.DeviceResource{
		Attributes: map[string]interface{}{
			"primaryTable":    COILS,
			"startingAddress": "10",
		},
		Properties: models.ProfileProperty{
			Value: models.PropertyValue{
				Type: BOOL,
			},
		},
	}
	commandInfo := createCommandInfo(deviceObject)

	ro := models.ResourceOperation{}
	resTime := time.Now().UnixNano() / int64(time.Millisecond)
	val, _ := sdkModel.NewBoolValue(&ro, resTime, true)

	expected := []byte{1}

	dataBytes, err := TransformCommandValueToDataBytes(commandInfo, val)

	if err != nil || !bytes.Equal(dataBytes, expected) {
		t.Fatalf("Fail to tramsform command value to data bytes. Error: %v", err)
	}
}
