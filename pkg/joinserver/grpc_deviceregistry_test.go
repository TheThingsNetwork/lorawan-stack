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
	"io"
	"sync/atomic"
	"testing"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/mohae/deepcopy"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/crypto/cryptoutil"
	"go.thethings.network/lorawan-stack/pkg/errors"
	. "go.thethings.network/lorawan-stack/pkg/joinserver"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/unique"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

var errNotFound = errors.DefineNotFound("not_found", "not found")

func eui64Ptr(eui types.EUI64) *types.EUI64 { return &eui }

func TestDeviceRegistryGet(t *testing.T) {
	registeredApplicationID := "foo-application"
	registeredDeviceID := "foo-device"
	registeredRootKeyID := "testKey"
	registeredKEKLabel := "test"
	unregisteredDeviceID := "bar-device"
	registeredJoinEUI := eui64Ptr(types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
	registeredDevEUI := eui64Ptr(types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
	registeredNwkKey := &types.AES128Key{0xff, 0xee, 0xdd, 0xcc, 0xbb, 0xaa, 0x99, 0x88, 0x77, 0x66, 0x55, 0x44, 0x33, 0x22, 0x11, 0x0}
	registeredNwkKeyEnc := []byte{0xa3, 0x34, 0x38, 0x1c, 0xca, 0x1c, 0x12, 0x7a, 0x5b, 0xb1, 0xa8, 0x97, 0x39, 0xc7, 0x5, 0x34, 0x91, 0x26, 0x9b, 0x21, 0x4f, 0x27, 0x80, 0x19}
	registeredAppKey := &types.AES128Key{0x0, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}
	registeredAppKeyEnc := []byte{0x1f, 0xa6, 0x8b, 0xa, 0x81, 0x12, 0xb4, 0x47, 0xae, 0xf3, 0x4b, 0xd8, 0xfb, 0x5a, 0x7b, 0x82, 0x9d, 0x3e, 0x86, 0x23, 0x71, 0xd2, 0xcf, 0xe5}
	unregisteredJoinEUI := eui64Ptr(types.EUI64{0x42, 0x42, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	unregisteredDevEUI := eui64Ptr(types.EUI64{0x42, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	keyVault := cryptoutil.NewMemKeyVault(map[string][]byte{
		"test": {0x0, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf},
	})
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
				Key:          registeredNwkKey,
				KEKLabel:     registeredKEKLabel,
				EncryptedKey: registeredNwkKeyEnc,
			},
			AppKey: &ttnpb.KeyEnvelope{
				Key:          registeredAppKey,
				KEKLabel:     registeredKEKLabel,
				EncryptedKey: registeredAppKeyEnc,
			},
		},
	}
	for _, tc := range []struct {
		Name            string
		ContextFunc     func(context.Context) context.Context
		GetByIDFunc     func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string) (*ttnpb.EndDevice, error)
		DeviceRequest   *ttnpb.GetEndDeviceRequest
		ErrorAssertion  func(*testing.T, error) bool
		DeviceAssertion func(*testing.T, *ttnpb.EndDevice) bool
		GetByIDCalls    uint64
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
			GetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string) (*ttnpb.EndDevice, error) {
				test.MustTFromContext(ctx).Errorf("GetByIDFunc must not be called")
				return nil, errors.New("GetByIDFunc must not be called")
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
		},

		{
			Name: "Not found",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, ttnpb.ApplicationIdentifiers{ApplicationID: registeredApplicationID}): ttnpb.RightsFrom(
							ttnpb.RIGHT_APPLICATION_DEVICES_READ,
						),
					},
				})
			},
			GetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(appID, should.Resemble, ttnpb.ApplicationIdentifiers{
					ApplicationID: registeredApplicationID,
				})
				a.So(devID, should.Equal, unregisteredDeviceID)
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
			GetByIDCalls: 1,
		},

		{
			Name: "Invalid application ID",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, ttnpb.ApplicationIdentifiers{ApplicationID: "bar-application"}): ttnpb.RightsFrom(
							ttnpb.RIGHT_APPLICATION_DEVICES_READ,
						),
					},
				})
			},
			GetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(appID, should.Resemble, ttnpb.ApplicationIdentifiers{
					ApplicationID: "bar-application",
				})
				a.So(devID, should.Equal, "bar-device")
				a.So(paths, should.HaveSameElementsDeep, []string{"ids"})
				return deepcopy.Copy(registeredDevice).(*ttnpb.EndDevice), nil
			},
			DeviceRequest: &ttnpb.GetEndDeviceRequest{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
						ApplicationID: "bar-application",
					},
					DeviceID: "bar-device",
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
			GetByIDCalls: 1,
		},

		{
			Name: "Get without key",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, ttnpb.ApplicationIdentifiers{ApplicationID: registeredApplicationID}): ttnpb.RightsFrom(
							ttnpb.RIGHT_APPLICATION_DEVICES_READ,
						),
					},
				})
			},
			GetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(appID, should.Resemble, ttnpb.ApplicationIdentifiers{
					ApplicationID: registeredApplicationID,
				})
				a.So(devID, should.Equal, registeredDeviceID)
				a.So(paths, should.HaveSameElementsDeep, []string{"ids"})
				return deepcopy.Copy(registeredDevice).(*ttnpb.EndDevice), nil
			},
			DeviceRequest: &ttnpb.GetEndDeviceRequest{
				EndDeviceIdentifiers: deepcopy.Copy(registeredDevice.EndDeviceIdentifiers).(ttnpb.EndDeviceIdentifiers),
				FieldMask: pbtypes.FieldMask{
					Paths: []string{"ids"},
				},
			},
			GetByIDCalls: 1,
		},

		{
			Name: "Get keys without permission",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, ttnpb.ApplicationIdentifiers{ApplicationID: registeredApplicationID}): ttnpb.RightsFrom(
							ttnpb.RIGHT_APPLICATION_DEVICES_READ,
						),
					},
				})
			},
			GetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string) (*ttnpb.EndDevice, error) {
				test.MustTFromContext(ctx).Errorf("GetByIDFunc must not be called")
				return nil, errors.New("GetByIDFunc must not be called")
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
		},

		{
			Name: "Get keys",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, ttnpb.ApplicationIdentifiers{ApplicationID: registeredApplicationID}): ttnpb.RightsFrom(
							ttnpb.RIGHT_APPLICATION_DEVICES_READ,
							ttnpb.RIGHT_APPLICATION_DEVICES_READ_KEYS,
						),
					},
				})
			},
			GetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(appID, should.Resemble, ttnpb.ApplicationIdentifiers{
					ApplicationID: registeredApplicationID,
				})
				a.So(devID, should.Equal, registeredDeviceID)
				a.So(paths, should.HaveSameElementsDeep, []string{"ids", "root_keys", "provisioner_id", "provisioning_data"})
				return deepcopy.Copy(registeredDevice).(*ttnpb.EndDevice), nil
			},
			DeviceRequest: &ttnpb.GetEndDeviceRequest{
				EndDeviceIdentifiers: deepcopy.Copy(registeredDevice.EndDeviceIdentifiers).(ttnpb.EndDeviceIdentifiers),
				FieldMask: pbtypes.FieldMask{
					Paths: []string{"ids", "root_keys"},
				},
			},
			DeviceAssertion: func(t *testing.T, dev *ttnpb.EndDevice) bool {
				a := assertions.New(t)
				expected := deepcopy.Copy(registeredDevice).(*ttnpb.EndDevice)
				expected.RootKeys = &ttnpb.RootKeys{
					RootKeyID: registeredDevice.RootKeys.RootKeyID,
					NwkKey: &ttnpb.KeyEnvelope{
						Key: registeredNwkKey,
					},
					AppKey: &ttnpb.KeyEnvelope{
						Key: registeredAppKey,
					},
				}
				return a.So(dev, should.Resemble, expected)
			},
			GetByIDCalls: 1,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			var getByIDCalls uint64

			js := test.Must(New(
				component.MustNew(test.GetLogger(t), &component.Config{}),
				&Config{
					Devices: &MockDeviceRegistry{
						GetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string) (*ttnpb.EndDevice, error) {
							atomic.AddUint64(&getByIDCalls, 1)
							return tc.GetByIDFunc(ctx, appID, devID, paths)
						},
					},
				},
			)).(*JoinServer)
			js.KeyVault = keyVault

			js.AddContextFiller(tc.ContextFunc)
			js.AddContextFiller(func(ctx context.Context) context.Context {
				ctx, cancel := context.WithDeadline(ctx, time.Now().Add(Timeout))
				_ = cancel
				return ctx
			})
			js.AddContextFiller(func(ctx context.Context) context.Context {
				return test.ContextWithT(ctx, t)
			})
			test.Must(nil, js.Start())
			defer js.Close()

			ctx := js.FillContext(test.Context())
			req := deepcopy.Copy(tc.DeviceRequest).(*ttnpb.GetEndDeviceRequest)

			dev, err := ttnpb.NewJsEndDeviceRegistryClient(js.LoopbackConn()).Get(ctx, req)
			a.So(getByIDCalls, should.Equal, tc.GetByIDCalls)
			if tc.ErrorAssertion != nil && a.So(tc.ErrorAssertion(t, err), should.BeTrue) {
				a.So(dev, should.BeNil)
			} else if a.So(err, should.BeNil) {
				if tc.DeviceAssertion != nil {
					a.So(tc.DeviceAssertion(t, dev), should.BeTrue)
				} else {
					a.So(dev, should.Resemble, registeredDevice)
				}
			}
			a.So(req, should.Resemble, tc.DeviceRequest)
		})
	}
}

