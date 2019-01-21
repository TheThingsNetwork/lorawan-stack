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
	"context"
	"testing"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/errors"
	. "go.thethings.network/lorawan-stack/pkg/networkserver"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/unique"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

type getByIDFuncKey struct{}
type setByIDFuncKey struct{}

var registeredDevice = &ttnpb.EndDevice{
	EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
		ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
			ApplicationID: "foo-application",
		},
		DeviceID: "foo-device",
		JoinEUI:  eui64Ptr(types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}),
		DevEUI:   eui64Ptr(types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}),
	},
	RootKeys: &ttnpb.RootKeys{
		RootKeyID: "test",
		NwkKey: &ttnpb.KeyEnvelope{
			KEKLabel: "test",
			Key:      []byte{0x1, 0x2},
		},
		AppKey: &ttnpb.KeyEnvelope{
			KEKLabel: "test",
			Key:      []byte{0x3, 0x4},
		},
	},
}

func TestDeviceRegistryGet(t *testing.T) {
	for _, tc := range []struct {
		Name             string
		ContextFunc      func(context.Context) context.Context
		GetByIDFunc      func(context.Context, ttnpb.ApplicationIdentifiers, string, []string) (*ttnpb.EndDevice, error)
		DeviceRequest    *ttnpb.GetEndDeviceRequest
		ErrorAssertion   func(*testing.T, error) bool
		ContextAssertion func(context.Context) bool
	}{
		{
			Name: "Permission denied",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, registeredDevice.ApplicationIdentifiers): nil,
					},
				})
			},
			GetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string) (*ttnpb.EndDevice, error) {
				defer test.MustIncrementContextCounter(ctx, getByIDFuncKey{}, 1)
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(appID, should.Resemble, registeredDevice.EndDeviceIdentifiers.ApplicationIdentifiers)
				a.So(devID, should.Equal, registeredDevice.EndDeviceIdentifiers.DeviceID)
				a.So(paths, should.HaveSameElementsDeep, []string{"ids"})
				return registeredDevice, nil
			},
			DeviceRequest: &ttnpb.GetEndDeviceRequest{
				EndDeviceIdentifiers: registeredDevice.EndDeviceIdentifiers,
				FieldMask: pbtypes.FieldMask{
					Paths: []string{"ids"},
				}},
			ErrorAssertion: func(t *testing.T, err error) bool {
				a := assertions.New(t)
				return a.So(errors.IsPermissionDenied(err), should.BeTrue)
			},
			ContextAssertion: func(ctx context.Context) bool {
				a := assertions.New(test.MustTFromContext(ctx))
				return a.So(test.MustCounterFromContext(ctx, getByIDFuncKey{}), should.Equal, 0)
			},
		},
		{
			Name: "Get without key",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, registeredDevice.ApplicationIdentifiers): ttnpb.RightsFrom(
							ttnpb.RIGHT_APPLICATION_DEVICES_READ,
						),
					},
				})
			},
			GetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string) (*ttnpb.EndDevice, error) {
				defer test.MustIncrementContextCounter(ctx, getByIDFuncKey{}, 1)
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(appID, should.Resemble, registeredDevice.EndDeviceIdentifiers.ApplicationIdentifiers)
				a.So(devID, should.Equal, registeredDevice.EndDeviceIdentifiers.DeviceID)
				a.So(paths, should.HaveSameElementsDeep, []string{"ids"})
				return registeredDevice, nil
			},
			DeviceRequest: &ttnpb.GetEndDeviceRequest{
				EndDeviceIdentifiers: registeredDevice.EndDeviceIdentifiers,
				FieldMask: pbtypes.FieldMask{
					Paths: []string{"ids"},
				},
			},
			ContextAssertion: func(ctx context.Context) bool {
				a := assertions.New(test.MustTFromContext(ctx))
				return a.So(test.MustCounterFromContext(ctx, getByIDFuncKey{}), should.Equal, 1)
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)
			ctx := test.ContextWithCounter(tc.ContextFunc(test.ContextWithT(test.Context(), t)), getByIDFuncKey{})
			reg := &MockDeviceRegistry{
				GetByIDFunc: tc.GetByIDFunc,
			}
			ns := test.Must(New(
				component.MustNew(test.GetLogger(t), &component.Config{}),
				&Config{
					Devices:             reg,
					DeduplicationWindow: 42,
					CooldownWindow:      42,
					DownlinkTasks:       &MockDownlinkTaskQueue{},
				})).(*NetworkServer)
			test.Must(nil, ns.Start())
			defer ns.Close()
			dev, err := ns.Get(ctx, tc.DeviceRequest)
			a.So(tc.ContextAssertion(ctx), should.BeTrue)
			if tc.ErrorAssertion != nil {
				a.So(tc.ErrorAssertion(t, err), should.BeTrue)
				a.So(dev, should.BeNil)
				return
			}
			a.So(err, should.BeNil)
			a.So(dev, should.Resemble, registeredDevice)
		})
	}
}

