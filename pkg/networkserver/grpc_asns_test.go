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
	"sync"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/config"
	"go.thethings.network/lorawan-stack/pkg/errors"
	. "go.thethings.network/lorawan-stack/pkg/networkserver"
	"go.thethings.network/lorawan-stack/pkg/networkserver/redis"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestLinkApplication(t *testing.T) {
	a := assertions.New(t)

	redisClient, flush := test.NewRedis(t, "networkserver_test")
	defer flush()
	defer redisClient.Close()
	devReg := &redis.DeviceRegistry{Redis: redisClient}

	ns := test.Must(New(
		component.MustNew(test.GetLogger(t), &component.Config{
			ServiceBase: config.ServiceBase{Cluster: config.Cluster{Keys: Keys}},
		}),
		&Config{
			Devices:             devReg,
			DeduplicationWindow: 42,
			CooldownWindow:      42,
		})).(*NetworkServer)
	test.Must(nil, ns.Start())
	defer ns.Close()

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

func TestDownlinkQueueReplace(t *testing.T) {
	ids := ttnpb.EndDeviceIdentifiers{
		ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
			ApplicationID: ApplicationID,
		},
		DeviceID: DeviceID,
		JoinEUI:  &JoinEUI,
		DevEUI:   &DevEUI,
	}

	for _, tc := range []struct {
		Name           string
		Context        context.Context
		Device         *ttnpb.EndDevice
		Request        *ttnpb.DownlinkQueueRequest
		ErrorAssertion func(*testing.T, error) bool
	}{
		{
			Name:    "no device",
			Context: test.Context(),
			Request: ttnpb.NewPopulatedDownlinkQueueRequest(test.Randy, false),
			ErrorAssertion: func(t *testing.T, err error) bool {
				a := assertions.New(t)
				return a.So(err, should.BeError) && a.So(errors.IsNotFound(err), should.BeTrue)
			},
		},
		{
			Name:    "empty queue/empty request",
			Context: test.Context(),
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers:       ids,
				QueuedApplicationDownlinks: nil,
			},
			Request: &ttnpb.DownlinkQueueRequest{
				EndDeviceIdentifiers: ids,
				Downlinks:            nil,
			},
		},
		{
			Name:    "empty queue/non-empty request",
			Context: test.Context(),
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers:       ids,
				QueuedApplicationDownlinks: nil,
			},
			Request: &ttnpb.DownlinkQueueRequest{
				EndDeviceIdentifiers: ids,
				Downlinks: []*ttnpb.ApplicationDownlink{
					Downlinks[0],
					Downlinks[2],
					Downlinks[1],
				},
			},
		},
		{
			Name:    "non-empty queue/empty request",
			Context: test.Context(),
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ids,
				QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
					Downlinks[2],
					Downlinks[1],
					Downlinks[0],
					Downlinks[4],
				},
			},
			Request: &ttnpb.DownlinkQueueRequest{
				EndDeviceIdentifiers: ids,
			},
		},
		{
			Name:    "non-empty queue/non-empty request",
			Context: test.Context(),
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ids,
				QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
					Downlinks[2],
					Downlinks[1],
					Downlinks[0],
					Downlinks[4],
				},
			},
			Request: &ttnpb.DownlinkQueueRequest{
				EndDeviceIdentifiers: ids,
				Downlinks: []*ttnpb.ApplicationDownlink{
					Downlinks[0],
					Downlinks[2],
					Downlinks[1],
				},
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

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
				})).(*NetworkServer)
			test.Must(nil, ns.Start())
			defer ns.Close()

			pb := CopyEndDevice(tc.Device)
			if tc.Device != nil {
				start := time.Now()
				ret, err := CreateDevice(tc.Context, devReg, CopyEndDevice(tc.Device))
				if !a.So(err, should.BeNil) {
					t.FailNow()
				}
				a.So(ret.CreatedAt, should.HappenAfter, start)
				a.So(ret.UpdatedAt, should.HappenAfter, start)
				a.So(ret.UpdatedAt, should.Equal, ret.CreatedAt)
				pb.CreatedAt = ret.CreatedAt
				pb.UpdatedAt = ret.UpdatedAt
				a.So(ret, should.Resemble, pb)
			}

			_, err := ns.DownlinkQueueReplace(tc.Context, tc.Request)
			if tc.ErrorAssertion != nil {
				a.So(tc.ErrorAssertion(t, err), should.BeTrue)
				return
			}
			a.So(err, should.BeNil)

			pb.QueuedApplicationDownlinks = tc.Request.Downlinks

			ret, err := devReg.GetByID(tc.Context, pb.EndDeviceIdentifiers.ApplicationIdentifiers, pb.EndDeviceIdentifiers.DeviceID, ttnpb.EndDeviceFieldPathsTopLevel)
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}
			a.So(ret.UpdatedAt, should.HappenAfter, pb.UpdatedAt)
			pb.UpdatedAt = ret.UpdatedAt
			a.So(ret, should.HaveEmptyDiff, pb)
		})
	}
}

