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
	"net/http"
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
	simulatorServicePort   = 1503
	host                   = "0.0.0.0"
	startingPortEnvName    = "STARTING_PORT"
	simulatorNumberEnvName = "SIMULATOR_NUMBER"
	defaultStartingPort    = 10000
	defaultSimulatorNumber = 1000
)

var devices []*mbserver.Server
var mutex = &sync.Mutex{}
var mu sync.Mutex
var readingAmount uint

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

		log.Println("Start a HTTP server to provide the reading count.")
		http.HandleFunc("/reading/count/reset", func(w http.ResponseWriter, r *http.Request) {
			readingAmount = 0
		})

		http.HandleFunc("/reading/count", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "%d", readingAmount)
		})

		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", simulatorServicePort), nil))

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
	overrideRegisterFunctionHandler(device)
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
	overrideRegisterFunctionHandler(device)
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

func increaseReadingAmount() {
	mu.Lock()
	readingAmount++
	mu.Unlock()
}

// Override the modbus function https://github.com/tbrandon/mbserver/blob/master/functions.go
func overrideRegisterFunctionHandler(server *mbserver.Server) {
	// coils
	server.RegisterFunctionHandler(1,
		func(s *mbserver.Server, frame mbserver.Framer) ([]byte, *mbserver.Exception) {
			increaseReadingAmount()
			return mbserver.ReadCoils(s, frame)
		})
	// discrete inputs
	server.RegisterFunctionHandler(2,
		func(s *mbserver.Server, frame mbserver.Framer) ([]byte, *mbserver.Exception) {
			increaseReadingAmount()
			return mbserver.ReadDiscreteInputs(s, frame)
		})
	// holding registers
	server.RegisterFunctionHandler(3,
		func(s *mbserver.Server, frame mbserver.Framer) ([]byte, *mbserver.Exception) {
			increaseReadingAmount()
			return mbserver.ReadHoldingRegisters(s, frame)
		})
	// input registers
	server.RegisterFunctionHandler(4,
		func(s *mbserver.Server, frame mbserver.Framer) ([]byte, *mbserver.Exception) {
			increaseReadingAmount()
			return mbserver.ReadInputRegisters(s, frame)
		})
}
