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

// Package alcsyncv1 provides the LoRa Application Layer Clock Synchronization Package.
package alcsyncv1

import (
	"encoding/binary"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/gpstime"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	lorautil "go.thethings.network/lorawan-stack/v3/pkg/util/lora"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// newTimeSyncCommand builds a new TimeSyncCommand.
func newTimeSyncCommand(
	data []byte,
	threshold time.Duration,
	receivedAt time.Time,
	fPort uint32,
) (*TimeSyncCommand, []byte, error) {
	// DeviceTime - bytes [0, 3].
	// Param - byte 4 (bits: RFU [7:5]; AnsRequired 4; TokenReq [3:0]).

	lenBytes := 5
	if len(data) < lenBytes {
		return nil, data, errInsufficientLength.WithAttributes(
			"expected_length", lenBytes,
			"actual_length", len(data),
		).New()
	}

	cPayload, rest := data[:lenBytes], data[lenBytes:]
	deviceTimeGPSSeconds := binary.LittleEndian.Uint32(cPayload[:4])
	durationGPS := time.Duration(deviceTimeGPSSeconds) * time.Second
	deviceTime := gpstime.Parse(durationGPS)
	tokenReq := cPayload[4] & 0x0F
	ansRequired := (cPayload[4] & 0x10) != 0

	cmd := &TimeSyncCommand{
		req: &ttnpb.ALCSyncCommand_AppTimeReq{
			DeviceTime:  timestamppb.New(deviceTime),
			TokenReq:    uint32(tokenReq),
			AnsRequired: ansRequired,
		},
		receivedAt: receivedAt,
		threshold:  threshold,
		fPort:      fPort,
	}
	return cmd, rest, nil
}

// MakeCommands parses the uplink payload and returns the commands.
func MakeCommands(up *ttnpb.ApplicationUplink, fPort uint32, data *packageData) ([]Command, events.Builders, error) {
	cID, cPayload := ttnpb.ALCSyncCommandIdentifier(up.FrmPayload[0]), up.FrmPayload[1:]
	commands := make([]Command, 0)
	evts := make(events.Builders, 0)
	for {
		cmd, rest, err := makeCommand(cID, cPayload, up, fPort, data)
		if err != nil {
			err := errCommandCreationFailed.WithCause(err).WithAttributes(
				"command_id", cID,
				"command_payload", cPayload,
				"remaining_payload", rest,
				"received_at", up.ReceivedAt,
			).New()
			evts = append(evts, EvtPkgFail.With(events.WithData(err)))
			return commands, evts, err
		}
		commands = append(commands, cmd)
		evts = append(evts, cmd.CommandReceivedEventBuilder())

		if len(rest) == 0 { // No commands left.
			break
		}
		cID, cPayload = ttnpb.ALCSyncCommandIdentifier(rest[0]), rest[1:]
	}
	return commands, evts, nil
}

// makeCommand parses the payload based on the command ID.
func makeCommand(
	cID ttnpb.ALCSyncCommandIdentifier,
	cPayload []byte,
	up *ttnpb.ApplicationUplink,
	fPort uint32,
	data *packageData,
) (Command, []byte, error) {
	receivedAt := lorautil.GetAdjustedReceivedAt(up)
	threshold := data.Threshold
	switch cID {
	case ttnpb.ALCSyncCommandIdentifier_ALCSYNC_CID_APP_TIME:
		return newTimeSyncCommand(cPayload, threshold, receivedAt.AsTime(), fPort)
	case ttnpb.ALCSyncCommandIdentifier_ALCSYNC_CID_PKG_VERSION,
		ttnpb.ALCSyncCommandIdentifier_ALCSYNC_CID_APP_DEV_TIME_PERIODICITY,
		ttnpb.ALCSyncCommandIdentifier_ALCSYNC_CID_FORCE_DEV_RESYNC:
		return nil, cPayload, errUnsuportedCommand.WithAttributes(
			"command_id", cID,
			"command_payload", cPayload,
		).New()
	default:
		return nil, cPayload, errUnknownCommand.WithAttributes(
			"command_id", cID,
			"command_payload", cPayload,
		).New()
	}
}

// MakeDownlink builds a single downlink message from the results.
func MakeDownlink(results []Result, fPort uint32) (*ttnpb.ApplicationDownlink, error) {
	frmPayload := make([]byte, 0)
	for _, result := range results {
		if result == nil {
			continue
		}
		b, err := result.MarshalBinary()
		if err != nil {
			return nil, errDownlinkCreationFailed.WithCause(err).New()
		}
		frmPayload = append(frmPayload, b...)
	}
	downlink := &ttnpb.ApplicationDownlink{
		FPort:      fPort,
		FrmPayload: frmPayload,
	}
	return downlink, nil
}
