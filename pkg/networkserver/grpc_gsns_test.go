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

package networkserver_test

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"testing"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	clusterauth "go.thethings.network/lorawan-stack/pkg/auth/cluster"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/crypto"
	"go.thethings.network/lorawan-stack/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/log"
	. "go.thethings.network/lorawan-stack/pkg/networkserver"
	"go.thethings.network/lorawan-stack/pkg/networkserver/redis"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var (
	DuplicateCount = 6
	DeviceCount    = 100

	Keys = []string{"AEAEAEAEAEAEAEAEAEAEAEAEAEAEAEAE"}

	Downlinks = [...]*ttnpb.ApplicationDownlink{
		{FCnt: 0},
		{FCnt: 1},
		{FCnt: 2},
		{FCnt: 3},
		{FCnt: 4},
		{FCnt: 5},
		{FCnt: 6},
		{FCnt: 7},
		{FCnt: 8},
		{FCnt: 9},
	}
)

func init() {
	SetAppQueueUpdateTimeout(0)
}

func contextWithKey() context.Context {
	ctx := test.Context()
	md := metadata.Pairs("authorization", fmt.Sprintf("Basic %s", Keys[0]))
	if ctxMd, ok := metadata.FromIncomingContext(ctx); ok {
		md = metadata.Join(ctxMd, md)
	}
	return metadata.NewIncomingContext(ctx, md)
}

type UplinkHandler interface {
	HandleUplink(context.Context, *ttnpb.UplinkMessage) (*pbtypes.Empty, error)
}

func sendUplinkDuplicates(t *testing.T, h UplinkHandler, windowEndCh chan windowEnd, ctx context.Context, msg *ttnpb.UplinkMessage, n int) []*ttnpb.RxMetadata {
	t.Helper()

	a := assertions.New(t)

	var weCh chan<- time.Time
	select {
	case we := <-windowEndCh:
		msg := CopyUplinkMessage(msg)

		a.So(we.msg.ReceivedAt, should.HappenBefore, time.Now())
		msg.ReceivedAt = we.msg.ReceivedAt

		a.So(we.msg.CorrelationIDs, should.NotBeEmpty)
		msg.CorrelationIDs = we.msg.CorrelationIDs

		a.So(we.msg, should.Resemble, msg)
		a.So(we.ctx, should.HaveParentContext, ctx)

		weCh = we.ch

	case <-time.After(Timeout):
		t.Fatal("Timed out while waiting for window end request to arrive")
		return nil
	}

	mdCh := make(chan *ttnpb.RxMetadata, n)
	if !t.Run("duplicates", func(t *testing.T) {
		a := assertions.New(t)

		wg := &sync.WaitGroup{}
		wg.Add(n)

		for i := 0; i < n; i++ {
			go func() {
				defer wg.Done()

				msg := CopyUplinkMessage(msg)

				msg.RxMetadata = nil
				n := rand.Intn(10)
				for i := 0; i < n; i++ {
					md := ttnpb.NewPopulatedRxMetadata(test.Randy, false)
					msg.RxMetadata = append(msg.RxMetadata, md)
					mdCh <- md
				}

				_, err := h.HandleUplink(ctx, msg)
				a.So(err, should.BeNil)
			}()
		}

		go func() {
			if !a.So(test.WaitTimeout(Timeout, wg.Wait), should.BeTrue) {
				t.FailNow()
			}

			select {
			case weCh <- time.Now():

			case <-time.After(Timeout):
				t.Fatal("Timed out while waiting for metadata collection to stop")
			}

			close(mdCh)
		}()
	}) {
		t.Fatal("Failed to send duplicates")
		return nil
	}

	mds := append([]*ttnpb.RxMetadata{}, msg.RxMetadata...)
	for md := range mdCh {
		mds = append(mds, md)
	}
	return mds
}

type windowEnd struct {
	ctx context.Context
	msg *ttnpb.UplinkMessage
	ch  chan<- time.Time
}

