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

package joinserver_test

import (
	"context"
	"testing"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/mohae/deepcopy"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/errors"
	. "go.thethings.network/lorawan-stack/pkg/joinserver"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/unique"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

type getByEUIFuncKey struct{}
type setByEUIFuncKey struct{}

var (
	registeredJoinEUI = eui64Ptr(types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
	registeredDevEUI  = eui64Ptr(types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
	registeredNwkKey  = []byte{0x1, 0x2}
	registeredAppKey  = []byte{0x3, 0x4}

	unregisteredJoinEUI = eui64Ptr(types.EUI64{0x42, 0x42, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	unregisteredDevEUI  = eui64Ptr(types.EUI64{0x42, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
)

const (
	registeredApplicationID = "foo-application"
	registeredDeviceID      = "foo-device"
	registeredRootKeyID     = "testKey"
	registeredKEKLabel      = "test"

	unregisteredDeviceID = "bar-device"
)

var errNotFound = errors.DefineNotFound("not_found", "not found")

func eui64Ptr(eui types.EUI64) *types.EUI64 { return &eui }

func TestDeviceRegistryGet(t *testing.T) {
	registeredDevice := &ttnpb.EndDevice{
		EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
			ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
				ApplicationID: registeredApplicationID,
			},
			DeviceID: registeredDeviceID,
			JoinEUI:  registeredJoinEUI,
			DevEUI:   registeredDevEUI,
		},
		RootKeys: &ttnpb.RootKeys{
			RootKeyID: registeredRootKeyID,
			NwkKey: &ttnpb.KeyEnvelope{
				KEKLabel: registeredKEKLabel,
				Key:      registeredNwkKey,
			},
			AppKey: &ttnpb.KeyEnvelope{
				KEKLabel: registeredKEKLabel,
				Key:      registeredAppKey,
			},
		},
	}
	for _, tc := range []struct {
		Name             string
		ContextFunc      func(context.Context) context.Context
		GetByEUIFunc     func(ctx context.Context, joinEUI, devEUI types.EUI64, paths []string) (*ttnpb.EndDevice, error)
		DeviceRequest    *ttnpb.GetEndDeviceRequest
		ErrorAssertion   func(*testing.T, error) bool
		ContextAssertion func(context.Context) bool
	}{
		{
			Name: "Permission denied",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, ttnpb.ApplicationIdentifiers{ApplicationID: registeredApplicationID}): nil,
					},
				})
			},
			GetByEUIFunc: func(ctx context.Context, joinEUI, devEUI types.EUI64, paths []string) (*ttnpb.EndDevice, error) {
				defer test.MustIncrementContextCounter(ctx, getByEUIFuncKey{}, 1)
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(joinEUI, should.Equal, *registeredJoinEUI)
				a.So(devEUI, should.Equal, *registeredDevEUI)
				a.So(paths, should.HaveSameElementsDeep, []string{"ids"})
				return deepcopy.Copy(registeredDevice).(*ttnpb.EndDevice), nil
			},
			DeviceRequest: &ttnpb.GetEndDeviceRequest{
				EndDeviceIdentifiers: deepcopy.Copy(registeredDevice.EndDeviceIdentifiers).(ttnpb.EndDeviceIdentifiers),
				FieldMask: pbtypes.FieldMask{
					Paths: []string{"ids"},
				}},
			ErrorAssertion: func(t *testing.T, err error) bool {
				a := assertions.New(t)
				return a.So(errors.IsPermissionDenied(err), should.BeTrue)
			},
			ContextAssertion: func(ctx context.Context) bool {
				a := assertions.New(test.MustTFromContext(ctx))
				return a.So(test.MustCounterFromContext(ctx, getByEUIFuncKey{}), should.Equal, 0)
			},
		},
		{
			Name: "No DevEUI",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, deepcopy.Copy(registeredDevice.EndDeviceIdentifiers.ApplicationIdentifiers).(ttnpb.ApplicationIdentifiers)): ttnpb.RightsFrom(
							ttnpb.RIGHT_APPLICATION_DEVICES_READ,
						),
					},
				})
			},
			GetByEUIFunc: func(ctx context.Context, joinEUI, devEUI types.EUI64, paths []string) (*ttnpb.EndDevice, error) {
				defer test.MustIncrementContextCounter(ctx, getByEUIFuncKey{}, 1)
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(joinEUI, should.Equal, registeredJoinEUI)
				a.So(devEUI, should.Equal, registeredDevEUI)
				a.So(paths, should.HaveSameElementsDeep, []string{"ids"})
				return deepcopy.Copy(registeredDevice).(*ttnpb.EndDevice), nil
			},
			DeviceRequest: &ttnpb.GetEndDeviceRequest{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
						ApplicationID: registeredApplicationID,
					},
					DeviceID: registeredDeviceID,
					JoinEUI:  registeredJoinEUI,
				},
				FieldMask: pbtypes.FieldMask{
					Paths: []string{"ids"},
				},
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				a := assertions.New(t)
				return a.So(errors.IsInvalidArgument(err), should.BeTrue)
			},
			ContextAssertion: func(ctx context.Context) bool {
				a := assertions.New(test.MustTFromContext(ctx))
				return a.So(test.MustCounterFromContext(ctx, getByEUIFuncKey{}), should.Equal, 0)
			},
		},
		{
			Name: "No JoinEUI",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, deepcopy.Copy(registeredDevice.EndDeviceIdentifiers.ApplicationIdentifiers).(ttnpb.ApplicationIdentifiers)): ttnpb.RightsFrom(
							ttnpb.RIGHT_APPLICATION_DEVICES_READ,
						),
					},
				})
			},
			GetByEUIFunc: func(ctx context.Context, joinEUI, devEUI types.EUI64, paths []string) (*ttnpb.EndDevice, error) {
				defer test.MustIncrementContextCounter(ctx, getByEUIFuncKey{}, 1)
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(joinEUI, should.Equal, registeredJoinEUI)
				a.So(devEUI, should.Equal, registeredDevEUI)
				a.So(paths, should.HaveSameElementsDeep, []string{"ids"})
				return deepcopy.Copy(registeredDevice).(*ttnpb.EndDevice), nil
			},
			DeviceRequest: &ttnpb.GetEndDeviceRequest{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
						ApplicationID: registeredApplicationID,
					},
					DeviceID: registeredDeviceID,
					DevEUI:   registeredDevEUI,
				},
				FieldMask: pbtypes.FieldMask{
					Paths: []string{"ids"},
				},
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				a := assertions.New(t)
				return a.So(errors.IsInvalidArgument(err), should.BeTrue)
			},
			ContextAssertion: func(ctx context.Context) bool {
				a := assertions.New(test.MustTFromContext(ctx))
				return a.So(test.MustCounterFromContext(ctx, getByEUIFuncKey{}), should.Equal, 0)
			},
		},
		{
			Name: "Not found",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, deepcopy.Copy(registeredDevice.EndDeviceIdentifiers.ApplicationIdentifiers).(ttnpb.ApplicationIdentifiers)): ttnpb.RightsFrom(
							ttnpb.RIGHT_APPLICATION_DEVICES_READ,
						),
					},
				})
			},
			GetByEUIFunc: func(ctx context.Context, joinEUI, devEUI types.EUI64, paths []string) (*ttnpb.EndDevice, error) {
				defer test.MustIncrementContextCounter(ctx, getByEUIFuncKey{}, 1)
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(joinEUI, should.Equal, *unregisteredJoinEUI)
				a.So(devEUI, should.Equal, *unregisteredDevEUI)
				a.So(paths, should.HaveSameElementsDeep, []string{"ids"})
				return nil, errNotFound
			},
			DeviceRequest: &ttnpb.GetEndDeviceRequest{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
						ApplicationID: registeredApplicationID,
					},
					DeviceID: unregisteredDeviceID,
					JoinEUI:  unregisteredJoinEUI,
					DevEUI:   unregisteredDevEUI,
				},
				FieldMask: pbtypes.FieldMask{
					Paths: []string{"ids"},
				},
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				a := assertions.New(t)
				return a.So(errors.IsNotFound(err), should.BeTrue)
			},
			ContextAssertion: func(ctx context.Context) bool {
				a := assertions.New(test.MustTFromContext(ctx))
				return a.So(test.MustCounterFromContext(ctx, getByEUIFuncKey{}), should.Equal, 1)
			},
		},
		{
			Name: "Invalid application ID",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, ttnpb.ApplicationIdentifiers{ApplicationID: "other-app"}): ttnpb.RightsFrom(
							ttnpb.RIGHT_APPLICATION_DEVICES_READ,
						),
					},
				})
			},
			GetByEUIFunc: func(ctx context.Context, joinEUI, devEUI types.EUI64, paths []string) (*ttnpb.EndDevice, error) {
				defer test.MustIncrementContextCounter(ctx, getByEUIFuncKey{}, 1)
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(joinEUI, should.Equal, *registeredJoinEUI)
				a.So(devEUI, should.Equal, *registeredDevEUI)
				a.So(paths, should.HaveSameElementsDeep, []string{"ids"})
				return deepcopy.Copy(registeredDevice).(*ttnpb.EndDevice), nil
			},
			DeviceRequest: &ttnpb.GetEndDeviceRequest{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
						ApplicationID: "other-app",
					},
					DeviceID: "other-device",
					JoinEUI:  registeredJoinEUI,
					DevEUI:   registeredDevEUI,
				},
				FieldMask: pbtypes.FieldMask{
					Paths: []string{"ids"},
				},
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				a := assertions.New(t)
				return a.So(errors.IsNotFound(err), should.BeTrue)
			},
			ContextAssertion: func(ctx context.Context) bool {
				a := assertions.New(test.MustTFromContext(ctx))
				return a.So(test.MustCounterFromContext(ctx, getByEUIFuncKey{}), should.Equal, 1)
			},
		},
		{
			Name: "Get without key",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, deepcopy.Copy(registeredDevice.EndDeviceIdentifiers.ApplicationIdentifiers).(ttnpb.ApplicationIdentifiers)): ttnpb.RightsFrom(
							ttnpb.RIGHT_APPLICATION_DEVICES_READ,
						),
					},
				})
			},
			GetByEUIFunc: func(ctx context.Context, joinEUI, devEUI types.EUI64, paths []string) (*ttnpb.EndDevice, error) {
				defer test.MustIncrementContextCounter(ctx, getByEUIFuncKey{}, 1)
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(joinEUI, should.Equal, *registeredJoinEUI)
				a.So(devEUI, should.Equal, *registeredDevEUI)
				a.So(paths, should.HaveSameElementsDeep, []string{"ids"})
				return deepcopy.Copy(registeredDevice).(*ttnpb.EndDevice), nil
			},
			DeviceRequest: &ttnpb.GetEndDeviceRequest{
				EndDeviceIdentifiers: deepcopy.Copy(registeredDevice.EndDeviceIdentifiers).(ttnpb.EndDeviceIdentifiers),
				FieldMask: pbtypes.FieldMask{
					Paths: []string{"ids"},
				},
			},
			ContextAssertion: func(ctx context.Context) bool {
				a := assertions.New(test.MustTFromContext(ctx))
				return a.So(test.MustCounterFromContext(ctx, getByEUIFuncKey{}), should.Equal, 1)
			},
		},
		{
			Name: "Get keys without permission",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, deepcopy.Copy(registeredDevice.EndDeviceIdentifiers.ApplicationIdentifiers).(ttnpb.ApplicationIdentifiers)): ttnpb.RightsFrom(
							ttnpb.RIGHT_APPLICATION_DEVICES_READ,
						),
					},
				})
			},
			GetByEUIFunc: func(ctx context.Context, joinEUI, devEUI types.EUI64, paths []string) (*ttnpb.EndDevice, error) {
				defer test.MustIncrementContextCounter(ctx, getByEUIFuncKey{}, 1)
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(joinEUI, should.Equal, *registeredJoinEUI)
				a.So(devEUI, should.Equal, *registeredDevEUI)
				a.So(paths, should.HaveSameElementsDeep, []string{"ids", "root_keys"})
				return deepcopy.Copy(registeredDevice).(*ttnpb.EndDevice), nil
			},
			DeviceRequest: &ttnpb.GetEndDeviceRequest{
				EndDeviceIdentifiers: deepcopy.Copy(registeredDevice.EndDeviceIdentifiers).(ttnpb.EndDeviceIdentifiers),
				FieldMask: pbtypes.FieldMask{
					Paths: []string{"ids", "root_keys"},
				},
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				a := assertions.New(t)
				return a.So(errors.IsPermissionDenied(err), should.BeTrue)
			},
			ContextAssertion: func(ctx context.Context) bool {
				a := assertions.New(test.MustTFromContext(ctx))
				return a.So(test.MustCounterFromContext(ctx, getByEUIFuncKey{}), should.Equal, 0)
			},
		},
		{
			Name: "Get keys",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, deepcopy.Copy(registeredDevice.EndDeviceIdentifiers.ApplicationIdentifiers).(ttnpb.ApplicationIdentifiers)): ttnpb.RightsFrom(
							ttnpb.RIGHT_APPLICATION_DEVICES_READ,
							ttnpb.RIGHT_APPLICATION_DEVICES_READ_KEYS,
						),
					},
				})
			},
			GetByEUIFunc: func(ctx context.Context, joinEUI, devEUI types.EUI64, paths []string) (*ttnpb.EndDevice, error) {
				defer test.MustIncrementContextCounter(ctx, getByEUIFuncKey{}, 1)
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(joinEUI, should.Equal, *registeredJoinEUI)
				a.So(devEUI, should.Equal, *registeredDevEUI)
				a.So(paths, should.HaveSameElementsDeep, []string{"ids", "root_keys"})
				return deepcopy.Copy(registeredDevice).(*ttnpb.EndDevice), nil
			},
			DeviceRequest: &ttnpb.GetEndDeviceRequest{
				EndDeviceIdentifiers: deepcopy.Copy(registeredDevice.EndDeviceIdentifiers).(ttnpb.EndDeviceIdentifiers),
				FieldMask: pbtypes.FieldMask{
					Paths: []string{"ids", "root_keys"},
				},
			},
			ContextAssertion: func(ctx context.Context) bool {
				a := assertions.New(test.MustTFromContext(ctx))
				return a.So(test.MustCounterFromContext(ctx, getByEUIFuncKey{}), should.Equal, 1)
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)
			ctx := test.ContextWithCounter(tc.ContextFunc(test.ContextWithT(test.Context(), t)), getByEUIFuncKey{})
			reg := &MockDeviceRegistry{
				GetByEUIFunc: tc.GetByEUIFunc,
			}
			srv := &JsDeviceServer{
				Registry: reg,
			}
			dev, err := srv.Get(ctx, tc.DeviceRequest)
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
	registeredDevice := &ttnpb.EndDevice{
		EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
			ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
				ApplicationID: registeredApplicationID,
			},
			DeviceID: registeredDeviceID,
			JoinEUI:  registeredJoinEUI,
			DevEUI:   registeredDevEUI,
		},
		RootKeys: &ttnpb.RootKeys{
			RootKeyID: registeredRootKeyID,
			NwkKey: &ttnpb.KeyEnvelope{
				KEKLabel: registeredKEKLabel,
				Key:      registeredNwkKey,
			},
			AppKey: &ttnpb.KeyEnvelope{
				KEKLabel: registeredKEKLabel,
				Key:      registeredAppKey,
			},
		},
	}
	for _, tc := range []struct {
		Name             string
		ContextFunc      func(context.Context) context.Context
		SetByEUIFunc     func(ctx context.Context, joinEUI, devEUI types.EUI64, paths []string, cb func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error)
		DeviceRequest    *ttnpb.SetEndDeviceRequest
		ErrorAssertion   func(*testing.T, error) bool
		ContextAssertion func(context.Context) bool
		DeviceAssertion  func(*testing.T, *ttnpb.EndDevice) bool
	}{
		{
			Name: "Permission denied",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, deepcopy.Copy(registeredDevice.EndDeviceIdentifiers.ApplicationIdentifiers).(ttnpb.ApplicationIdentifiers)): nil,
					},
				})
			},
			DeviceRequest: &ttnpb.SetEndDeviceRequest{
				Device: deepcopy.Copy(*registeredDevice).(ttnpb.EndDevice),
				FieldMask: pbtypes.FieldMask{
					Paths: []string{"ids"},
				},
			},
			SetByEUIFunc: func(ctx context.Context, joinEUI, devEUI types.EUI64, paths []string, cb func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				defer test.MustIncrementContextCounter(ctx, setByEUIFuncKey{}, 1)
				a.So(joinEUI, should.Equal, *registeredJoinEUI)
				a.So(devEUI, should.Equal, *registeredDevEUI)
				a.So(paths, should.Contain, "ids")
				dev, _, err := cb(deepcopy.Copy(registeredDevice).(*ttnpb.EndDevice))
				return dev, err
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				a := assertions.New(t)
				return a.So(errors.IsPermissionDenied(err), should.BeTrue)
			},
			ContextAssertion: func(ctx context.Context) bool {
				a := assertions.New(test.MustTFromContext(ctx))
				return a.So(test.MustCounterFromContext(ctx, setByEUIFuncKey{}), should.Equal, 0)
			},
		},
		{
			Name: "No DevEUI",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, deepcopy.Copy(registeredDevice.EndDeviceIdentifiers.ApplicationIdentifiers).(ttnpb.ApplicationIdentifiers)): ttnpb.RightsFrom(
							ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
						),
					},
				})
			},
			DeviceRequest: &ttnpb.SetEndDeviceRequest{
				Device: ttnpb.EndDevice{
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
							ApplicationID: registeredApplicationID,
						},
						DeviceID: registeredDeviceID,
						JoinEUI:  registeredJoinEUI,
					},
				},
				FieldMask: pbtypes.FieldMask{
					Paths: []string{"ids"},
				},
			},
			SetByEUIFunc: func(ctx context.Context, joinEUI, devEUI types.EUI64, paths []string, cb func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				defer test.MustIncrementContextCounter(ctx, setByEUIFuncKey{}, 1)
				a.So(joinEUI, should.Equal, *registeredJoinEUI)
				a.So(devEUI, should.Equal, *registeredDevEUI)
				a.So(paths, should.Contain, "ids")
				dev, _, err := cb(deepcopy.Copy(registeredDevice).(*ttnpb.EndDevice))
				return dev, err
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				a := assertions.New(t)
				return a.So(errors.IsInvalidArgument(err), should.BeTrue)
			},
			ContextAssertion: func(ctx context.Context) bool {
				a := assertions.New(test.MustTFromContext(ctx))
				return a.So(test.MustCounterFromContext(ctx, setByEUIFuncKey{}), should.Equal, 0)
			},
		},
		{
			Name: "No JoinEUI",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, deepcopy.Copy(registeredDevice.EndDeviceIdentifiers.ApplicationIdentifiers).(ttnpb.ApplicationIdentifiers)): ttnpb.RightsFrom(
							ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
						),
					},
				})
			},
			DeviceRequest: &ttnpb.SetEndDeviceRequest{
				Device: ttnpb.EndDevice{
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
							ApplicationID: registeredApplicationID,
						},
						DeviceID: registeredDeviceID,
						DevEUI:   registeredDevEUI,
					},
				},
				FieldMask: pbtypes.FieldMask{
					Paths: []string{"ids"},
				},
			},
			SetByEUIFunc: func(ctx context.Context, joinEUI, devEUI types.EUI64, paths []string, cb func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				defer test.MustIncrementContextCounter(ctx, setByEUIFuncKey{}, 1)
				a.So(joinEUI, should.Equal, registeredJoinEUI)
				a.So(devEUI, should.Equal, registeredDevEUI)
				a.So(paths, should.Contain, "ids")
				dev, _, err := cb(deepcopy.Copy(registeredDevice).(*ttnpb.EndDevice))
				return dev, err
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				a := assertions.New(t)
				return a.So(errors.IsInvalidArgument(err), should.BeTrue)
			},
			ContextAssertion: func(ctx context.Context) bool {
				a := assertions.New(test.MustTFromContext(ctx))
				return a.So(test.MustCounterFromContext(ctx, setByEUIFuncKey{}), should.Equal, 0)
			},
		},
		{
			Name: "Create",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, ttnpb.ApplicationIdentifiers{ApplicationID: registeredApplicationID}): ttnpb.RightsFrom(
							ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
						),
					},
				})
			},
			DeviceRequest: &ttnpb.SetEndDeviceRequest{
				Device: ttnpb.EndDevice{
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
							ApplicationID: registeredApplicationID,
						},
						DeviceID: "new-device",
						JoinEUI:  unregisteredJoinEUI,
						DevEUI:   unregisteredDevEUI,
					},
				},
				FieldMask: pbtypes.FieldMask{
					Paths: []string{"ids"},
				},
			},
			SetByEUIFunc: func(ctx context.Context, joinEUI, devEUI types.EUI64, paths []string, cb func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				defer test.MustIncrementContextCounter(ctx, setByEUIFuncKey{}, 1)
				a.So(joinEUI, should.Equal, *unregisteredJoinEUI)
				a.So(devEUI, should.Equal, *unregisteredDevEUI)
				a.So(paths, should.Contain, "ids")
				dev, _, err := cb(nil)
				return dev, err
			},
			ContextAssertion: func(ctx context.Context) bool {
				a := assertions.New(test.MustTFromContext(ctx))
				return a.So(test.MustCounterFromContext(ctx, setByEUIFuncKey{}), should.Equal, 1)
			},
			DeviceAssertion: func(t *testing.T, dev *ttnpb.EndDevice) bool {
				a := assertions.New(t)
				return a.So(dev, should.Resemble, &ttnpb.EndDevice{
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
							ApplicationID: registeredApplicationID,
						},
						DeviceID: "new-device",
						JoinEUI:  unregisteredJoinEUI,
						DevEUI:   unregisteredDevEUI,
					},
				})
			},
		},
		{
			Name: "Create, but registered in other application",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, ttnpb.ApplicationIdentifiers{ApplicationID: "other-app"}): ttnpb.RightsFrom(
							ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
						),
					},
				})
			},
			DeviceRequest: &ttnpb.SetEndDeviceRequest{
				Device: ttnpb.EndDevice{
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
							ApplicationID: "other-app",
						},
						DeviceID: "new-device",
						JoinEUI:  registeredJoinEUI,
						DevEUI:   registeredDevEUI,
					},
				},
				FieldMask: pbtypes.FieldMask{
					Paths: []string{"ids"},
				},
			},
			SetByEUIFunc: func(ctx context.Context, joinEUI, devEUI types.EUI64, paths []string, cb func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				defer test.MustIncrementContextCounter(ctx, setByEUIFuncKey{}, 1)
				a.So(joinEUI, should.Equal, *registeredJoinEUI)
				a.So(devEUI, should.Equal, *registeredDevEUI)
				a.So(paths, should.Contain, "ids")
				dev, _, err := cb(deepcopy.Copy(registeredDevice).(*ttnpb.EndDevice))
				return dev, err
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				a := assertions.New(t)
				return a.So(errors.IsInvalidArgument(err), should.BeTrue)
			},
			ContextAssertion: func(ctx context.Context) bool {
				a := assertions.New(test.MustTFromContext(ctx))
				return a.So(test.MustCounterFromContext(ctx, setByEUIFuncKey{}), should.Equal, 1)
			},
		},
		{
			Name: "Set without keys",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, deepcopy.Copy(registeredDevice.EndDeviceIdentifiers.ApplicationIdentifiers).(ttnpb.ApplicationIdentifiers)): ttnpb.RightsFrom(
							ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
						),
					},
				})
			},
			DeviceRequest: &ttnpb.SetEndDeviceRequest{
				Device: deepcopy.Copy(*registeredDevice).(ttnpb.EndDevice),
				FieldMask: pbtypes.FieldMask{
					Paths: []string{"ids"},
				},
			},
			SetByEUIFunc: func(ctx context.Context, joinEUI, devEUI types.EUI64, paths []string, cb func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				defer test.MustIncrementContextCounter(ctx, setByEUIFuncKey{}, 1)
				a.So(joinEUI, should.Equal, *registeredJoinEUI)
				a.So(devEUI, should.Equal, *registeredDevEUI)
				a.So(paths, should.Contain, "ids")
				dev, _, err := cb(deepcopy.Copy(registeredDevice).(*ttnpb.EndDevice))
				return dev, err
			},
			ContextAssertion: func(ctx context.Context) bool {
				a := assertions.New(test.MustTFromContext(ctx))
				return a.So(test.MustCounterFromContext(ctx, setByEUIFuncKey{}), should.Equal, 1)
			},
		},
		{
			Name: "Set keys, permission denied",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, deepcopy.Copy(registeredDevice.EndDeviceIdentifiers.ApplicationIdentifiers).(ttnpb.ApplicationIdentifiers)): ttnpb.RightsFrom(
							ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
						),
					},
				})
			},
			DeviceRequest: &ttnpb.SetEndDeviceRequest{
				Device: deepcopy.Copy(*registeredDevice).(ttnpb.EndDevice),
				FieldMask: pbtypes.FieldMask{
					Paths: []string{"ids", "root_keys"},
				},
			},
			SetByEUIFunc: func(ctx context.Context, joinEUI, devEUI types.EUI64, paths []string, cb func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				defer test.MustIncrementContextCounter(ctx, setByEUIFuncKey{}, 1)
				a.So(joinEUI, should.Equal, *registeredJoinEUI)
				a.So(devEUI, should.Equal, *registeredDevEUI)
				a.So(paths, should.Contain, "ids")
				dev, _, err := cb(deepcopy.Copy(registeredDevice).(*ttnpb.EndDevice))
				return dev, err
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				a := assertions.New(t)
				return a.So(errors.IsPermissionDenied(err), should.BeTrue)
			},
			ContextAssertion: func(ctx context.Context) bool {
				a := assertions.New(test.MustTFromContext(ctx))
				return a.So(test.MustCounterFromContext(ctx, setByEUIFuncKey{}), should.Equal, 0)
			},
		},
		{
			Name: "Set keys",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, deepcopy.Copy(registeredDevice.EndDeviceIdentifiers.ApplicationIdentifiers).(ttnpb.ApplicationIdentifiers)): ttnpb.RightsFrom(
							ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
							ttnpb.RIGHT_APPLICATION_DEVICES_WRITE_KEYS,
						),
					},
				})
			},
			DeviceRequest: &ttnpb.SetEndDeviceRequest{
				Device: deepcopy.Copy(*registeredDevice).(ttnpb.EndDevice),
				FieldMask: pbtypes.FieldMask{
					Paths: []string{"ids", "root_keys"},
				},
			},
			SetByEUIFunc: func(ctx context.Context, joinEUI, devEUI types.EUI64, paths []string, cb func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				defer test.MustIncrementContextCounter(ctx, setByEUIFuncKey{}, 1)
				a.So(joinEUI, should.Equal, *registeredJoinEUI)
				a.So(devEUI, should.Equal, *registeredDevEUI)
				a.So(paths, should.Contain, "ids")
				dev, _, err := cb(deepcopy.Copy(registeredDevice).(*ttnpb.EndDevice))
				return dev, err
			},
			ContextAssertion: func(ctx context.Context) bool {
				a := assertions.New(test.MustTFromContext(ctx))
				return a.So(test.MustCounterFromContext(ctx, setByEUIFuncKey{}), should.Equal, 1)
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)
			ctx := test.ContextWithCounter(tc.ContextFunc(test.ContextWithT(test.Context(), t)), setByEUIFuncKey{})
			reg := &MockDeviceRegistry{
				SetByEUIFunc: tc.SetByEUIFunc,
			}
			srv := &JsDeviceServer{
				Registry: reg,
			}
			dev, err := srv.Set(ctx, tc.DeviceRequest)
			a.So(tc.ContextAssertion(ctx), should.BeTrue)
			if tc.ErrorAssertion != nil {
				a.So(tc.ErrorAssertion(t, err), should.BeTrue)
				a.So(dev, should.BeNil)
				return
			}
			a.So(err, should.BeNil)
			if tc.DeviceAssertion != nil {
				a.So(tc.DeviceAssertion(t, dev), should.BeTrue)
				return
			}
			a.So(dev, should.Resemble, registeredDevice)
		})
	}
}

