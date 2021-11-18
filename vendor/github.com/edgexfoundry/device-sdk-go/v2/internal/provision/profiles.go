// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package provision

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/edgexfoundry/device-sdk-go/v2/internal/cache"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

const (
	jsonExt = ".json"
	yamlExt = ".yaml"
	ymlExt  = ".yml"
)

func LoadProfiles(path string, dic *di.Container) errors.EdgeX {
	if path == "" {
		return nil
	}
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

	absPath, err := filepath.Abs(path)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, "failed to create absolute path", err)
	}

	fileInfo, err := os.ReadDir(absPath)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, "failed to read directory", err)
	}

	var addProfilesReq []requests.DeviceProfileRequest
	dpc := bootstrapContainer.MetadataDeviceProfileClientFrom(dic.Get)
	lc.Infof("Loading pre-defined profiles from %s", absPath)
	for _, file := range fileInfo {
		var profile dtos.DeviceProfile
		fullPath := filepath.Join(absPath, file.Name())
		if strings.HasSuffix(fullPath, yamlExt) || strings.HasSuffix(fullPath, ymlExt) {
			content, err := os.ReadFile(fullPath)
			if err != nil {
				lc.Errorf("Failed to read %s: %v", fullPath, err)
				continue
			}

			err = yaml.Unmarshal(content, &profile)
			if err != nil {
				lc.Errorf("Failed to decode profile %s: %v", file.Name(), err)
				continue
			}
		} else if strings.HasSuffix(fullPath, jsonExt) {
			content, err := os.ReadFile(fullPath)
			if err != nil {
				lc.Errorf("Failed to read %s: %v", fullPath, err)
				continue
			}

			err = json.Unmarshal(content, &profile)
			if err != nil {
				lc.Errorf("Failed to decode profile %s: %v", file.Name(), err)
				continue
			}
		} else {
			continue
		}

		res, err := dpc.DeviceProfileByName(context.Background(), profile.Name)
		if err == nil {
			lc.Infof("Profile %s exists, using the existing one", profile.Name)
			_, exist := cache.Profiles().ForName(profile.Name)
			if !exist {
				err = cache.Profiles().Add(dtos.ToDeviceProfileModel(res.Profile))
				if err != nil {
					return errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to cache the profile %s", res.Profile.Name), err)
				}
			}
		} else {
			lc.Infof("Profile %s not found in Metadata, adding it ...", profile.Name)
			req := requests.NewDeviceProfileRequest(profile)
			addProfilesReq = append(addProfilesReq, req)
		}
	}

	if len(addProfilesReq) == 0 {
		return nil
	}
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.NewString())
	_, edgexErr := dpc.Add(ctx, addProfilesReq)
	return edgexErr
}