func TestDeviceRegistrySet(t *testing.T) {
	for _, tc := range []struct {
		Name             string
		ContextFunc      func(context.Context) context.Context
		SetByIDFunc      func(context.Context, ttnpb.ApplicationIdentifiers, string, []string, func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error)
		DeviceRequest    *ttnpb.SetEndDeviceRequest
		ErrorAssertion   func(*testing.T, error) bool
		ContextAssertion func(context.Context) bool
	}{
		{
			Name: "Permission denied",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, registeredDevice.ApplicationIdentifiers): nil,
					},
				})
			},
			DeviceRequest: &ttnpb.SetEndDeviceRequest{
				Device: *registeredDevice,
				FieldMask: pbtypes.FieldMask{
					Paths: []string{"ids"},
				},
			},
			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string, cb func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				defer test.MustIncrementContextCounter(ctx, setByIDFuncKey{}, 1)
				a.So(appID, should.Resemble, registeredDevice.EndDeviceIdentifiers.ApplicationIdentifiers)
				a.So(devID, should.Equal, registeredDevice.EndDeviceIdentifiers.DeviceID)
				a.So(paths, should.HaveSameElementsDeep, []string{"ids"})
				dev, fieldmask, err := cb(registeredDevice)
				if dev != nil {
					a.So(fieldmask, should.HaveSameElementsDeep, []string{"ids"})
				}
				return dev, err
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				a := assertions.New(t)
				return a.So(errors.IsPermissionDenied(err), should.BeTrue)
			},
			ContextAssertion: func(ctx context.Context) bool {
				a := assertions.New(test.MustTFromContext(ctx))
				return a.So(test.MustCounterFromContext(ctx, setByIDFuncKey{}), should.Equal, 0)
			},
		},
		{
			Name: "Set without keys",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, registeredDevice.ApplicationIdentifiers): ttnpb.RightsFrom(
							ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
						),
					},
				})
			},
			DeviceRequest: &ttnpb.SetEndDeviceRequest{
				Device: *registeredDevice,
				FieldMask: pbtypes.FieldMask{
					Paths: []string{"ids"},
				},
			},
			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string, cb func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				defer test.MustIncrementContextCounter(ctx, setByIDFuncKey{}, 1)
				a.So(appID, should.Resemble, registeredDevice.EndDeviceIdentifiers.ApplicationIdentifiers)
				a.So(devID, should.Equal, registeredDevice.EndDeviceIdentifiers.DeviceID)
				a.So(paths, should.HaveSameElementsDeep, []string{"ids"})
				dev, fieldmask, err := cb(registeredDevice)
				if dev != nil {
					a.So(fieldmask, should.HaveSameElementsDeep, []string{"ids"})
				}
				return dev, err
			},
			ContextAssertion: func(ctx context.Context) bool {
				a := assertions.New(test.MustTFromContext(ctx))
				return a.So(test.MustCounterFromContext(ctx, setByIDFuncKey{}), should.Equal, 1)
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)
			ctx := test.ContextWithCounter(tc.ContextFunc(test.ContextWithT(test.Context(), t)), setByIDFuncKey{})
			reg := &MockDeviceRegistry{
				SetByIDFunc: tc.SetByIDFunc,
			}
			ns := test.Must(New(
				component.MustNew(test.GetLogger(t), &component.Config{}),
				&Config{
					Devices:             reg,
					DeduplicationWindow: 42,
					CooldownWindow:      42,
					DownlinkTasks:       &MockDownlinkTaskQueue{},
				})).(*NetworkServer)
			test.Must(nil, ns.Start())
			defer ns.Close()
			dev, err := ns.Set(ctx, tc.DeviceRequest)
			a.So(tc.ContextAssertion(ctx), should.BeTrue)
			if tc.ErrorAssertion != nil {
				a.So(tc.ErrorAssertion(t, err), should.BeTrue)
				a.So(dev, should.BeNil)
				return
			}
			a.So(err, should.BeNil)
			a.So(dev, should.Resemble, registeredDevice)
		})
	}
}

