// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018-2019 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"encoding/binary"
	"fmt"
	"math"
	"strings"
	"time"

	sdkModel "github.com/edgexfoundry/device-sdk-go/pkg/models"
)

// DeviceClient is a interface for modbus client lib to implementation
// It's responsibility are handle connection, read data bytes value and write data bytes value
type DeviceClient interface {
	OpenConnection() error
	GetValue(commandInfo interface{}) ([]byte, error)
	SetValue(commandInfo interface{}, value []byte) error
	CloseConnection() error
}

// ConnectionInfo is command info
type CommandInfo struct {
	PrimaryTable    string
	StartingAddress uint16
	ValueType       sdkModel.ValueType
	// how many register need to read
	Length     uint16
	IsByteSwap bool
	IsWordSwap bool
}

func createCommandInfo(req *sdkModel.CommandRequest) *CommandInfo {
	primaryTable, _ := req.Attributes["primaryTable"]
	primaryTable = strings.ToUpper(primaryTable)

	startingAddress, _ := toUint16(req.Attributes["startingAddress"])
	startingAddress = startingAddress - 1

	length := calculateAddressLength(primaryTable, req.Type)

	var isByteSwap = false
	_, ok := req.Attributes["isByteSwap"]
	if ok {
		isByteSwap, _ = toBool(req.Attributes["isByteSwap"])
	}

	var isWordSwap = false
	_, ok = req.Attributes["isWordSwap"]
	if ok {
		isWordSwap, _ = toBool(req.Attributes["isWordSwap"])
	}

	return &CommandInfo{
		PrimaryTable:    primaryTable,
		StartingAddress: startingAddress,
		ValueType:       req.Type,
		Length:          length,
		IsByteSwap:      isByteSwap,
		IsWordSwap:      isWordSwap,
	}
}

func calculateAddressLength(primaryTable string, valueType sdkModel.ValueType) uint16 {
	var primaryTableBit = PrimaryTableBitCountMap[primaryTable]
	var valueTypeBitCount = ValueTypeBitCountMap[valueType]

	var length = valueTypeBitCount / primaryTableBit
	if length < 1 {
		length = 1
	}

	return length
}

func TransformDateBytesToResult(req *sdkModel.CommandRequest, dataBytes []byte, commandInfo *CommandInfo) (*sdkModel.CommandValue, error) {
	var result = &sdkModel.CommandValue{}
	var stringResult string
	var err error
	var resTime = time.Now().UnixNano()

	switch commandInfo.ValueType {
	case sdkModel.Uint16:
		var res = binary.BigEndian.Uint16(dataBytes)
		result, err = sdkModel.NewUint16Value(req.DeviceResourceName, resTime, res)
		stringResult = fmt.Sprint(res)
	case sdkModel.Uint32:
		var res = binary.BigEndian.Uint32(swap32BitDataBytes(dataBytes, commandInfo.IsByteSwap, commandInfo.IsWordSwap))
		result, err = sdkModel.NewUint32Value(req.DeviceResourceName, resTime, res)
		stringResult = fmt.Sprint(res)
	case sdkModel.Uint64:
		var res = binary.BigEndian.Uint64(dataBytes)
		result, err = sdkModel.NewUint64Value(req.DeviceResourceName, resTime, res)
		stringResult = fmt.Sprint(res)

	case sdkModel.Int16:
		var res = int16(binary.BigEndian.Uint16(dataBytes))
		result, err = sdkModel.NewInt16Value(req.DeviceResourceName, resTime, res)
		stringResult = fmt.Sprint(res)
	case sdkModel.Int32:
		var res = int32(binary.BigEndian.Uint32(swap32BitDataBytes(dataBytes, commandInfo.IsByteSwap, commandInfo.IsWordSwap)))
		result, err = sdkModel.NewInt32Value(req.DeviceResourceName, resTime, res)
		stringResult = fmt.Sprint(res)
	case sdkModel.Int64:
		var res = int64(binary.BigEndian.Uint64(dataBytes))
		result, err = sdkModel.NewInt64Value(req.DeviceResourceName, resTime, res)
		stringResult = fmt.Sprint(res)

	case sdkModel.Float32:
		var res = binary.BigEndian.Uint32(dataBytes)
		var floatResult = math.Float32frombits(res)
		result, err = sdkModel.NewFloat32Value(req.DeviceResourceName, resTime, floatResult)
		stringResult = fmt.Sprint(floatResult)
	case sdkModel.Float64:
		var res = binary.BigEndian.Uint64(dataBytes)
		var floatResult = math.Float64frombits(res)
		result, err = sdkModel.NewFloat64Value(req.DeviceResourceName, resTime, floatResult)
		stringResult = fmt.Sprint(floatResult)

	case sdkModel.Bool:
		var res = false
		// to find the 1st bit of the dataBytes by mask it with 2^0 = 1 (00000001)
		if (dataBytes[0] & 1) > 0 {
			res = true
		}
		result, err = sdkModel.NewBoolValue(req.DeviceResourceName, resTime, res)
		stringResult = fmt.Sprint(res)

	default:
		err = fmt.Errorf("return result fail, none supported value type: %v", commandInfo.ValueType)
	}

	driver.Logger.Info(fmt.Sprintf("Transfer dataBytes to CommandValue(%v, %v) successful.", commandInfo.ValueType, stringResult))
	return result, err
}

func TransformCommandValueToDataBytes(commandInfo *CommandInfo, value *sdkModel.CommandValue) ([]byte, error) {
	var byteCount = calculateByteCount(commandInfo)
	var err error
	var maxSize = uint16(len(value.NumericValue))
	var dataBytes = value.NumericValue[maxSize-byteCount : maxSize]

	_, ok := ValueTypeBitCountMap[commandInfo.ValueType]
	if !ok {
		err = fmt.Errorf("none supported value type : %v \n", commandInfo.ValueType)
		return dataBytes, err
	}

	if commandInfo.ValueType == sdkModel.Uint32 || commandInfo.ValueType == sdkModel.Int32 {
		dataBytes = swap32BitDataBytes(dataBytes, commandInfo.IsByteSwap, commandInfo.IsWordSwap)
	}

	driver.Logger.Info(fmt.Sprintf("Transfer CommandValue to dataBytes for write command, %v, %v", commandInfo.ValueType, dataBytes))
	return dataBytes, err
}

func calculateByteCount(commandInfo *CommandInfo) uint16 {
	var byteCount uint16
	if commandInfo.PrimaryTable == HOLDING_REGISTERS || commandInfo.PrimaryTable == INPUT_REGISTERS {
		byteCount = commandInfo.Length * 2
	} else {
		byteCount = commandInfo.Length
	}

	return byteCount
}

func swap32BitDataBytes(dataBytes []byte, isByteSwap bool, isWordSwap bool) []byte {

	if !isByteSwap && !isWordSwap {
		return dataBytes
	}

	if len(dataBytes) < 4 {
		return dataBytes
	}

	var newDataBytes = make([]byte, len(dataBytes))

	if isByteSwap {
		newDataBytes[0] = dataBytes[1]
		newDataBytes[1] = dataBytes[0]
		newDataBytes[2] = dataBytes[3]
		newDataBytes[3] = dataBytes[2]
	}
	if isWordSwap {
		newDataBytes[0] = dataBytes[2]
		newDataBytes[1] = dataBytes[3]
		newDataBytes[2] = dataBytes[0]
		newDataBytes[3] = dataBytes[1]
	}

	return newDataBytes
}
