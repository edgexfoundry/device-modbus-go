/*
 * Copyright (C) 2022 Canonical Ltd
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
	hooks "github.com/canonical/edgex-snap-hooks/v3"
	"github.com/canonical/edgex-snap-hooks/v3/env"
	"github.com/canonical/edgex-snap-hooks/v3/log"
)

// installConfig copies all config files from $SNAP to $SNAP_DATA
func installConfig() error {
	path := "/config/" + app + "/res"

	err := hooks.CopyDir(
		env.Snap+path,
		env.SnapData+path)
	if err != nil {
		return err
	}

	return nil
}

func install() {
	log.SetComponentName("install")

	if err := installConfig(); err != nil {
		log.Fatalf("Error installing config files: %s", err)
	}
}
