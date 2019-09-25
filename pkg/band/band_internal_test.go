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
	"math"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestParseChMask(t *testing.T) {
	for _, tc := range []struct {
		name string

		parseChMask func(mask [16]bool, cntl uint8) (map[uint8]bool, error)

		mask [16]bool
		cntl uint8

		enabledChannels  []uint8
		disabledChannels []uint8
		fails            bool
	}{
		{
			name: "16 channels/cntl=0",

			parseChMask: parseChMask16,

			mask: [16]bool{
				true, false, false, true, false, false, false, false,
				false, false, false, false, false, false, false, false,
			},
			cntl: 0,

			enabledChannels:  []uint8{0, 3},
			disabledChannels: []uint8{1, 2, 4, 5, 11},
		},
		{
			name: "16 channels/cntl=6",

			parseChMask: parseChMask16,

			mask: [16]bool{
				true, false, false, true, false, false, false, false,
				false, false, false, false, false, false, false, false,
			},
			cntl: 6,

			enabledChannels: func() (chans []uint8) {
				for i := uint8(0); i < 16; i++ {
					chans = append(chans, i)
				}
				return
			}(),
		},
		{
			name:        "16 channels/cntl=3",
			parseChMask: parseChMask16,
			cntl:        3,
			fails:       true,
		},
		{
			name: "72 channels/cntl=1",

			parseChMask: parseChMask72,

			mask: [16]bool{
				true, false, false, true, false, false, false, false,
				false, false, false, false, false, false, false, false,
			},
			cntl: 1,

			enabledChannels:  []uint8{16, 19},
			disabledChannels: []uint8{0, 3, 4, 5, 17, 18, 20, 32, 64},
		},
		{
			name: "72 channels/cntl=5",

			parseChMask: parseChMask72,

			mask: [16]bool{
				true, false, false, true, false, false, false, false,
				false, false, false, false, false, false, false, false,
			},
			cntl: 5,

			enabledChannels:  []uint8{0, 3, 7, 24, 25, 26, 30, 31, 64, 67},
			disabledChannels: []uint8{8, 9, 10, 11, 32, 33, 55, 65, 66, 68, 70},
		},
		{
			name: "72 channels/cntl=6",

			parseChMask: parseChMask72,

			mask: [16]bool{
				true, false, false, true, false, false, false, false,
				false, false, false, false, false, false, false, false,
			},
			cntl: 6,

			enabledChannels:  []uint8{0, 3, 7, 8, 9, 10, 11, 24, 25, 26, 30, 32, 33, 55, 31, 64, 67},
			disabledChannels: []uint8{65, 66, 68},
		},
		{
			name: "72 channels/cntl=7",

			parseChMask: parseChMask72,

			mask: [16]bool{
				true, false, false, true, true, false, false, false,
				false, false, false, false, false, false, false, false,
			},
			cntl: 7,

			enabledChannels:  []uint8{64, 67, 68},
			disabledChannels: []uint8{0, 3, 7, 8, 9, 10, 11, 24, 25, 26, 30, 32, 33, 55, 31, 65, 66, 69, 70},
		},
		{
			name:        "72 channels/cntl=math.MaxUint8",
			parseChMask: parseChMask72,
			cntl:        math.MaxUint8,
			fails:       true,
		},
		{
			name: "96 channels/cntl=3",

			parseChMask: parseChMask96,

			mask: [16]bool{
				true, false, false, true, true, false, false, false,
				true, true, true, false, false, false, false, false,
			},
			cntl: 3,

			enabledChannels:  []uint8{48, 51, 52, 56, 57},
			disabledChannels: []uint8{0, 16, 17, 49, 50, 55, 66},
		},
		{
			name:            "96 channels/cntl=6",
			parseChMask:     parseChMask96,
			cntl:            6,
			enabledChannels: []uint8{0, 3, 16, 17, 55, 90},
		},
		{
			name:        "96 channels/cntl=math.MaxUint8",
			parseChMask: parseChMask96,
			cntl:        math.MaxUint8,
			fails:       true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a := assertions.New(t)

			res, err := tc.parseChMask(tc.mask, tc.cntl)
			if tc.fails {
				if !a.So(err, should.NotBeNil) {
					t.FailNow()
				}
				return
			}
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}

			for _, enabledChannel := range tc.enabledChannels {
				a.So(res[enabledChannel], should.BeTrue)
			}
			for _, disabledChannel := range tc.disabledChannels {
				a.So(res[disabledChannel], should.BeFalse)
			}
		})
	}
}