func TestDeviceRegistrySet(t *testing.T) {
	registeredApplicationID := "foo-application"
	registeredDeviceID := "foo-device"
	registeredRootKeyID := "testKey"
	registeredKEKLabel := "test"
	registeredJoinEUI := eui64Ptr(types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
	registeredDevEUI := eui64Ptr(types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
	registeredNwkKey := &types.AES128Key{0xff, 0xee, 0xdd, 0xcc, 0xbb, 0xaa, 0x99, 0x88, 0x77, 0x66, 0x55, 0x44, 0x33, 0x22, 0x11, 0x0}
	registeredAppKey := &types.AES128Key{0x0, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}
	unregisteredJoinEUI := eui64Ptr(types.EUI64{0x42, 0x42, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	unregisteredDevEUI := eui64Ptr(types.EUI64{0x42, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	keyVault := cryptoutil.NewMemKeyVault(map[string][]byte{
		"test": {0x0, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf},
	})
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
				Key:      registeredNwkKey,
				KEKLabel: registeredKEKLabel,
			},
			AppKey: &ttnpb.KeyEnvelope{
				Key:      registeredAppKey,
				KEKLabel: registeredKEKLabel,
			},
		},
	}
	for _, tc := range []struct {
		Name            string
		ContextFunc     func(context.Context) context.Context
		SetByIDFunc     func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string, cb func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error)
		DeviceRequest   *ttnpb.SetEndDeviceRequest
		ErrorAssertion  func(*testing.T, error) bool
		DeviceAssertion func(*testing.T, *ttnpb.EndDevice) bool
		SetByIDCalls    uint64
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
			DeviceRequest: &ttnpb.SetEndDeviceRequest{
				EndDevice: deepcopy.Copy(*registeredDevice).(ttnpb.EndDevice),
				FieldMask: pbtypes.FieldMask{
					Paths: []string{"ids"},
				},
			},
			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string, cb func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				test.MustTFromContext(ctx).Errorf("SetByIDFunc must not be called")
				return nil, errors.New("SetByIDFunc must not be called")
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				a := assertions.New(t)
				return a.So(errors.IsPermissionDenied(err), should.BeTrue)
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
				EndDevice: ttnpb.EndDevice{
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
			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string, cb func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				test.MustTFromContext(ctx).Errorf("SetByIDFunc must not be called")
				return nil, errors.New("SetByIDFunc must not be called")
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				a := assertions.New(t)
				return a.So(errors.IsInvalidArgument(err), should.BeTrue)
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
				EndDevice: ttnpb.EndDevice{
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
			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string, cb func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				test.MustTFromContext(ctx).Errorf("SetByIDFunc must not be called")
				return nil, errors.New("SetByIDFunc must not be called")
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				a := assertions.New(t)
				return a.So(errors.IsInvalidArgument(err), should.BeTrue)
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
				EndDevice: ttnpb.EndDevice{
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
			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string, cb func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(appID, should.Resemble, ttnpb.ApplicationIdentifiers{
					ApplicationID: registeredApplicationID,
				})
				a.So(devID, should.Equal, "new-device")
				a.So(paths, should.HaveSameElementsDeep, []string{"ids"})
				dev, _, err := cb(nil)
				return dev, err
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
			SetByIDCalls: 1,
		},

		{
			Name: "Set without keys",
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
				EndDevice: deepcopy.Copy(*registeredDevice).(ttnpb.EndDevice),
				FieldMask: pbtypes.FieldMask{
					Paths: []string{"ids"},
				},
			},
			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string, cb func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(appID, should.Resemble, ttnpb.ApplicationIdentifiers{
					ApplicationID: registeredApplicationID,
				})
				a.So(devID, should.Equal, registeredDeviceID)
				a.So(paths, should.HaveSameElementsDeep, []string{"ids"})
				dev, _, err := cb(deepcopy.Copy(registeredDevice).(*ttnpb.EndDevice))
				return dev, err
			},
			SetByIDCalls: 1,
		},

		{
			Name: "Set keys without permission",
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
				EndDevice: deepcopy.Copy(*registeredDevice).(ttnpb.EndDevice),
				FieldMask: pbtypes.FieldMask{
					Paths: []string{"ids", "root_keys"},
				},
			},
			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string, cb func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				test.MustTFromContext(ctx).Errorf("SetByIDFunc must not be called")
				return nil, errors.New("SetByIDFunc must not be called")
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				a := assertions.New(t)
				return a.So(errors.IsPermissionDenied(err), should.BeTrue)
			},
		},

		{
			Name: "Set keys",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, ttnpb.ApplicationIdentifiers{ApplicationID: registeredApplicationID}): ttnpb.RightsFrom(
							ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
							ttnpb.RIGHT_APPLICATION_DEVICES_WRITE_KEYS,
						),
					},
				})
			},
			DeviceRequest: &ttnpb.SetEndDeviceRequest{
				EndDevice: deepcopy.Copy(*registeredDevice).(ttnpb.EndDevice),
				FieldMask: pbtypes.FieldMask{
					Paths: []string{"ids", "root_keys"},
				},
			},
			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string, cb func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(appID, should.Resemble, ttnpb.ApplicationIdentifiers{
					ApplicationID: registeredApplicationID,
				})
				a.So(devID, should.Equal, registeredDeviceID)
				a.So(paths, should.HaveSameElementsDeep, []string{"ids", "root_keys"})
				dev, _, err := cb(deepcopy.Copy(registeredDevice).(*ttnpb.EndDevice))
				return dev, err
			},
			SetByIDCalls: 1,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			var setByIDCalls uint64

			js := test.Must(New(
				component.MustNew(test.GetLogger(t), &component.Config{}),
				&Config{
					Devices: &MockDeviceRegistry{
						SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string, cb func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
							atomic.AddUint64(&setByIDCalls, 1)
							return tc.SetByIDFunc(ctx, appID, devID, paths, cb)
						},
					},
				},
			)).(*JoinServer)
			js.KeyVault = keyVault

			js.AddContextFiller(tc.ContextFunc)
			js.AddContextFiller(func(ctx context.Context) context.Context {
				ctx, cancel := context.WithDeadline(ctx, time.Now().Add(Timeout))
				_ = cancel
				return ctx
			})
			js.AddContextFiller(func(ctx context.Context) context.Context {
				return test.ContextWithT(ctx, t)
			})
			test.Must(nil, js.Start())
			defer js.Close()

			ctx := js.FillContext(test.Context())
			req := deepcopy.Copy(tc.DeviceRequest).(*ttnpb.SetEndDeviceRequest)

			dev, err := ttnpb.NewJsEndDeviceRegistryClient(js.LoopbackConn()).Set(ctx, req)
			a.So(setByIDCalls, should.Equal, tc.SetByIDCalls)
			if tc.ErrorAssertion != nil && a.So(tc.ErrorAssertion(t, err), should.BeTrue) {
				a.So(dev, should.BeNil)
			} else if a.So(err, should.BeNil) {
				if tc.DeviceAssertion != nil {
					a.So(tc.DeviceAssertion(t, dev), should.BeTrue)
				} else {
					a.So(dev, should.Resemble, registeredDevice)
				}
			}
			a.So(req, should.Resemble, tc.DeviceRequest)
		})
	}
}

