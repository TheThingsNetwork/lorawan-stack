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
	"math"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

var errUnknownCommand = errors.DefineNotFound("unknown_command", "unknown command")

// TimeSyncCommand is the command for time synchronization.
type TimeSyncCommand struct {
	req        *AppTimeReq
	receivedAt time.Time
	threshold  time.Duration
	fPort      uint32
}

// Code implements commands.Command.
func (*TimeSyncCommand) Code() uint8 {
	return 0x01
}

// Execute implements commands.Command.
func (cmd *TimeSyncCommand) Execute() ([]*ttnpb.ApplicationDownlink, error) {
	difference := cmd.receivedAt.Sub(cmd.req.DeviceTime)
	exceedsThreshold := math.Abs(difference.Seconds()) > cmd.threshold.Seconds()
	if !cmd.req.AnsRequired && !exceedsThreshold {
		return nil, nil
	}

	ans := &AppTimeAns{
		TimeCorrection: int32(difference.Seconds()),
		TokenAns:       cmd.req.TokenReq,
	}
	frmPayload, err := ans.MarshalBinary()
	if err != nil {
		return nil, err
	}
	downlink := &ttnpb.ApplicationDownlink{
		FPort:      cmd.fPort,
		FrmPayload: frmPayload,
	}
	return []*ttnpb.ApplicationDownlink{downlink}, nil
}
