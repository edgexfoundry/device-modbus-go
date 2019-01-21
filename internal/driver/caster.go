package driver

import (
	"fmt"
	"strconv"
)

func toString(i interface{}) (string, error) {
	if i == nil {
		return "", fmt.Errorf("value is undefined")
	}
	return fmt.Sprintf("%v", i), nil
}

func toUint16(i interface{}) (uint16, error) {
	var val uint16 = 0
	if i == nil {
		return val, fmt.Errorf("value is undefined")
	}
	stringVal := fmt.Sprintf("%v", i)
	uint64Val, err := strconv.ParseUint(stringVal, 10, 16)
	if err != nil {
		return val, fmt.Errorf("parse to uint16 failed, error: %v", err)
	}
	val = uint16(uint64Val)
	return val, nil
}

func toBool(i interface{}) (bool, error) {
	var val = false
	if i == nil {
		return val, fmt.Errorf("value is undefined")
	}
	stringVal := fmt.Sprintf("%v", i)
	val, err := strconv.ParseBool(stringVal)
	if err != nil {
		return val, fmt.Errorf("parse to bool failed, error: %v", err)
	}
	return val, nil
}
