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
	"go.thethings.network/lorawan-stack/pkg/rpcserver"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
)

func newISPeer(ctx context.Context, is interface {
	ttnpb.ApplicationAccessServer
}) cluster.Peer {
	return test.Must(test.NewGRPCServerPeer(ctx, is, ttnpb.RegisterApplicationAccessServer)).(cluster.Peer)
}

func newGSPeer(ctx context.Context, gs interface {
	ttnpb.NsGsServer
}) cluster.Peer {
	return test.Must(test.NewGRPCServerPeer(ctx, gs, ttnpb.RegisterNsGsServer)).(cluster.Peer)
}

func newJSPeer(ctx context.Context, js interface {
	ttnpb.NsJsServer
}) cluster.Peer {
	return test.Must(test.NewGRPCServerPeer(ctx, js, ttnpb.RegisterNsJsServer)).(cluster.Peer)
}

func assertLinkApplication(t *testing.T, ctx context.Context, conn *grpc.ClientConn, getPeerCh <-chan test.GetPeerRequest, timeout time.Duration, appID string) (ttnpb.AsNs_LinkApplicationClient, bool) {
	t.Helper()

	a := assertions.New(t)

	listRightsCh := make(chan test.ApplicationAccessListRightsRequest)
	defer func() {
		close(listRightsCh)
	}()

	var link ttnpb.AsNs_LinkApplicationClient
	var err error
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		link, err = ttnpb.NewAsNsClient(conn).LinkApplication(
			(rpcmetadata.MD{
				ID: appID,
			}).ToOutgoingContext(ctx),
			grpc.PerRPCCredentials(rpcmetadata.MD{
				AuthType:      "Bearer",
				AuthValue:     "link-application-key",
				AllowInsecure: true,
			}),
		)
		wg.Done()
	}()

	if !a.So(test.AssertGetPeerRequest(t, getPeerCh, Timeout,
		func(ctx context.Context, role ttnpb.PeerInfo_Role, ids ttnpb.Identifiers) bool {
			return a.So(role, should.Equal, ttnpb.PeerInfo_ACCESS) && a.So(ids, should.BeNil)
		},
		newISPeer(ctx, &test.MockApplicationAccessServer{
			ListRightsFunc: test.MakeApplicationAccessListRightsChFunc(listRightsCh),
		}),
	), should.BeTrue) {
		return nil, false
	}

	if !a.So(test.AssertListRightsRequest(t, listRightsCh, Timeout,
		func(ctx context.Context, ids ttnpb.Identifiers) bool {
			md := rpcmetadata.FromIncomingContext(ctx)
			return a.So(md.AuthType, should.Equal, "Bearer") &&
				a.So(md.AuthValue, should.Equal, "link-application-key") &&
				a.So(ids, should.Resemble, &ttnpb.ApplicationIdentifiers{ApplicationID: appID})
		}, ttnpb.RIGHT_APPLICATION_LINK,
	), should.BeTrue) {
		return nil, false
	}

	if !a.So(test.WaitTimeout(Timeout, wg.Wait), should.BeTrue) {
		t.Error("Timed out while waiting for AS link to open")
		return nil, false
	}
	return link, a.So(err, should.BeNil)
}

func assertSetDevice(t *testing.T, ctx context.Context, conn *grpc.ClientConn, getPeerCh <-chan test.GetPeerRequest, timeout time.Duration, appID string, req *ttnpb.SetEndDeviceRequest) (*ttnpb.EndDevice, bool) {
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

	if !a.So(test.AssertGetPeerRequest(t, getPeerCh, Timeout,
		func(ctx context.Context, role ttnpb.PeerInfo_Role, ids ttnpb.Identifiers) bool {
			return a.So(role, should.Equal, ttnpb.PeerInfo_ACCESS) && a.So(ids, should.BeNil)
		},
		newISPeer(ctx, &test.MockApplicationAccessServer{
			ListRightsFunc: test.MakeApplicationAccessListRightsChFunc(listRightsCh),
		}),
	), should.BeTrue) {
		return nil, false
	}

	if !a.So(test.AssertListRightsRequest(t, listRightsCh, Timeout,
		func(ctx context.Context, ids ttnpb.Identifiers) bool {
			md := rpcmetadata.FromIncomingContext(ctx)
			return a.So(md.AuthType, should.Equal, "Bearer") &&
				a.So(md.AuthValue, should.Equal, "set-key") &&
				a.So(ids, should.Resemble, &ttnpb.ApplicationIdentifiers{ApplicationID: appID})
		}, ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
	), should.BeTrue) {
		return nil, false
	}

	if !a.So(test.WaitTimeout(Timeout, wg.Wait), should.BeTrue) {
		t.Error("Timed out while waiting for device to be created")
		return nil, false
	}
	return dev, a.So(err, should.BeNil)
}

