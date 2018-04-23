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
	"math"
	"math/rand"
	"sync"
	"testing"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/kr/pretty"
	"github.com/mohae/deepcopy"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
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
	"google.golang.org/grpc"
)

var (
	FNwkSIntKey   = types.AES128Key{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	SNwkSIntKey   = types.AES128Key{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	DevAddr       = types.DevAddr{0x42, 0x42, 0xff, 0xff}
	DevEUI        = types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	JoinEUI       = types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	ApplicationID = "test"

	DuplicateCount = 6
	DeviceCount    = 100
	Timeout        = (2 << 10) * test.Delay
)

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

	ed := ttnpb.NewPopulatedEndDevice(test.Randy, false)
	ed.QueuedApplicationDownlinks = nil

	dev, err := reg.Create(ed)
	if !a.So(err, should.BeNil) {
		return
	}

	_, err = ns.DownlinkQueueReplace(context.Background(), &ttnpb.DownlinkQueueRequest{})
	a.So(err, should.NotBeNil)

	req := ttnpb.NewPopulatedDownlinkQueueRequest(test.Randy, false)
	req.EndDeviceIdentifiers = ed.EndDeviceIdentifiers

	_, err = ns.DownlinkQueueReplace(context.Background(), req)
	a.So(err, should.BeNil)

	dev.EndDevice, err = dev.Load()
	if !a.So(err, should.BeNil) ||
		!a.So(dev.EndDevice, should.NotBeNil) {
		return
	}

	a.So(pretty.Diff(dev.EndDevice.GetQueuedApplicationDownlinks(), req.GetDownlinks()), should.BeEmpty)

	req = ttnpb.NewPopulatedDownlinkQueueRequest(test.Randy, false)
	for len(req.GetDownlinks()) == 0 {
		req = ttnpb.NewPopulatedDownlinkQueueRequest(test.Randy, false)
	}
	req.EndDeviceIdentifiers = ed.EndDeviceIdentifiers

	_, err = ns.DownlinkQueueReplace(context.Background(), req)
	a.So(err, should.BeNil)

	dev.EndDevice, err = dev.Load()
	if !a.So(err, should.BeNil) ||
		!a.So(dev.EndDevice, should.NotBeNil) {
		return
	}

	a.So(pretty.Diff(dev.EndDevice.GetQueuedApplicationDownlinks(), req.GetDownlinks()), should.BeEmpty)
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

	ed := ttnpb.NewPopulatedEndDevice(test.Randy, false)
	ed.QueuedApplicationDownlinks = nil

	dev, err := reg.Create(ed)
	if !a.So(err, should.BeNil) {
		return
	}

	_, err = ns.DownlinkQueuePush(context.Background(), &ttnpb.DownlinkQueueRequest{})
	a.So(err, should.NotBeNil)

	req := ttnpb.NewPopulatedDownlinkQueueRequest(test.Randy, false)
	for len(req.GetDownlinks()) == 0 {
		req = ttnpb.NewPopulatedDownlinkQueueRequest(test.Randy, false)
	}
	req.EndDeviceIdentifiers = ed.EndDeviceIdentifiers

	downlinks := append(ed.QueuedApplicationDownlinks, req.GetDownlinks()...)

	_, err = ns.DownlinkQueuePush(context.Background(), req)
	a.So(err, should.BeNil)

	dev.EndDevice, err = dev.Load()
	if !a.So(err, should.BeNil) ||
		!a.So(dev.EndDevice, should.NotBeNil) {
		return
	}

	a.So(pretty.Diff(dev.EndDevice.GetQueuedApplicationDownlinks(), downlinks), should.BeEmpty)

	req = ttnpb.NewPopulatedDownlinkQueueRequest(test.Randy, false)
	req.EndDeviceIdentifiers = ed.EndDeviceIdentifiers
	downlinks = append(downlinks, req.GetDownlinks()...)

	_, err = ns.DownlinkQueuePush(context.Background(), req)
	a.So(err, should.BeNil)

	dev.EndDevice, err = dev.Load()
	if !a.So(err, should.BeNil) ||
		!a.So(dev.EndDevice, should.NotBeNil) {
		return
	}
	a.So(pretty.Diff(dev.EndDevice.GetQueuedApplicationDownlinks(), downlinks), should.BeEmpty)
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

	ed := ttnpb.NewPopulatedEndDevice(test.Randy, false)
	ed.QueuedApplicationDownlinks = nil

	dev, err := reg.Create(ed)
	if !a.So(err, should.BeNil) {
		return
	}

	_, err = ns.DownlinkQueueList(context.Background(), &ttnpb.EndDeviceIdentifiers{})
	a.So(err, should.NotBeNil)

	downlinks, err := ns.DownlinkQueueList(context.Background(), &dev.EndDevice.EndDeviceIdentifiers)
	a.So(err, should.BeNil)
	a.So(pretty.Diff(downlinks, &ttnpb.ApplicationDownlinks{ed.QueuedApplicationDownlinks}), should.BeEmpty)

	ed = ttnpb.NewPopulatedEndDevice(test.Randy, false)
	for len(ed.QueuedApplicationDownlinks) == 0 {
		ed = ttnpb.NewPopulatedEndDevice(test.Randy, false)
	}
	ed.EndDeviceIdentifiers = dev.EndDevice.EndDeviceIdentifiers
	dev.EndDevice = ed

	err = dev.Store()
	if !a.So(err, should.BeNil) {
		return
	}

	downlinks, err = ns.DownlinkQueueList(context.Background(), &dev.EndDevice.EndDeviceIdentifiers)
	a.So(err, should.BeNil)
	a.So(pretty.Diff(downlinks, &ttnpb.ApplicationDownlinks{ed.QueuedApplicationDownlinks}), should.BeEmpty)
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

	ed := ttnpb.NewPopulatedEndDevice(test.Randy, false)
	ed.QueuedApplicationDownlinks = nil

	dev, err := reg.Create(ed)
	if !a.So(err, should.BeNil) {
		return
	}

	e, err := ns.DownlinkQueueClear(context.Background(), &ttnpb.EndDeviceIdentifiers{})
	a.So(err, should.NotBeNil)
	a.So(e, should.BeNil)

	e, err = ns.DownlinkQueueClear(context.Background(), &dev.EndDevice.EndDeviceIdentifiers)
	a.So(err, should.BeNil)
	a.So(e, should.NotBeNil)

	dev.EndDevice, err = dev.Load()
	if !a.So(err, should.BeNil) ||
		!a.So(dev.EndDevice, should.NotBeNil) {
		return
	}
	a.So(dev.EndDevice.GetQueuedApplicationDownlinks(), should.BeEmpty)

	ed = ttnpb.NewPopulatedEndDevice(test.Randy, false)
	for len(ed.QueuedApplicationDownlinks) == 0 {
		ed = ttnpb.NewPopulatedEndDevice(test.Randy, false)
	}
	ed.EndDeviceIdentifiers = dev.EndDevice.EndDeviceIdentifiers
	dev.EndDevice = ed

	err = dev.Store()
	if !a.So(err, should.BeNil) {
		return
	}

	e, err = ns.DownlinkQueueClear(context.Background(), &dev.EndDevice.EndDeviceIdentifiers)
	a.So(err, should.BeNil)
	a.So(e, should.NotBeNil)

	dev.EndDevice, err = dev.Load()
	if !a.So(err, should.BeNil) ||
		!a.So(dev.EndDevice, should.NotBeNil) {
		return
	}
	a.So(dev.EndDevice.GetQueuedApplicationDownlinks(), should.BeEmpty)
}

type UplinkHandler interface {
	HandleUplink(context.Context, *ttnpb.UplinkMessage) (*pbtypes.Empty, error)
}

func sendUplinkDuplicates(t *testing.T, h UplinkHandler, windowEndCh chan windowEnd, ctx context.Context, msg *ttnpb.UplinkMessage, n int, waitForFirst bool) ([]*ttnpb.RxMetadata, <-chan error) {
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
		a.So(we.msg.GetReceivedAt(), should.HappenBefore, time.Now())
		msg.ReceivedAt = we.msg.GetReceivedAt()
		a.So(we.msg, should.Resemble, msg)
		a.So(we.ctx, should.Resemble, ctx)
		weCh = we.ch

	case <-time.After(Timeout):
		t.Fatal("Timeout")
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
				t.Fatal("Timeout")
			}
		}

		select {
		case weCh <- time.Now():

		case <-time.After(Timeout):
			t.Fatal("Timeout")
		}

		close(mdCh)
	}()

	mds := append([]*ttnpb.RxMetadata{}, msg.GetRxMetadata()...)
	for md := range mdCh {
		mds = append(mds, md)
	}
	return mds, errch
}

