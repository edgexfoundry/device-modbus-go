// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2019-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package autoevent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/OneOfOne/xxhash"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
	"github.com/google/uuid"

	"github.com/edgexfoundry/device-sdk-go/v2/internal/application"
	sdkCommon "github.com/edgexfoundry/device-sdk-go/v2/internal/common"
)

type Executor struct {
	deviceName   string
	sourceName   string
	onChange     bool
	lastReadings map[string]interface{}
	duration     time.Duration
	stop         bool
	mutex        *sync.Mutex
}

// Run triggers this Executor executes the handler for the event source periodically
func (e *Executor) Run(ctx context.Context, wg *sync.WaitGroup, buffer chan bool, dic *di.Container) {
	wg.Add(1)
	defer wg.Done()

	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(e.duration):
			if e.stop {
				return
			}
			lc.Debugf("AutoEvent - reading %s", e.sourceName)
			evt, err := readResource(e, dic)
			if err != nil {
				lc.Errorf("AutoEvent - error occurs when reading resource %s: %v", e.sourceName, err)
				continue
			}

			if evt != nil {
				if e.onChange {
					if e.compareReadings(evt.Readings) {
						lc.Debugf("AutoEvent - readings are the same as previous one")
						continue
					}
				}
				// After the auto event executes a read command, it will create a goroutine to send out events.
				// When the concurrent auto event amount becomes large, core-data might be hard to handle so many HTTP requests at the same time.
				// The device service will get some network errors like EOF or Connection reset by peer.
				// By adding a buffer here, the user can use the Service.AsyncBufferSize configuration to control the goroutine for sending events.
				go func() {
					buffer <- true
					correlationId := uuid.NewString()
					sdkCommon.SendEvent(evt, correlationId, dic)
					lc.Tracef("AutoEvent - Sent new Event/Reading for '%s' source with Correlation Id '%s'", evt.SourceName, correlationId)
					<-buffer
				}()
			} else {
				lc.Debugf("AutoEvent - no event generated when reading resource %s", e.sourceName)
			}
		}
	}
}

func readResource(e *Executor, dic *di.Container) (event *dtos.Event, err errors.EdgeX) {
	vars := make(map[string]string, 2)
	vars[common.Name] = e.deviceName
	vars[common.Command] = e.sourceName

	res, err := application.CommandHandler(true, false, "", vars, nil, "", dic)
	if err != nil {
		return event, err
	}
	return res, nil
}

func (e *Executor) compareReadings(readings []dtos.BaseReading) bool {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if len(e.lastReadings) != len(readings) {
		e.renewLastReadings(readings)
		return false
	}

	var result = true
	for _, reading := range readings {
		if lastReading, ok := e.lastReadings[reading.ResourceName]; ok {
			if reading.ValueType == common.ValueTypeBinary {
				checksum := xxhash.Checksum64(reading.BinaryValue)
				if lastReading != checksum {
					e.lastReadings[reading.ResourceName] = checksum
					result = false
				}
			} else {
				if lastReading != reading.Value {
					e.lastReadings[reading.ResourceName] = reading.Value
					result = false
				}
			}
		} else {
			e.renewLastReadings(readings)
			return false
		}
	}

	return result
}

func (e *Executor) renewLastReadings(readings []dtos.BaseReading) {
	e.lastReadings = make(map[string]interface{}, len(readings))
	for _, r := range readings {
		if r.ValueType == common.ValueTypeBinary {
			e.lastReadings[r.ResourceName] = xxhash.Checksum64(r.BinaryValue)
		} else {
			e.lastReadings[r.ResourceName] = r.Value
		}
	}
}

// Stop marks this Executor stopped
func (e *Executor) Stop() {
	e.stop = true
}

// NewExecutor creates an Executor for an AutoEvent
func NewExecutor(deviceName string, ae models.AutoEvent) (*Executor, errors.EdgeX) {
	// check Frequency
	duration, err := time.ParseDuration(ae.Interval)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to parse AutoEvent %s duration", ae.SourceName), err)
	}

	return &Executor{
		deviceName: deviceName,
		sourceName: ae.SourceName,
		onChange:   ae.OnChange,
		duration:   duration,
		stop:       false,
		mutex:      &sync.Mutex{}}, nil
}
