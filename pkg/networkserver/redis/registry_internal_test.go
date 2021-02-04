// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

package redis

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
)

func TestMsgpackCompatibility(t *testing.T) {
	_, ctx := test.New(t)
	cl, flush := test.NewRedis(ctx, "test", "devices")
	t.Cleanup(func() {
		flush()
		if err := cl.Close(); err != nil {
			t.Errorf("Failed to close Redis device registry client: %s", test.FormatError(err))
		}
	})
	for _, v := range []interface{}{
		uplinkMatchSession{},
		uplinkMatchSession{
			FNwkSIntKey: &keyEnvelope{
				Key: &types.AES128Key{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			},
		},
		uplinkMatchSession{
			LoRaWANVersion: ttnpb.MAC_V1_0_3,
		},
		uplinkMatchSession{
			FNwkSIntKey: &keyEnvelope{
				Key: &types.AES128Key{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			},
			LoRaWANVersion: ttnpb.MAC_V1_0_3,
		},
		uplinkMatchSession{
			FNwkSIntKey: &keyEnvelope{
				Key: &types.AES128Key{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			},
			LoRaWANVersion:    ttnpb.MAC_V1_0_3,
			Supports32BitFCnt: &boolValue{Value: true},
		},
		uplinkMatchSession{
			FNwkSIntKey: &keyEnvelope{
				Key: &types.AES128Key{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			},
			LoRaWANVersion: ttnpb.MAC_V1_0_3,
			ResetsFCnt:     &boolValue{Value: true},
		},
		uplinkMatchSession{
			FNwkSIntKey: &keyEnvelope{
				Key: &types.AES128Key{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			},
			LoRaWANVersion: ttnpb.MAC_V1_0_3,
			LastFCnt:       42,
		},
		uplinkMatchSession{
			FNwkSIntKey: &keyEnvelope{
				Key: &types.AES128Key{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			},
			LoRaWANVersion:    ttnpb.MAC_V1_0_3,
			ResetsFCnt:        &boolValue{Value: true},
			Supports32BitFCnt: &boolValue{Value: false},
			LastFCnt:          42,
		},
		uplinkMatchPendingSession{},
		uplinkMatchResult{},
	} {
		v := v
		b := test.Must(marshalMsgpack(v)).([]byte)
		test.RunSubtestFromContext(ctx, test.SubtestConfig{
			Name: fmt.Sprintf("%v %s", v, b),
			Func: func(ctx context.Context, _ *testing.T, a *assertions.Assertion) {
				a.So(redis.NewScript(`cmsgpack.unpack(ARGV[1])`).Run(ctx, cl, nil, b).Err(), should.Resemble, redis.Nil)
			},
		})
	}
}
