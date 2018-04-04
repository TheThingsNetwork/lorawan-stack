// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package networkserver_test

import (
	"context"
	"math"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/component"
	"github.com/TheThingsNetwork/ttn/pkg/crypto"
	"github.com/TheThingsNetwork/ttn/pkg/deviceregistry"
	"github.com/TheThingsNetwork/ttn/pkg/errors"
	. "github.com/TheThingsNetwork/ttn/pkg/networkserver"
	"github.com/TheThingsNetwork/ttn/pkg/store"
	"github.com/TheThingsNetwork/ttn/pkg/store/mapstore"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/TheThingsNetwork/ttn/pkg/types"
	"github.com/TheThingsNetwork/ttn/pkg/util/test"
	"github.com/kr/pretty"
	"github.com/mohae/deepcopy"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"google.golang.org/grpc"
)

var (
	// Slowdown slows down the test in a linear fashion(the higher the value - the slower the test),
	// which may be required on slower systems
	Slowdown = 0

	DuplicateCount = 5

	FNwkSIntKey   = types.AES128Key{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	SNwkSIntKey   = types.AES128Key{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	DevAddr       = types.DevAddr{0x42, 0x42, 0xff, 0xff}
	DevEUI        = types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	JoinEUI       = types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	ApplicationID = "test"

	Delay               = 20 * time.Millisecond * time.Duration(1+Slowdown)
	DeduplicationWindow = 4 * Delay
	CooldownWindow      = 8 * Delay
)

func TestDownlinkQueueReplace(t *testing.T) {
	a := assertions.New(t)
	reg := deviceregistry.New(store.NewTypedStoreClient(mapstore.New()))
	ns := New(
		component.MustNew(test.GetLogger(t), &component.Config{}),
		&Config{
			Registry:    reg,
			JoinServers: nil,
		})

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
	reg := deviceregistry.New(store.NewTypedStoreClient(mapstore.New()))
	ns := New(
		component.MustNew(test.GetLogger(t), &component.Config{}),
		&Config{
			Registry:    reg,
			JoinServers: nil,
		})

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
	reg := deviceregistry.New(store.NewTypedStoreClient(mapstore.New()))
	ns := New(
		component.MustNew(test.GetLogger(t), &component.Config{}),
		&Config{
			Registry:    reg,
			JoinServers: nil,
		})

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
	reg := deviceregistry.New(store.NewTypedStoreClient(mapstore.New()))
	ns := New(
		component.MustNew(test.GetLogger(t), &component.Config{}),
		&Config{
			Registry:    reg,
			JoinServers: nil,
		})

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

type mockAsNsLinkApplicationStream struct {
	*test.MockServerStream
	send func(*ttnpb.ApplicationUp) error
}

func (s *mockAsNsLinkApplicationStream) Send(msg *ttnpb.ApplicationUp) error {
	return s.send(msg)
}

func TestLinkApplication(t *testing.T) {
	a := assertions.New(t)
	reg := deviceregistry.New(store.NewTypedStoreClient(mapstore.New()))
	ns := New(
		component.MustNew(test.GetLogger(t), &component.Config{}),
		&Config{
			Registry:    reg,
			JoinServers: nil,
		})

	id := ttnpb.NewPopulatedApplicationIdentifier(test.Randy, false)

	stream := &mockAsNsLinkApplicationStream{
		MockServerStream: &test.MockServerStream{
			MockStream: &test.MockStream{
				ContextFunc: context.Background,
			},
		},
		send: func(*ttnpb.ApplicationUp) error {
			t.Error("Send should not be called")
			return nil
		},
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	time.AfterFunc(Delay, func() {
		stream.ContextFunc = func() context.Context {
			ctx, cancel := context.WithCancel(context.Background())
			time.AfterFunc(Delay, cancel)
			return ctx
		}

		err := ns.LinkApplication(id, stream)
		a.So(err, should.Resemble, context.Canceled)
		wg.Done()
	})

	err := ns.LinkApplication(id, stream)
	a.So(err, should.NotBeNil)

	wg.Wait()
}

func HandleUplinkTest() func(t *testing.T) {
	return func(t *testing.T) {
		a := assertions.New(t)

		reg := deviceregistry.New(store.NewTypedStoreClient(mapstore.New()))
		ns := New(
			component.MustNew(test.GetLogger(t), &component.Config{}),
			&Config{
				Registry:    reg,
				JoinServers: nil,
			},
		)

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
						ApplicationID: ApplicationID,
					},
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
						ApplicationID: ApplicationID,
					},
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
						ApplicationID: ApplicationID,
					},
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
						ApplicationID: ApplicationID,
					},
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
						ApplicationID: ApplicationID,
					},
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
						ApplicationID: ApplicationID,
					},
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
						ApplicationID: ApplicationID,
					},
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
						ApplicationID: ApplicationID,
					},
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

				reg := deviceregistry.New(store.NewTypedStoreClient(mapstore.New()))
				ns := New(
					component.MustNew(test.GetLogger(t), &component.Config{}),
					&Config{
						Registry:            reg,
						DeduplicationWindow: DeduplicationWindow,
						CooldownWindow:      CooldownWindow,
					},
				)

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

				for i := 0; i < 100; i++ {
					ed := ttnpb.NewPopulatedEndDevice(test.Randy, false)

					if s := ed.Session; s != nil {
						populateSessionKeys(s)

						s.DevAddr = DevAddr
						for ed.SessionFallback != nil && ed.SessionFallback.DevAddr.Equal(DevAddr) {
							ed.SessionFallback.DevAddr = *types.NewPopulatedDevAddr(test.Randy)
						}
					} else if s := ed.SessionFallback; s != nil {
						populateSessionKeys(s)

						s.DevAddr = DevAddr
						for ed.Session != nil && ed.Session.DevAddr.Equal(DevAddr) {
							ed.Session.DevAddr = *types.NewPopulatedDevAddr(test.Randy)
						}
					}
					_, err := reg.Create(ed)
					if !a.So(err, should.BeNil) {
						return
					}
				}

				dev, err := reg.Create(tc.Device)
				if !a.So(err, should.BeNil) {
					return
				}

				wg := &sync.WaitGroup{}
				wg.Add(1)

				mdCh := make(chan []*ttnpb.RxMetadata, 2)

				stream := &mockAsNsLinkApplicationStream{
					MockServerStream: &test.MockServerStream{
						MockStream: &test.MockStream{
							ContextFunc: context.Background,
						},
					},
					send: func(up *ttnpb.ApplicationUp) error {
						t.Run("application uplink", func(t *testing.T) {
							a := assertions.New(t)

							a.So(test.SameElementsDiff(<-mdCh, up.GetUplinkMessage().GetRxMetadata()), should.BeTrue)

							upc := deepcopy.Copy(up).(*ttnpb.ApplicationUp)
							upc.GetUplinkMessage().RxMetadata = nil

							a.So(upc, should.Resemble, &ttnpb.ApplicationUp{&ttnpb.ApplicationUp_UplinkMessage{&ttnpb.ApplicationUplink{
								FCnt:       tc.NextNextFCntUp - 1,
								FPort:      tc.UplinkMessage.Payload.GetMACPayload().GetFPort(),
								FRMPayload: tc.UplinkMessage.Payload.GetMACPayload().GetFRMPayload(),
							}}})
							wg.Done()
						})
						return nil
					},
				}

				id := ttnpb.NewPopulatedApplicationIdentifier(test.Randy, false)
				id.ApplicationID = ApplicationID

				go func() {
					err := ns.LinkApplication(id, stream)
					// LinkApplication should not return
					t.Errorf("LinkApplication should not return, returned with error: %s", err)
				}()

				md := make([]*ttnpb.RxMetadata, 0, DuplicateCount+1)

				sleepInterval := (DeduplicationWindow + CooldownWindow) / time.Duration(DuplicateCount)

				var deduplicationEnd time.Time
				var cooldownEnd time.Time

				if !t.Run("handle uplink", func(t *testing.T) {
					a := assertions.New(t)

					deduplicationEnd = time.Now().Add(DeduplicationWindow)
					cooldownEnd = deduplicationEnd.Add(CooldownWindow)

					var deduplicated uint64 = 1
					for ; time.Now().Before(cooldownEnd); time.Sleep(sleepInterval) {
						go func() {
							_, err = ns.HandleUplink(context.Background(), deepcopy.Copy(tc.UplinkMessage).(*ttnpb.UplinkMessage))
							if a.So(err, should.BeNil) && time.Now().Before(deduplicationEnd) {
								atomic.AddUint64(&deduplicated, 1)
							}
						}()
					}

					for i := uint64(0); i < deduplicated; i++ {
						md = append(md, tc.UplinkMessage.GetRxMetadata()...)
					}
					mdCh <- md
				}) {
					return
				}

				t.Run("updated NextFCntUp and RecentUplinks", func(t *testing.T) {
					a := assertions.New(t)

					msg := deepcopy.Copy(tc.UplinkMessage).(*ttnpb.UplinkMessage)
					msg.RxMetadata = md

					ed := deepcopy.Copy(tc.Device).(*ttnpb.EndDevice)
					ed.GetSession().NextFCntUp = tc.NextNextFCntUp

					ed.RecentUplinks = append(tc.Device.GetRecentUplinks(), msg)
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
					dUp := dev.EndDevice.GetRecentUplinks()[len(dev.EndDevice.RecentUplinks)-1]
					eUp := ed.GetRecentUplinks()[len(ed.RecentUplinks)-1]

					a.So(test.SameElementsDiff(dUp.GetRxMetadata(), eUp.GetRxMetadata()), should.BeTrue)

					dUp.RxMetadata = nil
					eUp.RxMetadata = nil

					a.So(pretty.Diff(dev.EndDevice, ed), should.BeEmpty)
				})

				t.Run("duplicates after cooldown window end", func(t *testing.T) {
					a := assertions.New(t)

					md := make([]*ttnpb.RxMetadata, 0, DuplicateCount+1)

					deduplicationEnd := time.Now().Add(DeduplicationWindow)

					_, err = ns.HandleUplink(context.Background(), deepcopy.Copy(tc.UplinkMessage).(*ttnpb.UplinkMessage))
					if dev.FCntResets {
						// Replay attack is possible if FCnt resets
						wg.Add(1)
						a.So(err, should.BeNil)
					} else {
						a.So(err, should.BeError)
					}

					cooldownEnd := deduplicationEnd.Add(CooldownWindow)

					var deduplicated uint64 = 1
					for ; time.Now().Before(cooldownEnd); time.Sleep(sleepInterval) {
						go func() {
							_, err = ns.HandleUplink(context.Background(), deepcopy.Copy(tc.UplinkMessage).(*ttnpb.UplinkMessage))
							if a.So(err, should.BeNil) && time.Now().Before(deduplicationEnd) {
								atomic.AddUint64(&deduplicated, 1)
							}
						}()
					}

					for i := uint64(0); i < deduplicated; i++ {
						md = append(md, tc.UplinkMessage.GetRxMetadata()...)
					}
					mdCh <- md
				})

				after := time.AfterFunc(Delay+CooldownWindow+DeduplicationWindow, func() {
					t.Error("Second uplink did not get sent")
					wg.Done()
				})

				wg.Wait()
				after.Stop()
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

