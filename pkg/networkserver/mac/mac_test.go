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

package mac_test

import (
	"testing"

	"github.com/smartystreets/assertions"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/mac"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestHandleMACResponseBlock(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		Name string

		CID          ttnpb.MACCommandIdentifier
		AllowMissing bool
		F            func(*assertions.Assertion, *ttnpb.MACCommand) error
		Commands     []*ttnpb.MACCommand

		Result      []*ttnpb.MACCommand
		AssertError func(*assertions.Assertion, error)
	}{
		{
			Name: "LinkADRReq",

			CID:          ttnpb.MACCommandIdentifier_CID_LINK_ADR,
			AllowMissing: false,
			F: func(a *assertions.Assertion, cmd *ttnpb.MACCommand) error {
				a.So(cmd, should.Resemble, &ttnpb.MACCommand{
					Cid: ttnpb.MACCommandIdentifier_CID_LINK_ADR,
					Payload: &ttnpb.MACCommand_LinkAdrReq{
						LinkAdrReq: &ttnpb.MACCommand_LinkADRReq{
							DataRateIndex:      1,
							TxPowerIndex:       2,
							ChannelMask:        []bool{false, true, false},
							ChannelMaskControl: 3,
							NbTrans:            4,
						},
					},
				})
				return nil
			},
			Commands: []*ttnpb.MACCommand{
				{
					Cid: ttnpb.MACCommandIdentifier_CID_LINK_ADR,
					Payload: &ttnpb.MACCommand_LinkAdrReq{
						LinkAdrReq: &ttnpb.MACCommand_LinkADRReq{
							DataRateIndex:      1,
							TxPowerIndex:       2,
							ChannelMask:        []bool{false, true, false},
							ChannelMaskControl: 3,
							NbTrans:            4,
						},
					},
				},
			},

			Result: []*ttnpb.MACCommand{},
		},
		{
			Name: "LinkADRReq+LinkADRReq",

			CID:          ttnpb.MACCommandIdentifier_CID_LINK_ADR,
			AllowMissing: false,
			F: func(a *assertions.Assertion, cmd *ttnpb.MACCommand) error {
				a.So(cmd, should.Resemble, &ttnpb.MACCommand{
					Cid: ttnpb.MACCommandIdentifier_CID_LINK_ADR,
					Payload: &ttnpb.MACCommand_LinkAdrReq{
						LinkAdrReq: &ttnpb.MACCommand_LinkADRReq{
							DataRateIndex:      1,
							TxPowerIndex:       2,
							ChannelMask:        []bool{false, true, false},
							ChannelMaskControl: 3,
							NbTrans:            4,
						},
					},
				})
				return nil
			},
			Commands: []*ttnpb.MACCommand{
				{
					Cid: ttnpb.MACCommandIdentifier_CID_LINK_ADR,
					Payload: &ttnpb.MACCommand_LinkAdrReq{
						LinkAdrReq: &ttnpb.MACCommand_LinkADRReq{
							DataRateIndex:      1,
							TxPowerIndex:       2,
							ChannelMask:        []bool{false, true, false},
							ChannelMaskControl: 3,
							NbTrans:            4,
						},
					},
				},
				{
					Cid: ttnpb.MACCommandIdentifier_CID_LINK_ADR,
					Payload: &ttnpb.MACCommand_LinkAdrReq{
						LinkAdrReq: &ttnpb.MACCommand_LinkADRReq{
							DataRateIndex:      1,
							TxPowerIndex:       2,
							ChannelMask:        []bool{false, true, false},
							ChannelMaskControl: 3,
							NbTrans:            4,
						},
					},
				},
			},

			Result: []*ttnpb.MACCommand{},
		},
		{
			Name: "LinkADRReq+DlChannelReq",

			CID:          ttnpb.MACCommandIdentifier_CID_LINK_ADR,
			AllowMissing: false,
			F: func(a *assertions.Assertion, cmd *ttnpb.MACCommand) error {
				a.So(cmd, should.Resemble, &ttnpb.MACCommand{
					Cid: ttnpb.MACCommandIdentifier_CID_LINK_ADR,
					Payload: &ttnpb.MACCommand_LinkAdrReq{
						LinkAdrReq: &ttnpb.MACCommand_LinkADRReq{
							DataRateIndex:      1,
							TxPowerIndex:       2,
							ChannelMask:        []bool{false, true, false},
							ChannelMaskControl: 3,
							NbTrans:            4,
						},
					},
				})
				return nil
			},
			Commands: []*ttnpb.MACCommand{
				{
					Cid: ttnpb.MACCommandIdentifier_CID_LINK_ADR,
					Payload: &ttnpb.MACCommand_LinkAdrReq{
						LinkAdrReq: &ttnpb.MACCommand_LinkADRReq{
							DataRateIndex:      1,
							TxPowerIndex:       2,
							ChannelMask:        []bool{false, true, false},
							ChannelMaskControl: 3,
							NbTrans:            4,
						},
					},
				},
				{
					Cid: ttnpb.MACCommandIdentifier_CID_DL_CHANNEL,
					Payload: &ttnpb.MACCommand_DlChannelReq{
						DlChannelReq: &ttnpb.MACCommand_DLChannelReq{
							ChannelIndex: 1,
							Frequency:    2,
						},
					},
				},
			},

			Result: []*ttnpb.MACCommand{
				{
					Cid: ttnpb.MACCommandIdentifier_CID_DL_CHANNEL,
					Payload: &ttnpb.MACCommand_DlChannelReq{
						DlChannelReq: &ttnpb.MACCommand_DLChannelReq{
							ChannelIndex: 1,
							Frequency:    2,
						},
					},
				},
			},
		},
		{
			Name: "DlChannelReq+LinkADRReq",

			CID:          ttnpb.MACCommandIdentifier_CID_LINK_ADR,
			AllowMissing: false,
			F: func(a *assertions.Assertion, cmd *ttnpb.MACCommand) error {
				a.So(cmd, should.Resemble, &ttnpb.MACCommand{
					Cid: ttnpb.MACCommandIdentifier_CID_LINK_ADR,
					Payload: &ttnpb.MACCommand_LinkAdrReq{
						LinkAdrReq: &ttnpb.MACCommand_LinkADRReq{
							DataRateIndex:      1,
							TxPowerIndex:       2,
							ChannelMask:        []bool{false, true, false},
							ChannelMaskControl: 3,
							NbTrans:            4,
						},
					},
				})
				return nil
			},
			Commands: []*ttnpb.MACCommand{
				{
					Cid: ttnpb.MACCommandIdentifier_CID_DL_CHANNEL,
					Payload: &ttnpb.MACCommand_DlChannelReq{
						DlChannelReq: &ttnpb.MACCommand_DLChannelReq{
							ChannelIndex: 1,
							Frequency:    2,
						},
					},
				},
				{
					Cid: ttnpb.MACCommandIdentifier_CID_LINK_ADR,
					Payload: &ttnpb.MACCommand_LinkAdrReq{
						LinkAdrReq: &ttnpb.MACCommand_LinkADRReq{
							DataRateIndex:      1,
							TxPowerIndex:       2,
							ChannelMask:        []bool{false, true, false},
							ChannelMaskControl: 3,
							NbTrans:            4,
						},
					},
				},
			},

			Result: []*ttnpb.MACCommand{
				{
					Cid: ttnpb.MACCommandIdentifier_CID_DL_CHANNEL,
					Payload: &ttnpb.MACCommand_DlChannelReq{
						DlChannelReq: &ttnpb.MACCommand_DLChannelReq{
							ChannelIndex: 1,
							Frequency:    2,
						},
					},
				},
			},
		},
		{
			Name: "LinkADRReq+LinkADRReq+DlChannelReq",

			CID:          ttnpb.MACCommandIdentifier_CID_LINK_ADR,
			AllowMissing: false,
			F: func(a *assertions.Assertion, cmd *ttnpb.MACCommand) error {
				a.So(cmd, should.Resemble, &ttnpb.MACCommand{
					Cid: ttnpb.MACCommandIdentifier_CID_LINK_ADR,
					Payload: &ttnpb.MACCommand_LinkAdrReq{
						LinkAdrReq: &ttnpb.MACCommand_LinkADRReq{
							DataRateIndex:      1,
							TxPowerIndex:       2,
							ChannelMask:        []bool{false, true, false},
							ChannelMaskControl: 3,
							NbTrans:            4,
						},
					},
				})
				return nil
			},
			Commands: []*ttnpb.MACCommand{
				{
					Cid: ttnpb.MACCommandIdentifier_CID_LINK_ADR,
					Payload: &ttnpb.MACCommand_LinkAdrReq{
						LinkAdrReq: &ttnpb.MACCommand_LinkADRReq{
							DataRateIndex:      1,
							TxPowerIndex:       2,
							ChannelMask:        []bool{false, true, false},
							ChannelMaskControl: 3,
							NbTrans:            4,
						},
					},
				},
				{
					Cid: ttnpb.MACCommandIdentifier_CID_LINK_ADR,
					Payload: &ttnpb.MACCommand_LinkAdrReq{
						LinkAdrReq: &ttnpb.MACCommand_LinkADRReq{
							DataRateIndex:      1,
							TxPowerIndex:       2,
							ChannelMask:        []bool{false, true, false},
							ChannelMaskControl: 3,
							NbTrans:            4,
						},
					},
				},
				{
					Cid: ttnpb.MACCommandIdentifier_CID_DL_CHANNEL,
					Payload: &ttnpb.MACCommand_DlChannelReq{
						DlChannelReq: &ttnpb.MACCommand_DLChannelReq{
							ChannelIndex: 1,
							Frequency:    2,
						},
					},
				},
			},

			Result: []*ttnpb.MACCommand{
				{
					Cid: ttnpb.MACCommandIdentifier_CID_DL_CHANNEL,
					Payload: &ttnpb.MACCommand_DlChannelReq{
						DlChannelReq: &ttnpb.MACCommand_DLChannelReq{
							ChannelIndex: 1,
							Frequency:    2,
						},
					},
				},
			},
		},
		{
			Name: "DlChannelReq+LinkADRReq+LinkADRReq+DlChannelReq",

			CID:          ttnpb.MACCommandIdentifier_CID_LINK_ADR,
			AllowMissing: false,
			F: func(a *assertions.Assertion, cmd *ttnpb.MACCommand) error {
				a.So(cmd, should.Resemble, &ttnpb.MACCommand{
					Cid: ttnpb.MACCommandIdentifier_CID_LINK_ADR,
					Payload: &ttnpb.MACCommand_LinkAdrReq{
						LinkAdrReq: &ttnpb.MACCommand_LinkADRReq{
							DataRateIndex:      1,
							TxPowerIndex:       2,
							ChannelMask:        []bool{false, true, false},
							ChannelMaskControl: 3,
							NbTrans:            4,
						},
					},
				})
				return nil
			},
			Commands: []*ttnpb.MACCommand{
				{
					Cid: ttnpb.MACCommandIdentifier_CID_DL_CHANNEL,
					Payload: &ttnpb.MACCommand_DlChannelReq{
						DlChannelReq: &ttnpb.MACCommand_DLChannelReq{
							ChannelIndex: 1,
							Frequency:    2,
						},
					},
				},
				{
					Cid: ttnpb.MACCommandIdentifier_CID_LINK_ADR,
					Payload: &ttnpb.MACCommand_LinkAdrReq{
						LinkAdrReq: &ttnpb.MACCommand_LinkADRReq{
							DataRateIndex:      1,
							TxPowerIndex:       2,
							ChannelMask:        []bool{false, true, false},
							ChannelMaskControl: 3,
							NbTrans:            4,
						},
					},
				},
				{
					Cid: ttnpb.MACCommandIdentifier_CID_LINK_ADR,
					Payload: &ttnpb.MACCommand_LinkAdrReq{
						LinkAdrReq: &ttnpb.MACCommand_LinkADRReq{
							DataRateIndex:      1,
							TxPowerIndex:       2,
							ChannelMask:        []bool{false, true, false},
							ChannelMaskControl: 3,
							NbTrans:            4,
						},
					},
				},
				{
					Cid: ttnpb.MACCommandIdentifier_CID_DL_CHANNEL,
					Payload: &ttnpb.MACCommand_DlChannelReq{
						DlChannelReq: &ttnpb.MACCommand_DLChannelReq{
							ChannelIndex: 1,
							Frequency:    2,
						},
					},
				},
			},

			Result: []*ttnpb.MACCommand{
				{
					Cid: ttnpb.MACCommandIdentifier_CID_DL_CHANNEL,
					Payload: &ttnpb.MACCommand_DlChannelReq{
						DlChannelReq: &ttnpb.MACCommand_DLChannelReq{
							ChannelIndex: 1,
							Frequency:    2,
						},
					},
				},
				{
					Cid: ttnpb.MACCommandIdentifier_CID_DL_CHANNEL,
					Payload: &ttnpb.MACCommand_DlChannelReq{
						DlChannelReq: &ttnpb.MACCommand_DLChannelReq{
							ChannelIndex: 1,
							Frequency:    2,
						},
					},
				},
			},
		},
	} {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			a := assertions.New(t)
			rest, err := HandleMACResponseBlock(tc.CID, tc.AllowMissing, func(cmd *ttnpb.MACCommand) error {
				return tc.F(a, cmd)
			}, tc.Commands...)
			if tc.AssertError == nil {
				a.So(err, should.BeNil)
				a.So(rest, should.Resemble, tc.Result)
			} else {
				tc.AssertError(a, err)
			}
		})
	}
}
