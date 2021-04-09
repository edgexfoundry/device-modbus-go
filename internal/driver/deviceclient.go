// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018-2021 IOTech Ltd
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

	"github.com/edgexfoundry/device-sdk-go/v2/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2"
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
	ValueType       string
	// how many register need to read
	Length     uint16
	IsByteSwap bool
	IsWordSwap bool
	RawType    string
}

func createCommandInfo(req *models.CommandRequest) (*CommandInfo, error) {
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

func getRawType(req *models.CommandRequest) (valueType string, err error) {
	rawType, ok := req.Attributes[RAW_TYPE]
	if !ok || rawType == "" {
		return req.Type, err
	}

	switch rawType {
	case UINT16:
		valueType = v2.ValueTypeUint16
	case INT16:
		valueType = v2.ValueTypeInt16
	default:
		driver.Logger.Warnf("The raw type %v is not supported, use value type %v instead", rawType, req.Type)
		err = fmt.Errorf("the raw type %v is not supported", rawType)
	}
	return valueType, err
}

func calculateAddressLength(primaryTable string, valueType string) uint16 {
	var primaryTableBit = PrimaryTableBitCountMap[primaryTable]
	var valueTypeBitCount = ValueTypeBitCountMap[valueType]

	var length = valueTypeBitCount / primaryTableBit
	if length < 1 {
		length = 1
	}

	return length
}

// TransformDataBytesToResult is used to transform the device's binary data to the specified value type as the actual result.
func TransformDataBytesToResult(req *models.CommandRequest, dataBytes []byte, commandInfo *CommandInfo) (*models.CommandValue, error) {
	var err error
	var res interface{}
	var result = &models.CommandValue{}

	switch commandInfo.ValueType {
	case v2.ValueTypeUint16:
		res = binary.BigEndian.Uint16(dataBytes)
	case v2.ValueTypeUint32:
		res = binary.BigEndian.Uint32(swap32BitDataBytes(dataBytes, commandInfo.IsByteSwap, commandInfo.IsWordSwap))
	case v2.ValueTypeUint64:
		res = binary.BigEndian.Uint64(dataBytes)
	case v2.ValueTypeInt16:
		res = int16(binary.BigEndian.Uint16(dataBytes))
	case v2.ValueTypeInt32:
		res = int32(binary.BigEndian.Uint32(swap32BitDataBytes(dataBytes, commandInfo.IsByteSwap, commandInfo.IsWordSwap)))
	case v2.ValueTypeInt64:
		res = int64(binary.BigEndian.Uint64(dataBytes))
	case v2.ValueTypeFloat32:
		switch commandInfo.RawType {
		case v2.ValueTypeFloat32:
			raw := binary.BigEndian.Uint32(swap32BitDataBytes(dataBytes, commandInfo.IsByteSwap, commandInfo.IsWordSwap))
			res = math.Float32frombits(raw)
		case v2.ValueTypeInt16:
			raw := int16(binary.BigEndian.Uint16(dataBytes))
			res = float32(raw)
			driver.Logger.Debugf("According to the rawType %s and the value type %s, convert integer %d to float %v ", INT16, FLOAT32, res, result.ValueToString())
		case v2.ValueTypeUint16:
			raw := binary.BigEndian.Uint16(dataBytes)
			res = float32(raw)
			driver.Logger.Debugf("According to the rawType %s and the value type %s, convert integer %d to float %v ", UINT16, FLOAT32, res, result.ValueToString())
		}
	case v2.ValueTypeFloat64:
		switch commandInfo.RawType {
		case v2.ValueTypeFloat64:
			raw := binary.BigEndian.Uint64(dataBytes)
			res = math.Float64frombits(raw)
		case v2.ValueTypeInt16:
			raw := int16(binary.BigEndian.Uint16(dataBytes))
			res = float64(raw)
			driver.Logger.Debugf("According to the rawType %s and the value type %s, convert integer %d to float %v ", INT16, FLOAT64, res, result.ValueToString())
		case v2.ValueTypeUint16:
			raw := binary.BigEndian.Uint16(dataBytes)
			res = float64(raw)
			driver.Logger.Debugf("According to the rawType %s and the value type %s, convert integer %d to float %v ", UINT16, FLOAT64, res, result.ValueToString())
		}
	case v2.ValueTypeBool:
		res = false
		// to find the 1st bit of the dataBytes by mask it with 2^0 = 1 (00000001)
		if (dataBytes[0] & 1) > 0 {
			res = true
		}
	default:
		return nil, fmt.Errorf("return result fail, none supported value type: %v", commandInfo.ValueType)
	}

	result, err = models.NewCommandValue(req.DeviceResourceName, commandInfo.ValueType, res)
	if err != nil {
		return nil, err
	}
	result.Origin = time.Now().UnixNano()

	driver.Logger.Debugf("Transfer dataBytes to CommandValue(%v) successful.", result.ValueToString())
	return result, nil
}

// TransformCommandValueToDataBytes transforms the reading value to binary data which is used to transfer data via Modbus protocol.
func TransformCommandValueToDataBytes(commandInfo *CommandInfo, value *models.CommandValue) ([]byte, error) {
	var err error
	var byteCount = calculateByteCount(commandInfo)
	buf := new(bytes.Buffer)
	err = binary.Write(buf, binary.BigEndian, value.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to transform %v to []byte", value.Value)
	}

	numericValue := buf.Bytes()
	var maxSize = uint16(len(numericValue))
	var dataBytes = numericValue[maxSize-byteCount : maxSize]

	_, ok := ValueTypeBitCountMap[commandInfo.ValueType]
	if !ok {
		err = fmt.Errorf("none supported value type : %v \n", commandInfo.ValueType)
		return dataBytes, err
	}

	if commandInfo.ValueType == v2.ValueTypeUint32 || commandInfo.ValueType == v2.ValueTypeInt32 || commandInfo.ValueType == v2.ValueTypeFloat32 {
		dataBytes = swap32BitDataBytes(dataBytes, commandInfo.IsByteSwap, commandInfo.IsWordSwap)
	}

	// Cast value according to the rawType, this feature only converts float value to integer 32bit value
	if commandInfo.ValueType == v2.ValueTypeFloat32 {
		val, edgexErr := value.Float32Value()
		if edgexErr != nil {
			return dataBytes, edgexErr
		}
		if commandInfo.RawType == v2.ValueTypeInt16 {
			dataBytes, err = getBinaryData(int16(val))
			if err != nil {
				return dataBytes, err
			}
		} else if commandInfo.RawType == v2.ValueTypeUint16 {
			dataBytes, err = getBinaryData(uint16(val))
			if err != nil {
				return dataBytes, err
			}
		}
	} else if commandInfo.ValueType == v2.ValueTypeFloat64 {
		val, edgexErr := value.Float64Value()
		if edgexErr != nil {
			return dataBytes, edgexErr
		}
		if commandInfo.RawType == v2.ValueTypeInt16 {
			dataBytes, err = getBinaryData(int16(val))
			if err != nil {
				return dataBytes, err
			}
		} else if commandInfo.RawType == v2.ValueTypeUint16 {
			dataBytes, err = getBinaryData(uint16(val))
			if err != nil {
				return dataBytes, err
			}
		}
	}

	driver.Logger.Debugf("Transfer CommandValue to dataBytes for write command, %v, %v", commandInfo.ValueType, dataBytes)
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
