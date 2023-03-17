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

	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	timeSyncPayload      = []byte{0xB2, 0x87, 0x2C, 0x51, 0x12}
	frmPayload           = []byte{0x01, 0xB2, 0x87, 0x2C, 0x51, 0x12}
	receivedAtTime       = time.Date(2023, 3, 3, 10, 0, 0, 0, time.UTC)
	threeSecondsDuration = time.Duration(3) * time.Second
)

func TestNewTimeSyncCommandValidInput(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		Name string
		In   struct {
			TimeSyncPayload []byte
			ReceivedAt      time.Time
			Threshold       time.Duration
			FPort           uint32
		}
		Expected struct {
			Cmd  Command
			Rest []byte
			Err  error
		}
	}{
		{
			Name: "WithNoExtraBytes",
			In: struct {
				TimeSyncPayload []byte
				ReceivedAt      time.Time
				Threshold       time.Duration
				FPort           uint32
			}{
				TimeSyncPayload: timeSyncPayload,
				ReceivedAt:      receivedAtTime,
				Threshold:       threeSecondsDuration,
				FPort:           202,
			},
			Expected: struct {
				Cmd  Command
				Rest []byte
				Err  error
			}{
				Cmd: &TimeSyncCommand{
					req: &ttnpb.ALCSyncCommand_AppTimeReq{
						DeviceTime:  timestamppb.New(receivedAtTime),
						TokenReq:    2,
						AnsRequired: true,
					},
					threshold:  threeSecondsDuration,
					receivedAt: receivedAtTime,
					fPort:      202,
				},
				Rest: []byte{},
				Err:  nil,
			},
		},
		{
			Name: "WithExtraBytes",
			In: struct {
				TimeSyncPayload []byte
				ReceivedAt      time.Time
				Threshold       time.Duration
				FPort           uint32
			}{
				TimeSyncPayload: append(timeSyncPayload, 0x53, 0x31),
				Threshold:       threeSecondsDuration,
				ReceivedAt:      receivedAtTime,
				FPort:           202,
			},
			Expected: struct {
				Cmd  Command
				Rest []byte
				Err  error
			}{
				Cmd: &TimeSyncCommand{
					req: &ttnpb.ALCSyncCommand_AppTimeReq{
						DeviceTime:  timestamppb.New(receivedAtTime),
						TokenReq:    2,
						AnsRequired: true,
					},
					threshold:  threeSecondsDuration,
					receivedAt: receivedAtTime,
					fPort:      202,
				},
				Rest: []byte{0x53, 0x31},
				Err:  nil,
			},
		},
		{
			Name: "WithPayloadTooShort",
			In: struct {
				TimeSyncPayload []byte
				ReceivedAt      time.Time
				Threshold       time.Duration
				FPort           uint32
			}{
				TimeSyncPayload: timeSyncPayload[:4],
				ReceivedAt:      receivedAtTime,
				Threshold:       threeSecondsDuration,
				FPort:           202,
			},
			Expected: struct {
				Cmd  Command
				Rest []byte
				Err  error
			}{
				Cmd:  (*TimeSyncCommand)(nil),
				Rest: timeSyncPayload[:4],
				Err:  errUnknownCommand,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			a, _ := test.New(t)
			cmd, rest, err := newTimeSyncCommand(tc.In.TimeSyncPayload, tc.In.Threshold, tc.In.ReceivedAt, tc.In.FPort)
			a.So(cmd, should.Resemble, tc.Expected.Cmd)
			a.So(rest, should.Resemble, tc.Expected.Rest)
			a.So(err, should.Resemble, tc.Expected.Err)
		})
	}
}

