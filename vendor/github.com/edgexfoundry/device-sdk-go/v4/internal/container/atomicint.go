// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2025 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package container

import "sync"

type AtomicInt struct {
	mutex sync.RWMutex
	value int
}

func (i *AtomicInt) Value() int {
	i.mutex.RLock()
	defer i.mutex.RUnlock()
	v := i.value
	return v
}

func (i *AtomicInt) Decrease() int {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	i.value--
	return i.value
}

func (i *AtomicInt) Set(v int) {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	i.value = v
}