func handleOTAAClassA868FlowTest(t *testing.T, reg DeviceRegistry, tq DownlinkTaskQueue) {
	a := assertions.New(t)

	getPeerCh := make(chan test.GetPeerRequest)

	collectionDoneCh := make(chan WindowEndRequest)
	deduplicationDoneCh := make(chan WindowEndRequest)

	netID := test.Must(types.NewNetID(2, []byte{1, 2, 3})).(types.NetID)

	ns := test.Must(New(
		component.MustNew(
			test.GetLogger(t),
			&component.Config{
				ServiceBase: config.ServiceBase{
					GRPC: config.GRPC{
						AllowInsecureForCredentials: true,
					},
				},
			},
			component.WithClusterNew(func(_ context.Context, conf *config.ServiceBase, registerers ...rpcserver.Registerer) (cluster.Cluster, error) {
				a.So(conf, should.Resemble, &config.ServiceBase{
					GRPC: config.GRPC{
						AllowInsecureForCredentials: true,
					},
				})
				if a.So(registerers, should.HaveLength, 1) {
					a.So(registerers[0].Roles(), should.Resemble, []ttnpb.PeerInfo_Role{
						ttnpb.PeerInfo_NETWORK_SERVER,
					})
				}
				return &test.MockCluster{
					GetPeerFunc: test.MakeGetPeerChFunc(getPeerCh),
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
		},
		WithCollectionDoneFunc(MakeWindowEndChFunc(collectionDoneCh)),
		WithDeduplicationDoneFunc(MakeWindowEndChFunc(deduplicationDoneCh)),
	)).(*NetworkServer)
	ns.FrequencyPlans = frequencyplans.NewStore(test.FrequencyPlansFetcher)
	test.Must(nil, ns.Start())
	defer ns.Close()

	conn := ns.LoopbackConn()

	start := time.Now()
	ctx := test.Context()

	link, ok := assertLinkApplication(t, ctx, conn, getPeerCh, Timeout, "test-app-id")
	if !a.So(ok, should.BeTrue) || !a.So(link, should.NotBeNil) {
		t.Fatal("Failed to link application")
	}

	dev, ok := assertSetDevice(t, ctx, conn, getPeerCh, Timeout, "test-app-id", &ttnpb.SetEndDeviceRequest{
		EndDevice: ttnpb.EndDevice{
			EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
				DeviceID:               "test-dev-id",
				ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
				JoinEUI:                &types.EUI64{0x42, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
				DevEUI:                 &types.EUI64{0x42, 0x42, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			},
			FrequencyPlanID:   test.EUFrequencyPlanID,
			LoRaWANPHYVersion: ttnpb.PHY_V1_0,
			LoRaWANVersion:    ttnpb.MAC_V1_0,
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
		t.Fatal("Failed to create device")
	}
	a.So(dev.CreatedAt, should.HappenAfter, start)
	a.So(dev.UpdatedAt, should.Equal, dev.CreatedAt)
	a.So([]time.Time{start, dev.CreatedAt, time.Now()}, should.BeChronological)
	a.So(dev, should.Resemble, &ttnpb.EndDevice{
		EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
			DeviceID:               "test-dev-id",
			ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
			JoinEUI:                &types.EUI64{0x42, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			DevEUI:                 &types.EUI64{0x42, 0x42, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		FrequencyPlanID:   test.EUFrequencyPlanID,
		LoRaWANPHYVersion: ttnpb.PHY_V1_0,
		LoRaWANVersion:    ttnpb.MAC_V1_0,
		SupportsJoin:      true,
		CreatedAt:         dev.CreatedAt,
		UpdatedAt:         dev.UpdatedAt,
	})

	scheduleDownlinkCh := make(chan ScheduleDownlinkRequest)
	gsPeer := newGSPeer(test.Context(), &MockNsGsServer{
		ScheduleDownlinkFunc: MakeScheduleDownlinkChFunc(scheduleDownlinkCh),
	})

	handleJoinCh := make(chan HandleJoinRequest)
	jsPeer := newJSPeer(test.Context(), &MockNsJsServer{
		HandleJoinFunc: MakeHandleJoinChFunc(handleJoinCh),
	})
	gsns := ttnpb.NewGsNsClient(conn)

	appSKey := types.AES128Key{0x42, 0x42, 0x42, 0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	fNwkSIntKey := types.AES128Key{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

	var devAddr types.DevAddr
	t.Run("join-request", func(t *testing.T) {
		a := assertions.New(t)

		uplink := &ttnpb.UplinkMessage{
			RawPayload: []byte{
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
			},
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
				{
					GatewayIdentifiers: ttnpb.GatewayIdentifiers{
						GatewayID: "test-gtw-1",
					},
					UplinkToken: []byte("join-request-token"),
				},
			},
			ReceivedAt:          time.Now(),
			CorrelationIDs:      []string{"GsNs-1", "GsNs-2"},
			GatewayChannelIndex: 2,
		}
		handleUplinkErrCh := make(chan error)
		go func() {
			_, err := gsns.HandleUplink(ctx, uplink)
			handleUplinkErrCh <- err
			close(handleUplinkErrCh)
		}()

		a.So(test.AssertGetPeerRequest(t, getPeerCh, Timeout,
			func(ctx context.Context, role ttnpb.PeerInfo_Role, ids ttnpb.Identifiers) bool {
				return a.So(role, should.Equal, ttnpb.PeerInfo_JOIN_SERVER) &&
					a.So(ids, should.Resemble, ttnpb.EndDeviceIdentifiers{
						DeviceID:               "test-dev-id",
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
						JoinEUI:                &types.EUI64{0x42, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
						DevEUI:                 &types.EUI64{0x42, 0x42, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
					})
			},
			jsPeer,
		), should.BeTrue)

		select {
		case req := <-handleJoinCh:
			if !a.So(req.Message, should.NotBeNil) {
				t.Fatal("Nil join-request sent to JS")
			}

			a.So(req.Message.CorrelationIDs, should.Contain, "GsNs-1")
			a.So(req.Message.CorrelationIDs, should.Contain, "GsNs-2")
			a.So(req.Message.CorrelationIDs, should.HaveLength, 4)
			a.So(req.Message.DevAddr, should.NotBeEmpty)
			a.So(req.Message.DevAddr.NwkID(), should.Resemble, netID.ID())
			a.So(req.Message.DevAddr.NetIDType(), should.Equal, netID.Type())
			a.So(req.Message, should.Resemble, &ttnpb.JoinRequest{
				RawPayload:         uplink.RawPayload,
				DevAddr:            req.Message.DevAddr,
				SelectedMACVersion: ttnpb.MAC_V1_0,
				NetID:              netID,
				RxDelay:            ttnpb.RX_DELAY_6,
				CFList: &ttnpb.CFList{
					Type: ttnpb.CFListType_FREQUENCIES,
					Freq: []uint32{8671000, 8673000, 8675000, 8677000, 8679000},
				},
				CorrelationIDs: req.Message.CorrelationIDs,
			})
			req.Response <- HandleJoinResponse{
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
			}
			devAddr = req.Message.DevAddr

		case <-time.After(Timeout):
			t.Fatal("Timed out while waiting for join-request to be sent to JS")
		}

		select {
		case req := <-deduplicationDoneCh:
			req.Response <- time.Now()
			close(req.Response)

		case <-time.After(Timeout):
			t.Fatal("Timed out while waiting for deduplication window to close")
		}

		var asUp *ttnpb.ApplicationUp
		var err error
		if !a.So(test.WaitTimeout(Timeout, func() {
			asUp, err = link.Recv()
		}), should.BeTrue) {
			t.Fatal("Timed out while waiting for join-accept to be sent to AS")
		}
		if !a.So(err, should.BeNil) {
			t.Fatalf("Failed to receive AS uplink: %s", err)
		}
		a.So(asUp.CorrelationIDs, should.Contain, "GsNs-1")
		a.So(asUp.CorrelationIDs, should.Contain, "GsNs-2")
		a.So(asUp.CorrelationIDs, should.Contain, "NsJs-1")
		a.So(asUp.CorrelationIDs, should.Contain, "NsJs-2")
		a.So(asUp.CorrelationIDs, should.HaveLength, 6)
		if !a.So(asUp.ReceivedAt, should.NotBeNil) {
			a.So([]time.Time{start, *asUp.ReceivedAt, time.Now()}, should.BeChronological)
		}
		a.So(asUp, should.Resemble, &ttnpb.ApplicationUp{
			EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
				DeviceID:               "test-dev-id",
				ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
				JoinEUI:                &types.EUI64{0x42, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
				DevEUI:                 &types.EUI64{0x42, 0x42, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
				DevAddr:                &devAddr,
			},
			CorrelationIDs: asUp.CorrelationIDs,
			ReceivedAt:     asUp.ReceivedAt,
			Up: &ttnpb.ApplicationUp_JoinAccept{JoinAccept: &ttnpb.ApplicationJoinAccept{
				AppSKey: &ttnpb.KeyEnvelope{
					Key: &appSKey,
				},
				SessionKeyID: []byte("session-key-id"),
			}},
		})

		if !a.So(test.WaitTimeout(Timeout, func() {
			err = link.Send(ttnpb.Empty)
		}), should.BeTrue) {
			t.Fatal("Timed out while waiting for NS to process AS response")
		}
		if !a.So(err, should.BeNil) {
			t.Fatalf("Failed to send AS uplink response: %s", err)
		}

		select {
		case req := <-collectionDoneCh:
			req.Response <- time.Now()
			close(req.Response)

		case <-time.After(Timeout):
			t.Fatal("Timed out while waiting for collection window to close")
		}

		select {
		case err := <-handleUplinkErrCh:
			if !a.So(err, should.BeNil) {
				t.Fatalf("Failed to handle uplink: %s", err)
			}

		case <-time.After(Timeout):
			t.Fatal("Timed out while waiting for HandleUplink to return")
		}

		a.So(test.AssertGetPeerRequest(t, getPeerCh, Timeout,
			func(ctx context.Context, role ttnpb.PeerInfo_Role, ids ttnpb.Identifiers) bool {
				return a.So(role, should.Equal, ttnpb.PeerInfo_GATEWAY_SERVER) &&
					a.So(ids, should.Resemble, ttnpb.GatewayIdentifiers{
						GatewayID: "test-gtw-1",
					})
			},
			gsPeer,
		), should.BeTrue)

		a.So(AssertScheduleDownlinkRequest(t, scheduleDownlinkCh, Timeout,
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
											UplinkToken: []byte("join-request-token"),
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
			ScheduleDownlinkResponse{
				Response: &ttnpb.ScheduleDownlinkResponse{},
			},
		), should.BeTrue)
	})

	t.Logf("Device successfully joined. DevAddr: %s", devAddr)

	t.Run("uplink", func(t *testing.T) {
		a := assertions.New(t)

		uplinkFRMPayload := test.Must(crypto.EncryptUplink(fNwkSIntKey, devAddr, 0, []byte("test"))).([]byte)
		uplink := &ttnpb.UplinkMessage{
			RawPayload: func() []byte {
				b := append([]byte{
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
				)
				mic := test.Must(crypto.ComputeLegacyUplinkMIC(fNwkSIntKey, devAddr, 0, b)).([4]byte)
				return append(b, mic[:]...)
			}(),
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
				{
					GatewayIdentifiers: ttnpb.GatewayIdentifiers{
						GatewayID: "test-gtw-2",
					},
					UplinkToken: []byte("test-uplink-token"),
				},
			},
			ReceivedAt:          time.Now(),
			CorrelationIDs:      []string{"GsNs-1", "GsNs-2"},
			GatewayChannelIndex: 2,
		}

		handleUplinkErrCh := make(chan error)
		go func() {
			_, err := gsns.HandleUplink(ctx, uplink)
			handleUplinkErrCh <- err
			close(handleUplinkErrCh)
		}()

		select {
		case req := <-deduplicationDoneCh:
			req.Response <- time.Now()
			close(req.Response)

		case <-time.After(Timeout):
			t.Fatal("Timed out while waiting for deduplication window to close")
		}

		var asUp *ttnpb.ApplicationUp
		var err error
		if !a.So(test.WaitTimeout(Timeout, func() {
			asUp, err = link.Recv()
		}), should.BeTrue) {
			t.Fatal("Timed out while waiting for uplink to be sent to AS")
		}
		if !a.So(err, should.BeNil) {
			t.Fatalf("Failed to receive AS uplink: %s", err)
		}
		a.So(asUp.CorrelationIDs, should.Contain, "GsNs-1")
		a.So(asUp.CorrelationIDs, should.Contain, "GsNs-2")
		a.So(asUp.CorrelationIDs, should.HaveLength, 4)
		if !a.So(asUp.ReceivedAt, should.NotBeNil) {
			a.So([]time.Time{start, *asUp.ReceivedAt, time.Now()}, should.BeChronological)
		}
		a.So(asUp, should.Resemble, &ttnpb.ApplicationUp{
			EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
				DeviceID:               "test-dev-id",
				ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
				JoinEUI:                &types.EUI64{0x42, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
				DevEUI:                 &types.EUI64{0x42, 0x42, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
				DevAddr:                &devAddr,
			},
			CorrelationIDs: asUp.CorrelationIDs,
			ReceivedAt:     asUp.ReceivedAt,
			Up: &ttnpb.ApplicationUp_UplinkMessage{UplinkMessage: &ttnpb.ApplicationUplink{
				SessionKeyID: []byte("session-key-id"),
				FPort:        0x42,
				FCnt:         0,
				FRMPayload:   uplinkFRMPayload,
				RxMetadata:   uplink.RxMetadata,
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

		if !a.So(test.WaitTimeout(Timeout, func() {
			err = link.Send(ttnpb.Empty)
		}), should.BeTrue) {
			t.Fatal("Timed out while waiting for NS to process AS response")
		}
		if !a.So(err, should.BeNil) {
			t.Fatalf("Failed to send AS uplink response: %s", err)
		}

		select {
		case req := <-collectionDoneCh:
			req.Response <- time.Now()
			close(req.Response)

		case <-time.After(Timeout):
			t.Fatal("Timed out while waiting for collection window to close")
		}

		select {
		case err := <-handleUplinkErrCh:
			if !a.So(err, should.BeNil) {
				t.Fatalf("Failed to handle uplink: %s", err)
			}

		case <-time.After(Timeout):
			t.Fatal("Timed out while waiting for HandleUplink to return")
		}

		a.So(test.AssertGetPeerRequest(t, getPeerCh, Timeout,
			func(ctx context.Context, role ttnpb.PeerInfo_Role, ids ttnpb.Identifiers) bool {
				return a.So(role, should.Equal, ttnpb.PeerInfo_GATEWAY_SERVER) &&
					a.So(ids, should.Resemble, ttnpb.GatewayIdentifiers{
						GatewayID: "test-gtw-2",
					})
			},
			gsPeer,
		), should.BeTrue)

		a.So(AssertScheduleDownlinkRequest(t, scheduleDownlinkCh, Timeout,
			func(ctx context.Context, msg *ttnpb.DownlinkMessage) bool {
				return a.So(msg.CorrelationIDs, should.Contain, "GsNs-1") &&
					a.So(msg.CorrelationIDs, should.Contain, "GsNs-2") &&
					a.So(msg.CorrelationIDs, should.HaveLength, 5) &&
					a.So(msg, should.Resemble, &ttnpb.DownlinkMessage{
						RawPayload: func() []byte {
							b := append([]byte{
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
							)
							mic := test.Must(crypto.ComputeLegacyDownlinkMIC(fNwkSIntKey, devAddr, 1, b)).([4]byte)
							return append(b, mic[:]...)
						}(),
						Settings: &ttnpb.DownlinkMessage_Request{
							Request: &ttnpb.TxRequest{
								Class: ttnpb.CLASS_A,
								DownlinkPaths: []*ttnpb.DownlinkPath{
									{
										Path: &ttnpb.DownlinkPath_UplinkToken{
											UplinkToken: []byte("test-uplink-token"),
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
			ScheduleDownlinkResponse{
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
						t.Fatalf("Failed to add mock device to task queue: %s", err)
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

			for flow, handleFlowTest := range map[string]func(*testing.T, DeviceRegistry, DownlinkTaskQueue){
				"Class A/OTAA/EU868": handleOTAAClassA868FlowTest,
			} {
				t.Run(flow, func(t *testing.T) {
					t.Parallel()

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
					handleFlowTest(t, reg, tq)
				})
			}
		})
	}
}
