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
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
)

// DeviceClient is a interface for modbus client lib to implementation
// It's responsibility are handle connection, read data bytes value and write data bytes value
type DeviceClient interface {
	OpenConnection() error
	GetValue(commandInfo interface{}) ([]byte, error)
	SetValue(commandInfo interface{}, value []byte) error
	CloseConnection() error
}

// CommandInfo is command info
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
	if _, ok := req.Attributes[PRIMARY_TABLE]; !ok {
		return nil, errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("attribute %s not exists", PRIMARY_TABLE), nil)
	}
	primaryTable := fmt.Sprintf("%v", req.Attributes[PRIMARY_TABLE])
	primaryTable = strings.ToUpper(primaryTable)

	if _, ok := req.Attributes[STARTING_ADDRESS]; !ok {
		return nil, errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("attribute %s not exists", STARTING_ADDRESS), nil)
	}
	startingAddress, err := castStartingAddress(req.Attributes[STARTING_ADDRESS])
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.Kind(err), fmt.Sprintf("fail to cast %s", STARTING_ADDRESS), err)
	}

	var rawType = req.Type
	if _, ok := req.Attributes[RAW_TYPE]; ok {
		rawType = fmt.Sprintf("%v", req.Attributes[RAW_TYPE])
		rawType, err = normalizeRawType(rawType)
		if err != nil {
			return nil, err
		}
	}
	var length uint16
	if req.Type == common.ValueTypeString {
		length, err = castStartingAddress(req.Attributes[STRING_REGISTER_SIZE])
		if err != nil {
			return nil, err
		} else if (length > 123) || (length < 1) {
			return nil, errors.NewCommonEdgeX(errors.KindLimitExceeded, fmt.Sprintf("register size should be within the range of 1~123, get %v.", length), nil)
		}
	} else {
		length = calculateAddressLength(primaryTable, rawType)
	}

	var isByteSwap = false
	if _, ok := req.Attributes[IS_BYTE_SWAP]; ok {
		isByteSwap, err = castSwapAttribute(req.Attributes[IS_BYTE_SWAP])
		if err != nil {
			return nil, errors.NewCommonEdgeX(errors.Kind(err), fmt.Sprintf("fail to cast %s", IS_BYTE_SWAP), err)
		}
	}

	var isWordSwap = false
	if _, ok := req.Attributes[IS_WORD_SWAP]; ok {
		isWordSwap, err = castSwapAttribute(req.Attributes[IS_WORD_SWAP])
		if err != nil {
			return nil, errors.NewCommonEdgeX(errors.Kind(err), fmt.Sprintf("fail to cast %s", IS_WORD_SWAP), err)
		}
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
	case common.ValueTypeUint16:
		res = binary.BigEndian.Uint16(dataBytes)
	case common.ValueTypeUint32:
		res = binary.BigEndian.Uint32(swap32BitDataBytes(dataBytes, commandInfo.IsByteSwap, commandInfo.IsWordSwap))
	case common.ValueTypeUint64:
		res = binary.BigEndian.Uint64(dataBytes)
	case common.ValueTypeInt16:
		res = int16(binary.BigEndian.Uint16(dataBytes))
	case common.ValueTypeInt32:
		res = int32(binary.BigEndian.Uint32(swap32BitDataBytes(dataBytes, commandInfo.IsByteSwap, commandInfo.IsWordSwap)))
	case common.ValueTypeInt64:
		res = int64(binary.BigEndian.Uint64(dataBytes))
	case common.ValueTypeFloat32:
		switch commandInfo.RawType {
		case common.ValueTypeFloat32:
			raw := binary.BigEndian.Uint32(swap32BitDataBytes(dataBytes, commandInfo.IsByteSwap, commandInfo.IsWordSwap))
			res = math.Float32frombits(raw)
		case common.ValueTypeInt16:
			raw := int16(binary.BigEndian.Uint16(dataBytes))
			res = float32(raw)
			driver.Logger.Debugf("According to the rawType %s and the value type %s, convert integer %d to float %v ", INT16, FLOAT32, res, result.ValueToString())
		case common.ValueTypeUint16:
			raw := binary.BigEndian.Uint16(dataBytes)
			res = float32(raw)
			driver.Logger.Debugf("According to the rawType %s and the value type %s, convert integer %d to float %v ", UINT16, FLOAT32, res, result.ValueToString())
		}
	case common.ValueTypeFloat64:
		switch commandInfo.RawType {
		case common.ValueTypeFloat64:
			raw := binary.BigEndian.Uint64(dataBytes)
			res = math.Float64frombits(raw)
		case common.ValueTypeInt16:
			raw := int16(binary.BigEndian.Uint16(dataBytes))
			res = float64(raw)
			driver.Logger.Debugf("According to the rawType %s and the value type %s, convert integer %d to float %v ", INT16, FLOAT64, res, result.ValueToString())
		case common.ValueTypeUint16:
			raw := binary.BigEndian.Uint16(dataBytes)
			res = float64(raw)
			driver.Logger.Debugf("According to the rawType %s and the value type %s, convert integer %d to float %v ", UINT16, FLOAT64, res, result.ValueToString())
		}
	case common.ValueTypeBool:
		res = false
		// to find the 1st bit of the dataBytes by mask it with 2^0 = 1 (00000001)
		if (dataBytes[0] & 1) > 0 {
			res = true
		}
	case common.ValueTypeString:
		res = string(bytes.Trim(dataBytes, string(rune(0))))
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
	var dataBytes []byte
	buf := new(bytes.Buffer)
	if commandInfo.ValueType != common.ValueTypeString {
		err = binary.Write(buf, binary.BigEndian, value.Value)
		if err != nil {
			return nil, fmt.Errorf("failed to transform %v to []byte", value.Value)
		}

		numericValue := buf.Bytes()
		var maxSize = uint16(len(numericValue))
		dataBytes = numericValue[maxSize-byteCount : maxSize]
	}

	_, ok := ValueTypeBitCountMap[commandInfo.ValueType]
	if !ok {
		err = fmt.Errorf("none supported value type : %v \n", commandInfo.ValueType)
		return dataBytes, err
	}

	if commandInfo.ValueType == common.ValueTypeUint32 || commandInfo.ValueType == common.ValueTypeInt32 || commandInfo.ValueType == common.ValueTypeFloat32 {
		dataBytes = swap32BitDataBytes(dataBytes, commandInfo.IsByteSwap, commandInfo.IsWordSwap)
	}

	// Cast value according to the rawType, this feature converts float value to integer 32bit value
	if commandInfo.ValueType == common.ValueTypeFloat32 {
		val, edgexErr := value.Float32Value()
		if edgexErr != nil {
			return dataBytes, edgexErr
		}
		if commandInfo.RawType == common.ValueTypeInt16 {
			dataBytes, err = getBinaryData(int16(val))
			if err != nil {
				return dataBytes, err
			}
		} else if commandInfo.RawType == common.ValueTypeUint16 {
			dataBytes, err = getBinaryData(uint16(val))
			if err != nil {
				return dataBytes, err
			}
		}
	} else if commandInfo.ValueType == common.ValueTypeFloat64 {
		val, edgexErr := value.Float64Value()
		if edgexErr != nil {
			return dataBytes, edgexErr
		}
		if commandInfo.RawType == common.ValueTypeInt16 {
			dataBytes, err = getBinaryData(int16(val))
			if err != nil {
				return dataBytes, err
			}
		} else if commandInfo.RawType == common.ValueTypeUint16 {
			dataBytes, err = getBinaryData(uint16(val))
			if err != nil {
				return dataBytes, err
			}
		}
	} else if commandInfo.ValueType == common.ValueTypeString {
		// Cast value of string type
		oriStr := value.ValueToString()
		tempBytes := []byte(oriStr)
		bytesL := len(tempBytes)
		oriByteL := int(commandInfo.Length * 2)
		if bytesL < oriByteL {
			less := make([]byte, oriByteL-bytesL)
			dataBytes = append(tempBytes, less...)
		} else if bytesL > oriByteL {
			dataBytes = tempBytes[:oriByteL]
		} else {
			dataBytes = []byte(oriStr)
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
