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

func TestDefaultTopics(t *testing.T) {
	for _, tc := range []struct {
		UID      string
		Func     func(string) []string
		Expected []string
		Is       func([]string) bool
		IsNot    []func([]string) bool
	}{
		{
			UID:      "test",
			Func:     topics.Default.UplinkTopic,
			Expected: []string{"v3", "test", "up"},
			Is:       topics.Default.IsUplinkTopic,
			IsNot:    []func([]string) bool{topics.Default.IsStatusTopic, topics.Default.IsTxAckTopic},
		},
		{
			UID:      "test",
			Func:     topics.Default.StatusTopic,
			Expected: []string{"v3", "test", "status"},
			Is:       topics.Default.IsStatusTopic,
			IsNot:    []func([]string) bool{topics.Default.IsUplinkTopic, topics.Default.IsTxAckTopic},
		},
		{
			UID:      "test",
			Func:     topics.Default.TxAckTopic,
			Expected: []string{"v3", "test", "down", "ack"},
			Is:       topics.Default.IsTxAckTopic,
			IsNot:    []func([]string) bool{topics.Default.IsUplinkTopic, topics.Default.IsStatusTopic},
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