type mockAsNsLinkApplicationStream struct {
	*test.MockServerStream
	send func(*ttnpb.ApplicationUp) error
}

func (s *mockAsNsLinkApplicationStream) Send(msg *ttnpb.ApplicationUp) error {
	return s.send(msg)
}

func TestLinkApplication(t *testing.T) {
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

	id := ttnpb.NewPopulatedApplicationIdentifiers(test.Randy, false)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	sendFunc := func(*ttnpb.ApplicationUp) error {
		t.Error("Send should not be called")
		return nil
	}

	time.AfterFunc(test.Delay, func() {
		err := ns.LinkApplication(id, &mockAsNsLinkApplicationStream{
			MockServerStream: &test.MockServerStream{
				MockStream: &test.MockStream{
					ContextFunc: func() context.Context {
						ctx, cancel := context.WithCancel(context.Background())
						time.AfterFunc(test.Delay, cancel)
						return ctx
					},
				},
			},
			send: sendFunc,
		})
		a.So(err, should.Resemble, context.Canceled)
		wg.Done()
	})

	err := ns.LinkApplication(id, &mockAsNsLinkApplicationStream{
		MockServerStream: &test.MockServerStream{
			MockStream: &test.MockStream{
				ContextFunc: context.Background,
			},
		},
		send: sendFunc,
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
			return
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

					mac := msg.Payload.GetMACPayload()
					mac.DevAddr = DevAddr
					mac.FCnt = 0x42

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

					mac := msg.Payload.GetMACPayload()
					mac.DevAddr = DevAddr
					mac.FCnt = 0x42

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

					mac := msg.Payload.GetMACPayload()
					mac.DevAddr = DevAddr
					mac.FCnt = 0x42

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

					mac := msg.Payload.GetMACPayload()
					mac.DevAddr = DevAddr
					mac.FCnt = 0x42

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
						},
					},
				},
				0x43,
				func() *ttnpb.UplinkMessage {
					msg := ttnpb.NewPopulatedUplinkMessageUplink(test.Randy, SNwkSIntKey, FNwkSIntKey, false)

					mac := msg.Payload.GetMACPayload()
					mac.DevAddr = DevAddr
					mac.FCnt = 0x42

					msg.Payload.MIC = nil
					mic := test.Must(crypto.ComputeUplinkMIC(SNwkSIntKey, FNwkSIntKey, 0,
						uint8(msg.Settings.GetDataRateIndex()), uint8(msg.Settings.GetChannelIndex()),
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

					mac := msg.Payload.GetMACPayload()
					mac.DevAddr = DevAddr
					mac.FCnt = 0x42

					msg.Payload.MIC = nil
					mic := test.Must(crypto.ComputeUplinkMIC(SNwkSIntKey, FNwkSIntKey, 0x24,
						uint8(msg.Settings.GetDataRateIndex()), uint8(msg.Settings.GetChannelIndex()),
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
						},
					},
					FCntResets: true,
				},
				0x43,
				func() *ttnpb.UplinkMessage {
					msg := ttnpb.NewPopulatedUplinkMessageUplink(test.Randy, SNwkSIntKey, FNwkSIntKey, false)

					mac := msg.Payload.GetMACPayload()
					mac.DevAddr = DevAddr
					mac.FCnt = 0x42

					msg.Payload.MIC = nil
					mic := test.Must(crypto.ComputeUplinkMIC(SNwkSIntKey, FNwkSIntKey, 0,
						uint8(msg.Settings.GetDataRateIndex()), uint8(msg.Settings.GetChannelIndex()),
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

					mac := msg.Payload.GetMACPayload()
					mac.DevAddr = DevAddr
					mac.FCnt = 0x42

					msg.Payload.MIC = nil
					mic := test.Must(crypto.ComputeUplinkMIC(SNwkSIntKey, FNwkSIntKey, 0x24,
						uint8(msg.Settings.GetDataRateIndex()), uint8(msg.Settings.GetChannelIndex()),
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
						return
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

				asSendCh := make(chan *ttnpb.ApplicationUp)

				go func() {
					id := ttnpb.NewPopulatedApplicationIdentifiers(test.Randy, false)
					id.ApplicationID = ApplicationID

					err := ns.LinkApplication(id, &mockAsNsLinkApplicationStream{
						MockServerStream: &test.MockServerStream{
							MockStream: &test.MockStream{
								ContextFunc: context.Background,
							},
						},
						send: func(up *ttnpb.ApplicationUp) error {
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
					return
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
							if !a.So(test.SameElementsDeep(md, up.GetUplinkMessage().GetRxMetadata()), should.BeTrue) {
								metadataLdiff(t, up.GetUplinkMessage().GetRxMetadata(), md)
							}

							expected := &ttnpb.ApplicationUp{
								EndDeviceIdentifiers: dev.EndDeviceIdentifiers,
								SessionKeyID:         dev.GetSession().SessionKeys.GetSessionKeyID(),
								Up: &ttnpb.ApplicationUp_UplinkMessage{&ttnpb.ApplicationUplink{
									FCnt:       tc.NextNextFCntUp - 1,
									FPort:      tc.UplinkMessage.Payload.GetMACPayload().GetFPort(),
									FRMPayload: tc.UplinkMessage.Payload.GetMACPayload().GetFRMPayload(),
									RxMetadata: up.GetUplinkMessage().GetRxMetadata(),
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

						msg := deepcopy.Copy(tc.UplinkMessage).(*ttnpb.UplinkMessage)
						msg.RxMetadata = md

						ed := deepcopy.Copy(tc.Device).(*ttnpb.EndDevice)
						ed.GetSession().NextFCntUp = tc.NextNextFCntUp

						ed.RecentUplinks = append(ed.GetRecentUplinks(), msg)
						if len(ed.RecentUplinks) > RecentUplinkCount {
							ed.RecentUplinks = ed.RecentUplinks[len(ed.RecentUplinks)-RecentUplinkCount:]
						}

						dev.EndDevice, err = dev.Load()
						if !a.So(err, should.BeNil) ||
							!a.So(dev.EndDevice, should.NotBeNil) {
							return
						}

						if !a.So(dev.EndDevice.GetRecentUplinks(), should.NotBeEmpty) {
							return
						}

						storedUp := dev.EndDevice.GetRecentUplinks()[len(dev.EndDevice.RecentUplinks)-1]
						expectedUp := ed.GetRecentUplinks()[len(ed.RecentUplinks)-1]

						a.So(storedUp.GetReceivedAt(), should.HappenBetween, start, time.Now())
						expectedUp.ReceivedAt = storedUp.GetReceivedAt()

						storedMD := storedUp.GetRxMetadata()
						expectedMD := expectedUp.GetRxMetadata()

						if !a.So(test.SameElementsDiff(storedMD, expectedMD), should.BeTrue) {
							metadataLdiff(t, storedMD, expectedMD)
						}

						copy(expectedMD, storedMD)

						ed.CreatedAt = dev.GetCreatedAt()
						ed.UpdatedAt = dev.GetUpdatedAt()
						a.So(pretty.Diff(dev.EndDevice, ed), should.BeEmpty)
					})
				}) {
					return
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
					if len(msg.GetRxMetadata()) > 1 {
						// Deduplication may change the order of metadata in the slice
						msg.RxMetadata = []*ttnpb.RxMetadata{msg.GetRxMetadata()[0]}
					}

					errch := make(chan error, 1)
					go func(msg *ttnpb.UplinkMessage) {
						_, err = ns.HandleUplink(ctx, msg)
						errch <- err
					}(deepcopy.Copy(msg).(*ttnpb.UplinkMessage))

					select {
					case err := <-errch:
						if !dev.GetFCntResets() {
							a.So(err, should.BeError)
							return
						}

						a.So(err, should.BeNil)
						t.Fatal("Uplink not sent to AS")

					case de := <-deduplicationDoneCh:
						a.So(de.msg.GetReceivedAt(), should.HappenBetween, start, time.Now())
						msg.ReceivedAt = de.msg.GetReceivedAt()
						a.So(de.msg, should.Resemble, msg)
						a.So(de.ctx, should.Resemble, ctx)
						de.ch <- time.Now()

					case <-time.After(Timeout):
						t.Fatal("Timeout")
					}

					select {
					case up := <-asSendCh:
						a.So(up, should.Resemble, &ttnpb.ApplicationUp{
							EndDeviceIdentifiers: dev.EndDeviceIdentifiers,
							SessionKeyID:         dev.GetSession().SessionKeys.GetSessionKeyID(),
							Up: &ttnpb.ApplicationUp_UplinkMessage{&ttnpb.ApplicationUplink{
								FCnt:       tc.NextNextFCntUp - 1,
								FPort:      msg.Payload.GetMACPayload().GetFPort(),
								FRMPayload: msg.Payload.GetMACPayload().GetFRMPayload(),
								RxMetadata: msg.GetRxMetadata(),
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

var _ ttnpb.NsJsClient = &mockNsJsClient{}

type mockNsJsClient struct {
	handleJoin  func(ctx context.Context, req *ttnpb.JoinRequest, opts ...grpc.CallOption) (*ttnpb.JoinResponse, error)
	getNwkSKeys func(ctx context.Context, req *ttnpb.SessionKeyRequest, opts ...grpc.CallOption) (*ttnpb.NwkSKeysResponse, error)
}

func (c *mockNsJsClient) HandleJoin(ctx context.Context, req *ttnpb.JoinRequest, opts ...grpc.CallOption) (*ttnpb.JoinResponse, error) {
	return c.handleJoin(ctx, req, opts...)
}

func (c *mockNsJsClient) GetNwkSKeys(ctx context.Context, req *ttnpb.SessionKeyRequest, opts ...grpc.CallOption) (*ttnpb.NwkSKeysResponse, error) {
	return c.getNwkSKeys(ctx, req, opts...)
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

		_, err := ns.HandleUplink(context.Background(), ttnpb.NewPopulatedUplinkMessageJoinRequest(test.Randy))
		a.So(err, should.NotBeNil)

		req := ttnpb.NewPopulatedUplinkMessageJoinRequest(test.Randy)
		ed := ttnpb.NewPopulatedEndDevice(test.Randy, false)

		ed.EndDeviceIdentifiers.DevEUI = &req.Payload.GetJoinRequestPayload().DevEUI
		ed.EndDeviceIdentifiers.JoinEUI = &req.Payload.GetJoinRequestPayload().JoinEUI

		_, err = reg.Create(ed)
		if !a.So(err, should.BeNil) {
			return
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
						return
					}
				}

				keys := ttnpb.NewPopulatedSessionKeys(test.Randy, false)

				getNwkSKeys := func(ctx context.Context, req *ttnpb.SessionKeyRequest, opts ...grpc.CallOption) (*ttnpb.NwkSKeysResponse, error) {
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
						JoinServers: []ttnpb.NsJsClient{&mockNsJsClient{
							handleJoin: func(ctx context.Context, req *ttnpb.JoinRequest, opts ...grpc.CallOption) (*ttnpb.JoinResponse, error) {
								return nil, errors.New("test")
							},
							getNwkSKeys: getNwkSKeys,
						},
							&mockNsJsClient{
								handleJoin: func(ctx context.Context, req *ttnpb.JoinRequest, opts ...grpc.CallOption) (*ttnpb.JoinResponse, error) {
									ch := make(chan *ttnpb.JoinResponse, 1)
									errch := make(chan error, 1)
									handleJoinCh <- handleJoinRequest{ctx, req, opts, ch, errch}
									return <-ch, <-errch
								},
								getNwkSKeys: getNwkSKeys,
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
				)).(*NetworkServer)

				asSendCh := make(chan *ttnpb.ApplicationUp)

				go func() {
					id := ttnpb.NewPopulatedApplicationIdentifiers(test.Randy, false)
					id.ApplicationID = ApplicationID

					err := ns.LinkApplication(id, &mockAsNsLinkApplicationStream{
						MockServerStream: &test.MockServerStream{
							MockStream: &test.MockStream{
								ContextFunc: context.Background,
							},
						},
						send: func(up *ttnpb.ApplicationUp) error {
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
					return
				}

				expectedRequest := &ttnpb.JoinRequest{
					RawPayload: tc.UplinkMessage.GetRawPayload(),
					Payload:    tc.UplinkMessage.GetPayload(),
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						DevEUI:  &DevEUI,
						JoinEUI: &JoinEUI,
					},
					NetID:              ns.NetID,
					SelectedMacVersion: tc.Device.GetLoRaWANVersion(),
					RxDelay:            tc.Device.GetMACStateDesired().GetRxDelay(),
					CFList:             nil,
					DownlinkSettings: ttnpb.DLSettings{
						Rx1DROffset: tc.Device.GetMACStateDesired().GetRx1DataRateOffset(),
						Rx2DR:       tc.Device.GetMACStateDesired().GetRx2DataRateIndex(),
					},
				}

				ctx := context.WithValue(context.Background(), "answer", 42)

				start := time.Now()
				if !t.Run("deduplication window", func(t *testing.T) {
					var md []*ttnpb.RxMetadata

					t.Run("message send", func(t *testing.T) {
						a := assertions.New(t)

						resp := ttnpb.NewPopulatedJoinResponse(test.Randy, false)
						resp.SessionKeys = *keys

						wg := &sync.WaitGroup{}
						wg.Add(1)
						go func() {
							defer wg.Done()

							select {
							case req := <-handleJoinCh:
								if ses := tc.Device.GetSession(); ses != nil {
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
									DeviceID:               tc.Device.EndDeviceIdentifiers.GetDeviceID(),
									JoinEUI:                tc.Device.EndDeviceIdentifiers.JoinEUI,
									ApplicationIdentifiers: tc.Device.EndDeviceIdentifiers.ApplicationIdentifiers,
								},
								SessionKeyID: test.Must(dev.Load()).(*ttnpb.EndDevice).GetSession().SessionKeys.GetSessionKeyID(),
								Up: &ttnpb.ApplicationUp_JoinAccept{&ttnpb.ApplicationJoinAccept{
									AppSKey: resp.SessionKeys.GetAppSKey(),
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

						msg := deepcopy.Copy(tc.UplinkMessage).(*ttnpb.UplinkMessage)
						msg.RxMetadata = md

						ed := deepcopy.Copy(tc.Device).(*ttnpb.EndDevice)

						ed.RecentUplinks = append(ed.GetRecentUplinks(), msg)
						if len(ed.RecentUplinks) > RecentUplinkCount {
							ed.RecentUplinks = ed.RecentUplinks[len(ed.RecentUplinks)-RecentUplinkCount:]
						}

						dev.EndDevice, err = dev.Load()
						if !a.So(err, should.BeNil) ||
							!a.So(dev.EndDevice, should.NotBeNil) {
							return
						}

						if !a.So(dev.EndDevice.GetRecentUplinks(), should.NotBeEmpty) {
							return
						}

						storedUp := dev.EndDevice.GetRecentUplinks()[len(dev.EndDevice.RecentUplinks)-1]
						expectedUp := ed.GetRecentUplinks()[len(ed.RecentUplinks)-1]

						a.So(storedUp.GetReceivedAt(), should.HappenBetween, start, time.Now())
						a.So(dev.EndDevice.Session.StartedAt, should.HappenBetween, start, time.Now())
						expectedUp.ReceivedAt = storedUp.GetReceivedAt()

						storedMD := storedUp.GetRxMetadata()
						expectedMD := expectedUp.GetRxMetadata()

						if !a.So(test.SameElementsDiff(storedMD, expectedMD), should.BeTrue) {
							metadataLdiff(t, storedMD, expectedMD)
						}

						storedUp.RxMetadata = expectedUp.RxMetadata

						a.So(dev.EndDevice.SessionFallback, should.BeNil)
						if a.So(dev.EndDevice.GetSession(), should.NotBeNil) {
							a.So(dev.EndDevice.Session.SessionKeys, should.Resemble, *keys)
							a.So(dev.EndDevice.Session.StartedAt, should.HappenBetween, start, time.Now())
							a.So(dev.EndDevice.EndDeviceIdentifiers.DevAddr, should.Resemble, &dev.EndDevice.Session.DevAddr)
							if ed.Session != nil {
								a.So(dev.EndDevice.Session.DevAddr, should.NotResemble, ed.Session.DevAddr)
							}
						}

						ed.EndDeviceIdentifiers.DevAddr = dev.EndDevice.EndDeviceIdentifiers.DevAddr
						ed.Session = dev.EndDevice.GetSession()
						ed.CreatedAt = dev.GetCreatedAt()
						ed.UpdatedAt = dev.GetUpdatedAt()
						a.So(pretty.Diff(dev.EndDevice, ed), should.BeEmpty)
					})
				}) {
					return
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
							a.So(req.req.EndDeviceIdentifiers.DevAddr, should.NotResemble, dev.EndDevice.GetSession().DevAddr)

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
						a.So(de.msg.GetReceivedAt(), should.HappenBetween, start, time.Now())
						msg.ReceivedAt = de.msg.GetReceivedAt()
						a.So(de.msg, should.Resemble, msg)
						a.So(de.ctx, should.Resemble, ctx)
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

	msg := ttnpb.NewPopulatedUplinkMessage(test.Randy, false)
	msg.Payload.Payload = nil
	msg.RawPayload = nil
	e, err := ns.HandleUplink(context.Background(), msg)
	a.So(err, should.NotBeNil)
	a.So(e, should.BeNil)

	msg = ttnpb.NewPopulatedUplinkMessage(test.Randy, false)
	msg.Payload.Payload = nil
	msg.RawPayload = []byte{}
	e, err = ns.HandleUplink(context.Background(), msg)
	a.So(err, should.NotBeNil)
	a.So(e, should.BeNil)

	msg = ttnpb.NewPopulatedUplinkMessage(test.Randy, false)
	msg.Payload.Major = 1
	e, err = ns.HandleUplink(context.Background(), msg)
	a.So(err, should.NotBeNil)
	a.So(e, should.BeNil)

	t.Run("Uplink", HandleUplinkTest(conf))
	t.Run("Join", HandleJoinTest(conf))
	t.Run("Rejoin", HandleRejoinTest(conf))
}