func handleUplinkTest() func(t *testing.T) {
	return func(t *testing.T) {
		a := assertions.New(t)

		authorizedCtx := clusterauth.NewContext(test.Context(), nil)

		redisClient, flush := test.NewRedis(t, "networkserver_test")
		defer flush()
		defer redisClient.Close()
		devReg := &redis.DeviceRegistry{Redis: redisClient}

		ns := test.Must(New(
			component.MustNew(test.GetLogger(t), &component.Config{}),
			&Config{
				Devices:             devReg,
				DeduplicationWindow: 42,
				CooldownWindow:      42,
				DownlinkTasks:       &MockDownlinkTaskQueue{},
			})).(*NetworkServer)
		ns.FrequencyPlans.Fetcher = test.FrequencyPlansFetcher
		test.Must(nil, ns.Start())
		defer ns.Close()

		pb := &ttnpb.EndDevice{
			EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
				ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: ApplicationID},
				DeviceID:               DeviceID,
				DevAddr:                &DevAddr,
				JoinEUI:                &JoinEUI,
				DevEUI:                 &DevEUI,
			},
			LoRaWANVersion: ttnpb.MAC_V1_1,
			Session: &ttnpb.Session{
				DevAddr: DevAddr,
				SessionKeys: ttnpb.SessionKeys{
					FNwkSIntKey: &ttnpb.KeyEnvelope{
						Key: FNwkSIntKey[:],
					},
					SNwkSIntKey: &ttnpb.KeyEnvelope{
						Key: SNwkSIntKey[:],
					},
				},
			},
			FrequencyPlanID: test.EUFrequencyPlanID,
		}

		ret, err := CreateDevice(authorizedCtx, devReg, pb)
		if !a.So(err, should.BeNil) {
			t.FailNow()
		}

		pb.CreatedAt = ret.CreatedAt
		pb.UpdatedAt = ret.UpdatedAt
		a.So(ret, should.Resemble, pb)

		t.Run("Empty DevAddr", func(t *testing.T) {
			a := assertions.New(t)

			msg := ttnpb.NewPopulatedUplinkMessageUplink(test.Randy, SNwkSIntKey, FNwkSIntKey, false)
			msg.Payload.GetMACPayload().DevAddr = types.DevAddr{}
			_, err := ns.HandleUplink(authorizedCtx, msg)
			a.So(err, should.NotBeNil)
		})

		t.Run("FCnt too high", func(t *testing.T) {
			a := assertions.New(t)

			msg := ttnpb.NewPopulatedUplinkMessageUplink(test.Randy, SNwkSIntKey, FNwkSIntKey, false)
			msg.Payload.GetMACPayload().DevAddr = DevAddr
			msg.Payload.GetMACPayload().FCnt = math.MaxUint16 + 1
			_, err := ns.HandleUplink(authorizedCtx, msg)
			a.So(err, should.NotBeNil)
		})

		t.Run("ChannelIndex too high", func(t *testing.T) {
			a := assertions.New(t)

			msg := ttnpb.NewPopulatedUplinkMessageUplink(test.Randy, SNwkSIntKey, FNwkSIntKey, false)
			msg.Payload.GetMACPayload().DevAddr = DevAddr
			msg.Settings.ChannelIndex = math.MaxUint8 + 1
			_, err := ns.HandleUplink(authorizedCtx, msg)
			a.So(err, should.NotBeNil)
		})

		t.Run("DataRateIndex too high", func(t *testing.T) {
			a := assertions.New(t)

			msg := ttnpb.NewPopulatedUplinkMessageUplink(test.Randy, SNwkSIntKey, FNwkSIntKey, false)
			msg.Payload.GetMACPayload().DevAddr = DevAddr
			msg.Settings.DataRateIndex = math.MaxUint8 + 1
			_, err = ns.HandleUplink(authorizedCtx, msg)
			a.So(err, should.NotBeNil)
		})

		for _, tc := range []struct {
			Name string

			Device        *ttnpb.EndDevice
			NextFCntUp    uint32
			UplinkMessage *ttnpb.UplinkMessage
		}{
			{
				"1.0/unconfirmed",
				&ttnpb.EndDevice{
					MACSettings: &ttnpb.MACSettings{},
					MACState: &ttnpb.MACState{
						LoRaWANVersion: ttnpb.MAC_V1_0,
						CurrentParameters: ttnpb.MACParameters{
							Channels: []*ttnpb.MACParameters_Channel{
								ttnpb.NewPopulatedMACParameters_Channel(test.Randy, false),
								ttnpb.NewPopulatedMACParameters_Channel(test.Randy, false),
								{
									UplinkFrequency:   42,
									DownlinkFrequency: 420,
									MinDataRateIndex:  3,
									MaxDataRateIndex:  4,
								},
								ttnpb.NewPopulatedMACParameters_Channel(test.Randy, false),
							},
						},
					},
					LoRaWANVersion: ttnpb.MAC_V1_0,
					ResetsFCnt:     true,
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
							ApplicationID: ApplicationID,
						},
						DeviceID: DeviceID,
						DevAddr:  &DevAddr,
						JoinEUI:  &JoinEUI,
						DevEUI:   &DevEUI,
					},
					FrequencyPlanID: test.EUFrequencyPlanID,
					Session: &ttnpb.Session{
						DevAddr:    DevAddr,
						LastFCntUp: 0x41,
						SessionKeys: ttnpb.SessionKeys{
							FNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: FNwkSIntKey[:],
							},
							NwkSEncKey: &ttnpb.KeyEnvelope{
								Key: NwkSEncKey[:],
							},
						},
					},
				},
				0x42,
				func() *ttnpb.UplinkMessage {
					msg := ttnpb.NewPopulatedUplinkMessageUplink(test.Randy, SNwkSIntKey, FNwkSIntKey, false)
					msg.Payload.GetMACPayload().FHDR.Ack = false
					msg.Payload.GetMACPayload().FHDR.ADR = false
					msg.Settings.ChannelIndex = 2
					msg.Settings.DataRateIndex = 3

					pld := msg.Payload.GetMACPayload()
					pld.DevAddr = DevAddr
					pld.FCnt = 0x42

					msg.Payload.MIC = nil
					mic := test.Must(crypto.ComputeLegacyUplinkMIC(FNwkSIntKey, DevAddr, 0x42, test.Must(lorawan.MarshalMessage(*msg.Payload)).([]byte))).([4]byte)
					msg.Payload.MIC = mic[:]
					msg.RawPayload = test.Must(lorawan.MarshalMessage(*msg.Payload)).([]byte)

					return msg
				}(),
			},
			{
				"1.0/unconfirmed/FCnt resets",
				&ttnpb.EndDevice{
					MACSettings: &ttnpb.MACSettings{},
					MACState: &ttnpb.MACState{
						LoRaWANVersion: ttnpb.MAC_V1_0,
						CurrentParameters: ttnpb.MACParameters{
							Channels: []*ttnpb.MACParameters_Channel{
								ttnpb.NewPopulatedMACParameters_Channel(test.Randy, false),
								ttnpb.NewPopulatedMACParameters_Channel(test.Randy, false),
								{
									UplinkFrequency:   42,
									DownlinkFrequency: 420,
									MinDataRateIndex:  3,
									MaxDataRateIndex:  4,
								},
								ttnpb.NewPopulatedMACParameters_Channel(test.Randy, false),
							},
						},
					},
					LoRaWANVersion: ttnpb.MAC_V1_0,
					ResetsFCnt:     true,
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
							ApplicationID: ApplicationID,
						},
						DeviceID: DeviceID,
						DevAddr:  &DevAddr,
						JoinEUI:  &JoinEUI,
						DevEUI:   &DevEUI,
					},
					FrequencyPlanID: test.EUFrequencyPlanID,
					Session: &ttnpb.Session{
						DevAddr:    DevAddr,
						LastFCntUp: 0x42424249,
						SessionKeys: ttnpb.SessionKeys{
							FNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: FNwkSIntKey[:],
							},
							NwkSEncKey: &ttnpb.KeyEnvelope{
								Key: NwkSEncKey[:],
							},
						},
					},
				},
				0x42,
				func() *ttnpb.UplinkMessage {
					msg := ttnpb.NewPopulatedUplinkMessageUplink(test.Randy, SNwkSIntKey, FNwkSIntKey, false)
					msg.Payload.GetMACPayload().FHDR.Ack = false
					msg.Payload.GetMACPayload().FHDR.ADR = false
					msg.Settings.ChannelIndex = 2
					msg.Settings.DataRateIndex = 3

					pld := msg.Payload.GetMACPayload()
					pld.DevAddr = DevAddr
					pld.FCnt = 0x42

					msg.Payload.MIC = nil
					mic := test.Must(crypto.ComputeLegacyUplinkMIC(FNwkSIntKey, DevAddr, 0x42, test.Must(lorawan.MarshalMessage(*msg.Payload)).([]byte))).([4]byte)
					msg.Payload.MIC = mic[:]
					msg.RawPayload = test.Must(lorawan.MarshalMessage(*msg.Payload)).([]byte)

					return msg
				}(),
			},
			{
				"1.0/confirmed/ack",
				&ttnpb.EndDevice{
					MACSettings: &ttnpb.MACSettings{},
					MACState: &ttnpb.MACState{
						LoRaWANVersion:             ttnpb.MAC_V1_0,
						PendingApplicationDownlink: ttnpb.NewPopulatedApplicationDownlink(test.Randy, false),
						CurrentParameters: ttnpb.MACParameters{
							Channels: []*ttnpb.MACParameters_Channel{
								ttnpb.NewPopulatedMACParameters_Channel(test.Randy, false),
								ttnpb.NewPopulatedMACParameters_Channel(test.Randy, false),
								{
									UplinkFrequency:   42,
									DownlinkFrequency: 420,
									MinDataRateIndex:  3,
									MaxDataRateIndex:  4,
								},
								ttnpb.NewPopulatedMACParameters_Channel(test.Randy, false),
							},
						},
					},
					LoRaWANVersion: ttnpb.MAC_V1_0,
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
							ApplicationID: ApplicationID,
						},
						DeviceID: DeviceID,
						DevAddr:  &DevAddr,
						JoinEUI:  &JoinEUI,
						DevEUI:   &DevEUI,
					},
					FrequencyPlanID: test.EUFrequencyPlanID,
					Session: &ttnpb.Session{
						DevAddr:    DevAddr,
						LastFCntUp: 0x41,
						SessionKeys: ttnpb.SessionKeys{
							FNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: FNwkSIntKey[:],
							},
							NwkSEncKey: &ttnpb.KeyEnvelope{
								Key: NwkSEncKey[:],
							},
						},
					},
					RecentDownlinks: []*ttnpb.DownlinkMessage{
						ttnpb.NewPopulatedDownlinkMessage(test.Randy, false),
						ttnpb.NewPopulatedDownlinkMessage(test.Randy, false),
						ttnpb.NewPopulatedDownlinkMessage(test.Randy, false),
					},
				},
				0x42,
				func() *ttnpb.UplinkMessage {
					msg := ttnpb.NewPopulatedUplinkMessageUplink(test.Randy, SNwkSIntKey, FNwkSIntKey, true)
					msg.Payload.GetMACPayload().FHDR.Ack = true
					msg.Payload.GetMACPayload().FHDR.ADR = false
					msg.Settings.ChannelIndex = 2
					msg.Settings.DataRateIndex = 3

					pld := msg.Payload.GetMACPayload()
					pld.DevAddr = DevAddr
					pld.FCnt = 0x42

					msg.Payload.MIC = nil
					mic := test.Must(crypto.ComputeLegacyUplinkMIC(FNwkSIntKey, DevAddr, 0x42, test.Must(lorawan.MarshalMessage(*msg.Payload)).([]byte))).([4]byte)
					msg.Payload.MIC = mic[:]
					msg.RawPayload = test.Must(lorawan.MarshalMessage(*msg.Payload)).([]byte)

					return msg
				}(),
			},
			{
				"1.0/confirmed/ack/FCnt resets",
				&ttnpb.EndDevice{
					MACSettings: &ttnpb.MACSettings{},
					MACState: &ttnpb.MACState{
						LoRaWANVersion:             ttnpb.MAC_V1_0,
						PendingApplicationDownlink: ttnpb.NewPopulatedApplicationDownlink(test.Randy, false),
						CurrentParameters: ttnpb.MACParameters{
							Channels: []*ttnpb.MACParameters_Channel{
								ttnpb.NewPopulatedMACParameters_Channel(test.Randy, false),
								ttnpb.NewPopulatedMACParameters_Channel(test.Randy, false),
								{
									UplinkFrequency:   42,
									DownlinkFrequency: 420,
									MinDataRateIndex:  3,
									MaxDataRateIndex:  4,
								},
								ttnpb.NewPopulatedMACParameters_Channel(test.Randy, false),
							},
						},
					},
					LoRaWANVersion: ttnpb.MAC_V1_0,
					ResetsFCnt:     true,
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
							ApplicationID: ApplicationID,
						},
						DeviceID: DeviceID,
						DevAddr:  &DevAddr,
						JoinEUI:  &JoinEUI,
						DevEUI:   &DevEUI,
					},
					FrequencyPlanID: test.EUFrequencyPlanID,
					Session: &ttnpb.Session{
						DevAddr:    DevAddr,
						LastFCntUp: 0x42424249,
						SessionKeys: ttnpb.SessionKeys{
							FNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: FNwkSIntKey[:],
							},
							NwkSEncKey: &ttnpb.KeyEnvelope{
								Key: NwkSEncKey[:],
							},
						},
					},
					RecentDownlinks: []*ttnpb.DownlinkMessage{
						ttnpb.NewPopulatedDownlinkMessage(test.Randy, false),
						ttnpb.NewPopulatedDownlinkMessage(test.Randy, false),
						ttnpb.NewPopulatedDownlinkMessage(test.Randy, false),
					},
				},
				0x42,
				func() *ttnpb.UplinkMessage {
					msg := ttnpb.NewPopulatedUplinkMessageUplink(test.Randy, SNwkSIntKey, FNwkSIntKey, true)
					msg.Payload.GetMACPayload().FHDR.Ack = true
					msg.Payload.GetMACPayload().FHDR.ADR = false
					msg.Settings.ChannelIndex = 2
					msg.Settings.DataRateIndex = 3

					pld := msg.Payload.GetMACPayload()
					pld.DevAddr = DevAddr
					pld.FCnt = 0x42

					msg.Payload.MIC = nil
					mic := test.Must(crypto.ComputeLegacyUplinkMIC(FNwkSIntKey, DevAddr, 0x42, test.Must(lorawan.MarshalMessage(*msg.Payload)).([]byte))).([4]byte)
					msg.Payload.MIC = mic[:]
					msg.RawPayload = test.Must(lorawan.MarshalMessage(*msg.Payload)).([]byte)

					return msg
				}(),
			},
			{
				"1.1/unconfirmed",
				&ttnpb.EndDevice{
					MACSettings: &ttnpb.MACSettings{},
					MACState: &ttnpb.MACState{
						LoRaWANVersion: ttnpb.MAC_V1_1,
						CurrentParameters: ttnpb.MACParameters{
							Channels: []*ttnpb.MACParameters_Channel{
								ttnpb.NewPopulatedMACParameters_Channel(test.Randy, false),
								ttnpb.NewPopulatedMACParameters_Channel(test.Randy, false),
								{
									UplinkFrequency:   42,
									DownlinkFrequency: 420,
									MinDataRateIndex:  3,
									MaxDataRateIndex:  4,
								},
								ttnpb.NewPopulatedMACParameters_Channel(test.Randy, false),
							},
						},
					},
					LoRaWANVersion: ttnpb.MAC_V1_1,
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
							ApplicationID: ApplicationID,
						},
						DeviceID: DeviceID,
						DevAddr:  &DevAddr,
						JoinEUI:  &JoinEUI,
						DevEUI:   &DevEUI,
					},
					FrequencyPlanID: test.EUFrequencyPlanID,
					Session: &ttnpb.Session{
						DevAddr:    DevAddr,
						LastFCntUp: 0x41,
						SessionKeys: ttnpb.SessionKeys{
							FNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: FNwkSIntKey[:],
							},
							SNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: SNwkSIntKey[:],
							},
							NwkSEncKey: &ttnpb.KeyEnvelope{
								Key: NwkSEncKey[:],
							},
						},
					},
				},
				0x42,
				func() *ttnpb.UplinkMessage {
					msg := ttnpb.NewPopulatedUplinkMessageUplink(test.Randy, SNwkSIntKey, FNwkSIntKey, false)
					msg.Payload.GetMACPayload().FHDR.Ack = false
					msg.Payload.GetMACPayload().FHDR.ADR = false
					msg.Settings.ChannelIndex = 2
					msg.Settings.DataRateIndex = 3

					pld := msg.Payload.GetMACPayload()
					pld.DevAddr = DevAddr
					pld.FCnt = 0x42

					msg.Payload.MIC = nil
					mic := test.Must(crypto.ComputeUplinkMIC(SNwkSIntKey, FNwkSIntKey, 0,
						uint8(msg.Settings.DataRateIndex), uint8(msg.Settings.ChannelIndex),
						DevAddr, 0x42, test.Must(lorawan.MarshalMessage(*msg.Payload)).([]byte))).([4]byte)
					msg.Payload.MIC = mic[:]
					msg.RawPayload = test.Must(lorawan.MarshalMessage(*msg.Payload)).([]byte)

					return msg
				}(),
			},
			{
				"1.1/confirmed/ack",
				&ttnpb.EndDevice{
					MACSettings: &ttnpb.MACSettings{},
					MACState: &ttnpb.MACState{
						LoRaWANVersion:             ttnpb.MAC_V1_1,
						PendingApplicationDownlink: ttnpb.NewPopulatedApplicationDownlink(test.Randy, false),
						CurrentParameters: ttnpb.MACParameters{
							Channels: []*ttnpb.MACParameters_Channel{
								ttnpb.NewPopulatedMACParameters_Channel(test.Randy, false),
								ttnpb.NewPopulatedMACParameters_Channel(test.Randy, false),
								{
									UplinkFrequency:   42,
									DownlinkFrequency: 420,
									MinDataRateIndex:  3,
									MaxDataRateIndex:  4,
								},
								ttnpb.NewPopulatedMACParameters_Channel(test.Randy, false),
							},
						},
					},
					LoRaWANVersion: ttnpb.MAC_V1_1,
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
							ApplicationID: ApplicationID,
						},
						DeviceID: DeviceID,
						DevAddr:  &DevAddr,
						JoinEUI:  &JoinEUI,
						DevEUI:   &DevEUI,
					},
					FrequencyPlanID: test.EUFrequencyPlanID,
					Session: &ttnpb.Session{
						DevAddr:          DevAddr,
						LastFCntUp:       0x41,
						LastConfFCntDown: 0x24,
						SessionKeys: ttnpb.SessionKeys{
							FNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: FNwkSIntKey[:],
							},
							SNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: SNwkSIntKey[:],
							},
							NwkSEncKey: &ttnpb.KeyEnvelope{
								Key: NwkSEncKey[:],
							},
						},
					},
					RecentDownlinks: []*ttnpb.DownlinkMessage{
						ttnpb.NewPopulatedDownlinkMessage(test.Randy, false),
						ttnpb.NewPopulatedDownlinkMessage(test.Randy, false),
						func() *ttnpb.DownlinkMessage {
							msg := ttnpb.NewPopulatedDownlinkMessage(test.Randy, false)
							msg.Payload.MHDR.MType = ttnpb.MType_CONFIRMED_DOWN
							msg.Payload.GetMACPayload().FCnt = 0x24
							return msg
						}(),
					},
				},
				0x42,
				func() *ttnpb.UplinkMessage {
					msg := ttnpb.NewPopulatedUplinkMessageUplink(test.Randy, SNwkSIntKey, FNwkSIntKey, true)
					msg.Payload.GetMACPayload().FHDR.Ack = true
					msg.Payload.GetMACPayload().FHDR.ADR = false
					msg.Settings.ChannelIndex = 2
					msg.Settings.DataRateIndex = 3

					pld := msg.Payload.GetMACPayload()
					pld.DevAddr = DevAddr
					pld.FCnt = 0x42

					msg.Payload.MIC = nil
					mic := test.Must(crypto.ComputeUplinkMIC(SNwkSIntKey, FNwkSIntKey, 0x24,
						uint8(msg.Settings.DataRateIndex), uint8(msg.Settings.ChannelIndex),
						DevAddr, 0x42, test.Must(lorawan.MarshalMessage(*msg.Payload)).([]byte))).([4]byte)
					msg.Payload.MIC = mic[:]
					msg.RawPayload = test.Must(lorawan.MarshalMessage(*msg.Payload)).([]byte)

					return msg
				}(),
			},
			{
				"1.1/unconfirmed/FCnt resets",
				&ttnpb.EndDevice{
					MACSettings: &ttnpb.MACSettings{},
					MACState: &ttnpb.MACState{
						LoRaWANVersion: ttnpb.MAC_V1_1,
						CurrentParameters: ttnpb.MACParameters{
							Channels: []*ttnpb.MACParameters_Channel{
								ttnpb.NewPopulatedMACParameters_Channel(test.Randy, false),
								ttnpb.NewPopulatedMACParameters_Channel(test.Randy, false),
								{
									UplinkFrequency:   42,
									DownlinkFrequency: 420,
									MinDataRateIndex:  3,
									MaxDataRateIndex:  4,
								},
								ttnpb.NewPopulatedMACParameters_Channel(test.Randy, false),
							},
						},
					},
					LoRaWANVersion: ttnpb.MAC_V1_1,
					ResetsFCnt:     true,
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
							ApplicationID: ApplicationID,
						},
						DeviceID: DeviceID,
						DevAddr:  &DevAddr,
						JoinEUI:  &JoinEUI,
						DevEUI:   &DevEUI,
					},
					FrequencyPlanID: test.EUFrequencyPlanID,
					Session: &ttnpb.Session{
						DevAddr:    DevAddr,
						LastFCntUp: 0x42424249,
						SessionKeys: ttnpb.SessionKeys{
							FNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: FNwkSIntKey[:],
							},
							SNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: SNwkSIntKey[:],
							},
							NwkSEncKey: &ttnpb.KeyEnvelope{
								Key: NwkSEncKey[:],
							},
						},
					},
				},
				0x42,
				func() *ttnpb.UplinkMessage {
					msg := ttnpb.NewPopulatedUplinkMessageUplink(test.Randy, SNwkSIntKey, FNwkSIntKey, false)
					msg.Payload.GetMACPayload().FHDR.Ack = false
					msg.Payload.GetMACPayload().FHDR.ADR = false
					msg.Settings.ChannelIndex = 2
					msg.Settings.DataRateIndex = 3

					pld := msg.Payload.GetMACPayload()
					pld.DevAddr = DevAddr
					pld.FCnt = 0x42

					msg.Payload.MIC = nil
					mic := test.Must(crypto.ComputeUplinkMIC(SNwkSIntKey, FNwkSIntKey, 0,
						uint8(msg.Settings.DataRateIndex), uint8(msg.Settings.ChannelIndex),
						DevAddr, 0x42, test.Must(lorawan.MarshalMessage(*msg.Payload)).([]byte))).([4]byte)
					msg.Payload.MIC = mic[:]
					msg.RawPayload = test.Must(lorawan.MarshalMessage(*msg.Payload)).([]byte)

					return msg
				}(),
			},
			{
				"1.1/confirmed/ack/FCnt resets",
				&ttnpb.EndDevice{
					MACSettings: &ttnpb.MACSettings{},
					MACState: &ttnpb.MACState{
						LoRaWANVersion:             ttnpb.MAC_V1_1,
						PendingApplicationDownlink: ttnpb.NewPopulatedApplicationDownlink(test.Randy, false),
						CurrentParameters: ttnpb.MACParameters{
							Channels: []*ttnpb.MACParameters_Channel{
								ttnpb.NewPopulatedMACParameters_Channel(test.Randy, false),
								ttnpb.NewPopulatedMACParameters_Channel(test.Randy, false),
								{
									UplinkFrequency:   42,
									DownlinkFrequency: 420,
									MinDataRateIndex:  3,
									MaxDataRateIndex:  4,
								},
								ttnpb.NewPopulatedMACParameters_Channel(test.Randy, false),
							},
						},
					},
					LoRaWANVersion: ttnpb.MAC_V1_1,
					ResetsFCnt:     true,
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
							ApplicationID: ApplicationID,
						},
						DeviceID: DeviceID,
						DevAddr:  &DevAddr,
						JoinEUI:  &JoinEUI,
						DevEUI:   &DevEUI,
					},
					FrequencyPlanID: test.EUFrequencyPlanID,
					Session: &ttnpb.Session{
						DevAddr:          DevAddr,
						LastFCntUp:       0x42424249,
						LastConfFCntDown: 0x24,
						SessionKeys: ttnpb.SessionKeys{
							FNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: FNwkSIntKey[:],
							},
							SNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: SNwkSIntKey[:],
							},
							NwkSEncKey: &ttnpb.KeyEnvelope{
								Key: NwkSEncKey[:],
							},
						},
					},
					RecentDownlinks: []*ttnpb.DownlinkMessage{
						ttnpb.NewPopulatedDownlinkMessage(test.Randy, false),
						ttnpb.NewPopulatedDownlinkMessage(test.Randy, false),
						func() *ttnpb.DownlinkMessage {
							msg := ttnpb.NewPopulatedDownlinkMessage(test.Randy, false)
							msg.Payload.GetMACPayload().FCnt = 0x24
							return msg
						}(),
					},
				},
				0x42,
				func() *ttnpb.UplinkMessage {
					msg := ttnpb.NewPopulatedUplinkMessageUplink(test.Randy, SNwkSIntKey, FNwkSIntKey, true)
					msg.Payload.GetMACPayload().FHDR.Ack = true
					msg.Payload.GetMACPayload().FHDR.ADR = false
					msg.Settings.ChannelIndex = 2
					msg.Settings.DataRateIndex = 3

					pld := msg.Payload.GetMACPayload()
					pld.DevAddr = DevAddr
					pld.FCnt = 0x42

					msg.Payload.MIC = nil
					mic := test.Must(crypto.ComputeUplinkMIC(SNwkSIntKey, FNwkSIntKey, 0x24,
						uint8(msg.Settings.DataRateIndex), uint8(msg.Settings.ChannelIndex),
						DevAddr, 0x42, test.Must(lorawan.MarshalMessage(*msg.Payload)).([]byte))).([4]byte)
					msg.Payload.MIC = mic[:]
					msg.RawPayload = test.Must(lorawan.MarshalMessage(*msg.Payload)).([]byte)

					return msg
				}(),
			},
		} {
			t.Run(tc.Name, func(t *testing.T) {
				a := assertions.New(t)

				redisClient, flush := test.NewRedis(t, "networkserver_test")
				defer flush()
				defer redisClient.Close()
				devReg := &redis.DeviceRegistry{Redis: redisClient}

				populateSessionKeys := func(s *ttnpb.Session) {
					for s.SessionKeys.FNwkSIntKey == nil ||
						len(s.SessionKeys.FNwkSIntKey.Key) == 0 ||
						s.SessionKeys.FNwkSIntKey.KEKLabel == "" && bytes.Equal(s.SessionKeys.FNwkSIntKey.Key, FNwkSIntKey[:]) {

						s.SessionKeys.FNwkSIntKey = ttnpb.NewPopulatedKeyEnvelope(test.Randy, false)
					}

					for s.SessionKeys.SNwkSIntKey == nil ||
						len(s.SessionKeys.SNwkSIntKey.Key) == 0 ||
						s.SessionKeys.SNwkSIntKey.KEKLabel == "" && bytes.Equal(s.SessionKeys.SNwkSIntKey.Key, SNwkSIntKey[:]) {

						s.SessionKeys.SNwkSIntKey = ttnpb.NewPopulatedKeyEnvelope(test.Randy, false)
					}
				}

				ctx := context.WithValue(authorizedCtx, struct{}{}, 42)
				ctx = log.NewContext(ctx, test.GetLogger(t))

				// Fill DeviceRegistry with devices
				for i := 0; i < DeviceCount; i++ {
					pb := ttnpb.NewPopulatedEndDevice(test.Randy, false)
					for pb.Equal(tc.Device) {
						pb = ttnpb.NewPopulatedEndDevice(test.Randy, false)
					}

					if s := pb.Session; s != nil {
						populateSessionKeys(s)

						s.DevAddr = DevAddr
						for pb.PendingSession != nil && pb.PendingSession.DevAddr.Equal(s.DevAddr) {
							pb.PendingSession.DevAddr = *types.NewPopulatedDevAddr(test.Randy)
						}
					} else if s := pb.PendingSession; s != nil {
						populateSessionKeys(s)

						s.DevAddr = DevAddr
						for pb.Session != nil && pb.Session.DevAddr.Equal(s.DevAddr) {
							pb.Session.DevAddr = *types.NewPopulatedDevAddr(test.Randy)
						}
					}

					ret, err := CreateDevice(ctx, devReg, pb)
					if !a.So(err, should.BeNil) {
						t.FailNow()
					}
					pb.CreatedAt = ret.CreatedAt
					pb.UpdatedAt = ret.UpdatedAt
					a.So(ret, should.Resemble, pb)
				}

				deduplicationDoneCh := make(chan windowEnd, 1)
				collectionDoneCh := make(chan windowEnd, 1)

				type asSendReq struct {
					up    *ttnpb.ApplicationUp
					errch chan error
				}
				asSendCh := make(chan asSendReq)

				type downlinkTasksAddRequest struct {
					ctx   context.Context
					devID ttnpb.EndDeviceIdentifiers
					t     time.Time
				}
				downlinkAddCh := make(chan downlinkTasksAddRequest, 1)

				ns := test.Must(New(
					component.MustNew(test.GetLogger(t), &component.Config{}),
					&Config{
						Devices:             devReg,
						DeduplicationWindow: 42,
						CooldownWindow:      42,
						DownlinkTasks: &MockDownlinkTaskQueue{
							AddFunc: func(ctx context.Context, devID ttnpb.EndDeviceIdentifiers, t time.Time) error {
								downlinkAddCh <- downlinkTasksAddRequest{
									ctx:   ctx,
									devID: devID,
									t:     t,
								}
								return nil
							},
						},
					},
					WithDeduplicationDoneFunc(func(ctx context.Context, msg *ttnpb.UplinkMessage) <-chan time.Time {
						ch := make(chan time.Time, 1)
						deduplicationDoneCh <- windowEnd{ctx, msg, ch}
						return ch
					}),
					WithCollectionDoneFunc(func(ctx context.Context, msg *ttnpb.UplinkMessage) <-chan time.Time {
						ch := make(chan time.Time, 1)
						collectionDoneCh <- windowEnd{ctx, msg, ch}
						return ch
					}),
					WithASUplinkHandler(func(ctx context.Context, ids ttnpb.ApplicationIdentifiers, up *ttnpb.ApplicationUp) (bool, error) {
						req := asSendReq{
							up:    up,
							errch: make(chan error),
						}
						asSendCh <- req
						return true, <-req.errch
					}),
				)).(*NetworkServer)
				ns.FrequencyPlans.Fetcher = test.FrequencyPlansFetcher
				test.Must(nil, ns.Start())
				defer ns.Close()

				pb := CopyEndDevice(tc.Device)

				start := time.Now()

				ret, err := CreateDevice(ctx, devReg, CopyEndDevice(tc.Device))
				if !a.So(err, should.BeNil) {
					t.FailNow()
				}
				pb.CreatedAt = ret.CreatedAt
				a.So(ret.UpdatedAt, should.HappenAfter, start)
				pb.UpdatedAt = ret.UpdatedAt
				a.So(ret, should.HaveEmptyDiff, pb)

				errch := make(chan error, 1)
				go func() {
					_, err := ns.HandleUplink(ctx, CopyUplinkMessage(tc.UplinkMessage))
					errch <- err
				}()

				if pb.MACState != nil && pb.MACState.PendingApplicationDownlink != nil {
					select {
					case req := <-asSendCh:
						if tc.UplinkMessage.Payload.GetMACPayload().Ack {
							a.So(req.up.GetDownlinkAck(), should.HaveEmptyDiff, pb.MACState.PendingApplicationDownlink)
						} else {
							a.So(req.up.GetDownlinkNack(), should.HaveEmptyDiff, pb.MACState.PendingApplicationDownlink)
						}
						close(req.errch)

					case we := <-collectionDoneCh:
						close(we.ch)
						err := <-errch
						a.So(err, should.BeNil)
						t.Fatalf("Downlink (n)ack not sent to AS: %v", errors.Stack(err))

					case <-time.After(Timeout):
						t.Fatal("Timed out while waiting for (n)ack to be sent to AS")
					}
				}

				md := sendUplinkDuplicates(t, ns, deduplicationDoneCh, ctx, tc.UplinkMessage, DuplicateCount)

				var asUpReq asSendReq
				select {
				case asUpReq = <-asSendCh:
					a.So(md, should.HaveSameElementsDeep, asUpReq.up.GetUplinkMessage().RxMetadata)
					a.So(asUpReq.up.CorrelationIDs, should.NotBeEmpty)

					a.So(asUpReq.up, should.Resemble, &ttnpb.ApplicationUp{
						EndDeviceIdentifiers: pb.EndDeviceIdentifiers,
						CorrelationIDs:       asUpReq.up.CorrelationIDs,
						Up: &ttnpb.ApplicationUp_UplinkMessage{UplinkMessage: &ttnpb.ApplicationUplink{
							FCnt:         tc.NextFCntUp,
							FPort:        tc.UplinkMessage.Payload.GetMACPayload().FPort,
							FRMPayload:   tc.UplinkMessage.Payload.GetMACPayload().FRMPayload,
							RxMetadata:   asUpReq.up.GetUplinkMessage().RxMetadata,
							SessionKeyID: pb.Session.SessionKeys.SessionKeyID,
							Settings:     asUpReq.up.GetUplinkMessage().Settings,
						}},
					})

				case we := <-collectionDoneCh:
					close(we.ch)
					a.So(<-errch, should.BeNil)
					t.Fatal("Uplink not sent to AS")

				case <-time.After(Timeout):
					t.Fatal("Timed out while waiting for uplink to be sent to AS")
				}

				if !t.Run("device update", func(t *testing.T) {
					a := assertions.New(t)

					ret, err := devReg.GetByID(ctx, pb.EndDeviceIdentifiers.ApplicationIdentifiers, pb.EndDeviceIdentifiers.DeviceID, ttnpb.EndDeviceFieldPathsTopLevel)
					if !a.So(err, should.BeNil) ||
						!a.So(ret, should.NotBeNil) {
						t.FailNow()
					}

					if !a.So(ret.RecentUplinks, should.NotBeEmpty) {
						t.FailNow()
					}

					pb.Session.LastFCntUp = tc.NextFCntUp
					pb.PendingSession = nil
					pb.CreatedAt = ret.CreatedAt
					pb.UpdatedAt = ret.UpdatedAt
					if pb.MACState == nil {
						err := ResetMACState(ns.Component.FrequencyPlans, pb)
						if !a.So(err, should.BeNil) {
							t.FailNow()
						}
					}
					pb.MACState.RxWindowsAvailable = true
					pb.MACState.PendingApplicationDownlink = nil

					msg := CopyUplinkMessage(tc.UplinkMessage)
					msg.RxMetadata = md

					pb.RecentUplinks = append(pb.RecentUplinks, msg)
					if len(pb.RecentUplinks) > RecentUplinkCount {
						pb.RecentUplinks = pb.RecentUplinks[len(pb.RecentUplinks)-RecentUplinkCount:]
					}

					retUp := ret.RecentUplinks[len(ret.RecentUplinks)-1]
					pbUp := pb.RecentUplinks[len(pb.RecentUplinks)-1]

					a.So(retUp.ReceivedAt, should.HappenBetween, start, time.Now())
					pbUp.ReceivedAt = retUp.ReceivedAt

					a.So(retUp.CorrelationIDs, should.NotBeEmpty)
					pbUp.CorrelationIDs = retUp.CorrelationIDs

					a.So(retUp.RxMetadata, should.HaveSameElementsDiff, pbUp.RxMetadata)
					pbUp.RxMetadata = retUp.RxMetadata

					a.So(ret, should.HaveEmptyDiff, pb)
				}) {
					t.FailNow()
				}

				close(deduplicationDoneCh)
				close(asUpReq.errch)

				select {
				case req := <-downlinkAddCh:
					a.So(req.ctx, should.HaveParentContext, ctx)
					a.So(req.devID, should.Resemble, pb.EndDeviceIdentifiers)
					a.So([]time.Time{start, req.t, time.Now()}, should.BeChronological)

				case <-time.After(Timeout):
					t.Fatal("Timeout waiting for Add to be called")
				}

				_ = sendUplinkDuplicates(t, ns, collectionDoneCh, ctx, tc.UplinkMessage, DuplicateCount)
				close(collectionDoneCh)

				select {
				case err := <-errch:
					a.So(err, should.BeNil)

				case <-time.After(Timeout):
					t.Fatal("Timed out while waiting for HandleUplink to return")
				}

				t.Run("after cooldown window", func(t *testing.T) {
					a := assertions.New(t)

					deduplicationDoneCh = make(chan windowEnd, 1)
					collectionDoneCh = make(chan windowEnd, 1)

					errch := make(chan error, 1)
					go func() {
						_, err = ns.HandleUplink(ctx, CopyUplinkMessage(tc.UplinkMessage))
						errch <- err
					}()

					if !pb.ResetsFCnt {
						close(deduplicationDoneCh)
						_ = sendUplinkDuplicates(t, ns, collectionDoneCh, ctx, tc.UplinkMessage, DuplicateCount)
						close(collectionDoneCh)

						select {
						case err := <-errch:
							a.So(err, should.BeError)

						case <-time.After(Timeout):
							t.Fatal("Timed out while waiting for HandleUplink to return")
						}

						return
					}

					if pb.MACState != nil && pb.MACState.PendingApplicationDownlink != nil {
						select {
						case req := <-asSendCh:
							if tc.UplinkMessage.Payload.GetMACPayload().Ack {
								a.So(req.up.GetDownlinkAck(), should.Resemble, pb.MACState.PendingApplicationDownlink)
							} else {
								a.So(req.up.GetDownlinkNack(), should.Resemble, pb.MACState.PendingApplicationDownlink)
							}
							close(req.errch)

						case we := <-collectionDoneCh:
							close(we.ch)
							a.So(<-errch, should.BeNil)
							t.Fatal("Downlink (n)ack not sent to AS")

						case <-time.After(Timeout):
							t.Fatal("Timed out while waiting for (n)ack to be sent to AS")
						}
					}

					md := sendUplinkDuplicates(t, ns, deduplicationDoneCh, ctx, tc.UplinkMessage, DuplicateCount)

					select {
					case asUpReq = <-asSendCh:
						a.So(md, should.HaveSameElementsDeep, asUpReq.up.GetUplinkMessage().RxMetadata)
						a.So(asUpReq.up.CorrelationIDs, should.NotBeEmpty)

						a.So(asUpReq.up, should.Resemble, &ttnpb.ApplicationUp{
							EndDeviceIdentifiers: pb.EndDeviceIdentifiers,
							CorrelationIDs:       asUpReq.up.CorrelationIDs,
							Up: &ttnpb.ApplicationUp_UplinkMessage{UplinkMessage: &ttnpb.ApplicationUplink{
								FCnt:         tc.NextFCntUp,
								FPort:        tc.UplinkMessage.Payload.GetMACPayload().FPort,
								FRMPayload:   tc.UplinkMessage.Payload.GetMACPayload().FRMPayload,
								RxMetadata:   asUpReq.up.GetUplinkMessage().RxMetadata,
								SessionKeyID: pb.Session.SessionKeys.SessionKeyID,
								Settings:     asUpReq.up.GetUplinkMessage().Settings,
							}},
						})

					case we := <-collectionDoneCh:
						close(we.ch)
						a.So(<-errch, should.BeNil)
						t.Fatal("Uplink not sent to AS")

					case <-time.After(Timeout):
						t.Fatal("Timed out while waiting for uplink to be sent to AS")
					}

					close(deduplicationDoneCh)
					close(asUpReq.errch)

					select {
					case req := <-downlinkAddCh:
						a.So(req.ctx, should.HaveParentContext, ctx)
						a.So(req.devID, should.Resemble, pb.EndDeviceIdentifiers)
						a.So([]time.Time{start, req.t, time.Now()}, should.BeChronological)

					case <-time.After(Timeout):
						t.Fatal("Timeout waiting for Add to be called")
					}

					_ = sendUplinkDuplicates(t, ns, collectionDoneCh, ctx, tc.UplinkMessage, DuplicateCount)
					close(collectionDoneCh)

					select {
					case err := <-errch:
						a.So(err, should.BeNil)

					case <-time.After(Timeout):
						t.Fatal("Timed out while waiting for HandleUplink to return")
					}
				})
			})
		}
	}
}

