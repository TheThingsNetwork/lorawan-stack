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
	"context"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"testing"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/kr/pretty"
	"github.com/mohae/deepcopy"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/config"
	"go.thethings.network/lorawan-stack/pkg/crypto"
	"go.thethings.network/lorawan-stack/pkg/deviceregistry"
	"go.thethings.network/lorawan-stack/pkg/errors"
	. "go.thethings.network/lorawan-stack/pkg/networkserver"
	"go.thethings.network/lorawan-stack/pkg/store"
	"go.thethings.network/lorawan-stack/pkg/store/mapstore"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
	ttnshould "go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var (
	FNwkSIntKey   = types.AES128Key{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	SNwkSIntKey   = types.AES128Key{0x42, 0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	NwkSEncKey    = types.AES128Key{0x42, 0x42, 0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	DevAddr       = types.DevAddr{0x42, 0x42, 0xff, 0xff}
	DevEUI        = types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	JoinEUI       = types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	ApplicationID = "test"

	DuplicateCount = 6
	DeviceCount    = 100
	Timeout        = (2 << 10) * test.Delay

	Keys = []string{"AEAEAEAEAEAEAEAEAEAEAEAEAEAEAEAE"}
)

func contextWithKey() context.Context {
	return metadata.NewIncomingContext(context.Background(), metadata.MD{
		"authorization": []string{fmt.Sprintf("Basic %s", Keys[0])},
	})
}

func metadataLdiff(l pretty.Logfer, xs, ys []*ttnpb.RxMetadata) {
	if len(xs) != len(ys) {
		l.Logf("Length mismatch: %d != %d", len(xs), len(ys))
		return
	}

	xm := make(map[*ttnpb.RxMetadata]struct{})
	for _, x := range xs {
		xm[x] = struct{}{}
	}

	ym := make(map[*ttnpb.RxMetadata]struct{})
	for _, y := range ys {
		ym[y] = struct{}{}
	}
	pretty.Ldiff(l, xm, ym)
}

func TestDownlinkQueueReplace(t *testing.T) {
	a := assertions.New(t)
	reg := deviceregistry.New(store.NewTypedMapStoreClient(mapstore.New()))
	ns := test.Must(New(
		component.MustNew(test.GetLogger(t), &component.Config{}),
		&Config{
			Registry:            reg,
			JoinServers:         nil,
			DeduplicationWindow: 42,
			CooldownWindow:      42,
		})).(*NetworkServer)
	test.Must(nil, ns.Start())

	ed := ttnpb.NewPopulatedEndDevice(test.Randy, false)
	ed.QueuedApplicationDownlinks = nil

	dev, err := reg.Create(ed)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	_, err = ns.DownlinkQueueReplace(context.Background(), &ttnpb.DownlinkQueueRequest{})
	a.So(err, should.NotBeNil)

	req := ttnpb.NewPopulatedDownlinkQueueRequest(test.Randy, false)
	req.EndDeviceIdentifiers = ed.EndDeviceIdentifiers

	_, err = ns.DownlinkQueueReplace(context.Background(), req)
	a.So(err, should.BeNil)

	dev, err = dev.Load()
	if !a.So(err, should.BeNil) ||
		!a.So(dev.EndDevice, should.NotBeNil) {
		t.FailNow()
	}

	a.So(pretty.Diff(dev.QueuedApplicationDownlinks, req.Downlinks), should.BeEmpty)

	req = ttnpb.NewPopulatedDownlinkQueueRequest(test.Randy, false)
	for len(req.Downlinks) == 0 {
		req = ttnpb.NewPopulatedDownlinkQueueRequest(test.Randy, false)
	}
	req.EndDeviceIdentifiers = ed.EndDeviceIdentifiers

	_, err = ns.DownlinkQueueReplace(context.Background(), req)
	a.So(err, should.BeNil)

	dev, err = dev.Load()
	if !a.So(err, should.BeNil) ||
		!a.So(dev.EndDevice, should.NotBeNil) {
		t.FailNow()
	}

	a.So(pretty.Diff(dev.QueuedApplicationDownlinks, req.Downlinks), should.BeEmpty)
}

func TestDownlinkQueuePush(t *testing.T) {
	a := assertions.New(t)
	reg := deviceregistry.New(store.NewTypedMapStoreClient(mapstore.New()))
	ns := test.Must(New(
		component.MustNew(test.GetLogger(t), &component.Config{}),
		&Config{
			Registry:            reg,
			JoinServers:         nil,
			DeduplicationWindow: 42,
			CooldownWindow:      42,
		})).(*NetworkServer)
	test.Must(nil, ns.Start())

	ed := ttnpb.NewPopulatedEndDevice(test.Randy, false)
	ed.QueuedApplicationDownlinks = nil

	dev, err := reg.Create(ed)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	_, err = ns.DownlinkQueuePush(context.Background(), &ttnpb.DownlinkQueueRequest{})
	a.So(err, should.NotBeNil)

	req := ttnpb.NewPopulatedDownlinkQueueRequest(test.Randy, false)
	for len(req.Downlinks) == 0 {
		req = ttnpb.NewPopulatedDownlinkQueueRequest(test.Randy, false)
	}
	req.EndDeviceIdentifiers = ed.EndDeviceIdentifiers

	downlinks := append(ed.QueuedApplicationDownlinks, req.Downlinks...)

	_, err = ns.DownlinkQueuePush(context.Background(), req)
	a.So(err, should.BeNil)

	dev, err = dev.Load()
	if !a.So(err, should.BeNil) ||
		!a.So(dev.EndDevice, should.NotBeNil) {
		t.FailNow()
	}

	a.So(pretty.Diff(dev.QueuedApplicationDownlinks, downlinks), should.BeEmpty)

	req = ttnpb.NewPopulatedDownlinkQueueRequest(test.Randy, false)
	req.EndDeviceIdentifiers = ed.EndDeviceIdentifiers
	downlinks = append(downlinks, req.Downlinks...)

	_, err = ns.DownlinkQueuePush(context.Background(), req)
	a.So(err, should.BeNil)

	dev, err = dev.Load()
	if !a.So(err, should.BeNil) ||
		!a.So(dev.EndDevice, should.NotBeNil) {
		t.FailNow()
	}
	a.So(pretty.Diff(dev.QueuedApplicationDownlinks, downlinks), should.BeEmpty)
}

func TestDownlinkQueueList(t *testing.T) {
	a := assertions.New(t)
	reg := deviceregistry.New(store.NewTypedMapStoreClient(mapstore.New()))
	ns := test.Must(New(
		component.MustNew(test.GetLogger(t), &component.Config{}),
		&Config{
			Registry:            reg,
			JoinServers:         nil,
			DeduplicationWindow: 42,
			CooldownWindow:      42,
		})).(*NetworkServer)
	test.Must(nil, ns.Start())

	ed := ttnpb.NewPopulatedEndDevice(test.Randy, false)
	ed.QueuedApplicationDownlinks = nil

	dev, err := reg.Create(ed)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	_, err = ns.DownlinkQueueList(context.Background(), &ttnpb.EndDeviceIdentifiers{})
	a.So(err, should.NotBeNil)

	downlinks, err := ns.DownlinkQueueList(context.Background(), &dev.EndDeviceIdentifiers)
	a.So(err, should.BeNil)
	a.So(pretty.Diff(downlinks, &ttnpb.ApplicationDownlinks{Downlinks: ed.QueuedApplicationDownlinks}), should.BeEmpty)

	ed = ttnpb.NewPopulatedEndDevice(test.Randy, false)
	for len(ed.QueuedApplicationDownlinks) == 0 {
		ed = ttnpb.NewPopulatedEndDevice(test.Randy, false)
	}
	ed.EndDeviceIdentifiers = dev.EndDeviceIdentifiers
	dev.EndDevice = ed

	err = dev.Store()
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	downlinks, err = ns.DownlinkQueueList(context.Background(), &dev.EndDeviceIdentifiers)
	a.So(err, should.BeNil)
	a.So(pretty.Diff(downlinks, &ttnpb.ApplicationDownlinks{Downlinks: ed.QueuedApplicationDownlinks}), should.BeEmpty)
}

func TestDownlinkQueueClear(t *testing.T) {
	a := assertions.New(t)
	reg := deviceregistry.New(store.NewTypedMapStoreClient(mapstore.New()))
	ns := test.Must(New(
		component.MustNew(test.GetLogger(t), &component.Config{}),
		&Config{
			Registry:            reg,
			JoinServers:         nil,
			DeduplicationWindow: 42,
			CooldownWindow:      42,
		})).(*NetworkServer)
	test.Must(nil, ns.Start())

	ed := ttnpb.NewPopulatedEndDevice(test.Randy, false)
	ed.QueuedApplicationDownlinks = nil

	dev, err := reg.Create(ed)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	e, err := ns.DownlinkQueueClear(context.Background(), &ttnpb.EndDeviceIdentifiers{})
	a.So(err, should.NotBeNil)
	a.So(e, should.BeNil)

	e, err = ns.DownlinkQueueClear(context.Background(), &dev.EndDeviceIdentifiers)
	a.So(err, should.BeNil)
	a.So(e, should.NotBeNil)

	dev, err = dev.Load()
	if !a.So(err, should.BeNil) ||
		!a.So(dev.EndDevice, should.NotBeNil) {
		t.FailNow()
	}
	a.So(dev.QueuedApplicationDownlinks, should.BeEmpty)

	ed = ttnpb.NewPopulatedEndDevice(test.Randy, false)
	for len(ed.QueuedApplicationDownlinks) == 0 {
		ed = ttnpb.NewPopulatedEndDevice(test.Randy, false)
	}
	ed.EndDeviceIdentifiers = dev.EndDeviceIdentifiers
	dev.EndDevice = ed

	err = dev.Store()
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	e, err = ns.DownlinkQueueClear(context.Background(), &dev.EndDeviceIdentifiers)
	a.So(err, should.BeNil)
	a.So(e, should.NotBeNil)

	dev, err = dev.Load()
	if !a.So(err, should.BeNil) ||
		!a.So(dev.EndDevice, should.NotBeNil) {
		t.FailNow()
	}
	a.So(dev.QueuedApplicationDownlinks, should.BeEmpty)
}

type UplinkHandler interface {
	HandleUplink(context.Context, *ttnpb.UplinkMessage) (*pbtypes.Empty, error)
}

func sendUplinkDuplicates(t *testing.T, h UplinkHandler, windowEndCh chan windowEnd, ctx context.Context, msg *ttnpb.UplinkMessage, n int, waitForFirst bool) ([]*ttnpb.RxMetadata, <-chan error) {
	t.Helper()

	a := assertions.New(t)
	msg = deepcopy.Copy(msg).(*ttnpb.UplinkMessage)

	errch := make(chan error, 1)
	go func(msg *ttnpb.UplinkMessage) {
		_, err := h.HandleUplink(ctx, msg)
		errch <- err
	}(deepcopy.Copy(msg).(*ttnpb.UplinkMessage))

	var weCh chan<- time.Time
	select {
	case we := <-windowEndCh:
		a.So(we.msg.ReceivedAt, should.HappenBefore, time.Now())
		msg.ReceivedAt = we.msg.ReceivedAt

		a.So(we.msg.CorrelationIDs, should.NotBeEmpty)
		msg.CorrelationIDs = we.msg.CorrelationIDs

		a.So(we.msg, should.Resemble, msg)
		a.So(we.ctx, ttnshould.HaveParentContext, ctx)
		weCh = we.ch

	case <-time.After(Timeout):
		select {
		case err := <-errch:
			t.Fatalf("Error processing first message: %s", err)

		default:
			t.Fatal("Timed out while waiting for first message to arrive")
		}
	}

	mdCh := make(chan *ttnpb.RxMetadata)

	wg := &sync.WaitGroup{}
	wg.Add(n)

	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()

			msg := deepcopy.Copy(msg).(*ttnpb.UplinkMessage)

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
		wg.Wait()

		if waitForFirst {
			select {
			case errch <- <-errch:

			case <-time.After(Timeout):
				t.Fatal("Timed out while waiting for first call to HandleUplink to return ")
			}
		}

		select {
		case weCh <- time.Now():

		case <-time.After(Timeout):
			t.Fatal("Timed out while waiting for metadata collection to stop")
		}

		close(mdCh)
	}()

	mds := append([]*ttnpb.RxMetadata{}, msg.RxMetadata...)
	for md := range mdCh {
		mds = append(mds, md)
	}
	return mds, errch
}

type MockAsNsLinkApplicationStream struct {
	*test.MockServerStream
	SendFunc func(*ttnpb.ApplicationUp) error
}

func (s *MockAsNsLinkApplicationStream) Send(msg *ttnpb.ApplicationUp) error {
	if s.SendFunc == nil {
		return nil
	}
	return s.SendFunc(msg)
}

func TestLinkApplication(t *testing.T) {
	a := assertions.New(t)
	reg := deviceregistry.New(store.NewTypedMapStoreClient(mapstore.New()))
	ns := test.Must(New(
		component.MustNew(test.GetLogger(t), &component.Config{
			ServiceBase: config.ServiceBase{Cluster: config.Cluster{Keys: Keys}},
		}),
		&Config{
			Registry:            reg,
			JoinServers:         nil,
			DeduplicationWindow: 42,
			CooldownWindow:      42,
		})).(*NetworkServer)
	test.Must(nil, ns.Start())

	id := ttnpb.NewPopulatedApplicationIdentifiers(test.Randy, false)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	sendFunc := func(*ttnpb.ApplicationUp) error {
		t.Error("Send should not be called")
		return nil
	}

	time.AfterFunc(test.Delay, func() {
		err := ns.LinkApplication(id, &MockAsNsLinkApplicationStream{
			MockServerStream: &test.MockServerStream{
				MockStream: &test.MockStream{
					ContextFunc: func() context.Context {
						ctx, cancel := context.WithCancel(contextWithKey())
						time.AfterFunc(test.Delay, cancel)
						return ctx
					},
				},
			},
			SendFunc: sendFunc,
		})
		a.So(err, should.Resemble, context.Canceled)
		wg.Done()
	})

	err := ns.LinkApplication(id, &MockAsNsLinkApplicationStream{
		MockServerStream: &test.MockServerStream{
			MockStream: &test.MockStream{
				ContextFunc: contextWithKey,
			},
		},
		SendFunc: sendFunc,
	})
	a.So(err, should.NotBeNil)

	wg.Wait()
}

type windowEnd struct {
	ctx context.Context
	msg *ttnpb.UplinkMessage
	ch  chan<- time.Time
}

func HandleUplinkTest(conf *component.Config) func(t *testing.T) {
	return func(t *testing.T) {
		a := assertions.New(t)

		reg := deviceregistry.New(store.NewTypedMapStoreClient(mapstore.New()))
		ns := test.Must(New(
			component.MustNew(test.GetLogger(t), conf),
			&Config{
				Registry:            reg,
				JoinServers:         nil,
				DeduplicationWindow: 42,
				CooldownWindow:      42,
			})).(*NetworkServer)
		test.Must(nil, ns.Start())

		_, err := reg.Create(&ttnpb.EndDevice{
			LoRaWANVersion: ttnpb.MAC_V1_1,
			Session: &ttnpb.Session{
				DevAddr: DevAddr,
				SessionKeys: ttnpb.SessionKeys{
					FNwkSIntKey: &ttnpb.KeyEnvelope{
						Key: &FNwkSIntKey,
					},
					SNwkSIntKey: &ttnpb.KeyEnvelope{
						Key: &SNwkSIntKey,
					},
				},
			},
			FrequencyPlanID: test.EUFrequencyPlanID,
		})
		if !a.So(err, should.BeNil) {
			t.FailNow()
		}

		t.Run("Empty DevAddr", func(t *testing.T) {
			a := assertions.New(t)

			msg := ttnpb.NewPopulatedUplinkMessageUplink(test.Randy, SNwkSIntKey, FNwkSIntKey, false)
			msg.Payload.GetMACPayload().DevAddr = types.DevAddr{}
			_, err := ns.HandleUplink(context.Background(), msg)
			a.So(err, should.NotBeNil)
		})

		t.Run("FCnt too high", func(t *testing.T) {
			a := assertions.New(t)

			msg := ttnpb.NewPopulatedUplinkMessageUplink(test.Randy, SNwkSIntKey, FNwkSIntKey, false)
			msg.Payload.GetMACPayload().DevAddr = DevAddr
			msg.Payload.GetMACPayload().FCnt = math.MaxUint16 + 1
			_, err := ns.HandleUplink(context.Background(), msg)
			a.So(err, should.NotBeNil)
		})

		t.Run("ChannelIndex too high", func(t *testing.T) {
			a := assertions.New(t)

			msg := ttnpb.NewPopulatedUplinkMessageUplink(test.Randy, SNwkSIntKey, FNwkSIntKey, false)
			msg.Payload.GetMACPayload().DevAddr = DevAddr
			msg.Settings.ChannelIndex = math.MaxUint8 + 1
			_, err := ns.HandleUplink(context.Background(), msg)
			a.So(err, should.NotBeNil)
		})

		t.Run("DataRateIndex too high", func(t *testing.T) {
			a := assertions.New(t)

			msg := ttnpb.NewPopulatedUplinkMessageUplink(test.Randy, SNwkSIntKey, FNwkSIntKey, false)
			msg.Payload.GetMACPayload().DevAddr = DevAddr
			msg.Settings.DataRateIndex = math.MaxUint8 + 1
			_, err = ns.HandleUplink(context.Background(), msg)
			a.So(err, should.NotBeNil)
		})

		for _, tc := range []struct {
			Name string

			Device         *ttnpb.EndDevice
			NextNextFCntUp uint32
			UplinkMessage  *ttnpb.UplinkMessage
		}{
			{
				"1.0/unconfirmed",
				&ttnpb.EndDevice{
					LoRaWANVersion: ttnpb.MAC_V1_0,
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
							ApplicationID: ApplicationID,
						},
						DevAddr: &DevAddr,
					},
					FrequencyPlanID: test.EUFrequencyPlanID,
					Session: &ttnpb.Session{
						DevAddr:    DevAddr,
						NextFCntUp: 0x42,
						SessionKeys: ttnpb.SessionKeys{
							FNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: &FNwkSIntKey,
							},
						},
					},
				},
				0x43,
				func() *ttnpb.UplinkMessage {
					msg := ttnpb.NewPopulatedUplinkMessageUplink(test.Randy, SNwkSIntKey, FNwkSIntKey, false)

					pld := msg.Payload.GetMACPayload()
					pld.DevAddr = DevAddr
					pld.FCnt = 0x42

					msg.Payload.MIC = nil
					mic := test.Must(crypto.ComputeLegacyUplinkMIC(FNwkSIntKey, DevAddr, 0x42, test.Must(msg.Payload.MarshalLoRaWAN()).([]byte))).([4]byte)
					msg.Payload.MIC = mic[:]
					msg.RawPayload = test.Must(msg.Payload.MarshalLoRaWAN()).([]byte)

					return msg
				}(),
			},
			{
				"1.0/unconfirmed/FCnt resets",
				&ttnpb.EndDevice{
					LoRaWANVersion: ttnpb.MAC_V1_0,
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
							ApplicationID: ApplicationID,
						},
						DevAddr: &DevAddr,
					},
					FrequencyPlanID: test.EUFrequencyPlanID,
					Session: &ttnpb.Session{
						DevAddr:    DevAddr,
						NextFCntUp: 0x42424249,
						SessionKeys: ttnpb.SessionKeys{
							FNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: &FNwkSIntKey,
							},
						},
					},
					FCntResets: true,
				},
				0x43,
				func() *ttnpb.UplinkMessage {
					msg := ttnpb.NewPopulatedUplinkMessageUplink(test.Randy, SNwkSIntKey, FNwkSIntKey, false)

					pld := msg.Payload.GetMACPayload()
					pld.DevAddr = DevAddr
					pld.FCnt = 0x42

					msg.Payload.MIC = nil
					mic := test.Must(crypto.ComputeLegacyUplinkMIC(FNwkSIntKey, DevAddr, 0x42, test.Must(msg.Payload.MarshalLoRaWAN()).([]byte))).([4]byte)
					msg.Payload.MIC = mic[:]
					msg.RawPayload = test.Must(msg.Payload.MarshalLoRaWAN()).([]byte)

					return msg
				}(),
			},
			{
				"1.0/confirmed",
				&ttnpb.EndDevice{
					LoRaWANVersion: ttnpb.MAC_V1_0,
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
							ApplicationID: ApplicationID,
						},
						DevAddr: &DevAddr,
					},
					FrequencyPlanID: test.EUFrequencyPlanID,
					Session: &ttnpb.Session{
						DevAddr:    DevAddr,
						NextFCntUp: 0x42,
						SessionKeys: ttnpb.SessionKeys{
							FNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: &FNwkSIntKey,
							},
						},
					},
					RecentDownlinks: []*ttnpb.DownlinkMessage{
						ttnpb.NewPopulatedDownlinkMessage(test.Randy, false),
						ttnpb.NewPopulatedDownlinkMessage(test.Randy, false),
						ttnpb.NewPopulatedDownlinkMessage(test.Randy, false),
					},
				},
				0x43,
				func() *ttnpb.UplinkMessage {
					msg := ttnpb.NewPopulatedUplinkMessageUplink(test.Randy, SNwkSIntKey, FNwkSIntKey, true)

					pld := msg.Payload.GetMACPayload()
					pld.DevAddr = DevAddr
					pld.FCnt = 0x42

					msg.Payload.MIC = nil
					mic := test.Must(crypto.ComputeLegacyUplinkMIC(FNwkSIntKey, DevAddr, 0x42, test.Must(msg.Payload.MarshalLoRaWAN()).([]byte))).([4]byte)
					msg.Payload.MIC = mic[:]
					msg.RawPayload = test.Must(msg.Payload.MarshalLoRaWAN()).([]byte)

					return msg
				}(),
			},
			{
				"1.0/confirmed/FCnt resets",
				&ttnpb.EndDevice{
					LoRaWANVersion: ttnpb.MAC_V1_0,
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
							ApplicationID: ApplicationID,
						},
						DevAddr: &DevAddr,
					},
					FrequencyPlanID: test.EUFrequencyPlanID,
					Session: &ttnpb.Session{
						DevAddr:    DevAddr,
						NextFCntUp: 0x42424249,
						SessionKeys: ttnpb.SessionKeys{
							FNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: &FNwkSIntKey,
							},
						},
					},
					RecentDownlinks: []*ttnpb.DownlinkMessage{
						ttnpb.NewPopulatedDownlinkMessage(test.Randy, false),
						ttnpb.NewPopulatedDownlinkMessage(test.Randy, false),
						ttnpb.NewPopulatedDownlinkMessage(test.Randy, false),
					},
					FCntResets: true,
				},
				0x43,
				func() *ttnpb.UplinkMessage {
					msg := ttnpb.NewPopulatedUplinkMessageUplink(test.Randy, SNwkSIntKey, FNwkSIntKey, true)

					pld := msg.Payload.GetMACPayload()
					pld.DevAddr = DevAddr
					pld.FCnt = 0x42

					msg.Payload.MIC = nil
					mic := test.Must(crypto.ComputeLegacyUplinkMIC(FNwkSIntKey, DevAddr, 0x42, test.Must(msg.Payload.MarshalLoRaWAN()).([]byte))).([4]byte)
					msg.Payload.MIC = mic[:]
					msg.RawPayload = test.Must(msg.Payload.MarshalLoRaWAN()).([]byte)

					return msg
				}(),
			},
			{
				"1.1/unconfirmed",
				&ttnpb.EndDevice{
					LoRaWANVersion: ttnpb.MAC_V1_1,
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
							ApplicationID: ApplicationID,
						},
						DevAddr: &DevAddr,
					},
					FrequencyPlanID: test.EUFrequencyPlanID,
					Session: &ttnpb.Session{
						DevAddr:    DevAddr,
						NextFCntUp: 0x42,
						SessionKeys: ttnpb.SessionKeys{
							FNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: &FNwkSIntKey,
							},
							SNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: &SNwkSIntKey,
							},
							NwkSEncKey: &ttnpb.KeyEnvelope{
								Key: &NwkSEncKey,
							},
						},
					},
				},
				0x43,
				func() *ttnpb.UplinkMessage {
					msg := ttnpb.NewPopulatedUplinkMessageUplink(test.Randy, SNwkSIntKey, FNwkSIntKey, false)

					pld := msg.Payload.GetMACPayload()
					pld.DevAddr = DevAddr
					pld.FCnt = 0x42

					msg.Payload.MIC = nil
					mic := test.Must(crypto.ComputeUplinkMIC(SNwkSIntKey, FNwkSIntKey, 0,
						uint8(msg.Settings.DataRateIndex), uint8(msg.Settings.ChannelIndex),
						DevAddr, 0x42, test.Must(msg.Payload.MarshalLoRaWAN()).([]byte))).([4]byte)
					msg.Payload.MIC = mic[:]
					msg.RawPayload = test.Must(msg.Payload.MarshalLoRaWAN()).([]byte)

					return msg
				}(),
			},
			{
				"1.1/confirmed",
				&ttnpb.EndDevice{
					LoRaWANVersion: ttnpb.MAC_V1_1,
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
							ApplicationID: ApplicationID,
						},
						DevAddr: &DevAddr,
					},
					FrequencyPlanID: test.EUFrequencyPlanID,
					Session: &ttnpb.Session{
						DevAddr:    DevAddr,
						NextFCntUp: 0x42,
						SessionKeys: ttnpb.SessionKeys{
							FNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: &FNwkSIntKey,
							},
							SNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: &SNwkSIntKey,
							},
							NwkSEncKey: &ttnpb.KeyEnvelope{
								Key: &NwkSEncKey,
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
				0x43,
				func() *ttnpb.UplinkMessage {
					msg := ttnpb.NewPopulatedUplinkMessageUplink(test.Randy, SNwkSIntKey, FNwkSIntKey, true)

					pld := msg.Payload.GetMACPayload()
					pld.DevAddr = DevAddr
					pld.FCnt = 0x42

					msg.Payload.MIC = nil
					mic := test.Must(crypto.ComputeUplinkMIC(SNwkSIntKey, FNwkSIntKey, 0x24,
						uint8(msg.Settings.DataRateIndex), uint8(msg.Settings.ChannelIndex),
						DevAddr, 0x42, test.Must(msg.Payload.MarshalLoRaWAN()).([]byte))).([4]byte)
					msg.Payload.MIC = mic[:]
					msg.RawPayload = test.Must(msg.Payload.MarshalLoRaWAN()).([]byte)

					return msg
				}(),
			},
			{
				"1.1/unconfirmed/FCnt resets",
				&ttnpb.EndDevice{
					LoRaWANVersion: ttnpb.MAC_V1_1,
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
							ApplicationID: ApplicationID,
						},
						DevAddr: &DevAddr,
					},
					FrequencyPlanID: test.EUFrequencyPlanID,
					Session: &ttnpb.Session{
						DevAddr:    DevAddr,
						NextFCntUp: 0x42424249,
						SessionKeys: ttnpb.SessionKeys{
							FNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: &FNwkSIntKey,
							},
							SNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: &SNwkSIntKey,
							},
							NwkSEncKey: &ttnpb.KeyEnvelope{
								Key: &NwkSEncKey,
							},
						},
					},
					FCntResets: true,
				},
				0x43,
				func() *ttnpb.UplinkMessage {
					msg := ttnpb.NewPopulatedUplinkMessageUplink(test.Randy, SNwkSIntKey, FNwkSIntKey, false)

					pld := msg.Payload.GetMACPayload()
					pld.DevAddr = DevAddr
					pld.FCnt = 0x42

					msg.Payload.MIC = nil
					mic := test.Must(crypto.ComputeUplinkMIC(SNwkSIntKey, FNwkSIntKey, 0,
						uint8(msg.Settings.DataRateIndex), uint8(msg.Settings.ChannelIndex),
						DevAddr, 0x42, test.Must(msg.Payload.MarshalLoRaWAN()).([]byte))).([4]byte)
					msg.Payload.MIC = mic[:]
					msg.RawPayload = test.Must(msg.Payload.MarshalLoRaWAN()).([]byte)

					return msg
				}(),
			},
			{
				"1.1/confirmed/FCnt resets",
				&ttnpb.EndDevice{
					LoRaWANVersion: ttnpb.MAC_V1_1,
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
							ApplicationID: ApplicationID,
						},
						DevAddr: &DevAddr,
					},
					FrequencyPlanID: test.EUFrequencyPlanID,
					Session: &ttnpb.Session{
						DevAddr:    DevAddr,
						NextFCntUp: 0x42424249,
						SessionKeys: ttnpb.SessionKeys{
							FNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: &FNwkSIntKey,
							},
							SNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: &SNwkSIntKey,
							},
							NwkSEncKey: &ttnpb.KeyEnvelope{
								Key: &NwkSEncKey,
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
					FCntResets: true,
				},
				0x43,
				func() *ttnpb.UplinkMessage {
					msg := ttnpb.NewPopulatedUplinkMessageUplink(test.Randy, SNwkSIntKey, FNwkSIntKey, true)

					pld := msg.Payload.GetMACPayload()
					pld.DevAddr = DevAddr
					pld.FCnt = 0x42

					msg.Payload.MIC = nil
					mic := test.Must(crypto.ComputeUplinkMIC(SNwkSIntKey, FNwkSIntKey, 0x24,
						uint8(msg.Settings.DataRateIndex), uint8(msg.Settings.ChannelIndex),
						DevAddr, 0x42, test.Must(msg.Payload.MarshalLoRaWAN()).([]byte))).([4]byte)
					msg.Payload.MIC = mic[:]
					msg.RawPayload = test.Must(msg.Payload.MarshalLoRaWAN()).([]byte)

					return msg
				}(),
			},
		} {
			t.Run(tc.Name, func(t *testing.T) {
				a := assertions.New(t)

				reg := deviceregistry.New(store.NewTypedMapStoreClient(mapstore.New()))

				populateSessionKeys := func(s *ttnpb.Session) {
					for s.SessionKeys.FNwkSIntKey == nil ||
						s.SessionKeys.FNwkSIntKey.Key.IsZero() ||
						s.SessionKeys.FNwkSIntKey.Key.Equal(FNwkSIntKey) {

						s.SessionKeys.FNwkSIntKey = ttnpb.NewPopulatedKeyEnvelope(test.Randy, false)
					}

					for s.SessionKeys.SNwkSIntKey == nil ||
						s.SessionKeys.SNwkSIntKey.Key.IsZero() ||
						s.SessionKeys.SNwkSIntKey.Key.Equal(SNwkSIntKey) {

						s.SessionKeys.SNwkSIntKey = ttnpb.NewPopulatedKeyEnvelope(test.Randy, false)
					}
				}

				// Fill Registry with devices
				for i := 0; i < DeviceCount; i++ {
					ed := ttnpb.NewPopulatedEndDevice(test.Randy, false)
					for ed.Equal(tc.Device) {
						ed = ttnpb.NewPopulatedEndDevice(test.Randy, false)
					}

					if s := ed.Session; s != nil {
						populateSessionKeys(s)

						s.DevAddr = DevAddr
						for ed.SessionFallback != nil && ed.SessionFallback.DevAddr.Equal(s.DevAddr) {
							ed.SessionFallback.DevAddr = *types.NewPopulatedDevAddr(test.Randy)
						}
					} else if s := ed.SessionFallback; s != nil {
						populateSessionKeys(s)

						s.DevAddr = DevAddr
						for ed.Session != nil && ed.Session.DevAddr.Equal(s.DevAddr) {
							ed.Session.DevAddr = *types.NewPopulatedDevAddr(test.Randy)
						}
					}

					_, err := reg.Create(ed)
					if !a.So(err, should.BeNil) {
						t.FailNow()
					}
				}

				deduplicationDoneCh := make(chan windowEnd, 1)
				collectionDoneCh := make(chan windowEnd, 1)

				ns := test.Must(New(
					component.MustNew(test.GetLogger(t), conf),
					&Config{
						Registry:            reg,
						DeduplicationWindow: 42,
						CooldownWindow:      42,
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
				)).(*NetworkServer)
				test.Must(nil, ns.Start())

				asSendCh := make(chan *ttnpb.ApplicationUp)

				go func() {
					id := ttnpb.NewPopulatedApplicationIdentifiers(test.Randy, false)
					id.ApplicationID = ApplicationID

					err := ns.LinkApplication(id, &MockAsNsLinkApplicationStream{
						MockServerStream: &test.MockServerStream{
							MockStream: &test.MockStream{
								ContextFunc: context.Background,
							},
						},
						SendFunc: func(up *ttnpb.ApplicationUp) error {
							asSendCh <- up
							return nil
						},
					})
					// LinkApplication should not return
					t.Errorf("LinkApplication should not return, returned with error: %s", err)
				}()

				time.Sleep(test.Delay)

				dev, err := reg.Create(deepcopy.Copy(tc.Device).(*ttnpb.EndDevice))
				if !a.So(err, should.BeNil) {
					t.FailNow()
				}

				ctx := context.WithValue(context.Background(), "answer", 42)

				start := time.Now()
				if !t.Run("deduplication window", func(t *testing.T) {
					var md []*ttnpb.RxMetadata

					t.Run("message send", func(t *testing.T) {
						a := assertions.New(t)

						var errch <-chan error
						md, errch = sendUplinkDuplicates(t, ns, deduplicationDoneCh, ctx, tc.UplinkMessage, DuplicateCount, false)

						select {
						case up := <-asSendCh:
							if !a.So(md, should.HaveSameElementsDeep, up.GetUplinkMessage().RxMetadata) {
								metadataLdiff(t, up.GetUplinkMessage().RxMetadata, md)
							}
							a.So(up.GetUplinkMessage().CorrelationIDs, should.NotBeEmpty)

							expected := &ttnpb.ApplicationUp{
								EndDeviceIdentifiers: dev.EndDeviceIdentifiers,
								SessionKeyID:         dev.Session.SessionKeys.SessionKeyID,
								Up: &ttnpb.ApplicationUp_UplinkMessage{UplinkMessage: &ttnpb.ApplicationUplink{
									FCnt:           tc.NextNextFCntUp - 1,
									FPort:          tc.UplinkMessage.Payload.GetMACPayload().FPort,
									FRMPayload:     tc.UplinkMessage.Payload.GetMACPayload().FRMPayload,
									RxMetadata:     up.GetUplinkMessage().RxMetadata,
									CorrelationIDs: up.GetUplinkMessage().CorrelationIDs,
								}},
							}
							if !a.So(up, should.Resemble, expected) {
								pretty.Ldiff(t, up, expected)
							}

						case err := <-errch:
							a.So(err, should.BeNil)
							t.Fatal("Uplink not sent to AS")

						case <-time.After(Timeout):
							t.Fatal("Timeout")
						}

						select {
						case err := <-errch:
							a.So(err, should.BeNil)

						case <-time.After(Timeout):
							t.Fatal("Timeout")
						}
					})

					t.Run("device update", func(t *testing.T) {
						a := assertions.New(t)

						dev, err = dev.Load()
						if !a.So(err, should.BeNil) ||
							!a.So(dev.EndDevice, should.NotBeNil) {
							t.FailNow()
						}

						if !a.So(dev.RecentUplinks, should.NotBeEmpty) {
							t.FailNow()
						}

						msg := deepcopy.Copy(tc.UplinkMessage).(*ttnpb.UplinkMessage)
						msg.RxMetadata = md

						expected := deepcopy.Copy(tc.Device).(*ttnpb.EndDevice)
						expected.Session.NextFCntUp = tc.NextNextFCntUp
						expected.SessionFallback = nil
						expected.CreatedAt = dev.CreatedAt
						expected.UpdatedAt = dev.UpdatedAt
						if expected.MACState == nil {
							err := ResetMACState(ns.Component.FrequencyPlans, expected)
							if !a.So(err, should.BeNil) {
								t.FailNow()
							}
						}
						expected.MACState.ADRDataRateIndex = msg.Settings.DataRateIndex

						if expected.MACInfo == nil {
							expected.MACInfo = &ttnpb.MACInfo{}
						}
						if expected.MACSettings == nil {
							expected.MACSettings = &ttnpb.MACSettings{}
						}

						expected.RecentUplinks = append(expected.RecentUplinks, msg)
						if len(expected.RecentUplinks) > RecentUplinkCount {
							expected.RecentUplinks = expected.RecentUplinks[len(expected.RecentUplinks)-RecentUplinkCount:]
						}

						storedUp := dev.RecentUplinks[len(dev.RecentUplinks)-1]
						expectedUp := expected.RecentUplinks[len(expected.RecentUplinks)-1]

						a.So(storedUp.ReceivedAt, should.HappenBetween, start, time.Now())
						expectedUp.ReceivedAt = storedUp.ReceivedAt

						a.So(storedUp.CorrelationIDs, should.NotBeEmpty)
						expectedUp.CorrelationIDs = storedUp.CorrelationIDs

						if !a.So(storedUp.RxMetadata, should.HaveSameElementsDiff, expectedUp.RxMetadata) {
							metadataLdiff(t, storedUp.RxMetadata, expectedUp.RxMetadata)
						}
						expectedUp.RxMetadata = storedUp.RxMetadata

						a.So(pretty.Diff(dev.EndDevice, expected), should.BeEmpty)
					})
				}) {
					t.FailNow()
				}

				t.Run("cooldown window", func(t *testing.T) {
					a := assertions.New(t)

					_, errch := sendUplinkDuplicates(t, ns, collectionDoneCh, ctx, tc.UplinkMessage, DuplicateCount, true)

					select {
					case err := <-errch:
						a.So(err, should.BeNil)

					case <-time.After(Timeout):
						t.Fatal("Timeout")
					}
				})

				time.Sleep(test.Delay) // Ensure the message hash is removed from deduplication table

				t.Run("after cooldown window", func(t *testing.T) {
					a := assertions.New(t)

					msg := deepcopy.Copy(tc.UplinkMessage).(*ttnpb.UplinkMessage)
					if len(msg.RxMetadata) > 1 {
						// Deduplication may change the order of metadata in the slice
						msg.RxMetadata = []*ttnpb.RxMetadata{msg.RxMetadata[0]}
					}

					errch := make(chan error, 1)
					go func(ctx context.Context, msg *ttnpb.UplinkMessage) {
						_, err = ns.HandleUplink(ctx, msg)
						errch <- err
					}(ctx, deepcopy.Copy(msg).(*ttnpb.UplinkMessage))

					select {
					case err := <-errch:
						if !dev.FCntResets {
							a.So(err, should.BeError)
							return
						}

						a.So(err, should.BeNil)
						t.Fatal("Uplink not sent to AS")

					case de := <-deduplicationDoneCh:
						a.So(de.msg.ReceivedAt, should.HappenBetween, start, time.Now())
						msg.ReceivedAt = de.msg.ReceivedAt

						a.So(de.msg.CorrelationIDs, should.NotBeEmpty)
						msg.CorrelationIDs = de.msg.CorrelationIDs

						a.So(de.msg, should.Resemble, msg)
						a.So(de.ctx, ttnshould.HaveParentContext, ctx)

						de.ch <- time.Now()

					case <-time.After(Timeout):
						t.Fatal("Timeout")
					}

					select {
					case up := <-asSendCh:
						a.So(up, should.Resemble, &ttnpb.ApplicationUp{
							EndDeviceIdentifiers: dev.EndDeviceIdentifiers,
							SessionKeyID:         dev.Session.SessionKeys.SessionKeyID,
							Up: &ttnpb.ApplicationUp_UplinkMessage{UplinkMessage: &ttnpb.ApplicationUplink{
								FCnt:           tc.NextNextFCntUp - 1,
								FPort:          msg.Payload.GetMACPayload().FPort,
								FRMPayload:     msg.Payload.GetMACPayload().FRMPayload,
								RxMetadata:     msg.RxMetadata,
								CorrelationIDs: msg.CorrelationIDs,
							}},
						})

					case <-time.After(Timeout):
						t.Fatal("Timeout")
					}

					select {
					case err := <-errch:
						a.So(err, should.BeNil)

					case <-time.After(Timeout):
						t.Fatal("Timeout")
					}
				})
			})
		}
	}
}

