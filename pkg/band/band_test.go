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

package band_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/smarty/assertions"
	. "go.thethings.network/lorawan-stack/v3/pkg/band"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestParseChMask(t *testing.T) {
	a := assertions.New(t)
	a.So(ParseChMask(0), should.Resemble, map[uint8]bool{})
	a.So(ParseChMask(42), should.Resemble, map[uint8]bool{})
	a.So(ParseChMask(0, false), should.Resemble, map[uint8]bool{
		0: false,
	})
	a.So(ParseChMask(253, false, true, true), should.Resemble, map[uint8]bool{
		253: false,
		254: true,
		255: true,
	})
	a.So(func() { ParseChMask(253, false, true, true, false) }, should.Panic)
	a.So(ParseChMask(42, true, true, true, false, false, true), should.Resemble, map[uint8]bool{
		42: true,
		43: true,
		44: true,
		45: false,
		46: false,
		47: true,
	})

	for _, tc := range []struct {
		Name           string
		ParseChMask    func(Mask [16]bool, ChMaskCntl uint8) (map[uint8]bool, error)
		Mask           [16]bool
		ChMaskCntl     uint8
		Expected       map[uint8]bool
		ErrorAssertion func(t *testing.T, err error) bool
	}{
		{
			Name:        "16 channels/ChMaskCntl=0",
			ParseChMask: ParseChMask16,
			Mask: [16]bool{
				true, false, false, true, false, false, false, false,
				true, false, true, false, false, false, false, false,
			},
			Expected: ParseChMask(0,
				true, false, false, true, false, false, false, false,
				true, false, true, false, false, false, false, false,
			),
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:        "16 channels/ChMaskCntl=1",
			ParseChMask: ParseChMask16,
			ChMaskCntl:  1,
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.HaveSameErrorDefinitionAs, ErrUnsupportedChMaskCntl.WithAttributes("chmaskcntl", 1))
			},
		},
		{
			Name:        "16 channels/ChMaskCntl=2",
			ParseChMask: ParseChMask16,
			ChMaskCntl:  2,
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.HaveSameErrorDefinitionAs, ErrUnsupportedChMaskCntl.WithAttributes("chmaskcntl", 2))
			},
		},
		{
			Name:        "16 channels/ChMaskCntl=3",
			ParseChMask: ParseChMask16,
			ChMaskCntl:  3,
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.HaveSameErrorDefinitionAs, ErrUnsupportedChMaskCntl.WithAttributes("chmaskcntl", 3))
			},
		},
		{
			Name:        "16 channels/ChMaskCntl=4",
			ParseChMask: ParseChMask16,
			ChMaskCntl:  4,
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.HaveSameErrorDefinitionAs, ErrUnsupportedChMaskCntl.WithAttributes("chmaskcntl", 4))
			},
		},
		{
			Name:        "16 channels/ChMaskCntl=5",
			ParseChMask: ParseChMask16,
			ChMaskCntl:  5,
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.HaveSameErrorDefinitionAs, ErrUnsupportedChMaskCntl.WithAttributes("chmaskcntl", 5))
			},
		},
		{
			Name:        "16 channels/ChMaskCntl=6",
			ParseChMask: ParseChMask16,
			Mask: [16]bool{
				true, false, false, true, false, false, false, false,
				true, false, true, false, false, false, false, false,
			},
			ChMaskCntl: 6,
			Expected: ParseChMask(0,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
			),
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:        "16 channels/ChMaskCntl=7",
			ParseChMask: ParseChMask16,
			ChMaskCntl:  7,
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.HaveSameErrorDefinitionAs, ErrUnsupportedChMaskCntl.WithAttributes("chmaskcntl", 7))
			},
		},
		{
			Name:        "72 channels/ChMaskCntl=0",
			ParseChMask: ParseChMask72,
			Mask: [16]bool{
				true, false, false, true, false, false, false, false,
				true, false, true, false, false, false, false, false,
			},
			Expected: ParseChMask(0,
				true, false, false, true, false, false, false, false,
				true, false, true, false, false, false, false, false,
			),
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:        "72 channels/ChMaskCntl=1",
			ParseChMask: ParseChMask72,
			Mask: [16]bool{
				true, false, false, true, false, false, false, false,
				true, false, true, false, false, false, false, false,
			},
			ChMaskCntl: 1,
			Expected: ParseChMask(16,
				true, false, false, true, false, false, false, false,
				true, false, true, false, false, false, false, false,
			),
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:        "72 channels/ChMaskCntl=2",
			ParseChMask: ParseChMask72,
			Mask: [16]bool{
				true, false, false, true, false, false, false, false,
				true, false, true, false, false, false, false, false,
			},
			ChMaskCntl: 2,
			Expected: ParseChMask(32,
				true, false, false, true, false, false, false, false,
				true, false, true, false, false, false, false, false,
			),
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:        "72 channels/ChMaskCntl=3",
			ParseChMask: ParseChMask72,
			Mask: [16]bool{
				true, false, false, true, false, false, false, false,
				true, false, true, false, false, false, false, false,
			},
			ChMaskCntl: 3,
			Expected: ParseChMask(48,
				true, false, false, true, false, false, false, false,
				true, false, true, false, false, false, false, false,
			),
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:        "72 channels/ChMaskCntl=4",
			ParseChMask: ParseChMask72,
			Mask: [16]bool{
				true, false, false, true, false, false, false, false,
				true, false, true, false, false, false, false, false,
			},
			ChMaskCntl: 4,
			Expected: ParseChMask(64,
				true, false, false, true, false, false, false, false,
			),
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:        "72 channels/ChMaskCntl=5",
			ParseChMask: ParseChMask72,
			Mask: [16]bool{
				true, false, false, true, false, false, false, false,
				true, false, true, false, false, false, false, false,
			},
			ChMaskCntl: 5,
			Expected: ParseChMask(0,
				true, true, true, true, true, true, true, true,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				true, true, true, true, true, true, true, true,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
			),
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:        "72 channels/ChMaskCntl=6",
			ParseChMask: ParseChMask72,
			Mask: [16]bool{
				true, false, false, true, false, false, false, false,
				true, false, true, false, false, false, false, false,
			},
			ChMaskCntl: 6,
			Expected: ParseChMask(0,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, false, false, true, false, false, false, false,
			),
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:        "72 channels/ChMaskCntl=7",
			ParseChMask: ParseChMask72,
			Mask: [16]bool{
				true, false, false, true, false, false, false, false,
				true, false, true, false, false, false, false, false,
			},
			ChMaskCntl: 7,
			Expected: ParseChMask(0,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				true, false, false, true, false, false, false, false,
			),
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:        "72 channels/ChMaskCntl=8",
			ParseChMask: ParseChMask72,
			Mask: [16]bool{
				true, false, false, true, false, false, false, false,
				true, false, true, false, false, false, false, false,
			},
			ChMaskCntl: 8,
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.HaveSameErrorDefinitionAs, ErrUnsupportedChMaskCntl.WithAttributes("chmaskcntl", 8))
			},
		},
		{
			Name:        "96 channels/ChMaskCntl=0",
			ParseChMask: ParseChMask96,
			Mask: [16]bool{
				true, false, false, true, false, false, false, false,
				true, false, true, false, false, false, false, false,
			},
			Expected: ParseChMask(0,
				true, false, false, true, false, false, false, false,
				true, false, true, false, false, false, false, false,
			),
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:        "96 channels/ChMaskCntl=1",
			ParseChMask: ParseChMask96,
			Mask: [16]bool{
				true, false, false, true, false, false, false, false,
				true, false, true, false, false, false, false, false,
			},
			ChMaskCntl: 1,
			Expected: ParseChMask(16,
				true, false, false, true, false, false, false, false,
				true, false, true, false, false, false, false, false,
			),
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:        "96 channels/ChMaskCntl=2",
			ParseChMask: ParseChMask96,
			Mask: [16]bool{
				true, false, false, true, false, false, false, false,
				true, false, true, false, false, false, false, false,
			},
			ChMaskCntl: 2,
			Expected: ParseChMask(32,
				true, false, false, true, false, false, false, false,
				true, false, true, false, false, false, false, false,
			),
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:        "96 channels/ChMaskCntl=3",
			ParseChMask: ParseChMask96,
			Mask: [16]bool{
				true, false, false, true, false, false, false, false,
				true, false, true, false, false, false, false, false,
			},
			ChMaskCntl: 3,
			Expected: ParseChMask(48,
				true, false, false, true, false, false, false, false,
				true, false, true, false, false, false, false, false,
			),
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:        "96 channels/ChMaskCntl=4",
			ParseChMask: ParseChMask96,
			Mask: [16]bool{
				true, false, false, true, false, false, false, false,
				true, false, true, false, false, false, false, false,
			},
			ChMaskCntl: 4,
			Expected: ParseChMask(64,
				true, false, false, true, false, false, false, false,
				true, false, true, false, false, false, false, false,
			),
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:        "96 channels/ChMaskCntl=5",
			ParseChMask: ParseChMask96,
			Mask: [16]bool{
				true, false, false, true, false, false, false, false,
				true, false, true, false, false, false, false, false,
			},
			ChMaskCntl: 5,
			Expected: ParseChMask(80,
				true, false, false, true, false, false, false, false,
				true, false, true, false, false, false, false, false,
			),
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:        "96 channels/ChMaskCntl=6",
			ParseChMask: ParseChMask96,
			Mask: [16]bool{
				true, false, false, true, false, false, false, false,
				true, false, true, false, false, false, false, false,
			},
			ChMaskCntl: 6,
			Expected: ParseChMask(0,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
			),
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:        "96 channels/ChMaskCntl=7",
			ParseChMask: ParseChMask96,
			ChMaskCntl:  7,
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.HaveSameErrorDefinitionAs, ErrUnsupportedChMaskCntl.WithAttributes("chmaskcntl", 7))
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			var mask [16]bool
			copy(mask[:], tc.Mask[:])
			res, err := tc.ParseChMask(mask, tc.ChMaskCntl)
			a.So(mask, should.Equal, tc.Mask)
			if a.So(tc.ErrorAssertion(t, err), should.BeTrue) {
				a.So(res, should.Resemble, tc.Expected)
			}
		})
	}
}