func TestDeviceRegistryDelete(t *testing.T) {
	registeredApplicationID := "foo-application"
	registeredDeviceID := "foo-device"
	registeredRootKeyID := "testKey"
	registeredKEKLabel := "test"
	unregisteredDeviceID := "bar-device"
	registeredJoinEUI := eui64Ptr(types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
	registeredDevEUI := eui64Ptr(types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
	registeredNwkKey := &types.AES128Key{0xff, 0xee, 0xdd, 0xcc, 0xbb, 0xaa, 0x99, 0x88, 0x77, 0x66, 0x55, 0x44, 0x33, 0x22, 0x11, 0x0}
	registeredAppKey := &types.AES128Key{0x0, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}
	unregisteredJoinEUI := eui64Ptr(types.EUI64{0x42, 0x42, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	unregisteredDevEUI := eui64Ptr(types.EUI64{0x42, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	keyVault := cryptoutil.NewMemKeyVault(map[string][]byte{
		"test": {0x0, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf},
	})
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
				Key:      registeredNwkKey,
				KEKLabel: registeredKEKLabel,
			},
			AppKey: &ttnpb.KeyEnvelope{
				Key:      registeredAppKey,
				KEKLabel: registeredKEKLabel,
			},
		},
	}
	for _, tc := range []struct {
		Name           string
		ContextFunc    func(context.Context) context.Context
		SetByIDFunc    func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string, cb func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error)
		DeviceRequest  *ttnpb.EndDeviceIdentifiers
		ErrorAssertion func(*testing.T, error) bool
		SetByIDCalls   uint64
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
			DeviceRequest: deepcopy.Copy(&registeredDevice.EndDeviceIdentifiers).(*ttnpb.EndDeviceIdentifiers),
			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string, cb func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				test.MustTFromContext(ctx).Errorf("SetByIDFunc must not be called")
				return nil, errors.New("SetByIDFunc must not be called")
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				a := assertions.New(t)
				return a.So(errors.IsPermissionDenied(err), should.BeTrue)
			},
		},

		{
			Name: "Invalid application ID",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, ttnpb.ApplicationIdentifiers{ApplicationID: "bar-application"}): ttnpb.RightsFrom(
							ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
						),
					},
				})
			},
			DeviceRequest: &ttnpb.EndDeviceIdentifiers{
				ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
					ApplicationID: "bar-application",
				},
				DeviceID: "bar-device",
				JoinEUI:  registeredJoinEUI,
				DevEUI:   registeredDevEUI,
			},
			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string, cb func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(appID, should.Resemble, ttnpb.ApplicationIdentifiers{
					ApplicationID: "bar-application",
				})
				a.So(devID, should.Equal, "bar-device")
				a.So(paths, should.BeNil)
				dev, _, err := cb(deepcopy.Copy(registeredDevice).(*ttnpb.EndDevice))
				return dev, err
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				a := assertions.New(t)
				return a.So(errors.IsNotFound(err), should.BeTrue)
			},
			SetByIDCalls: 1,
		},

		{
			Name: "Not found",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, ttnpb.ApplicationIdentifiers{ApplicationID: registeredApplicationID}): ttnpb.RightsFrom(
							ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
						),
					},
				})
			},
			DeviceRequest: &ttnpb.EndDeviceIdentifiers{
				ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
					ApplicationID: registeredApplicationID,
				},
				DeviceID: unregisteredDeviceID,
				JoinEUI:  unregisteredJoinEUI,
				DevEUI:   unregisteredDevEUI,
			},
			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string, cb func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(appID, should.Resemble, ttnpb.ApplicationIdentifiers{
					ApplicationID: registeredApplicationID,
				})
				a.So(devID, should.Equal, unregisteredDeviceID)
				a.So(paths, should.BeNil)
				return nil, errNotFound
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				a := assertions.New(t)
				return a.So(errors.IsNotFound(err), should.BeTrue)
			},
			SetByIDCalls: 1,
		},

		{
			Name: "Delete",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, ttnpb.ApplicationIdentifiers{ApplicationID: registeredApplicationID}): ttnpb.RightsFrom(
							ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
						),
					},
				})
			},
			DeviceRequest: deepcopy.Copy(&registeredDevice.EndDeviceIdentifiers).(*ttnpb.EndDeviceIdentifiers),
			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string, cb func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(appID, should.Resemble, ttnpb.ApplicationIdentifiers{
					ApplicationID: registeredApplicationID,
				})
				a.So(devID, should.Equal, registeredDeviceID)
				a.So(paths, should.BeNil)
				dev, _, err := cb(deepcopy.Copy(registeredDevice).(*ttnpb.EndDevice))
				return dev, err
			},
			SetByIDCalls: 1,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			var setByIDCalls uint64

			js := test.Must(New(
				component.MustNew(test.GetLogger(t), &component.Config{}),
				&Config{
					Devices: &MockDeviceRegistry{
						SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string, cb func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
							atomic.AddUint64(&setByIDCalls, 1)
							return tc.SetByIDFunc(ctx, appID, devID, paths, cb)
						},
					},
				},
			)).(*JoinServer)
			js.KeyVault = keyVault

			js.AddContextFiller(tc.ContextFunc)
			js.AddContextFiller(func(ctx context.Context) context.Context {
				ctx, cancel := context.WithDeadline(ctx, time.Now().Add(Timeout))
				_ = cancel
				return ctx
			})
			js.AddContextFiller(func(ctx context.Context) context.Context {
				return test.ContextWithT(ctx, t)
			})
			test.Must(nil, js.Start())
			defer js.Close()

			ctx := js.FillContext(test.Context())
			req := deepcopy.Copy(tc.DeviceRequest).(*ttnpb.EndDeviceIdentifiers)

			_, err := ttnpb.NewJsEndDeviceRegistryClient(js.LoopbackConn()).Delete(ctx, req)
			a.So(setByIDCalls, should.Equal, tc.SetByIDCalls)
			if tc.ErrorAssertion != nil {
				a.So(tc.ErrorAssertion(t, err), should.BeTrue)
			} else {
				a.So(err, should.BeNil)
			}
			a.So(req, should.Resemble, tc.DeviceRequest)
		})
	}
}