var _ ttnpb.NsJsClient = &MockNsJsClient{}

type MockNsJsClient struct {
	HandleJoinFunc  func(ctx context.Context, req *ttnpb.JoinRequest, opts ...grpc.CallOption) (*ttnpb.JoinResponse, error)
	GetNwkSKeysFunc func(ctx context.Context, req *ttnpb.SessionKeyRequest, opts ...grpc.CallOption) (*ttnpb.NwkSKeysResponse, error)
}

func (c *MockNsJsClient) HandleJoin(ctx context.Context, req *ttnpb.JoinRequest, opts ...grpc.CallOption) (*ttnpb.JoinResponse, error) {
	return c.HandleJoinFunc(ctx, req, opts...)
}

func (c *MockNsJsClient) GetNwkSKeys(ctx context.Context, req *ttnpb.SessionKeyRequest, opts ...grpc.CallOption) (*ttnpb.NwkSKeysResponse, error) {
	return c.GetNwkSKeysFunc(ctx, req, opts...)
}

func HandleJoinTest(conf *component.Config) func(t *testing.T) {
	return func(t *testing.T) {
		a := assertions.New(t)

		reg := deviceregistry.New(store.NewTypedMapStoreClient(mapstore.New()))
		ns := test.Must(New(
			component.MustNew(test.GetLogger(t), conf),
			&Config{
				Registry:            reg,
				DeduplicationWindow: 42,
				CooldownWindow:      42,
			},
		)).(*NetworkServer)
		test.Must(nil, ns.Start())

		_, err := ns.HandleUplink(context.Background(), ttnpb.NewPopulatedUplinkMessageJoinRequest(test.Randy))
		a.So(err, should.NotBeNil)

		req := ttnpb.NewPopulatedUplinkMessageJoinRequest(test.Randy)
		ed := ttnpb.NewPopulatedEndDevice(test.Randy, false)

		ed.EndDeviceIdentifiers.DevEUI = &req.Payload.GetJoinRequestPayload().DevEUI
		ed.EndDeviceIdentifiers.JoinEUI = &req.Payload.GetJoinRequestPayload().JoinEUI

		_, err = reg.Create(ed)
		if !a.So(err, should.BeNil) {
			t.FailNow()
		}

		_, err = ns.HandleUplink(context.Background(), req)
		a.So(err, should.NotBeNil)

		for _, tc := range []struct {
			Name string

			Device        *ttnpb.EndDevice
			UplinkMessage *ttnpb.UplinkMessage
		}{
			{
				"1.1",
				&ttnpb.EndDevice{
					LoRaWANVersion: ttnpb.MAC_V1_1,
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						DevEUI:                 &DevEUI,
						JoinEUI:                &JoinEUI,
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: ApplicationID},
					},
					FrequencyPlanID: test.EUFrequencyPlanID,
					Session:         nil,
					MACStateDesired: ttnpb.NewPopulatedMACState(test.Randy, false),
				},
				func() *ttnpb.UplinkMessage {
					msg := ttnpb.NewPopulatedUplinkMessageJoinRequest(test.Randy)

					jr := msg.Payload.GetJoinRequestPayload()
					jr.DevEUI = DevEUI
					jr.JoinEUI = JoinEUI

					return msg
				}(),
			},
			{
				"1.1",
				&ttnpb.EndDevice{
					LoRaWANVersion: ttnpb.MAC_V1_1,
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						DevEUI:                 &DevEUI,
						JoinEUI:                &JoinEUI,
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: ApplicationID},
					},
					FrequencyPlanID: test.EUFrequencyPlanID,
					Session:         ttnpb.NewPopulatedSession(test.Randy, false),
					MACStateDesired: ttnpb.NewPopulatedMACState(test.Randy, false),
				},
				func() *ttnpb.UplinkMessage {
					msg := ttnpb.NewPopulatedUplinkMessageJoinRequest(test.Randy)

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
					},
					FrequencyPlanID: test.EUFrequencyPlanID,
					Session:         ttnpb.NewPopulatedSession(test.Randy, false),
					MACStateDesired: ttnpb.NewPopulatedMACState(test.Randy, false),
				},
				func() *ttnpb.UplinkMessage {
					msg := ttnpb.NewPopulatedUplinkMessageJoinRequest(test.Randy)

					jr := msg.Payload.GetJoinRequestPayload()
					jr.DevEUI = DevEUI
					jr.JoinEUI = JoinEUI

					return msg
				}(),
			},
		} {
			t.Run(tc.Name, func(t *testing.T) {
				a := assertions.New(t)

				reg := deviceregistry.New(store.NewTypedMapStoreClient(mapstore.New()))

				// Fill Registry with devices
				for i := 0; i < DeviceCount; i++ {
					ed := ttnpb.NewPopulatedEndDevice(test.Randy, false)
					for ed.Equal(tc.Device) {
						ed = ttnpb.NewPopulatedEndDevice(test.Randy, false)
					}

					_, err := reg.Create(ed)
					if !a.So(err, should.BeNil) {
						t.FailNow()
					}
				}

				keys := ttnpb.NewPopulatedSessionKeys(test.Randy, false)

				getNwkSKeysFunc := func(ctx context.Context, req *ttnpb.SessionKeyRequest, opts ...grpc.CallOption) (*ttnpb.NwkSKeysResponse, error) {
					err := errors.New("GetNwkSKeys should not be called")
					t.Fatal(err)
					return nil, err
				}

				type handleJoinRequest struct {
					ctx   context.Context
					req   *ttnpb.JoinRequest
					opts  []grpc.CallOption
					ch    chan<- *ttnpb.JoinResponse
					errch chan<- error
				}

				deduplicationDoneCh := make(chan windowEnd, 1)
				collectionDoneCh := make(chan windowEnd, 1)
				handleJoinCh := make(chan handleJoinRequest, 1)

				ns := test.Must(New(
					component.MustNew(test.GetLogger(t), conf),
					&Config{
						Registry: reg,
						JoinServers: []ttnpb.NsJsClient{&MockNsJsClient{
							HandleJoinFunc: func(ctx context.Context, req *ttnpb.JoinRequest, opts ...grpc.CallOption) (*ttnpb.JoinResponse, error) {
								return nil, errors.New("test")
							},
							GetNwkSKeysFunc: getNwkSKeysFunc,
						},
							&MockNsJsClient{
								HandleJoinFunc: func(ctx context.Context, req *ttnpb.JoinRequest, opts ...grpc.CallOption) (*ttnpb.JoinResponse, error) {
									ch := make(chan *ttnpb.JoinResponse, 1)
									errch := make(chan error, 1)
									handleJoinCh <- handleJoinRequest{ctx, req, opts, ch, errch}
									return <-ch, <-errch
								},
								GetNwkSKeysFunc: getNwkSKeysFunc,
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
				)).(*NetworkServer)

				test.Must(nil, ns.Start())

				asSendCh := make(chan *ttnpb.ApplicationUp)

				go func() {
					id := ttnpb.NewPopulatedApplicationIdentifiers(test.Randy, false)
					id.ApplicationID = ApplicationID

					err := ns.LinkApplication(id, &MockAsNsLinkApplicationStream{
						MockServerStream: &test.MockServerStream{
							MockStream: &test.MockStream{
								ContextFunc: context.Background,
							},
						},
						SendFunc: func(up *ttnpb.ApplicationUp) error {
							asSendCh <- up
							return nil
						},
					})
					// LinkApplication should not return
					t.Errorf("LinkApplication should not return, returned with error: %s", err)
				}()

				time.Sleep(test.Delay)

				dev, err := reg.Create(deepcopy.Copy(tc.Device).(*ttnpb.EndDevice))
				if !a.So(err, should.BeNil) {
					t.FailNow()
				}

				expectedRequest := &ttnpb.JoinRequest{
					RawPayload: tc.UplinkMessage.RawPayload,
					Payload:    tc.UplinkMessage.Payload,
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						DevEUI:  &DevEUI,
						JoinEUI: &JoinEUI,
					},
					NetID:              ns.NetID,
					SelectedMacVersion: tc.Device.LoRaWANVersion,
					RxDelay:            tc.Device.MACStateDesired.RxDelay,
					CFList:             nil,
					DownlinkSettings: ttnpb.DLSettings{
						Rx1DROffset: tc.Device.MACStateDesired.Rx1DataRateOffset,
						Rx2DR:       tc.Device.MACStateDesired.Rx2DataRateIndex,
					},
				}

				ctx := context.WithValue(context.Background(), "answer", 42)

				start := time.Now()
				if !t.Run("deduplication window", func(t *testing.T) {
					var md []*ttnpb.RxMetadata

					resp := ttnpb.NewPopulatedJoinResponse(test.Randy, false)
					resp.SessionKeys = *keys

					t.Run("message send", func(t *testing.T) {
						a := assertions.New(t)

						wg := &sync.WaitGroup{}
						wg.Add(1)
						go func() {
							defer wg.Done()

							select {
							case req := <-handleJoinCh:
								if ses := tc.Device.Session; ses != nil {
									a.So(req.req.EndDeviceIdentifiers.DevAddr, should.NotResemble, ses.DevAddr)
								}

								expectedRequest.EndDeviceIdentifiers.DevAddr = req.req.EndDeviceIdentifiers.DevAddr
								a.So(req.req, should.Resemble, expectedRequest)

								req.ch <- resp
								req.errch <- nil

							case <-time.After(Timeout):
								t.Fatal("Timeout")
							}
						}()

						var errch <-chan error
						md, errch = sendUplinkDuplicates(t, ns, deduplicationDoneCh, ctx, tc.UplinkMessage, DuplicateCount, false)

						select {
						case err := <-errch:
							a.So(err, should.BeNil)

						case <-time.After(Timeout):
							t.Fatal("Timeout")
						}

						select {
						case up := <-asSendCh:
							expected := &ttnpb.ApplicationUp{
								EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
									DevAddr:                expectedRequest.EndDeviceIdentifiers.DevAddr,
									DevEUI:                 tc.Device.EndDeviceIdentifiers.DevEUI,
									DeviceID:               tc.Device.EndDeviceIdentifiers.DeviceID,
									JoinEUI:                tc.Device.EndDeviceIdentifiers.JoinEUI,
									ApplicationIdentifiers: tc.Device.EndDeviceIdentifiers.ApplicationIdentifiers,
								},
								SessionKeyID: test.Must(dev.Load()).(*deviceregistry.Device).Session.SessionKeys.SessionKeyID,
								Up: &ttnpb.ApplicationUp_JoinAccept{JoinAccept: &ttnpb.ApplicationJoinAccept{
									AppSKey: resp.SessionKeys.AppSKey,
								}},
							}
							expected.DevAddr = expectedRequest.EndDeviceIdentifiers.DevAddr
							a.So(up, should.Resemble, expected)

						case <-time.After(Timeout):
							t.Fatal("Timeout")
						}

						wg.Wait()
					})

					t.Run("device update", func(t *testing.T) {
						a := assertions.New(t)

						dev, err = dev.Load()
						if !a.So(err, should.BeNil) ||
							!a.So(dev.EndDevice, should.NotBeNil) {
							t.FailNow()
						}
						if a.So(dev.Session, should.NotBeNil) {
							ses := dev.Session
							a.So(ses.StartedAt, should.HappenBetween, start, time.Now())
							a.So(dev.EndDeviceIdentifiers.DevAddr, should.Resemble, &ses.DevAddr)
							if tc.Device.Session != nil {
								a.So(ses.DevAddr, should.NotResemble, tc.Device.Session.DevAddr)
							}
						}

						if !a.So(dev.RecentUplinks, should.NotBeEmpty) {
							t.FailNow()
						}

						if !a.So(dev.RecentDownlinks, should.NotBeEmpty) {
							t.FailNow()
						}

						a.So(pretty.Diff(dev.RecentDownlinks[len(dev.RecentDownlinks)-1].RawPayload, resp.RawPayload), should.BeEmpty)

						msg := deepcopy.Copy(tc.UplinkMessage).(*ttnpb.UplinkMessage)
						msg.RxMetadata = md

						expected := deepcopy.Copy(tc.Device).(*ttnpb.EndDevice)

						err := ResetMACState(ns.Component.FrequencyPlans, expected)
						if !a.So(err, should.BeNil) {
							t.FailNow()
						}

						expected.MACState.RxDelay = tc.Device.MACStateDesired.RxDelay
						expected.MACState.Rx1DataRateOffset = tc.Device.MACStateDesired.Rx1DataRateOffset
						expected.MACState.Rx2DataRateIndex = tc.Device.MACStateDesired.Rx2DataRateIndex

						expected.MACStateDesired.RxDelay = expected.MACState.RxDelay
						expected.MACStateDesired.Rx1DataRateOffset = expected.MACState.Rx1DataRateOffset
						expected.MACStateDesired.Rx2DataRateIndex = expected.MACState.Rx2DataRateIndex

						expected.EndDeviceIdentifiers.DevAddr = dev.EndDeviceIdentifiers.DevAddr
						expected.Session = &ttnpb.Session{
							SessionKeys: *keys,
							StartedAt:   dev.Session.StartedAt,
							DevAddr:     *dev.EndDeviceIdentifiers.DevAddr,
						}
						expected.SessionFallback = tc.Device.Session
						expected.CreatedAt = dev.CreatedAt
						expected.UpdatedAt = dev.UpdatedAt
						expected.RecentDownlinks = dev.RecentDownlinks

						expected.RecentUplinks = append(expected.RecentUplinks, msg)
						if len(expected.RecentUplinks) > RecentUplinkCount {
							expected.RecentUplinks = expected.RecentUplinks[len(expected.RecentUplinks)-RecentUplinkCount:]
						}

						storedUp := dev.RecentUplinks[len(dev.RecentUplinks)-1]
						expectedUp := expected.RecentUplinks[len(expected.RecentUplinks)-1]

						a.So(storedUp.ReceivedAt, should.HappenBetween, start, time.Now())
						expectedUp.ReceivedAt = storedUp.ReceivedAt

						a.So(storedUp.CorrelationIDs, should.NotBeEmpty)
						expectedUp.CorrelationIDs = storedUp.CorrelationIDs

						if !a.So(storedUp.RxMetadata, should.HaveSameElementsDiff, expectedUp.RxMetadata) {
							metadataLdiff(t, storedUp.RxMetadata, expectedUp.RxMetadata)
						}
						expectedUp.RxMetadata = storedUp.RxMetadata

						a.So(pretty.Diff(dev.EndDevice, expected), should.BeEmpty)
					})
				}) {
					t.FailNow()
				}

				t.Run("cooldown window", func(t *testing.T) {
					a := assertions.New(t)

					_, errch := sendUplinkDuplicates(t, ns, collectionDoneCh, ctx, tc.UplinkMessage, DuplicateCount, true)

					select {
					case err := <-errch:
						a.So(err, should.BeNil)

					case <-time.After(Timeout):
						t.Fatal("Timeout")
					}
				})

				time.Sleep(test.Delay) // Ensure the message hash is removed from deduplication table

				t.Run("after cooldown window", func(t *testing.T) {
					a := assertions.New(t)

					wg := &sync.WaitGroup{}
					wg.Add(1)
					go func() {
						defer wg.Done()

						select {
						case req := <-handleJoinCh:
							a.So(req.req.EndDeviceIdentifiers.DevAddr, should.NotResemble, dev.Session.DevAddr)

							expectedRequest.EndDeviceIdentifiers.DevAddr = req.req.EndDeviceIdentifiers.DevAddr
							a.So(req.req, should.Resemble, expectedRequest)

							resp := ttnpb.NewPopulatedJoinResponse(test.Randy, false)
							resp.SessionKeys = *keys

							req.ch <- resp
							req.errch <- nil

						case <-time.After(Timeout):
							t.Error("Timeout")
						}
					}()

					msg := deepcopy.Copy(tc.UplinkMessage).(*ttnpb.UplinkMessage)

					errch := make(chan error, 1)
					go func() {
						_, err = ns.HandleUplink(ctx, deepcopy.Copy(msg).(*ttnpb.UplinkMessage))
						errch <- err
					}()

					select {
					case de := <-deduplicationDoneCh:
						a.So(de.msg.ReceivedAt, should.HappenBetween, start, time.Now())
						msg.ReceivedAt = de.msg.ReceivedAt

						a.So(de.msg.CorrelationIDs, should.NotBeEmpty)
						msg.CorrelationIDs = de.msg.CorrelationIDs

						a.So(de.msg, should.Resemble, msg)
						a.So(de.ctx, ttnshould.HaveParentContext, ctx)

						de.ch <- time.Now()

					case <-time.After(Timeout):
						t.Fatal("Timeout")
					}

					select {
					case err := <-errch:
						a.So(err, should.BeNil)

					case <-time.After(Timeout):
						t.Fatal("Timeout")
					}

					wg.Wait()
				})
			})
		}
	}
}

