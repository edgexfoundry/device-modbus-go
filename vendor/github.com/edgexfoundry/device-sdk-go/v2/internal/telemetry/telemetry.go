/*******************************************************************************
 * Copyright 2019 Dell Inc., Intel Corporation
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *******************************************************************************/

package telemetry

import (
	"context"
	"runtime"
	"sync"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
)

// SystemUsage
// swagger:model
type SystemUsage struct {
	Memory     memoryUsage
	CpuBusyAvg float64
}

// swagger:model
type memoryUsage struct {
	Alloc       uint64
	TotalAlloc  uint64
	Sys         uint64
	Mallocs     uint64
	Frees       uint64
	LiveObjects uint64
}

type CpuUsage struct {
	Busy, // time used by all processes. this ideally does not include system processes.
	Idle, // time used by the idle process
	Total uint64 // reported sum total of all usage
}

var once sync.Once
var lastSample CpuUsage
var usageAvg float64

func NewSystemUsage() (s SystemUsage) {
	// The micro-service is to be considered the System Of Record (SOR) in terms of accurate information.
	// Fetch metrics for the metadata service.
	var rtm runtime.MemStats

	// Read full memory stats
	runtime.ReadMemStats(&rtm)

	// Miscellaneous memory stats
	s.Memory.Alloc = rtm.Alloc
	s.Memory.TotalAlloc = rtm.TotalAlloc
	s.Memory.Sys = rtm.Sys
	s.Memory.Mallocs = rtm.Mallocs
	s.Memory.Frees = rtm.Frees

	// Live objects = Mallocs - Frees
	s.Memory.LiveObjects = s.Memory.Mallocs - s.Memory.Frees

	s.CpuBusyAvg = usageAvg

	return s
}

func StartCpuUsageAverage(appWg *sync.WaitGroup, appCtx context.Context, logger logger.LoggingClient) {
	defer appWg.Done()

	logger.Info("Starting CPU Usage Average loop")

	for {
		select {
		case <-appCtx.Done():
			logger.Info("Exiting CPU Usage Average loop")
			return

		case <-time.After(time.Second * 10):
			nextUsage := PollCpu()
			usageAvg = AvgCpuUsage(lastSample, nextUsage)
			lastSample = nextUsage
		}
	}
}