func HandleJoinTest() func(t *testing.T) {
	return func(t *testing.T) {
		a := assertions.New(t)

		reg := deviceregistry.New(store.NewTypedStoreClient(mapstore.New()))
		ns := New(
			component.MustNew(test.GetLogger(t), &component.Config{}),
			&Config{
				Registry: reg,
			},
		)

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
						DevEUI:  &DevEUI,
						JoinEUI: &JoinEUI,
					},
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
						DevEUI:  &DevEUI,
						JoinEUI: &JoinEUI,
					},
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
						DevEUI:  &DevEUI,
						JoinEUI: &JoinEUI,
					},
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

				getNwkSKeys := func(ctx context.Context, req *ttnpb.SessionKeyRequest, opts ...grpc.CallOption) (*ttnpb.NwkSKeysResponse, error) {
					err := errors.New("GetNwkSKeys should not be called")
					t.Error(err)
					return nil, err
				}

				reqExpected := &ttnpb.JoinRequest{
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

				keys := ttnpb.NewPopulatedSessionKeys(test.Randy, false)

				reg := deviceregistry.New(store.NewTypedStoreClient(mapstore.New()))

				ns := New(
					component.MustNew(test.GetLogger(t), &component.Config{}),
					&Config{
						Registry:            reg,
						DeduplicationWindow: DeduplicationWindow,
						CooldownWindow:      CooldownWindow,
						JoinServers: []ttnpb.NsJsClient{&mockNsJsClient{
							handleJoin: func(ctx context.Context, req *ttnpb.JoinRequest, opts ...grpc.CallOption) (*ttnpb.JoinResponse, error) {
								return nil, errors.New("test")
							},
							getNwkSKeys: getNwkSKeys,
						},
							&mockNsJsClient{
								handleJoin: func(ctx context.Context, req *ttnpb.JoinRequest, opts ...grpc.CallOption) (*ttnpb.JoinResponse, error) {
									a := assertions.New(t)

									resp := ttnpb.NewPopulatedJoinResponse(test.Randy, false)
									resp.SessionKeys = *keys

									if tc.Device.GetSession() != nil {
										a.So(req.EndDeviceIdentifiers.DevAddr, should.NotResemble, tc.Device.Session.DevAddr)
									}
									req.EndDeviceIdentifiers.DevAddr = nil

									a.So(pretty.Diff(req, reqExpected), should.BeEmpty)

									return resp, nil
								},
								getNwkSKeys: getNwkSKeys,
							},
						},
					},
				)

				for i := 0; i < 100; i++ {
					_, err := reg.Create(ttnpb.NewPopulatedEndDevice(test.Randy, false))
					if !a.So(err, should.BeNil) {
						return
					}
				}

				dev, err := reg.Create(tc.Device)
				if !a.So(err, should.BeNil) {
					return
				}

				md := make([]*ttnpb.RxMetadata, 0, DuplicateCount+1)

				sleepInterval := (DeduplicationWindow + CooldownWindow) / time.Duration(DuplicateCount)

				var start time.Time
				var deduplicationEnd time.Time
				var cooldownEnd time.Time

				if !t.Run("handle join", func(t *testing.T) {
					a := assertions.New(t)

					start := time.Now()
					deduplicationEnd = start.Add(DeduplicationWindow)
					cooldownEnd = deduplicationEnd.Add(CooldownWindow)

					var deduplicated uint64 = 1
					for ; time.Now().Before(cooldownEnd); time.Sleep(sleepInterval) {
						go func() {
							_, err = ns.HandleUplink(context.Background(), deepcopy.Copy(tc.UplinkMessage).(*ttnpb.UplinkMessage))
							if a.So(err, should.BeNil) && time.Now().Before(deduplicationEnd) {
								atomic.AddUint64(&deduplicated, 1)
							}
						}()
					}

					for i := uint64(0); i < deduplicated; i++ {
						md = append(md, tc.UplinkMessage.GetRxMetadata()...)
					}
				}) {
					return
				}

				t.Run("updated Session and RecentUplinks", func(t *testing.T) {
					a := assertions.New(t)

					msg := deepcopy.Copy(tc.UplinkMessage).(*ttnpb.UplinkMessage)
					msg.RxMetadata = md

					ed := deepcopy.Copy(tc.Device).(*ttnpb.EndDevice)

					ed.RecentUplinks = append(tc.Device.GetRecentUplinks(), msg)
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
					dUp := dev.EndDevice.GetRecentUplinks()[len(dev.EndDevice.RecentUplinks)-1]
					eUp := ed.GetRecentUplinks()[len(ed.RecentUplinks)-1]

					a.So(test.SameElementsDiff(dUp.GetRxMetadata(), eUp.GetRxMetadata()), should.BeTrue)

					dUp.RxMetadata = nil
					eUp.RxMetadata = nil

					a.So(dev.EndDevice.SessionFallback, should.BeNil)
					if a.So(dev.EndDevice.GetSession(), should.NotBeNil) {
						a.So(dev.EndDevice.Session.SessionKeys, should.Resemble, *keys)
						a.So([]time.Time{start, dev.EndDevice.Session.StartedAt, time.Now()}, should.BeChronological)
						if ed.Session != nil {
							a.So(dev.EndDevice.Session.DevAddr, should.NotResemble, ed.Session.DevAddr)
						}
					}

					ed.Session = nil
					dev.EndDevice.Session = nil

					a.So(pretty.Diff(dev.EndDevice, ed), should.BeEmpty)
				})

				t.Run("duplicates after cooldown window end", func(t *testing.T) {
					a := assertions.New(t)

					deduplicationEnd := time.Now().Add(DeduplicationWindow)

					_, err = ns.HandleUplink(context.Background(), deepcopy.Copy(tc.UplinkMessage).(*ttnpb.UplinkMessage))
					a.So(err, should.BeNil)

					cooldownEnd := deduplicationEnd.Add(CooldownWindow)

					var deduplicated uint64 = 1
					for ; time.Now().Before(cooldownEnd); time.Sleep(sleepInterval) {
						go func() {
							_, err = ns.HandleUplink(context.Background(), deepcopy.Copy(tc.UplinkMessage).(*ttnpb.UplinkMessage))
							if a.So(err, should.BeNil) && time.Now().Before(deduplicationEnd) {
								atomic.AddUint64(&deduplicated, 1)
							}
						}()
					}
				})
			})
		}
	}
}

func HandleRejoinTest() func(t *testing.T) {
	return func(t *testing.T) {
		// TODO: Implement https://github.com/TheThingsIndustries/ttn/issues/557
	}
}

func TestHandleUplink(t *testing.T) {
	a := assertions.New(t)

	reg := deviceregistry.New(store.NewTypedStoreClient(mapstore.New()))
	ns := New(
		component.MustNew(test.GetLogger(t), &component.Config{}),
		&Config{
			Registry:    reg,
			JoinServers: nil,
		},
	)

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

	t.Run("Uplink", HandleUplinkTest())
	t.Run("Join", HandleJoinTest())
	t.Run("Rejoin", HandleRejoinTest())
}
