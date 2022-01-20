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

package redis_test

import (
	"testing"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestScheduledDownlinkMatcher(t *testing.T) {
	a, ctx := test.New(t)

	cl, flush := test.NewRedis(ctx, "redis_test")
	defer flush()
	defer cl.Close()

	m := redis.ScheduledDownlinkMatcher{cl}

	ids := &ttnpb.EndDeviceIdentifiers{
		ApplicationIds: &ttnpb.ApplicationIdentifiers{
			ApplicationId: "app1",
		},
		DeviceId: "dev1",
	}

	stored := &ttnpb.DownlinkMessage{
		RawPayload:   []byte{1, 2, 3},
		EndDeviceIds: ids,
		Settings: &ttnpb.DownlinkMessage_Request{
			Request: &ttnpb.TxRequest{
				Class: ttnpb.Class_CLASS_A,
			},
		},
		CorrelationIds: []string{"corr1", "corr2", "ns:downlink:CORRELATIONID"},
	}

	ack := &ttnpb.TxAcknowledgment{
		Result: ttnpb.TxAcknowledgment_SUCCESS,
		DownlinkMessage: &ttnpb.DownlinkMessage{
			Settings: &ttnpb.DownlinkMessage_Scheduled{
				Scheduled: &ttnpb.TxSettings{
					DataRate: &ttnpb.DataRate{
						Modulation: &ttnpb.DataRate_Lora{
							Lora: &ttnpb.LoRaDataRate{
								SpreadingFactor: 7,
								Bandwidth:       125000,
							},
						},
					},
				},
			},
			CorrelationIds: []string{"corr1", "corr2", "ns:downlink:CORRELATIONID"},
		},
	}

	err := m.Add(test.Context(), stored)
	a.So(err, should.BeNil)

	t.Run("MissingCorrelationID", func(t *testing.T) {
		a, ctx := test.New(t)
		down, err := m.Match(ctx, &ttnpb.TxAcknowledgment{})
		a.So(errors.IsNotFound(err), should.BeTrue)
		a.So(down, should.BeNil)
	})

	t.Run("InvalidCorrelationID", func(t *testing.T) {
		a, ctx := test.New(t)
		down, err := m.Match(ctx, &ttnpb.TxAcknowledgment{
			DownlinkMessage: &ttnpb.DownlinkMessage{
				CorrelationIds: []string{"ns:downlink:OTHERCORRELATIONID"},
			},
		})
		a.So(errors.IsNotFound(err), should.BeTrue)
		a.So(down, should.BeNil)
	})

	t.Run("Match", func(t *testing.T) {
		a, ctx := test.New(t)
		down, err := m.Match(ctx, ack)
		a.So(err, should.BeNil)
		a.So(down, should.Resemble, stored)
	})

	t.Run("DoNotMatchTwice", func(t *testing.T) {
		a, ctx := test.New(t)
		down, err := m.Match(ctx, ack)
		a.So(errors.IsNotFound(err), should.BeTrue)
		a.So(down, should.BeNil)
	})
}
