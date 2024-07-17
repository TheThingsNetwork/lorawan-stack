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
	"testing"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestTimeSynchronizationCommandCalculatesCorrection(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		Name     string
		Command  Command
		Expected Result
	}{
		{
			Name: "NegativeTimeCorrection",
			Command: &TimeSyncCommand{
				req: &ttnpb.ALCSyncCommand_AppTimeReq{
					DeviceTime:  timestamppb.New(receivedAtTime.Add(10 * time.Second)),
					TokenReq:    1,
					AnsRequired: true,
				},
				receivedAt: receivedAtTime,
				fPort:      202,
				threshold:  threeSecondsDuration,
			},
			Expected: &TimeSyncCommandResult{
				ans: &ttnpb.ALCSyncCommand_AppTimeAns{
					TimeCorrection: -10,
					TokenAns:       1,
				},
			},
		},
		{
			Name: "PositiveTimeCorrection",
			Command: &TimeSyncCommand{
				req: &ttnpb.ALCSyncCommand_AppTimeReq{
					DeviceTime:  timestamppb.New(receivedAtTime.Add(-10 * time.Second)),
					TokenReq:    1,
					AnsRequired: true,
				},
				receivedAt: receivedAtTime,
				fPort:      202,
				threshold:  threeSecondsDuration,
			},
			Expected: &TimeSyncCommandResult{
				ans: &ttnpb.ALCSyncCommand_AppTimeAns{
					TimeCorrection: 10,
					TokenAns:       1,
				},
			},
		},
		{
			Name: "NoTimeCorrectionWithAnswer",
			Command: &TimeSyncCommand{
				req: &ttnpb.ALCSyncCommand_AppTimeReq{
					DeviceTime:  timestamppb.New(receivedAtTime),
					TokenReq:    1,
					AnsRequired: true,
				},
				receivedAt: receivedAtTime,
				fPort:      202,
				threshold:  threeSecondsDuration,
			},
			Expected: &TimeSyncCommandResult{
				ans: &ttnpb.ALCSyncCommand_AppTimeAns{
					TimeCorrection: 0,
					TokenAns:       1,
				},
			},
		},
		{
			Name: "NoTimeCorrectionWithNoAnswer",
			Command: &TimeSyncCommand{
				req: &ttnpb.ALCSyncCommand_AppTimeReq{
					DeviceTime:  timestamppb.New(receivedAtTime),
					TokenReq:    1,
					AnsRequired: false,
				},
				receivedAt: receivedAtTime,
				fPort:      202,
				threshold:  threeSecondsDuration,
			},
			Expected: nil,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			a, _ := test.New(t)
			result, err := tc.Command.Execute()
			if tc.Expected != nil {
				a.So(err, should.BeNil)
				a.So(result, should.NotBeNil)
				a.So(result, should.Resemble, tc.Expected)
			} else {
				a.So(result, should.BeNil)
				a.So(errors.IsUnavailable(err), should.BeTrue)
			}
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
				req: &ttnpb.ALCSyncCommand_AppTimeReq{
					DeviceTime:  timestamppb.New(receivedAtTime.Add(2 * time.Second)),
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
				req: &ttnpb.ALCSyncCommand_AppTimeReq{
					DeviceTime:  timestamppb.New(receivedAtTime.Add(-2 * time.Second)),
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
				req: &ttnpb.ALCSyncCommand_AppTimeReq{
					DeviceTime:  timestamppb.New(receivedAtTime),
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
			a, _ := test.New(t)
			result, err := tc.Command.Execute()
			a.So(err, should.NotBeNil)
			a.So(errors.IsUnavailable(err), should.BeTrue)
			a.So(result, should.BeNil)
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
				req: &ttnpb.ALCSyncCommand_AppTimeReq{
					DeviceTime:  timestamppb.New(receivedAtTime.Add(2 * time.Second)),
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
				req: &ttnpb.ALCSyncCommand_AppTimeReq{
					DeviceTime:  timestamppb.New(receivedAtTime.Add(-2 * time.Second)),
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
				req: &ttnpb.ALCSyncCommand_AppTimeReq{
					DeviceTime:  timestamppb.New(receivedAtTime),
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
			a, _ := test.New(t)
			result, err := tc.Command.Execute()
			a.So(err, should.BeNil)
			a.So(result, should.NotBeNil)
		})
	}
}

func TestTimeSyncCommandResultMarshalsBytesCorrectly(t *testing.T) {
	t.Parallel()
	a, _ := test.New(t)
	expected := []byte{0x01, 0x05, 0x00, 0x00, 0x00, 0x02}
	result := &TimeSyncCommandResult{
		ans: &ttnpb.ALCSyncCommand_AppTimeAns{
			TimeCorrection: 5,
			TokenAns:       2,
		},
	}
	actual, err := result.MarshalBinary()
	a.So(err, should.BeNil)
	a.So(actual, should.Resemble, expected)
}
