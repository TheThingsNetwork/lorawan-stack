// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

func TestChannelMask(t *testing.T) {
	for _, tc := range []struct {
		name string

		chMaskFunc ChannelMaskFunc

		mask [16]bool
		cntl uint8

		enabledChannels  []int
		disabledChannels []int
		fails            bool
	}{
		{
			name: "16 channels/cntl=0",

			chMaskFunc: chMask16Channels,

			mask: [16]bool{
				true, false, false, true, false, false, false, false,
				false, false, false, false, false, false, false, false,
			},
			cntl: 0,

			enabledChannels:  []int{0, 3},
			disabledChannels: []int{1, 2, 4, 5, 11},
		},
		{
			name: "16 channels/cntl=6",

			chMaskFunc: chMask16Channels,

			mask: [16]bool{
				true, false, false, true, false, false, false, false,
				false, false, false, false, false, false, false, false,
			},
			cntl: 6,

			enabledChannels: func() (chans []int) {
				for i := 0; i < 16; i++ {
					chans = append(chans, i)
				}
				return
			}(),
		},
		{
			name:       "16 channels/cntl=3",
			chMaskFunc: chMask16Channels,
			cntl:       3,
			fails:      true,
		},
		{
			name: "72 channels/cntl=1",

			chMaskFunc: chMask72Channels,

			mask: [16]bool{
				true, false, false, true, false, false, false, false,
				false, false, false, false, false, false, false, false,
			},
			cntl: 1,

			enabledChannels:  []int{16, 19},
			disabledChannels: []int{0, 3, 4, 5, 17, 18, 20, 32, 64},
		},
		{
			name: "72 channels/cntl=5",

			chMaskFunc: chMask72Channels,

			mask: [16]bool{
				true, false, false, true, false, false, false, false,
				false, false, false, false, false, false, false, false,
			},
			cntl: 5,

			enabledChannels:  []int{0, 3, 7, 24, 25, 26, 30, 31, 64, 67},
			disabledChannels: []int{8, 9, 10, 11, 32, 33, 55, 65, 66, 68, 70},
		},
		{
			name: "72 channels/cntl=6",

			chMaskFunc: chMask72Channels,

			mask: [16]bool{
				true, false, false, true, false, false, false, false,
				false, false, false, false, false, false, false, false,
			},
			cntl: 6,

			enabledChannels:  []int{0, 3, 7, 8, 9, 10, 11, 24, 25, 26, 30, 32, 33, 55, 31, 64, 67},
			disabledChannels: []int{65, 66, 68},
		},
		{
			name: "72 channels/cntl=7",

			chMaskFunc: chMask72Channels,

			mask: [16]bool{
				true, false, false, true, true, false, false, false,
				false, false, false, false, false, false, false, false,
			},
			cntl: 7,

			enabledChannels:  []int{64, 67, 68},
			disabledChannels: []int{0, 3, 7, 8, 9, 10, 11, 24, 25, 26, 30, 32, 33, 55, 31, 65, 66, 69, 70},
		},
		{
			name:       "72 channels/cntl=math.MaxUint8",
			chMaskFunc: chMask72Channels,
			cntl:       math.MaxUint8,
			fails:      true,
		},
		{
			name: "96 channels/cntl=3",

			chMaskFunc: chMask96Channels,

			mask: [16]bool{
				true, false, false, true, true, false, false, false,
				true, true, true, false, false, false, false, false,
			},
			cntl: 3,

			enabledChannels:  []int{48, 51, 52, 56, 57},
			disabledChannels: []int{0, 16, 17, 49, 50, 55, 66},
		},
		{
			name:            "96 channels/cntl=6",
			chMaskFunc:      chMask96Channels,
			cntl:            6,
			enabledChannels: []int{0, 3, 16, 17, 55, 90},
		},
		{
			name:       "96 channels/cntl=math.MaxUint8",
			chMaskFunc: chMask96Channels,
			cntl:       math.MaxUint8,
			fails:      true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a := assertions.New(t)

			res, err := tc.chMaskFunc(tc.mask, tc.cntl)
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
