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

package networkserver_test

import (
	"bytes"
	"context"
	"math/rand"
	"sync"
	"testing"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/events"
	. "go.thethings.network/lorawan-stack/pkg/networkserver"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
)

func sendUplinkDuplicates(ctx context.Context, handle func(ctx context.Context, up *ttnpb.UplinkMessage) <-chan error, windowEndCh <-chan WindowEndRequest, makeMessage func(decoded bool) *ttnpb.UplinkMessage, start time.Time, n int) []*ttnpb.RxMetadata {
	t := test.MustTFromContext(ctx)
	t.Helper()

	a := assertions.New(t)

	var weResp chan<- time.Time
	select {
	case <-ctx.Done():
		t.Fatal("Timed out while waiting for window end request to arrive")
		return nil

	case req := <-windowEndCh:
		expectedMsg := makeMessage(true)
		for _, id := range expectedMsg.CorrelationIDs {
			a.So(req.Message.CorrelationIDs, should.Contain, id)
		}
		a.So(len(req.Message.CorrelationIDs), should.BeGreaterThan, len(expectedMsg.CorrelationIDs))
		for _, md := range expectedMsg.RxMetadata {
			a.So(req.Message.RxMetadata, should.Contain, md)
		}
		a.So(len(req.Message.RxMetadata), should.BeGreaterThanOrEqualTo, len(expectedMsg.RxMetadata))
		a.So([]time.Time{start, req.Message.ReceivedAt, time.Now()}, should.BeChronological)
		expectedMsg.ReceivedAt = req.Message.ReceivedAt
		expectedMsg.CorrelationIDs = req.Message.CorrelationIDs
		expectedMsg.RxMetadata = req.Message.RxMetadata
		a.So(req.Message, should.HaveEmptyDiff, expectedMsg)

		weResp = req.Response
	}

	mdCh := make(chan *ttnpb.RxMetadata, n)
	if !t.Run("duplicates", func(t *testing.T) {
		a := assertions.New(t)

		wg := &sync.WaitGroup{}
		wg.Add(n)

		for i := 0; i < n; i++ {
			go func() {
				defer wg.Done()

				msg := makeMessage(false)

				msg.RxMetadata = nil
				n := 1 + rand.Intn(10)
				for i := 0; i < n; i++ {
					md := ttnpb.NewPopulatedRxMetadata(test.Randy, false)
					msg.RxMetadata = append(msg.RxMetadata, md)
					mdCh <- md
				}

				err := <-handle(ctx, msg)
				a.So(err, should.BeNil)
			}()
		}

		go func() {
			if !test.WaitContext(ctx, wg.Wait) {
				t.Log("Timed out while waiting for duplicate uplinks to be processed")
				return
			}

			select {
			case <-ctx.Done():
				t.Log("Timed out while waiting for metadata collection to stop")
				return

			case weResp <- time.Now():
			}

			close(mdCh)
		}()
	}) {
		t.Error("Failed to send duplicates")
		return nil
	}

	var mds []*ttnpb.RxMetadata
	for md := range mdCh {
		mds = append(mds, md)
	}
	return mds
}

