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
	"sync"
	"testing"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	clusterauth "go.thethings.network/lorawan-stack/pkg/auth/cluster"
	"go.thethings.network/lorawan-stack/pkg/cluster"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/config"
	"go.thethings.network/lorawan-stack/pkg/crypto"
	"go.thethings.network/lorawan-stack/pkg/frequencyplans"
	. "go.thethings.network/lorawan-stack/pkg/networkserver"
	"go.thethings.network/lorawan-stack/pkg/networkserver/redis"
	"go.thethings.network/lorawan-stack/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
)

func AssertSetDevice(ctx context.Context, conn *grpc.ClientConn, getPeerCh <-chan test.ClusterGetPeerRequest, appID ttnpb.ApplicationIdentifiers, req *ttnpb.SetEndDeviceRequest) (*ttnpb.EndDevice, bool) {
	t := test.MustTFromContext(ctx)
	t.Helper()

	a := assertions.New(t)

	listRightsCh := make(chan test.ApplicationAccessListRightsRequest)
	defer func() {
		close(listRightsCh)
	}()

	var dev *ttnpb.EndDevice
	var err error
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		dev, err = ttnpb.NewNsEndDeviceRegistryClient(conn).Set(
			ctx,
			req,
			grpc.PerRPCCredentials(rpcmetadata.MD{
				AuthType:      "Bearer",
				AuthValue:     "set-key",
				AllowInsecure: true,
			}),
		)
		wg.Done()
	}()

	if !a.So(test.AssertClusterGetPeerRequest(ctx, getPeerCh,
		func(ctx context.Context, role ttnpb.ClusterRole, ids ttnpb.Identifiers) bool {
			return a.So(role, should.Equal, ttnpb.ClusterRole_ACCESS) && a.So(ids, should.BeNil)
		},
		test.ClusterGetPeerResponse{
			Peer: NewISPeer(ctx, &test.MockApplicationAccessServer{
				ListRightsFunc: test.MakeApplicationAccessListRightsChFunc(listRightsCh),
			}),
		},
	), should.BeTrue) {
		return nil, false
	}

	if !a.So(test.AssertListRightsRequest(ctx, listRightsCh,
		func(ctx context.Context, ids ttnpb.Identifiers) bool {
			md := rpcmetadata.FromIncomingContext(ctx)
			return a.So(md.AuthType, should.Equal, "Bearer") &&
				a.So(md.AuthValue, should.Equal, "set-key") &&
				a.So(ids, should.Resemble, &appID)
		}, ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
	), should.BeTrue) {
		return nil, false
	}

	if !a.So(test.WaitContext(ctx, wg.Wait), should.BeTrue) {
		t.Error("Timed out while waiting for device to be set")
		return nil, false
	}
	return dev, a.So(err, should.BeNil)
}