func TestDeviceRegistryDelete(t *testing.T) {
	registeredDevice := &ttnpb.EndDevice{
		EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
			ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
				ApplicationID: registeredApplicationID,
			},
			DeviceID: registeredDeviceID,
			JoinEUI:  registeredJoinEUI,
			DevEUI:   registeredDevEUI,
		},
		RootKeys: &ttnpb.RootKeys{
			RootKeyID: registeredRootKeyID,
			NwkKey: &ttnpb.KeyEnvelope{
				KEKLabel: registeredKEKLabel,
				Key:      registeredNwkKey,
			},
			AppKey: &ttnpb.KeyEnvelope{
				KEKLabel: registeredKEKLabel,
				Key:      registeredAppKey,
			},
		},
	}
	for _, tc := range []struct {
		Name             string
		ContextFunc      func(context.Context) context.Context
		SetByEUIFunc     func(ctx context.Context, joinEUI, devEUI types.EUI64, paths []string, cb func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error)
		Device           *ttnpb.EndDeviceIdentifiers
		ErrorAssertion   func(*testing.T, error) bool
		ContextAssertion func(context.Context) bool
	}{
		{
			Name: "Permission denied",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, ttnpb.ApplicationIdentifiers{
							ApplicationID: registeredApplicationID,
						}): nil,
					},
				})
			},
			Device: deepcopy.Copy(&registeredDevice.EndDeviceIdentifiers).(*ttnpb.EndDeviceIdentifiers),
			SetByEUIFunc: func(ctx context.Context, joinEUI, devEUI types.EUI64, paths []string, cb func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				defer test.MustIncrementContextCounter(ctx, setByEUIFuncKey{}, 1)
				a.So(joinEUI, should.Equal, *registeredJoinEUI)
				a.So(devEUI, should.Equal, *registeredDevEUI)
				a.So(paths, should.BeNil)
				dev, _, err := cb(deepcopy.Copy(registeredDevice).(*ttnpb.EndDevice))
				return dev, err
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				a := assertions.New(t)
				return a.So(errors.IsPermissionDenied(err), should.BeTrue)
			},
			ContextAssertion: func(ctx context.Context) bool {
				a := assertions.New(test.MustTFromContext(ctx))
				return a.So(test.MustCounterFromContext(ctx, setByEUIFuncKey{}), should.Equal, 0)
			},
		},
		{
			Name: "No DevEUI",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, deepcopy.Copy(registeredDevice.EndDeviceIdentifiers.ApplicationIdentifiers).(ttnpb.ApplicationIdentifiers)): ttnpb.RightsFrom(
							ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
						),
					},
				})
			},
			Device: &ttnpb.EndDeviceIdentifiers{
				ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
					ApplicationID: registeredApplicationID,
				},
				DeviceID: registeredDeviceID,
				JoinEUI:  registeredJoinEUI,
			},
			SetByEUIFunc: func(ctx context.Context, joinEUI, devEUI types.EUI64, paths []string, cb func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				defer test.MustIncrementContextCounter(ctx, setByEUIFuncKey{}, 1)
				a.So(joinEUI, should.Equal, *registeredJoinEUI)
				a.So(devEUI, should.Equal, *registeredDevEUI)
				a.So(paths, should.BeNil)
				dev, _, err := cb(deepcopy.Copy(registeredDevice).(*ttnpb.EndDevice))
				return dev, err
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				a := assertions.New(t)
				return a.So(errors.IsInvalidArgument(err), should.BeTrue)
			},
			ContextAssertion: func(ctx context.Context) bool {
				a := assertions.New(test.MustTFromContext(ctx))
				return a.So(test.MustCounterFromContext(ctx, setByEUIFuncKey{}), should.Equal, 0)
			},
		},
		{
			Name: "No JoinEUI",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, deepcopy.Copy(registeredDevice.EndDeviceIdentifiers.ApplicationIdentifiers).(ttnpb.ApplicationIdentifiers)): ttnpb.RightsFrom(
							ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
						),
					},
				})
			},
			Device: &ttnpb.EndDeviceIdentifiers{
				ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
					ApplicationID: registeredApplicationID,
				},
				DeviceID: registeredDeviceID,
				DevEUI:   registeredDevEUI,
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				a := assertions.New(t)
				return a.So(errors.IsInvalidArgument(err), should.BeTrue)
			},
			ContextAssertion: func(ctx context.Context) bool {
				a := assertions.New(test.MustTFromContext(ctx))
				return a.So(test.MustCounterFromContext(ctx, setByEUIFuncKey{}), should.Equal, 0)
			},
		},
		{
			Name: "Not found",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, deepcopy.Copy(registeredDevice.EndDeviceIdentifiers.ApplicationIdentifiers).(ttnpb.ApplicationIdentifiers)): ttnpb.RightsFrom(
							ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
						),
					},
				})
			},
			Device: &ttnpb.EndDeviceIdentifiers{
				ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
					ApplicationID: registeredApplicationID,
				},
				DeviceID: unregisteredDeviceID,
				JoinEUI:  unregisteredJoinEUI,
				DevEUI:   unregisteredDevEUI,
			},
			SetByEUIFunc: func(ctx context.Context, joinEUI, devEUI types.EUI64, paths []string, cb func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				defer test.MustIncrementContextCounter(ctx, setByEUIFuncKey{}, 1)
				a.So(joinEUI, should.Equal, *unregisteredJoinEUI)
				a.So(devEUI, should.Equal, *unregisteredDevEUI)
				a.So(paths, should.BeNil)
				return nil, errNotFound
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				a := assertions.New(t)
				return a.So(errors.IsNotFound(err), should.BeTrue)
			},
			ContextAssertion: func(ctx context.Context) bool {
				a := assertions.New(test.MustTFromContext(ctx))
				return a.So(test.MustCounterFromContext(ctx, setByEUIFuncKey{}), should.Equal, 1)
			},
		},
		{
			Name: "Invalid application ID",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, ttnpb.ApplicationIdentifiers{ApplicationID: "other-app"}): ttnpb.RightsFrom(
							ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
						),
					},
				})
			},
			Device: &ttnpb.EndDeviceIdentifiers{
				ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
					ApplicationID: "other-app",
				},
				DeviceID: "other-device",
				JoinEUI:  registeredJoinEUI,
				DevEUI:   registeredDevEUI,
			},
			SetByEUIFunc: func(ctx context.Context, joinEUI, devEUI types.EUI64, paths []string, cb func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				defer test.MustIncrementContextCounter(ctx, setByEUIFuncKey{}, 1)
				a.So(joinEUI, should.Equal, *registeredJoinEUI)
				a.So(devEUI, should.Equal, *registeredDevEUI)
				a.So(paths, should.BeNil)
				dev, _, err := cb(deepcopy.Copy(registeredDevice).(*ttnpb.EndDevice))
				return dev, err
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				a := assertions.New(t)
				return a.So(errors.IsNotFound(err), should.BeTrue)
			},
			ContextAssertion: func(ctx context.Context) bool {
				a := assertions.New(test.MustTFromContext(ctx))
				return a.So(test.MustCounterFromContext(ctx, setByEUIFuncKey{}), should.Equal, 1)
			},
		},
		{
			Name: "Delete",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, deepcopy.Copy(registeredDevice.EndDeviceIdentifiers.ApplicationIdentifiers).(ttnpb.ApplicationIdentifiers)): ttnpb.RightsFrom(
							ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
						),
					},
				})
			},
			Device: deepcopy.Copy(&registeredDevice.EndDeviceIdentifiers).(*ttnpb.EndDeviceIdentifiers),
			SetByEUIFunc: func(ctx context.Context, joinEUI, devEUI types.EUI64, paths []string, cb func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				defer test.MustIncrementContextCounter(ctx, setByEUIFuncKey{}, 1)
				a.So(joinEUI, should.Equal, *registeredJoinEUI)
				a.So(devEUI, should.Equal, *registeredDevEUI)
				a.So(paths, should.BeNil)
				dev, _, err := cb(deepcopy.Copy(registeredDevice).(*ttnpb.EndDevice))
				return dev, err
			},
			ContextAssertion: func(ctx context.Context) bool {
				a := assertions.New(test.MustTFromContext(ctx))
				return a.So(test.MustCounterFromContext(ctx, setByEUIFuncKey{}), should.Equal, 1)
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)
			ctx := test.ContextWithCounter(tc.ContextFunc(test.ContextWithT(test.Context(), t)), setByEUIFuncKey{})
			reg := &MockDeviceRegistry{
				SetByEUIFunc: tc.SetByEUIFunc,
			}
			srv := &JsDeviceServer{
				Registry: reg,
			}
			dev, err := srv.Delete(ctx, tc.Device)
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
