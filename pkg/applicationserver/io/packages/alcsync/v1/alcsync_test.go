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
	timeSyncCID          = byte(0x01)
	timeSyncPayload      = []byte{0xB2, 0x87, 0x2C, 0x51, 0x12}
	frmPayload           = []byte{0x01, 0xB2, 0x87, 0x2C, 0x51, 0x12}
	receivedAtTime       = time.Date(2023, 3, 3, 10, 0, 0, 0, time.UTC)
	threeSecondsDuration = time.Duration(3) * time.Second
)

func TestNewTimeSyncCommandValidInput(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name     string
		uplink   *ttnpb.ApplicationUplink
		expected struct {
			cmd  Command
			rest []byte
			err  error
		}
	}{
		{
			name: "WithNoExtraBytes",
			uplink: &ttnpb.ApplicationUplink{
				FrmPayload: frmPayload,
				FPort:      202,
				ReceivedAt: timestamppb.New(receivedAtTime),
			},
			expected: struct {
				cmd  Command
				rest []byte
				err  error
			}{
				cmd: &TimeSyncCommand{
					req: &AppTimeReq{
						DeviceTime:  receivedAtTime,
						TokenReq:    2,
						AnsRequired: true,
					},
					receivedAt: receivedAtTime,
					fPort:      202,
				},
				rest: []byte{},
				err:  nil,
			},
		},
		{
			name: "WithExtraBytes",
			uplink: &ttnpb.ApplicationUplink{
				FrmPayload: append(frmPayload, 0x53, 0x31),
				FPort:      202,
				ReceivedAt: timestamppb.New(receivedAtTime),
			},
			expected: struct {
				cmd  Command
				rest []byte
				err  error
			}{
				cmd: &TimeSyncCommand{
					req: &AppTimeReq{
						DeviceTime:  receivedAtTime,
						TokenReq:    2,
						AnsRequired: true,
					},
					receivedAt: receivedAtTime,
					fPort:      202,
				},
				rest: []byte{0x53, 0x31},
				err:  nil,
			},
		},
		{
			name: "WithPayloadTooShort",
			uplink: &ttnpb.ApplicationUplink{
				FrmPayload: frmPayload[:4],
				FPort:      202,
				ReceivedAt: timestamppb.New(receivedAtTime),
			},
			expected: struct {
				cmd  Command
				rest []byte
				err  error
			}{
				cmd:  (*TimeSyncCommand)(nil),
				rest: frmPayload[1:4],
				err:  errUnknownCommand.New(),
			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a, _ := test.New(t)
			cPayload := tc.uplink.FrmPayload[1:]
			cmd, rest, err := newTimeSyncCommand(cPayload, 0, tc.uplink.ReceivedAt.AsTime(), tc.uplink.FPort)

			a.So(cmd, should.Resemble, tc.expected.cmd)
			a.So(rest, should.Resemble, tc.expected.rest)
			a.So(err, should.Resemble, tc.expected.err)
		})
	}
}

func TestMakeCommandValidInput(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name string
		in   struct {
			cID      byte
			cPayload []byte
		}
		expected struct {
			cID  byte
			rest []byte
		}
	}{
		{
			name: "TimeSyncCommandWithNoExtraBytes",
			in: struct {
				cID      byte
				cPayload []byte
			}{
				cID:      timeSyncCID,
				cPayload: timeSyncPayload,
			},
			expected: struct {
				cID  byte
				rest []byte
			}{
				cID:  timeSyncCID,
				rest: []byte{},
			},
		},
		{
			name: "TimeSyncCommandWithExtraBytes",
			in: struct {
				cID      byte
				cPayload []byte
			}{
				cID:      timeSyncCID,
				cPayload: append(timeSyncPayload, 0x53, 0x31),
			},
			expected: struct {
				cID  byte
				rest []byte
			}{
				cID:  timeSyncCID,
				rest: []byte{0x53, 0x31},
			},
		},
	}

	uplink := &ttnpb.ApplicationUplink{
		FPort:      202,
		ReceivedAt: timestamppb.New(receivedAtTime),
	}
	data := &packageData{
		FPort:     202,
		Threshold: threeSecondsDuration,
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a, _ := test.New(t)
			cmd, rest, err := makeCommand(tc.in.cID, tc.in.cPayload, uplink, data)
			a.So(err, should.Equal, nil)
			a.So(cmd.Code(), should.Resemble, tc.expected.cID)
			a.So(rest, should.Resemble, tc.expected.rest)
		})
	}
}

func TestMakeCommandInvalidInput(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name string
		in   struct {
			cID      byte
			cPayload []byte
		}
		expected struct {
			cID  byte
			rest []byte
			err  error
		}
	}{
		{
			name: "WithUnknownCID",
			in: struct {
				cID      byte
				cPayload []byte
			}{
				cID:      0x43,
				cPayload: timeSyncPayload,
			},
			expected: struct {
				cID  byte
				rest []byte
				err  error
			}{
				cID:  0x43,
				rest: timeSyncPayload,
				err:  errUnknownCommand.New(),
			},
		},
	}
	uplink := &ttnpb.ApplicationUplink{
		FPort:      202,
		ReceivedAt: timestamppb.New(receivedAtTime),
	}
	data := &packageData{
		FPort:     202,
		Threshold: threeSecondsDuration,
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a, _ := test.New(t)
			cmd, rest, err := makeCommand(tc.in.cID, tc.in.cPayload, uplink, data)
			a.So(cmd, should.Equal, nil)
			a.So(rest, should.Resemble, tc.expected.rest)
			a.So(err, should.Resemble, tc.expected.err)
		})
	}
}

