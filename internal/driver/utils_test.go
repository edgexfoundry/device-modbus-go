// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022 Schneider Electric
// Copyright (C) 2026 encentive GmbH
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
)

func Test_castStartingAddress(t *testing.T) {
	tests := []struct {
		name    string
		i       interface{}
		want    uint16
		wantErr bool
	}{
		{
			name: "OK - integer provided",
			i:    42,
			want: uint16(42),
		},
		{
			name: "OK - string provided",
			i:    "42",
			want: uint16(42),
		},
		{
			name:    "NOK - uncastable string provided",
			i:       "test",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := castStartingAddress(tt.i)
			if err != nil && !tt.wantErr || err == nil && tt.wantErr {
				t.Errorf("castStartingAddress() got error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("castStartingAddress() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_normalizeRawType(t *testing.T) {
	tests := []struct {
		name    string
		rawType string
		want    string
		wantErr bool
	}{
		{name: "OK - Uint32", rawType: "Uint32", want: common.ValueTypeUint32},
		{name: "OK - uint32", rawType: "uint32", want: common.ValueTypeUint32},
		{name: "OK - Uint64", rawType: "Uint64", want: common.ValueTypeUint64},
		{name: "NOK - unsupported", rawType: "Float80", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeRawType(tt.rawType)
			if err != nil && !tt.wantErr || err == nil && tt.wantErr {
				t.Errorf("normalizeRawType() got error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("normalizeRawType() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_castSwapAttribute(t *testing.T) {
	tests := []struct {
		name    string
		i       interface{}
		want    bool
		wantErr bool
	}{
		{
			name: "OK - boolean provided",
			i:    true,
			want: true,
		},
		{
			name: "OK - string provided",
			i:    "false",
			want: false,
		},
		{
			name:    "NOK - uncastable string provided",
			i:       "test",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := castSwapAttribute(tt.i)
			if err != nil && !tt.wantErr || err == nil && tt.wantErr {
				t.Errorf("castSwapAttribute() got error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("castSwapAttribute() got = %v, want %v", got, tt.want)
			}
		})
	}
}
