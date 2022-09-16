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
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

var (
	ParseChMask   = parseChMask
	ParseChMask16 = parseChMask16
	ParseChMask72 = parseChMask72
	ParseChMask96 = parseChMask96

	ErrUnsupportedChMaskCntl = errUnsupportedChMaskCntl
)

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
				t.Helper()
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
				t.Helper()
				return assertions.New(t).So(err, should.BeNil)
			},
		},

		{
			Name:     "72 channels/no cntl5/current(1-72)/desired(1-72)",
			Generate: makeGenerateChMask72(false, true),
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
			Generate: makeGenerateChMask72(true, true),
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
			Generate: makeGenerateChMask72(false, true),
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
			Generate: makeGenerateChMask72(true, true),
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
			Generate: makeGenerateChMask72(false, true),
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
			Generate: makeGenerateChMask72(false, true),
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
			Generate: makeGenerateChMask72(false, true),
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
			Generate: makeGenerateChMask72(false, true),
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
			Generate: makeGenerateChMask72(false, true),
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
			Generate: makeGenerateChMask72(false, true),
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
			Generate: makeGenerateChMask72(false, true),
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
			Generate: makeGenerateChMask72(false, true),
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
			Generate: makeGenerateChMask72(false, true),
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
			Generate: makeGenerateChMask72(true, true),
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
			Generate: makeGenerateChMask72(false, true),
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
			Generate: makeGenerateChMask72(false, false),
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
			Name:     "72 channels/cntl5/current(1-72)/desired(9-24)",
			Generate: makeGenerateChMask72(true, true),
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
				t.Helper()
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
				t.Helper()
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

			res := compareDatarates(tc.A, tc.B, tc.Strict)

			if !a.So(res, should.Equal, tc.Expected) {
				t.Fatalf("Unexpected outcome received. Expected :%v, got: %v", tc.Expected, res)
			}
		})
	}
}
