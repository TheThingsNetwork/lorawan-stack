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
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestHandleDutyCycleAns(t *testing.T) {
	for _, tc := range []struct {
		Name             string
		Device, Expected *ttnpb.EndDevice
		Error            error
	}{
		{
			Name:     "no request",
			Device:   &ttnpb.EndDevice{},
			Expected: &ttnpb.EndDevice{},
			Error:    ErrMACRequestNotFound.New(nil),
		},
		{
			Name: "2048",
			Device: &ttnpb.EndDevice{
				PendingMACCommands: []*ttnpb.MACCommand{
					(&ttnpb.MACCommand_DutyCycleReq{
						MaxDutyCycle: ttnpb.DUTY_CYCLE_2048,
					}).MACCommand(),
				},
			},
			Expected: &ttnpb.EndDevice{
				PendingMACCommands: []*ttnpb.MACCommand{},
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			dev := deepcopy.Copy(tc.Device).(*ttnpb.EndDevice)

			err := handleDutyCycleAns(test.Context(), dev)
			if tc.Error != nil {
				a.So(err, should.BeError)
				return
			}

			a.So(err, should.BeNil)
			a.So(dev, should.Resemble, tc.Expected)
		})
	}
}
