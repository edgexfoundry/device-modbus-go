/*
 * Copyright (C) 2021 Canonical Ltd
 *
 *  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 *  in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *
 * SPDX-License-Identifier: Apache-2.0'
 */

package main

import (
	hooks "github.com/canonical/edgex-snap-hooks/v2"
	"github.com/canonical/edgex-snap-hooks/v2/log"
	"github.com/canonical/edgex-snap-hooks/v2/options"
)

// ConfToEnv defines mappings from snap config keys to EdgeX environment variable
// names that are used to override individual device-modbus's [Device]  configuration
// values via a .env file read by the snap service wrapper.
//
// The syntax to set a configuration key is:
//
// env.<section>.<keyname>
var ConfToEnv = map[string]string{
	// [Device]
	"device.update-last-connected": "DEVICE_UPDATELASTCONNECTED",
	"device.use-message-bus":       "DEVICE_USEMESSAGEBUS",
}

// configure is called by the main function
func configure() {

	const service = "device-modbus"

	log.SetComponentName("configure")

	log.Info("Processing legacy env options")
	envJSON, err := hooks.NewSnapCtl().Config(hooks.EnvConfig)
	if err != nil {
		log.Fatalf("Reading config 'env' failed: %v", err)
	}
	if envJSON != "" {
		log.Debugf("envJSON: %s", envJSON)
		err = hooks.HandleEdgeXConfig(service, envJSON, ConfToEnv)
		if err != nil {
			log.Fatalf("HandleEdgeXConfig failed: %v", err)
		}
	}

	log.Info("Processing config options")
	err = options.ProcessConfig(service)
	if err != nil {
		log.Fatalf("could not process config options: %v", err)
	}

	log.Info("Processing autostart options")
	err = options.ProcessAutostart(service)
	if err != nil {
		log.Fatalf("could not process autostart options: %v", err)
	}
}
