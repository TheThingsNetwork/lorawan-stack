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
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/TheThingsIndustries/mystique/pkg/topic"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/applicationserver/io/mqtt/topics"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestV3AcceptedTopic(t *testing.T) {
	uid := unique.ID(test.Context(), ttnpb.ApplicationIdentifiers{ApplicationID: "foo-app"})
	for i, tc := range []struct {
		Requested,
		Accepted string
		OK bool
	}{
		{
			Requested: "v3",
		},
		{
			Requested: "+",
		},
		{
			Requested: "#",
			Accepted:  fmt.Sprintf("v3/%s/#", uid),
			OK:        true,
		},
		{
			Requested: "v3/#",
			Accepted:  fmt.Sprintf("v3/%s/#", uid),
			OK:        true,
		},
		{
			Requested: "v3/+/uplink",
			Accepted:  fmt.Sprintf("v3/%s/uplink", uid),
			OK:        true,
		},
		{
			Requested: fmt.Sprintf("v3/%s/uplink", uid),
			Accepted:  fmt.Sprintf("v3/%s/uplink", uid),
			OK:        true,
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			a := assertions.New(t)
			actual, ok := topics.Default.AcceptedTopic(uid, topic.Split(tc.Requested))
			if !a.So(ok, should.Equal, tc.OK) {
				t.FailNow()
			}
			a.So(topic.Join(actual), should.Equal, tc.Accepted)
		})
	}
}

func TestV3Topics(t *testing.T) {
	appUID := unique.ID(test.Context(), ttnpb.ApplicationIdentifiers{ApplicationID: "foo-app"})
	devID := "foo-device"

	for _, tc := range []struct {
		Fn       func(applicationUID, deviceUID string) []string
		Expected string
	}{
		{
			Fn:       topics.Default.UplinkTopic,
			Expected: fmt.Sprintf("v3/%s/devices/%s/up", appUID, devID),
		},
		{
			Fn:       topics.Default.JoinAcceptTopic,
			Expected: fmt.Sprintf("v3/%s/devices/%s/join", appUID, devID),
		},
		{
			Fn:       topics.Default.DownlinkAckTopic,
			Expected: fmt.Sprintf("v3/%s/devices/%s/down/ack", appUID, devID),
		},
		{
			Fn:       topics.Default.DownlinkNackTopic,
			Expected: fmt.Sprintf("v3/%s/devices/%s/down/nack", appUID, devID),
		},
		{
			Fn:       topics.Default.DownlinkSentTopic,
			Expected: fmt.Sprintf("v3/%s/devices/%s/down/sent", appUID, devID),
		},
		{
			Fn:       topics.Default.DownlinkFailedTopic,
			Expected: fmt.Sprintf("v3/%s/devices/%s/down/failed", appUID, devID),
		},
		{
			Fn:       topics.Default.DownlinkQueuedTopic,
			Expected: fmt.Sprintf("v3/%s/devices/%s/down/queued", appUID, devID),
		},
		{
			Fn:       topics.Default.LocationSolvedTopic,
			Expected: fmt.Sprintf("v3/%s/devices/%s/location/solved", appUID, devID),
		},
		{
			Fn:       topics.Default.DownlinkPushTopic,
			Expected: fmt.Sprintf("v3/%s/devices/%s/down/push", appUID, devID),
		},
		{
			Fn:       topics.Default.DownlinkReplaceTopic,
			Expected: fmt.Sprintf("v3/%s/devices/%s/down/replace", appUID, devID),
		},
	} {
		t.Run(tc.Expected, func(t *testing.T) {
			actual := strings.Join(tc.Fn(appUID, devID), "/")
			assertions.New(t).So(actual, should.Equal, tc.Expected)
		})
	}
}