func HandleRejoinTest(conf *component.Config) func(t *testing.T) {
	return func(t *testing.T) {
		// TODO: Implement https://github.com/TheThingsIndustries/ttn/issues/557
	}
}

func TestHandleUplink(t *testing.T) {
	a := assertions.New(t)

	fpStore, err := test.NewFrequencyPlansStore()
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	defer fpStore.Destroy()

	conf := &component.Config{ServiceBase: config.ServiceBase{
		FrequencyPlans: config.FrequencyPlans{
			StoreDirectory: fpStore.Directory(),
		},
	}}

	reg := deviceregistry.New(store.NewTypedMapStoreClient(mapstore.New()))
	ns := test.Must(New(
		component.MustNew(test.GetLogger(t), conf),
		&Config{
			Registry:            reg,
			JoinServers:         nil,
			DeduplicationWindow: 42,
			CooldownWindow:      42,
		},
	)).(*NetworkServer)
	test.Must(nil, ns.Start())

	msg := ttnpb.NewPopulatedUplinkMessage(test.Randy, false)
	msg.Payload.Payload = nil
	msg.RawPayload = nil
	_, err = ns.HandleUplink(context.Background(), msg)
	a.So(err, should.NotBeNil)

	msg = ttnpb.NewPopulatedUplinkMessage(test.Randy, false)
	msg.Payload.Payload = nil
	msg.RawPayload = []byte{}
	_, err = ns.HandleUplink(context.Background(), msg)
	a.So(err, should.NotBeNil)

	msg = ttnpb.NewPopulatedUplinkMessage(test.Randy, false)
	msg.Payload.Major = 1
	_, err = ns.HandleUplink(context.Background(), msg)
	a.So(err, should.NotBeNil)

	t.Run("Uplink", HandleUplinkTest(conf))
	t.Run("Join", HandleJoinTest(conf))
	t.Run("Rejoin", HandleRejoinTest(conf))
}
