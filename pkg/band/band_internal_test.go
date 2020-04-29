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

package band

import (
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestParseChMask(t *testing.T) {
	a := assertions.New(t)
	a.So(parseChMask(0), should.Resemble, map[uint8]bool{})
	a.So(parseChMask(42), should.Resemble, map[uint8]bool{})
	a.So(parseChMask(0, false), should.Resemble, map[uint8]bool{
		0: false,
	})
	a.So(parseChMask(253, false, true, true), should.Resemble, map[uint8]bool{
		253: false,
		254: true,
		255: true,
	})
	a.So(func() { parseChMask(253, false, true, true, false) }, should.Panic)
	a.So(parseChMask(42, true, true, true, false, false, true), should.Resemble, map[uint8]bool{
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
			ParseChMask: parseChMask16,
			Mask: [16]bool{
				true, false, false, true, false, false, false, false,
				true, false, true, false, false, false, false, false,
			},
			Expected: parseChMask(0,
				true, false, false, true, false, false, false, false,
				true, false, true, false, false, false, false, false,
			),
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:        "16 channels/ChMaskCntl=1",
			ParseChMask: parseChMask16,
			ChMaskCntl:  1,
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.HaveSameErrorDefinitionAs, errUnsupportedChMaskCntl.WithAttributes("chmaskcntl", 1))
			},
		},
		{
			Name:        "16 channels/ChMaskCntl=2",
			ParseChMask: parseChMask16,
			ChMaskCntl:  2,
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.HaveSameErrorDefinitionAs, errUnsupportedChMaskCntl.WithAttributes("chmaskcntl", 2))
			},
		},
		{
			Name:        "16 channels/ChMaskCntl=3",
			ParseChMask: parseChMask16,
			ChMaskCntl:  3,
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.HaveSameErrorDefinitionAs, errUnsupportedChMaskCntl.WithAttributes("chmaskcntl", 3))
			},
		},
		{
			Name:        "16 channels/ChMaskCntl=4",
			ParseChMask: parseChMask16,
			ChMaskCntl:  4,
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.HaveSameErrorDefinitionAs, errUnsupportedChMaskCntl.WithAttributes("chmaskcntl", 4))
			},
		},
		{
			Name:        "16 channels/ChMaskCntl=5",
			ParseChMask: parseChMask16,
			ChMaskCntl:  5,
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.HaveSameErrorDefinitionAs, errUnsupportedChMaskCntl.WithAttributes("chmaskcntl", 5))
			},
		},
		{
			Name:        "16 channels/ChMaskCntl=6",
			ParseChMask: parseChMask16,
			Mask: [16]bool{
				true, false, false, true, false, false, false, false,
				true, false, true, false, false, false, false, false,
			},
			ChMaskCntl: 6,
			Expected: parseChMask(0,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
			),
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:        "16 channels/ChMaskCntl=7",
			ParseChMask: parseChMask16,
			ChMaskCntl:  7,
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.HaveSameErrorDefinitionAs, errUnsupportedChMaskCntl.WithAttributes("chmaskcntl", 7))
			},
		},
		{
			Name:        "72 channels/ChMaskCntl=0",
			ParseChMask: parseChMask72,
			Mask: [16]bool{
				true, false, false, true, false, false, false, false,
				true, false, true, false, false, false, false, false,
			},
			Expected: parseChMask(0,
				true, false, false, true, false, false, false, false,
				true, false, true, false, false, false, false, false,
			),
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:        "72 channels/ChMaskCntl=1",
			ParseChMask: parseChMask72,
			Mask: [16]bool{
				true, false, false, true, false, false, false, false,
				true, false, true, false, false, false, false, false,
			},
			ChMaskCntl: 1,
			Expected: parseChMask(16,
				true, false, false, true, false, false, false, false,
				true, false, true, false, false, false, false, false,
			),
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:        "72 channels/ChMaskCntl=2",
			ParseChMask: parseChMask72,
			Mask: [16]bool{
				true, false, false, true, false, false, false, false,
				true, false, true, false, false, false, false, false,
			},
			ChMaskCntl: 2,
			Expected: parseChMask(32,
				true, false, false, true, false, false, false, false,
				true, false, true, false, false, false, false, false,
			),
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:        "72 channels/ChMaskCntl=3",
			ParseChMask: parseChMask72,
			Mask: [16]bool{
				true, false, false, true, false, false, false, false,
				true, false, true, false, false, false, false, false,
			},
			ChMaskCntl: 3,
			Expected: parseChMask(48,
				true, false, false, true, false, false, false, false,
				true, false, true, false, false, false, false, false,
			),
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:        "72 channels/ChMaskCntl=4",
			ParseChMask: parseChMask72,
			Mask: [16]bool{
				true, false, false, true, false, false, false, false,
				true, false, true, false, false, false, false, false,
			},
			ChMaskCntl: 4,
			Expected: parseChMask(64,
				true, false, false, true, false, false, false, false,
			),
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:        "72 channels/ChMaskCntl=5",
			ParseChMask: parseChMask72,
			Mask: [16]bool{
				true, false, false, true, false, false, false, false,
				true, false, true, false, false, false, false, false,
			},
			ChMaskCntl: 5,
			Expected: parseChMask(0,
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
			ParseChMask: parseChMask72,
			Mask: [16]bool{
				true, false, false, true, false, false, false, false,
				true, false, true, false, false, false, false, false,
			},
			ChMaskCntl: 6,
			Expected: parseChMask(0,
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
			Name:        "72 channels/ChMaskCntl=7",
			ParseChMask: parseChMask72,
			Mask: [16]bool{
				true, false, false, true, false, false, false, false,
				true, false, true, false, false, false, false, false,
			},
			ChMaskCntl: 7,
			Expected: parseChMask(0,
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
			Name:        "72 channels/ChMaskCntl=8",
			ParseChMask: parseChMask72,
			Mask: [16]bool{
				true, false, false, true, false, false, false, false,
				true, false, true, false, false, false, false, false,
			},
			ChMaskCntl: 8,
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.HaveSameErrorDefinitionAs, errUnsupportedChMaskCntl.WithAttributes("chmaskcntl", 8))
			},
		},
		{
			Name:        "96 channels/ChMaskCntl=0",
			ParseChMask: parseChMask96,
			Mask: [16]bool{
				true, false, false, true, false, false, false, false,
				true, false, true, false, false, false, false, false,
			},
			Expected: parseChMask(0,
				true, false, false, true, false, false, false, false,
				true, false, true, false, false, false, false, false,
			),
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:        "96 channels/ChMaskCntl=1",
			ParseChMask: parseChMask96,
			Mask: [16]bool{
				true, false, false, true, false, false, false, false,
				true, false, true, false, false, false, false, false,
			},
			ChMaskCntl: 1,
			Expected: parseChMask(16,
				true, false, false, true, false, false, false, false,
				true, false, true, false, false, false, false, false,
			),
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:        "96 channels/ChMaskCntl=2",
			ParseChMask: parseChMask96,
			Mask: [16]bool{
				true, false, false, true, false, false, false, false,
				true, false, true, false, false, false, false, false,
			},
			ChMaskCntl: 2,
			Expected: parseChMask(32,
				true, false, false, true, false, false, false, false,
				true, false, true, false, false, false, false, false,
			),
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:        "96 channels/ChMaskCntl=3",
			ParseChMask: parseChMask96,
			Mask: [16]bool{
				true, false, false, true, false, false, false, false,
				true, false, true, false, false, false, false, false,
			},
			ChMaskCntl: 3,
			Expected: parseChMask(48,
				true, false, false, true, false, false, false, false,
				true, false, true, false, false, false, false, false,
			),
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:        "96 channels/ChMaskCntl=4",
			ParseChMask: parseChMask96,
			Mask: [16]bool{
				true, false, false, true, false, false, false, false,
				true, false, true, false, false, false, false, false,
			},
			ChMaskCntl: 4,
			Expected: parseChMask(64,
				true, false, false, true, false, false, false, false,
				true, false, true, false, false, false, false, false,
			),
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:        "96 channels/ChMaskCntl=5",
			ParseChMask: parseChMask96,
			Mask: [16]bool{
				true, false, false, true, false, false, false, false,
				true, false, true, false, false, false, false, false,
			},
			ChMaskCntl: 5,
			Expected: parseChMask(80,
				true, false, false, true, false, false, false, false,
				true, false, true, false, false, false, false, false,
			),
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:        "96 channels/ChMaskCntl=6",
			ParseChMask: parseChMask96,
			Mask: [16]bool{
				true, false, false, true, false, false, false, false,
				true, false, true, false, false, false, false, false,
			},
			ChMaskCntl: 6,
			Expected: parseChMask(0,
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
			ParseChMask: parseChMask96,
			ChMaskCntl:  7,
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.HaveSameErrorDefinitionAs, errUnsupportedChMaskCntl.WithAttributes("chmaskcntl", 7))
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
	for _, tc := range []struct {
		Name            string
		Generate        func([]bool, []bool) ([]ChMaskCntlPair, error)
		CurrentChannels []bool
		DesiredChannels []bool
		Expected        []ChMaskCntlPair
		ErrorAssertion  func(t *testing.T, err error) bool
	}{
		// NOTE: generateChMask16 always generates singleton ChMaskCntlPair slice regardless of CurrentChannels.
		{
			Name:     "16 channels/2,4",
			Generate: generateChMask16,
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
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:     "16 channels/1-16",
			Generate: generateChMask16,
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
				return assertions.New(t).So(err, should.BeNil)
			},
		},

		{
			Name:     "72 channels/no cntl5/current(1-72)/desired(1-72)",
			Generate: makeGenerateChMask72(false),
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
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:     "72 channels/cntl5/current(1-72)/desired(1-72)",
			Generate: makeGenerateChMask72(true),
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
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:     "72 channels/no cntl5/current:(1-16,42,67,69);desired:(1-16,42,67,69)",
			Generate: makeGenerateChMask72(false),
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
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:     "72 channels/cntl5/current:(1-16,42,67,69);desired:(1-16,42,67,69)",
			Generate: makeGenerateChMask72(true),
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
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:     "72 channels/no cntl5/current:(1-4,6-16,42,67,69);desired:(1-16,42,67,69)",
			Generate: makeGenerateChMask72(false),
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
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:     "72 channels/cntl5/current:(1-4,6-16,42,67,69);desired:(1-16,42,67,69)",
			Generate: makeGenerateChMask72(false),
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
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:     "72 channels/no cntl5/current(1-12,14-33,36-42,44-72)/desired(1-69)",
			Generate: makeGenerateChMask72(false),
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
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:     "72 channels/cntl5/current(1-12,14-33,36-42,44-72)/desired(1-69)",
			Generate: makeGenerateChMask72(false),
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
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:     "72 channels/no cntl5/current(1-12,14-33,36-42,44-71)/desired(1-3,5-72)",
			Generate: makeGenerateChMask72(false),
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
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:     "72 channels/cntl5/current(1-12,14-33,36-42,44-71)/desired(1-3,5-72)",
			Generate: makeGenerateChMask72(false),
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
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:     "72 channels/no cntl5/current(1-12,14-33,36-42,44-63,65-72)/desired(1-3,5-72)",
			Generate: makeGenerateChMask72(false),
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
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:     "72 channels/cntl5/current(1-12,14-33,36-42,44-63,65-72)/desired(1-3,5-72)",
			Generate: makeGenerateChMask72(false),
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
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:     "72 channels/no cntl5/current(1-72)/desired(9-16,65-72)",
			Generate: makeGenerateChMask72(false),
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
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:     "72 channels/cntl5/current(1-72)/desired(9-16,65-72)",
			Generate: makeGenerateChMask72(true),
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
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:     "72 channels/no cntl5/current(1-72)/desired(9-24)",
			Generate: makeGenerateChMask72(false),
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
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:     "72 channels/cntl5/current(1-72)/desired(9-24)",
			Generate: makeGenerateChMask72(true),
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
				return assertions.New(t).So(err, should.BeNil)
			},
		},

		{
			Name:     "96 channels/current(1-96)/desired(1-96)",
			Generate: generateChMask96,
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
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:     "96 channels/current(1-12,14-33,36-42,44-96)/desired(1-96)",
			Generate: generateChMask96,
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
				return assertions.New(t).So(err, should.BeNil)
			},
		},
		{
			Name:     "96 channels/current(1-12,14-33,36-42,44-95)/desired(1-3,5-96)",
			Generate: generateChMask96,
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
				return assertions.New(t).So(err, should.BeNil)
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
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