func TestMakeCommandsValidInput(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name string
		in   struct {
			uplink *ttnpb.ApplicationUplink
			data   *packageData
		}
		expected struct {
			cmds []Command
			err  error
		}
	}{
		{
			name: "HandlesSingleCommand",
			in: struct {
				uplink *ttnpb.ApplicationUplink
				data   *packageData
			}{
				uplink: &ttnpb.ApplicationUplink{
					FrmPayload: frmPayload,
					FPort:      202,
					ReceivedAt: timestamppb.New(receivedAtTime),
				},
				data: &packageData{
					FPort:     202,
					Threshold: threeSecondsDuration,
				},
			},
			expected: struct {
				cmds []Command
				err  error
			}{
				cmds: []Command{
					&TimeSyncCommand{
						req: &AppTimeReq{
							DeviceTime:  receivedAtTime,
							TokenReq:    2,
							AnsRequired: true,
						},
						fPort:      202,
						receivedAt: receivedAtTime,
						threshold:  threeSecondsDuration,
					},
				},
				err: nil,
			},
		},
		{
			name: "HandlesSingleCommandAndRest",
			in: struct {
				uplink *ttnpb.ApplicationUplink
				data   *packageData
			}{
				uplink: &ttnpb.ApplicationUplink{
					FrmPayload: append(frmPayload, 0x53, 0x31),
					FPort:      202,
					ReceivedAt: timestamppb.New(receivedAtTime),
				},
				data: &packageData{
					FPort:     202,
					Threshold: threeSecondsDuration,
				},
			},
			expected: struct {
				cmds []Command
				err  error
			}{
				cmds: []Command{
					&TimeSyncCommand{
						req: &AppTimeReq{
							DeviceTime:  receivedAtTime,
							TokenReq:    2,
							AnsRequired: true,
						},
						fPort:      202,
						receivedAt: receivedAtTime,
						threshold:  threeSecondsDuration,
					},
				},
				err: errUnknownCommand.New(),
			},
		},
		{
			name: "HandlesMultipleCommands",
			in: struct {
				uplink *ttnpb.ApplicationUplink
				data   *packageData
			}{
				uplink: &ttnpb.ApplicationUplink{
					FrmPayload: append(frmPayload, frmPayload...),
					FPort:      202,
					ReceivedAt: timestamppb.New(receivedAtTime),
				},
				data: &packageData{
					FPort:     202,
					Threshold: threeSecondsDuration,
				},
			},
			expected: struct {
				cmds []Command
				err  error
			}{
				cmds: []Command{
					&TimeSyncCommand{
						req: &AppTimeReq{
							DeviceTime:  receivedAtTime,
							TokenReq:    2,
							AnsRequired: true,
						},
						fPort:      202,
						receivedAt: receivedAtTime,
						threshold:  threeSecondsDuration,
					},
					&TimeSyncCommand{
						req: &AppTimeReq{
							DeviceTime:  receivedAtTime,
							TokenReq:    2,
							AnsRequired: true,
						},
						fPort:      202,
						receivedAt: receivedAtTime,
						threshold:  threeSecondsDuration,
					},
				},
				err: nil,
			},
		},
		{
			name: "HandlesMultipleCommandsAndRest",
			in: struct {
				uplink *ttnpb.ApplicationUplink
				data   *packageData
			}{
				uplink: &ttnpb.ApplicationUplink{
					FrmPayload: append(
						append(frmPayload, frmPayload...),
						0x53, 0x31,
					),
					FPort:      202,
					ReceivedAt: timestamppb.New(receivedAtTime),
				},
				data: &packageData{
					FPort:     202,
					Threshold: threeSecondsDuration,
				},
			},
			expected: struct {
				cmds []Command
				err  error
			}{
				cmds: []Command{
					&TimeSyncCommand{
						req: &AppTimeReq{
							DeviceTime:  receivedAtTime,
							TokenReq:    2,
							AnsRequired: true,
						},
						fPort:      202,
						receivedAt: receivedAtTime,
						threshold:  threeSecondsDuration,
					},
					&TimeSyncCommand{
						req: &AppTimeReq{
							DeviceTime:  receivedAtTime,
							TokenReq:    2,
							AnsRequired: true,
						},
						fPort:      202,
						receivedAt: receivedAtTime,
						threshold:  threeSecondsDuration,
					},
				},
				err: errUnknownCommand.New(),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a, _ := test.New(t)
			cmds, err := MakeCommands(tc.in.uplink, tc.in.data)
			a.So(err, should.Resemble, tc.expected.err)
			a.So(cmds, should.Resemble, tc.expected.cmds)
		})
	}
}
