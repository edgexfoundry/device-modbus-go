// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/goburrow/serial"
	"github.com/tbrandon/mbserver"
)

const (
	ProtocolTCP = "TCP"
	ProtocolRTU = "RTU"
	ModbusPort  = "1502"
)

var port, protocol *string

// Run a Modbus TCP or RTU simulator
// In RTU protocol, you can use socat to simulate serial port:
// socat -d -d -d -d pty,link=/tmp/master,raw,echo=0 pty,link=/tmp/slave,raw,echo=0
func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	protocol = flag.String("protocol", ProtocolTCP, "Modbus protocol")
	port = flag.String("port", ModbusPort, "Modbus port")

	server := mbserver.NewServer()
	defer func() {
		server.Close()
		fmt.Printf("Simulator shutdown. \n")
	}()

	if *protocol == ProtocolTCP {
		url := fmt.Sprintf("0.0.0.0:%s", *port)
		fmt.Printf("Modbus TCP address: %v \n", url)

		if err := server.ListenTCP(url); err != nil {
			fmt.Printf("Failed to start the Modbus TCP server, %v\n", err)
		}
	} else {
		config := &serial.Config{
			Address:  "/tmp/master",
			BaudRate: 19200,
			DataBits: 8,
			StopBits: 1,
			Parity:   "N",
		}
		fmt.Printf("Modbus RTU config %v \n", config)

		if err := server.ListenRTU(config); err != nil {
			fmt.Printf("Failed to start the Modbus RTU server, %v\n", err)
		}
	}

	fmt.Printf("Start up a Modbus %s simulator. \n", *protocol)
	<-c
}
