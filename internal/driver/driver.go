// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018-2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

// Package driver is used to execute device-sdk's commands
package driver

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/edgexfoundry/device-sdk-go/v4/pkg/interfaces"
	sdkModel "github.com/edgexfoundry/device-sdk-go/v4/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
)

var once sync.Once
var driver *Driver

type Driver struct {
	Logger              logger.LoggingClient
	AsyncCh             chan<- *sdkModel.AsyncValues
	addressMutex        sync.Mutex
	addressMap          map[string]chan bool
	workingAddressCount map[string]int
	stopped             bool
	clientMutex         sync.Mutex
	clientMap           map[string]DeviceClient
}

var concurrentCommandLimit = 100

func (d *Driver) createDeviceClient(info *ConnectionInfo) (DeviceClient, error) {
	d.clientMutex.Lock()
	defer d.clientMutex.Unlock()
	key := info.String()
	c, ok := d.clientMap[key]
	if ok {
		return c, nil
	}
	c, err := NewDeviceClient(info)
	if err != nil {
		driver.Logger.Errorf("create device client failed. err:%v \n", err)
		return nil, err
	}
	d.clientMap[key] = c
	return c, nil
}

func (d *Driver) removeDeviceClient(info *ConnectionInfo) {
	d.clientMutex.Lock()
	defer d.clientMutex.Unlock()
	key := info.String()
	if _, ok := d.clientMap[key]; ok {
		client := d.clientMap[key]
		err := client.CloseConnection()
		if err != nil {
			driver.Logger.Errorf("Device client close error, client key = %s, err = %v", key, err)
		} else {
			driver.Logger.Infof("device client closed,client key = %s", key)
		}
		delete(d.clientMap, key)
	}
}

// lockAddress mark address is unavailable because real device handle one request at a time
func (d *Driver) lockAddress(address string) error {
	if d.stopped {
		return fmt.Errorf("service attempts to stop and unable to handle new request")
	}
	d.addressMutex.Lock()
	lock, ok := d.addressMap[address]
	if !ok {
		lock = make(chan bool, 1)
		d.addressMap[address] = lock
	}

	// workingAddressCount used to check high-frequency command execution to avoid goroutine block
	count, ok := d.workingAddressCount[address]
	if !ok {
		d.workingAddressCount[address] = 1
	} else if count >= concurrentCommandLimit {
		d.addressMutex.Unlock()
		errorMessage := fmt.Sprintf("High-frequency command execution. There are %v commands with the same address in the queue", concurrentCommandLimit)
		d.Logger.Error(errorMessage)
		return fmt.Errorf("%s", errorMessage)
	} else {
		d.workingAddressCount[address] = count + 1
	}

	d.addressMutex.Unlock()
	lock <- true

	return nil
}

// unlockAddress remove token after command finish
func (d *Driver) unlockAddress(address string) {
	d.addressMutex.Lock()
	lock := d.addressMap[address]
	d.workingAddressCount[address] = d.workingAddressCount[address] - 1
	d.addressMutex.Unlock()
	<-lock
}

// lockableAddress return the lockable address according to the protocol
func (d *Driver) lockableAddress(info *ConnectionInfo) string {
	var address string
	if info.Protocol == ProtocolTCP {
		address = fmt.Sprintf("%s:%d", info.Address, info.Port)
	} else {
		address = info.Address
	}
	return address
}

