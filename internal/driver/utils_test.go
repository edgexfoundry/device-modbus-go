// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022 Schneider Electric
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"testing"
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
