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
	"sync/atomic"
	"testing"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/mohae/deepcopy"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/pkg/component/test"
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
						unique.ID(test.Context(), ttnpb.ApplicationIdentifiers{ApplicationID: registeredApplicationID}): nil,
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
						unique.ID(test.Context(), ttnpb.ApplicationIdentifiers{ApplicationID: registeredApplicationID}): ttnpb.RightsFrom(
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
						unique.ID(test.Context(), ttnpb.ApplicationIdentifiers{ApplicationID: "bar-application"}): ttnpb.RightsFrom(
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
				return CopyEndDevice(&ttnpb.EndDevice{
					EndDeviceIdentifiers: registeredDevice.EndDeviceIdentifiers,
				}), nil
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
						unique.ID(test.Context(), ttnpb.ApplicationIdentifiers{ApplicationID: registeredApplicationID}): ttnpb.RightsFrom(
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
				return CopyEndDevice(&ttnpb.EndDevice{
					EndDeviceIdentifiers: registeredDevice.EndDeviceIdentifiers,
				}), nil
			},
			DeviceRequest: &ttnpb.GetEndDeviceRequest{
				EndDeviceIdentifiers: deepcopy.Copy(registeredDevice.EndDeviceIdentifiers).(ttnpb.EndDeviceIdentifiers),
				FieldMask: pbtypes.FieldMask{
					Paths: []string{"ids"},
				},
			},
			DeviceAssertion: func(t *testing.T, dev *ttnpb.EndDevice) bool {
				return assertions.New(t).So(dev, should.Resemble, &ttnpb.EndDevice{
					EndDeviceIdentifiers: registeredDevice.EndDeviceIdentifiers,
				})
			},
			GetByIDCalls: 1,
		},

		{
			Name: "Get keys without permission",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(test.Context(), ttnpb.ApplicationIdentifiers{ApplicationID: registeredApplicationID}): ttnpb.RightsFrom(
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
					Paths: []string{"ids", "root_keys.app_key.key", "root_keys.nwk_key.key"},
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
						unique.ID(test.Context(), ttnpb.ApplicationIdentifiers{ApplicationID: registeredApplicationID}): ttnpb.RightsFrom(
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
				a.So(paths, should.HaveSameElementsDeep, []string{
					"ids",
					"provisioner_id",
					"provisioning_data",
					"root_keys.app_key.encrypted_key",
					"root_keys.app_key.kek_label",
					"root_keys.app_key.key",
					"root_keys.nwk_key.encrypted_key",
					"root_keys.nwk_key.kek_label",
					"root_keys.nwk_key.key",
				})
				return CopyEndDevice(registeredDevice), nil
			},
			DeviceRequest: &ttnpb.GetEndDeviceRequest{
				EndDeviceIdentifiers: deepcopy.Copy(registeredDevice.EndDeviceIdentifiers).(ttnpb.EndDeviceIdentifiers),
				FieldMask: pbtypes.FieldMask{
					Paths: []string{"ids", "root_keys.app_key.key", "root_keys.nwk_key.key"},
				},
			},
			DeviceAssertion: func(t *testing.T, dev *ttnpb.EndDevice) bool {
				a := assertions.New(t)
				expected := CopyEndDevice(registeredDevice)
				expected.RootKeys = &ttnpb.RootKeys{
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
				componenttest.NewComponent(t, &component.Config{}),
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
			componenttest.StartComponent(t, js.Component)
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
						unique.ID(test.Context(), ttnpb.ApplicationIdentifiers{ApplicationID: registeredApplicationID}): nil,
					},
				})
			},
			DeviceRequest: &ttnpb.SetEndDeviceRequest{
				EndDevice: *CopyEndDevice(registeredDevice),
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
						unique.ID(test.Context(), deepcopy.Copy(registeredDevice.EndDeviceIdentifiers.ApplicationIdentifiers).(ttnpb.ApplicationIdentifiers)): ttnpb.RightsFrom(
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
						unique.ID(test.Context(), deepcopy.Copy(registeredDevice.EndDeviceIdentifiers.ApplicationIdentifiers).(ttnpb.ApplicationIdentifiers)): ttnpb.RightsFrom(
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
						unique.ID(test.Context(), ttnpb.ApplicationIdentifiers{ApplicationID: registeredApplicationID}): ttnpb.RightsFrom(
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
			},
			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string, cb func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(appID, should.Resemble, ttnpb.ApplicationIdentifiers{
					ApplicationID: registeredApplicationID,
				})
				a.So(devID, should.Equal, "new-device")
				a.So(gets, should.BeEmpty)
				dev, sets, err := cb(nil)
				a.So(sets, should.HaveSameElementsDeep, []string{
					"ids.application_ids",
					"ids.dev_eui",
					"ids.device_id",
					"ids.join_eui",
				})
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
						unique.ID(test.Context(), ttnpb.ApplicationIdentifiers{ApplicationID: registeredApplicationID}): ttnpb.RightsFrom(
							ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
						),
					},
				})
			},
			DeviceRequest: &ttnpb.SetEndDeviceRequest{
				EndDevice: *CopyEndDevice(&ttnpb.EndDevice{
					EndDeviceIdentifiers: registeredDevice.EndDeviceIdentifiers,
					NetID:                &types.NetID{0x42, 0x00, 0x00},
				}),
				FieldMask: pbtypes.FieldMask{
					Paths: []string{"net_id"},
				},
			},
			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string, cb func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(appID, should.Resemble, ttnpb.ApplicationIdentifiers{
					ApplicationID: registeredApplicationID,
				})
				a.So(devID, should.Equal, registeredDeviceID)
				a.So(gets, should.HaveSameElementsDeep, []string{
					"net_id",
				})
				dev, sets, err := cb(CopyEndDevice(&ttnpb.EndDevice{
					EndDeviceIdentifiers: registeredDevice.EndDeviceIdentifiers,
					NetID:                registeredDevice.NetID,
				}))
				a.So(sets, should.HaveSameElementsDeep, []string{
					"net_id",
				})
				return dev, err
			},
			DeviceAssertion: func(t *testing.T, dev *ttnpb.EndDevice) bool {
				return assertions.New(t).So(dev, should.Resemble, &ttnpb.EndDevice{
					EndDeviceIdentifiers: registeredDevice.EndDeviceIdentifiers,
					NetID:                &types.NetID{0x42, 0x00, 0x00},
				})
			},
			SetByIDCalls: 1,
		},

		{
			Name: "Set keys without permission",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(test.Context(), ttnpb.ApplicationIdentifiers{ApplicationID: registeredApplicationID}): ttnpb.RightsFrom(
							ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
						),
					},
				})
			},
			DeviceRequest: &ttnpb.SetEndDeviceRequest{
				EndDevice: *CopyEndDevice(registeredDevice),
				FieldMask: pbtypes.FieldMask{
					Paths: []string{"root_keys.app_key.key", "root_keys.nwk_key.key"},
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
						unique.ID(test.Context(), ttnpb.ApplicationIdentifiers{ApplicationID: registeredApplicationID}): ttnpb.RightsFrom(
							ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
							ttnpb.RIGHT_APPLICATION_DEVICES_WRITE_KEYS,
						),
					},
				})
			},
			DeviceRequest: &ttnpb.SetEndDeviceRequest{
				EndDevice: *CopyEndDevice(&ttnpb.EndDevice{
					EndDeviceIdentifiers: registeredDevice.EndDeviceIdentifiers,
					RootKeys: &ttnpb.RootKeys{
						AppKey: &ttnpb.KeyEnvelope{
							Key: &types.AES128Key{0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff},
						},
						NwkKey: &ttnpb.KeyEnvelope{
							Key: &types.AES128Key{0xff, 0xee, 0xdd, 0xcc, 0xbb, 0xaa, 0x99, 0x88, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42},
						},
					},
				}),
				FieldMask: pbtypes.FieldMask{
					Paths: []string{"root_keys.app_key.key", "root_keys.nwk_key.key"},
				},
			},
			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string, cb func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(appID, should.Resemble, ttnpb.ApplicationIdentifiers{
					ApplicationID: registeredApplicationID,
				})
				a.So(devID, should.Equal, registeredDeviceID)
				a.So(gets, should.HaveSameElementsDeep, []string{
					"root_keys.app_key.key", "root_keys.nwk_key.key",
				})
				dev, sets, err := cb(CopyEndDevice(registeredDevice))
				a.So(sets, should.HaveSameElementsDeep, []string{
					"root_keys.app_key.key", "root_keys.nwk_key.key",
				})
				return dev, err
			},
			DeviceAssertion: func(t *testing.T, dev *ttnpb.EndDevice) bool {
				return assertions.New(t).So(dev, should.Resemble, &ttnpb.EndDevice{
					EndDeviceIdentifiers: registeredDevice.EndDeviceIdentifiers,
					RootKeys: &ttnpb.RootKeys{
						AppKey: &ttnpb.KeyEnvelope{
							Key: &types.AES128Key{0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff},
						},
						NwkKey: &ttnpb.KeyEnvelope{
							Key: &types.AES128Key{0xff, 0xee, 0xdd, 0xcc, 0xbb, 0xaa, 0x99, 0x88, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42},
						},
					},
				})
			},
			SetByIDCalls: 1,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			var setByIDCalls uint64

			js := test.Must(New(
				componenttest.NewComponent(t, &component.Config{}),
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
			componenttest.StartComponent(t, js.Component)
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
						unique.ID(test.Context(), ttnpb.ApplicationIdentifiers{ApplicationID: registeredApplicationID}): nil,
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
						unique.ID(test.Context(), ttnpb.ApplicationIdentifiers{ApplicationID: "bar-application"}): ttnpb.RightsFrom(
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
				dev, _, err := cb(CopyEndDevice(registeredDevice))
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
						unique.ID(test.Context(), ttnpb.ApplicationIdentifiers{ApplicationID: registeredApplicationID}): ttnpb.RightsFrom(
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
						unique.ID(test.Context(), ttnpb.ApplicationIdentifiers{ApplicationID: registeredApplicationID}): ttnpb.RightsFrom(
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
				dev, _, err := cb(CopyEndDevice(registeredDevice))
				return dev, err
			},
			SetByIDCalls: 1,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			var setByIDCalls uint64

			js := test.Must(New(
				componenttest.NewComponent(t, &component.Config{}),
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
			componenttest.StartComponent(t, js.Component)
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
