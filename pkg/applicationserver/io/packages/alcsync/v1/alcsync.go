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

	"go.thethings.network/lorawan-stack/v3/pkg/gpstime"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	lorautil "go.thethings.network/lorawan-stack/v3/pkg/util/lora"
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
		return nil, data, errUnknownCommand.New()
	}

	cPayload, rest := data[:lenBytes], data[lenBytes:]
	deviceTimeGPSSeconds := binary.LittleEndian.Uint32(cPayload[:4])
	durationGPS := time.Duration(deviceTimeGPSSeconds) * time.Second
	deviceTime := gpstime.Parse(durationGPS)
	tokenReq := cPayload[4] & 0x0F
	ansRequired := (cPayload[4] & 0x10) != 0

	cmd := &TimeSyncCommand{
		req: &AppTimeReq{
			DeviceTime:  deviceTime,
			TokenReq:    tokenReq,
			AnsRequired: ansRequired,
		},
		receivedAt: receivedAt,
		threshold:  threshold,
		fPort:      fPort,
	}
	return cmd, rest, nil
}

// MakeCommands parses the uplink payload and returns the commands.
func MakeCommands(up *ttnpb.ApplicationUplink, fPort uint32, data *packageData) ([]Command, error) {
	cID, cPayload := up.FrmPayload[0], up.FrmPayload[1:]
	commands := make([]Command, 0)
	for {
		cmd, rest, err := makeCommand(cID, cPayload, up, fPort, data)
		if err != nil {
			return commands, err
		}
		commands = append(commands, cmd)

		if len(rest) == 0 { // No commands left.
			break
		}
		cID, cPayload = rest[0], rest[1:]
	}
	return commands, nil
}

// makeCommand parses the payload based on the command ID.
func makeCommand(
	cID byte,
	cPayload []byte,
	up *ttnpb.ApplicationUplink,
	fPort uint32,
	data *packageData,
) (Command, []byte, error) {
	receivedAt := lorautil.GetAdjustedReceivedAt(up)
	threshold := data.Threshold
	switch cID {
	case 0x01:
		return newTimeSyncCommand(cPayload, threshold, receivedAt.AsTime(), fPort)
	default:
		return nil, cPayload, errUnknownCommand.New()
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
			return nil, err
		}
		frmPayload = append(frmPayload, b...)
	}
	downlink := &ttnpb.ApplicationDownlink{
		FPort:      fPort,
		FrmPayload: frmPayload,
	}
	return downlink, nil
}
