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

package gatewayserver_test

import (
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestConfig(t *testing.T) {
	a := assertions.New(t)

	{
		conf := gatewayserver.Config{
			Forward: map[string][]string{
				"":                []string{"00000000/0"},
				"packetbroker.io": []string{"00000000/3", "26000000/7"},
			},
		}
		forward, err := conf.ForwardDevAddrPrefixes()
		a.So(err, should.BeNil)
		a.So(forward, should.HaveEmptyDiff, map[string][]types.DevAddrPrefix{
			"": []types.DevAddrPrefix{{}},
			"packetbroker.io": []types.DevAddrPrefix{
				{DevAddr: types.DevAddr{}, Length: 3},
				{DevAddr: types.DevAddr{0x26, 0x0, 0x0, 0x0}, Length: 7},
			},
		})
	}

	{
		conf := gatewayserver.Config{
			Forward: map[string][]string{
				"packetbroker.io": []string{"00000000/3", "invalid"},
			},
		}
		_, err := conf.ForwardDevAddrPrefixes()
		a.So(err, should.NotBeNil)
	}
}