func TestMakeCommandValidInput(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		Name string
		In   struct {
			CID      ttnpb.ALCSyncCommandIdentifier
			CPayload []byte
		}
		Expected struct {
			CID  ttnpb.ALCSyncCommandIdentifier
			Rest []byte
		}
	}{
		{
			Name: "TimeSyncCommandWithNoExtraBytes",
			In: struct {
				CID      ttnpb.ALCSyncCommandIdentifier
				CPayload []byte
			}{
				CID:      ttnpb.ALCSyncCommandIdentifier_CID_APP_TIME,
				CPayload: timeSyncPayload,
			},
			Expected: struct {
				CID  ttnpb.ALCSyncCommandIdentifier
				Rest []byte
			}{
				CID:  ttnpb.ALCSyncCommandIdentifier_CID_APP_TIME,
				Rest: []byte{},
			},
		},
		{
			Name: "TimeSyncCommandWithExtraBytes",
			In: struct {
				CID      ttnpb.ALCSyncCommandIdentifier
				CPayload []byte
			}{
				CID:      ttnpb.ALCSyncCommandIdentifier_CID_APP_TIME,
				CPayload: append(timeSyncPayload, 0x53, 0x31),
			},
			Expected: struct {
				CID  ttnpb.ALCSyncCommandIdentifier
				Rest []byte
			}{
				CID:  ttnpb.ALCSyncCommandIdentifier_CID_APP_TIME,
				Rest: []byte{0x53, 0x31},
			},
		},
	}

	// The uplink, fPort and data can be omitted since they are not targeted by this test.

	uplink := &ttnpb.ApplicationUplink{}
	data := &packageData{}
	var fPort uint32

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			a, _ := test.New(t)
			cmd, rest, err := makeCommand(tc.In.CID, tc.In.CPayload, uplink, fPort, data)
			a.So(err, should.Equal, nil)
			a.So(cmd.Code(), should.Resemble, tc.Expected.CID)
			a.So(rest, should.Resemble, tc.Expected.Rest)
		})
	}
}

func TestMakeCommandInvalidInput(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		Name string
		In   struct {
			CID      ttnpb.ALCSyncCommandIdentifier
			CPayload []byte
		}
		Expected struct {
			Rest []byte
			Err  error
		}
	}{
		{
			Name: "WithUnsupportedCID",
			In: struct {
				CID      ttnpb.ALCSyncCommandIdentifier
				CPayload []byte
			}{
				CID:      ttnpb.ALCSyncCommandIdentifier_CID_PKG_VERSION,
				CPayload: timeSyncPayload,
			},
			Expected: struct {
				Rest []byte
				Err  error
			}{
				Rest: timeSyncPayload,
				Err:  errUnknownCommand.New(),
			},
		},
		{
			Name: "WithUnknownCID",
			In: struct {
				CID      ttnpb.ALCSyncCommandIdentifier
				CPayload []byte
			}{
				CID:      42,
				CPayload: timeSyncPayload,
			},
			Expected: struct {
				Rest []byte
				Err  error
			}{
				Rest: timeSyncPayload,
				Err:  errUnknownCommand.New(),
			},
		},
	}

	// The uplink, fPort and data can be omitted since they are not targeted by this test.

	uplink := &ttnpb.ApplicationUplink{}
	data := &packageData{}
	var fPort uint32

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			a, _ := test.New(t)
			cmd, rest, err := makeCommand(tc.In.CID, tc.In.CPayload, uplink, fPort, data)
			a.So(cmd, should.Equal, nil)
			a.So(rest, should.Resemble, tc.Expected.Rest)
			a.So(err, should.Resemble, tc.Expected.Err)
		})
	}
}

