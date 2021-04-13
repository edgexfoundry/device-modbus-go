// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"fmt"

	"github.com/edgexfoundry/device-sdk-go/v2/pkg/models"
)

func checkAttributes(reqs []models.CommandRequest) error {
	var err error = nil
	for _, req := range reqs {
		attributes := req.Attributes

		_, err = toString(attributes[PRIMARY_TABLE])
		if err != nil {
			return fmt.Errorf("primaryTable fail to convert interface inoto string: %v", err)
		}

		_, err = toUint16(attributes[STARTING_ADDRESS])
		if err != nil {
			return fmt.Errorf("startingAddress fail to convert interface inoto unit16: %v", err)
		}

		_, ok := attributes[IS_BYTE_SWAP]
		if ok {
			_, err = toBool(attributes[IS_BYTE_SWAP])
			if err != nil {
				return fmt.Errorf("isByteSwap fail to convert interface inoto boolean: %v", err)
			}
		}

		_, ok = attributes[IS_WORD_SWAP]
		if ok {
			_, err = toBool(attributes[IS_WORD_SWAP])
			if err != nil {
				return fmt.Errorf("isWordSwap fail to convert interface inoto boolean: %v", err)
			}
		}
	}
	return err
}