func handleOTAAClassA868FlowTest1_0_2(ctx context.Context, reg DeviceRegistry, tq DownlinkTaskQueue) {
	t := test.MustTFromContext(ctx)
	a := assertions.New(t)

	authCh := make(chan test.ClusterAuthRequest)
	getPeerCh := make(chan test.ClusterGetPeerRequest)

	netID := test.Must(types.NewNetID(2, []byte{1, 2, 3})).(types.NetID)

	appID := ttnpb.ApplicationIdentifiers{ApplicationID: "flow-test-app-id"}
	devID := "flow-test-dev-id"

	ns := test.Must(New(
		component.MustNew(
			test.GetLogger(t),
			&component.Config{},
			component.WithClusterNew(func(_ context.Context, conf *config.Cluster, options ...cluster.Option) (cluster.Cluster, error) {
				return &test.MockCluster{
					AuthFunc:    test.MakeClusterAuthChFunc(authCh),
					GetPeerFunc: test.MakeClusterGetPeerChFunc(getPeerCh),
					JoinFunc:    test.ClusterJoinNilFunc,
					WithVerifiedSourceFunc: func(ctx context.Context) context.Context {
						return clusterauth.NewContext(ctx, nil)
					},
				}, nil
			}),
		),
		&Config{
			NetID:         netID,
			Devices:       reg,
			DownlinkTasks: tq,
			DownlinkPriorities: DownlinkPriorityConfig{
				JoinAccept:             "highest",
				MACCommands:            "highest",
				MaxApplicationDownlink: "high",
			},
			DefaultMACSettings: MACSettingConfig{
				DesiredRx1Delay: func(v ttnpb.RxDelay) *ttnpb.RxDelay { return &v }(ttnpb.RX_DELAY_6),
			},
			DeduplicationWindow: (1 << 5) * test.Delay,
			CooldownWindow:      (1 << 6) * test.Delay,
		},
	)).(*NetworkServer)
	ns.FrequencyPlans = frequencyplans.NewStore(test.FrequencyPlansFetcher)
	test.Must(nil, ns.Start())
	defer ns.Close()

	conn := ns.LoopbackConn()

	start := time.Now()

	link, ok := AssertLinkApplication(ctx, conn, getPeerCh, appID)
	if !a.So(ok, should.BeTrue) || !a.So(link, should.NotBeNil) {
		t.Error("Failed to link application")
		return
	}

	dev, ok := AssertSetDevice(ctx, conn, getPeerCh, appID, &ttnpb.SetEndDeviceRequest{
		EndDevice: ttnpb.EndDevice{
			EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
				DeviceID:               devID,
				ApplicationIdentifiers: appID,
				JoinEUI:                &types.EUI64{0x42, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
				DevEUI:                 &types.EUI64{0x42, 0x42, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			},
			FrequencyPlanID:   test.EUFrequencyPlanID,
			LoRaWANPHYVersion: ttnpb.PHY_V1_0_2_REV_B,
			LoRaWANVersion:    ttnpb.MAC_V1_0_2,
			SupportsJoin:      true,
		},
		FieldMask: pbtypes.FieldMask{
			Paths: []string{
				"frequency_plan_id",
				"lorawan_phy_version",
				"lorawan_version",
				"supports_join",
			},
		},
	})
	if !a.So(ok, should.BeTrue) || !a.So(dev, should.NotBeNil) {
		t.Error("Failed to create device")
		return
	}
	t.Log("Device created")
	a.So(dev.CreatedAt, should.HappenAfter, start)
	a.So(dev.UpdatedAt, should.Equal, dev.CreatedAt)
	a.So([]time.Time{start, dev.CreatedAt, time.Now()}, should.BeChronological)
	a.So(dev, should.Resemble, &ttnpb.EndDevice{
		EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
			DeviceID:               devID,
			ApplicationIdentifiers: appID,
			JoinEUI:                &types.EUI64{0x42, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			DevEUI:                 &types.EUI64{0x42, 0x42, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		FrequencyPlanID:   test.EUFrequencyPlanID,
		LoRaWANPHYVersion: ttnpb.PHY_V1_0_2_REV_B,
		LoRaWANVersion:    ttnpb.MAC_V1_0_2,
		SupportsJoin:      true,
		CreatedAt:         dev.CreatedAt,
		UpdatedAt:         dev.UpdatedAt,
	})

	scheduleDownlinkCh := make(chan NsGsScheduleDownlinkRequest)
	gsPeer := NewGSPeer(ctx, &MockNsGsServer{
		ScheduleDownlinkFunc: MakeNsGsScheduleDownlinkChFunc(scheduleDownlinkCh),
	})

	handleJoinCh := make(chan NsJsHandleJoinRequest)
	jsPeer := NewJSPeer(ctx, &MockNsJsServer{
		HandleJoinFunc: MakeNsJsHandleJoinChFunc(handleJoinCh),
	})
	gsns := ttnpb.NewGsNsClient(conn)

	appSKey := types.AES128Key{0x42, 0x42, 0x42, 0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	fNwkSIntKey := types.AES128Key{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

	var devAddr types.DevAddr
	if !t.Run("join-request", func(t *testing.T) {
		a := assertions.New(t)

		ctx := test.ContextWithT(ctx, t)
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		payload := []byte{
			/* MHDR */
			0x00,
			/* Join-request */
			/** JoinEUI **/
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x42,
			/** DevEUI **/
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x42, 0x42,
			/** DevNonce **/
			0x01, 0x00,
			/* MIC */
			0x03, 0x02, 0x01, 0x00,
		}

		makeUplink := func(rxMetadata *ttnpb.RxMetadata, correlationIDs ...string) *ttnpb.UplinkMessage {
			return &ttnpb.UplinkMessage{
				RawPayload: payload,
				Settings: ttnpb.TxSettings{
					DataRate: ttnpb.DataRate{
						Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{
							Bandwidth:       125000,
							SpreadingFactor: 12,
						}},
					},
					Frequency: 868100000,
					EnableCRC: true,
					Timestamp: 42,
				},
				RxMetadata: []*ttnpb.RxMetadata{
					rxMetadata,
				},
				ReceivedAt:     time.Now(),
				CorrelationIDs: correlationIDs,
			}
		}

		handleUplinkErrCh := make(chan error)
		go func() {
			_, err := gsns.HandleUplink(ctx, makeUplink(
				&ttnpb.RxMetadata{
					GatewayIdentifiers: ttnpb.GatewayIdentifiers{
						GatewayID: "test-gtw-1",
					},
					SNR:         -1,
					UplinkToken: []byte("join-request-token-1"),
				},
				"GsNs-1", "GsNs-2",
			))
			t.Logf("HandleUplink returned %v", err)
			handleUplinkErrCh <- err
			close(handleUplinkErrCh)
		}()

		defer time.AfterFunc((1<<3)*test.Delay, func() {
			_, err := gsns.HandleUplink(ctx, makeUplink(
				&ttnpb.RxMetadata{
					GatewayIdentifiers: ttnpb.GatewayIdentifiers{
						GatewayID: "test-gtw-2",
					},
					SNR:         -2,
					UplinkToken: []byte("join-request-token-2"),
				},
				"GsNs-1", "GsNs-3",
			))
			t.Logf("Duplicate HandleUplink returned %v", err)
			handleUplinkErrCh <- err
		}).Stop()

		if !a.So(test.AssertClusterGetPeerRequest(ctx, getPeerCh,
			func(ctx context.Context, role ttnpb.ClusterRole, ids ttnpb.Identifiers) bool {
				return a.So(role, should.Equal, ttnpb.ClusterRole_JOIN_SERVER) &&
					a.So(ids, should.Resemble, ttnpb.EndDeviceIdentifiers{
						DeviceID:               devID,
						ApplicationIdentifiers: appID,
						JoinEUI:                &types.EUI64{0x42, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
						DevEUI:                 &types.EUI64{0x42, 0x42, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
					})
			},
			test.ClusterGetPeerResponse{Peer: jsPeer},
		), should.BeTrue) {
			return
		}

		if !a.So(AssertAuthNsJsHandleJoinRequest(ctx, authCh, handleJoinCh, func(ctx context.Context, req *ttnpb.JoinRequest) bool {
			devAddr = req.DevAddr
			return a.So(req.CorrelationIDs, should.Contain, "GsNs-1") &&
				a.So(req.CorrelationIDs, should.Contain, "GsNs-2") &&
				a.So(req.CorrelationIDs, should.HaveLength, 4) &&
				a.So(req.DevAddr, should.NotBeEmpty) &&
				a.So(req.DevAddr.NwkID(), should.Resemble, netID.ID()) &&
				a.So(req.DevAddr.NetIDType(), should.Equal, netID.Type()) &&
				a.So(req, should.Resemble, &ttnpb.JoinRequest{
					Payload: &ttnpb.Message{
						Payload: &ttnpb.Message_JoinRequestPayload{
							JoinRequestPayload: &ttnpb.JoinRequestPayload{
								JoinEUI:  types.EUI64{0x42, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
								DevEUI:   types.EUI64{0x42, 0x42, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
								DevNonce: [2]byte{0x00, 0x01},
							},
						},
						MIC: []byte{3, 2, 1, 0},
					},
					RawPayload:         payload,
					DevAddr:            req.DevAddr,
					SelectedMACVersion: ttnpb.MAC_V1_0_2,
					NetID:              netID,
					RxDelay:            ttnpb.RX_DELAY_6,
					CFList: &ttnpb.CFList{
						Type: ttnpb.CFListType_FREQUENCIES,
						Freq: []uint32{8671000, 8673000, 8675000, 8677000, 8679000},
					},
					CorrelationIDs: req.CorrelationIDs,
				})
		},
			&grpc.EmptyCallOption{},
			NsJsHandleJoinResponse{
				Response: &ttnpb.JoinResponse{
					RawPayload: bytes.Repeat([]byte{0x42}, 33),
					SessionKeys: ttnpb.SessionKeys{
						SessionKeyID: []byte("session-key-id"),
						AppSKey: &ttnpb.KeyEnvelope{
							Key: &appSKey,
						},
						FNwkSIntKey: &ttnpb.KeyEnvelope{
							Key: &fNwkSIntKey,
						},
					},
					CorrelationIDs: []string{"NsJs-1", "NsJs-2"},
				},
			},
		), should.BeTrue) {
			t.Error("Join-request send assertion failed")
			return
		}

		select {
		case <-ctx.Done():
			t.Error("Timed out while waiting for duplicate HandleUplink to return")
			return

		case err := <-handleUplinkErrCh:
			if !a.So(err, should.BeNil) {
				t.Errorf("Failed to handle duplicate uplink: %s", err)
				return
			}
		}

		var asUp *ttnpb.ApplicationUp
		var err error
		if !a.So(test.WaitContext(ctx, func() {
			asUp, err = link.Recv()
		}), should.BeTrue) {
			t.Error("Timed out while waiting for join-accept to be sent to AS")
			return
		}
		if !a.So(err, should.BeNil) {
			t.Errorf("Failed to receive AS uplink: %s", err)
			return
		}
		a.So(asUp.CorrelationIDs, should.Contain, "GsNs-1")
		a.So(asUp.CorrelationIDs, should.Contain, "GsNs-2")
		a.So(asUp.CorrelationIDs, should.Contain, "NsJs-1")
		a.So(asUp.CorrelationIDs, should.Contain, "NsJs-2")
		a.So(asUp.CorrelationIDs, should.HaveLength, 6)
		a.So(asUp, should.Resemble, &ttnpb.ApplicationUp{
			EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
				DeviceID:               devID,
				ApplicationIdentifiers: appID,
				JoinEUI:                &types.EUI64{0x42, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
				DevEUI:                 &types.EUI64{0x42, 0x42, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
				DevAddr:                &devAddr,
			},
			CorrelationIDs: asUp.CorrelationIDs,
			Up: &ttnpb.ApplicationUp_JoinAccept{JoinAccept: &ttnpb.ApplicationJoinAccept{
				AppSKey: &ttnpb.KeyEnvelope{
					Key: &appSKey,
				},
				SessionKeyID: []byte("session-key-id"),
			}},
		})

		if !a.So(test.WaitContext(ctx, func() {
			err = link.Send(ttnpb.Empty)
		}), should.BeTrue) {
			t.Error("Timed out while waiting for NS to process AS response")
			return
		}
		if !a.So(err, should.BeNil) {
			t.Errorf("Failed to send AS uplink response: %s", err)
			return
		}

		select {
		case err := <-handleUplinkErrCh:
			if !a.So(err, should.BeNil) {
				t.Errorf("Failed to handle uplink: %s", err)
				return
			}

		case <-ctx.Done():
			t.Error("Timed out while waiting for HandleUplink to return")
			return
		}

		if !a.So(test.AssertClusterGetPeerRequest(ctx, getPeerCh,
			func(ctx context.Context, role ttnpb.ClusterRole, ids ttnpb.Identifiers) bool {
				return a.So(role, should.Equal, ttnpb.ClusterRole_GATEWAY_SERVER) &&
					a.So(ids, should.Resemble, ttnpb.GatewayIdentifiers{
						GatewayID: "test-gtw-1",
					})
			},
			test.ClusterGetPeerResponse{Peer: gsPeer},
		), should.BeTrue) {
			return
		}

		if !a.So(test.AssertClusterGetPeerRequest(ctx, getPeerCh,
			func(ctx context.Context, role ttnpb.ClusterRole, ids ttnpb.Identifiers) bool {
				return a.So(role, should.Equal, ttnpb.ClusterRole_GATEWAY_SERVER) &&
					a.So(ids, should.Resemble, ttnpb.GatewayIdentifiers{
						GatewayID: "test-gtw-2",
					})
			},
			test.ClusterGetPeerResponse{Peer: gsPeer},
		), should.BeTrue) {
			return
		}

		a.So(AssertAuthNsGsScheduleDownlinkRequest(ctx, authCh, scheduleDownlinkCh,
			func(ctx context.Context, msg *ttnpb.DownlinkMessage) bool {
				return a.So(msg.CorrelationIDs, should.Contain, "GsNs-1") &&
					a.So(msg.CorrelationIDs, should.Contain, "GsNs-2") &&
					a.So(msg.CorrelationIDs, should.HaveLength, 5) &&
					a.So(msg, should.Resemble, &ttnpb.DownlinkMessage{
						RawPayload: bytes.Repeat([]byte{0x42}, 33),
						Settings: &ttnpb.DownlinkMessage_Request{
							Request: &ttnpb.TxRequest{
								Class: ttnpb.CLASS_A,
								DownlinkPaths: []*ttnpb.DownlinkPath{
									{
										Path: &ttnpb.DownlinkPath_UplinkToken{
											UplinkToken: []byte("join-request-token-1"),
										},
									},
									{
										Path: &ttnpb.DownlinkPath_UplinkToken{
											UplinkToken: []byte("join-request-token-2"),
										},
									},
								},
								Rx1Delay:         ttnpb.RX_DELAY_5,
								Rx1DataRateIndex: ttnpb.DATA_RATE_0,
								Rx1Frequency:     868100000,
								Rx2DataRateIndex: ttnpb.DATA_RATE_0,
								Rx2Frequency:     869525000,
								Priority:         ttnpb.TxSchedulePriority_HIGHEST,
							},
						},
						CorrelationIDs: msg.CorrelationIDs,
					})
			},
			&grpc.EmptyCallOption{},
			NsGsScheduleDownlinkResponse{
				Response: &ttnpb.ScheduleDownlinkResponse{},
			},
		), should.BeTrue)
	}) {
		t.Error("Join-accept schedule assertion failed")
		return
	}

	t.Logf("Device successfully joined. DevAddr: %s", devAddr)

	t.Run("uplink", func(t *testing.T) {
		a := assertions.New(t)

		ctx := test.ContextWithT(ctx, t)
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		uplinkFRMPayload := test.Must(crypto.EncryptUplink(appSKey, devAddr, 0, []byte("test"))).([]byte)

		makeUplink := func(rxMetadata *ttnpb.RxMetadata, correlationIDs ...string) *ttnpb.UplinkMessage {
			return &ttnpb.UplinkMessage{
				RawPayload: MustAppendLegacyUplinkMIC(
					fNwkSIntKey,
					devAddr,
					0,
					append([]byte{
						/* MHDR */
						0x40,
						/* MACPayload */
						/** FHDR **/
						/*** DevAddr ***/
						devAddr[3], devAddr[2], devAddr[1], devAddr[0],
						/*** FCtrl ***/
						0x80,
						/*** FCnt ***/
						0x00, 0x00,
						/** FPort **/
						0x42,
					},
						uplinkFRMPayload...,
					)...,
				),
				Settings: ttnpb.TxSettings{
					DataRate: ttnpb.DataRate{
						Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{
							Bandwidth:       125000,
							SpreadingFactor: 11,
						}},
					},
					EnableCRC: true,
					Frequency: 867100000,
					Timestamp: 42,
				},
				RxMetadata: []*ttnpb.RxMetadata{
					rxMetadata,
				},
				ReceivedAt:     time.Now(),
				CorrelationIDs: correlationIDs,
			}
		}

		mds := [...]*ttnpb.RxMetadata{
			{
				GatewayIdentifiers: ttnpb.GatewayIdentifiers{
					GatewayID: "test-gtw-2",
				},
				SNR:         -3.42,
				UplinkToken: []byte("test-uplink-token-2"),
			},
			{
				GatewayIdentifiers: ttnpb.GatewayIdentifiers{
					GatewayID: "test-gtw-3",
				},
				SNR:         -2.3,
				UplinkToken: []byte("test-uplink-token-3"),
			},
		}

		handleUplinkErrCh := make(chan error)
		go func() {
			_, err := gsns.HandleUplink(ctx, makeUplink(
				mds[0],
				"GsNs-1", "GsNs-2",
			))
			t.Logf("HandleUplink returned %v", err)
			handleUplinkErrCh <- err
			close(handleUplinkErrCh)
		}()

		defer time.AfterFunc((1<<3)*test.Delay, func() {
			_, err := gsns.HandleUplink(ctx, makeUplink(
				mds[1],
				"GsNs-1", "GsNs-3",
			))
			t.Logf("Duplicate HandleUplink returned %v", err)
			handleUplinkErrCh <- err
		}).Stop()

		var asUp *ttnpb.ApplicationUp
		var err error
		if !a.So(test.WaitContext(ctx, func() {
			asUp, err = link.Recv()
		}), should.BeTrue) {
			t.Error("Timed out while waiting for uplink to be sent to AS")
			return
		}
		if !a.So(err, should.BeNil) {
			t.Errorf("Failed to receive AS uplink: %s", err)
			return
		}

		a.So(asUp.GetUplinkMessage().GetRxMetadata(), should.HaveSameElementsDeep, mds)
		a.So(asUp.CorrelationIDs, should.Contain, "GsNs-1")
		a.So(asUp.CorrelationIDs, should.Contain, "GsNs-2")
		a.So(asUp.CorrelationIDs, should.HaveLength, 4)
		a.So(asUp, should.Resemble, &ttnpb.ApplicationUp{
			EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
				DeviceID:               devID,
				ApplicationIdentifiers: appID,
				JoinEUI:                &types.EUI64{0x42, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
				DevEUI:                 &types.EUI64{0x42, 0x42, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
				DevAddr:                &devAddr,
			},
			CorrelationIDs: asUp.CorrelationIDs,
			Up: &ttnpb.ApplicationUp_UplinkMessage{UplinkMessage: &ttnpb.ApplicationUplink{
				SessionKeyID: []byte("session-key-id"),
				FPort:        0x42,
				FRMPayload:   uplinkFRMPayload,
				RxMetadata:   asUp.GetUplinkMessage().GetRxMetadata(),
				Settings: ttnpb.TxSettings{
					DataRate: ttnpb.DataRate{
						Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{
							Bandwidth:       125000,
							SpreadingFactor: 11,
						}},
					},
					DataRateIndex: ttnpb.DATA_RATE_1,
					EnableCRC:     true,
					Frequency:     867100000,
					Timestamp:     42,
				},
			}},
		})

		if !a.So(test.WaitContext(ctx, func() {
			err = link.Send(ttnpb.Empty)
		}), should.BeTrue) {
			t.Error("Timed out while waiting for NS to process AS response")
			return
		}
		if !a.So(err, should.BeNil) {
			t.Errorf("Failed to send AS uplink response: %s", err)
			return
		}

		select {
		case err := <-handleUplinkErrCh:
			if !a.So(err, should.BeNil) {
				t.Errorf("Failed to handle uplink: %s", err)
				return
			}

		case <-ctx.Done():
			t.Error("Timed out while waiting for HandleUplink to return")
			return
		}

		a.So(test.AssertClusterGetPeerRequest(ctx, getPeerCh,
			func(ctx context.Context, role ttnpb.ClusterRole, ids ttnpb.Identifiers) bool {
				return a.So(role, should.Equal, ttnpb.ClusterRole_GATEWAY_SERVER) &&
					a.So(ids, should.Resemble, ttnpb.GatewayIdentifiers{
						GatewayID: "test-gtw-3",
					})
			},
			test.ClusterGetPeerResponse{Peer: gsPeer},
		), should.BeTrue)

		a.So(test.AssertClusterGetPeerRequest(ctx, getPeerCh,
			func(ctx context.Context, role ttnpb.ClusterRole, ids ttnpb.Identifiers) bool {
				return a.So(role, should.Equal, ttnpb.ClusterRole_GATEWAY_SERVER) &&
					a.So(ids, should.Resemble, ttnpb.GatewayIdentifiers{
						GatewayID: "test-gtw-2",
					})
			},
			test.ClusterGetPeerResponse{Peer: gsPeer},
		), should.BeTrue)

		a.So(AssertAuthNsGsScheduleDownlinkRequest(ctx, authCh, scheduleDownlinkCh,
			func(ctx context.Context, msg *ttnpb.DownlinkMessage) bool {
				return a.So(msg.CorrelationIDs, should.Contain, "GsNs-1") &&
					a.So(msg.CorrelationIDs, should.Contain, "GsNs-2") &&
					a.So(msg.CorrelationIDs, should.HaveLength, 5) &&
					a.So(msg, should.Resemble, &ttnpb.DownlinkMessage{
						RawPayload: MustAppendLegacyDownlinkMIC(
							fNwkSIntKey,
							devAddr,
							1,
							append([]byte{
								/* MHDR */
								0x60,
								/* MACPayload */
								/** FHDR **/
								/*** DevAddr ***/
								devAddr[3], devAddr[2], devAddr[1], devAddr[0],
								/*** FCtrl ***/
								0x80,
								/*** FCnt ***/
								0x01, 0x00,
								/** FPort **/
								0x0,
							},
								test.Must(crypto.EncryptDownlink(fNwkSIntKey, devAddr, 1, []byte{
									/* DevStatusReq */
									0x06,
								})).([]byte)...,
							)...,
						),
						Settings: &ttnpb.DownlinkMessage_Request{
							Request: &ttnpb.TxRequest{
								Class: ttnpb.CLASS_A,
								DownlinkPaths: []*ttnpb.DownlinkPath{
									{
										Path: &ttnpb.DownlinkPath_UplinkToken{
											UplinkToken: []byte("test-uplink-token-3"),
										},
									},
									{
										Path: &ttnpb.DownlinkPath_UplinkToken{
											UplinkToken: []byte("test-uplink-token-2"),
										},
									},
								},
								Rx1Delay:         ttnpb.RX_DELAY_6,
								Rx1DataRateIndex: ttnpb.DATA_RATE_1,
								Rx1Frequency:     867100000,
								Rx2DataRateIndex: ttnpb.DATA_RATE_0,
								Rx2Frequency:     869525000,
								Priority:         ttnpb.TxSchedulePriority_HIGHEST,
							},
						},
						CorrelationIDs: msg.CorrelationIDs,
					})
			},
			&grpc.EmptyCallOption{},
			NsGsScheduleDownlinkResponse{
				Response: &ttnpb.ScheduleDownlinkResponse{},
			},
		), should.BeTrue)
	})
}

func TestFlow(t *testing.T) {
	t.Parallel()

	namespace := [...]string{
		"networkserver_test",
	}

	for _, tc := range []struct {
		Name                 string
		NewRegistry          func(t testing.TB) (reg DeviceRegistry, closeFn func() error)
		NewDownlinkTaskQueue func(t testing.TB) (tq DownlinkTaskQueue, closeFn func() error)
	}{
		{
			Name: "Redis registry/Redis downlink task queue",
			NewRegistry: func(t testing.TB) (DeviceRegistry, func() error) {
				cl, flush := test.NewRedis(t, append(namespace[:], "devices")...)
				reg := &redis.DeviceRegistry{Redis: cl}
				return reg, func() error {
					flush()
					return cl.Close()
				}
			},
			NewDownlinkTaskQueue: func(t testing.TB) (DownlinkTaskQueue, func() error) {
				cl, flush := test.NewRedis(t, append(namespace[:], "tasks")...)
				tq := redis.NewDownlinkTaskQueue(cl, 100000, "ns", "test")
				ctx, cancel := context.WithCancel(test.Context())
				errch := make(chan error)
				go func() {
					errch <- tq.Run(ctx)
				}()
				return tq, func() error {
					cancel()
					if err := tq.Add(ctx, ttnpb.EndDeviceIdentifiers{
						DeviceID:               "test",
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test"},
					}, time.Now(), false); err != nil {
						t.Errorf("Failed to add mock device to task queue: %s", err)
						return err
					}
					runErr := <-errch
					flush()
					closeErr := cl.Close()
					if runErr != nil && runErr != context.Canceled {
						return runErr
					}
					return closeErr
				}
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			for flow, handleFlowTest := range map[string]func(context.Context, DeviceRegistry, DownlinkTaskQueue){
				"Class A/OTAA/EU868/1.0.2": handleOTAAClassA868FlowTest1_0_2,
			} {
				t.Run(flow, func(t *testing.T) {
					reg, regClose := tc.NewRegistry(t)
					if regClose != nil {
						defer func() {
							if err := regClose(); err != nil {
								t.Errorf("Failed to close device registry: %s", err)
							}
						}()
					}

					tq, tqClose := tc.NewDownlinkTaskQueue(t)
					if tqClose != nil {
						defer func() {
							if err := tqClose(); err != nil {
								t.Errorf("Failed to close downlink task queue: %s", err)
							}
						}()
					}

					ctx := test.ContextWithT(test.Context(), t)
					ctx, cancel := context.WithTimeout(ctx, (1<<13)*test.Delay)
					defer cancel()
					handleFlowTest(ctx, reg, tq)
				})
			}
		})
	}
}
