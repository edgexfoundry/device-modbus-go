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
	"github.com/canonical/edgex-snap-hooks/v3/log"
	"github.com/canonical/edgex-snap-hooks/v3/options"
)

func configure() {
	log.SetComponentName("configure")

	if err := options.ProcessConfig(app); err != nil {
		log.Fatalf("Error processing config options: %v", err)
	}

	if err := options.ProcessAutostart(app); err != nil {
		log.Fatalf("Error processing autostart options: %v", err)
	}
}
