// -*- Mode: Go; indent-tabs-mode: t -*-
//
// # Copyright (C) 2023 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
package provision

import (
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/utils"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"net/url"
	"path"
	"strings"
)

type FileType int

const (
	YAML FileType = iota
	JSON
	OTHER
)

func GetFileType(fullPath string) FileType {
	res := strings.Split(fullPath, "?")
	if strings.HasSuffix(res[0], yamlExt) || strings.HasSuffix(res[0], ymlExt) {
		return YAML
	} else if strings.HasSuffix(res[0], ".json") {
		return JSON
	} else {
		return OTHER
	}
}

func GetFullAndRedactedURI(baseURI *url.URL, file, description string, lc logger.LoggingClient) (string, string) {
	basePath, _ := path.Split(baseURI.Path)
	newPath, err := url.JoinPath(basePath, file)
	if err != nil {
		lc.Error("could not join URI path for %s %s/%s: %v", description, basePath, file, err)
		return "", ""
	}
	var fullURI url.URL
	err = utils.DeepCopy(baseURI, &fullURI)
	if err != nil {
		lc.Error("could not copy URI for %s %s: %v", description, newPath, err)
		return "", ""
	}
	fullURI.User = baseURI.User
	fullURI.Path = newPath
	lc.Debugf("%s URI for %s: %s", description, file, fullURI.String())
	return fullURI.String(), fullURI.Redacted()
}
