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

	"go.thethings.network/lorawan-stack/v3/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/v3/pkg/component/test"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
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
	a.So(errors.IsNotFound(err), should.BeTrue)

	// Test Upstream.
	upstream := test.Must(NewUpstream(ctx, c, Config{
		NetID:     test.DefaultNetID,
		Source:    "directory",
		Directory: "testdata",
	}))

	unsupportedJoinEUI := types.EUI64{0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0D}
	claimer := upstream.JoinEUIClaimer(ctx, unsupportedJoinEUI)
	a.So(claimer, should.BeNil)

	supportedJoinEUI := types.EUI64{0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0C}
	claimer = upstream.JoinEUIClaimer(ctx, supportedJoinEUI)
	a.So(claimer, should.NotBeNil)
}
