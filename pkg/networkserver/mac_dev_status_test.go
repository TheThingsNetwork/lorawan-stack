// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

package networkserver

import (
	"testing"

	"github.com/mohae/deepcopy"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/errors/common"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestHandleDevStatusAns(t *testing.T) {
	for _, tc := range []struct {
		Name             string
		Device, Expected *ttnpb.EndDevice
		Payload          *ttnpb.MACCommand_DevStatusAns
		Error            error
	}{
		{
			Name:     "nil payload",
			Device:   &ttnpb.EndDevice{},
			Expected: &ttnpb.EndDevice{},
			Payload:  nil,
			Error:    common.ErrMissingPayload.New(nil),
		},
		{
			Name:     "no request",
			Device:   &ttnpb.EndDevice{},
			Expected: &ttnpb.EndDevice{},
			Payload:  ttnpb.NewPopulatedMACCommand_DevStatusAns(test.Randy, false),
			Error:    ErrMACRequestNotFound.New(nil),
		},
		{
			Name: "battery 42, margin 4",
			Device: &ttnpb.EndDevice{
				MACInfo: &ttnpb.MACInfo{},
				QueuedMACCommands: []*ttnpb.MACCommand{
					ttnpb.CID_DEV_STATUS.MACCommand(),
				},
			},
			Expected: &ttnpb.EndDevice{
				MACInfo: &ttnpb.MACInfo{
					// TODO: https://github.com/TheThingsIndustries/ttn/issues/292
				},
				QueuedMACCommands: []*ttnpb.MACCommand{},
			},
			Payload: &ttnpb.MACCommand_DevStatusAns{
				Battery: 42,
				Margin:  4,
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			dev := deepcopy.Copy(tc.Device).(*ttnpb.EndDevice)

			err := handleDevStatusAns(test.Context(), dev, tc.Payload)
			if tc.Error != nil {
				a.So(err, should.BeError)
				return
			}

			a.So(err, should.BeNil)
			a.So(dev, should.Resemble, tc.Expected)
		})
	}
}
