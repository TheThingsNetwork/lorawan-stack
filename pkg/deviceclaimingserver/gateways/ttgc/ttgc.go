// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package ttgc provides functions to use The Things Gateway Controller.
package ttgc

import (
	"context"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

// Config is the configuration for the client.
type Config struct{}

// TTGC is the client for The Things Gateway Controller.
type TTGC struct {
	config Config
}

// NewClient returns a new TTGC client.
func (c Config) NewClient(context.Context) (*TTGC, error) {
	return &TTGC{
		config: c,
	}, nil
}

var errUnimplemented = errors.DefineUnimplemented("not_implemented", "not implemented")

// Claim implements gateways.GatewayClaimer.
func (TTGC) Claim(context.Context, types.EUI64, string, string) error {
	return errUnimplemented.New()
}

// Unclaim implements gateways.GatewayClaimer.
func (TTGC) Unclaim(context.Context, types.EUI64, string) error {
	return errUnimplemented.New()
}
