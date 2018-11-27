// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"github.com/goburrow/serial"
	"github.com/tbrandon/mbserver"
)

func main() {
	go runMdbusTCPserver()
	go runMdbusRTUserver()
	select {}
}

// runMdbusTCPserver run a simulate modbus tcp server
func runMdbusTCPserver() {
	serv := mbserver.NewServer()
	err := serv.ListenTCP("0.0.0.0:1502")
	if err != nil {
		fmt.Printf("%v\n", err)
	}

	defer serv.Close()
	fmt.Println("Run Mdbus TCP server.")
	select {}
}

// runMdbusRTUserver run a simulate modbus RTU server
//
// you can use socat to simulate serial port
// socat -d -d -d -d pty,link=/tmp/master,raw,echo=0 pty,link=/tmp/slave,raw,echo=0
func runMdbusRTUserver() {
	serv := mbserver.NewServer()
	err := serv.ListenRTU(&serial.Config{
		Address:  "/tmp/master",
		BaudRate: 19200,
		DataBits: 8,
		StopBits: 1,
		Parity:   "N",
	})
	if err != nil {
		fmt.Printf("failed to listen, got %v\n", err)
	}

	defer serv.Close()
	fmt.Println("Run Mdbus RTU server.")
	select {}
}
