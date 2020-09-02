// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"

	"github.com/goburrow/serial"
	"github.com/tbrandon/mbserver"
)

const (
	protocolTCP            = "TCP"
	protocolRTU            = "RTU"
	defaultDevicePort      = 1502
	host                   = "0.0.0.0"
	startingPortEnvName    = "STARTING_PORT"
	simulatorNumberEnvName = "SIMULATOR_NUMBER"
	defaultStartingPort    = 10000
	defaultSimulatorNumber = 1000
)

var devices []*mbserver.Server
var mutex = &sync.Mutex{}

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	protocol := flag.String("protocol", protocolTCP, "Modbus protocol")
	flag.Parse()

	_, scalabilityTestMode := os.LookupEnv(startingPortEnvName)

	if scalabilityTestMode {
		err := createScalabilityTestSimulators()
		if err != nil {
			log.Fatalf("Fail to create devices for scalability test, %v. \n", err)
		}

		defer func() {
			for _, device := range devices {
				device.Close()
			}
			log.Printf("Close %d mock devices. \n", len(devices))
		}()

	} else {
		if *protocol == protocolTCP {
			err := createTCPDevice(defaultDevicePort)
			if err != nil {
				log.Fatalf("Fail to create TCP device, %v. \n", err)
			}
		} else {
			err := createRTUDevice()
			if err != nil {
				log.Fatalf("Fail to create RTU device, %v. \n", err)
			}
		}

	}

	<-c
	log.Println("Modbus simulator shutdown.")
}

func createTCPDevice(port int) error {
	device := mbserver.NewServer()
	url := fmt.Sprintf("%s:%d", host, port)
	if err := device.ListenTCP(url); err != nil {
		return fmt.Errorf("the Modbus TCP device cann't listen on %s, %v", url, err)
	}
	devices = append(devices, device)
	log.Printf("Start up a Modbus mock device with address %s \n", url)
	return nil
}

func createRTUDevice() error {
	device := mbserver.NewServer()
	config := &serial.Config{
		Address:  "/tmp/master",
		BaudRate: 19200,
		DataBits: 8,
		StopBits: 1,
		Parity:   "N",
	}
	if err := device.ListenRTU(config); err != nil {
		return fmt.Errorf("the Modbus RTU device cann't listen on %v, %v", config, err)
	}
	devices = append(devices, device)
	log.Printf("Start up a Modbus mock device with address %v \n", config)
	return nil
}

func scaleDevice(startingPort int, simulatorNumber int) (scaledDevicePorts []int, err error) {
	log.Printf("Create simulator, startingPort is %d, simulatorNumber is %d \n", startingPort, simulatorNumber)
	count := 0
	for count < simulatorNumber {
		err := createTCPDevice(startingPort)
		if err != nil {
			return nil, err
		}
		scaledDevicePorts = append(scaledDevicePorts, startingPort)
		startingPort++
		count++
	}
	return scaledDevicePorts, nil
}

func createScalabilityTestSimulators() error {
	var err error
	startingPort := defaultStartingPort
	simulatorNumber := defaultSimulatorNumber
	startingPortEnv, ok := os.LookupEnv(startingPortEnvName)
	if ok {
		startingPort, err = strconv.Atoi(startingPortEnv)
		if err != nil {
			return fmt.Errorf("fail to parse STARTING_PORT %s. \n", startingPortEnv)
		}
	}
	simulatorNumberEnv, ok := os.LookupEnv(simulatorNumberEnvName)
	if ok {
		simulatorNumber, err = strconv.Atoi(simulatorNumberEnv)
		if err != nil {
			return fmt.Errorf("fail to parse SIMULATOR_NUMBER %s. \n", simulatorNumberEnv)
		}
	}

	_, err = scaleDevice(startingPort, simulatorNumber)
	if err != nil {
		return fmt.Errorf("fail to scale %d devices, %v \n", simulatorNumber, err)
	}
	return nil
}
