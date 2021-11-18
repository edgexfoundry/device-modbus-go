// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020-2021 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-messaging/v2/messaging"
)

var MessagingClientName = di.TypeInstanceToName((*messaging.MessageClient)(nil))

func MessagingClientFrom(get di.Get) messaging.MessageClient {
	client, ok := get(MessagingClientName).(messaging.MessageClient)
	if !ok {
		return nil
	}

	return client
}
