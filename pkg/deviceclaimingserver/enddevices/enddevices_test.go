// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

package enddevices

import (
	"testing"

	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/v3/pkg/component/test"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

var (
	supportedJoinEUI   = &types.EUI64{0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0C}
	unsupportedJoinEUI = &types.EUI64{0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0D}
)

func TestUpstream(t *testing.T) {
	t.Parallel()
	a, ctx := test.New(t)

	c := componenttest.NewComponent(t, &component.Config{})
	componenttest.StartComponent(t, c)
	t.Cleanup(func() {
		c.Close()
	})

	// Invalid configs
	_, err := NewUpstream(ctx, c, Config{
		Source: "directory",
	})
	a.So(err, should.NotBeNil)

	// Upstream test
	upstream := test.Must(NewUpstream(ctx, c, Config{
		NetID:     test.DefaultNetID,
		Source:    "directory",
		Directory: "testdata",
	}, WithDeviceRegistry(&mockDeviceRegistry{})))

	ctx = rights.NewContext(ctx, &rights.Rights{
		ApplicationRights: *rights.NewMap(map[string]*ttnpb.Rights{
			unique.ID(test.Context(), &ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"}): ttnpb.RightsFrom(
				ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ,
				ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE,
			),
		}),
	})

	// Invalid JoinEUI.
	err = upstream.Claim(ctx, *unsupportedJoinEUI,
		types.EUI64{0x00, 0x04, 0xA3, 0x0B, 0x00, 0x1C, 0x05, 0x30},
		"secret",
	)
	a.So(errors.IsAborted(err), should.BeTrue)

	_, err = upstream.Unclaim(ctx, &ttnpb.EndDeviceIdentifiers{
		DeviceId: "test-dev",
		ApplicationIds: &ttnpb.ApplicationIdentifiers{
			ApplicationId: "test-app",
		},
		JoinEui: unsupportedJoinEUI.Bytes(),
		DevEui:  types.EUI64{0x00, 0x04, 0xA3, 0x0B, 0x00, 0x1C, 0x05, 0x30}.Bytes(),
	})
	a.So(errors.IsUnauthenticated(err), should.BeTrue)

	resp, err := upstream.GetInfoByJoinEUI(ctx, &ttnpb.GetInfoByJoinEUIRequest{
		JoinEui: unsupportedJoinEUI.Bytes(),
	})
	a.So(err, should.BeNil)
	a.So(resp.SupportsClaiming, should.BeFalse)

	// Valid JoinEUI.
	inf, err := upstream.GetInfoByJoinEUI(ctx, &ttnpb.GetInfoByJoinEUIRequest{
		JoinEui: supportedJoinEUI.Bytes(),
	})
	a.So(err, should.BeNil)
	a.So(inf.JoinEui, should.Resemble, supportedJoinEUI.Bytes())
	a.So(inf.SupportsClaiming, should.BeTrue)

	err = upstream.Claim(ctx, *supportedJoinEUI,
		types.EUI64{0x00, 0x04, 0xA3, 0x0B, 0x00, 0x1C, 0x05, 0x30},
		"secret",
	)
	a.So(!errors.IsUnimplemented(err), should.BeTrue)

	_, err = upstream.Unclaim(ctx, &ttnpb.EndDeviceIdentifiers{
		DeviceId: "test-dev",
		ApplicationIds: &ttnpb.ApplicationIdentifiers{
			ApplicationId: "test-app",
		},
		JoinEui: supportedJoinEUI.Bytes(),
		DevEui:  types.EUI64{0x00, 0x04, 0xA3, 0x0B, 0x00, 0x1C, 0x05, 0x30}.Bytes(),
	})
	a.So(!errors.IsUnavailable(err), should.BeTrue)
}
