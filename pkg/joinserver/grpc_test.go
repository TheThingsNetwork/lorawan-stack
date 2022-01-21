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

package joinserver_test

import (
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/v3/pkg/component/test"
	. "go.thethings.network/lorawan-stack/v3/pkg/joinserver"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestGetJoinEUIPrefixes(t *testing.T) {
	for _, tc := range []struct {
		Name            string
		JoinEUIPrefixes []types.EUI64Prefix
		Response        *ttnpb.JoinEUIPrefixes
	}{
		{
			Name: "Defined JoinEUIPrefixes Set 1",
			JoinEUIPrefixes: []types.EUI64Prefix{
				{EUI64: types.EUI64{0xff, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, Length: 12},
				{EUI64: types.EUI64{0x10, 0x00, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, Length: 12},
				{EUI64: types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x00}, Length: 56},
			},
			Response: &ttnpb.JoinEUIPrefixes{
				Prefixes: []*ttnpb.JoinEUIPrefix{
					{JoinEui: types.EUI64{0xff, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, Length: 12},
					{JoinEui: types.EUI64{0x10, 0x00, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, Length: 12},
					{JoinEui: types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x00}, Length: 56},
				},
			},
		},
		{
			Name: "Defined JoinEUIPrefixes Set 2",
			JoinEUIPrefixes: []types.EUI64Prefix{
				{EUI64: types.EUI64{0xaf, 0xb2, 0x11, 0x00, 0x4f, 0x99, 0x75, 0x01}, Length: 1},
				{EUI64: types.EUI64{0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, Length: 40},
				{EUI64: types.EUI64{0x11, 0xff, 0x11, 0xff, 0x11, 0xff, 0x11, 0x00}, Length: 56},
			},
			Response: &ttnpb.JoinEUIPrefixes{
				Prefixes: []*ttnpb.JoinEUIPrefix{
					{JoinEui: types.EUI64{0xaf, 0xb2, 0x11, 0x00, 0x4f, 0x99, 0x75, 0x01}, Length: 1},
					{JoinEui: types.EUI64{0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, Length: 40},
					{JoinEui: types.EUI64{0x11, 0xff, 0x11, 0xff, 0x11, 0xff, 0x11, 0x00}, Length: 56},
				},
			},
		},
		{
			Name: "Defined JoinEUIPrefixes Set 3",
			JoinEUIPrefixes: []types.EUI64Prefix{
				{EUI64: types.EUI64{0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56}, Length: 4},
				{EUI64: types.EUI64{0x00, 0x00, 0x00, 0x00, 0xff, 0xff, 0xff, 0x45}, Length: 8},
				{EUI64: types.EUI64{0x45, 0xff, 0xff, 0xff, 0x00, 0x00, 0x00, 0x00}, Length: 16},
			},
			Response: &ttnpb.JoinEUIPrefixes{
				Prefixes: []*ttnpb.JoinEUIPrefix{
					{JoinEui: types.EUI64{0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56}, Length: 4},
					{JoinEui: types.EUI64{0x00, 0x00, 0x00, 0x00, 0xff, 0xff, 0xff, 0x45}, Length: 8},
					{JoinEui: types.EUI64{0x45, 0xff, 0xff, 0xff, 0x00, 0x00, 0x00, 0x00}, Length: 16},
				},
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			js := test.Must(New(
				componenttest.NewComponent(t, &component.Config{}),
				&Config{
					JoinEUIPrefixes: tc.JoinEUIPrefixes,
					DevNonceLimit:   defaultDevNonceLimit,
				})).(*JoinServer)
			componenttest.StartComponent(t, js.Component)
			defer js.Close()

			resp, err := ttnpb.NewJsClient(js.LoopbackConn()).GetJoinEUIPrefixes(test.Context(), ttnpb.Empty)
			if a.So(err, should.BeNil) {
				a.So(resp, should.Resemble, tc.Response)
			}
		})
	}
}

func TestGetDefaultJoinEUI(t *testing.T) {
	for _, tc := range []struct {
		Name           string
		DefaultJoinEUI types.EUI64
		Response       ttnpb.GetDefaultJoinEUIResponse
	}{
		{
			Name:           "Default",
			DefaultJoinEUI: types.EUI64{0xff, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			Response: ttnpb.GetDefaultJoinEUIResponse{
				JoinEui: &types.EUI64{0xff, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			},
		},
		{
			Name:           "AnotherDefault",
			DefaultJoinEUI: types.EUI64{0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			Response: ttnpb.GetDefaultJoinEUIResponse{
				JoinEui: &types.EUI64{0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			js := test.Must(New(
				componenttest.NewComponent(t, &component.Config{}),
				&Config{
					DefaultJoinEUI: tc.DefaultJoinEUI,
					DevNonceLimit:  defaultDevNonceLimit,
				})).(*JoinServer)
			componenttest.StartComponent(t, js.Component)
			defer js.Close()

			resp, err := ttnpb.NewJsClient(js.LoopbackConn()).GetDefaultJoinEUI(test.Context(), ttnpb.Empty)
			if a.So(err, should.BeNil) {
				a.So(resp, should.Resemble, tc.Response)
			}
		})
	}
}