func TestDeviceRegistryDelete(t *testing.T) {

	for _, tc := range []struct {
		Name             string
		ContextFunc      func(context.Context) context.Context
		SetByIDFunc      func(context.Context, ttnpb.ApplicationIdentifiers, string, []string, func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error)
		Device           *ttnpb.EndDeviceIdentifiers
		ErrorAssertion   func(*testing.T, error) bool
		ContextAssertion func(context.Context) bool
	}{
		{
			Name: "Permission denied",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, registeredDevice.ApplicationIdentifiers): nil,
					},
				})
			},
			Device: &registeredDevice.EndDeviceIdentifiers,
			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string, cb func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				defer test.MustIncrementContextCounter(ctx, setByIDFuncKey{}, 1)
				a.So(appID, should.Resemble, registeredDevice.EndDeviceIdentifiers.ApplicationIdentifiers)
				a.So(devID, should.Equal, registeredDevice.EndDeviceIdentifiers.DeviceID)
				a.So(paths, should.BeNil)
				dev, _, err := cb(registeredDevice)
				return dev, err
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				a := assertions.New(t)
				return a.So(errors.IsPermissionDenied(err), should.BeTrue)
			},
			ContextAssertion: func(ctx context.Context) bool {
				a := assertions.New(test.MustTFromContext(ctx))
				return a.So(test.MustCounterFromContext(ctx, setByIDFuncKey{}), should.Equal, 0)
			},
		},
		{
			Name: "Delete",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, registeredDevice.ApplicationIdentifiers): ttnpb.RightsFrom(
							ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
						),
					},
				})
			},
			Device: &registeredDevice.EndDeviceIdentifiers,
			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string, cb func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				defer test.MustIncrementContextCounter(ctx, setByIDFuncKey{}, 1)
				a.So(appID, should.Resemble, registeredDevice.EndDeviceIdentifiers.ApplicationIdentifiers)
				a.So(devID, should.Equal, registeredDevice.EndDeviceIdentifiers.DeviceID)
				a.So(paths, should.BeNil)
				dev, _, err := cb(registeredDevice)
				return dev, err
			},
			ContextAssertion: func(ctx context.Context) bool {
				a := assertions.New(test.MustTFromContext(ctx))
				return a.So(test.MustCounterFromContext(ctx, setByIDFuncKey{}), should.Equal, 1)
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)
			ctx := test.ContextWithCounter(tc.ContextFunc(test.ContextWithT(test.Context(), t)), setByIDFuncKey{})
			reg := &MockDeviceRegistry{
				SetByIDFunc: tc.SetByIDFunc,
			}
			ns := test.Must(New(
				component.MustNew(test.GetLogger(t), &component.Config{}),
				&Config{
					Devices:             reg,
					DeduplicationWindow: 42,
					CooldownWindow:      42,
					DownlinkTasks:       &MockDownlinkTaskQueue{},
				})).(*NetworkServer)
			test.Must(nil, ns.Start())
			defer ns.Close()
			dev, err := ns.Delete(ctx, tc.Device)
			a.So(tc.ContextAssertion(ctx), should.BeTrue)
			if tc.ErrorAssertion != nil {
				a.So(tc.ErrorAssertion(t, err), should.BeTrue)
				a.So(dev, should.BeNil)
				return
			}
			a.So(err, should.BeNil)
			a.So(dev, should.Resemble, ttnpb.Empty)
		})
	}
}
