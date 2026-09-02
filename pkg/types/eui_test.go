// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

package types_test

import (
	"encoding/json"
	"testing"

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	. "go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

var (
	_ config.Configurable = &EUI64Prefix{}
	_ config.Stringer     = EUI64Prefix{}
	_ config.Configurable = &EUI64Range{}
	_ config.Stringer     = EUI64Range{}
)

func TestEUI64(t *testing.T) {
	a := assertions.New(t)

	eui := EUI64{0x26, 0x12, 0x34, 0x56, 0x42, 0x42, 0x42, 0x42}

	prefix := EUI64Prefix{EUI64{0x26}, 7}
	a.So(prefix.Matches(eui), should.BeTrue)

	// Prefix list
	{
		addr := EUI64{1, 2, 3, 4}
		a.So(addr.HasPrefix(EUI64Prefix{EUI64{0, 0, 0, 0}, 0}), should.BeTrue)
		a.So(addr.HasPrefix(EUI64Prefix{EUI64{1, 2, 3, 0}, 24}), should.BeTrue)
		a.So(addr.HasPrefix(EUI64Prefix{EUI64{2, 2, 3, 4}, 31}), should.BeFalse)
		a.So(addr.HasPrefix(EUI64Prefix{EUI64{1, 1, 3, 4}, 31}), should.BeFalse)
		a.So(addr.HasPrefix(EUI64Prefix{EUI64{1, 1, 1, 1}, 15}), should.BeFalse)
	}

	t.Run("JSON", func(t *testing.T) {
		a := assertions.New(t)

		const encodedEUI = `"2612345642424242"`

		b, err := json.Marshal(eui)
		if a.So(err, should.BeNil) {
			a.So(string(b), should.Equal, encodedEUI)
		}

		var decodedEUI EUI64
		err = json.Unmarshal([]byte(encodedEUI), &decodedEUI)
		if a.So(err, should.BeNil) {
			a.So(decodedEUI, should.Equal, eui)
		}

		const encodedPrefix = `"2600000000000000/7"`

		b, err = json.Marshal(prefix)
		if a.So(err, should.BeNil) {
			a.So(string(b), should.Equal, encodedPrefix)
		}

		var decodedPrefix EUI64Prefix
		err = json.Unmarshal([]byte(encodedPrefix), &decodedPrefix)
		if a.So(err, should.BeNil) {
			a.So(decodedPrefix, should.Equal, prefix)
		}
	})

	t.Run("Text", func(t *testing.T) {
		a := assertions.New(t)

		const encodedEUI = `2612345642424242`

		b, err := eui.MarshalText()
		if a.So(err, should.BeNil) {
			a.So(string(b), should.Equal, encodedEUI)
		}

		var decodedEUI EUI64
		err = decodedEUI.UnmarshalText([]byte(encodedEUI))
		if a.So(err, should.BeNil) {
			a.So(decodedEUI, should.Equal, eui)
		}

		const encodedPrefix = `2600000000000000/7`

		b, err = prefix.MarshalText()
		if a.So(err, should.BeNil) {
			a.So(string(b), should.Equal, encodedPrefix)
		}

		var decodedPrefix EUI64Prefix
		err = decodedPrefix.UnmarshalText([]byte(encodedPrefix))
		if a.So(err, should.BeNil) {
			a.So(decodedPrefix, should.Equal, prefix)
		}
	})

	t.Run("Number", func(t *testing.T) {
		a := assertions.New(t)

		const encodedEUI uint64 = 2743312668105523778

		n := eui.MarshalNumber()
		a.So(n, should.Resemble, encodedEUI)

		var decodedEUI EUI64
		decodedEUI.UnmarshalNumber(encodedEUI)
		a.So(decodedEUI, should.Equal, eui)
	})
}

func TestParseEUI64Range(t *testing.T) {
	t.Parallel()
	a := assertions.New(t)

	for _, tc := range []struct {
		value    string
		str      string // Expected canonical string, if different from value.
		contains []EUI64
		excludes []EUI64
	}{
		{
			value: "001616FFFE42DFAD-001616FFFE42E395",
			contains: []EUI64{
				{0x00, 0x16, 0x16, 0xFF, 0xFE, 0x42, 0xDF, 0xAD},
				{0x00, 0x16, 0x16, 0xFF, 0xFE, 0x42, 0xE0, 0x00},
				{0x00, 0x16, 0x16, 0xFF, 0xFE, 0x42, 0xE3, 0x95},
			},
			excludes: []EUI64{
				{0x00, 0x16, 0x16, 0xFF, 0xFE, 0x42, 0xDF, 0xAC},
				{0x00, 0x16, 0x16, 0xFF, 0xFE, 0x42, 0xE3, 0x96},
			},
		},
		{
			value: "001616FFFE300500-001616FFFE30FFFF",
			contains: []EUI64{
				{0x00, 0x16, 0x16, 0xFF, 0xFE, 0x30, 0x05, 0x00},
				{0x00, 0x16, 0x16, 0xFF, 0xFE, 0x30, 0xFF, 0xFF},
			},
			excludes: []EUI64{
				{0x00, 0x16, 0x16, 0xFF, 0xFE, 0x30, 0x04, 0xFF},
				{0x00, 0x16, 0x16, 0xFF, 0xFE, 0x31, 0x00, 0x00},
			},
		},
		{
			value: "001616FFFE2A32AA-001616FFFE2F5454",
			contains: []EUI64{
				{0x00, 0x16, 0x16, 0xFF, 0xFE, 0x2A, 0x32, 0xAA},
				{0x00, 0x16, 0x16, 0xFF, 0xFE, 0x2B, 0x00, 0x00},
				{0x00, 0x16, 0x16, 0xFF, 0xFE, 0x2F, 0x54, 0x54},
			},
			excludes: []EUI64{
				{0x00, 0x16, 0x16, 0xFF, 0xFE, 0x2A, 0x32, 0xA9},
				{0x00, 0x16, 0x16, 0xFF, 0xFE, 0x2F, 0x54, 0x55},
			},
		},
		{
			value: "58A0CBFFFE800000/48",
			contains: []EUI64{
				{0x58, 0xA0, 0xCB, 0xFF, 0xFE, 0x80, 0x00, 0x00},
				{0x58, 0xA0, 0xCB, 0xFF, 0xFE, 0x80, 0xFF, 0xFF},
			},
			excludes: []EUI64{
				{0x58, 0xA0, 0xCB, 0xFF, 0xFE, 0x81, 0x00, 0x00},
				{0x58, 0xA0, 0xCB, 0xFF, 0xFD, 0x80, 0x00, 0x00},
			},
		},
		{
			value: "EC656EFFFE000000/40",
			contains: []EUI64{
				{0xEC, 0x65, 0x6E, 0xFF, 0xFE, 0x00, 0x00, 0x00},
				{0xEC, 0x65, 0x6E, 0xFF, 0xFE, 0xFF, 0xFF, 0xFF},
			},
			excludes: []EUI64{
				{0xEC, 0x65, 0x6E, 0xFF, 0xFF, 0x00, 0x00, 0x00},
				{0xEC, 0x65, 0x6D, 0xFF, 0xFE, 0x00, 0x00, 0x00},
			},
		},
		// A prefix of the full EUI64 length contains only the exact EUI64.
		{
			value: "58A0CBFFFE800000/64",
			contains: []EUI64{
				{0x58, 0xA0, 0xCB, 0xFF, 0xFE, 0x80, 0x00, 0x00},
			},
			excludes: []EUI64{
				{0x58, 0xA0, 0xCB, 0xFF, 0xFE, 0x80, 0x00, 0x01},
				{0x58, 0xA0, 0xCB, 0xFF, 0xFE, 0x7F, 0xFF, 0xFF},
			},
		},
		// A zero length prefix contains all EUI64s.
		{
			value: "0000000000000000/0",
			contains: []EUI64{
				{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
				{0x58, 0xA0, 0xCB, 0xFF, 0xFE, 0x80, 0x00, 0x00},
				{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
			},
		},
		// Bits beyond the prefix length are masked out.
		{
			value: "58A0CBFFFE80FFFF/48",
			str:   "58A0CBFFFE800000/48",
			contains: []EUI64{
				{0x58, 0xA0, 0xCB, 0xFF, 0xFE, 0x80, 0x00, 0x00},
				{0x58, 0xA0, 0xCB, 0xFF, 0xFE, 0x80, 0xFF, 0xFF},
			},
			excludes: []EUI64{
				{0x58, 0xA0, 0xCB, 0xFF, 0xFE, 0x81, 0x00, 0x00},
			},
		},
		// Lowercase hexadecimal is accepted and canonicalized to uppercase.
		{
			value: "58a0cbfffe800000/48",
			str:   "58A0CBFFFE800000/48",
			contains: []EUI64{
				{0x58, 0xA0, 0xCB, 0xFF, 0xFE, 0x80, 0x00, 0x00},
			},
		},
		{
			value: "001616fffe42dfad-001616fffe42e395",
			str:   "001616FFFE42DFAD-001616FFFE42E395",
			contains: []EUI64{
				{0x00, 0x16, 0x16, 0xFF, 0xFE, 0x42, 0xE0, 0x00},
			},
		},
		// The start of the range equals the end.
		{
			value: "58A0CBFFFE800000-58A0CBFFFE800000",
			contains: []EUI64{
				{0x58, 0xA0, 0xCB, 0xFF, 0xFE, 0x80, 0x00, 0x00},
			},
			excludes: []EUI64{
				{0x58, 0xA0, 0xCB, 0xFF, 0xFE, 0x7F, 0xFF, 0xFF},
				{0x58, 0xA0, 0xCB, 0xFF, 0xFE, 0x80, 0x00, 0x01},
			},
		},
		// The range covers the full EUI64 space without overflowing.
		{
			value: "0000000000000000-FFFFFFFFFFFFFFFF",
			contains: []EUI64{
				{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
				{0x58, 0xA0, 0xCB, 0xFF, 0xFE, 0x80, 0x00, 0x00},
				{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
			},
		},
	} {
		r, err := ParseEUI64Range(tc.value)
		if !a.So(err, should.BeNil) {
			t.Fatalf("Failed to parse %q: %v", tc.value, err)
		}
		for _, eui := range tc.contains {
			a.So(r.Contains(eui), should.BeTrue)
		}
		for _, eui := range tc.excludes {
			a.So(r.Contains(eui), should.BeFalse)
		}
		// The string representation must be canonical and must round-trip.
		str := tc.str
		if str == "" {
			str = tc.value
		}
		a.So(r.String(), should.Equal, str)
		roundtrip, err := ParseEUI64Range(str)
		if a.So(err, should.BeNil) {
			a.So(roundtrip, should.Resemble, r)
		}
		var configured EUI64Range
		if err := configured.UnmarshalConfigString(tc.value); a.So(err, should.BeNil) {
			a.So(configured, should.Resemble, r)
		}
	}

	for _, value := range []string{
		"",
		"58A0CBFFFE800000",
		"58A0CBFFFE800000/",
		// The prefix length is out of bounds.
		"58A0CBFFFE800000/65",
		"58A0CBFFFE800000/99",
		"58A0CBFFFE800000/-1",
		"58A0CBFFFE800000/4X",
		"58A0CBFFFE800000/123456",
		// The EUI64 of the prefix is too short or too long.
		"58A0CBFFFE8000/48",
		"58A0CBFFFE80000000/48",
		// The range has too many or empty parts.
		"58A0CBFFFE800000-58A0CBFFFE800000-58A0CBFFFE800000",
		"-58A0CBFFFE800000",
		"58A0CBFFFE800000-",
		"-",
		// The range bounds are not valid EUI64s.
		"001616FFFEWXUSDA-001616FFFETGENDE",
		"001616FFFE42DFAD-001616FFFETGENDE",
		"58A0CBFFFE8000-58A0CBFFFE80FFFF",
		" 58A0CBFFFE800000-58A0CBFFFE80FFFF",
		// The start of the range is after the end.
		"001616FFFE42E395-001616FFFE42DFAD",
		"FFFFFFFFFFFFFFFF-0000000000000000",
	} {
		_, err := ParseEUI64Range(value)
		if !a.So(err, should.NotBeNil) {
			t.Fatalf("Expected error for %q", value)
		}
	}
}

func TestParseEUI64RangesMap(t *testing.T) {
	t.Parallel()
	a := assertions.New(t)

	m, err := ParseEUI64RangesMap(map[string][]string{
		"ttgc": {
			"58A0CBFFFE800000/48",
			"001616FFFE42DFAD-001616FFFE42E395",
		},
		"semtech-rjs": {
			"EC656EFFFE000000/40",
		},
	})
	if a.So(err, should.BeNil) {
		a.So(m, should.Resemble, map[string][]EUI64Range{
			"ttgc": {
				EUI64Prefix{
					EUI64:  EUI64{0x58, 0xA0, 0xCB, 0xFF, 0xFE, 0x80, 0x00, 0x00},
					Length: 48,
				}.EUI64Range(),
				EUI64RangeFromInterval(
					EUI64{0x00, 0x16, 0x16, 0xFF, 0xFE, 0x42, 0xDF, 0xAD},
					EUI64{0x00, 0x16, 0x16, 0xFF, 0xFE, 0x42, 0xE3, 0x95},
				),
			},
			"semtech-rjs": {
				EUI64Prefix{
					EUI64:  EUI64{0xEC, 0x65, 0x6E, 0xFF, 0xFE, 0x00, 0x00, 0x00},
					Length: 40,
				}.EUI64Range(),
			},
		})
	}

	_, err = ParseEUI64RangesMap(map[string][]string{"ttgc": {"invalid"}})
	a.So(err, should.NotBeNil)

	m, err = ParseEUI64RangesMap(nil)
	if a.So(err, should.BeNil) {
		a.So(m, should.BeEmpty)
	}
}