var _ ttnpb.NsJsClient = &MockNsJsClient{}

type MockNsJsClient struct {
	*test.MockClientStream
	HandleJoinFunc  func(context.Context, *ttnpb.JoinRequest, ...grpc.CallOption) (*ttnpb.JoinResponse, error)
	GetNwkSKeysFunc func(context.Context, *ttnpb.SessionKeyRequest, ...grpc.CallOption) (*ttnpb.NwkSKeysResponse, error)
}

func (js *MockNsJsClient) HandleJoin(ctx context.Context, req *ttnpb.JoinRequest, opts ...grpc.CallOption) (*ttnpb.JoinResponse, error) {
	if js.HandleJoinFunc == nil {
		return nil, errors.New("HandleJoinFunc not set")
	}
	return js.HandleJoinFunc(ctx, req, opts...)
}

func (js *MockNsJsClient) GetNwkSKeys(ctx context.Context, req *ttnpb.SessionKeyRequest, opts ...grpc.CallOption) (*ttnpb.NwkSKeysResponse, error) {
	if js.GetNwkSKeysFunc == nil {
		return nil, errors.New("GetNwkSKeysFunc not set")
	}
	return js.GetNwkSKeysFunc(ctx, req, opts...)
}

func handleJoinTest() func(t *testing.T) {
	return func(t *testing.T) {
		a := assertions.New(t)

		authorizedCtx := clusterauth.NewContext(test.Context(), nil)

		redisClient, flush := test.NewRedis(t, "networkserver_test")
		defer flush()
		defer redisClient.Close()
		devReg := &redis.DeviceRegistry{Redis: redisClient}

		ns := test.Must(New(
			component.MustNew(test.GetLogger(t), &component.Config{}),
			&Config{
				Devices:             devReg,
				DeduplicationWindow: 42,
				CooldownWindow:      42,
				DownlinkTasks:       &MockDownlinkTaskQueue{},
			},
		)).(*NetworkServer)
		ns.FrequencyPlans.Fetcher = test.FrequencyPlansFetcher
		test.Must(nil, ns.Start())
		defer ns.Close()

		_, err := ns.HandleUplink(authorizedCtx, ttnpb.NewPopulatedUplinkMessageJoinRequest(test.Randy))
		a.So(err, should.NotBeNil)

		req := ttnpb.NewPopulatedUplinkMessageJoinRequest(test.Randy)
		pb := ttnpb.NewPopulatedEndDevice(test.Randy, false)

		pb.EndDeviceIdentifiers.ApplicationID = ApplicationID
		pb.EndDeviceIdentifiers.DeviceID = DeviceID
		pb.EndDeviceIdentifiers.DevEUI = &req.Payload.GetJoinRequestPayload().DevEUI
		pb.EndDeviceIdentifiers.JoinEUI = &req.Payload.GetJoinRequestPayload().JoinEUI

		_, err = CreateDevice(authorizedCtx, devReg, pb)
		if !a.So(err, should.BeNil) {
			t.FailNow()
		}

		_, err = ns.HandleUplink(authorizedCtx, req)
		a.So(err, should.NotBeNil)

		for _, tc := range []struct {
			Name string

			Device        *ttnpb.EndDevice
			UplinkMessage *ttnpb.UplinkMessage
		}{
			{
				"1.1/nil session",
				&ttnpb.EndDevice{
					LoRaWANVersion: ttnpb.MAC_V1_1,
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						DevEUI:                 &DevEUI,
						JoinEUI:                &JoinEUI,
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: ApplicationID},
						DeviceID:               DeviceID,
					},
					FrequencyPlanID: test.EUFrequencyPlanID,
					Session:         nil,
					MACState:        ttnpb.NewPopulatedMACState(test.Randy, false),
				},
				func() *ttnpb.UplinkMessage {
					msg := ttnpb.NewPopulatedUplinkMessageJoinRequest(test.Randy)
					msg.Settings.ChannelIndex = 2
					msg.Settings.DataRateIndex = 3

					jr := msg.Payload.GetJoinRequestPayload()
					jr.DevEUI = DevEUI
					jr.JoinEUI = JoinEUI

					return msg
				}(),
			},
			{
				"1.1/active session",
				&ttnpb.EndDevice{
					LoRaWANVersion: ttnpb.MAC_V1_1,
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						DevEUI:                 &DevEUI,
						JoinEUI:                &JoinEUI,
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: ApplicationID},
						DeviceID:               DeviceID,
					},
					FrequencyPlanID: test.EUFrequencyPlanID,
					Session:         ttnpb.NewPopulatedSession(test.Randy, false),
					MACState:        ttnpb.NewPopulatedMACState(test.Randy, false),
				},
				func() *ttnpb.UplinkMessage {
					msg := ttnpb.NewPopulatedUplinkMessageJoinRequest(test.Randy)
					msg.Settings.ChannelIndex = 2
					msg.Settings.DataRateIndex = 3

					jr := msg.Payload.GetJoinRequestPayload()
					jr.DevEUI = DevEUI
					jr.JoinEUI = JoinEUI

					return msg
				}(),
			},
			{
				"1.0.2",
				&ttnpb.EndDevice{
					LoRaWANVersion: ttnpb.MAC_V1_0_2,
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						DevEUI:                 &DevEUI,
						JoinEUI:                &JoinEUI,
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: ApplicationID},
						DeviceID:               DeviceID,
					},
					FrequencyPlanID: test.EUFrequencyPlanID,
					Session:         ttnpb.NewPopulatedSession(test.Randy, false),
					MACState:        ttnpb.NewPopulatedMACState(test.Randy, false),
				},
				func() *ttnpb.UplinkMessage {
					msg := ttnpb.NewPopulatedUplinkMessageJoinRequest(test.Randy)
					msg.Settings.ChannelIndex = 2
					msg.Settings.DataRateIndex = 3

					jr := msg.Payload.GetJoinRequestPayload()
					jr.DevEUI = DevEUI
					jr.JoinEUI = JoinEUI

					return msg
				}(),
			},
			{
				"1.0.1",
				&ttnpb.EndDevice{
					LoRaWANVersion: ttnpb.MAC_V1_0_1,
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						DevEUI:                 &DevEUI,
						JoinEUI:                &JoinEUI,
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: ApplicationID},
						DeviceID:               DeviceID,
					},
					FrequencyPlanID: test.EUFrequencyPlanID,
					Session:         ttnpb.NewPopulatedSession(test.Randy, false),
					MACState:        ttnpb.NewPopulatedMACState(test.Randy, false),
				},
				func() *ttnpb.UplinkMessage {
					msg := ttnpb.NewPopulatedUplinkMessageJoinRequest(test.Randy)
					msg.Settings.ChannelIndex = 2
					msg.Settings.DataRateIndex = 3

					jr := msg.Payload.GetJoinRequestPayload()
					jr.DevEUI = DevEUI
					jr.JoinEUI = JoinEUI

					return msg
				}(),
			},
			{
				"1.0",
				&ttnpb.EndDevice{
					LoRaWANVersion: ttnpb.MAC_V1_0,
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						DevEUI:                 &DevEUI,
						JoinEUI:                &JoinEUI,
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: ApplicationID},
						DeviceID:               DeviceID,
					},
					FrequencyPlanID: test.EUFrequencyPlanID,
					Session:         ttnpb.NewPopulatedSession(test.Randy, false),
					MACState:        ttnpb.NewPopulatedMACState(test.Randy, false),
				},
				func() *ttnpb.UplinkMessage {
					msg := ttnpb.NewPopulatedUplinkMessageJoinRequest(test.Randy)
					msg.Settings.ChannelIndex = 2
					msg.Settings.DataRateIndex = 3

					jr := msg.Payload.GetJoinRequestPayload()
					jr.DevEUI = DevEUI
					jr.JoinEUI = JoinEUI

					return msg
				}(),
			},
		} {
			t.Run(tc.Name, func(t *testing.T) {
				a := assertions.New(t)

				authorizedCtx = test.ContextWithT(authorizedCtx, t)

				redisClient, flush := test.NewRedis(t, "networkserver_test")
				defer flush()
				defer redisClient.Close()
				devReg := &redis.DeviceRegistry{Redis: redisClient}

				// Fill DeviceRegistry with devices
				for i := 0; i < DeviceCount; i++ {
					pb := ttnpb.NewPopulatedEndDevice(test.Randy, false)
					for pb.Equal(tc.Device) {
						pb = ttnpb.NewPopulatedEndDevice(test.Randy, false)
					}

					_, err = CreateDevice(authorizedCtx, devReg, pb)
					if !a.So(err, should.BeNil) {
						t.FailNow()
					}
				}

				type handleJoinRequest struct {
					ctx   context.Context
					req   *ttnpb.JoinRequest
					ch    chan<- *ttnpb.JoinResponse
					errch chan<- error
				}

				type downlinkTasksAddRequest struct {
					ctx   context.Context
					devID ttnpb.EndDeviceIdentifiers
					t     time.Time
				}
				downlinkAddCh := make(chan downlinkTasksAddRequest, 1)

				deduplicationDoneCh := make(chan windowEnd, 1)
				collectionDoneCh := make(chan windowEnd, 1)
				handleJoinCh := make(chan handleJoinRequest, 1)
				asSendCh := make(chan *ttnpb.ApplicationUp)

				keys := ttnpb.NewPopulatedSessionKeys(test.Randy, false)

				ns := test.Must(New(
					component.MustNew(test.GetLogger(t), &component.Config{}),
					&Config{
						Devices: devReg,
						DownlinkTasks: &MockDownlinkTaskQueue{
							AddFunc: func(ctx context.Context, devID ttnpb.EndDeviceIdentifiers, t time.Time) error {
								downlinkAddCh <- downlinkTasksAddRequest{
									ctx:   ctx,
									devID: devID,
									t:     t,
								}
								return nil
							},
						},
					},
					WithDeduplicationDoneFunc(func(ctx context.Context, msg *ttnpb.UplinkMessage) <-chan time.Time {
						ch := make(chan time.Time, 1)
						deduplicationDoneCh <- windowEnd{ctx, msg, ch}
						return ch
					}),
					WithCollectionDoneFunc(func(ctx context.Context, msg *ttnpb.UplinkMessage) <-chan time.Time {
						ch := make(chan time.Time, 1)
						collectionDoneCh <- windowEnd{ctx, msg, ch}
						return ch
					}),
					WithNsGsClientFunc(func(ctx context.Context, id ttnpb.GatewayIdentifiers) (ttnpb.NsGsClient, error) {
						return &MockNsGsClient{}, nil
					}),
					WithNsJsClientFunc(func(ctx context.Context, id ttnpb.EndDeviceIdentifiers) (ttnpb.NsJsClient, error) {
						return &MockNsJsClient{
							GetNwkSKeysFunc: func(ctx context.Context, req *ttnpb.SessionKeyRequest, _ ...grpc.CallOption) (*ttnpb.NwkSKeysResponse, error) {
								err := errors.New("GetNwkSKeys should not be called")
								test.MustTFromContext(ctx).Error(err)
								return nil, err
							},
							HandleJoinFunc: func(ctx context.Context, req *ttnpb.JoinRequest, _ ...grpc.CallOption) (*ttnpb.JoinResponse, error) {
								ch := make(chan *ttnpb.JoinResponse, 1)
								errch := make(chan error, 1)
								handleJoinCh <- handleJoinRequest{ctx, req, ch, errch}
								return <-ch, <-errch
							},
						}, nil
					}),
					WithASUplinkHandler(func(ctx context.Context, ids ttnpb.ApplicationIdentifiers, up *ttnpb.ApplicationUp) (bool, error) {
						asSendCh <- up
						return true, nil
					}),
				)).(*NetworkServer)
				ns.Component.FrequencyPlans.Fetcher = test.FrequencyPlansFetcher

				test.Must(nil, ns.Start())
				defer ns.Close()

				pb := CopyEndDevice(tc.Device)

				ret, err := CreateDevice(authorizedCtx, devReg, CopyEndDevice(pb))
				a.So(err, should.BeNil)
				a.So(ret.CreatedAt, should.Equal, ret.UpdatedAt)
				pb.CreatedAt = ret.CreatedAt
				pb.UpdatedAt = ret.UpdatedAt
				a.So(ret, should.Resemble, pb)

				expectedRequest := &ttnpb.JoinRequest{
					RawPayload: tc.UplinkMessage.RawPayload,
					Payload:    tc.UplinkMessage.Payload,
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						DevEUI:  &DevEUI,
						JoinEUI: &JoinEUI,
					},
					NetID:              ns.NetID,
					SelectedMACVersion: tc.Device.LoRaWANVersion,
					RxDelay:            tc.Device.MACState.DesiredParameters.Rx1Delay,
					CFList:             nil,
					DownlinkSettings: ttnpb.DLSettings{
						Rx1DROffset: tc.Device.MACState.DesiredParameters.Rx1DataRateOffset,
						Rx2DR:       tc.Device.MACState.DesiredParameters.Rx2DataRateIndex,
						OptNeg:      true,
					},
				}

				ctx := context.WithValue(authorizedCtx, struct{}{}, 42)
				ctx = log.NewContext(ctx, test.GetLogger(t))

				start := time.Now()

				resp := ttnpb.NewPopulatedJoinResponse(test.Randy, false)
				resp.SessionKeys = *keys

				errch := make(chan error, 1)
				go func() {
					_, err := ns.HandleUplink(ctx, CopyUplinkMessage(tc.UplinkMessage))
					errch <- err
				}()

				select {
				case req := <-handleJoinCh:
					if ses := tc.Device.Session; ses != nil {
						a.So(req.req.EndDeviceIdentifiers.DevAddr, should.NotResemble, ses.DevAddr)
					}

					expectedRequest.EndDeviceIdentifiers.DevAddr = req.req.EndDeviceIdentifiers.DevAddr
					a.So(req.req, should.Resemble, expectedRequest)

					req.ch <- resp
					req.errch <- nil

				case we := <-collectionDoneCh:
					close(we.ch)
					a.So(<-errch, should.BeNil)
					t.Fatal("Join request not sent to JS")

				case <-time.After(Timeout):
					t.Fatal("Timed out while waiting for join to be sent to JS")
				}

				md := sendUplinkDuplicates(t, ns, deduplicationDoneCh, ctx, tc.UplinkMessage, DuplicateCount)

				close(deduplicationDoneCh)

				select {
				case up := <-asSendCh:
					if !t.Run("device update", func(t *testing.T) {
						a := assertions.New(t)

						ret, err := devReg.GetByID(authorizedCtx, pb.EndDeviceIdentifiers.ApplicationIdentifiers, pb.EndDeviceIdentifiers.DeviceID, ttnpb.EndDeviceFieldPathsTopLevel)
						if !a.So(err, should.BeNil) ||
							!a.So(ret, should.NotBeNil) {
							t.FailNow()
						}
						if a.So(ret.Session, should.NotBeNil) {
							a.So(ret.Session.StartedAt, should.HappenBetween, start, time.Now())
							a.So(ret.EndDeviceIdentifiers.DevAddr, should.Resemble, &ret.Session.DevAddr)
							if tc.Device.Session != nil {
								a.So(ret.Session.DevAddr, should.NotResemble, tc.Device.Session.DevAddr)
							}
						}

						if !a.So(ret.RecentUplinks, should.NotBeEmpty) {
							t.FailNow()
						}

						err = ResetMACState(ns.Component.FrequencyPlans, pb)
						if !a.So(err, should.BeNil) {
							t.FailNow()
						}

						pb.MACState.RxWindowsAvailable = true
						pb.MACState.QueuedJoinAccept = resp.RawPayload
						pb.MACState.CurrentParameters.Rx1Delay = tc.Device.MACState.DesiredParameters.Rx1Delay
						pb.MACState.CurrentParameters.Rx1DataRateOffset = tc.Device.MACState.DesiredParameters.Rx1DataRateOffset
						pb.MACState.CurrentParameters.Rx2DataRateIndex = tc.Device.MACState.DesiredParameters.Rx2DataRateIndex

						pb.MACState.DesiredParameters.Rx1Delay = pb.MACState.CurrentParameters.Rx1Delay
						pb.MACState.DesiredParameters.Rx1DataRateOffset = pb.MACState.CurrentParameters.Rx1DataRateOffset
						pb.MACState.DesiredParameters.Rx2DataRateIndex = pb.MACState.CurrentParameters.Rx2DataRateIndex

						pb.EndDeviceIdentifiers.DevAddr = ret.EndDeviceIdentifiers.DevAddr
						pb.Session = &ttnpb.Session{
							DevAddr:     *ret.EndDeviceIdentifiers.DevAddr,
							SessionKeys: *keys,
							StartedAt:   ret.Session.StartedAt,
						}
						pb.CreatedAt = ret.CreatedAt
						pb.UpdatedAt = ret.UpdatedAt
						pb.QueuedApplicationDownlinks = nil

						msg := CopyUplinkMessage(tc.UplinkMessage)
						msg.RxMetadata = md

						pb.RecentUplinks = append(pb.RecentUplinks, msg)
						if len(pb.RecentUplinks) > RecentUplinkCount {
							pb.RecentUplinks = pb.RecentUplinks[len(pb.RecentUplinks)-RecentUplinkCount:]
						}

						retUp := ret.RecentUplinks[len(ret.RecentUplinks)-1]
						pbUp := pb.RecentUplinks[len(pb.RecentUplinks)-1]

						a.So(retUp.ReceivedAt, should.HappenBetween, start, time.Now())
						pbUp.ReceivedAt = retUp.ReceivedAt

						a.So(retUp.CorrelationIDs, should.NotBeEmpty)
						pbUp.CorrelationIDs = retUp.CorrelationIDs

						a.So(retUp.RxMetadata, should.HaveSameElementsDiff, pbUp.RxMetadata)
						pbUp.RxMetadata = retUp.RxMetadata

						a.So(ret, should.HaveEmptyDiff, pb)
					}) {
						t.FailNow()
					}

					a.So(up.CorrelationIDs, should.NotBeEmpty)
					a.So(up, should.HaveEmptyDiff, &ttnpb.ApplicationUp{
						CorrelationIDs:       up.CorrelationIDs,
						EndDeviceIdentifiers: pb.EndDeviceIdentifiers,
						Up: &ttnpb.ApplicationUp_JoinAccept{JoinAccept: &ttnpb.ApplicationJoinAccept{
							AppSKey:              resp.SessionKeys.AppSKey,
							SessionKeyID:         test.Must(devReg.GetByID(ctx, tc.Device.EndDeviceIdentifiers.ApplicationIdentifiers, tc.Device.EndDeviceIdentifiers.DeviceID, ttnpb.EndDeviceFieldPathsTopLevel)).(*ttnpb.EndDevice).Session.SessionKeys.SessionKeyID,
							InvalidatedDownlinks: tc.Device.QueuedApplicationDownlinks,
							SessionStartedAt:     pb.Session.StartedAt,
						}},
					})

				case <-time.After(Timeout):
					t.Fatal("Timed out while waiting for join to be sent to AS")
				}

				select {
				case req := <-downlinkAddCh:
					a.So(req.ctx, should.HaveParentContext, ctx)
					a.So(req.devID, should.Resemble, pb.EndDeviceIdentifiers)
					a.So([]time.Time{start, req.t, time.Now()}, should.BeChronological)

				case <-time.After(Timeout):
					t.Fatal("Timeout waiting for Add to be called")
				}

				_ = sendUplinkDuplicates(t, ns, collectionDoneCh, ctx, tc.UplinkMessage, DuplicateCount)
				close(collectionDoneCh)

				select {
				case err := <-errch:
					a.So(err, should.BeNil)

				case <-time.After(Timeout):
					t.Fatal("Timed out while waiting for HandleUplink to return")
				}

				deduplicationDoneCh = make(chan windowEnd, 1)
				collectionDoneCh = make(chan windowEnd, 1)

				t.Run("after cooldown window", func(t *testing.T) {
					a := assertions.New(t)

					errch := make(chan error, 1)
					go func() {
						_, err = ns.HandleUplink(ctx, CopyUplinkMessage(tc.UplinkMessage))
						errch <- err
					}()

					select {
					case req := <-handleJoinCh:
						a.So(req.req.EndDeviceIdentifiers.DevAddr, should.NotResemble, pb.Session.DevAddr)

						expectedRequest.EndDeviceIdentifiers.DevAddr = req.req.EndDeviceIdentifiers.DevAddr
						a.So(req.req, should.Resemble, expectedRequest)

						resp := ttnpb.NewPopulatedJoinResponse(test.Randy, false)
						resp.SessionKeys = *keys

						req.ch <- resp
						req.errch <- nil

					case err := <-errch:
						a.So(err, should.BeNil)
						t.Fatal("Join not sent to JS")

					case <-time.After(Timeout):
						t.Fatal("Timed out while waiting for join to be sent to JS")
					}

					_ = sendUplinkDuplicates(t, ns, deduplicationDoneCh, ctx, tc.UplinkMessage, DuplicateCount)
					close(deduplicationDoneCh)

					select {
					case up := <-asSendCh:
						a.So(up.CorrelationIDs, should.NotBeEmpty)

						a.So(up, should.HaveEmptyDiff, &ttnpb.ApplicationUp{
							CorrelationIDs: up.CorrelationIDs,
							EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
								DevAddr:                expectedRequest.EndDeviceIdentifiers.DevAddr,
								DevEUI:                 tc.Device.EndDeviceIdentifiers.DevEUI,
								DeviceID:               tc.Device.EndDeviceIdentifiers.DeviceID,
								JoinEUI:                tc.Device.EndDeviceIdentifiers.JoinEUI,
								ApplicationIdentifiers: tc.Device.EndDeviceIdentifiers.ApplicationIdentifiers,
							},
							Up: &ttnpb.ApplicationUp_JoinAccept{JoinAccept: &ttnpb.ApplicationJoinAccept{
								AppSKey:          resp.SessionKeys.AppSKey,
								SessionKeyID:     test.Must(devReg.GetByID(ctx, tc.Device.EndDeviceIdentifiers.ApplicationIdentifiers, tc.Device.EndDeviceIdentifiers.DeviceID, ttnpb.EndDeviceFieldPathsTopLevel)).(*ttnpb.EndDevice).Session.SessionKeys.SessionKeyID,
								SessionStartedAt: up.GetJoinAccept().SessionStartedAt,
							}},
						})

					case <-time.After(Timeout):
						t.Fatal("Timed out while waiting for join to be sent to AS")
					}

					pb.EndDeviceIdentifiers.DevAddr = expectedRequest.EndDeviceIdentifiers.DevAddr
					select {
					case req := <-downlinkAddCh:
						a.So(req.ctx, should.HaveParentContext, ctx)
						a.So(req.devID, should.Resemble, pb.EndDeviceIdentifiers)
						a.So([]time.Time{start, req.t, time.Now()}, should.BeChronological)

					case <-time.After(Timeout):
						t.Fatal("Timeout waiting for Add to be called")
					}

					_ = sendUplinkDuplicates(t, ns, collectionDoneCh, ctx, tc.UplinkMessage, DuplicateCount)
					close(collectionDoneCh)

					select {
					case err := <-errch:
						a.So(err, should.BeNil)

					case <-time.After(Timeout):
						t.Fatal("Timed out while waiting for HandleUplink to return")
					}
				})
			})
		}
	}
}