func (d *Driver) HandleReadCommands(deviceName string, protocols map[string]models.ProtocolProperties, reqs []sdkModel.CommandRequest) (responses []*sdkModel.CommandValue, err error) {
	d.Logger.Debugf("HandleReadCommands called for device %s", deviceName)
	d.Logger.Debugf("Read command requests: %+v", reqs)
	connectionInfo, err := createConnectionInfo(protocols)
	if err != nil {
		driver.Logger.Errorf("Fail to create read command connection info. err:%v \n", err)
		return responses, err
	}

	err = d.lockAddress(d.lockableAddress(connectionInfo))
	if err != nil {
		return responses, err
	}
	defer d.unlockAddress(d.lockableAddress(connectionInfo))

	responses = make([]*sdkModel.CommandValue, len(reqs))

	// create device client and open connection
	deviceClient, err := d.createDeviceClient(connectionInfo)
	if err != nil {
		driver.Logger.Errorf("Read command OpenConnection failed. err:%v \n", err)
		return responses, err
	}
	driver.Logger.Debugf("key = %s,client = %+v", connectionInfo.String(), deviceClient)

	reqsMeta, reqsAct, err := aggregateReadRequests(reqs)
	if err != nil {
		return responses, err
	}
	d.Logger.Infof("Actual Read requests: %+v", reqsAct)
	// For each group, do a single read and split
	for _, g := range reqsAct {
		baseReq := reqsMeta[g.startIdx].req
		baseCmdInfo, err := createCommandInfo(&baseReq)
		if err != nil {
			return responses, err
		}
		baseCmdInfo.Length = g.totalLen
		deviceClient.SetSlaveID(byte(connectionInfo.UnitID))
		data, err := deviceClient.GetValue(baseCmdInfo)
		if err != nil {
			driver.Logger.Errorf("Read command failed, remove the Modbus client from the cache to allow re-establish connection next time. Cmd: %v err:%v", baseReq.DeviceResourceName, err)
			d.removeDeviceClient(connectionInfo)
			return responses, err
		}

		// Split the grouped response data into individual responses
		splitResults, err := splitGroupReadResponse(data, g, reqsMeta)
		if err != nil {
			driver.Logger.Errorf("Failed to split group read response: %v", err)
			return responses, err
		}

		for _, split := range splitResults {
			res, err := handleReadRequestAndTransformData(split.Request, split.Data)
			if err != nil {
				driver.Logger.Errorf("Read command failed, remove the Modbus client from the cache to allow re-establish connection next time. Cmd: %v err:%v", split.Request.DeviceResourceName, err)
				d.removeDeviceClient(connectionInfo)
				return responses, err
			}
			responses[split.OriginalIdx] = res
		}
	}
	driver.Logger.Debugf("get response %v", responses)
	return responses, nil
}

func handleReadRequestAndTransformData(req sdkModel.CommandRequest, data []byte) (*sdkModel.CommandValue, error) {
	var result *sdkModel.CommandValue
	var err error

	commandInfo, err := createCommandInfo(&req)
	if err != nil {
		return nil, err
	}

	result, err = TransformDataBytesToResult(&req, data, commandInfo)

	if err != nil {
		return result, err
	} else {
		driver.Logger.Tracef("Read command finished. Cmd:%v, %v \n", req.DeviceResourceName, result)
	}

	return result, nil
}

func (d *Driver) HandleWriteCommands(deviceName string, protocols map[string]models.ProtocolProperties, reqs []sdkModel.CommandRequest, params []*sdkModel.CommandValue) error {
	connectionInfo, err := createConnectionInfo(protocols)
	if err != nil {
		driver.Logger.Errorf("Fail to create write command connection info. err:%v \n", err)
		return err
	}

	err = d.lockAddress(d.lockableAddress(connectionInfo))
	if err != nil {
		return err
	}
	defer d.unlockAddress(d.lockableAddress(connectionInfo))

	// create device client and open connection
	deviceClient, err := d.createDeviceClient(connectionInfo)
	if err != nil {
		return err
	}

	// handle command requests
	for i, req := range reqs {
		err = handleWriteCommandRequest(deviceClient, connectionInfo, req, params[i])
		if err != nil {
			driver.Logger.Errorf("Write command failed, remove the Modbus client from the cache to allow re-establish connection next time. Cmd: %v err:%v", req.DeviceResourceName, err)
			d.removeDeviceClient(connectionInfo)
			break
		}
	}

	return err
}

