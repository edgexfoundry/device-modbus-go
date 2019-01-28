// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

// Package driver is used to execute device-sdk's commands
package driver

import (
	"fmt"
	"sync"

	sdkModel "github.com/edgexfoundry/device-sdk-go/pkg/models"
	"github.com/edgexfoundry/edgex-go/pkg/clients/logging"
	"github.com/edgexfoundry/edgex-go/pkg/models"
)

var once sync.Once
var driver *Driver

type Driver struct {
	Logger     logger.LoggingClient
	AsyncCh    chan<- *sdkModel.AsyncValues
	addressMap sync.Map
}

func (*Driver) DisconnectDevice(address *models.Addressable) error {
	panic("implement me")
}

// lockAddress mark address is unavailable because real device handle one request at a time
func (d *Driver) lockAddress(address *models.Addressable) {
	lock, ok := d.addressMap.Load(address.Address)
	if !ok {
		lock = make(chan bool, 1)
		d.addressMap.Store(address.Address, lock)
	}

	lock.(chan bool) <- true
}

// unlockAddress remove token after command finish
func (d *Driver) unlockAddress(address *models.Addressable) {
	lock, _ := d.addressMap.Load(address.Address)
	<-lock.(chan bool)
}

func (d *Driver) HandleReadCommands(addr *models.Addressable, reqs []sdkModel.CommandRequest) ([]*sdkModel.CommandValue, error) {
	d.lockAddress(addr)
	defer d.unlockAddress(addr)

	var responses = make([]*sdkModel.CommandValue, len(reqs))
	var deviceClient DeviceClient

	// Check request's attribute to avoid interface cast error
	err := checkAttributes(reqs)
	if err != nil {
		driver.Logger.Info(fmt.Sprintf("CommandRequest's attribute are invalid. err:%v \n", err))
		return responses, err
	}

	// create device client and open connection
	connectionInfo, err := createConnectionInfo(*addr)
	if err != nil {
		driver.Logger.Info(fmt.Sprintf("Read command createConnectionInfo failed. err:%v \n", err))
		return responses, err
	}

	deviceClient, err = NewDeviceClient(connectionInfo)
	if err != nil {
		driver.Logger.Info(fmt.Sprintf("Read command NewDeviceClient failed. err:%v \n", err))
		return responses, err
	}

	err = deviceClient.OpenConnection()
	if err != nil {
		driver.Logger.Info(fmt.Sprintf("Read command OpenConnection failed. err:%v \n", err))
		return responses, err
	}

	defer deviceClient.CloseConnection()

	// handle command requests
	for i, req := range reqs {
		res, err := d.handleReadCommandRequest(deviceClient, req)
		if err != nil {
			driver.Logger.Info(fmt.Sprintf("Read command failed. Cmd:%v err:%v \n", req.DeviceObject.Name, err))
			return responses, err
		}

		responses[i] = res
	}

	return responses, nil
}

func (d *Driver) handleReadCommandRequest(deviceClient DeviceClient, req sdkModel.CommandRequest) (*sdkModel.CommandValue, error) {
	var response []byte
	var result = &sdkModel.CommandValue{}
	var err error

	commandInfo := createCommandInfo(req.DeviceObject)

	response, err = deviceClient.GetValue(commandInfo)
	if err != nil {
		return result, err
	}

	//stringResult := TransformDateBytesToString(response, commandInfo)
	result, err = TransformDateBytesToResult(&req.RO, response, commandInfo)

	if err != nil {
		return result, err
	} else {
		driver.Logger.Info(fmt.Sprintf("Read command finished. Cmd:%v, %v \n", req.DeviceObject.Name, result))
	}

	return result, nil
}

func (d *Driver) HandleWriteCommands(addr *models.Addressable, reqs []sdkModel.CommandRequest, params []*sdkModel.CommandValue) error {
	d.lockAddress(addr)
	defer d.unlockAddress(addr)

	var deviceClient DeviceClient

	// Check request's attribute to avoid interface cast error
	err := checkAttributes(reqs)
	if err != nil {
		driver.Logger.Info(fmt.Sprintf("CommandRequest's attribute are invalid. err:%v \n", err))
		return err
	}

	// create device client and open connection
	connectionInfo, err := createConnectionInfo(*addr)
	if err != nil {
		return err
	}

	deviceClient, err = NewDeviceClient(connectionInfo)
	if err != nil {
		return err
	}

	err = deviceClient.OpenConnection()
	if err != nil {
		return err
	}

	defer deviceClient.CloseConnection()

	// handle command requests
	for i, req := range reqs {
		err = d.handleWriteCommandRequest(deviceClient, req, params[i])
		if err != nil {
			d.Logger.Error(err.Error())
			break
		}
	}

	return err
}

func (d *Driver) handleWriteCommandRequest(deviceClient DeviceClient, req sdkModel.CommandRequest, param *sdkModel.CommandValue) error {
	var err error

	commandInfo := createCommandInfo(req.DeviceObject)

	dataBytes, err := TransformCommandValueToDataBytes(commandInfo, param)
	if err != nil {
		return fmt.Errorf("transform command value failed, err: %v", err)
	}

	err = deviceClient.SetValue(commandInfo, dataBytes)
	if err != nil {
		return fmt.Errorf("handle write command request failed, err: %v", err)
	}

	driver.Logger.Info(fmt.Sprintf("Write command finished. Cmd:%v \n", req.DeviceObject.Name))
	return nil
}

func (d *Driver) Initialize(lc logger.LoggingClient, asyncCh chan<- *sdkModel.AsyncValues) error {
	d.Logger = lc
	d.AsyncCh = asyncCh
	return nil
}

func (*Driver) Stop(force bool) error {
	panic("implement me")
}

func NewProtocolDriver() sdkModel.ProtocolDriver {
	once.Do(func() {
		driver = new(Driver)
	})
	return driver
}
