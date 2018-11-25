// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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
	"strconv"
	"testing"

	"github.com/TheThingsIndustries/mystique/pkg/topic"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/applicationserver/io/mqtt/topics"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestV3AcceptedTopic(t *testing.T) {
	uid := "foo-app"
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
			Accepted:  "v3/foo-app/#",
			OK:        true,
		},
		{
			Requested: "v3/#",
			Accepted:  "v3/foo-app/#",
			OK:        true,
		},
		{
			Requested: "v3/+/uplink",
			Accepted:  "v3/foo-app/uplink",
			OK:        true,
		},
		{
			Requested: "v3/foo-app/uplink",
			Accepted:  "v3/foo-app/uplink",
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
