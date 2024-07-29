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

// Package types provides types for the Device Claiming Server.
package types

import "go.thethings.network/lorawan-stack/v3/pkg/types"

// EUI64Range is a range of EUI64s.
type EUI64Range interface {
	// Contains returns true if the EUI64 is in the range.
	Contains(types.EUI64) bool
}

type eui64PrefixRange types.EUI64Prefix

var _ EUI64Range = eui64PrefixRange{}

// Contains implements EUI64Range.
func (r eui64PrefixRange) Contains(eui types.EUI64) bool {
	return eui.HasPrefix(types.EUI64Prefix(r))
}

// RangeFromEUI64Prefix returns a range that contains all EUI64s with the given prefix.
func RangeFromEUI64Prefix(prefix types.EUI64Prefix) EUI64Range {
	return eui64PrefixRange(prefix)
}

type eui64Range struct {
	start, end uint64
}

var _ EUI64Range = eui64Range{}

// Contains implements EUI64Range.
func (r eui64Range) Contains(eui types.EUI64) bool {
	n := eui.MarshalNumber()
	return n >= r.start && n <= r.end
}

// RangeFromEUI64Range returns a range that contains all EUI64s between start and end.
func RangeFromEUI64Range(start, end types.EUI64) EUI64Range {
	return eui64Range{
		start: start.MarshalNumber(),
		end:   end.MarshalNumber(),
	}
}
