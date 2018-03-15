// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package networkserver_test

import (
	"context"
	"math"
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/component"
	"github.com/TheThingsNetwork/ttn/pkg/deviceregistry"
	. "github.com/TheThingsNetwork/ttn/pkg/networkserver"
	"github.com/TheThingsNetwork/ttn/pkg/store"
	"github.com/TheThingsNetwork/ttn/pkg/store/mapstore"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/TheThingsNetwork/ttn/pkg/types"
	"github.com/TheThingsNetwork/ttn/pkg/util/test"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
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

	_, err := reg.Create(ed)
	if !a.So(err, should.BeNil) {
		return
	}

	_, err = ns.DownlinkQueueReplace(context.Background(), &ttnpb.DownlinkQueueRequest{})
	a.So(err, should.NotBeNil)

	req := ttnpb.NewPopulatedDownlinkQueueRequest(test.Randy, false)
	req.EndDeviceIdentifiers = ed.EndDeviceIdentifiers

	_, err = ns.DownlinkQueueReplace(context.Background(), req)
	a.So(err, should.BeNil)

	dev, err := deviceregistry.FindOneDeviceByIdentifiers(reg, &ed.EndDeviceIdentifiers)
	if !a.So(err, should.BeNil) ||
		!a.So(dev, should.NotBeNil) ||
		!a.So(dev.EndDevice, should.NotBeNil) {
		return
	}
	a.So(dev.EndDevice.GetQueuedApplicationDownlinks(), should.Resemble, req.GetDownlinks())

	req = ttnpb.NewPopulatedDownlinkQueueRequest(test.Randy, false)
	for len(req.GetDownlinks()) == 0 {
		req = ttnpb.NewPopulatedDownlinkQueueRequest(test.Randy, false)
	}
	req.EndDeviceIdentifiers = ed.EndDeviceIdentifiers

	_, err = ns.DownlinkQueueReplace(context.Background(), req)
	a.So(err, should.BeNil)

	dev, err = deviceregistry.FindOneDeviceByIdentifiers(reg, &ed.EndDeviceIdentifiers)
	if !a.So(err, should.BeNil) ||
		!a.So(dev, should.NotBeNil) ||
		!a.So(dev.EndDevice, should.NotBeNil) {
		return
	}
	a.So(dev.EndDevice.GetQueuedApplicationDownlinks(), should.Resemble, req.GetDownlinks())
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

	_, err := reg.Create(ed)
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

	dev, err := deviceregistry.FindOneDeviceByIdentifiers(reg, &ed.EndDeviceIdentifiers)
	if !a.So(err, should.BeNil) ||
		!a.So(dev, should.NotBeNil) ||
		!a.So(dev.EndDevice, should.NotBeNil) {
		return
	}
	a.So(dev.EndDevice.GetQueuedApplicationDownlinks(), should.Resemble, downlinks)

	req = ttnpb.NewPopulatedDownlinkQueueRequest(test.Randy, false)
	req.EndDeviceIdentifiers = ed.EndDeviceIdentifiers
	downlinks = append(downlinks, req.GetDownlinks()...)

	_, err = ns.DownlinkQueuePush(context.Background(), req)
	a.So(err, should.BeNil)

	dev, err = deviceregistry.FindOneDeviceByIdentifiers(reg, &ed.EndDeviceIdentifiers)
	if !a.So(err, should.BeNil) ||
		!a.So(dev, should.NotBeNil) ||
		!a.So(dev.EndDevice, should.NotBeNil) {
		return
	}
	a.So(dev.EndDevice.GetQueuedApplicationDownlinks(), should.Resemble, downlinks)
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
	a.So(downlinks, should.Resemble, &ttnpb.ApplicationDownlinks{ed.QueuedApplicationDownlinks})

	ed = ttnpb.NewPopulatedEndDevice(test.Randy, false)
	for len(ed.QueuedApplicationDownlinks) == 0 {
		ed = ttnpb.NewPopulatedEndDevice(test.Randy, false)
	}
	ed.EndDeviceIdentifiers = dev.EndDevice.EndDeviceIdentifiers
	dev.EndDevice = ed

	err = dev.Update()
	if !a.So(err, should.BeNil) {
		return
	}

	downlinks, err = ns.DownlinkQueueList(context.Background(), &dev.EndDevice.EndDeviceIdentifiers)
	a.So(err, should.BeNil)
	a.So(downlinks, should.Resemble, &ttnpb.ApplicationDownlinks{ed.QueuedApplicationDownlinks})
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

	dev, err = deviceregistry.FindOneDeviceByIdentifiers(reg, &ed.EndDeviceIdentifiers)
	if !a.So(err, should.BeNil) ||
		!a.So(dev, should.NotBeNil) ||
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

	err = dev.Update()
	if !a.So(err, should.BeNil) {
		return
	}

	e, err = ns.DownlinkQueueClear(context.Background(), &dev.EndDevice.EndDeviceIdentifiers)
	a.So(err, should.BeNil)
	a.So(e, should.NotBeNil)

	dev, err = deviceregistry.FindOneDeviceByIdentifiers(reg, &ed.EndDeviceIdentifiers)
	if !a.So(err, should.BeNil) ||
		!a.So(dev, should.NotBeNil) ||
		!a.So(dev.EndDevice, should.NotBeNil) {
		return
	}
	a.So(dev.EndDevice.GetQueuedApplicationDownlinks(), should.BeEmpty)
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

	msg = ttnpb.NewPopulatedUplinkMessage(test.Randy, false)
	msg.Payload = *ttnpb.NewPopulatedMessageDownlink(test.Randy, *types.NewPopulatedAES128Key(test.Randy), false)
	e, err = ns.HandleUplink(context.Background(), msg)
	a.So(err, should.NotBeNil)
	a.So(e, should.BeNil)

	t.Run("Uplink", handleUplinkTest())
	// TODO: Test Join/Rejoin
}

