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

const gatewayID = "test"

func TestDefaultTopics(t *testing.T) {
	ctx := test.Context()
	v3 := topics.New(ctx)
	uid := unique.ID(ctx, &ttnpb.GatewayIdentifiers{GatewayId: gatewayID})
	for _, tc := range []struct {
		UID      string
		Func     func(string) []string
		Expected []string
		Is       func([]string) bool
		IsNot    []func([]string) bool
	}{
		{
			UID:      uid,
			Func:     v3.UplinkTopic,
			Expected: []string{"v3", uid, "up"},
			Is:       v3.IsUplinkTopic,
			IsNot:    []func([]string) bool{v3.IsStatusTopic, v3.IsTxAckTopic},
		},
		{
			UID:      uid,
			Func:     v3.StatusTopic,
			Expected: []string{"v3", uid, "status"},
			Is:       v3.IsStatusTopic,
			IsNot:    []func([]string) bool{v3.IsUplinkTopic, v3.IsTxAckTopic},
		},
		{
			UID:      uid,
			Func:     v3.TxAckTopic,
			Expected: []string{"v3", uid, "down", "ack"},
			Is:       v3.IsTxAckTopic,
			IsNot:    []func([]string) bool{v3.IsUplinkTopic, v3.IsStatusTopic},
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
