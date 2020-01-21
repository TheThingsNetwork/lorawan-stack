// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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
	"context"
	"testing"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/band"
	"go.thethings.network/lorawan-stack/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/pkg/component/test"
	"go.thethings.network/lorawan-stack/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestNewDevAddr(t *testing.T) {
	a := assertions.New(t)

	// Use DevAddr prefix from NetID.
	{
		ns := test.Must(New(
			componenttest.NewComponent(t, &component.Config{}),
			&Config{
				NetID:               types.NetID{0x00, 0x00, 0x13},
				DeduplicationWindow: 42,
				CooldownWindow:      42,
				DownlinkTasks: &MockDownlinkTaskQueue{
					PopFunc: DownlinkTaskPopBlockFunc,
				},
			})).(*NetworkServer)

		if !a.So(ns.devAddrPrefixes, should.HaveLength, 1) {
			t.FailNow()
		}
		a.So(ns.devAddrPrefixes[0], should.Resemble, types.DevAddrPrefix{
			DevAddr: types.DevAddr{0x26, 0, 0, 0},
			Length:  7,
		})

		devAddr := ns.newDevAddr(test.Context(), nil)
		a.So(devAddr.HasPrefix(ns.devAddrPrefixes[0]), should.BeTrue)
	}

	// Configured DevAddr prefixes.
	{
		ns := test.Must(New(
			componenttest.NewComponent(t, &component.Config{}),
			&Config{
				NetID: types.NetID{0x00, 0x00, 0x13},
				DevAddrPrefixes: []types.DevAddrPrefix{
					{
						DevAddr: types.DevAddr{0x26, 0x01, 0x00, 0x00},
						Length:  16,
					},
					{
						DevAddr: types.DevAddr{0x26, 0xff, 0x01, 0x00},
						Length:  24,
					},
					{
						DevAddr: types.DevAddr{0x27, 0x00, 0x00, 0x00},
						Length:  8,
					},
				},
				DeduplicationWindow: 42,
				CooldownWindow:      42,
				DownlinkTasks: &MockDownlinkTaskQueue{
					PopFunc: DownlinkTaskPopBlockFunc,
				},
			})).(*NetworkServer)

		seen := map[types.DevAddrPrefix]int{}
		for i := 0; i < 100; i++ {
			devAddr := ns.newDevAddr(test.Context(), nil)
			for _, prefix := range ns.devAddrPrefixes {
				if devAddr.HasPrefix(prefix) {
					seen[prefix]++
					break
				}
			}
		}
		a.So(seen[ns.devAddrPrefixes[0]], should.BeGreaterThan, 0)
		a.So(seen[ns.devAddrPrefixes[1]], should.BeGreaterThan, 0)
		a.So(seen[ns.devAddrPrefixes[2]], should.BeGreaterThan, 0)
	}
}