func TestMakeCommandsValidInput(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		Name string
		In   struct {
			Uplink *ttnpb.ApplicationUplink
			FPort  uint32
			Data   *packageData
		}
		Expected struct {
			Cmds []Command
			Err  error
		}
	}{
		{
			Name: "HandlesSingleCommand",
			In: struct {
				Uplink *ttnpb.ApplicationUplink
				FPort  uint32
				Data   *packageData
			}{
				Uplink: &ttnpb.ApplicationUplink{
					FrmPayload: frmPayload,
					ReceivedAt: timestamppb.New(receivedAtTime),
				},
				Data: &packageData{
					Threshold: threeSecondsDuration,
				},
				FPort: 202,
			},
			Expected: struct {
				Cmds []Command
				Err  error
			}{
				Cmds: []Command{
					&TimeSyncCommand{
						req: &ttnpb.ALCSyncCommand_AppTimeReq{
							DeviceTime:  timestamppb.New(receivedAtTime),
							TokenReq:    2,
							AnsRequired: true,
						},
						fPort:      202,
						receivedAt: receivedAtTime,
						threshold:  threeSecondsDuration,
					},
				},
				Err: nil,
			},
		},
		{
			Name: "HandlesSingleCommandAndRest",
			In: struct {
				Uplink *ttnpb.ApplicationUplink
				FPort  uint32
				Data   *packageData
			}{
				Uplink: &ttnpb.ApplicationUplink{
					FrmPayload: append(frmPayload, 0x53, 0x31),
					ReceivedAt: timestamppb.New(receivedAtTime),
				},
				Data: &packageData{
					Threshold: threeSecondsDuration,
				},
				FPort: 202,
			},
			Expected: struct {
				Cmds []Command
				Err  error
			}{
				Cmds: []Command{
					&TimeSyncCommand{
						req: &ttnpb.ALCSyncCommand_AppTimeReq{
							DeviceTime:  timestamppb.New(receivedAtTime),
							TokenReq:    2,
							AnsRequired: true,
						},
						fPort:      202,
						receivedAt: receivedAtTime,
						threshold:  threeSecondsDuration,
					},
				},
				Err: errUnknownCommand.New(),
			},
		},
		{
			Name: "HandlesMultipleCommands",
			In: struct {
				Uplink *ttnpb.ApplicationUplink
				FPort  uint32
				Data   *packageData
			}{
				Uplink: &ttnpb.ApplicationUplink{
					FrmPayload: append(frmPayload, frmPayload...),
					FPort:      202,
					ReceivedAt: timestamppb.New(receivedAtTime),
				},
				Data: &packageData{
					Threshold: threeSecondsDuration,
				},
				FPort: 202,
			},
			Expected: struct {
				Cmds []Command
				Err  error
			}{
				Cmds: []Command{
					&TimeSyncCommand{
						req: &ttnpb.ALCSyncCommand_AppTimeReq{
							DeviceTime:  timestamppb.New(receivedAtTime),
							TokenReq:    2,
							AnsRequired: true,
						},
						fPort:      202,
						receivedAt: receivedAtTime,
						threshold:  threeSecondsDuration,
					},
					&TimeSyncCommand{
						req: &ttnpb.ALCSyncCommand_AppTimeReq{
							DeviceTime:  timestamppb.New(receivedAtTime),
							TokenReq:    2,
							AnsRequired: true,
						},
						fPort:      202,
						receivedAt: receivedAtTime,
						threshold:  threeSecondsDuration,
					},
				},
				Err: nil,
			},
		},
		{
			Name: "HandlesMultipleCommandsAndRest",
			In: struct {
				Uplink *ttnpb.ApplicationUplink
				FPort  uint32
				Data   *packageData
			}{
				Uplink: &ttnpb.ApplicationUplink{
					FrmPayload: append(
						append(frmPayload, frmPayload...),
						0x53, 0x31,
					),
					FPort:      202,
					ReceivedAt: timestamppb.New(receivedAtTime),
				},
				Data: &packageData{
					Threshold: threeSecondsDuration,
				},
				FPort: 202,
			},
			Expected: struct {
				Cmds []Command
				Err  error
			}{
				Cmds: []Command{
					&TimeSyncCommand{
						req: &ttnpb.ALCSyncCommand_AppTimeReq{
							DeviceTime:  timestamppb.New(receivedAtTime),
							TokenReq:    2,
							AnsRequired: true,
						},
						fPort:      202,
						receivedAt: receivedAtTime,
						threshold:  threeSecondsDuration,
					},
					&TimeSyncCommand{
						req: &ttnpb.ALCSyncCommand_AppTimeReq{
							DeviceTime:  timestamppb.New(receivedAtTime),
							TokenReq:    2,
							AnsRequired: true,
						},
						fPort:      202,
						receivedAt: receivedAtTime,
						threshold:  threeSecondsDuration,
					},
				},
				Err: errUnknownCommand.New(),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			a, _ := test.New(t)
			cmds, err := MakeCommands(tc.In.Uplink, tc.In.FPort, tc.In.Data)
			a.So(err, should.Resemble, tc.Expected.Err)
			a.So(cmds, should.Resemble, tc.Expected.Cmds)
		})
	}
}