func TestHandleUplink(t *testing.T) {
	dataGetPaths := [...]string{
		"frequency_plan_id",
		"last_dev_status_received_at",
		"lorawan_phy_version",
		"lorawan_version",
		"mac_settings",
		"mac_state",
		"multicast",
		"pending_mac_state",
		"pending_session",
		"recent_downlinks",
		"recent_uplinks",
		"session",
		"supports_class_b",
		"supports_class_c",
		"supports_join",
	}

	joinGetByEUIPaths := [...]string{
		"frequency_plan_id",
		"lorawan_phy_version",
		"lorawan_version",
		"mac_settings",
		"session",
		"supports_class_b",
		"supports_class_c",
		"supports_join",
	}

	joinSetByEUIGetPaths := [...]string{
		"frequency_plan_id",
		"lorawan_phy_version",
		"queued_application_downlinks",
		"recent_uplinks",
	}

	joinSetByEUISetPaths := [...]string{
		"pending_mac_state",
		"queued_application_downlinks",
		"recent_uplinks",
	}

	const duplicateCount = 6
	const fPort = 0x42

	netID := test.Must(types.NewNetID(2, []byte{1, 2, 3})).(types.NetID)

	const appIDString = "handle-uplink-test-app-id"
	appID := ttnpb.ApplicationIdentifiers{ApplicationID: appIDString}
	const devID = "handle-uplink-test-dev-id"

	joinEUI := types.EUI64{0x42, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	devEUI := types.EUI64{0x42, 0x42, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

	devAddr := types.DevAddr{0x42, 0x00, 0x00, 0x00}

	fNwkSIntKey := types.AES128Key{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	nwkSEncKey := types.AES128Key{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	sNwkSIntKey := types.AES128Key{0x42, 0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	appSKey := types.AES128Key{0x42, 0x42, 0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

	correlationIDs := [...]string{
		"handle-uplink-test-1",
		"handle-uplink-test-2",
	}

	now := time.Now().UTC()

	makeOTAAIdentifiers := func(devAddr *types.DevAddr) *ttnpb.EndDeviceIdentifiers {
		return &ttnpb.EndDeviceIdentifiers{
			ApplicationIdentifiers: appID,
			DeviceID:               devID,

			DevEUI:  devEUI.Copy(&types.EUI64{}),
			JoinEUI: joinEUI.Copy(&types.EUI64{}),

			DevAddr: devAddr,
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
			SessionKeyID: []byte("handle-uplink-test-session-key-id"),
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
			StartedAt:   now,
		}
	}

	makeJoinRequest := func(decodePayload bool) *ttnpb.UplinkMessage {
		msg := &ttnpb.UplinkMessage{
			CorrelationIDs: correlationIDs[:],
			RawPayload: []byte{
				/* MHDR */
				0x00,
				/* Join-request */
				/** JoinEUI **/
				joinEUI[7], joinEUI[6], joinEUI[5], joinEUI[4], joinEUI[3], joinEUI[2], joinEUI[1], joinEUI[0],
				/** DevEUI **/
				devEUI[7], devEUI[6], devEUI[5], devEUI[4], devEUI[3], devEUI[2], devEUI[1], devEUI[0],
				/** DevNonce **/
				0x01, 0x00,
				/* MIC */
				0x03, 0x02, 0x01, 0x00,
			},
			RxMetadata: MakeRxMetadataSlice(),
			Settings: ttnpb.TxSettings{
				DataRate: ttnpb.DataRate{
					Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{
						Bandwidth:       125000,
						SpreadingFactor: 11,
					}},
				},
				EnableCRC: true,
				Frequency: 868500000,
				Timestamp: 42,
			},
		}
		if decodePayload {
			msg.Payload = &ttnpb.Message{
				MHDR: ttnpb.MHDR{
					MType: ttnpb.MType_JOIN_REQUEST,
					Major: ttnpb.Major_LORAWAN_R1,
				},
				MIC: []byte{0x03, 0x02, 0x01, 0x00},
				Payload: &ttnpb.Message_JoinRequestPayload{
					JoinRequestPayload: &ttnpb.JoinRequestPayload{
						DevEUI:   devEUI,
						JoinEUI:  joinEUI,
						DevNonce: types.DevNonce{0x00, 0x01},
					},
				},
			}
		}
		return msg
	}

	makeRejoinRequest := func(decodePayload bool) *ttnpb.UplinkMessage {
		msg := &ttnpb.UplinkMessage{
			CorrelationIDs: correlationIDs[:],
			RawPayload: []byte{
				/* MHDR */
				0xc0,
				/* Rejoin-request */
				/** Rejoin Type **/
				0x00,
				/** JoinEUI **/
				netID[2], netID[1], netID[0],
				/** DevEUI **/
				devEUI[7], devEUI[6], devEUI[5], devEUI[4], devEUI[3], devEUI[2], devEUI[1], devEUI[0],
				/** RJcount0 **/
				0x01, 0x00,
				/* MIC */
				0x03, 0x02, 0x01, 0x00,
			},
			RxMetadata: MakeRxMetadataSlice(),
			Settings: ttnpb.TxSettings{
				DataRate: ttnpb.DataRate{
					Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{
						Bandwidth:       125000,
						SpreadingFactor: 11,
					}},
				},
				EnableCRC: true,
				Frequency: 868500000,
				Timestamp: 42,
			},
		}
		if decodePayload {
			msg.Payload = &ttnpb.Message{
				MHDR: ttnpb.MHDR{
					MType: ttnpb.MType_REJOIN_REQUEST,
					Major: ttnpb.Major_LORAWAN_R1,
				},
				MIC: []byte{0x03, 0x02, 0x01, 0x00},
				Payload: &ttnpb.Message_RejoinRequestPayload{
					RejoinRequestPayload: &ttnpb.RejoinRequestPayload{
						DevEUI:     devEUI,
						NetID:      netID,
						RejoinCnt:  1,
						RejoinType: ttnpb.RejoinType_CONTEXT,
					},
				},
			}
		}
		return msg
	}

	makeDataUplinkFRMPayload := func(fCnt uint32) []byte {
		return MustEncryptUplink(appSKey, devAddr, fCnt, 't', 'e', 's', 't')
	}

	makeDataUplinkSettings := func() ttnpb.TxSettings {
		return ttnpb.TxSettings{
			DataRate: ttnpb.DataRate{
				Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{
					Bandwidth:       125000,
					SpreadingFactor: 10,
				}},
			},
			EnableCRC: true,
			Frequency: 868300000,
			Timestamp: 42,
		}
	}

	makeDecodedDataUplinkPayload := func(fCnt uint32, fOpts []byte, payload ...byte) *ttnpb.Message {
		return &ttnpb.Message{
			MHDR: ttnpb.MHDR{
				MType: ttnpb.MType_UNCONFIRMED_UP,
				Major: ttnpb.Major_LORAWAN_R1,
			},
			MIC: payload[len(payload)-4:],
			Payload: &ttnpb.Message_MACPayload{
				MACPayload: &ttnpb.MACPayload{
					FHDR: ttnpb.FHDR{
						DevAddr: *devAddr.Copy(&types.DevAddr{}),
						FCnt:    fCnt,
						FOpts:   fOpts,
					},
					FPort:      fPort,
					FRMPayload: makeDataUplinkFRMPayload(fCnt),
				},
			},
		}
	}

	makeLegacyDataUplink := func(fCnt uint8, decodePayload bool) *ttnpb.UplinkMessage {
		msg := &ttnpb.UplinkMessage{
			CorrelationIDs: correlationIDs[:],
			RawPayload: MustAppendLegacyUplinkMIC(
				fNwkSIntKey,
				devAddr,
				uint32(fCnt),
				append([]byte{
					/* MHDR */
					0x40,
					/* MACPayload */
					/** FHDR **/
					/*** DevAddr ***/
					devAddr[3], devAddr[2], devAddr[1], devAddr[0],
					/*** FCtrl ***/
					0x01,
					/*** FCnt ***/
					fCnt, 0x00,
					/*** FOpts ***/
					/**** LinkCheckReq ****/
					0x02,
					/** FPort **/
					fPort,
				},
					makeDataUplinkFRMPayload(uint32(fCnt))...,
				)...,
			),
			RxMetadata: MakeRxMetadataSlice(),
			Settings:   makeDataUplinkSettings(),
		}
		if decodePayload {
			msg.Payload = makeDecodedDataUplinkPayload(uint32(fCnt), []byte{0x02}, msg.RawPayload...)
			msg.ReceivedAt = now
		}
		return msg
	}

	bindMakeLegacyDataUplinkFCnt := func(fCnt uint8) func(bool) *ttnpb.UplinkMessage {
		return func(decoded bool) *ttnpb.UplinkMessage {
			return makeLegacyDataUplink(fCnt, decoded)
		}
	}

	makeDataUplink := func(fCnt uint8, decodePayload bool) *ttnpb.UplinkMessage {
		sets := makeDataUplinkSettings()
		mds := MakeRxMetadataSlice()
		fOpts := MustEncryptUplink(nwkSEncKey, devAddr, uint32(fCnt), 0x02)
		msg := &ttnpb.UplinkMessage{
			CorrelationIDs: correlationIDs[:],
			RawPayload: MustAppendUplinkMIC(
				sNwkSIntKey,
				fNwkSIntKey,
				0,
				2,
				1,
				devAddr,
				uint32(fCnt),
				append(
					append(
						append([]byte{
							/* MHDR */
							0x40,
							/* MACPayload */
							/** FHDR **/
							/*** DevAddr ***/
							devAddr[3], devAddr[2], devAddr[1], devAddr[0],
							/*** FCtrl ***/
							0x01,
							/*** FCnt ***/
							fCnt, 0x00,
						},
							/*** FOpts ***/
							fOpts...,
						),
						/** FPort **/
						fPort,
					),
					makeDataUplinkFRMPayload(uint32(fCnt))...,
				)...,
			),
			RxMetadata: mds,
			Settings:   sets,
		}
		if decodePayload {
			msg.Payload = makeDecodedDataUplinkPayload(uint32(fCnt), fOpts, msg.RawPayload...)
			msg.ReceivedAt = now
		}
		return msg
	}

	bindMakeDataUplinkFCnt := func(fCnt uint8) func(bool) *ttnpb.UplinkMessage {
		return func(decoded bool) *ttnpb.UplinkMessage {
			return makeDataUplink(fCnt, decoded)
		}
	}

	makeApplicationDownlink := func() *ttnpb.ApplicationDownlink {
		return &ttnpb.ApplicationDownlink{
			SessionKeyID: []byte("app-down-1-session-key-id"),
			FPort:        fPort,
			FCnt:         0x32,
			FRMPayload:   []byte("app-down-1-frm-payload"),
			Confirmed:    true,
			Priority:     ttnpb.TxSchedulePriority_HIGH,
			CorrelationIDs: []string{
				"app-down-1-correlation-id-1",
			},
		}
	}

	makeJoinResponse := func(ver ttnpb.MACVersion) *ttnpb.JoinResponse {
		return &ttnpb.JoinResponse{
			RawPayload:  bytes.Repeat([]byte{0x42}, 17),
			SessionKeys: *makeSessionKeys(ver),
		}
	}

	assertHandleUplinkResponse := func(ctx context.Context, handleUplinkErrCh <-chan error, assert func(error) bool) bool {
		t := test.MustTFromContext(ctx)
		t.Helper()
		select {
		case <-ctx.Done():
			t.Error("Timed out while waiting for NetworkServer.HandleUplink to return")
			return false

		case err := <-handleUplinkErrCh:
			return assert(err)
		}
	}

	type AsNsLinkRecvRequest struct {
		Uplink   *ttnpb.ApplicationUp
		Response chan<- *pbtypes.Empty
	}

	var errTest = errors.New("testError")

	for _, tc := range []struct {
		Name    string
		Handler func(context.Context, TestEnvironment, <-chan AsNsLinkRecvRequest, func(context.Context, *ttnpb.UplinkMessage) <-chan error) bool
	}{
		{
			Name: "Invalid payload",
			Handler: func(ctx context.Context, env TestEnvironment, asRecvCh <-chan AsNsLinkRecvRequest, handle func(context.Context, *ttnpb.UplinkMessage) <-chan error) bool {
				t := test.MustTFromContext(ctx)
				a := assertions.New(t)

				makeMsg := func(_ bool) *ttnpb.UplinkMessage {
					return &ttnpb.UplinkMessage{
						RxMetadata: MakeRxMetadataSlice(),
					}
				}

				handleUplinkErrCh := handle(ctx, makeMsg(false))

				return assertHandleUplinkResponse(ctx, handleUplinkErrCh, func(err error) bool {
					return a.So(err, should.HaveSameErrorDefinitionAs, ErrDecodePayload)
				})
			},
		},

		{
			Name: "Unknown Major",
			Handler: func(ctx context.Context, env TestEnvironment, asRecvCh <-chan AsNsLinkRecvRequest, handle func(context.Context, *ttnpb.UplinkMessage) <-chan error) bool {
				t := test.MustTFromContext(ctx)
				a := assertions.New(t)

				makeMsg := func(_ bool) *ttnpb.UplinkMessage {
					return &ttnpb.UplinkMessage{
						RawPayload: []byte{
							/* MHDR */
							0x01,
							/* Join-request */
							/** JoinEUI **/
							joinEUI[7], joinEUI[6], joinEUI[5], joinEUI[4], joinEUI[3], joinEUI[2], joinEUI[1], joinEUI[0],
							/** DevEUI **/
							devEUI[7], devEUI[6], devEUI[5], devEUI[4], devEUI[3], devEUI[2], devEUI[1], devEUI[0],
							/** DevNonce **/
							0x01, 0x00,
							/* MIC */
							0x03, 0x02, 0x01, 0x00,
						},
						RxMetadata: MakeRxMetadataSlice(),
					}
				}

				handleUplinkErrCh := handle(ctx, makeMsg(false))

				return assertHandleUplinkResponse(ctx, handleUplinkErrCh, func(err error) bool {
					return a.So(err, should.HaveSameErrorDefinitionAs, ErrUnsupportedLoRaWANVersion)
				})
			},
		},

		{
			Name: "Invalid MType",
			Handler: func(ctx context.Context, env TestEnvironment, asRecvCh <-chan AsNsLinkRecvRequest, handle func(context.Context, *ttnpb.UplinkMessage) <-chan error) bool {
				t := test.MustTFromContext(ctx)
				a := assertions.New(t)

				makeMsg := func(_ bool) *ttnpb.UplinkMessage {
					return &ttnpb.UplinkMessage{
						RawPayload: bytes.Repeat([]byte{0x20}, 33),
						RxMetadata: MakeRxMetadataSlice(),
					}
				}

				handleUplinkErrCh := handle(ctx, makeMsg(false))

				return assertHandleUplinkResponse(ctx, handleUplinkErrCh, func(err error) bool {
					return a.So(err, should.BeNil)
				})
			},
		},

		{
			Name: "Proprietary MType",
			Handler: func(ctx context.Context, env TestEnvironment, asRecvCh <-chan AsNsLinkRecvRequest, handle func(context.Context, *ttnpb.UplinkMessage) <-chan error) bool {
				t := test.MustTFromContext(ctx)
				a := assertions.New(t)

				makeMsg := func(_ bool) *ttnpb.UplinkMessage {
					return &ttnpb.UplinkMessage{
						RawPayload: []byte{
							/* MHDR */
							0xe0,
						},
						RxMetadata: MakeRxMetadataSlice(),
					}
				}

				handleUplinkErrCh := handle(ctx, makeMsg(false))

				return assertHandleUplinkResponse(ctx, handleUplinkErrCh, func(err error) bool {
					return a.So(err, should.HaveSameErrorDefinitionAs, ErrDecodePayload)
				})
			},
		},

		{
			Name: "Join-request/Get fail",
			Handler: func(ctx context.Context, env TestEnvironment, asRecvCh <-chan AsNsLinkRecvRequest, handle func(context.Context, *ttnpb.UplinkMessage) <-chan error) bool {
				t := test.MustTFromContext(ctx)
				a := assertions.New(t)

				start := time.Now()

				msg := makeJoinRequest(false)

				handleUplinkErrCh := handle(ctx, msg)

				select {
				case <-ctx.Done():
					t.Error("Timed out while waiting for DeviceRegistry.GetByEUI to be called")
					return false

				case req := <-env.DeviceRegistry.GetByEUI:
					reqCorrelationIDs := events.CorrelationIDsFromContext(req.Context)
					for _, id := range correlationIDs {
						a.So(reqCorrelationIDs, should.Contain, id)
					}
					a.So(reqCorrelationIDs, should.HaveLength, len(correlationIDs)+2)
					a.So(req.JoinEUI, should.Resemble, joinEUI)
					a.So(req.DevEUI, should.Resemble, devEUI)
					a.So(req.Paths, should.HaveSameElementsDeep, joinGetByEUIPaths[:])
					req.Response <- DeviceRegistryGetByEUIResponse{
						Error: errTest,
					}
				}

				_ = sendUplinkDuplicates(ctx, handle, env.CollectionDone, makeJoinRequest, start, duplicateCount)

				return assertHandleUplinkResponse(ctx, handleUplinkErrCh, func(err error) bool {
					return a.So(err, should.EqualErrorOrDefinition, errTest)
				})
			},
		},

		{
			Name: "Join-request/Get ABP device",
			Handler: func(ctx context.Context, env TestEnvironment, asRecvCh <-chan AsNsLinkRecvRequest, handle func(context.Context, *ttnpb.UplinkMessage) <-chan error) bool {
				t := test.MustTFromContext(ctx)
				a := assertions.New(t)

				start := time.Now()

				msg := makeJoinRequest(false)

				handleUplinkErrCh := handle(ctx, msg)

				getDevice := &ttnpb.EndDevice{
					EndDeviceIdentifiers: *makeOTAAIdentifiers(nil),
					FrequencyPlanID:      test.EUFrequencyPlanID,
					LoRaWANPHYVersion:    ttnpb.PHY_V1_0,
					LoRaWANVersion:       ttnpb.MAC_V1_0,
					CreatedAt:            start,
					UpdatedAt:            time.Now(),
				}

				var reqCtx context.Context
				select {
				case <-ctx.Done():
					t.Error("Timed out while waiting for DeviceRegistry.GetByEUI to be called")
					return false

				case req := <-env.DeviceRegistry.GetByEUI:
					reqCtx = req.Context
					reqCorrelationIDs := events.CorrelationIDsFromContext(req.Context)
					for _, id := range correlationIDs {
						a.So(reqCorrelationIDs, should.Contain, id)
					}
					a.So(reqCorrelationIDs, should.HaveLength, len(correlationIDs)+2)
					a.So(req.JoinEUI, should.Resemble, joinEUI)
					a.So(req.DevEUI, should.Resemble, devEUI)
					a.So(req.Paths, should.HaveSameElementsDeep, joinGetByEUIPaths[:])
					req.Response <- DeviceRegistryGetByEUIResponse{
						Device: CopyEndDevice(getDevice),
					}
				}

				if !a.So(test.AssertEventPubSubPublishRequest(ctx, env.Events, func(ev events.Event) bool {
					return a.So(ev, should.ResembleEvent, EvtDropJoinRequest(reqCtx, makeOTAAIdentifiers(nil), ErrABPJoinRequest))
				}), should.BeTrue) {
					return false
				}

				_ = sendUplinkDuplicates(ctx, handle, env.CollectionDone, makeJoinRequest, start, duplicateCount)

				return assertHandleUplinkResponse(ctx, handleUplinkErrCh, func(err error) bool {
					return a.So(err, should.EqualErrorOrDefinition, ErrABPJoinRequest)
				})
			},
		},

		{
			Name: "Join-request/Get multicast device",
			Handler: func(ctx context.Context, env TestEnvironment, asRecvCh <-chan AsNsLinkRecvRequest, handle func(context.Context, *ttnpb.UplinkMessage) <-chan error) bool {
				t := test.MustTFromContext(ctx)
				a := assertions.New(t)

				start := time.Now()

				msg := makeJoinRequest(false)

				handleUplinkErrCh := handle(ctx, msg)

				getDevice := &ttnpb.EndDevice{
					EndDeviceIdentifiers: *makeOTAAIdentifiers(nil),
					FrequencyPlanID:      test.EUFrequencyPlanID,
					LoRaWANPHYVersion:    ttnpb.PHY_V1_0,
					LoRaWANVersion:       ttnpb.MAC_V1_0,
					Multicast:            true,
					CreatedAt:            start,
					UpdatedAt:            time.Now(),
				}

				var reqCtx context.Context
				select {
				case <-ctx.Done():
					t.Error("Timed out while waiting for DeviceRegistry.GetByEUI to be called")
					return false

				case req := <-env.DeviceRegistry.GetByEUI:
					reqCtx = req.Context
					reqCorrelationIDs := events.CorrelationIDsFromContext(req.Context)
					for _, id := range correlationIDs {
						a.So(reqCorrelationIDs, should.Contain, id)
					}
					a.So(reqCorrelationIDs, should.HaveLength, len(correlationIDs)+2)
					a.So(req.JoinEUI, should.Resemble, joinEUI)
					a.So(req.DevEUI, should.Resemble, devEUI)
					a.So(req.Paths, should.HaveSameElementsDeep, joinGetByEUIPaths[:])
					req.Response <- DeviceRegistryGetByEUIResponse{
						Device: CopyEndDevice(getDevice),
					}
				}

				if !a.So(test.AssertEventPubSubPublishRequest(ctx, env.Events, func(ev events.Event) bool {
					return a.So(ev, should.ResembleEvent, EvtDropJoinRequest(reqCtx, makeOTAAIdentifiers(nil), ErrABPJoinRequest))
				}), should.BeTrue) {
					return false
				}

				_ = sendUplinkDuplicates(ctx, handle, env.CollectionDone, makeJoinRequest, start, duplicateCount)

				return assertHandleUplinkResponse(ctx, handleUplinkErrCh, func(err error) bool {
					return a.So(err, should.EqualErrorOrDefinition, ErrABPJoinRequest)
				})
			},
		},

		{
			Name: "Join-request/Get OTAA device/1.0.2/JS fail",
			Handler: func(ctx context.Context, env TestEnvironment, asRecvCh <-chan AsNsLinkRecvRequest, handle func(context.Context, *ttnpb.UplinkMessage) <-chan error) bool {
				t := test.MustTFromContext(ctx)
				a := assertions.New(t)

				start := time.Now()

				msg := makeJoinRequest(false)

				handleUplinkErrCh := handle(ctx, msg)

				getDevice := &ttnpb.EndDevice{
					EndDeviceIdentifiers: *makeOTAAIdentifiers(nil),
					FrequencyPlanID:      test.EUFrequencyPlanID,
					LoRaWANPHYVersion:    ttnpb.PHY_V1_0_2_REV_B,
					LoRaWANVersion:       ttnpb.MAC_V1_0_2,
					MACSettings: &ttnpb.MACSettings{
						Rx1Delay: &ttnpb.MACSettings_RxDelayValue{
							Value: ttnpb.RX_DELAY_3,
						},
					},
					SupportsJoin: true,
					CreatedAt:    start,
					UpdatedAt:    time.Now(),
				}

				var reqCtx context.Context
				var reqCorrelationIDs []string
				select {
				case <-ctx.Done():
					t.Error("Timed out while waiting for DeviceRegistry.GetByEUI to be called")
					return false

				case req := <-env.DeviceRegistry.GetByEUI:
					reqCtx = req.Context
					reqCorrelationIDs = events.CorrelationIDsFromContext(req.Context)
					for _, id := range correlationIDs {
						a.So(reqCorrelationIDs, should.Contain, id)
					}
					a.So(reqCorrelationIDs, should.HaveLength, len(correlationIDs)+2)
					a.So(req.JoinEUI, should.Resemble, joinEUI)
					a.So(req.DevEUI, should.Resemble, devEUI)
					a.So(req.Paths, should.HaveSameElementsDeep, joinGetByEUIPaths[:])
					req.Response <- DeviceRegistryGetByEUIResponse{
						Device: CopyEndDevice(getDevice),
					}
				}

				if !a.So(AssertNsJsPeerHandleAuthJoinRequest(ctx, env.Cluster.GetPeer, env.Cluster.Auth,
					func(ctx context.Context, ids ttnpb.Identifiers) bool {
						return a.So(ctx, should.HaveParentContextOrEqual, reqCtx) &&
							a.So(ids, should.Resemble, *makeOTAAIdentifiers(nil))
					},
					func(ctx context.Context, req *ttnpb.JoinRequest) bool {
						return a.So(req.CorrelationIDs, should.HaveSameElementsDeep, reqCorrelationIDs) &&
							a.So(req.DevAddr, should.NotBeEmpty) &&
							a.So(req.DevAddr.NwkID(), should.Resemble, netID.ID()) &&
							a.So(req.DevAddr.NetIDType(), should.Equal, netID.Type()) &&
							a.So(req, should.Resemble, &ttnpb.JoinRequest{
								CFList: &ttnpb.CFList{
									Type: ttnpb.CFListType_FREQUENCIES,
									Freq: []uint32{8671000, 8673000, 8675000, 8677000, 8679000},
								},
								CorrelationIDs:     req.CorrelationIDs,
								DevAddr:            req.DevAddr,
								NetID:              netID,
								RawPayload:         msg.RawPayload,
								RxDelay:            ttnpb.RX_DELAY_3,
								SelectedMACVersion: ttnpb.MAC_V1_0_2,
							})
					},
					&grpc.EmptyCallOption{},
					NsJsHandleJoinResponse{
						Error: errTest,
					},
				), should.BeTrue) {
					return false
				}

				if !a.So(test.AssertEventPubSubPublishRequest(ctx, env.Events, func(ev events.Event) bool {
					if !a.So(ev.Data(), should.BeError) {
						return false
					}
					err, ok := errors.From(ev.Data().(error))
					if !a.So(ok, should.BeTrue) {
						return false
					}
					return a.So(ev, should.ResembleEvent, EvtDropJoinRequest(reqCtx, makeOTAAIdentifiers(nil), err))
				}), should.BeTrue) {
					return false
				}

				_ = sendUplinkDuplicates(ctx, handle, env.CollectionDone, makeJoinRequest, start, duplicateCount)

				return assertHandleUplinkResponse(ctx, handleUplinkErrCh, func(err error) bool {
					return a.So(err, should.BeError)
				})
			},
		},

		{
			Name: "Join-request/Get OTAA device/1.1/JS fail",
			Handler: func(ctx context.Context, env TestEnvironment, asRecvCh <-chan AsNsLinkRecvRequest, handle func(context.Context, *ttnpb.UplinkMessage) <-chan error) bool {
				t := test.MustTFromContext(ctx)
				a := assertions.New(t)

				start := time.Now()

				msg := makeJoinRequest(false)

				handleUplinkErrCh := handle(ctx, msg)

				getDevice := &ttnpb.EndDevice{
					EndDeviceIdentifiers: *makeOTAAIdentifiers(nil),
					FrequencyPlanID:      test.EUFrequencyPlanID,
					LoRaWANPHYVersion:    ttnpb.PHY_V1_1_REV_B,
					LoRaWANVersion:       ttnpb.MAC_V1_1,
					MACSettings: &ttnpb.MACSettings{
						Rx1Delay: &ttnpb.MACSettings_RxDelayValue{
							Value: ttnpb.RX_DELAY_3,
						},
					},
					SupportsJoin: true,
					CreatedAt:    start,
					UpdatedAt:    time.Now(),
				}

				var reqCtx context.Context
				var reqCorrelationIDs []string
				select {
				case <-ctx.Done():
					t.Error("Timed out while waiting for DeviceRegistry.GetByEUI to be called")
					return false

				case req := <-env.DeviceRegistry.GetByEUI:
					reqCtx = req.Context
					reqCorrelationIDs = events.CorrelationIDsFromContext(req.Context)
					for _, id := range correlationIDs {
						a.So(reqCorrelationIDs, should.Contain, id)
					}
					a.So(reqCorrelationIDs, should.HaveLength, len(correlationIDs)+2)
					a.So(req.JoinEUI, should.Resemble, joinEUI)
					a.So(req.DevEUI, should.Resemble, devEUI)
					a.So(req.Paths, should.HaveSameElementsDeep, joinGetByEUIPaths[:])
					req.Response <- DeviceRegistryGetByEUIResponse{
						Device: CopyEndDevice(getDevice),
					}
				}

				if !a.So(AssertNsJsPeerHandleAuthJoinRequest(ctx, env.Cluster.GetPeer, env.Cluster.Auth,
					func(ctx context.Context, ids ttnpb.Identifiers) bool {
						return a.So(ids, should.Resemble, *makeOTAAIdentifiers(nil))
					},
					func(ctx context.Context, req *ttnpb.JoinRequest) bool {
						return a.So(req.CorrelationIDs, should.HaveSameElementsDeep, reqCorrelationIDs) &&
							a.So(req.DevAddr, should.NotBeEmpty) &&
							a.So(req.DevAddr.NwkID(), should.Resemble, netID.ID()) &&
							a.So(req.DevAddr.NetIDType(), should.Equal, netID.Type()) &&
							a.So(req, should.Resemble, &ttnpb.JoinRequest{
								CFList: &ttnpb.CFList{
									Type: ttnpb.CFListType_FREQUENCIES,
									Freq: []uint32{8671000, 8673000, 8675000, 8677000, 8679000},
								},
								CorrelationIDs: req.CorrelationIDs,
								DevAddr:        req.DevAddr,
								DownlinkSettings: ttnpb.DLSettings{
									OptNeg: true,
								},
								NetID:              netID,
								RawPayload:         msg.RawPayload,
								RxDelay:            ttnpb.RX_DELAY_3,
								SelectedMACVersion: ttnpb.MAC_V1_1,
							})
					},
					&grpc.EmptyCallOption{},
					NsJsHandleJoinResponse{
						Error: errTest,
					},
				), should.BeTrue) {
					return false
				}

				if !a.So(test.AssertEventPubSubPublishRequest(ctx, env.Events, func(ev events.Event) bool {
					if !a.So(ev.Data(), should.BeError) {
						return false
					}
					err, ok := errors.From(ev.Data().(error))
					if !a.So(ok, should.BeTrue) {
						return false
					}
					return a.So(ev, should.ResembleEvent, EvtDropJoinRequest(reqCtx, makeOTAAIdentifiers(nil), err))
				}), should.BeTrue) {
					return false
				}

				_ = sendUplinkDuplicates(ctx, handle, env.CollectionDone, makeJoinRequest, start, duplicateCount)

				return assertHandleUplinkResponse(ctx, handleUplinkErrCh, func(err error) bool {
					return a.So(err, should.BeError)
				})
			},
		},
		{
			Name: "Join-request/Get OTAA device/1.1/JS not found",
			Handler: func(ctx context.Context, env TestEnvironment, asRecvCh <-chan AsNsLinkRecvRequest, handle func(context.Context, *ttnpb.UplinkMessage) <-chan error) bool {
				t := test.MustTFromContext(ctx)
				a := assertions.New(t)

				start := time.Now()

				msg := makeJoinRequest(false)

				handleUplinkErrCh := handle(ctx, msg)

				getDevice := &ttnpb.EndDevice{
					EndDeviceIdentifiers: *makeOTAAIdentifiers(nil),
					FrequencyPlanID:      test.EUFrequencyPlanID,
					LoRaWANPHYVersion:    ttnpb.PHY_V1_1_REV_B,
					LoRaWANVersion:       ttnpb.MAC_V1_1,
					MACSettings: &ttnpb.MACSettings{
						Rx1Delay: &ttnpb.MACSettings_RxDelayValue{
							Value: ttnpb.RX_DELAY_3,
						},
					},
					SupportsJoin: true,
					CreatedAt:    start,
					UpdatedAt:    time.Now(),
				}

				var reqCtx context.Context
				var reqCorrelationIDs []string
				select {
				case <-ctx.Done():
					t.Error("Timed out while waiting for DeviceRegistry.GetByEUI to be called")
					return false

				case req := <-env.DeviceRegistry.GetByEUI:
					reqCtx = req.Context
					reqCorrelationIDs = events.CorrelationIDsFromContext(req.Context)
					for _, id := range correlationIDs {
						a.So(reqCorrelationIDs, should.Contain, id)
					}
					a.So(reqCorrelationIDs, should.HaveLength, len(correlationIDs)+2)
					a.So(req.JoinEUI, should.Resemble, joinEUI)
					a.So(req.DevEUI, should.Resemble, devEUI)
					a.So(req.Paths, should.HaveSameElementsDeep, joinGetByEUIPaths[:])
					req.Response <- DeviceRegistryGetByEUIResponse{
						Device: CopyEndDevice(getDevice),
					}
				}

				if !a.So(test.AssertClusterGetPeerRequest(ctx, env.Cluster.GetPeer, func(ctx context.Context, role ttnpb.PeerInfo_Role, ids ttnpb.Identifiers) bool {
					return a.So(ctx, should.HaveParentContextOrEqual, reqCtx) &&
						a.So(role, should.Equal, ttnpb.PeerInfo_JOIN_SERVER) &&
						a.So(ids, should.Resemble, getDevice.EndDeviceIdentifiers)
				},
					nil,
				), should.BeTrue) {
					return false
				}

				if !a.So(test.AssertEventPubSubPublishRequest(ctx, env.Events, func(ev events.Event) bool {
					if !a.So(ev.Data(), should.BeError) {
						return false
					}
					err, ok := errors.From(ev.Data().(error))
					if !a.So(ok, should.BeTrue) {
						return false
					}
					return a.So(ev, should.ResembleEvent, EvtDropJoinRequest(reqCtx, makeOTAAIdentifiers(nil), err))
				}), should.BeTrue) {
					return false
				}

				_ = sendUplinkDuplicates(ctx, handle, env.CollectionDone, makeJoinRequest, start, duplicateCount)

				return assertHandleUplinkResponse(ctx, handleUplinkErrCh, func(err error) bool {
					return a.So(err, should.BeError)
				})
			},
		},

		{
			Name: "Join-request/Get OTAA device/1.0.2/JS accept/Set fail",
			Handler: func(ctx context.Context, env TestEnvironment, asRecvCh <-chan AsNsLinkRecvRequest, handle func(context.Context, *ttnpb.UplinkMessage) <-chan error) bool {
				t := test.MustTFromContext(ctx)
				a := assertions.New(t)

				start := time.Now()

				msg := makeJoinRequest(false)

				handleUplinkErrCh := handle(ctx, msg)

				getDevice := &ttnpb.EndDevice{
					EndDeviceIdentifiers: *makeOTAAIdentifiers(nil),
					FrequencyPlanID:      test.EUFrequencyPlanID,
					LoRaWANPHYVersion:    ttnpb.PHY_V1_0_2_REV_B,
					LoRaWANVersion:       ttnpb.MAC_V1_0_2,
					MACSettings: &ttnpb.MACSettings{
						DesiredRx1Delay: &ttnpb.MACSettings_RxDelayValue{
							Value: ttnpb.RX_DELAY_3,
						},
					},
					RecentUplinks: []*ttnpb.UplinkMessage{
						makeLegacyDataUplink(33, true),
					},
					SupportsJoin: true,
					CreatedAt:    start,
					UpdatedAt:    time.Now(),
				}

				var reqCtx context.Context
				var reqCorrelationIDs []string
				select {
				case <-ctx.Done():
					t.Error("Timed out while waiting for DeviceRegistry.GetByEUI to be called")
					return false

				case req := <-env.DeviceRegistry.GetByEUI:
					reqCtx = req.Context
					reqCorrelationIDs = events.CorrelationIDsFromContext(req.Context)
					for _, id := range correlationIDs {
						a.So(reqCorrelationIDs, should.Contain, id)
					}
					a.So(reqCorrelationIDs, should.HaveLength, len(correlationIDs)+2)
					a.So(req.JoinEUI, should.Resemble, joinEUI)
					a.So(req.DevEUI, should.Resemble, devEUI)
					a.So(req.Paths, should.HaveSameElementsDeep, joinGetByEUIPaths[:])
					req.Response <- DeviceRegistryGetByEUIResponse{
						Device: CopyEndDevice(getDevice),
					}
				}

				joinResp := makeJoinResponse(ttnpb.MAC_V1_0_2)

				var joinReq *ttnpb.JoinRequest
				if !a.So(AssertNsJsPeerHandleAuthJoinRequest(ctx, env.Cluster.GetPeer, env.Cluster.Auth,
					func(ctx context.Context, ids ttnpb.Identifiers) bool {
						return a.So(ctx, should.HaveParentContextOrEqual, reqCtx) &&
							a.So(ids, should.Resemble, *makeOTAAIdentifiers(nil))
					},
					func(ctx context.Context, req *ttnpb.JoinRequest) bool {
						joinReq = req
						return a.So(req.CorrelationIDs, should.HaveSameElementsDeep, reqCorrelationIDs) &&
							a.So(req.DevAddr, should.NotBeEmpty) &&
							a.So(req.DevAddr.NwkID(), should.Resemble, netID.ID()) &&
							a.So(req.DevAddr.NetIDType(), should.Equal, netID.Type()) &&
							a.So(req, should.Resemble, &ttnpb.JoinRequest{
								CFList: &ttnpb.CFList{
									Type: ttnpb.CFListType_FREQUENCIES,
									Freq: []uint32{8671000, 8673000, 8675000, 8677000, 8679000},
								},
								CorrelationIDs:     req.CorrelationIDs,
								DevAddr:            req.DevAddr,
								NetID:              netID,
								RawPayload:         msg.RawPayload,
								RxDelay:            ttnpb.RX_DELAY_3,
								SelectedMACVersion: ttnpb.MAC_V1_0_2,
							})
					},
					&grpc.EmptyCallOption{},
					NsJsHandleJoinResponse{
						Response: joinResp,
					},
				), should.BeTrue) {
					return false
				}

				if !a.So(test.AssertEventPubSubPublishRequest(ctx, env.Events, func(ev events.Event) bool {
					return a.So(ev, should.ResembleEvent, EvtForwardJoinRequest(reqCtx, makeOTAAIdentifiers(nil), nil))
				}), should.BeTrue) {
					return false
				}

				mds := sendUplinkDuplicates(ctx, handle, env.DeduplicationDone, makeJoinRequest, start, duplicateCount)
				mds = append(mds, msg.RxMetadata...)

				if !a.So(test.AssertEventPubSubPublishRequest(ctx, env.Events, func(ev events.Event) bool {
					return a.So(ev, should.ResembleEvent, EvtMergeMetadata(reqCtx, makeOTAAIdentifiers(nil), len(mds)))
				}), should.BeTrue) {
					return false
				}

				var recentUp *ttnpb.UplinkMessage
				select {
				case <-ctx.Done():
					t.Error("Timed out while waiting for DeviceRegistry.SetByID to be called")
					return false

				case req := <-env.DeviceRegistry.SetByID:
					a.So(req.Context, should.HaveParentContextOrEqual, reqCtx)
					a.So(req.ApplicationIdentifiers, should.Resemble, appID)
					a.So(req.DeviceID, should.Resemble, devID)
					a.So(req.Paths, should.HaveSameElementsDeep, joinSetByEUIGetPaths[:])
					dev, sets, err := req.Func(&ttnpb.EndDevice{
						FrequencyPlanID:   test.EUFrequencyPlanID,
						LoRaWANPHYVersion: ttnpb.PHY_V1_0_2_REV_B,
						RecentUplinks: []*ttnpb.UplinkMessage{
							makeLegacyDataUplink(33, true),
						},
						QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
							makeApplicationDownlink(),
						},
					})
					if !a.So(err, should.BeNil) || !a.So(dev, should.NotBeNil) {
						return false
					}
					a.So(sets, should.HaveSameElementsDeep, joinSetByEUISetPaths[:])

					macState := MakeDefaultEU868MACState(ttnpb.CLASS_A, ttnpb.MAC_V1_0_2)
					macState.DesiredParameters.Rx1Delay = ttnpb.RX_DELAY_3
					macState.CurrentParameters.Rx1Delay = macState.DesiredParameters.Rx1Delay
					macState.CurrentParameters.Channels = macState.DesiredParameters.Channels
					macState.RxWindowsAvailable = true
					macState.QueuedJoinAccept = &ttnpb.MACState_JoinAccept{
						Keys:    *makeSessionKeys(ttnpb.MAC_V1_0_2),
						Payload: joinResp.RawPayload,
						Request: *joinReq,
					}
					a.So(dev.PendingMACState, should.Resemble, macState)
					a.So(dev.QueuedApplicationDownlinks, should.BeNil)
					if a.So(dev.RecentUplinks, should.NotBeEmpty) {
						recentUp = dev.RecentUplinks[len(dev.RecentUplinks)-1]
						a.So([]time.Time{start, recentUp.ReceivedAt, time.Now()}, should.BeChronological)
						a.So(recentUp.RxMetadata, should.HaveSameElementsDiff, mds)
						expectedUp := makeJoinRequest(true)
						expectedUp.CorrelationIDs = reqCorrelationIDs
						expectedUp.DeviceChannelIndex = 2
						expectedUp.ReceivedAt = recentUp.ReceivedAt
						expectedUp.RxMetadata = recentUp.RxMetadata
						expectedUp.Settings.DataRateIndex = ttnpb.DATA_RATE_1
						a.So(dev.RecentUplinks, should.HaveEmptyDiff, append(CopyUplinkMessages(getDevice.RecentUplinks...), expectedUp))
					}
					req.Response <- DeviceRegistrySetByIDResponse{
						Error: errTest,
					}
				}

				if !a.So(test.AssertEventPubSubPublishRequest(ctx, env.Events, func(ev events.Event) bool {
					a.So(ev, should.ResembleEvent, EvtDropJoinRequest(reqCtx, makeOTAAIdentifiers(nil), errTest))
					return true
				}), should.BeTrue) {
					return false
				}

				_ = sendUplinkDuplicates(ctx, handle, env.CollectionDone, func(decoded bool) *ttnpb.UplinkMessage {
					msg := makeJoinRequest(decoded)
					if !decoded {
						return msg
					}
					msg.DeviceChannelIndex = 2
					msg.Settings.DataRateIndex = ttnpb.DATA_RATE_1
					return msg
				}, start, duplicateCount)

				return assertHandleUplinkResponse(ctx, handleUplinkErrCh, func(err error) bool {
					return a.So(err, should.EqualErrorOrDefinition, errTest)
				})
			},
		},

		{
			Name: "Join-request/Get OTAA device/1.1/JS accept/Set success/Downlink add fail",
			Handler: func(ctx context.Context, env TestEnvironment, asRecvCh <-chan AsNsLinkRecvRequest, handle func(context.Context, *ttnpb.UplinkMessage) <-chan error) bool {
				t := test.MustTFromContext(ctx)
				a := assertions.New(t)

				start := time.Now()

				msg := makeJoinRequest(false)

				handleUplinkErrCh := handle(ctx, msg)

				getDevice := &ttnpb.EndDevice{
					EndDeviceIdentifiers: *makeOTAAIdentifiers(nil),
					FrequencyPlanID:      test.EUFrequencyPlanID,
					LoRaWANPHYVersion:    ttnpb.PHY_V1_1_REV_B,
					LoRaWANVersion:       ttnpb.MAC_V1_1,
					MACSettings: &ttnpb.MACSettings{
						DesiredRx1Delay: &ttnpb.MACSettings_RxDelayValue{
							Value: ttnpb.RX_DELAY_3,
						},
					},
					RecentUplinks: []*ttnpb.UplinkMessage{
						makeDataUplink(33, true),
					},
					SupportsJoin: true,
					CreatedAt:    start,
					UpdatedAt:    time.Now(),
				}

				var reqCtx context.Context
				var reqCorrelationIDs []string
				select {
				case <-ctx.Done():
					t.Error("Timed out while waiting for DeviceRegistry.GetByEUI to be called")
					return false

				case req := <-env.DeviceRegistry.GetByEUI:
					reqCtx = req.Context
					reqCorrelationIDs = events.CorrelationIDsFromContext(req.Context)
					for _, id := range correlationIDs {
						a.So(reqCorrelationIDs, should.Contain, id)
					}
					a.So(reqCorrelationIDs, should.HaveLength, len(correlationIDs)+2)
					a.So(req.JoinEUI, should.Resemble, joinEUI)
					a.So(req.DevEUI, should.Resemble, devEUI)
					a.So(req.Paths, should.HaveSameElementsDeep, joinGetByEUIPaths[:])
					req.Response <- DeviceRegistryGetByEUIResponse{
						Device: CopyEndDevice(getDevice),
					}
				}

				joinResp := makeJoinResponse(ttnpb.MAC_V1_1)

				var joinReq *ttnpb.JoinRequest
				if !a.So(AssertNsJsPeerHandleAuthJoinRequest(ctx, env.Cluster.GetPeer, env.Cluster.Auth,
					func(ctx context.Context, ids ttnpb.Identifiers) bool {
						return a.So(ctx, should.HaveParentContextOrEqual, reqCtx) &&
							a.So(ids, should.Resemble, *makeOTAAIdentifiers(nil))
					},
					func(ctx context.Context, req *ttnpb.JoinRequest) bool {
						joinReq = req
						return a.So(req.CorrelationIDs, should.HaveSameElementsDeep, reqCorrelationIDs) &&
							a.So(req.DevAddr, should.NotBeEmpty) &&
							a.So(req.DevAddr.NwkID(), should.Resemble, netID.ID()) &&
							a.So(req.DevAddr.NetIDType(), should.Equal, netID.Type()) &&
							a.So(req, should.Resemble, &ttnpb.JoinRequest{
								CFList: &ttnpb.CFList{
									Type: ttnpb.CFListType_FREQUENCIES,
									Freq: []uint32{8671000, 8673000, 8675000, 8677000, 8679000},
								},
								CorrelationIDs: req.CorrelationIDs,
								DevAddr:        req.DevAddr,
								DownlinkSettings: ttnpb.DLSettings{
									OptNeg: true,
								},
								NetID:              netID,
								RawPayload:         msg.RawPayload,
								RxDelay:            ttnpb.RX_DELAY_3,
								SelectedMACVersion: ttnpb.MAC_V1_1,
							})
					},
					&grpc.EmptyCallOption{},
					NsJsHandleJoinResponse{
						Response: joinResp,
					},
				), should.BeTrue) {
					return false
				}

				if !a.So(test.AssertEventPubSubPublishRequest(ctx, env.Events, func(ev events.Event) bool {
					return a.So(ev, should.ResembleEvent, EvtForwardJoinRequest(reqCtx, makeOTAAIdentifiers(nil), nil))
				}), should.BeTrue) {
					return false
				}

				mds := sendUplinkDuplicates(ctx, handle, env.DeduplicationDone, makeJoinRequest, start, duplicateCount)
				mds = append(mds, msg.RxMetadata...)

				if !a.So(test.AssertEventPubSubPublishRequest(ctx, env.Events, func(ev events.Event) bool {
					return a.So(ev, should.ResembleEvent, EvtMergeMetadata(reqCtx, makeOTAAIdentifiers(nil), len(mds)))
				}), should.BeTrue) {
					return false
				}

				var recentUp *ttnpb.UplinkMessage
				select {
				case <-ctx.Done():
					t.Error("Timed out while waiting for DeviceRegistry.SetByID to be called")
					return false

				case req := <-env.DeviceRegistry.SetByID:
					a.So(req.Context, should.HaveParentContextOrEqual, reqCtx)
					a.So(req.ApplicationIdentifiers, should.Resemble, appID)
					a.So(req.DeviceID, should.Resemble, devID)
					a.So(req.Paths, should.HaveSameElementsDeep, joinSetByEUIGetPaths[:])
					dev, sets, err := req.Func(&ttnpb.EndDevice{
						FrequencyPlanID:   test.EUFrequencyPlanID,
						LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
						RecentUplinks: []*ttnpb.UplinkMessage{
							makeDataUplink(33, true),
						},
						QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
							makeApplicationDownlink(),
						},
					})
					if !a.So(err, should.BeNil) || !a.So(dev, should.NotBeNil) {
						return false
					}
					a.So(sets, should.HaveSameElementsDeep, joinSetByEUISetPaths[:])

					macState := MakeDefaultEU868MACState(ttnpb.CLASS_A, ttnpb.MAC_V1_1)
					macState.DesiredParameters.Rx1Delay = ttnpb.RX_DELAY_3
					macState.CurrentParameters.Rx1Delay = macState.DesiredParameters.Rx1Delay
					macState.CurrentParameters.Channels = macState.DesiredParameters.Channels
					macState.RxWindowsAvailable = true
					macState.QueuedJoinAccept = &ttnpb.MACState_JoinAccept{
						Keys:    *makeSessionKeys(ttnpb.MAC_V1_1),
						Payload: joinResp.RawPayload,
						Request: *joinReq,
					}
					a.So(dev.PendingMACState, should.Resemble, macState)
					a.So(dev.QueuedApplicationDownlinks, should.BeNil)
					if a.So(dev.RecentUplinks, should.NotBeEmpty) {
						recentUp = dev.RecentUplinks[len(dev.RecentUplinks)-1]
						a.So([]time.Time{start, recentUp.ReceivedAt, time.Now()}, should.BeChronological)
						a.So(recentUp.RxMetadata, should.HaveSameElementsDiff, mds)
						expectedUp := makeJoinRequest(true)
						expectedUp.CorrelationIDs = reqCorrelationIDs
						expectedUp.DeviceChannelIndex = 2
						expectedUp.ReceivedAt = recentUp.ReceivedAt
						expectedUp.RxMetadata = recentUp.RxMetadata
						expectedUp.Settings.DataRateIndex = ttnpb.DATA_RATE_1
						a.So(dev.RecentUplinks, should.HaveEmptyDiff, append(CopyUplinkMessages(getDevice.RecentUplinks...), expectedUp))
					}
					req.Response <- DeviceRegistrySetByIDResponse{
						Device: &ttnpb.EndDevice{
							EndDeviceIdentifiers:       *makeOTAAIdentifiers(nil),
							PendingMACState:            macState,
							QueuedApplicationDownlinks: dev.QueuedApplicationDownlinks,
							RecentUplinks:              dev.RecentUplinks,
							CreatedAt:                  start,
							UpdatedAt:                  time.Now(),
						},
					}
				}

				if !a.So(AssertDownlinkTaskAddRequest(ctx, env.DownlinkTasks.Add, func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, startAt time.Time, replace bool) bool {
					return a.So(ctx, should.HaveParentContextOrEqual, reqCtx) &&
						a.So(ids, should.Resemble, *makeOTAAIdentifiers(nil)) &&
						a.So(startAt, should.Resemble, recentUp.ReceivedAt.Add(5*time.Second-NSScheduleWindow())) &&
						a.So(replace, should.BeTrue)
				},
					errTest,
				), should.BeTrue) {
					return false
				}

				_ = sendUplinkDuplicates(ctx, handle, env.CollectionDone, func(decoded bool) *ttnpb.UplinkMessage {
					msg := makeJoinRequest(decoded)
					if !decoded {
						return msg
					}
					msg.DeviceChannelIndex = 2
					msg.Settings.DataRateIndex = ttnpb.DATA_RATE_1
					return msg
				}, start, duplicateCount)

				if !assertHandleUplinkResponse(ctx, handleUplinkErrCh, func(err error) bool {
					return a.So(err, should.BeNil)
				}) {
					return false
				}

				if asRecvCh != nil {
					select {
					case <-ctx.Done():
						t.Error("Timed out while waiting for NetworkServer.handleASUplink to be called")
						return false

					case req := <-asRecvCh:
						a.So(req.Uplink, should.Resemble, &ttnpb.ApplicationUp{
							CorrelationIDs:       reqCorrelationIDs,
							EndDeviceIdentifiers: *makeOTAAIdentifiers(&joinReq.DevAddr),
							Up: &ttnpb.ApplicationUp_JoinAccept{JoinAccept: &ttnpb.ApplicationJoinAccept{
								AppSKey: makeSessionKeys(ttnpb.MAC_V1_1).AppSKey,
								InvalidatedDownlinks: []*ttnpb.ApplicationDownlink{
									makeApplicationDownlink(),
								},
								SessionKeyID: makeSessionKeys(ttnpb.MAC_V1_1).SessionKeyID,
							}},
						})
						req.Response <- ttnpb.Empty
					}
				}
				return true
			},
		},

		{
			Name: "Join-request/Get OTAA device/1.0.2/JS accept/Set success/Downlink add success",
			Handler: func(ctx context.Context, env TestEnvironment, asRecvCh <-chan AsNsLinkRecvRequest, handle func(context.Context, *ttnpb.UplinkMessage) <-chan error) bool {
				t := test.MustTFromContext(ctx)
				a := assertions.New(t)

				start := time.Now()

				msg := makeJoinRequest(false)

				handleUplinkErrCh := handle(ctx, msg)

				getDevice := &ttnpb.EndDevice{
					EndDeviceIdentifiers: *makeOTAAIdentifiers(nil),
					FrequencyPlanID:      test.EUFrequencyPlanID,
					LoRaWANPHYVersion:    ttnpb.PHY_V1_0_2_REV_B,
					LoRaWANVersion:       ttnpb.MAC_V1_0_2,
					MACSettings: &ttnpb.MACSettings{
						DesiredRx1Delay: &ttnpb.MACSettings_RxDelayValue{
							Value: ttnpb.RX_DELAY_3,
						},
					},
					RecentUplinks: []*ttnpb.UplinkMessage{
						makeLegacyDataUplink(33, true),
					},
					SupportsJoin: true,
					CreatedAt:    start,
					UpdatedAt:    time.Now(),
				}

				var reqCtx context.Context
				var reqCorrelationIDs []string
				select {
				case <-ctx.Done():
					t.Error("Timed out while waiting for DeviceRegistry.GetByEUI to be called")
					return false

				case req := <-env.DeviceRegistry.GetByEUI:
					reqCtx = req.Context
					reqCorrelationIDs = events.CorrelationIDsFromContext(req.Context)
					for _, id := range correlationIDs {
						a.So(reqCorrelationIDs, should.Contain, id)
					}
					a.So(reqCorrelationIDs, should.HaveLength, len(correlationIDs)+2)
					a.So(req.JoinEUI, should.Resemble, joinEUI)
					a.So(req.DevEUI, should.Resemble, devEUI)
					a.So(req.Paths, should.HaveSameElementsDeep, joinGetByEUIPaths[:])
					req.Response <- DeviceRegistryGetByEUIResponse{
						Device: CopyEndDevice(getDevice),
					}
				}

				joinResp := makeJoinResponse(ttnpb.MAC_V1_0_2)

				var joinReq *ttnpb.JoinRequest
				if !a.So(AssertNsJsPeerHandleAuthJoinRequest(ctx, env.Cluster.GetPeer, env.Cluster.Auth,
					func(ctx context.Context, ids ttnpb.Identifiers) bool {
						return a.So(ctx, should.HaveParentContextOrEqual, reqCtx) &&
							a.So(ids, should.Resemble, *makeOTAAIdentifiers(nil))
					},
					func(ctx context.Context, req *ttnpb.JoinRequest) bool {
						joinReq = req
						return a.So(req.CorrelationIDs, should.HaveSameElementsDeep, reqCorrelationIDs) &&
							a.So(req.DevAddr, should.NotBeEmpty) &&
							a.So(req.DevAddr.NwkID(), should.Resemble, netID.ID()) &&
							a.So(req.DevAddr.NetIDType(), should.Equal, netID.Type()) &&
							a.So(req, should.Resemble, &ttnpb.JoinRequest{
								CFList: &ttnpb.CFList{
									Type: ttnpb.CFListType_FREQUENCIES,
									Freq: []uint32{8671000, 8673000, 8675000, 8677000, 8679000},
								},
								CorrelationIDs:     req.CorrelationIDs,
								DevAddr:            req.DevAddr,
								NetID:              netID,
								RawPayload:         msg.RawPayload,
								RxDelay:            ttnpb.RX_DELAY_3,
								SelectedMACVersion: ttnpb.MAC_V1_0_2,
							})
					},
					&grpc.EmptyCallOption{},
					NsJsHandleJoinResponse{
						Response: joinResp,
					},
				), should.BeTrue) {
					return false
				}

				if !a.So(test.AssertEventPubSubPublishRequest(ctx, env.Events, func(ev events.Event) bool {
					return a.So(ev, should.ResembleEvent, EvtForwardJoinRequest(reqCtx, makeOTAAIdentifiers(nil), nil))
				}), should.BeTrue) {
					return false
				}

				mds := sendUplinkDuplicates(ctx, handle, env.DeduplicationDone, makeJoinRequest, start, duplicateCount)
				mds = append(mds, msg.RxMetadata...)

				if !a.So(test.AssertEventPubSubPublishRequest(ctx, env.Events, func(ev events.Event) bool {
					return a.So(ev, should.ResembleEvent, EvtMergeMetadata(reqCtx, makeOTAAIdentifiers(nil), len(mds)))
				}), should.BeTrue) {
					return false
				}

				var recentUp *ttnpb.UplinkMessage
				select {
				case <-ctx.Done():
					t.Error("Timed out while waiting for DeviceRegistry.SetByID to be called")
					return false

				case req := <-env.DeviceRegistry.SetByID:
					a.So(req.Context, should.HaveParentContextOrEqual, reqCtx)
					a.So(req.ApplicationIdentifiers, should.Resemble, appID)
					a.So(req.DeviceID, should.Resemble, devID)
					a.So(req.Paths, should.HaveSameElementsDeep, joinSetByEUIGetPaths[:])
					dev, sets, err := req.Func(&ttnpb.EndDevice{
						FrequencyPlanID:   test.EUFrequencyPlanID,
						LoRaWANPHYVersion: ttnpb.PHY_V1_0_2_REV_B,
						RecentUplinks: []*ttnpb.UplinkMessage{
							makeLegacyDataUplink(33, true),
						},
						QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
							makeApplicationDownlink(),
						},
					})
					if !a.So(err, should.BeNil) || !a.So(dev, should.NotBeNil) {
						return false
					}
					a.So(sets, should.HaveSameElementsDeep, joinSetByEUISetPaths[:])

					macState := MakeDefaultEU868MACState(ttnpb.CLASS_A, ttnpb.MAC_V1_0_2)
					macState.DesiredParameters.Rx1Delay = ttnpb.RX_DELAY_3
					macState.CurrentParameters.Rx1Delay = macState.DesiredParameters.Rx1Delay
					macState.CurrentParameters.Channels = macState.DesiredParameters.Channels
					macState.RxWindowsAvailable = true
					macState.QueuedJoinAccept = &ttnpb.MACState_JoinAccept{
						Keys:    *makeSessionKeys(ttnpb.MAC_V1_0_2),
						Payload: joinResp.RawPayload,
						Request: *joinReq,
					}
					a.So(dev.PendingMACState, should.Resemble, macState)
					a.So(dev.QueuedApplicationDownlinks, should.BeNil)
					if a.So(dev.RecentUplinks, should.NotBeEmpty) {
						recentUp = dev.RecentUplinks[len(dev.RecentUplinks)-1]
						a.So([]time.Time{start, recentUp.ReceivedAt, time.Now()}, should.BeChronological)
						a.So(recentUp.RxMetadata, should.HaveSameElementsDiff, mds)
						expectedUp := makeJoinRequest(true)
						expectedUp.CorrelationIDs = reqCorrelationIDs
						expectedUp.DeviceChannelIndex = 2
						expectedUp.ReceivedAt = recentUp.ReceivedAt
						expectedUp.RxMetadata = recentUp.RxMetadata
						expectedUp.Settings.DataRateIndex = ttnpb.DATA_RATE_1
						a.So(dev.RecentUplinks, should.HaveEmptyDiff, append(CopyUplinkMessages(getDevice.RecentUplinks...), expectedUp))
					}
					req.Response <- DeviceRegistrySetByIDResponse{
						Device: &ttnpb.EndDevice{
							EndDeviceIdentifiers:       *makeOTAAIdentifiers(nil),
							PendingMACState:            macState,
							QueuedApplicationDownlinks: dev.QueuedApplicationDownlinks,
							RecentUplinks:              dev.RecentUplinks,
							CreatedAt:                  start,
							UpdatedAt:                  time.Now(),
						},
					}
				}

				if !a.So(AssertDownlinkTaskAddRequest(ctx, env.DownlinkTasks.Add, func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, startAt time.Time, replace bool) bool {
					return a.So(ctx, should.HaveParentContextOrEqual, reqCtx) &&
						a.So(ids, should.Resemble, *makeOTAAIdentifiers(nil)) &&
						a.So(startAt, should.Resemble, recentUp.ReceivedAt.Add(5*time.Second-NSScheduleWindow())) &&
						a.So(replace, should.BeTrue)
				},
					nil,
				), should.BeTrue) {
					return false
				}

				_ = sendUplinkDuplicates(ctx, handle, env.CollectionDone, func(decoded bool) *ttnpb.UplinkMessage {
					msg := makeJoinRequest(decoded)
					if !decoded {
						return msg
					}
					msg.DeviceChannelIndex = 2
					msg.Settings.DataRateIndex = ttnpb.DATA_RATE_1
					return msg
				}, start, duplicateCount)

				if !assertHandleUplinkResponse(ctx, handleUplinkErrCh, func(err error) bool {
					return a.So(err, should.BeNil)
				}) {
					return false
				}

				if asRecvCh != nil {
					select {
					case <-ctx.Done():
						t.Error("Timed out while waiting for NetworkServer.handleASUplink to be called")
						return false

					case req := <-asRecvCh:
						a.So(req.Uplink, should.Resemble, &ttnpb.ApplicationUp{
							CorrelationIDs:       reqCorrelationIDs,
							EndDeviceIdentifiers: *makeOTAAIdentifiers(&joinReq.DevAddr),
							Up: &ttnpb.ApplicationUp_JoinAccept{JoinAccept: &ttnpb.ApplicationJoinAccept{
								AppSKey: makeSessionKeys(ttnpb.MAC_V1_0_2).AppSKey,
								InvalidatedDownlinks: []*ttnpb.ApplicationDownlink{
									makeApplicationDownlink(),
								},
								SessionKeyID: makeSessionKeys(ttnpb.MAC_V1_0_2).SessionKeyID,
							}},
						})
						req.Response <- ttnpb.Empty
					}
				}
				return true
			},
		},

		{
			Name: "Join-request/Get OTAA device/1.1/JS accept/Set success/Downlink add success",
			Handler: func(ctx context.Context, env TestEnvironment, asRecvCh <-chan AsNsLinkRecvRequest, handle func(context.Context, *ttnpb.UplinkMessage) <-chan error) bool {
				t := test.MustTFromContext(ctx)
				a := assertions.New(t)

				start := time.Now()

				msg := makeJoinRequest(false)

				handleUplinkErrCh := handle(ctx, msg)

				getDevice := &ttnpb.EndDevice{
					EndDeviceIdentifiers: *makeOTAAIdentifiers(nil),
					FrequencyPlanID:      test.EUFrequencyPlanID,
					LoRaWANPHYVersion:    ttnpb.PHY_V1_1_REV_B,
					LoRaWANVersion:       ttnpb.MAC_V1_1,
					MACSettings: &ttnpb.MACSettings{
						DesiredRx1Delay: &ttnpb.MACSettings_RxDelayValue{
							Value: ttnpb.RX_DELAY_3,
						},
					},
					RecentUplinks: []*ttnpb.UplinkMessage{
						makeDataUplink(33, true),
					},
					SupportsJoin: true,
					CreatedAt:    start,
					UpdatedAt:    time.Now(),
				}

				var reqCtx context.Context
				var reqCorrelationIDs []string
				select {
				case <-ctx.Done():
					t.Error("Timed out while waiting for DeviceRegistry.GetByEUI to be called")
					return false

				case req := <-env.DeviceRegistry.GetByEUI:
					reqCtx = req.Context
					reqCorrelationIDs = events.CorrelationIDsFromContext(req.Context)
					for _, id := range correlationIDs {
						a.So(reqCorrelationIDs, should.Contain, id)
					}
					a.So(reqCorrelationIDs, should.HaveLength, len(correlationIDs)+2)
					a.So(req.JoinEUI, should.Resemble, joinEUI)
					a.So(req.DevEUI, should.Resemble, devEUI)
					a.So(req.Paths, should.HaveSameElementsDeep, joinGetByEUIPaths[:])
					req.Response <- DeviceRegistryGetByEUIResponse{
						Device: CopyEndDevice(getDevice),
					}
				}

				joinResp := makeJoinResponse(ttnpb.MAC_V1_1)

				var joinReq *ttnpb.JoinRequest
				if !a.So(AssertNsJsPeerHandleAuthJoinRequest(ctx, env.Cluster.GetPeer, env.Cluster.Auth,
					func(ctx context.Context, ids ttnpb.Identifiers) bool {
						return a.So(ctx, should.HaveParentContextOrEqual, reqCtx) &&
							a.So(ids, should.Resemble, *makeOTAAIdentifiers(nil))
					},
					func(ctx context.Context, req *ttnpb.JoinRequest) bool {
						joinReq = req
						return a.So(req.CorrelationIDs, should.HaveSameElementsDeep, reqCorrelationIDs) &&
							a.So(req.DevAddr, should.NotBeEmpty) &&
							a.So(req.DevAddr.NwkID(), should.Resemble, netID.ID()) &&
							a.So(req.DevAddr.NetIDType(), should.Equal, netID.Type()) &&
							a.So(req, should.Resemble, &ttnpb.JoinRequest{
								CFList: &ttnpb.CFList{
									Type: ttnpb.CFListType_FREQUENCIES,
									Freq: []uint32{8671000, 8673000, 8675000, 8677000, 8679000},
								},
								CorrelationIDs: req.CorrelationIDs,
								DevAddr:        req.DevAddr,
								DownlinkSettings: ttnpb.DLSettings{
									OptNeg: true,
								},
								NetID:              netID,
								RawPayload:         msg.RawPayload,
								RxDelay:            ttnpb.RX_DELAY_3,
								SelectedMACVersion: ttnpb.MAC_V1_1,
							})
					},
					&grpc.EmptyCallOption{},
					NsJsHandleJoinResponse{
						Response: joinResp,
					},
				), should.BeTrue) {
					return false
				}

				if !a.So(test.AssertEventPubSubPublishRequest(ctx, env.Events, func(ev events.Event) bool {
					return a.So(ev, should.ResembleEvent, EvtForwardJoinRequest(reqCtx, makeOTAAIdentifiers(nil), nil))
				}), should.BeTrue) {
					return false
				}

				mds := sendUplinkDuplicates(ctx, handle, env.DeduplicationDone, makeJoinRequest, start, duplicateCount)
				mds = append(mds, msg.RxMetadata...)

				if !a.So(test.AssertEventPubSubPublishRequest(ctx, env.Events, func(ev events.Event) bool {
					return a.So(ev, should.ResembleEvent, EvtMergeMetadata(reqCtx, makeOTAAIdentifiers(nil), len(mds)))
				}), should.BeTrue) {
					return false
				}

				var recentUp *ttnpb.UplinkMessage
				select {
				case <-ctx.Done():
					t.Error("Timed out while waiting for DeviceRegistry.SetByID to be called")
					return false

				case req := <-env.DeviceRegistry.SetByID:
					a.So(req.Context, should.HaveParentContextOrEqual, reqCtx)
					a.So(req.ApplicationIdentifiers, should.Resemble, appID)
					a.So(req.DeviceID, should.Resemble, devID)
					a.So(req.Paths, should.HaveSameElementsDeep, joinSetByEUIGetPaths[:])
					dev, sets, err := req.Func(&ttnpb.EndDevice{
						FrequencyPlanID:   test.EUFrequencyPlanID,
						LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
						RecentUplinks: []*ttnpb.UplinkMessage{
							makeDataUplink(33, true),
						},
						QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
							makeApplicationDownlink(),
						},
					})
					if !a.So(err, should.BeNil) || !a.So(dev, should.NotBeNil) {
						return false
					}
					a.So(sets, should.HaveSameElementsDeep, joinSetByEUISetPaths[:])

					macState := MakeDefaultEU868MACState(ttnpb.CLASS_A, ttnpb.MAC_V1_1)
					macState.DesiredParameters.Rx1Delay = ttnpb.RX_DELAY_3
					macState.CurrentParameters.Rx1Delay = macState.DesiredParameters.Rx1Delay
					macState.CurrentParameters.Channels = macState.DesiredParameters.Channels
					macState.RxWindowsAvailable = true
					macState.QueuedJoinAccept = &ttnpb.MACState_JoinAccept{
						Keys:    *makeSessionKeys(ttnpb.MAC_V1_1),
						Payload: joinResp.RawPayload,
						Request: *joinReq,
					}
					a.So(dev.PendingMACState, should.Resemble, macState)
					a.So(dev.QueuedApplicationDownlinks, should.BeNil)
					if a.So(dev.RecentUplinks, should.NotBeEmpty) {
						recentUp = dev.RecentUplinks[len(dev.RecentUplinks)-1]
						a.So([]time.Time{start, recentUp.ReceivedAt, time.Now()}, should.BeChronological)
						a.So(recentUp.RxMetadata, should.HaveSameElementsDiff, mds)
						expectedUp := makeJoinRequest(true)
						expectedUp.CorrelationIDs = reqCorrelationIDs
						expectedUp.DeviceChannelIndex = 2
						expectedUp.ReceivedAt = recentUp.ReceivedAt
						expectedUp.RxMetadata = recentUp.RxMetadata
						expectedUp.Settings.DataRateIndex = ttnpb.DATA_RATE_1
						a.So(dev.RecentUplinks, should.HaveEmptyDiff, append(CopyUplinkMessages(getDevice.RecentUplinks...), expectedUp))
					}
					req.Response <- DeviceRegistrySetByIDResponse{
						Device: &ttnpb.EndDevice{
							EndDeviceIdentifiers:       *makeOTAAIdentifiers(nil),
							PendingMACState:            macState,
							QueuedApplicationDownlinks: dev.QueuedApplicationDownlinks,
							RecentUplinks:              dev.RecentUplinks,
							CreatedAt:                  start,
							UpdatedAt:                  time.Now(),
						},
					}
				}

				if !a.So(AssertDownlinkTaskAddRequest(ctx, env.DownlinkTasks.Add, func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, startAt time.Time, replace bool) bool {
					return a.So(ctx, should.HaveParentContextOrEqual, reqCtx) &&
						a.So(ids, should.Resemble, *makeOTAAIdentifiers(nil)) &&
						a.So(startAt, should.Resemble, recentUp.ReceivedAt.Add(5*time.Second-NSScheduleWindow())) &&
						a.So(replace, should.BeTrue)
				},
					nil,
				), should.BeTrue) {
					return false
				}

				_ = sendUplinkDuplicates(ctx, handle, env.CollectionDone, func(decoded bool) *ttnpb.UplinkMessage {
					msg := makeJoinRequest(decoded)
					if !decoded {
						return msg
					}
					msg.DeviceChannelIndex = 2
					msg.Settings.DataRateIndex = ttnpb.DATA_RATE_1
					return msg
				}, start, duplicateCount)

				if !assertHandleUplinkResponse(ctx, handleUplinkErrCh, func(err error) bool {
					return a.So(err, should.BeNil)
				}) {
					return false
				}

				if asRecvCh != nil {
					select {
					case <-ctx.Done():
						t.Error("Timed out while waiting for NetworkServer.handleASUplink to be called")
						return false

					case req := <-asRecvCh:
						a.So(req.Uplink, should.Resemble, &ttnpb.ApplicationUp{
							CorrelationIDs:       reqCorrelationIDs,
							EndDeviceIdentifiers: *makeOTAAIdentifiers(&joinReq.DevAddr),
							Up: &ttnpb.ApplicationUp_JoinAccept{JoinAccept: &ttnpb.ApplicationJoinAccept{
								AppSKey: makeSessionKeys(ttnpb.MAC_V1_1).AppSKey,
								InvalidatedDownlinks: []*ttnpb.ApplicationDownlink{
									makeApplicationDownlink(),
								},
								SessionKeyID: makeSessionKeys(ttnpb.MAC_V1_1).SessionKeyID,
							}},
						})
						req.Response <- ttnpb.Empty
					}
				}
				return true
			},
		},

		{
			Name: "Rejoin-request",
			Handler: func(ctx context.Context, env TestEnvironment, asRecvCh <-chan AsNsLinkRecvRequest, handle func(context.Context, *ttnpb.UplinkMessage) <-chan error) bool {
				t := test.MustTFromContext(ctx)
				a := assertions.New(t)

				start := time.Now()

				msg := makeRejoinRequest(false)

				handleUplinkErrCh := handle(ctx, msg)

				_ = sendUplinkDuplicates(ctx, handle, env.CollectionDone, makeRejoinRequest, start, duplicateCount)

				return assertHandleUplinkResponse(ctx, handleUplinkErrCh, func(err error) bool {
					return a.So(err, should.NotBeNil)
				})
			},
		},

		{
			Name: "Data uplink/Matching device/No concurrent update/1.0.2/First transmission/No ADR/Set success/Downlink add success",
			Handler: func(ctx context.Context, env TestEnvironment, asRecvCh <-chan AsNsLinkRecvRequest, handle func(context.Context, *ttnpb.UplinkMessage) <-chan error) bool {
				t := test.MustTFromContext(ctx)
				a := assertions.New(t)

				start := time.Now()

				msg := makeLegacyDataUplink(34, false)

				handleUplinkErrCh := handle(ctx, msg)

				rangeDevice := &ttnpb.EndDevice{
					EndDeviceIdentifiers: *makeOTAAIdentifiers(&devAddr),
					FrequencyPlanID:      test.EUFrequencyPlanID,
					LoRaWANPHYVersion:    ttnpb.PHY_V1_0_2_REV_B,
					LoRaWANVersion:       ttnpb.MAC_V1_0_2,
					MACState:             MakeDefaultEU868MACState(ttnpb.CLASS_A, ttnpb.MAC_V1_0_2),
					RecentUplinks: []*ttnpb.UplinkMessage{
						makeLegacyDataUplink(31, true),
						makeLegacyDataUplink(32, true),
					},
					Session:   makeSession(ttnpb.MAC_V1_0_2, devAddr, 32),
					CreatedAt: start,
					UpdatedAt: time.Now(),
				}

				var upCtx context.Context
				var upCorrelationIDs []string
				select {
				case <-ctx.Done():
					t.Error("Timed out while waiting for DeviceRegistry.RangeByAddr to be called")
					return false

				case req := <-env.DeviceRegistry.RangeByAddr:
					upCtx = req.Context
					upCorrelationIDs = events.CorrelationIDsFromContext(req.Context)
					for _, id := range correlationIDs {
						a.So(upCorrelationIDs, should.Contain, id)
					}
					a.So(upCorrelationIDs, should.HaveLength, len(correlationIDs)+2)
					a.So(req.DevAddr, should.Resemble, devAddr)
					a.So(req.Paths, should.HaveSameElementsDeep, dataGetPaths[:])
					a.So(req.Func(CopyEndDevice(rangeDevice)), should.BeTrue)
					multicastDevice := CopyEndDevice(rangeDevice)
					multicastDevice.EndDeviceIdentifiers.DeviceID += "-multicast"
					multicastDevice.Multicast = true
					a.So(req.Func(multicastDevice), should.BeTrue)
					fCntTooHighDevice := CopyEndDevice(rangeDevice)
					fCntTooHighDevice.EndDeviceIdentifiers.DeviceID += "-too-high"
					fCntTooHighDevice.RecentUplinks = append(fCntTooHighDevice.RecentUplinks, makeLegacyDataUplink(42, true))
					fCntTooHighDevice.Session.LastFCntUp = 42
					a.So(req.Func(fCntTooHighDevice), should.BeTrue)
					req.Response <- nil
				}

				mds := sendUplinkDuplicates(ctx, handle, env.DeduplicationDone, bindMakeLegacyDataUplinkFCnt(34), start, duplicateCount)
				mds = append(mds, msg.RxMetadata...)

				var recentUp *ttnpb.UplinkMessage
				select {
				case <-ctx.Done():
					t.Error("Timed out while waiting for DeviceRegistry.SetByID to be called")
					return false

				case req := <-env.DeviceRegistry.SetByID:
					a.So(req.Context, should.HaveParentContextOrEqual, upCtx)
					a.So(req.ApplicationIdentifiers, should.Resemble, appID)
					a.So(req.DeviceID, should.Resemble, devID)
					a.So(req.Paths, should.HaveSameElementsDeep, dataGetPaths[:])
					dev, sets, err := req.Func(CopyEndDevice(rangeDevice))
					if !a.So(err, should.BeNil) || !a.So(dev, should.NotBeNil) {
						return false
					}
					a.So(sets, should.HaveSameElementsDeep, []string{
						"mac_state",
						"pending_mac_state",
						"pending_session",
						"recent_adr_uplinks",
						"recent_uplinks",
						"session",
					})

					macState := MakeDefaultEU868MACState(ttnpb.CLASS_A, ttnpb.MAC_V1_0_2)
					macState.RxWindowsAvailable = true
					macState.QueuedResponses = []*ttnpb.MACCommand{
						MakeLinkCheckAns(mds...),
					}
					a.So(dev.MACState, should.Resemble, macState)
					a.So(dev.PendingMACState, should.BeNil)
					a.So(dev.PendingSession, should.BeNil)
					a.So(dev.RecentADRUplinks, should.BeNil)
					a.So(dev.Session, should.Resemble, makeSession(ttnpb.MAC_V1_0_2, devAddr, 34))
					if a.So(dev.RecentUplinks, should.NotBeEmpty) {
						recentUp = dev.RecentUplinks[len(dev.RecentUplinks)-1]
						a.So([]time.Time{start, recentUp.ReceivedAt, time.Now()}, should.BeChronological)
						a.So(recentUp.RxMetadata, should.HaveSameElementsDiff, mds)
						expectedUp := makeLegacyDataUplink(34, true)
						expectedUp.CorrelationIDs = upCorrelationIDs
						expectedUp.DeviceChannelIndex = 1
						expectedUp.ReceivedAt = recentUp.ReceivedAt
						expectedUp.RxMetadata = recentUp.RxMetadata
						expectedUp.Settings.DataRateIndex = ttnpb.DATA_RATE_2
						a.So(dev.RecentUplinks, should.HaveEmptyDiff, append(CopyUplinkMessages(rangeDevice.RecentUplinks...), expectedUp))
					}
					req.Response <- DeviceRegistrySetByIDResponse{
						Device: &ttnpb.EndDevice{
							EndDeviceIdentifiers: *makeOTAAIdentifiers(&devAddr),
							FrequencyPlanID:      test.EUFrequencyPlanID,
							LoRaWANPHYVersion:    ttnpb.PHY_V1_0_2_REV_B,
							LoRaWANVersion:       ttnpb.MAC_V1_0_2,
							MACState:             macState,
							RecentUplinks: []*ttnpb.UplinkMessage{
								makeLegacyDataUplink(31, true),
								makeLegacyDataUplink(32, true),
								makeLegacyDataUplink(34, true),
							},
							Session:   makeSession(ttnpb.MAC_V1_0_2, devAddr, 34),
							CreatedAt: start,
							UpdatedAt: time.Now(),
						},
					}
				}

				if !a.So(AssertDownlinkTaskAddRequest(ctx, env.DownlinkTasks.Add, func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, startAt time.Time, replace bool) bool {
					return a.So(ctx, should.HaveParentContextOrEqual, upCtx) &&
						a.So(ids, should.Resemble, *makeOTAAIdentifiers(&devAddr)) &&
						a.So(startAt, should.Resemble, recentUp.ReceivedAt.Add(time.Second-NSScheduleWindow())) &&
						a.So(replace, should.BeTrue)
				},
					nil,
				), should.BeTrue) {
					return false
				}

				if !a.So(test.AssertEventPubSubPublishRequest(ctx, env.Events, func(ev events.Event) bool {
					return a.So(ev, should.ResembleEvent, EvtMergeMetadata(upCtx, rangeDevice.EndDeviceIdentifiers, len(mds)))
				}), should.BeTrue) {
					return false
				}

				if !a.So(test.AssertEventPubSubPublishRequest(ctx, env.Events, func(ev events.Event) bool {
					return a.So(ev, should.ResembleEvent, EvtReceiveLinkCheckRequest(upCtx, rangeDevice.EndDeviceIdentifiers, nil))
				}), should.BeTrue) {
					return false
				}

				if !a.So(test.AssertEventPubSubPublishRequest(ctx, env.Events, func(ev events.Event) bool {
					return a.So(ev, should.ResembleEvent, EvtEnqueueLinkCheckAnswer(upCtx, rangeDevice.EndDeviceIdentifiers, MakeLinkCheckAns(mds...).GetLinkCheckAns()))
				}), should.BeTrue) {
					return false
				}

				if !a.So(test.AssertEventPubSubPublishRequest(ctx, env.Events, func(ev events.Event) bool {
					return a.So(ev, should.ResembleEvent, EvtForwardDataUplink(upCtx, rangeDevice.EndDeviceIdentifiers, nil))
				}), should.BeTrue) {
					return false
				}

				_ = sendUplinkDuplicates(ctx, handle, env.CollectionDone, func(decoded bool) *ttnpb.UplinkMessage {
					msg := makeLegacyDataUplink(34, decoded)
					if !decoded {
						return msg
					}
					msg.DeviceChannelIndex = 1
					msg.Settings.DataRateIndex = ttnpb.DATA_RATE_2
					return msg
				}, start, duplicateCount)

				if !assertHandleUplinkResponse(ctx, handleUplinkErrCh, func(err error) bool {
					return a.So(err, should.BeNil)
				}) {
					return false
				}

				if asRecvCh != nil {
					select {
					case <-ctx.Done():
						t.Error("Timed out while waiting for NetworkServer.handleASUplink to be called")
						return false

					case req := <-asRecvCh:
						a.So(req.Uplink, should.Resemble, &ttnpb.ApplicationUp{
							CorrelationIDs:       upCorrelationIDs,
							EndDeviceIdentifiers: *makeOTAAIdentifiers(&devAddr),
							Up: &ttnpb.ApplicationUp_UplinkMessage{UplinkMessage: &ttnpb.ApplicationUplink{
								SessionKeyID: makeSessionKeys(ttnpb.MAC_V1_1).SessionKeyID,
								FPort:        fPort,
								FCnt:         34,
								FRMPayload:   makeDataUplinkFRMPayload(34),
								RxMetadata:   recentUp.RxMetadata,
								Settings: ttnpb.TxSettings{
									DataRateIndex: ttnpb.DATA_RATE_2,
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
							}},
						})
						req.Response <- ttnpb.Empty
					}
				}
				return true
			},
		},

		{
			Name: "Data uplink/Matching device/No concurrent update/1.1/First transmission/No ADR/Set success/Downlink add success",
			Handler: func(ctx context.Context, env TestEnvironment, asRecvCh <-chan AsNsLinkRecvRequest, handle func(context.Context, *ttnpb.UplinkMessage) <-chan error) bool {
				t := test.MustTFromContext(ctx)
				a := assertions.New(t)

				start := time.Now()

				msg := makeDataUplink(34, false)

				handleUplinkErrCh := handle(ctx, msg)

				rangeDevice := &ttnpb.EndDevice{
					EndDeviceIdentifiers: *makeOTAAIdentifiers(&devAddr),
					FrequencyPlanID:      test.EUFrequencyPlanID,
					LoRaWANPHYVersion:    ttnpb.PHY_V1_1_REV_B,
					LoRaWANVersion:       ttnpb.MAC_V1_1,
					MACState:             MakeDefaultEU868MACState(ttnpb.CLASS_A, ttnpb.MAC_V1_1),
					RecentUplinks: []*ttnpb.UplinkMessage{
						makeDataUplink(31, true),
						makeDataUplink(32, true),
					},
					Session:   makeSession(ttnpb.MAC_V1_1, devAddr, 32),
					CreatedAt: start,
					UpdatedAt: time.Now(),
				}

				var upCtx context.Context
				var upCorrelationIDs []string
				select {
				case <-ctx.Done():
					t.Error("Timed out while waiting for DeviceRegistry.RangeByAddr to be called")
					return false

				case req := <-env.DeviceRegistry.RangeByAddr:
					upCtx = req.Context
					upCorrelationIDs = events.CorrelationIDsFromContext(req.Context)
					for _, id := range correlationIDs {
						a.So(upCorrelationIDs, should.Contain, id)
					}
					a.So(upCorrelationIDs, should.HaveLength, len(correlationIDs)+2)
					a.So(req.DevAddr, should.Resemble, devAddr)
					a.So(req.Paths, should.HaveSameElementsDeep, dataGetPaths[:])
					a.So(req.Func(CopyEndDevice(rangeDevice)), should.BeTrue)
					multicastDevice := CopyEndDevice(rangeDevice)
					multicastDevice.EndDeviceIdentifiers.DeviceID += "-multicast"
					multicastDevice.Multicast = true
					a.So(req.Func(multicastDevice), should.BeTrue)
					fCntTooHighDevice := CopyEndDevice(rangeDevice)
					fCntTooHighDevice.EndDeviceIdentifiers.DeviceID += "-too-high"
					fCntTooHighDevice.RecentUplinks = append(fCntTooHighDevice.RecentUplinks, makeLegacyDataUplink(42, true))
					fCntTooHighDevice.Session.LastFCntUp = 42
					a.So(req.Func(fCntTooHighDevice), should.BeTrue)
					req.Response <- nil
				}

				mds := sendUplinkDuplicates(ctx, handle, env.DeduplicationDone, bindMakeDataUplinkFCnt(34), start, duplicateCount)
				mds = append(mds, msg.RxMetadata...)

				var recentUp *ttnpb.UplinkMessage
				select {
				case <-ctx.Done():
					t.Error("Timed out while waiting for DeviceRegistry.SetByID to be called")
					return false

				case req := <-env.DeviceRegistry.SetByID:
					a.So(req.Context, should.HaveParentContextOrEqual, upCtx)
					a.So(req.ApplicationIdentifiers, should.Resemble, appID)
					a.So(req.DeviceID, should.Resemble, devID)
					a.So(req.Paths, should.HaveSameElementsDeep, dataGetPaths[:])
					dev, sets, err := req.Func(CopyEndDevice(rangeDevice))
					if !a.So(err, should.BeNil) || !a.So(dev, should.NotBeNil) {
						return false
					}
					a.So(sets, should.HaveSameElementsDeep, []string{
						"mac_state",
						"pending_mac_state",
						"pending_session",
						"recent_adr_uplinks",
						"recent_uplinks",
						"session",
					})

					macState := MakeDefaultEU868MACState(ttnpb.CLASS_A, ttnpb.MAC_V1_1)
					macState.RxWindowsAvailable = true
					macState.QueuedResponses = []*ttnpb.MACCommand{
						MakeLinkCheckAns(mds...),
					}
					a.So(dev.MACState, should.Resemble, macState)
					a.So(dev.PendingMACState, should.BeNil)
					a.So(dev.PendingSession, should.BeNil)
					a.So(dev.RecentADRUplinks, should.BeNil)
					a.So(dev.Session, should.Resemble, makeSession(ttnpb.MAC_V1_1, devAddr, 34))
					if a.So(dev.RecentUplinks, should.NotBeEmpty) {
						recentUp = dev.RecentUplinks[len(dev.RecentUplinks)-1]
						a.So([]time.Time{start, recentUp.ReceivedAt, time.Now()}, should.BeChronological)
						a.So(recentUp.RxMetadata, should.HaveSameElementsDiff, mds)
						expectedUp := makeDataUplink(34, true)
						expectedUp.CorrelationIDs = upCorrelationIDs
						expectedUp.DeviceChannelIndex = 1
						expectedUp.ReceivedAt = recentUp.ReceivedAt
						expectedUp.RxMetadata = recentUp.RxMetadata
						expectedUp.Settings.DataRateIndex = ttnpb.DATA_RATE_2
						a.So(dev.RecentUplinks, should.HaveEmptyDiff, append(CopyUplinkMessages(rangeDevice.RecentUplinks...), expectedUp))
					}
					req.Response <- DeviceRegistrySetByIDResponse{
						Device: &ttnpb.EndDevice{
							EndDeviceIdentifiers: *makeOTAAIdentifiers(&devAddr),
							FrequencyPlanID:      test.EUFrequencyPlanID,
							LoRaWANPHYVersion:    ttnpb.PHY_V1_1_REV_B,
							LoRaWANVersion:       ttnpb.MAC_V1_1,
							MACState:             macState,
							RecentUplinks: []*ttnpb.UplinkMessage{
								makeDataUplink(31, true),
								makeDataUplink(32, true),
								makeDataUplink(34, true),
							},
							Session:   makeSession(ttnpb.MAC_V1_1, devAddr, 34),
							CreatedAt: start,
							UpdatedAt: time.Now(),
						},
					}
				}

				if !a.So(AssertDownlinkTaskAddRequest(ctx, env.DownlinkTasks.Add, func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, startAt time.Time, replace bool) bool {
					return a.So(ctx, should.HaveParentContextOrEqual, upCtx) &&
						a.So(ids, should.Resemble, *makeOTAAIdentifiers(&devAddr)) &&
						a.So(startAt, should.Resemble, recentUp.ReceivedAt.Add(time.Second-NSScheduleWindow())) &&
						a.So(replace, should.BeTrue)
				},
					nil,
				), should.BeTrue) {
					return false
				}

				if !a.So(test.AssertEventPubSubPublishRequest(ctx, env.Events, func(ev events.Event) bool {
					return a.So(ev, should.ResembleEvent, EvtMergeMetadata(upCtx, rangeDevice.EndDeviceIdentifiers, len(mds)))
				}), should.BeTrue) {
					return false
				}

				if !a.So(test.AssertEventPubSubPublishRequest(ctx, env.Events, func(ev events.Event) bool {
					return a.So(ev, should.ResembleEvent, EvtReceiveLinkCheckRequest(upCtx, rangeDevice.EndDeviceIdentifiers, nil))
				}), should.BeTrue) {
					return false
				}

				if !a.So(test.AssertEventPubSubPublishRequest(ctx, env.Events, func(ev events.Event) bool {
					return a.So(ev, should.ResembleEvent, EvtEnqueueLinkCheckAnswer(upCtx, rangeDevice.EndDeviceIdentifiers, MakeLinkCheckAns(mds...).GetLinkCheckAns()))
				}), should.BeTrue) {
					return false
				}

				if !a.So(test.AssertEventPubSubPublishRequest(ctx, env.Events, func(ev events.Event) bool {
					return a.So(ev, should.ResembleEvent, EvtForwardDataUplink(upCtx, rangeDevice.EndDeviceIdentifiers, nil))
				}), should.BeTrue) {
					return false
				}

				_ = sendUplinkDuplicates(ctx, handle, env.CollectionDone, func(decoded bool) *ttnpb.UplinkMessage {
					msg := makeDataUplink(34, decoded)
					if !decoded {
						return msg
					}
					msg.DeviceChannelIndex = 1
					msg.Settings.DataRateIndex = ttnpb.DATA_RATE_2
					return msg
				}, start, duplicateCount)

				if !assertHandleUplinkResponse(ctx, handleUplinkErrCh, func(err error) bool {
					return a.So(err, should.BeNil)
				}) {
					return false
				}

				if asRecvCh != nil {
					select {
					case <-ctx.Done():
						t.Error("Timed out while waiting for NetworkServer.handleASUplink to be called")
						return false

					case req := <-asRecvCh:
						a.So(req.Uplink, should.Resemble, &ttnpb.ApplicationUp{
							CorrelationIDs:       upCorrelationIDs,
							EndDeviceIdentifiers: *makeOTAAIdentifiers(&devAddr),
							Up: &ttnpb.ApplicationUp_UplinkMessage{UplinkMessage: &ttnpb.ApplicationUplink{
								SessionKeyID: makeSessionKeys(ttnpb.MAC_V1_1).SessionKeyID,
								FPort:        fPort,
								FCnt:         34,
								FRMPayload:   makeDataUplinkFRMPayload(34),
								RxMetadata:   recentUp.RxMetadata,
								Settings: ttnpb.TxSettings{
									DataRateIndex: ttnpb.DATA_RATE_2,
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
							}},
						})
						req.Response <- ttnpb.Empty
					}
				}
				return true
			},
		},

		{
			Name: "Data uplink/Matching device/Concurrent update/1.0.2/First transmission/No ADR/Set success/Downlink add success",
			Handler: func(ctx context.Context, env TestEnvironment, asRecvCh <-chan AsNsLinkRecvRequest, handle func(context.Context, *ttnpb.UplinkMessage) <-chan error) bool {
				t := test.MustTFromContext(ctx)
				a := assertions.New(t)

				start := time.Now()

				msg := makeLegacyDataUplink(34, false)

				handleUplinkErrCh := handle(ctx, msg)

				rangeDevice := &ttnpb.EndDevice{
					EndDeviceIdentifiers: *makeOTAAIdentifiers(&devAddr),
					FrequencyPlanID:      test.EUFrequencyPlanID,
					LoRaWANPHYVersion:    ttnpb.PHY_V1_0_2_REV_B,
					LoRaWANVersion:       ttnpb.MAC_V1_0_2,
					MACState:             MakeDefaultEU868MACState(ttnpb.CLASS_A, ttnpb.MAC_V1_0_2),
					RecentUplinks: []*ttnpb.UplinkMessage{
						makeLegacyDataUplink(31, true),
						makeLegacyDataUplink(32, true),
					},
					Session:   makeSession(ttnpb.MAC_V1_0_2, devAddr, 32),
					CreatedAt: start,
					UpdatedAt: time.Now(),
				}

				var upCtx context.Context
				var upCorrelationIDs []string
				select {
				case <-ctx.Done():
					t.Error("Timed out while waiting for DeviceRegistry.RangeByAddr to be called")
					return false

				case req := <-env.DeviceRegistry.RangeByAddr:
					upCtx = req.Context
					upCorrelationIDs = events.CorrelationIDsFromContext(req.Context)
					for _, id := range correlationIDs {
						a.So(upCorrelationIDs, should.Contain, id)
					}
					a.So(upCorrelationIDs, should.HaveLength, len(correlationIDs)+2)
					a.So(req.DevAddr, should.Resemble, devAddr)
					a.So(req.Paths, should.HaveSameElementsDeep, dataGetPaths[:])
					a.So(req.Func(CopyEndDevice(rangeDevice)), should.BeTrue)
					multicastDevice := CopyEndDevice(rangeDevice)
					multicastDevice.EndDeviceIdentifiers.DeviceID += "-multicast"
					multicastDevice.Multicast = true
					a.So(req.Func(multicastDevice), should.BeTrue)
					fCntTooHighDevice := CopyEndDevice(rangeDevice)
					fCntTooHighDevice.EndDeviceIdentifiers.DeviceID += "-too-high"
					fCntTooHighDevice.RecentUplinks = append(fCntTooHighDevice.RecentUplinks, makeLegacyDataUplink(42, true))
					fCntTooHighDevice.Session.LastFCntUp = 42
					a.So(req.Func(fCntTooHighDevice), should.BeTrue)
					req.Response <- nil
				}

				mds := sendUplinkDuplicates(ctx, handle, env.DeduplicationDone, bindMakeLegacyDataUplinkFCnt(34), start, duplicateCount)
				mds = append(mds, msg.RxMetadata...)

				var recentUp *ttnpb.UplinkMessage
				select {
				case <-ctx.Done():
					t.Error("Timed out while waiting for DeviceRegistry.SetByID to be called")
					return false

				case req := <-env.DeviceRegistry.SetByID:
					a.So(req.Context, should.HaveParentContextOrEqual, upCtx)
					a.So(req.ApplicationIdentifiers, should.Resemble, appID)
					a.So(req.DeviceID, should.Resemble, devID)
					a.So(req.Paths, should.HaveSameElementsDeep, dataGetPaths[:])
					updatedDevice := CopyEndDevice(rangeDevice)
					updatedDevice.UpdatedAt = time.Now()
					dev, sets, err := req.Func(updatedDevice)
					if !a.So(err, should.BeNil) || !a.So(dev, should.NotBeNil) {
						return false
					}
					a.So(sets, should.HaveSameElementsDeep, []string{
						"mac_state",
						"pending_mac_state",
						"pending_session",
						"recent_adr_uplinks",
						"recent_uplinks",
						"session",
					})

					macState := MakeDefaultEU868MACState(ttnpb.CLASS_A, ttnpb.MAC_V1_0_2)
					macState.RxWindowsAvailable = true
					macState.QueuedResponses = []*ttnpb.MACCommand{
						MakeLinkCheckAns(mds...),
					}
					a.So(dev.MACState, should.Resemble, macState)
					a.So(dev.PendingMACState, should.BeNil)
					a.So(dev.PendingSession, should.BeNil)
					a.So(dev.RecentADRUplinks, should.BeNil)
					a.So(dev.Session, should.Resemble, makeSession(ttnpb.MAC_V1_0_2, devAddr, 34))
					if a.So(dev.RecentUplinks, should.NotBeEmpty) {
						recentUp = dev.RecentUplinks[len(dev.RecentUplinks)-1]
						a.So([]time.Time{start, recentUp.ReceivedAt, time.Now()}, should.BeChronological)
						a.So(recentUp.RxMetadata, should.HaveSameElementsDiff, mds)
						expectedUp := makeLegacyDataUplink(34, true)
						expectedUp.CorrelationIDs = upCorrelationIDs
						expectedUp.DeviceChannelIndex = 1
						expectedUp.ReceivedAt = recentUp.ReceivedAt
						expectedUp.RxMetadata = recentUp.RxMetadata
						expectedUp.Settings.DataRateIndex = ttnpb.DATA_RATE_2
						a.So(dev.RecentUplinks, should.HaveEmptyDiff, append(CopyUplinkMessages(rangeDevice.RecentUplinks...), expectedUp))
					}
					req.Response <- DeviceRegistrySetByIDResponse{
						Device: &ttnpb.EndDevice{
							EndDeviceIdentifiers: *makeOTAAIdentifiers(&devAddr),
							FrequencyPlanID:      test.EUFrequencyPlanID,
							LoRaWANPHYVersion:    ttnpb.PHY_V1_0_2_REV_B,
							LoRaWANVersion:       ttnpb.MAC_V1_0_2,
							MACState:             macState,
							RecentUplinks: []*ttnpb.UplinkMessage{
								makeLegacyDataUplink(31, true),
								makeLegacyDataUplink(32, true),
								makeLegacyDataUplink(34, true),
							},
							Session:   makeSession(ttnpb.MAC_V1_0_2, devAddr, 34),
							CreatedAt: start,
							UpdatedAt: time.Now(),
						},
					}
				}

				if !a.So(AssertDownlinkTaskAddRequest(ctx, env.DownlinkTasks.Add, func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, startAt time.Time, replace bool) bool {
					return a.So(ctx, should.HaveParentContextOrEqual, upCtx) &&
						a.So(ids, should.Resemble, *makeOTAAIdentifiers(&devAddr)) &&
						a.So(startAt, should.Resemble, recentUp.ReceivedAt.Add(time.Second-NSScheduleWindow())) &&
						a.So(replace, should.BeTrue)
				},
					nil,
				), should.BeTrue) {
					return false
				}

				if !a.So(test.AssertEventPubSubPublishRequest(ctx, env.Events, func(ev events.Event) bool {
					return a.So(ev, should.ResembleEvent, EvtMergeMetadata(upCtx, rangeDevice.EndDeviceIdentifiers, len(mds)))
				}), should.BeTrue) {
					return false
				}

				if !a.So(test.AssertEventPubSubPublishRequest(ctx, env.Events, func(ev events.Event) bool {
					return a.So(ev, should.ResembleEvent, EvtReceiveLinkCheckRequest(upCtx, rangeDevice.EndDeviceIdentifiers, nil))
				}), should.BeTrue) {
					return false
				}

				if !a.So(test.AssertEventPubSubPublishRequest(ctx, env.Events, func(ev events.Event) bool {
					return a.So(ev, should.ResembleEvent, EvtEnqueueLinkCheckAnswer(upCtx, rangeDevice.EndDeviceIdentifiers, MakeLinkCheckAns(mds...).GetLinkCheckAns()))
				}), should.BeTrue) {
					return false
				}

				if !a.So(test.AssertEventPubSubPublishRequest(ctx, env.Events, func(ev events.Event) bool {
					return a.So(ev, should.ResembleEvent, EvtForwardDataUplink(upCtx, rangeDevice.EndDeviceIdentifiers, nil))
				}), should.BeTrue) {
					return false
				}

				_ = sendUplinkDuplicates(ctx, handle, env.CollectionDone, func(decoded bool) *ttnpb.UplinkMessage {
					msg := makeLegacyDataUplink(34, decoded)
					if !decoded {
						return msg
					}
					msg.DeviceChannelIndex = 1
					msg.Settings.DataRateIndex = ttnpb.DATA_RATE_2
					return msg
				}, start, duplicateCount)

				if !assertHandleUplinkResponse(ctx, handleUplinkErrCh, func(err error) bool {
					return a.So(err, should.BeNil)
				}) {
					return false
				}

				if asRecvCh != nil {
					select {
					case <-ctx.Done():
						t.Error("Timed out while waiting for NetworkServer.handleASUplink to be called")
						return false

					case req := <-asRecvCh:
						a.So(req.Uplink, should.Resemble, &ttnpb.ApplicationUp{
							CorrelationIDs:       upCorrelationIDs,
							EndDeviceIdentifiers: *makeOTAAIdentifiers(&devAddr),
							Up: &ttnpb.ApplicationUp_UplinkMessage{UplinkMessage: &ttnpb.ApplicationUplink{
								SessionKeyID: makeSessionKeys(ttnpb.MAC_V1_1).SessionKeyID,
								FPort:        fPort,
								FCnt:         34,
								FRMPayload:   makeDataUplinkFRMPayload(34),
								RxMetadata:   recentUp.RxMetadata,
								Settings: ttnpb.TxSettings{
									DataRateIndex: ttnpb.DATA_RATE_2,
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
							}},
						})
						req.Response <- ttnpb.Empty
					}
				}
				return true
			},
		},

		{
			Name: "Data uplink/Matching device/No concurrent update/1.0.2/Second transmission/No ADR/Set success/Downlink add success",
			Handler: func(ctx context.Context, env TestEnvironment, asRecvCh <-chan AsNsLinkRecvRequest, handle func(context.Context, *ttnpb.UplinkMessage) <-chan error) bool {
				t := test.MustTFromContext(ctx)
				a := assertions.New(t)

				start := time.Now()

				msg := makeLegacyDataUplink(34, false)

				handleUplinkErrCh := handle(ctx, msg)

				makeMACState := func() *ttnpb.MACState {
					macState := MakeDefaultEU868MACState(ttnpb.CLASS_A, ttnpb.MAC_V1_0_2)
					macState.CurrentParameters.ADRNbTrans = 2
					macState.DesiredParameters.ADRNbTrans = 2
					return macState
				}

				rangeDevice := &ttnpb.EndDevice{
					EndDeviceIdentifiers: *makeOTAAIdentifiers(&devAddr),
					FrequencyPlanID:      test.EUFrequencyPlanID,
					LoRaWANPHYVersion:    ttnpb.PHY_V1_0_2_REV_B,
					LoRaWANVersion:       ttnpb.MAC_V1_0_2,
					MACState:             makeMACState(),
					RecentUplinks: []*ttnpb.UplinkMessage{
						makeLegacyDataUplink(31, true),
						makeLegacyDataUplink(32, true),
						makeLegacyDataUplink(34, true),
					},
					Session:   makeSession(ttnpb.MAC_V1_0_2, devAddr, 34),
					CreatedAt: start,
					UpdatedAt: time.Now(),
				}

				var upCtx context.Context
				var upCorrelationIDs []string
				select {
				case <-ctx.Done():
					t.Error("Timed out while waiting for DeviceRegistry.RangeByAddr to be called")
					return false

				case req := <-env.DeviceRegistry.RangeByAddr:
					upCtx = req.Context
					upCorrelationIDs = events.CorrelationIDsFromContext(req.Context)
					for _, id := range correlationIDs {
						a.So(upCorrelationIDs, should.Contain, id)
					}
					a.So(upCorrelationIDs, should.HaveLength, len(correlationIDs)+2)
					a.So(req.DevAddr, should.Resemble, devAddr)
					a.So(req.Paths, should.HaveSameElementsDeep, dataGetPaths[:])
					a.So(req.Func(CopyEndDevice(rangeDevice)), should.BeTrue)
					multicastDevice := CopyEndDevice(rangeDevice)
					multicastDevice.EndDeviceIdentifiers.DeviceID += "-multicast"
					multicastDevice.Multicast = true
					a.So(req.Func(multicastDevice), should.BeTrue)
					fCntTooHighDevice := CopyEndDevice(rangeDevice)
					fCntTooHighDevice.EndDeviceIdentifiers.DeviceID += "-too-high"
					fCntTooHighDevice.RecentUplinks = append(fCntTooHighDevice.RecentUplinks, makeLegacyDataUplink(42, true))
					fCntTooHighDevice.Session.LastFCntUp = 42
					a.So(req.Func(fCntTooHighDevice), should.BeTrue)
					req.Response <- nil
				}

				mds := sendUplinkDuplicates(ctx, handle, env.DeduplicationDone, bindMakeLegacyDataUplinkFCnt(34), start, duplicateCount)
				mds = append(mds, msg.RxMetadata...)

				var recentUp *ttnpb.UplinkMessage
				select {
				case <-ctx.Done():
					t.Error("Timed out while waiting for DeviceRegistry.SetByID to be called")
					return false

				case req := <-env.DeviceRegistry.SetByID:
					a.So(req.Context, should.HaveParentContextOrEqual, upCtx)
					a.So(req.ApplicationIdentifiers, should.Resemble, appID)
					a.So(req.DeviceID, should.Resemble, devID)
					a.So(req.Paths, should.HaveSameElementsDeep, dataGetPaths[:])
					dev, sets, err := req.Func(CopyEndDevice(rangeDevice))
					if !a.So(err, should.BeNil) || !a.So(dev, should.NotBeNil) {
						return false
					}
					a.So(sets, should.HaveSameElementsDeep, []string{
						"mac_state",
						"pending_mac_state",
						"pending_session",
						"recent_adr_uplinks",
						"recent_uplinks",
						"session",
					})

					macState := makeMACState()
					macState.RxWindowsAvailable = true
					a.So(dev.MACState, should.Resemble, macState)
					a.So(dev.PendingMACState, should.BeNil)
					a.So(dev.PendingSession, should.BeNil)
					a.So(dev.RecentADRUplinks, should.BeNil)
					a.So(dev.Session, should.Resemble, makeSession(ttnpb.MAC_V1_0_2, devAddr, 34))
					if a.So(dev.RecentUplinks, should.NotBeEmpty) {
						recentUp = dev.RecentUplinks[len(dev.RecentUplinks)-1]
						a.So([]time.Time{start, recentUp.ReceivedAt, time.Now()}, should.BeChronological)
						a.So(recentUp.RxMetadata, should.HaveSameElementsDiff, mds)
						expectedUp := makeLegacyDataUplink(34, true)
						expectedUp.CorrelationIDs = upCorrelationIDs
						expectedUp.DeviceChannelIndex = 1
						expectedUp.ReceivedAt = recentUp.ReceivedAt
						expectedUp.RxMetadata = recentUp.RxMetadata
						expectedUp.Settings.DataRateIndex = ttnpb.DATA_RATE_2
						a.So(dev.RecentUplinks, should.HaveEmptyDiff, append(CopyUplinkMessages(rangeDevice.RecentUplinks...), expectedUp))
					}
					req.Response <- DeviceRegistrySetByIDResponse{
						Device: &ttnpb.EndDevice{
							EndDeviceIdentifiers: *makeOTAAIdentifiers(&devAddr),
							FrequencyPlanID:      test.EUFrequencyPlanID,
							LoRaWANPHYVersion:    ttnpb.PHY_V1_0_2_REV_B,
							LoRaWANVersion:       ttnpb.MAC_V1_0_2,
							MACState:             macState,
							RecentUplinks: []*ttnpb.UplinkMessage{
								makeLegacyDataUplink(31, true),
								makeLegacyDataUplink(32, true),
								makeLegacyDataUplink(34, true),
								makeLegacyDataUplink(34, true),
							},
							Session:   makeSession(ttnpb.MAC_V1_0_2, devAddr, 34),
							CreatedAt: start,
							UpdatedAt: time.Now(),
						},
					}
				}

				if !a.So(AssertDownlinkTaskAddRequest(ctx, env.DownlinkTasks.Add, func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, startAt time.Time, replace bool) bool {
					return a.So(ctx, should.HaveParentContextOrEqual, upCtx) &&
						a.So(ids, should.Resemble, *makeOTAAIdentifiers(&devAddr)) &&
						a.So(startAt, should.Resemble, recentUp.ReceivedAt.Add(time.Second-NSScheduleWindow())) &&
						a.So(replace, should.BeTrue)
				},
					nil,
				), should.BeTrue) {
					return false
				}

				if !a.So(test.AssertEventPubSubPublishRequest(ctx, env.Events, func(ev events.Event) bool {
					return a.So(ev, should.ResembleEvent, EvtMergeMetadata(upCtx, rangeDevice.EndDeviceIdentifiers, len(mds)))
				}), should.BeTrue) {
					return false
				}

				_ = sendUplinkDuplicates(ctx, handle, env.CollectionDone, func(decoded bool) *ttnpb.UplinkMessage {
					msg := makeLegacyDataUplink(34, decoded)
					if !decoded {
						return msg
					}
					msg.DeviceChannelIndex = 1
					msg.Settings.DataRateIndex = ttnpb.DATA_RATE_2
					return msg
				}, start, duplicateCount)

				return assertHandleUplinkResponse(ctx, handleUplinkErrCh, func(err error) bool {
					return a.So(err, should.BeNil)
				})
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			handleTest := func(ctx context.Context, ns *NetworkServer, env TestEnvironment, asRecvCh <-chan AsNsLinkRecvRequest, stop func()) {
				defer stop()

				<-env.DownlinkTasks.Pop

				if !tc.Handler(ctx, env, asRecvCh, func(ctx context.Context, msg *ttnpb.UplinkMessage) <-chan error {
					ch := make(chan error)
					go func() {
						_, err := ttnpb.NewGsNsClient(ns.LoopbackConn()).HandleUplink(ctx, CopyUplinkMessage(msg))
						ttnErr, ok := errors.From(err)
						if ok {
							ch <- ttnErr
						} else {
							ch <- err
						}
						close(ch)
					}()
					return ch
				}) {
					t.Error("Test handler failed")
				}
			}

			makeConfig := func() Config {
				return Config{
					NetID:              *netID.Copy(&types.NetID{}),
					DefaultMACSettings: MACSettingConfig{},
				}
			}

			timeout := (1 << 12) * test.Delay

			t.Run("no link", func(t *testing.T) {
				ns, ctx, env, stop := StartTest(t, makeConfig(), timeout)
				handleTest(ctx, ns, env, nil, func() {
					defer stop()
					assertions.New(t).So(AssertNetworkServerClose(ctx, ns), should.BeTrue)
				})
			})

			t.Run("active link", func(t *testing.T) {
				ns, ctx, env, stop := StartTest(t, makeConfig(), timeout)

				link, ok := AssertLinkApplication(ctx, ns.LoopbackConn(), env.Cluster.GetPeer, appID)
				if !ok {
					t.Fatal("Failed to link application")
				}

				a := assertions.New(t)

				var evCorrelationIDs []string
				if !a.So(test.AssertEventPubSubPublishRequest(ctx, env.Events, func(ev events.Event) bool {
					evCorrelationIDs = ev.CorrelationIDs()
					return a.So(evCorrelationIDs, should.HaveLength, 1) &&
						a.So(ev, should.ResembleEvent, EvtBeginApplicationLink(events.ContextWithCorrelationID(ctx, evCorrelationIDs...), appID, nil))
				}), should.BeTrue) {
					t.FailNow()
				}

				asRecvCh := make(chan AsNsLinkRecvRequest)
				wg := &sync.WaitGroup{}
				wg.Add(1)
				go func() {
					defer wg.Done()
					for {
						up, err := link.Recv()
						if err != nil {
							t.Logf("Receive on AS link returned error: %v", err)
							close(asRecvCh)
							return
						}

						respCh := make(chan *pbtypes.Empty)
						select {
						case <-ctx.Done():
							t.Error("Timed out while waiting for AS uplink to be processed")
							return
						case asRecvCh <- AsNsLinkRecvRequest{
							Uplink:   up,
							Response: respCh,
						}:
						}

						select {
						case <-ctx.Done():
							t.Error("Timed out while waiting for AS uplink response to be processed")
							return
						case resp := <-respCh:
							if err := link.Send(resp); err != nil {
								t.Logf("Send on the link returned error: %v", err)
							}
						}
					}
				}()
				handleTest(ctx, ns, env, asRecvCh, func() {
					defer stop()

					wg.Add(1)
					go func() {
						defer wg.Done()
						if !a.So(test.AssertEventPubSubPublishRequest(ctx, env.Events, func(ev events.Event) bool {
							return a.So(ev, should.ResembleEvent, EvtEndApplicationLink(events.ContextWithCorrelationID(ctx, evCorrelationIDs...), appID, nil))
						}), should.BeTrue) {
							return
						}
					}()
					a.So(AssertNetworkServerClose(ctx, ns), should.BeTrue)
					wg.Wait() // prevent panic when assertions in goroutines fail
				})
			})
		})
	}
}