func TestGenerateChMask(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		Name            string
		Generate        func([]bool, []bool) ([]ChMaskCntlPair, error)
		CurrentChannels []bool
		DesiredChannels []bool
		Expected        []ChMaskCntlPair
		ErrorAssertion  func(t *testing.T, err error) bool
	}{
		// NOTE: GenerateChMask16 always generates singleton ChMaskCntlPair slice regardless of CurrentChannels.
		{
			Name:     "16 channels/2,4",
			Generate: GenerateChMask16,
			CurrentChannels: []bool{
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
			},
			DesiredChannels: []bool{
				false, true, false, true, false, false, false, false,
				false, false, false, false, false, false, false, false,
			},
			Expected: []ChMaskCntlPair{
				{
					Mask: [16]bool{
						false, true, false, true, false, false, false, false,
						false, false, false, false, false, false, false, false,
					},
				},
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				t.Helper()
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:     "16 channels/1-16",
			Generate: GenerateChMask16,
			CurrentChannels: []bool{
				true, true, false, true, false, true, true, true,
				true, true, true, false, true, true, false, true,
			},
			DesiredChannels: []bool{
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
			},
			Expected: []ChMaskCntlPair{
				{
					Mask: [16]bool{
						true, true, true, true, true, true, true, true,
						true, true, true, true, true, true, true, true,
					},
				},
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				t.Helper()
				return assertions.New(t).So(err, should.BeNil)
			},
		},

		{
			Name:     "72 channels/no cntl5/current(1-72)/desired(1-72)",
			Generate: MakeGenerateChMask72(false, true),
			CurrentChannels: []bool{
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
			},
			DesiredChannels: []bool{
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
			},
			Expected: []ChMaskCntlPair{
				{
					Mask: [16]bool{
						true, true, true, true, true, true, true, true,
						true, true, true, true, true, true, true, true,
					},
				},
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				t.Helper()
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:     "72 channels/cntl5/current(1-72)/desired(1-72)",
			Generate: MakeGenerateChMask72(true, true),
			CurrentChannels: []bool{
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
			},
			DesiredChannels: []bool{
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
			},
			Expected: []ChMaskCntlPair{
				{
					Mask: [16]bool{
						true, true, true, true, true, true, true, true,
						true, true, true, true, true, true, true, true,
					},
				},
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				t.Helper()
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:     "72 channels/no cntl5/current:(1-16,42,67,69);desired:(1-16,42,67,69)",
			Generate: MakeGenerateChMask72(false, true),
			CurrentChannels: []bool{
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, true, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, true, false, true, false, false, false,
			},
			DesiredChannels: []bool{
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, true, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, true, false, true, false, false, false,
			},
			Expected: []ChMaskCntlPair{
				{
					Mask: [16]bool{
						true, true, true, true, true, true, true, true,
						true, true, true, true, true, true, true, true,
					},
				},
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				t.Helper()
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:     "72 channels/cntl5/current:(1-16,42,67,69);desired:(1-16,42,67,69)",
			Generate: MakeGenerateChMask72(true, true),
			CurrentChannels: []bool{
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, true, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, true, false, true, false, false, false,
			},
			DesiredChannels: []bool{
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, true, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, true, false, true, false, false, false,
			},
			Expected: []ChMaskCntlPair{
				{
					Mask: [16]bool{
						true, true, true, true, true, true, true, true,
						true, true, true, true, true, true, true, true,
					},
				},
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				t.Helper()
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:     "72 channels/no cntl5/current:(1-4,6-16,42,67,69);desired:(1-16,42,67,69)",
			Generate: MakeGenerateChMask72(false, true),
			CurrentChannels: []bool{
				false, false, false, false, true, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, true, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, true, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, true, false, true, false, false, false,
			},
			DesiredChannels: []bool{
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				false, false, false, false, false, false, false, false,
				false, false, true, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, true, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, true, false, true, false, false, false,
			},
			Expected: []ChMaskCntlPair{
				{
					Mask: [16]bool{
						true, true, true, true, true, true, true, true,
						true, true, true, true, true, true, true, true,
					},
				},
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				t.Helper()
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:     "72 channels/cntl5/current:(1-4,6-16,42,67,69);desired:(1-16,42,67,69)",
			Generate: MakeGenerateChMask72(false, true),
			CurrentChannels: []bool{
				false, false, false, false, true, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, true, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, true, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, true, false, true, false, false, false,
			},
			DesiredChannels: []bool{
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				false, false, false, false, false, false, false, false,
				false, false, true, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, true, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, true, false, true, false, false, false,
			},
			Expected: []ChMaskCntlPair{
				{
					Mask: [16]bool{
						true, true, true, true, true, true, true, true,
						true, true, true, true, true, true, true, true,
					},
				},
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				t.Helper()
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:     "72 channels/no cntl5/current(1-12,14-33,36-42,44-72)/desired(1-69)",
			Generate: MakeGenerateChMask72(false, true),
			CurrentChannels: []bool{
				true, true, true, true, true, true, true, true,
				true, true, true, true, false, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, false, false, true, true, true, true, true,
				true, true, false, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
			},
			DesiredChannels: []bool{
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, false, false, false,
			},
			Expected: []ChMaskCntlPair{
				{
					Cntl: 6,
					Mask: [16]bool{
						true, true, true, true, true, false, false, false,
						false, false, false, false, false, false, false, false,
					},
				},
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				t.Helper()
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:     "72 channels/cntl5/current(1-12,14-33,36-42,44-72)/desired(1-69)",
			Generate: MakeGenerateChMask72(false, true),
			CurrentChannels: []bool{
				true, true, true, true, true, true, true, true,
				true, true, true, true, false, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, false, false, true, true, true, true, true,
				true, true, false, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
			},
			DesiredChannels: []bool{
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, false, false, false,
			},
			Expected: []ChMaskCntlPair{
				{
					Cntl: 6,
					Mask: [16]bool{
						true, true, true, true, true, false, false, false,
						false, false, false, false, false, false, false, false,
					},
				},
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				t.Helper()
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:     "72 channels/no cntl5/current(1-12,14-33,36-42,44-71)/desired(1-3,5-72)",
			Generate: MakeGenerateChMask72(false, true),
			CurrentChannels: []bool{
				true, true, true, true, true, true, true, true,
				true, true, true, true, false, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, false, false, true, true, true, true, true,
				true, true, false, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, false,
			},
			DesiredChannels: []bool{
				true, true, true, false, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
			},
			Expected: []ChMaskCntlPair{
				{
					Cntl: 6,
					Mask: [16]bool{
						true, true, true, true, true, true, true, true,
						false, false, false, false, false, false, false, false,
					},
				},
				{
					Mask: [16]bool{
						true, true, true, false, true, true, true, true,
						true, true, true, true, true, true, true, true,
					},
				},
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				t.Helper()
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:     "72 channels/cntl5/current(1-12,14-33,36-42,44-71)/desired(1-3,5-72)",
			Generate: MakeGenerateChMask72(false, true),
			CurrentChannels: []bool{
				true, true, true, true, true, true, true, true,
				true, true, true, true, false, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, false, false, true, true, true, true, true,
				true, true, false, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, false,
			},
			DesiredChannels: []bool{
				true, true, true, false, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
			},
			Expected: []ChMaskCntlPair{
				{
					Cntl: 6,
					Mask: [16]bool{
						true, true, true, true, true, true, true, true,
						false, false, false, false, false, false, false, false,
					},
				},
				{
					Mask: [16]bool{
						true, true, true, false, true, true, true, true,
						true, true, true, true, true, true, true, true,
					},
				},
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				t.Helper()
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:     "72 channels/no cntl5/current(1-12,14-33,36-42,44-63,65-72)/desired(1-3,5-72)",
			Generate: MakeGenerateChMask72(false, true),
			CurrentChannels: []bool{
				true, true, true, true, true, true, true, true,
				true, true, true, true, false, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, false, false, true, true, true, true, true,
				true, true, false, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, false,
				true, true, true, true, true, true, true, true,
			},
			DesiredChannels: []bool{
				true, true, true, false, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
			},
			Expected: []ChMaskCntlPair{
				{
					Cntl: 6,
					Mask: [16]bool{
						true, true, true, true, true, true, true, true,
						false, false, false, false, false, false, false, false,
					},
				},
				{
					Mask: [16]bool{
						true, true, true, false, true, true, true, true,
						true, true, true, true, true, true, true, true,
					},
				},
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				t.Helper()
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:     "72 channels/cntl5/current(1-12,14-33,36-42,44-63,65-72)/desired(1-3,5-72)",
			Generate: MakeGenerateChMask72(false, true),
			CurrentChannels: []bool{
				true, true, true, true, true, true, true, true,
				true, true, true, true, false, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, false, false, true, true, true, true, true,
				true, true, false, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, false,
				true, true, true, true, true, true, true, true,
			},
			DesiredChannels: []bool{
				true, true, true, false, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
			},
			Expected: []ChMaskCntlPair{
				{
					Cntl: 6,
					Mask: [16]bool{
						true, true, true, true, true, true, true, true,
						false, false, false, false, false, false, false, false,
					},
				},
				{
					Mask: [16]bool{
						true, true, true, false, true, true, true, true,
						true, true, true, true, true, true, true, true,
					},
				},
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				t.Helper()
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:     "72 channels/no cntl5/current(1-72)/desired(9-16,65-72)",
			Generate: MakeGenerateChMask72(false, true),
			CurrentChannels: []bool{
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
			},
			DesiredChannels: []bool{
				false, false, false, false, false, false, false, false,
				true, true, true, true, true, true, true, true,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				true, true, true, true, true, true, true, true,
			},
			Expected: []ChMaskCntlPair{
				{
					Cntl: 7,
					Mask: [16]bool{
						true, true, true, true, true, true, true, true,
						false, false, false, false, false, false, false, false,
					},
				},
				{
					Mask: [16]bool{
						false, false, false, false, false, false, false, false,
						true, true, true, true, true, true, true, true,
					},
				},
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				t.Helper()
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:     "72 channels/cntl5/current(1-72)/desired(9-16,65-72)",
			Generate: MakeGenerateChMask72(true, true),
			CurrentChannels: []bool{
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
			},
			DesiredChannels: []bool{
				false, false, false, false, false, false, false, false,
				true, true, true, true, true, true, true, true,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				true, true, true, true, true, true, true, true,
			},
			Expected: []ChMaskCntlPair{
				{
					Cntl: 5,
					Mask: [16]bool{
						false, true, false, false, false, false, false, false,
						false, false, false, false, false, false, false, false,
					},
				},
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				t.Helper()
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:     "72 channels/no cntl5/current(1-72)/desired(9-24)",
			Generate: MakeGenerateChMask72(false, true),
			CurrentChannels: []bool{
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
			},
			DesiredChannels: []bool{
				false, false, false, false, false, false, false, false,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
			},
			Expected: []ChMaskCntlPair{
				{
					Cntl: 7,
				},
				{
					Mask: [16]bool{
						false, false, false, false, false, false, false, false,
						true, true, true, true, true, true, true, true,
					},
				},
				{
					Cntl: 1,
					Mask: [16]bool{
						true, true, true, true, true, true, true, true,
						false, false, false, false, false, false, false, false,
					},
				},
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				t.Helper()
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:     "72 channels/no cntl5/non-atomic/current(1-72)/desired(40-48)",
			Generate: MakeGenerateChMask72(false, false),
			CurrentChannels: []bool{
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
			},
			DesiredChannels: []bool{
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				true, true, true, true, true, true, true, true,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
			},
			Expected: []ChMaskCntlPair{
				{
					Cntl: 2,
					Mask: [16]bool{
						false, false, false, false, false, false, false, false,
						true, true, true, true, true, true, true, true,
					},
				},
				{
					Cntl: 4,
					Mask: [16]bool{
						false, false, false, false, false, false, false, false,
						false, false, false, false, false, false, false, false,
					},
				},
				{
					Cntl: 3,
					Mask: [16]bool{
						false, false, false, false, false, false, false, false,
						false, false, false, false, false, false, false, false,
					},
				},
				{
					Cntl: 1,
					Mask: [16]bool{
						false, false, false, false, false, false, false, false,
						false, false, false, false, false, false, false, false,
					},
				},
				{
					Mask: [16]bool{
						false, false, false, false, false, false, false, false,
						false, false, false, false, false, false, false, false,
					},
				},
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				t.Helper()
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:     "72 channels/no cntl5/non-atomic/current(1-72)/desired(40-48+70)",
			Generate: MakeGenerateChMask72(false, false),
			CurrentChannels: []bool{
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
			},
			DesiredChannels: []bool{
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				true, true, true, true, true, true, true, true,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, true, false, false,
			},
			Expected: []ChMaskCntlPair{
				{
					Cntl: 2,
					Mask: [16]bool{
						false, false, false, false, false, false, false, false,
						true, true, true, true, true, true, true, true,
					},
				},
				{
					Cntl: 4,
					Mask: [16]bool{
						false, false, false, false, false, true, false, false,
						false, false, false, false, false, false, false, false,
					},
				},
				{
					Cntl: 3,
					Mask: [16]bool{
						false, false, false, false, false, false, false, false,
						false, false, false, false, false, false, false, false,
					},
				},
				{
					Cntl: 1,
					Mask: [16]bool{
						false, false, false, false, false, false, false, false,
						false, false, false, false, false, false, false, false,
					},
				},
				{
					Mask: [16]bool{
						false, false, false, false, false, false, false, false,
						false, false, false, false, false, false, false, false,
					},
				},
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				t.Helper()
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:     "72 channels/cntl5/current(1-72)/desired(9-24)",
			Generate: MakeGenerateChMask72(true, true),
			CurrentChannels: []bool{
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
			},
			DesiredChannels: []bool{
				false, false, false, false, false, false, false, false,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
			},
			Expected: []ChMaskCntlPair{
				{
					Cntl: 5,
					Mask: [16]bool{
						false, true, true, false, false, false, false, false,
						false, false, false, false, false, false, false, false,
					},
				},
				{
					Cntl: 4,
					Mask: [16]bool{
						false, false, false, false, false, false, false, false,
						false, false, false, false, false, false, false, false,
					},
				},
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				t.Helper()
				return assertions.New(t).So(err, should.BeNil)
			},
		},

		{
			Name:     "96 channels/current(1-96)/desired(1-96)",
			Generate: GenerateChMask96,
			CurrentChannels: []bool{
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
			},
			DesiredChannels: []bool{
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
			},
			Expected: []ChMaskCntlPair{
				{
					Mask: [16]bool{
						true, true, true, true, true, true, true, true,
						true, true, true, true, true, true, true, true,
					},
				},
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				t.Helper()
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:     "96 channels/current(1-12,14-33,36-42,44-96)/desired(1-96)",
			Generate: GenerateChMask96,
			CurrentChannels: []bool{
				true, true, true, true, true, true, true, true,
				true, true, true, true, false, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, false, false, true, true, true, true, true,
				true, true, false, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
			},
			DesiredChannels: []bool{
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
			},
			Expected: []ChMaskCntlPair{
				{
					Cntl: 6,
				},
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				t.Helper()
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:     "96 channels/current(1-12,14-33,36-42,44-95)/desired(1-3,5-96)",
			Generate: GenerateChMask96,
			CurrentChannels: []bool{
				true, true, true, true, true, true, true, true,
				true, true, true, true, false, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, false, false, true, true, true, true, true,
				true, true, false, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, false,
			},
			DesiredChannels: []bool{
				true, true, true, false, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
			},
			Expected: []ChMaskCntlPair{
				{
					Cntl: 6,
				},
				{
					Mask: [16]bool{
						true, true, true, false, true, true, true, true,
						true, true, true, true, true, true, true, true,
					},
				},
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				t.Helper()
				return assertions.New(t).So(err, should.BeNil)
			},
		},
	} {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			a := assertions.New(t)

			current := append(tc.CurrentChannels[:0:0], tc.CurrentChannels...)
			desired := append(tc.DesiredChannels[:0:0], tc.DesiredChannels...)
			res, err := tc.Generate(current, desired)
			a.So(current, should.Resemble, tc.CurrentChannels)
			a.So(desired, should.Resemble, tc.DesiredChannels)
			if a.So(tc.ErrorAssertion(t, err), should.BeTrue) {
				a.So(res, should.Resemble, tc.Expected)
			}
		})
	}
}

func TestCompareDatarates(t *testing.T) {
	t.Parallel()
	a := assertions.New(t)

	for _, tc := range []struct {
		Name   string
		A      *ttnpb.DataRate
		B      *ttnpb.DataRate
		Strict bool

		Expected bool
	}{
		{
			Name: "lorawan strict",
			A: &ttnpb.DataRate{Modulation: &ttnpb.DataRate_Lora{Lora: &ttnpb.LoRaDataRate{
				Bandwidth:       1,
				SpreadingFactor: 2,
				CodingRate:      Cr4_5,
			}}},
			B: &ttnpb.DataRate{Modulation: &ttnpb.DataRate_Lora{Lora: &ttnpb.LoRaDataRate{
				Bandwidth:       1,
				SpreadingFactor: 2,
				CodingRate:      Cr4_5,
			}}},
			Strict: true,

			Expected: true,
		},
		{
			Name: "lorawan strict not-equal",
			A: &ttnpb.DataRate{Modulation: &ttnpb.DataRate_Lora{Lora: &ttnpb.LoRaDataRate{
				Bandwidth:       1,
				SpreadingFactor: 2,
			}}},
			B: &ttnpb.DataRate{Modulation: &ttnpb.DataRate_Lora{Lora: &ttnpb.LoRaDataRate{
				Bandwidth:       1,
				SpreadingFactor: 2,
				CodingRate:      Cr4_5,
			}}},
			Strict: true,

			Expected: false,
		},
		{
			Name: "lorawan",
			A: &ttnpb.DataRate{Modulation: &ttnpb.DataRate_Lora{Lora: &ttnpb.LoRaDataRate{
				Bandwidth:       1,
				SpreadingFactor: 2,
			}}},
			B: &ttnpb.DataRate{Modulation: &ttnpb.DataRate_Lora{Lora: &ttnpb.LoRaDataRate{
				Bandwidth:       1,
				SpreadingFactor: 2,
				CodingRate:      Cr4_5,
			}}},
			Strict: false,

			Expected: true,
		},
		{
			Name: "fsk",
			A: &ttnpb.DataRate{Modulation: &ttnpb.DataRate_Fsk{Fsk: &ttnpb.FSKDataRate{
				BitRate: 1,
			}}},
			B: &ttnpb.DataRate{Modulation: &ttnpb.DataRate_Fsk{Fsk: &ttnpb.FSKDataRate{
				BitRate: 1,
			}}},

			Expected: true,
		},
		{
			Name: "lr-fhss",
			A: &ttnpb.DataRate{Modulation: &ttnpb.DataRate_Lrfhss{Lrfhss: &ttnpb.LRFHSSDataRate{
				ModulationType:        1,
				OperatingChannelWidth: 2,
				CodingRate:            Cr4_5,
			}}},
			B: &ttnpb.DataRate{Modulation: &ttnpb.DataRate_Lrfhss{Lrfhss: &ttnpb.LRFHSSDataRate{
				ModulationType:        1,
				OperatingChannelWidth: 2,
				CodingRate:            Cr4_5,
			}}},

			Expected: true,
		},
		{
			Name: "lorawan - lr-fhss",
			A: &ttnpb.DataRate{Modulation: &ttnpb.DataRate_Lora{Lora: &ttnpb.LoRaDataRate{
				Bandwidth:       1,
				SpreadingFactor: 2,
			}}},
			B: &ttnpb.DataRate{Modulation: &ttnpb.DataRate_Lrfhss{Lrfhss: &ttnpb.LRFHSSDataRate{
				ModulationType:        1,
				OperatingChannelWidth: 2,
				CodingRate:            Cr4_5,
			}}},

			Expected: false,
		},
	} {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			res := CompareDataRates(tc.A, tc.B, tc.Strict)

			if !a.So(res, should.Equal, tc.Expected) {
				t.Fatalf("Unexpected outcome received. Expected :%v, got: %v", tc.Expected, res)
			}
		})
	}
}

func TestRx1DataRate(t *testing.T) {
	for _, tc := range []struct {
		bandID string

		validIndexes []ttnpb.DataRateIndex
		validOffsets []ttnpb.DataRateOffset

		invalidIndexes []ttnpb.DataRateIndex
		invalidOffsets []ttnpb.DataRateOffset
	}{
		{
			bandID:         "AU_915_928",
			validIndexes:   []ttnpb.DataRateIndex{0, 1, 2, 3, 4, 5, 6, 7},
			invalidIndexes: []ttnpb.DataRateIndex{8, 9, 10, 11, 12, 13, 14, 15},
			validOffsets:   []ttnpb.DataRateOffset{0, 1, 2, 3, 4, 5},
			invalidOffsets: []ttnpb.DataRateOffset{6, 7},
		},
		{
			bandID:       "AS_923",
			validIndexes: []ttnpb.DataRateIndex{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
			validOffsets: []ttnpb.DataRateOffset{0, 1, 2, 3, 4, 5, 6, 7},
		},
		{
			bandID:         "CN_470_510",
			validIndexes:   []ttnpb.DataRateIndex{0, 1, 2, 3, 4, 5},
			invalidIndexes: []ttnpb.DataRateIndex{6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
			validOffsets:   []ttnpb.DataRateOffset{0, 1, 2, 3, 4, 5},
			invalidOffsets: []ttnpb.DataRateOffset{6, 7},
		},
		{
			bandID:         "CN_779_787",
			validIndexes:   []ttnpb.DataRateIndex{0, 1, 2, 3, 4, 5, 6, 7},
			invalidIndexes: []ttnpb.DataRateIndex{8, 9, 10, 11, 12, 13, 14, 15},
			validOffsets:   []ttnpb.DataRateOffset{0, 1, 2, 3, 4, 5},
			invalidOffsets: []ttnpb.DataRateOffset{6, 7},
		},
		{
			bandID:         "EU_433",
			validIndexes:   []ttnpb.DataRateIndex{0, 1, 2, 3, 4, 5, 6, 7},
			invalidIndexes: []ttnpb.DataRateIndex{8, 9, 10, 11, 12, 13, 14, 15},
			validOffsets:   []ttnpb.DataRateOffset{0, 1, 2, 3, 4, 5},
			invalidOffsets: []ttnpb.DataRateOffset{6, 7},
		},
		{
			bandID:         "EU_863_870",
			validIndexes:   []ttnpb.DataRateIndex{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
			invalidIndexes: []ttnpb.DataRateIndex{12, 13, 14, 15},
			validOffsets:   []ttnpb.DataRateOffset{0, 1, 2, 3, 4, 5},
			invalidOffsets: []ttnpb.DataRateOffset{6, 7},
		},
		{
			bandID:       "IN_865_867",
			validIndexes: []ttnpb.DataRateIndex{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
			validOffsets: []ttnpb.DataRateOffset{0, 1, 2, 3, 4, 5, 6, 7},
		},
		{
			bandID:         "KR_920_923",
			validIndexes:   []ttnpb.DataRateIndex{0, 1, 2, 3, 4, 5},
			invalidIndexes: []ttnpb.DataRateIndex{6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
			validOffsets:   []ttnpb.DataRateOffset{0, 1, 2, 3, 4, 5},
			invalidOffsets: []ttnpb.DataRateOffset{6, 7},
		},
		{
			bandID:         "RU_864_870",
			validIndexes:   []ttnpb.DataRateIndex{0, 1, 2, 3, 4, 5, 6, 7},
			invalidIndexes: []ttnpb.DataRateIndex{8, 9, 10, 11, 12, 13, 14, 15},
			validOffsets:   []ttnpb.DataRateOffset{0, 1, 2, 3, 4, 5},
			invalidOffsets: []ttnpb.DataRateOffset{6, 7},
		},
		{
			bandID:         "US_902_928",
			validIndexes:   []ttnpb.DataRateIndex{0, 1, 2, 3, 4, 5, 6},
			invalidIndexes: []ttnpb.DataRateIndex{7, 8, 9, 10, 11, 12, 13, 14, 15},
			validOffsets:   []ttnpb.DataRateOffset{0, 1, 2, 3},
			invalidOffsets: []ttnpb.DataRateOffset{4, 5, 6, 7},
		},
	} {
		t.Run(tc.bandID, func(t *testing.T) {
			a := assertions.New(t)

			b, err := GetLatest(tc.bandID)
			if !a.So(err, should.BeNil) {
				t.Fatalf("Error when getting band %s: %s", tc.bandID, err)
			}

			for _, validIndex := range tc.validIndexes {
				for _, validOffset := range tc.validOffsets {
					_, err := b.Rx1DataRate(validIndex, validOffset, true)
					if !a.So(err, should.BeNil) {
						t.Fatalf("Computing Rx1 data rate should have succeeded with index %d and offset %d", validIndex, validOffset)
					}
				}
			}
			for _, invalidIndex := range tc.invalidIndexes {
				for _, offset := range append(tc.validOffsets, tc.invalidOffsets...) {
					_, err := b.Rx1DataRate(invalidIndex, offset, true)
					if !a.So(err, should.NotBeNil) {
						t.Fatalf("Computing Rx1 data rate should not have succeeded with index %d and offset %d", invalidIndex, offset)
					}
				}
			}
			for _, index := range append(tc.validIndexes, tc.invalidIndexes...) {
				for _, invalidOffset := range tc.invalidOffsets {
					_, err := b.Rx1DataRate(index, invalidOffset, true)
					if !a.So(err, should.NotBeNil) {
						t.Fatalf("Computing Rx1 data rate should not have succeeded with index %d and offset %d", index, invalidOffset)
					}
				}
			}
		})
	}
}

func TestParseChMaskBands(t *testing.T) {
	a := assertions.New(t)

	for version, versions := range All {
		for _, b := range versions {
			if !a.So(b.ParseChMask, should.NotBeNil) {
				t.Fatalf("Band %s:%v should have a ParseChMask function defined", b.ID, version)
			}
		}
	}
}

func TestGenerateChMasksBands(t *testing.T) {
	a := assertions.New(t)

	for version, versions := range All {
		for _, b := range versions {
			if !a.So(b.GenerateChMasks, should.NotBeNil) {
				t.Fatalf("Band %s:%v should have a GenerateChMasks function defined", b.ID, version)
			}
		}
	}
}

func TestFindSubBand(t *testing.T) {
	for version, versions := range All {
		for _, b := range versions {
			t.Run(fmt.Sprintf("%v:%v", b.ID, version), func(t *testing.T) {
				a := assertions.New(t)
				for _, ch := range b.UplinkChannels {
					sb, ok := b.FindSubBand(ch.Frequency)
					if !a.So(ok, should.BeTrue) {
						t.Fatalf("Frequency %d not found in sub-bands", ch.Frequency)
					}
					a.So(sb.MinFrequency, should.BeLessThanOrEqualTo, ch.Frequency)
					a.So(sb.MaxFrequency, should.BeGreaterThanOrEqualTo, ch.Frequency)
				}
			})
		}
	}
}

func TestFindDataRate(t *testing.T) {
	t.Parallel()
	a := assertions.New(t)

	// US_902_928
	testBand, _ := Get(US_902_928, ttnpb.PHYVersion_RP001_V1_0_2_REV_B)
	dr := &ttnpb.DataRate{
		Modulation: &ttnpb.DataRate_Lora{
			Lora: &ttnpb.LoRaDataRate{
				Bandwidth:       500000,
				SpreadingFactor: 8,
				CodingRate:      Cr4_5,
			},
		},
	}
	index, _, ok := testBand.FindDownlinkDataRate(dr)
	a.So(ok, should.BeTrue)
	if index != ttnpb.DataRateIndex_DATA_RATE_12 {
		t.Fatalf("Invalid index, expected 12. Got %d", index)
	}

	dr = &ttnpb.DataRate{
		Modulation: &ttnpb.DataRate_Lora{
			Lora: &ttnpb.LoRaDataRate{
				Bandwidth:       500000,
				SpreadingFactor: 8,
				CodingRate:      Cr4_5,
			},
		},
	}
	index, _, ok = testBand.FindUplinkDataRate(dr)
	a.So(ok, should.BeTrue)
	if index != ttnpb.DataRateIndex_DATA_RATE_4 {
		t.Fatalf("Invalid index, expected 4. Got %d", index)
	}

	// AU_915_928
	testBand, _ = Get(AU_915_928, ttnpb.PHYVersion_RP001_V1_0_3_REV_A)
	dr = &ttnpb.DataRate{
		Modulation: &ttnpb.DataRate_Lora{
			Lora: &ttnpb.LoRaDataRate{
				Bandwidth:       500000,
				SpreadingFactor: 12,
				CodingRate:      Cr4_5,
			},
		},
	}
	index, _, ok = testBand.FindDownlinkDataRate(dr)
	a.So(ok, should.BeTrue)
	if index != ttnpb.DataRateIndex_DATA_RATE_8 {
		t.Fatalf("Invalid index, expected 8. Got %d", index)
	}

	dr = &ttnpb.DataRate{
		Modulation: &ttnpb.DataRate_Lora{
			Lora: &ttnpb.LoRaDataRate{
				Bandwidth:       500000,
				SpreadingFactor: 12,
				CodingRate:      Cr4_5,
			},
		},
	}
	index, _, ok = testBand.FindUplinkDataRate(dr)
	a.So(ok, should.BeTrue)
	if index != ttnpb.DataRateIndex_DATA_RATE_8 {
		t.Fatalf("Invalid index, expected 8. Got %d", index)
	}
}

func TestBeacon(t *testing.T) {
	t.Parallel()

	for name, version := range LatestVersion {
		b := All[name][version]
		t.Run(fmt.Sprintf("%v/%v", name, version), func(t *testing.T) {
			t.Parallel()

			a := assertions.New(t)
			beaconDR, ok := b.DataRates[b.Beacon.DataRateIndex]
			if a.So(ok, should.BeTrue) {
				// As of L2 1.0.4, the beacons are guaranteed to be LoRa modulated with a spreading
				// factor between 8 and 12.
				a.So(beaconDR.Rate.GetLora(), should.NotBeNil)
				a.So(beaconDR.Rate.GetLora().GetSpreadingFactor(), should.BeBetweenOrEqual, 8, 12)
			}
		})
	}
}

func TestStrictCodingRateSanityCheck(t *testing.T) {
	t.Parallel()
	for bandID, versions := range All {
		for version, b := range versions {
			bandID, version, b := bandID, version, b

			t.Run(fmt.Sprintf("%v/%v", bandID, version), func(t *testing.T) {
				t.Parallel()
				if version >= ttnpb.PHYVersion_RP002_V1_0_0 ||
					strings.HasPrefix(bandID, "MA") ||
					strings.HasPrefix(bandID, "ISM") {
					if !b.StrictCodingRate {
						t.Errorf("Strict coding rate doesn't match expected. Want true, got %v.", b.StrictCodingRate)
					}
				} else if strings.HasPrefix(bandID, "US") && version >= ttnpb.PHYVersion_PHY_V1_0_2_REV_A ||
					strings.HasPrefix(bandID, "AU") ||
					strings.HasPrefix(bandID, "CN_470") {
					if !b.StrictCodingRate {
						t.Errorf("Strict coding rate doesn't match expected. Want true, got %v.", b.StrictCodingRate)
					}
				} else if b.StrictCodingRate {
					t.Errorf("Strict coding rate doesn't match expected. Want false, got %v.", b.StrictCodingRate)
				}
			})
		}
	}
}

func TestChannelsWellDefined(t *testing.T) {
	t.Parallel()

	for name, versions := range All {
		for version, b := range versions {
			b := b
			t.Run(fmt.Sprintf("%v/%v", name, version), func(t *testing.T) {
				t.Parallel()

				a := assertions.New(t)
				assertCh := func(ch Channel) {
					a.So(ch.MinDataRate, should.BeLessThanOrEqualTo, ch.MaxDataRate)
					a.So(b.DataRates, should.ContainKey, ch.MinDataRate)
					a.So(b.DataRates, should.ContainKey, ch.MaxDataRate)
				}

				for _, ch := range b.UplinkChannels {
					assertCh(ch)
				}
				for _, ch := range b.DownlinkChannels {
					assertCh(ch)
				}
			})
		}
	}
}

func TestSubBandsWellDefined(t *testing.T) {
	t.Parallel()

	for name, versions := range All {
		for version, b := range versions {
			b := b
			t.Run(fmt.Sprintf("%v/%v", name, version), func(t *testing.T) {
				t.Parallel()

				checkSubBand := func(ch Channel) bool {
					for _, sb := range b.SubBands {
						if sb.MinFrequency <= ch.Frequency && ch.Frequency <= sb.MaxFrequency {
							return true
						}
					}
					return false
				}

				a := assertions.New(t)
				for _, ch := range b.UplinkChannels {
					a.So(checkSubBand(ch), should.BeTrue)
				}
				for _, ch := range b.DownlinkChannels {
					a.So(checkSubBand(ch), should.BeTrue)
				}
			})
		}
	}
}