func handleWriteCommandRequest(deviceClient DeviceClient, connectionInfo *ConnectionInfo, req sdkModel.CommandRequest, param *sdkModel.CommandValue) error {
	var err error

	commandInfo, err := createCommandInfo(&req)
	if err != nil {
		return err
	}

	deviceClient.SetSlaveID(byte(connectionInfo.UnitID))

	dataBytes, err := TransformCommandValueToDataBytes(commandInfo, param)
	if err != nil {
		return fmt.Errorf("transform command value failed, err: %v", err)
	}

	err = deviceClient.SetValue(commandInfo, dataBytes)
	if err != nil {
		return fmt.Errorf("handle write command request failed, err: %v", err)
	}

	driver.Logger.Tracef("Write command finished. Cmd:%v \n", req.DeviceResourceName)
	return nil
}

func (d *Driver) Initialize(sdk interfaces.DeviceServiceSDK) error {
	d.Logger = sdk.LoggingClient()
	d.AsyncCh = sdk.AsyncValuesChannel()
	d.addressMap = make(map[string]chan bool)
	d.workingAddressCount = make(map[string]int)
	d.clientMap = make(map[string]DeviceClient)
	return nil
}

func (d *Driver) Start() error {
	return nil
}

func (d *Driver) Stop(force bool) error {
	d.clientMutex.Lock()
	for key, client := range d.clientMap {
		err := client.CloseConnection()
		if err != nil {
			d.Logger.Errorf("device client closed,client key = %s,err = %v", key, err)
		}
	}
	d.clientMutex.Unlock()
	d.stopped = true
	if !force {
		d.waitAllCommandsToFinish()
	}
	for _, locked := range d.addressMap {
		close(locked)
	}
	return nil
}

// waitAllCommandsToFinish used to check and wait for the unfinished job
func (d *Driver) waitAllCommandsToFinish() {
loop:
	for {
		for _, count := range d.workingAddressCount {
			if count != 0 {
				// wait a moment and check again
				time.Sleep(time.Second * SERVICE_STOP_WAIT_TIME)
				continue loop
			}
		}
		break loop
	}
}

func (d *Driver) AddDevice(deviceName string, protocols map[string]models.ProtocolProperties, adminState models.AdminState) error {
	d.Logger.Debugf("Device %s is added", deviceName)
	return nil
}

func (d *Driver) UpdateDevice(deviceName string, protocols map[string]models.ProtocolProperties, adminState models.AdminState) error {
	d.Logger.Debugf("Device %s is updated", deviceName)
	return nil
}

func (d *Driver) RemoveDevice(deviceName string, protocols map[string]models.ProtocolProperties) error {
	connectionInfo, err := createConnectionInfo(protocols)
	if err != nil {
		driver.Logger.Errorf("Fail to create remove device connection info. err:%v \n", err)
		return err
	}
	d.removeDeviceClient(connectionInfo)
	d.Logger.Debugf("Device %s is removed", deviceName)
	return nil
}

func (d *Driver) Discover() error {
	return fmt.Errorf("driver's Discover function isn't implemented")
}

func (d *Driver) ValidateDevice(device models.Device) error {
	_, err := createConnectionInfo(device.Protocols)
	if err != nil {
		return fmt.Errorf("invalid protocol properties, %v", err)
	}
	return nil
}

func NewProtocolDriver() interfaces.ProtocolDriver {
	once.Do(func() {
		driver = new(Driver)
	})
	return driver
}

// ReqWithMeta holds metadata for grouping read requests
type ReqWithMeta struct {
	idx         int
	req         sdkModel.CommandRequest
	primaryType string
	dataType    string
	startAddr   uint16
	length      uint16
}

// DeviceAddressRange represents a DeviceAddressRange of contiguous, compatible read requests
type DeviceAddressRange struct {
	startIdx    int
	endIdx      int
	primaryType string
	rawType     string
	startAddr   uint16
	totalLen    uint16
}