func handleRejoinTest() func(t *testing.T) {
	return func(t *testing.T) {
		// TODO: Implement https://github.com/TheThingsIndustries/ttn/issues/557
	}
}

func TestHandleUplink(t *testing.T) {
	a := assertions.New(t)

	authorizedCtx := clusterauth.NewContext(test.Context(), nil)

	redisClient, flush := test.NewRedis(t, "networkserver_test")
	defer flush()
	defer redisClient.Close()
	devReg := &redis.DeviceRegistry{Redis: redisClient}

	ns := test.Must(New(
		component.MustNew(test.GetLogger(t), &component.Config{}),
		&Config{
			Devices:             devReg,
			DeduplicationWindow: 42,
			CooldownWindow:      42,
			DownlinkTasks:       &MockDownlinkTaskQueue{},
		},
	)).(*NetworkServer)
	test.Must(nil, ns.Start())
	defer ns.Close()

	msg := ttnpb.NewPopulatedUplinkMessage(test.Randy, false)
	msg.Payload.Payload = nil
	msg.RawPayload = nil
	_, err := ns.HandleUplink(authorizedCtx, msg)
	a.So(err, should.NotBeNil)

	msg = ttnpb.NewPopulatedUplinkMessage(test.Randy, false)
	msg.Payload.Payload = nil
	msg.RawPayload = []byte{}
	_, err = ns.HandleUplink(authorizedCtx, msg)
	a.So(err, should.NotBeNil)

	msg = ttnpb.NewPopulatedUplinkMessage(test.Randy, false)
	msg.Payload.Major = 1
	_, err = ns.HandleUplink(authorizedCtx, msg)
	a.So(err, should.NotBeNil)

	t.Run("Uplink", handleUplinkTest())
	t.Run("Join", handleJoinTest())
	t.Run("Rejoin", handleRejoinTest())
}
