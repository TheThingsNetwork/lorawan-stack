// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

package networkserver

import (
	"context"
	"fmt"
	"testing"

	"github.com/smartystreets/assertions"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal/test"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestAppendRecentUplink(t *testing.T) {
	ups := [...]*ttnpb.UplinkMessage{
		{
			RawPayload: []byte("test1"),
		},
		{
			RawPayload: []byte("test2"),
		},
		{
			RawPayload: []byte("test3"),
		},
	}
	for _, tc := range []struct {
		Recent   []*ttnpb.UplinkMessage
		Up       *ttnpb.UplinkMessage
		Window   int
		Expected []*ttnpb.UplinkMessage
	}{
		{
			Up:       ups[0],
			Window:   1,
			Expected: ups[:1],
		},
		{
			Recent:   ups[:1],
			Up:       ups[1],
			Window:   1,
			Expected: ups[1:2],
		},
		{
			Recent:   ups[:2],
			Up:       ups[2],
			Window:   1,
			Expected: ups[2:3],
		},
		{
			Recent:   ups[:1],
			Up:       ups[1],
			Window:   2,
			Expected: ups[:2],
		},
		{
			Recent:   ups[:2],
			Up:       ups[2],
			Window:   2,
			Expected: ups[1:3],
		},
	} {
		tc := tc
		test.RunSubtest(t, test.SubtestConfig{
			Name:     fmt.Sprintf("recent_length:%d,window:%v", len(tc.Recent), tc.Window),
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				recent := CopyUplinkMessages(tc.Recent...)
				up := CopyUplinkMessage(tc.Up)
				ret := appendRecentUplink(recent, up, tc.Window)
				a.So(recent, should.Resemble, tc.Recent)
				a.So(up, should.Resemble, tc.Up)
				a.So(ret, should.Resemble, tc.Expected)
			},
		})
	}
}