func TestMatchAndHandleUplink(t *testing.T) {
	netID := test.Must(types.NewNetID(2, []byte{1, 2, 3})).(types.NetID)

	const appIDString = "match-and-handle-uplink-test-app-id"
	appID := ttnpb.ApplicationIdentifiers{ApplicationID: appIDString}
	const devID = "match-and-handle-uplink-test-dev-id"

	devAddr := types.DevAddr{0x42, 0x00, 0x00, 0x00}

	fNwkSIntKey := types.AES128Key{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	nwkSEncKey := types.AES128Key{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	sNwkSIntKey := types.AES128Key{0x42, 0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

	correlationIDs := [...]string{
		"match-and-handle-uplink-test-1",
		"match-and-handle-uplink-test-2",
	}

	start := time.Now().UTC()

	makeABPIdentifiers := func(devAddr types.DevAddr) *ttnpb.EndDeviceIdentifiers {
		return &ttnpb.EndDeviceIdentifiers{
			ApplicationIdentifiers: appID,
			DeviceID:               devID,

			DevAddr: &devAddr,
		}
	}

	makeSessionKeys := func(ver ttnpb.MACVersion) *ttnpb.SessionKeys {
		sk := &ttnpb.SessionKeys{
			FNwkSIntKey: &ttnpb.KeyEnvelope{
				Key: &fNwkSIntKey,
			},
			NwkSEncKey: &ttnpb.KeyEnvelope{
				Key: &nwkSEncKey,
			},
			SNwkSIntKey: &ttnpb.KeyEnvelope{
				Key: &sNwkSIntKey,
			},
			SessionKeyID: []byte("match-and-handle-uplink-test-session-key-id"),
		}
		if ver.Compare(ttnpb.MAC_V1_1) < 0 {
			sk.NwkSEncKey = sk.FNwkSIntKey
			sk.SNwkSIntKey = sk.FNwkSIntKey
		}
		return CopySessionKeys(sk)
	}

	makeSession := func(ver ttnpb.MACVersion, devAddr types.DevAddr, lastFCntUp uint32) *ttnpb.Session {
		return &ttnpb.Session{
			DevAddr:     devAddr,
			LastFCntUp:  lastFCntUp,
			SessionKeys: *makeSessionKeys(ver),
		}
	}

	makeUplink := func(pld *ttnpb.MACPayload, confirmed bool, fCnt, confFCnt uint32, txDRIdx ttnpb.DataRateIndex, txChIdx uint8, sets ttnpb.TxSettings) *ttnpb.UplinkMessage {
		mType := ttnpb.MType_UNCONFIRMED_UP
		if confirmed {
			mType = ttnpb.MType_CONFIRMED_UP
		}
		msg := ttnpb.Message{
			MHDR: ttnpb.MHDR{
				MType: mType,
				Major: ttnpb.Major_LORAWAN_R1,
			},
			Payload: &ttnpb.Message_MACPayload{
				MACPayload: pld,
			},
		}

		rawPayload := MustAppendUplinkMIC(sNwkSIntKey, fNwkSIntKey, confFCnt, uint8(txDRIdx), txChIdx, pld.DevAddr, fCnt, test.Must(lorawan.MarshalMessage(msg)).([]byte)...)
		msg.MIC = rawPayload[len(rawPayload)-4:]
		return &ttnpb.UplinkMessage{
			CorrelationIDs: correlationIDs[:],
			Payload:        &msg,
			RawPayload:     rawPayload,
			ReceivedAt:     start,
			RxMetadata:     MakeRxMetadataSlice(),
			Settings:       sets,
		}
	}

	makeLinkCheckEvents := func(pld *ttnpb.MACCommand_LinkCheckAns) []events.DefinitionDataClosure {
		return []events.DefinitionDataClosure{
			evtReceiveLinkCheckRequest.BindData(nil),
			evtEnqueueLinkCheckAnswer.BindData(pld),
		}
	}

	applicationDownlinkCorrelationIDs := [...]string{
		"application-downlink-correlation-id-1",
		"application-downlink-correlation-id-2",
	}

	makeApplicationDownlink := func() *ttnpb.ApplicationDownlink {
		return &ttnpb.ApplicationDownlink{
			SessionKeyID:   []byte("application-downlink-key"),
			FPort:          0x01,
			FCnt:           0x44,
			FRMPayload:     []byte("application-downlink-frm-payload"),
			CorrelationIDs: applicationDownlinkCorrelationIDs[:],
		}
	}

	makeDownlinkNack := func() *ttnpb.ApplicationUp {
		return &ttnpb.ApplicationUp{
			EndDeviceIdentifiers: *makeABPIdentifiers(devAddr),
			CorrelationIDs:       append(applicationDownlinkCorrelationIDs[:], correlationIDs[:]...),
			Up: &ttnpb.ApplicationUp_DownlinkNack{
				DownlinkNack: makeApplicationDownlink(),
			},
		}
	}

	makeDownlinkAck := func() *ttnpb.ApplicationUp {
		return &ttnpb.ApplicationUp{
			EndDeviceIdentifiers: *makeABPIdentifiers(devAddr),
			CorrelationIDs:       append(applicationDownlinkCorrelationIDs[:], correlationIDs[:]...),
			Up: &ttnpb.ApplicationUp_DownlinkAck{
				DownlinkAck: makeApplicationDownlink(),
			},
		}
	}

	for _, tc := range []struct {
		Name            string
		Uplink          *ttnpb.UplinkMessage
		Deduplicated    bool
		Devices         []*ttnpb.EndDevice
		DeviceAssertion func(ctx context.Context, dev *matchedDevice, up *ttnpb.UplinkMessage) bool
		ErrorAssertion  func(ctx context.Context, err error) bool
	}{
		{
			Name: "1.1/Does not support 32-bit FCnt/FCnt reset/No pending application downlink",
			Uplink: makeUplink(
				&ttnpb.MACPayload{
					FHDR: ttnpb.FHDR{
						DevAddr: devAddr,
						FCnt:    12,
						FOpts:   MustEncryptUplink(nwkSEncKey, devAddr, 12, 0x02),
					},
					FPort:      0x01,
					FRMPayload: []byte("test-frm-payload"),
				},
				false,
				12,
				0,
				ttnpb.DATA_RATE_2,
				1,
				ttnpb.TxSettings{
					DataRate: ttnpb.DataRate{
						Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{
							Bandwidth:       125000,
							SpreadingFactor: 10,
						}},
					},
					EnableCRC: true,
					Frequency: 868300000,
					Timestamp: 42,
				},
			),
			Devices: []*ttnpb.EndDevice{
				{
					EndDeviceIdentifiers: *makeABPIdentifiers(devAddr),
					FrequencyPlanID:      test.EUFrequencyPlanID,
					LoRaWANPHYVersion:    ttnpb.PHY_V1_1_REV_B,
					LoRaWANVersion:       ttnpb.MAC_V1_1,
					MACState:             MakeDefaultUS915FSB2MACState(ttnpb.CLASS_B, ttnpb.MAC_V1_1),
					Session:              makeSession(ttnpb.MAC_V1_1, devAddr, 33),
					MACSettings: &ttnpb.MACSettings{
						ResetsFCnt:        &pbtypes.BoolValue{Value: true},
						Supports32BitFCnt: &pbtypes.BoolValue{Value: false},
					},
				},
			},
			DeviceAssertion: func(ctx context.Context, dev *matchedDevice, up *ttnpb.UplinkMessage) bool {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(dev, should.NotBeNil) ||
					!a.So(dev.Device, should.NotBeNil) ||
					!a.So(dev.Device.Session, should.NotBeNil) {
					return false
				}
				session := makeSession(ttnpb.MAC_V1_1, devAddr, 12)
				session.StartedAt = dev.Device.Session.StartedAt
				macState := MakeDefaultEU868MACState(ttnpb.CLASS_A, ttnpb.MAC_V1_1)
				macState.RxWindowsAvailable = true
				expectedDev := &matchedDevice{
					logger:              dev.logger,
					phy:                 test.Must(band.All[band.EU_863_870].Version(ttnpb.PHY_V1_1_REV_B)).(band.Band),
					ChannelIndex:        1,
					DataRateIndex:       ttnpb.DATA_RATE_2,
					DeferredMACHandlers: dev.DeferredMACHandlers,
					Device: &ttnpb.EndDevice{
						EndDeviceIdentifiers: *makeABPIdentifiers(devAddr),
						FrequencyPlanID:      test.EUFrequencyPlanID,
						LoRaWANPHYVersion:    ttnpb.PHY_V1_1_REV_B,
						LoRaWANVersion:       ttnpb.MAC_V1_1,
						MACState:             macState,
						Session:              session,
						MACSettings: &ttnpb.MACSettings{
							ResetsFCnt:        &pbtypes.BoolValue{Value: true},
							Supports32BitFCnt: &pbtypes.BoolValue{Value: false},
						},
					},
					FCnt:      12,
					FCntReset: true,
					NbTrans:   1,
					SetPaths: []string{
						"mac_state",
						"pending_mac_state",
						"pending_session",
						"session",
					},
				}

				if !a.So([]time.Time{start, dev.Device.Session.StartedAt, time.Now()}, should.BeChronological) ||
					!a.So(dev.DeferredMACHandlers, should.HaveLength, 1) ||
					!a.So(dev, should.HaveEmptyDiff, expectedDev) {
					return false
				}

				linkCheckAns := MakeLinkCheckAns(MakeRxMetadataSlice()...)
				expectedEvents := map[int][]events.DefinitionDataClosure{
					0: makeLinkCheckEvents(linkCheckAns.GetLinkCheckAns()),
				}
				for i, h := range dev.DeferredMACHandlers {
					evs, err := h(ctx, dev.Device, up)
					if !a.So(err, should.BeNil) || !a.So(evs, should.ResembleEventDefinitionDataClosures, expectedEvents[i]) {
						return false
					}
				}
				expectedDev.Device.MACState.QueuedResponses = []*ttnpb.MACCommand{
					linkCheckAns,
				}
				return a.So(dev, should.HaveEmptyDiff, expectedDev)
			},
			ErrorAssertion: func(ctx context.Context, err error) bool {
				return assertions.New(test.MustTFromContext(ctx)).So(err, should.BeNil)
			},
		},

		{
			Name: "1.1/Does not support 32-bit FCnt/FCnt reset/Pending application downlink",
			Uplink: makeUplink(
				&ttnpb.MACPayload{
					FHDR: ttnpb.FHDR{
						DevAddr: devAddr,
						FCnt:    12,
						FOpts:   MustEncryptUplink(nwkSEncKey, devAddr, 12, 0x02),
					},
					FPort:      0x01,
					FRMPayload: []byte("test-frm-payload"),
				},
				false,
				12,
				0,
				ttnpb.DATA_RATE_2,
				1,
				ttnpb.TxSettings{
					DataRate: ttnpb.DataRate{
						Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{
							Bandwidth:       125000,
							SpreadingFactor: 10,
						}},
					},
					EnableCRC: true,
					Frequency: 868300000,
					Timestamp: 42,
				},
			),
			Devices: []*ttnpb.EndDevice{
				{
					EndDeviceIdentifiers: *makeABPIdentifiers(devAddr),
					FrequencyPlanID:      test.EUFrequencyPlanID,
					LoRaWANPHYVersion:    ttnpb.PHY_V1_1_REV_B,
					LoRaWANVersion:       ttnpb.MAC_V1_1,
					MACState: func() *ttnpb.MACState {
						macState := MakeDefaultUS915FSB2MACState(ttnpb.CLASS_B, ttnpb.MAC_V1_1)
						macState.PendingApplicationDownlink = makeApplicationDownlink()
						return macState
					}(),
					Session: makeSession(ttnpb.MAC_V1_1, devAddr, 33),
					MACSettings: &ttnpb.MACSettings{
						ResetsFCnt:        &pbtypes.BoolValue{Value: true},
						Supports32BitFCnt: &pbtypes.BoolValue{Value: false},
					},
				},
			},
			DeviceAssertion: func(ctx context.Context, dev *matchedDevice, up *ttnpb.UplinkMessage) bool {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(dev, should.NotBeNil) ||
					!a.So(dev.Device, should.NotBeNil) ||
					!a.So(dev.Device.Session, should.NotBeNil) {
					return false
				}
				session := makeSession(ttnpb.MAC_V1_1, devAddr, 12)
				session.StartedAt = dev.Device.Session.StartedAt
				macState := MakeDefaultEU868MACState(ttnpb.CLASS_A, ttnpb.MAC_V1_1)
				macState.RxWindowsAvailable = true
				expectedDev := &matchedDevice{
					logger:              dev.logger,
					phy:                 test.Must(band.All[band.EU_863_870].Version(ttnpb.PHY_V1_1_REV_B)).(band.Band),
					ChannelIndex:        1,
					DataRateIndex:       ttnpb.DATA_RATE_2,
					DeferredMACHandlers: dev.DeferredMACHandlers,
					Device: &ttnpb.EndDevice{
						EndDeviceIdentifiers: *makeABPIdentifiers(devAddr),
						FrequencyPlanID:      test.EUFrequencyPlanID,
						LoRaWANPHYVersion:    ttnpb.PHY_V1_1_REV_B,
						LoRaWANVersion:       ttnpb.MAC_V1_1,
						MACState:             macState,
						Session:              session,
						MACSettings: &ttnpb.MACSettings{
							ResetsFCnt:        &pbtypes.BoolValue{Value: true},
							Supports32BitFCnt: &pbtypes.BoolValue{Value: false},
						},
					},
					FCnt:      12,
					FCntReset: true,
					NbTrans:   1,
					SetPaths: []string{
						"mac_state",
						"pending_mac_state",
						"pending_session",
						"session",
					},
					QueuedApplicationUplinks: []*ttnpb.ApplicationUp{
						makeDownlinkNack(),
					},
				}
				if !a.So([]time.Time{start, dev.Device.Session.StartedAt, time.Now()}, should.BeChronological) ||
					!a.So(dev.DeferredMACHandlers, should.HaveLength, 1) ||
					!a.So(dev, should.HaveEmptyDiff, expectedDev) {
					return false
				}

				linkCheckAns := MakeLinkCheckAns(MakeRxMetadataSlice()...)
				expectedEvents := map[int][]events.DefinitionDataClosure{
					0: makeLinkCheckEvents(linkCheckAns.GetLinkCheckAns()),
				}
				for i, h := range dev.DeferredMACHandlers {
					evs, err := h(ctx, dev.Device, up)
					if !a.So(err, should.BeNil) || !a.So(evs, should.ResembleEventDefinitionDataClosures, expectedEvents[i]) {
						return false
					}
				}
				expectedDev.Device.MACState.QueuedResponses = []*ttnpb.MACCommand{
					linkCheckAns,
				}
				return a.So(dev, should.HaveEmptyDiff, expectedDev)
			},
			ErrorAssertion: func(ctx context.Context, err error) bool {
				return assertions.New(test.MustTFromContext(ctx)).So(err, should.BeNil)
			},
		},

		{
			Name: "1.1/Supports 32-bit FCnt/FCnt reset/No pending application downlink",
			Uplink: makeUplink(
				&ttnpb.MACPayload{
					FHDR: ttnpb.FHDR{
						DevAddr: devAddr,
						FCnt:    12,
						FOpts:   MustEncryptUplink(nwkSEncKey, devAddr, 12, 0x02),
					},
					FPort:      0x01,
					FRMPayload: []byte("test-frm-payload"),
				},
				false,
				12,
				0,
				ttnpb.DATA_RATE_2,
				1,
				ttnpb.TxSettings{
					DataRate: ttnpb.DataRate{
						Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{
							Bandwidth:       125000,
							SpreadingFactor: 10,
						}},
					},
					EnableCRC: true,
					Frequency: 868300000,
					Timestamp: 42,
				},
			),
			Devices: []*ttnpb.EndDevice{
				{
					EndDeviceIdentifiers: *makeABPIdentifiers(devAddr),
					FrequencyPlanID:      test.EUFrequencyPlanID,
					LoRaWANPHYVersion:    ttnpb.PHY_V1_1_REV_B,
					LoRaWANVersion:       ttnpb.MAC_V1_1,
					MACState:             MakeDefaultUS915FSB2MACState(ttnpb.CLASS_B, ttnpb.MAC_V1_1),
					Session:              makeSession(ttnpb.MAC_V1_1, devAddr, 33),
					MACSettings: &ttnpb.MACSettings{
						ResetsFCnt: &pbtypes.BoolValue{Value: true},
					},
				},
			},
			DeviceAssertion: func(ctx context.Context, dev *matchedDevice, up *ttnpb.UplinkMessage) bool {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(dev, should.NotBeNil) ||
					!a.So(dev.Device, should.NotBeNil) ||
					!a.So(dev.Device.Session, should.NotBeNil) {
					return false
				}
				session := makeSession(ttnpb.MAC_V1_1, devAddr, 12)
				session.StartedAt = dev.Device.Session.StartedAt
				macState := MakeDefaultEU868MACState(ttnpb.CLASS_A, ttnpb.MAC_V1_1)
				macState.RxWindowsAvailable = true
				expectedDev := &matchedDevice{
					logger:              dev.logger,
					phy:                 test.Must(band.All[band.EU_863_870].Version(ttnpb.PHY_V1_1_REV_B)).(band.Band),
					ChannelIndex:        1,
					DataRateIndex:       ttnpb.DATA_RATE_2,
					DeferredMACHandlers: dev.DeferredMACHandlers,
					Device: &ttnpb.EndDevice{
						EndDeviceIdentifiers: *makeABPIdentifiers(devAddr),
						FrequencyPlanID:      test.EUFrequencyPlanID,
						LoRaWANPHYVersion:    ttnpb.PHY_V1_1_REV_B,
						LoRaWANVersion:       ttnpb.MAC_V1_1,
						MACState:             macState,
						Session:              session,
						MACSettings: &ttnpb.MACSettings{
							ResetsFCnt: &pbtypes.BoolValue{Value: true},
						},
					},
					FCnt:      12,
					FCntReset: true,
					NbTrans:   1,
					SetPaths: []string{
						"mac_state",
						"pending_mac_state",
						"pending_session",
						"session",
					},
				}
				if !a.So([]time.Time{start, dev.Device.Session.StartedAt, time.Now()}, should.BeChronological) ||
					!a.So(dev.DeferredMACHandlers, should.HaveLength, 1) ||
					!a.So(dev, should.HaveEmptyDiff, expectedDev) {
					return false
				}

				linkCheckAns := MakeLinkCheckAns(MakeRxMetadataSlice()...)
				expectedEvents := map[int][]events.DefinitionDataClosure{
					0: makeLinkCheckEvents(linkCheckAns.GetLinkCheckAns()),
				}
				for i, h := range dev.DeferredMACHandlers {
					evs, err := h(ctx, dev.Device, up)
					if !a.So(err, should.BeNil) || !a.So(evs, should.ResembleEventDefinitionDataClosures, expectedEvents[i]) {
						return false
					}
				}
				expectedDev.Device.MACState.QueuedResponses = []*ttnpb.MACCommand{
					linkCheckAns,
				}
				return a.So(dev, should.HaveEmptyDiff, expectedDev)
			},
			ErrorAssertion: func(ctx context.Context, err error) bool {
				return assertions.New(test.MustTFromContext(ctx)).So(err, should.BeNil)
			},
		},

		{
			Name: "1.1/Ack",
			Uplink: makeUplink(
				&ttnpb.MACPayload{
					FHDR: ttnpb.FHDR{
						DevAddr: devAddr,
						FCnt:    12,
						FCtrl: ttnpb.FCtrl{
							Ack: true,
						},
						FOpts: MustEncryptUplink(nwkSEncKey, devAddr, 12, 0x02),
					},
					FPort:      0x01,
					FRMPayload: []byte("test-frm-payload"),
				},
				false,
				12,
				0,
				ttnpb.DATA_RATE_2,
				1,
				ttnpb.TxSettings{
					DataRate: ttnpb.DataRate{
						Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{
							Bandwidth:       125000,
							SpreadingFactor: 10,
						}},
					},
					EnableCRC: true,
					Frequency: 868300000,
					Timestamp: 42,
				},
			),
			Devices: []*ttnpb.EndDevice{
				{
					EndDeviceIdentifiers: *makeABPIdentifiers(devAddr),
					FrequencyPlanID:      test.EUFrequencyPlanID,
					LoRaWANPHYVersion:    ttnpb.PHY_V1_1_REV_B,
					LoRaWANVersion:       ttnpb.MAC_V1_1,
					MACState: func() *ttnpb.MACState {
						macState := MakeDefaultEU868MACState(ttnpb.CLASS_A, ttnpb.MAC_V1_1)
						macState.PendingApplicationDownlink = makeApplicationDownlink()
						macState.RecentDownlinks = []*ttnpb.DownlinkMessage{
							{
								Payload: &ttnpb.Message{
									MHDR: ttnpb.MHDR{
										MType: ttnpb.MType_CONFIRMED_DOWN,
									},
									Payload: &ttnpb.Message_MACPayload{
										MACPayload: &ttnpb.MACPayload{},
									},
								},
							},
						}
						return macState
					}(),
					Session: makeSession(ttnpb.MAC_V1_1, devAddr, 10),
				},
			},
			DeviceAssertion: func(ctx context.Context, dev *matchedDevice, up *ttnpb.UplinkMessage) bool {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(dev, should.NotBeNil) ||
					!a.So(dev.Device, should.NotBeNil) ||
					!a.So(dev.Device.Session, should.NotBeNil) {
					return false
				}
				macState := MakeDefaultEU868MACState(ttnpb.CLASS_A, ttnpb.MAC_V1_1)
				macState.RxWindowsAvailable = true
				macState.RecentDownlinks = []*ttnpb.DownlinkMessage{
					{
						Payload: &ttnpb.Message{
							MHDR: ttnpb.MHDR{
								MType: ttnpb.MType_CONFIRMED_DOWN,
							},
							Payload: &ttnpb.Message_MACPayload{
								MACPayload: &ttnpb.MACPayload{},
							},
						},
					},
				}
				expectedDev := &matchedDevice{
					logger:              dev.logger,
					phy:                 test.Must(band.All[band.EU_863_870].Version(ttnpb.PHY_V1_1_REV_B)).(band.Band),
					ChannelIndex:        1,
					DataRateIndex:       ttnpb.DATA_RATE_2,
					DeferredMACHandlers: dev.DeferredMACHandlers,
					Device: &ttnpb.EndDevice{
						EndDeviceIdentifiers: *makeABPIdentifiers(devAddr),
						FrequencyPlanID:      test.EUFrequencyPlanID,
						LoRaWANPHYVersion:    ttnpb.PHY_V1_1_REV_B,
						LoRaWANVersion:       ttnpb.MAC_V1_1,
						MACState:             macState,
						Session:              makeSession(ttnpb.MAC_V1_1, devAddr, 12),
					},
					FCnt:    12,
					NbTrans: 1,
					SetPaths: []string{
						"mac_state",
						"pending_mac_state",
						"pending_session",
						"session",
					},
					QueuedApplicationUplinks: []*ttnpb.ApplicationUp{
						makeDownlinkAck(),
					},
				}
				if !a.So(dev.DeferredMACHandlers, should.HaveLength, 1) ||
					!a.So(dev, should.HaveEmptyDiff, expectedDev) {
					return false
				}

				linkCheckAns := MakeLinkCheckAns(MakeRxMetadataSlice()...)
				expectedEvents := map[int][]events.DefinitionDataClosure{
					0: makeLinkCheckEvents(linkCheckAns.GetLinkCheckAns()),
				}
				for i, h := range dev.DeferredMACHandlers {
					evs, err := h(ctx, dev.Device, up)
					if !a.So(err, should.BeNil) || !a.So(evs, should.ResembleEventDefinitionDataClosures, expectedEvents[i]) {
						return false
					}
				}
				expectedDev.Device.MACState.QueuedResponses = []*ttnpb.MACCommand{
					linkCheckAns,
				}
				return a.So(dev, should.HaveEmptyDiff, expectedDev)
			},
			ErrorAssertion: func(ctx context.Context, err error) bool {
				return assertions.New(test.MustTFromContext(ctx)).So(err, should.BeNil)
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			ns, ctx, env, stop := StartTest(t, Config{NetID: netID}, (1<<10)*test.Delay, true)
			defer stop()

			<-env.DownlinkTasks.Pop

			dev, err := ns.matchAndHandleDataUplink(ctx, CopyUplinkMessage(tc.Uplink), tc.Deduplicated, CopyEndDevices(tc.Devices...)...)
			a.So(tc.DeviceAssertion(ctx, dev, CopyUplinkMessage(tc.Uplink)), should.BeTrue)
			a.So(tc.ErrorAssertion(ctx, err), should.BeTrue)
		})
	}
}