// aggregateReadRequests groups read requests by register type, value type, and contiguous addresses
func aggregateReadRequests(reqs []sdkModel.CommandRequest) ([]ReqWithMeta, []DeviceAddressRange, error) {
	reqsMeta := make([]ReqWithMeta, len(reqs))
	for i, r := range reqs {
		cmdInfo, err := createCommandInfo(&r)
		if err != nil {
			return nil, nil, err
		}
		reqsMeta[i] = ReqWithMeta{
			idx:         i,
			req:         r,
			primaryType: cmdInfo.PrimaryTable,
			dataType:    cmdInfo.ValueType,
			startAddr:   cmdInfo.StartingAddress,
			length:      cmdInfo.Length,
		}
	}
	// Sort by startAddr, primaryType, valueType
	sort.Slice(reqsMeta, func(i, j int) bool {
		if reqsMeta[i].startAddr != reqsMeta[j].startAddr {
			return reqsMeta[i].startAddr < reqsMeta[j].startAddr
		}
		if reqsMeta[i].primaryType != reqsMeta[j].primaryType {
			return reqsMeta[i].primaryType < reqsMeta[j].primaryType
		}
		return reqsMeta[i].dataType < reqsMeta[j].dataType
	})

	// Grouping
	groups := []DeviceAddressRange{}
	i := 0
	for i < len(reqsMeta) {
		g := DeviceAddressRange{
			startIdx:    i,
			endIdx:      i,
			primaryType: reqsMeta[i].primaryType,
			rawType:     reqsMeta[i].dataType,
			startAddr:   reqsMeta[i].startAddr,
			totalLen:    reqsMeta[i].length,
		}
		lastAddr := reqsMeta[i].startAddr
		lastLen := reqsMeta[i].length
		for j := i + 1; j < len(reqsMeta); j++ {
			cur := reqsMeta[j]
			// Only group if same register type, type, and contiguous
			if cur.primaryType == g.primaryType && cur.dataType == g.rawType && cur.startAddr == lastAddr+lastLen {
				g.endIdx = j
				g.totalLen += cur.length
				lastAddr = cur.startAddr
				lastLen = cur.length
			} else {
				break
			}
		}
		groups = append(groups, g)
		i = g.endIdx + 1
	}
	return reqsMeta, groups, nil
}

// SplitGroupResponse holds the result of splitting a group read response
type SplitGroupResponse struct {
	OriginalIdx int
	Request     sdkModel.CommandRequest
	Data        []byte
}

// splitGroupReadResponse splits raw data from a grouped read into individual data slices
// for each request in the group. It returns an array of SplitGroupResponse containing
// the original request index, the command request, and the corresponding data slice.
func splitGroupReadResponse(data []byte, group DeviceAddressRange, reqsMeta []ReqWithMeta) ([]SplitGroupResponse, error) {
	results := make([]SplitGroupResponse, 0, group.endIdx-group.startIdx+1)
	offset := uint16(0)

	for k := group.startIdx; k <= group.endIdx; k++ {
		meta := reqsMeta[k]
		cmdInfo, err := createCommandInfo(&meta.req)
		if err != nil {
			return nil, err
		}

		var byteLen uint16
		if cmdInfo.PrimaryTable == HOLDING_REGISTERS || cmdInfo.PrimaryTable == INPUT_REGISTERS {
			byteLen = cmdInfo.Length * 2
		} else {
			byteLen = cmdInfo.Length
		}

		dataStart := offset
		dataEnd := offset + byteLen
		if int(dataEnd) > len(data) {
			return nil, fmt.Errorf("data out of range for split: need %d bytes but only have %d", dataEnd, len(data))
		}

		reqData := make([]byte, byteLen)
		copy(reqData, data[dataStart:dataEnd])

		results = append(results, SplitGroupResponse{
			OriginalIdx: meta.idx,
			Request:     meta.req,
			Data:        reqData,
		})
		offset += byteLen
	}

	return results, nil
}