func TestDeviceRegistryProvision(t *testing.T) {
	registeredApplicationID := "foo-application"
	for _, tc := range []struct {
		Name              string
		ContextFunc       func(context.Context) context.Context
		ProvisionRequest  *ttnpb.ProvisionEndDevicesRequest
		ErrorAssertion    func(*testing.T, error) bool
		SingleAssertion   func(*testing.T, *ttnpb.EndDevice) bool
		MultipleAssertion func(*testing.T, []*ttnpb.EndDevice) bool
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
			ProvisionRequest: &ttnpb.ProvisionEndDevicesRequest{
				ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
					ApplicationID: registeredApplicationID,
				},
				EndDevices: &ttnpb.ProvisionEndDevicesRequest_List{
					List: &ttnpb.ProvisionEndDevicesRequest_IdentifiersList{},
				},
				ProvisionerID:    "mock",
				ProvisioningData: []byte{0x1},
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				a := assertions.New(t)
				return a.So(errors.IsPermissionDenied(err), should.BeTrue)
			},
		},

		{
			Name: "Unknown provisioner",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, ttnpb.ApplicationIdentifiers{ApplicationID: registeredApplicationID}): ttnpb.RightsFrom(
							ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
							ttnpb.RIGHT_APPLICATION_DEVICES_WRITE_KEYS,
						),
					},
				})
			},
			ProvisionRequest: &ttnpb.ProvisionEndDevicesRequest{
				ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
					ApplicationID: registeredApplicationID,
				},
				EndDevices: &ttnpb.ProvisionEndDevicesRequest_List{
					List: &ttnpb.ProvisionEndDevicesRequest_IdentifiersList{},
				},
				ProvisionerID:    "unknown",
				ProvisioningData: []byte{0x1},
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				a := assertions.New(t)
				return a.So(errors.IsNotFound(err), should.BeTrue)
			},
		},

		{
			Name: "Range without StartDevEUI",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, ttnpb.ApplicationIdentifiers{ApplicationID: registeredApplicationID}): ttnpb.RightsFrom(
							ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
							ttnpb.RIGHT_APPLICATION_DEVICES_WRITE_KEYS,
						),
					},
				})
			},
			ProvisionRequest: &ttnpb.ProvisionEndDevicesRequest{
				ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
					ApplicationID: registeredApplicationID,
				},
				EndDevices: &ttnpb.ProvisionEndDevicesRequest_Range{
					Range: &ttnpb.ProvisionEndDevicesRequest_IdentifiersRange{
						JoinEUI: eui64Ptr(types.EUI64{0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}),
					},
				},
				ProvisionerID:    "mock",
				ProvisioningData: []byte{0x1},
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				a := assertions.New(t)
				return a.So(errors.IsInvalidArgument(err), should.BeTrue)
			},
		},

		{
			Name: "List with wrong application ID",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, ttnpb.ApplicationIdentifiers{ApplicationID: registeredApplicationID}): ttnpb.RightsFrom(
							ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
							ttnpb.RIGHT_APPLICATION_DEVICES_WRITE_KEYS,
						),
					},
				})
			},
			ProvisionRequest: &ttnpb.ProvisionEndDevicesRequest{
				ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
					ApplicationID: registeredApplicationID,
				},
				EndDevices: &ttnpb.ProvisionEndDevicesRequest_List{
					List: &ttnpb.ProvisionEndDevicesRequest_IdentifiersList{
						EndDeviceIDs: []ttnpb.EndDeviceIdentifiers{
							{
								ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
									ApplicationID: "wrong-id",
								},
								DeviceID: "new-dev-1",
								JoinEUI:  eui64Ptr(types.EUI64{0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1}),
								DevEUI:   eui64Ptr(types.EUI64{0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1}),
							},
						},
					},
				},
				ProvisionerID:    "mock",
				ProvisioningData: []byte{0x1},
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				a := assertions.New(t)
				return a.So(errors.IsInvalidArgument(err), should.BeTrue)
			},
		},

		{
			Name: "List item without JoinEUI",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, ttnpb.ApplicationIdentifiers{ApplicationID: registeredApplicationID}): ttnpb.RightsFrom(
							ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
							ttnpb.RIGHT_APPLICATION_DEVICES_WRITE_KEYS,
						),
					},
				})
			},
			ProvisionRequest: &ttnpb.ProvisionEndDevicesRequest{
				ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
					ApplicationID: registeredApplicationID,
				},
				EndDevices: &ttnpb.ProvisionEndDevicesRequest_List{
					List: &ttnpb.ProvisionEndDevicesRequest_IdentifiersList{
						EndDeviceIDs: []ttnpb.EndDeviceIdentifiers{
							{
								ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
									ApplicationID: registeredApplicationID,
								},
								DeviceID: "new-dev-1",
								JoinEUI:  eui64Ptr(types.EUI64{}),
								DevEUI:   eui64Ptr(types.EUI64{0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1}),
							},
						},
					},
				},
				ProvisionerID:    "mock",
				ProvisioningData: []byte{0x1},
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				a := assertions.New(t)
				return a.So(errors.IsInvalidArgument(err), should.BeTrue)
			},
		},

		{
			Name: "List item without DevEUI",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, ttnpb.ApplicationIdentifiers{ApplicationID: registeredApplicationID}): ttnpb.RightsFrom(
							ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
							ttnpb.RIGHT_APPLICATION_DEVICES_WRITE_KEYS,
						),
					},
				})
			},
			ProvisionRequest: &ttnpb.ProvisionEndDevicesRequest{
				ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
					ApplicationID: registeredApplicationID,
				},
				EndDevices: &ttnpb.ProvisionEndDevicesRequest_List{
					List: &ttnpb.ProvisionEndDevicesRequest_IdentifiersList{
						EndDeviceIDs: []ttnpb.EndDeviceIdentifiers{
							{
								ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
									ApplicationID: registeredApplicationID,
								},
								DeviceID: "new-dev-1",
								JoinEUI:  eui64Ptr(types.EUI64{0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1}),
								DevEUI:   eui64Ptr(types.EUI64{}),
							},
						},
					},
				},
				ProvisionerID:    "mock",
				ProvisioningData: []byte{0x1},
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				a := assertions.New(t)
				return a.So(errors.IsInvalidArgument(err), should.BeTrue)
			},
		},

		{
			Name: "List one device",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, ttnpb.ApplicationIdentifiers{ApplicationID: registeredApplicationID}): ttnpb.RightsFrom(
							ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
							ttnpb.RIGHT_APPLICATION_DEVICES_WRITE_KEYS,
						),
					},
				})
			},
			ProvisionRequest: &ttnpb.ProvisionEndDevicesRequest{
				ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
					ApplicationID: registeredApplicationID,
				},
				EndDevices: &ttnpb.ProvisionEndDevicesRequest_List{
					List: &ttnpb.ProvisionEndDevicesRequest_IdentifiersList{
						EndDeviceIDs: []ttnpb.EndDeviceIdentifiers{
							{
								ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
									ApplicationID: registeredApplicationID,
								},
								DeviceID: "new-dev-1",
								JoinEUI:  eui64Ptr(types.EUI64{0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1}),
								DevEUI:   eui64Ptr(types.EUI64{0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1}),
							},
						},
					},
				},
				ProvisionerID:    "mock",
				ProvisioningData: []byte{0x1},
			},
			SingleAssertion: func(t *testing.T, dev *ttnpb.EndDevice) bool {
				a := assertions.New(t)
				return a.So(dev, should.Resemble, &ttnpb.EndDevice{
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
							ApplicationID: registeredApplicationID,
						},
						DeviceID: "new-dev-1",
						JoinEUI:  eui64Ptr(types.EUI64{0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1}),
						DevEUI:   eui64Ptr(types.EUI64{0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1}),
					},
					ProvisionerID: "mock",
					ProvisioningData: &pbtypes.Struct{
						Fields: map[string]*pbtypes.Value{
							"serial_number": {
								Kind: &pbtypes.Value_NumberValue{
									NumberValue: 1,
								},
							},
						},
					},
				})
			},
		},

		{
			Name: "List invalid number of entries",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, ttnpb.ApplicationIdentifiers{ApplicationID: registeredApplicationID}): ttnpb.RightsFrom(
							ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
							ttnpb.RIGHT_APPLICATION_DEVICES_WRITE_KEYS,
						),
					},
				})
			},
			ProvisionRequest: &ttnpb.ProvisionEndDevicesRequest{
				ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
					ApplicationID: registeredApplicationID,
				},
				EndDevices: &ttnpb.ProvisionEndDevicesRequest_List{
					List: &ttnpb.ProvisionEndDevicesRequest_IdentifiersList{
						EndDeviceIDs: []ttnpb.EndDeviceIdentifiers{
							{
								ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
									ApplicationID: registeredApplicationID,
								},
								DeviceID: "new-dev-1",
								JoinEUI:  eui64Ptr(types.EUI64{0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1}),
								DevEUI:   eui64Ptr(types.EUI64{0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1}),
							},
							{
								ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
									ApplicationID: registeredApplicationID,
								},
								DeviceID: "new-dev-42",
								JoinEUI:  eui64Ptr(types.EUI64{0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1}),
								DevEUI:   eui64Ptr(types.EUI64{0x2a, 0x2a, 0x2a, 0x2a, 0x2a, 0x2a, 0x2a, 0x2a}),
							},
						},
					},
				},
				ProvisionerID:    "mock",
				ProvisioningData: []byte{0x1, 0x2a, 0xff},
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				a := assertions.New(t)
				return a.So(errors.IsInvalidArgument(err), should.BeTrue)
			},
			MultipleAssertion: func(t *testing.T, devs []*ttnpb.EndDevice) bool {
				a := assertions.New(t)
				return a.So(devs, should.Resemble, []*ttnpb.EndDevice{
					{
						EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
							ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
								ApplicationID: registeredApplicationID,
							},
							DeviceID: "new-dev-1",
							JoinEUI:  eui64Ptr(types.EUI64{0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1}),
							DevEUI:   eui64Ptr(types.EUI64{0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1}),
						},
						ProvisionerID: "mock",
						ProvisioningData: &pbtypes.Struct{
							Fields: map[string]*pbtypes.Value{
								"serial_number": {
									Kind: &pbtypes.Value_NumberValue{
										NumberValue: 1,
									},
								},
							},
						},
					},
					{
						EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
							ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
								ApplicationID: registeredApplicationID,
							},
							DeviceID: "new-dev-42",
							JoinEUI:  eui64Ptr(types.EUI64{0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1, 0x1}),
							DevEUI:   eui64Ptr(types.EUI64{0x2a, 0x2a, 0x2a, 0x2a, 0x2a, 0x2a, 0x2a, 0x2a}),
						},
						ProvisionerID: "mock",
						ProvisioningData: &pbtypes.Struct{
							Fields: map[string]*pbtypes.Value{
								"serial_number": {
									Kind: &pbtypes.Value_NumberValue{
										NumberValue: 42,
									},
								},
							},
						},
					},
				})
			},
		},

		{
			Name: "Range three devices",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, ttnpb.ApplicationIdentifiers{ApplicationID: registeredApplicationID}): ttnpb.RightsFrom(
							ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
							ttnpb.RIGHT_APPLICATION_DEVICES_WRITE_KEYS,
						),
					},
				})
			},
			ProvisionRequest: &ttnpb.ProvisionEndDevicesRequest{
				ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
					ApplicationID: registeredApplicationID,
				},
				EndDevices: &ttnpb.ProvisionEndDevicesRequest_Range{
					Range: &ttnpb.ProvisionEndDevicesRequest_IdentifiersRange{
						JoinEUI:     eui64Ptr(types.EUI64{0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42}),
						StartDevEUI: types.EUI64{0x2, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
					},
				},
				ProvisionerID:    "mock",
				ProvisioningData: []byte{0x1, 0x2a, 0xff},
			},
			MultipleAssertion: func(t *testing.T, devs []*ttnpb.EndDevice) bool {
				a := assertions.New(t)
				return a.So(devs, should.Resemble, []*ttnpb.EndDevice{
					{
						EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
							ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
								ApplicationID: registeredApplicationID,
							},
							DeviceID: "sn-1",
							JoinEUI:  eui64Ptr(types.EUI64{0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42}),
							DevEUI:   eui64Ptr(types.EUI64{0x2, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}),
						},
						ProvisionerID: "mock",
						ProvisioningData: &pbtypes.Struct{
							Fields: map[string]*pbtypes.Value{
								"serial_number": {
									Kind: &pbtypes.Value_NumberValue{
										NumberValue: 1,
									},
								},
							},
						},
					},
					{
						EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
							ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
								ApplicationID: registeredApplicationID,
							},
							DeviceID: "sn-42",
							JoinEUI:  eui64Ptr(types.EUI64{0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42}),
							DevEUI:   eui64Ptr(types.EUI64{0x2, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1}),
						},
						ProvisionerID: "mock",
						ProvisioningData: &pbtypes.Struct{
							Fields: map[string]*pbtypes.Value{
								"serial_number": {
									Kind: &pbtypes.Value_NumberValue{
										NumberValue: 42,
									},
								},
							},
						},
					},
					{
						EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
							ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
								ApplicationID: registeredApplicationID,
							},
							DeviceID: "sn-255",
							JoinEUI:  eui64Ptr(types.EUI64{0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42}),
							DevEUI:   eui64Ptr(types.EUI64{0x2, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x2}),
						},
						ProvisionerID: "mock",
						ProvisioningData: &pbtypes.Struct{
							Fields: map[string]*pbtypes.Value{
								"serial_number": {
									Kind: &pbtypes.Value_NumberValue{
										NumberValue: 255,
									},
								},
							},
						},
					},
				})
			},
		},

		{
			Name: "From data with three devices",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, ttnpb.ApplicationIdentifiers{ApplicationID: registeredApplicationID}): ttnpb.RightsFrom(
							ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
							ttnpb.RIGHT_APPLICATION_DEVICES_WRITE_KEYS,
						),
					},
				})
			},
			ProvisionRequest: &ttnpb.ProvisionEndDevicesRequest{
				ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
					ApplicationID: registeredApplicationID,
				},
				EndDevices: &ttnpb.ProvisionEndDevicesRequest_FromData{
					FromData: &ttnpb.ProvisionEndDevicesRequest_IdentifiersFromData{
						JoinEUI: eui64Ptr(types.EUI64{0x42, 0x42, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}),
					},
				},
				ProvisionerID:    "mock",
				ProvisioningData: []byte{0x1, 0x2a, 0xff},
			},
			MultipleAssertion: func(t *testing.T, devs []*ttnpb.EndDevice) bool {
				a := assertions.New(t)
				return a.So(devs, should.Resemble, []*ttnpb.EndDevice{
					{
						EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
							ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
								ApplicationID: registeredApplicationID,
							},
							DeviceID: "sn-1",
							JoinEUI:  eui64Ptr(types.EUI64{0x42, 0x42, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}),
							DevEUI:   eui64Ptr(types.EUI64{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1}),
						},
						ProvisionerID: "mock",
						ProvisioningData: &pbtypes.Struct{
							Fields: map[string]*pbtypes.Value{
								"serial_number": {
									Kind: &pbtypes.Value_NumberValue{
										NumberValue: 1,
									},
								},
							},
						},
					},
					{
						EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
							ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
								ApplicationID: registeredApplicationID,
							},
							DeviceID: "sn-42",
							JoinEUI:  eui64Ptr(types.EUI64{0x42, 0x42, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}),
							DevEUI:   eui64Ptr(types.EUI64{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x2a}),
						},
						ProvisionerID: "mock",
						ProvisioningData: &pbtypes.Struct{
							Fields: map[string]*pbtypes.Value{
								"serial_number": {
									Kind: &pbtypes.Value_NumberValue{
										NumberValue: 42,
									},
								},
							},
						},
					},
					{
						EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
							ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
								ApplicationID: registeredApplicationID,
							},
							DeviceID: "sn-255",
							JoinEUI:  eui64Ptr(types.EUI64{0x42, 0x42, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}),
							DevEUI:   eui64Ptr(types.EUI64{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xff}),
						},
						ProvisionerID: "mock",
						ProvisioningData: &pbtypes.Struct{
							Fields: map[string]*pbtypes.Value{
								"serial_number": {
									Kind: &pbtypes.Value_NumberValue{
										NumberValue: 255,
									},
								},
							},
						},
					},
				})
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			js := test.Must(New(
				component.MustNew(test.GetLogger(t), &component.Config{}),
				&Config{},
			)).(*JoinServer)

			js.AddContextFiller(tc.ContextFunc)
			js.AddContextFiller(func(ctx context.Context) context.Context {
				ctx, cancel := context.WithDeadline(ctx, time.Now().Add(Timeout))
				_ = cancel
				return ctx
			})
			js.AddContextFiller(func(ctx context.Context) context.Context {
				return test.ContextWithT(ctx, t)
			})
			test.Must(nil, js.Start())
			defer js.Close()

			ctx := js.FillContext(test.Context())

			var devs []*ttnpb.EndDevice
			stream, err := ttnpb.NewJsEndDeviceRegistryClient(js.LoopbackConn()).Provision(ctx, tc.ProvisionRequest)
			a.So(err, should.BeNil)

			for {
				dev, err := stream.Recv()
				if err != nil {
					if err == io.EOF {
						a.So(tc.ErrorAssertion, should.BeNil)
					} else if a.So(tc.ErrorAssertion, should.NotBeNil) {
						a.So(tc.ErrorAssertion(t, err), should.BeTrue)
					}
					break
				}
				devs = append(devs, dev)
			}
			if tc.SingleAssertion != nil && a.So(devs, should.HaveLength, 1) {
				a.So(tc.SingleAssertion(t, devs[0]), should.BeTrue)
				return
			}
			if tc.MultipleAssertion != nil {
				a.So(tc.MultipleAssertion(t, devs), should.BeTrue)
				return
			}
			a.So(devs, should.BeEmpty)
		})
	}
}

type mockJsEndDeviceRegistryProvisionServer struct {
	*test.MockServerStream
	SendFunc func(*ttnpb.EndDevice) error
}

func (s *mockJsEndDeviceRegistryProvisionServer) Send(up *ttnpb.EndDevice) error {
	if s.SendFunc == nil {
		return nil
	}
	return s.SendFunc(up)
}
