// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018-2019 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"strings"
	"time"

	sdkModel "github.com/edgexfoundry/device-sdk-go/pkg/models"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
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
	RawType    sdkModel.ValueType
}

func createCommandInfo(req *sdkModel.CommandRequest) (*CommandInfo, error) {
	primaryTable, _ := req.Attributes[PRIMARY_TABLE]
	primaryTable = strings.ToUpper(primaryTable)

	startingAddress, _ := toUint16(req.Attributes[STARTING_ADDRESS])
	startingAddress = startingAddress - 1

	rawType, err := getRawType(req)
	if err != nil {
		return nil, err
	}
	length := calculateAddressLength(primaryTable, rawType)

	var isByteSwap = false
	_, ok := req.Attributes[IS_BYTE_SWAP]
	if ok {
		isByteSwap, _ = toBool(req.Attributes[IS_BYTE_SWAP])
	}

	var isWordSwap = false
	_, ok = req.Attributes[IS_WORD_SWAP]
	if ok {
		isWordSwap, _ = toBool(req.Attributes[IS_WORD_SWAP])
	}

	return &CommandInfo{
		PrimaryTable:    primaryTable,
		StartingAddress: startingAddress,
		ValueType:       req.Type,
		Length:          length,
		IsByteSwap:      isByteSwap,
		IsWordSwap:      isWordSwap,
		RawType:         rawType,
	}, nil
}