func TestDownlinkQueuePush(t *testing.T) {
	ids := ttnpb.EndDeviceIdentifiers{
		ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
			ApplicationID: ApplicationID,
		},
		DeviceID: DeviceID,
		JoinEUI:  &JoinEUI,
		DevEUI:   &DevEUI,
	}

	for _, tc := range []struct {
		Name           string
		Context        context.Context
		Device         *ttnpb.EndDevice
		Request        *ttnpb.DownlinkQueueRequest
		ErrorAssertion func(*testing.T, error) bool
	}{
		{
			Name:    "no device",
			Context: test.Context(),
			Request: ttnpb.NewPopulatedDownlinkQueueRequest(test.Randy, false),
			ErrorAssertion: func(t *testing.T, err error) bool {
				a := assertions.New(t)
				return a.So(err, should.BeError) && a.So(errors.IsNotFound(err), should.BeTrue)
			},
		},
		{
			Name:    "empty queue/empty request",
			Context: test.Context(),
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers:       ids,
				QueuedApplicationDownlinks: nil,
			},
			Request: &ttnpb.DownlinkQueueRequest{
				EndDeviceIdentifiers: ids,
				Downlinks:            nil,
			},
		},
		{
			Name:    "empty queue/non-empty request",
			Context: test.Context(),
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers:       ids,
				QueuedApplicationDownlinks: nil,
			},
			Request: &ttnpb.DownlinkQueueRequest{
				EndDeviceIdentifiers: ids,
				Downlinks: []*ttnpb.ApplicationDownlink{
					Downlinks[0],
					Downlinks[2],
					Downlinks[1],
				},
			},
		},
		{
			Name:    "non-empty queue/empty request",
			Context: test.Context(),
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ids,
				QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
					Downlinks[2],
					Downlinks[1],
					Downlinks[0],
					Downlinks[4],
				},
			},
			Request: &ttnpb.DownlinkQueueRequest{
				EndDeviceIdentifiers: ids,
			},
		},
		{
			Name:    "non-empty queue/non-empty request",
			Context: test.Context(),
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ids,
				QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
					Downlinks[2],
					Downlinks[1],
					Downlinks[0],
					Downlinks[4],
				},
			},
			Request: &ttnpb.DownlinkQueueRequest{
				EndDeviceIdentifiers: ids,
				Downlinks: []*ttnpb.ApplicationDownlink{
					Downlinks[0],
					Downlinks[2],
					Downlinks[1],
				},
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

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
				})).(*NetworkServer)
			test.Must(nil, ns.Start())
			defer ns.Close()

			pb := CopyEndDevice(tc.Device)
			if tc.Device != nil {
				start := time.Now()
				ret, err := CreateDevice(tc.Context, devReg, CopyEndDevice(tc.Device))
				if !a.So(err, should.BeNil) {
					t.FailNow()
				}
				a.So(ret.CreatedAt, should.HappenAfter, start)
				a.So(ret.UpdatedAt, should.HappenAfter, start)
				a.So(ret.UpdatedAt, should.Equal, ret.CreatedAt)
				pb.CreatedAt = ret.CreatedAt
				pb.UpdatedAt = ret.UpdatedAt
				a.So(ret, should.Resemble, pb)
			}

			_, err := ns.DownlinkQueuePush(tc.Context, tc.Request)
			if tc.ErrorAssertion != nil {
				a.So(tc.ErrorAssertion(t, err), should.BeTrue)
				return
			}
			a.So(err, should.BeNil)

			pb.QueuedApplicationDownlinks = append(pb.QueuedApplicationDownlinks, tc.Request.Downlinks...)

			ret, err := devReg.GetByID(tc.Context, pb.EndDeviceIdentifiers.ApplicationIdentifiers, pb.EndDeviceIdentifiers.DeviceID, ttnpb.EndDeviceFieldPathsTopLevel)
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}
			a.So(ret.UpdatedAt, should.HappenAfter, pb.UpdatedAt)
			pb.UpdatedAt = ret.UpdatedAt
			a.So(ret, should.HaveEmptyDiff, pb)
		})
	}
}