func handleUplinkTest() func(t *testing.T) {
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

		msg := ttnpb.NewPopulatedUplinkMessageUplink(test.Randy, *types.NewPopulatedAES128Key(test.Randy), *types.NewPopulatedAES128Key(test.Randy), false)
		msg.EndDeviceIdentifiers = ttnpb.EndDeviceIdentifiers{}
		_, err := ns.HandleUplink(context.Background(), msg)
		a.So(err, should.NotBeNil)

		msg = ttnpb.NewPopulatedUplinkMessageUplink(test.Randy, *types.NewPopulatedAES128Key(test.Randy), *types.NewPopulatedAES128Key(test.Randy), false)
		msg.Payload.GetMACPayload().DevAddr = types.DevAddr{}
		_, err = ns.HandleUplink(context.Background(), msg)
		a.So(err, should.NotBeNil)

		msg = ttnpb.NewPopulatedUplinkMessageUplink(test.Randy, *types.NewPopulatedAES128Key(test.Randy), *types.NewPopulatedAES128Key(test.Randy), false)
		msg.Payload.GetMACPayload().FCnt = math.MaxUint16 + 1
		_, err = ns.HandleUplink(context.Background(), msg)
		a.So(err, should.NotBeNil)

		msg = ttnpb.NewPopulatedUplinkMessageUplink(test.Randy, *types.NewPopulatedAES128Key(test.Randy), *types.NewPopulatedAES128Key(test.Randy), false)
		msg.RxMetadata = nil
		_, err = ns.HandleUplink(context.Background(), msg)
		a.So(err, should.NotBeNil)

		msg = ttnpb.NewPopulatedUplinkMessageUplink(test.Randy, *types.NewPopulatedAES128Key(test.Randy), *types.NewPopulatedAES128Key(test.Randy), false)
		for len(msg.GetRxMetadata()) == 0 {
			msg = ttnpb.NewPopulatedUplinkMessageUplink(test.Randy, *types.NewPopulatedAES128Key(test.Randy), *types.NewPopulatedAES128Key(test.Randy), false)
		}
		msg.RxMetadata[0].ChannelIndex = math.MaxUint8 + 1
		_, err = ns.HandleUplink(context.Background(), msg)
		a.So(err, should.NotBeNil)

		msg = ttnpb.NewPopulatedUplinkMessageUplink(test.Randy, *types.NewPopulatedAES128Key(test.Randy), *types.NewPopulatedAES128Key(test.Randy), false)
		msg.Settings.DataRateIndex = math.MaxUint8 + 1
		_, err = ns.HandleUplink(context.Background(), msg)
		a.So(err, should.NotBeNil)

		for _, tc := range []struct {
			Name string

			Device         *ttnpb.EndDevice
			NextNextFCntUp uint32
			UplinkMessage  *ttnpb.UplinkMessage
		}{
			// TODO: Add test cases
		} {
			t.Run(tc.Name, func(t *testing.T) {
				// TODO: Implement test
			})
		}
	}
}
