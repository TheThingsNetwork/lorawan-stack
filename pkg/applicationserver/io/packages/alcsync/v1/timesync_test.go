// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

package alcsyncv1

import (
	"context"
	"encoding/binary"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func assertCorrectAns(a *assertions.Assertion, ans []byte, expected AppTimeAns) {
	var actual AppTimeAns
	actual.TimeCorrection = int32(binary.LittleEndian.Uint32(ans))
	actual.TokenAns = ans[4] & 0x0F
	a.So(actual, should.Resemble, expected)
}

func TestTimeSynchronizationCommandCalculatesCorrection(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		Name     string
		Command  TimeSyncCommand
		Expected AppTimeAns
	}{
		{
			Name: "NegativeTimeCorrection",
			Command: TimeSyncCommand{
				req: &AppTimeReq{
					DeviceTime:  receivedAtTime.Add(10 * time.Second),
					TokenReq:    1,
					AnsRequired: true,
				},
				receivedAt: receivedAtTime,
				fPort:      202,
				threshold:  threeSecondsDuration,
			},
			Expected: AppTimeAns{
				TimeCorrection: -10,
				TokenAns:       1,
			},
		},
		{
			Name: "PositiveTimeCorrection",
			Command: TimeSyncCommand{
				req: &AppTimeReq{
					DeviceTime:  receivedAtTime.Add(-10 * time.Second),
					TokenReq:    1,
					AnsRequired: true,
				},
				receivedAt: receivedAtTime,
				fPort:      202,
				threshold:  threeSecondsDuration,
			},
			Expected: AppTimeAns{
				TimeCorrection: 10,
				TokenAns:       1,
			},
		},
		{
			Name: "NoTimeCorrection",
			Command: TimeSyncCommand{
				req: &AppTimeReq{
					DeviceTime:  receivedAtTime,
					TokenReq:    1,
					AnsRequired: true,
				},
				receivedAt: receivedAtTime,
				fPort:      202,
				threshold:  threeSecondsDuration,
			},
			Expected: AppTimeAns{
				TimeCorrection: 0,
				TokenAns:       1,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			a, ctx := test.New(t)
			err := tc.Command.Execute(ctx, func(ctx context.Context, ad *ttnpb.ApplicationDownlink) error {
				assertCorrectAns(a, ad.FrmPayload, tc.Expected)
				return nil
			})
			a.So(err, should.BeNil)
		})
	}
}

func TestTimeSynchronizationCommandRespectsThreshold(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		Name    string
		Command TimeSyncCommand
	}{
		{
			Name: "NegativeTimeCorrection",
			Command: TimeSyncCommand{
				req: &AppTimeReq{
					DeviceTime:  receivedAtTime.Add(2 * time.Second),
					TokenReq:    1,
					AnsRequired: false,
				},
				receivedAt: receivedAtTime,
				fPort:      202,
				threshold:  threeSecondsDuration,
			},
		},
		{
			Name: "PositiveTimeCorrection",
			Command: TimeSyncCommand{
				req: &AppTimeReq{
					DeviceTime:  receivedAtTime.Add(-2 * time.Second),
					TokenReq:    1,
					AnsRequired: false,
				},
				receivedAt: receivedAtTime,
				fPort:      202,
				threshold:  threeSecondsDuration,
			},
		},
		{
			Name: "NoTimeCorrection",
			Command: TimeSyncCommand{
				req: &AppTimeReq{
					DeviceTime:  receivedAtTime,
					TokenReq:    1,
					AnsRequired: false,
				},
				receivedAt: receivedAtTime,
				fPort:      202,
				threshold:  threeSecondsDuration,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			a, ctx := test.New(t)
			err := tc.Command.Execute(ctx, func(ctx context.Context, ad *ttnpb.ApplicationDownlink) error {
				a.So(ad, should.BeNil)
				return nil
			})
			a.So(err, should.BeNil)
		})
	}
}

func TestTimeSynchronizationCommandRespectsAnsRequired(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		Name    string
		Command TimeSyncCommand
	}{
		{
			Name: "NegativeTimeCorrection",
			Command: TimeSyncCommand{
				req: &AppTimeReq{
					DeviceTime:  receivedAtTime.Add(2 * time.Second),
					TokenReq:    1,
					AnsRequired: true,
				},
				receivedAt: receivedAtTime,
				fPort:      202,
				threshold:  threeSecondsDuration,
			},
		},
		{
			Name: "PositiveTimeCorrection",
			Command: TimeSyncCommand{
				req: &AppTimeReq{
					DeviceTime:  receivedAtTime.Add(-2 * time.Second),
					TokenReq:    1,
					AnsRequired: true,
				},
				receivedAt: receivedAtTime,
				fPort:      202,
				threshold:  threeSecondsDuration,
			},
		},
		{
			Name: "NoTimeCorrection",
			Command: TimeSyncCommand{
				req: &AppTimeReq{
					DeviceTime:  receivedAtTime,
					TokenReq:    1,
					AnsRequired: true,
				},
				receivedAt: receivedAtTime,
				fPort:      202,
				threshold:  3,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			a, ctx := test.New(t)
			err := tc.Command.Execute(ctx, func(ctx context.Context, ad *ttnpb.ApplicationDownlink) error {
				a.So(ad, should.NotBeNil)
				return nil
			})
			a.So(err, should.BeNil)
		})
	}
}