func getRawType(req *sdkModel.CommandRequest) (valueType sdkModel.ValueType, err error) {
	rawType, ok := req.Attributes[RAW_TYPE]
	if !ok || rawType == "" {
		return req.Type, err
	}

	switch rawType {
	case UINT16:
		valueType = sdkModel.Uint16
	case INT16:
		valueType = sdkModel.Int16
	default:
		driver.Logger.Warn(fmt.Sprintf("The raw type %v is not supportted, use value type %v instead", rawType, req.Type))
		err = fmt.Errorf("the raw type %v is not supportted", rawType)
	}
	return valueType, err
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

func TransformDataBytesToResult(req *sdkModel.CommandRequest, dataBytes []byte, commandInfo *CommandInfo) (*sdkModel.CommandValue, error) {
	var result = &sdkModel.CommandValue{}
	var err error
	var resTime = time.Now().UnixNano()

	switch commandInfo.ValueType {
	case sdkModel.Uint16:
		var res = binary.BigEndian.Uint16(dataBytes)
		result, err = sdkModel.NewUint16Value(req.DeviceResourceName, resTime, res)
	case sdkModel.Uint32:
		var res = binary.BigEndian.Uint32(swap32BitDataBytes(dataBytes, commandInfo.IsByteSwap, commandInfo.IsWordSwap))
		result, err = sdkModel.NewUint32Value(req.DeviceResourceName, resTime, res)
	case sdkModel.Uint64:
		var res = binary.BigEndian.Uint64(dataBytes)
		result, err = sdkModel.NewUint64Value(req.DeviceResourceName, resTime, res)
	case sdkModel.Int16:
		var res = int16(binary.BigEndian.Uint16(dataBytes))
		result, err = sdkModel.NewInt16Value(req.DeviceResourceName, resTime, res)
	case sdkModel.Int32:
		var res = int32(binary.BigEndian.Uint32(swap32BitDataBytes(dataBytes, commandInfo.IsByteSwap, commandInfo.IsWordSwap)))
		result, err = sdkModel.NewInt32Value(req.DeviceResourceName, resTime, res)
	case sdkModel.Int64:
		var res = int64(binary.BigEndian.Uint64(dataBytes))
		result, err = sdkModel.NewInt64Value(req.DeviceResourceName, resTime, res)
	case sdkModel.Float32:
		switch commandInfo.RawType {
		case sdkModel.Float32:
			var res = binary.BigEndian.Uint32(swap32BitDataBytes(dataBytes, commandInfo.IsByteSwap, commandInfo.IsWordSwap))
			var floatResult = math.Float32frombits(res)
			result, err = sdkModel.NewFloat32Value(req.DeviceResourceName, resTime, floatResult)
		case sdkModel.Int16:
			var res = int16(binary.BigEndian.Uint16(dataBytes))
			result, err = sdkModel.NewFloat32Value(req.DeviceResourceName, resTime, float32(res))
			driver.Logger.Debug(fmt.Sprintf("According to the rawType %s and the value type %s, convert integer %d to float %v ", INT16, FLOAT32, res, result.ValueToString(contract.ENotation)))
		case sdkModel.Uint16:
			var res = binary.BigEndian.Uint16(dataBytes)
			result, err = sdkModel.NewFloat32Value(req.DeviceResourceName, resTime, float32(res))
			driver.Logger.Debug(fmt.Sprintf("According to the rawType %s and the value type %s, convert integer %d to float %v ", UINT16, FLOAT32, res, result.ValueToString(contract.ENotation)))
		}
	case sdkModel.Float64:
		switch commandInfo.RawType {
		case sdkModel.Float64:
			var res = binary.BigEndian.Uint64(dataBytes)
			var floatResult = math.Float64frombits(res)
			result, err = sdkModel.NewFloat64Value(req.DeviceResourceName, resTime, floatResult)
		case sdkModel.Int16:
			var res = int16(binary.BigEndian.Uint16(dataBytes))
			result, err = sdkModel.NewFloat64Value(req.DeviceResourceName, resTime, float64(res))
			driver.Logger.Debug(fmt.Sprintf("According to the rawType %s and the value type %s, convert integer %d to float %v ", INT16, FLOAT64, res, result.ValueToString(contract.ENotation)))
		case sdkModel.Uint16:
			var res = binary.BigEndian.Uint16(dataBytes)
			result, err = sdkModel.NewFloat64Value(req.DeviceResourceName, resTime, float64(res))
			driver.Logger.Debug(fmt.Sprintf("According to the rawType %s and the value type %s, convert integer %d to float %v ", UINT16, FLOAT64, res, result.ValueToString(contract.ENotation)))
		}
	case sdkModel.Bool:
		var res = false
		// to find the 1st bit of the dataBytes by mask it with 2^0 = 1 (00000001)
		if (dataBytes[0] & 1) > 0 {
			res = true
		}
		result, err = sdkModel.NewBoolValue(req.DeviceResourceName, resTime, res)
	default:
		err = fmt.Errorf("return result fail, none supported value type: %v", commandInfo.ValueType)
	}
	driver.Logger.Debug(fmt.Sprintf("Transfer dataBytes to CommandValue(%v) successful.", result.ValueToString(contract.ENotation)))
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

	if commandInfo.ValueType == sdkModel.Uint32 || commandInfo.ValueType == sdkModel.Int32 || commandInfo.ValueType == sdkModel.Float32 {
		dataBytes = swap32BitDataBytes(dataBytes, commandInfo.IsByteSwap, commandInfo.IsWordSwap)
	}

	// Cast value according to the rawType, this feature only converts float value to integer 32bit value
	if commandInfo.ValueType == sdkModel.Float32 {
		val, err := value.Float32Value()
		if err != nil {
			return dataBytes, err
		}
		if commandInfo.RawType == sdkModel.Int16 {
			dataBytes, err = getBinaryData(int16(val))
			if err != nil {
				return dataBytes, err
			}
		} else if commandInfo.RawType == sdkModel.Uint16 {
			dataBytes, err = getBinaryData(uint16(val))
			if err != nil {
				return dataBytes, err
			}
		}
	} else if commandInfo.ValueType == sdkModel.Float64 {
		val, err := value.Float64Value()
		if err != nil {
			return dataBytes, err
		}
		if commandInfo.RawType == sdkModel.Int16 {
			dataBytes, err = getBinaryData(int16(val))
			if err != nil {
				return dataBytes, err
			}
		} else if commandInfo.RawType == sdkModel.Uint16 {
			dataBytes, err = getBinaryData(uint16(val))
			if err != nil {
				return dataBytes, err
			}
		}
	}

	driver.Logger.Debug(fmt.Sprintf("Transfer CommandValue to dataBytes for write command, %v, %v", commandInfo.ValueType, dataBytes))
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

func getBinaryData(val interface{}) (dataBytes []byte, err error) {
	buf := new(bytes.Buffer)
	err = binary.Write(buf, binary.BigEndian, val)
	if err != nil {
		return dataBytes, err
	}
	dataBytes = buf.Bytes()
	return dataBytes, err
}
