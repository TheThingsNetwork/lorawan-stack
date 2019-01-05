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
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/io/mqtt/topics"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestV2Topics(t *testing.T) {
	for _, tc := range []struct {
		UID      string
		Func     func(string) []string
		Expected []string
		Is       func([]string) bool
		IsNot    []func([]string) bool
	}{
		{
			UID:      "test",
			Func:     topics.V2.UplinkTopic,
			Expected: []string{"test", "up"},
			Is:       topics.V2.IsUplinkTopic,
			IsNot:    []func([]string) bool{topics.V2.IsStatusTopic, topics.V2.IsTxAckTopic},
		},
		{
			UID:      "test",
			Func:     topics.V2.StatusTopic,
			Expected: []string{"test", "status"},
			Is:       topics.V2.IsStatusTopic,
			IsNot:    []func([]string) bool{topics.V2.IsUplinkTopic, topics.V2.IsTxAckTopic},
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