func TestDownlinkQueueList(t *testing.T) {
	ids := ttnpb.EndDeviceIdentifiers{
		ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
			ApplicationID: ApplicationID,
		},
		DeviceID: DeviceID,
		JoinEUI:  &JoinEUI,
		DevEUI:   &DevEUI,
	}

	for _, tc := range []struct {
		Name           string
		Context        context.Context
		Device         *ttnpb.EndDevice
		Request        *ttnpb.EndDeviceIdentifiers
		ErrorAssertion func(*testing.T, error) bool
	}{
		{
			Name:    "no device",
			Context: test.Context(),
			Request: ttnpb.NewPopulatedEndDeviceIdentifiers(test.Randy, false),
			ErrorAssertion: func(t *testing.T, err error) bool {
				a := assertions.New(t)
				return a.So(err, should.BeError) && a.So(errors.IsNotFound(err), should.BeTrue)
			},
		},
		{
			Name:    "empty identifiers",
			Context: test.Context(),
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers:       ids,
				QueuedApplicationDownlinks: nil,
			},
			Request: &ttnpb.EndDeviceIdentifiers{},
			ErrorAssertion: func(t *testing.T, err error) bool {
				a := assertions.New(t)
				return a.So(err, should.BeError) && a.So(errors.IsInvalidArgument(err), should.BeTrue)
			},
		},
		{
			Name:    "empty queue",
			Context: test.Context(),
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers:       ids,
				QueuedApplicationDownlinks: nil,
			},
			Request: &ids,
		},
		{
			Name:    "non-empty queue",
			Context: test.Context(),
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ids,
				QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
					Downlinks[2],
					Downlinks[1],
					Downlinks[0],
					Downlinks[4],
				},
			},
			Request: &ids,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

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
				})).(*NetworkServer)
			test.Must(nil, ns.Start())
			defer ns.Close()

			pb := CopyEndDevice(tc.Device)
			if tc.Device != nil {
				start := time.Now()
				ret, err := CreateDevice(tc.Context, devReg, CopyEndDevice(tc.Device))
				if !a.So(err, should.BeNil) {
					t.FailNow()
				}
				if !a.So(err, should.BeNil) {
					t.FailNow()
				}
				a.So(ret.CreatedAt, should.HappenAfter, start)
				a.So(ret.UpdatedAt, should.HappenAfter, start)
				a.So(ret.UpdatedAt, should.Equal, ret.CreatedAt)
				pb.CreatedAt = ret.CreatedAt
				pb.UpdatedAt = ret.UpdatedAt
				a.So(ret, should.Resemble, pb)
			}

			resp, err := ns.DownlinkQueueList(tc.Context, tc.Request)
			if tc.ErrorAssertion != nil {
				a.So(tc.ErrorAssertion(t, err), should.BeTrue)
				a.So(resp, should.BeNil)
				return
			}
			a.So(err, should.BeNil)
			a.So(resp, should.HaveEmptyDiff, &ttnpb.ApplicationDownlinks{Downlinks: pb.QueuedApplicationDownlinks})
		})
	}
}