func TestMakeDownlinkSerializesAppTimeAns(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		Name string
		In   struct {
			Ans   []Result
			FPort uint32
		}
		Expected struct {
			Downlink *ttnpb.ApplicationDownlink
		}
	}{
		{
			Name: "SingleResult",
			In: struct {
				Ans   []Result
				FPort uint32
			}{
				Ans: []Result{
					&TimeSyncCommandResult{
						ans: &ttnpb.ALCSyncCommand_AppTimeAns{
							TimeCorrection: 1,
							TokenAns:       2,
						},
					},
				},
				FPort: 202,
			},
			Expected: struct {
				Downlink *ttnpb.ApplicationDownlink
			}{
				Downlink: &ttnpb.ApplicationDownlink{
					FPort: 202,
					FrmPayload: []byte{
						0x01, 0x01, 0x00, 0x00, 0x00, 0x02,
					},
				},
			},
		},
		{
			Name: "MultipleResults",
			In: struct {
				Ans   []Result
				FPort uint32
			}{
				Ans: []Result{
					&TimeSyncCommandResult{
						ans: &ttnpb.ALCSyncCommand_AppTimeAns{
							TimeCorrection: 1,
							TokenAns:       2,
						},
					},
					&TimeSyncCommandResult{
						ans: &ttnpb.ALCSyncCommand_AppTimeAns{
							TimeCorrection: 3,
							TokenAns:       4,
						},
					},
				},
				FPort: 202,
			},
			Expected: struct {
				Downlink *ttnpb.ApplicationDownlink
			}{
				Downlink: &ttnpb.ApplicationDownlink{
					FPort: 202,
					FrmPayload: []byte{
						0x01, 0x01, 0x00, 0x00, 0x00, 0x02,
						0x01, 0x03, 0x00, 0x00, 0x00, 0x04,
					},
				},
			},
		},
		{
			Name: "MultipleResultsContainingNil",
			In: struct {
				Ans   []Result
				FPort uint32
			}{
				Ans: []Result{
					&TimeSyncCommandResult{
						ans: &ttnpb.ALCSyncCommand_AppTimeAns{
							TimeCorrection: 1,
							TokenAns:       2,
						},
					},
					nil,
					&TimeSyncCommandResult{
						ans: &ttnpb.ALCSyncCommand_AppTimeAns{
							TimeCorrection: 3,
							TokenAns:       4,
						},
					},
				},
				FPort: 202,
			},
			Expected: struct {
				Downlink *ttnpb.ApplicationDownlink
			}{
				Downlink: &ttnpb.ApplicationDownlink{
					FPort: 202,
					FrmPayload: []byte{
						0x01, 0x01, 0x00, 0x00, 0x00, 0x02,
						0x01, 0x03, 0x00, 0x00, 0x00, 0x04,
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			a, _ := test.New(t)
			downlink, err := MakeDownlink(tc.In.Ans, tc.In.FPort)
			a.So(err, should.BeNil)
			a.So(downlink, should.Resemble, tc.Expected.Downlink)
		})
	}
}