func TestGenerateChMask(t *testing.T) {
	for _, tc := range []struct {
		Name     string
		Generate func([]bool) ([]ChMaskCntlPair, error)
		Mask     []bool
		Expected []ChMaskCntlPair
		Error    error
	}{
		{
			Name:     "16 channels/2,4 on",
			Generate: generateChMask16,
			Mask: []bool{
				false, true, false, true, false, false, false, false, false, false, false, false, false, false, false, false,
			},
			Expected: []ChMaskCntlPair{
				{0, [16]bool{false, true, false, true, false, false, false, false, false, false, false, false, false, false, false, false}},
			},
		},
		{
			Name:     "16 channels/all on",
			Generate: generateChMask16,
			Mask: []bool{
				true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true,
			},
			Expected: []ChMaskCntlPair{
				{6, [16]bool{}},
			},
		},
		{
			Name:     "72 channels/1-16 on, 42, 67, 69 on (Cntl5 off)",
			Generate: makeGenerateChMask72(false),
			Mask: []bool{
				true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true,
				false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false, false, true, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false,
				false, false, true, false, true, false, false, false,
			},
			Expected: []ChMaskCntlPair{
				{0, [16]bool{true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true}},
				{1, [16]bool{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false}},
				{2, [16]bool{false, false, false, false, false, false, false, false, false, true, false, false, false, false, false, false}},
				{3, [16]bool{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false}},
				{4, [16]bool{false, false, true, false, true, false, false, false, false, false, false, false, false, false, false, false}},
			},
		},
		{
			Name:     "72 channels/1-16 on, 42, 67, 69 on (Cntl5 on)",
			Generate: makeGenerateChMask72(true),
			Mask: []bool{
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
				{5, [16]bool{false, false, false, false, false, false, false, false, false, false, false, false, false, false, true, true}},
				{2, [16]bool{false, false, false, false, false, false, false, false, false, true, false, false, false, false, false, false}},
				{4, [16]bool{false, false, true, false, true, false, false, false, false, false, false, false, false, false, false, false}},
			},
		},
		{
			Name:     "72 channels/125Hz on, 66, 68 on",
			Generate: makeGenerateChMask72(false),
			Mask: []bool{
				true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true,
				false, true, false, true, false, false, false, false,
			},
			Expected: []ChMaskCntlPair{
				{6, [16]bool{false, true, false, true, false, false, false, false, false, false, false, false, false, false, false, false}},
			},
		},
		{
			Name:     "72 channels/125Hz off, 67, 69 on",
			Generate: makeGenerateChMask72(false),
			Mask: []bool{
				false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false,
				false, false, true, false, true, false, false, false,
			},
			Expected: []ChMaskCntlPair{
				{7, [16]bool{false, false, true, false, true, false, false, false, false, false, false, false, false, false, false, false}},
			},
		},
		{
			Name:     "72 channels/FSB 1 on",
			Generate: makeGenerateChMask72(true),
			Mask: []bool{
				true, true, true, true, true, true, true, true,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				true, false, false, false, false, false, false, false,
			},
			Expected: []ChMaskCntlPair{
				{5, [16]bool{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, true}},
			},
		},
		{
			Name:     "72 channels/FSB 1 on, ch 2 off",
			Generate: makeGenerateChMask72(true),
			Mask: []bool{
				true, false, true, true, true, true, true, true,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				true, false, false, false, false, false, false, false,
			},
			Expected: []ChMaskCntlPair{
				{5, [16]bool{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, true}},
				{0, [16]bool{true, false, true, true, true, true, true, true, false, false, false, false, false, false, false, false}},
			},
		},
		{
			Name:     "72 channels/FSB 3, 4 on",
			Generate: makeGenerateChMask72(true),
			Mask: []bool{
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, true, true, false, false, false, false,
			},
			Expected: []ChMaskCntlPair{
				{5, [16]bool{false, false, false, false, false, false, false, false, false, false, false, false, true, true, false, false}},
			},
		},
		{
			Name:     "72 channels/FSB 3, 4 on, ch 67,68 off",
			Generate: makeGenerateChMask72(true),
			Mask: []bool{
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false,
			},
			Expected: []ChMaskCntlPair{
				{5, [16]bool{false, false, false, false, false, false, false, false, false, false, false, false, true, true, false, false}},
				{4, [16]bool{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false}},
			},
		},
		{
			Name:     "96 channels/1-16 on, 42, 67, 69, 80 on",
			Generate: generateChMask96,
			Mask: []bool{
				true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true,
				false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false, false, true, false, false, false, false, false, false,
				false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false,
				false, false, true, false, true, false, false, false, false, false, false, false, false, false, false, false,
				true, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false,
			},
			Expected: []ChMaskCntlPair{
				{0, [16]bool{true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true}},
				{1, [16]bool{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false}},
				{2, [16]bool{false, false, false, false, false, false, false, false, false, true, false, false, false, false, false, false}},
				{3, [16]bool{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false}},
				{4, [16]bool{false, false, true, false, true, false, false, false, false, false, false, false, false, false, false, false}},
				{5, [16]bool{true, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false}},
			},
		},
		{
			Name:     "96 channels/all on",
			Generate: generateChMask96,
			Mask: []bool{
				true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true,
				true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true,
			},
			Expected: []ChMaskCntlPair{
				{6, [16]bool{}},
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			ret, err := tc.Generate(tc.Mask)
			if tc.Error != nil && !a.So(err, should.EqualErrorOrDefinition, tc.Error) ||
				tc.Error == nil && !a.So(err, should.BeNil) {
				t.FailNow()
			}
			a.So(ret, should.Resemble, tc.Expected)
		})
	}
}
