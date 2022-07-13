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

package topics_test

import (
	"testing"

	"github.com/TheThingsIndustries/mystique/pkg/topic"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io/mqtt/topics"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

const gatewayIDV2 = "test"

func TestV2Topics(t *testing.T) {
	ctx := test.Context()
	v2 := topics.NewV2(ctx)
	uid := unique.ID(ctx, &ttnpb.GatewayIdentifiers{GatewayId: gatewayIDV2})
	for _, tc := range []struct {
		UID      string
		Func     func(string) []string
		Expected []string
		Is       func([]string) bool
		IsNot    []func([]string) bool
	}{
		{
			UID:      uid,
			Func:     v2.UplinkTopic,
			Expected: []string{uid, "up"},
			Is:       v2.IsUplinkTopic,
			IsNot:    []func([]string) bool{v2.IsStatusTopic, v2.IsTxAckTopic},
		},
		{
			UID:      uid,
			Func:     v2.StatusTopic,
			Expected: []string{uid, "status"},
			Is:       v2.IsStatusTopic,
			IsNot:    []func([]string) bool{v2.IsUplinkTopic, v2.IsTxAckTopic},
		},
	} {
		t.Run(topic.Join(tc.Expected), func(t *testing.T) {
			a := assertions.New(t)
			actual := tc.Func(tc.UID)
			a.So(actual, should.Resemble, tc.Expected)
			a.So(tc.Is(actual), should.BeTrue)
			for _, isNot := range tc.IsNot {
				a.So(isNot(actual), should.BeFalse)
			}
		})
	}
}
