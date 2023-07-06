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
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/v3/pkg/component/test"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	. "go.thethings.network/lorawan-stack/v3/pkg/joinserver"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

var errNotFound = errors.DefineNotFound("not_found", "not found")

func TestDeviceRegistryGet(t *testing.T) { //nolint:paralleltest
	registeredApplicationID := "foo-application"
	registeredDeviceID := "foo-device"
	registeredRootKeyID := "testKey"
	registeredKEKLabel := "test"
	unregisteredDeviceID := "bar-device"
	registeredJoinEUI := types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	registeredDevEUI := types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	registeredNwkKey := &types.AES128Key{
		0xff, 0xee, 0xdd, 0xcc, 0xbb, 0xaa, 0x99, 0x88, 0x77, 0x66, 0x55, 0x44, 0x33, 0x22, 0x11, 0x0,
	}
	registeredNwkKeyEnc := []byte{
		0xa3, 0x34, 0x38, 0x1c, 0xca, 0x1c, 0x12, 0x7a, 0x5b, 0xb1, 0xa8, 0x97,
		0x39, 0xc7, 0x5, 0x34, 0x91, 0x26, 0x9b, 0x21, 0x4f, 0x27, 0x80, 0x19,
	}
	registeredAppKey := types.AES128Key{
		0x0, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff,
	}
	registeredAppKeyEnc := []byte{
		0x1f, 0xa6, 0x8b, 0xa, 0x81, 0x12, 0xb4, 0x47, 0xae, 0xf3, 0x4b, 0xd8,
		0xfb, 0x5a, 0x7b, 0x82, 0x9d, 0x3e, 0x86, 0x23, 0x71, 0xd2, 0xcf, 0xe5,
	}
	unregisteredJoinEUI := types.EUI64{0x42, 0x42, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	unregisteredDevEUI := types.EUI64{0x42, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	keyVault := map[string][]byte{
		"test": {0x0, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf},
	}
	registeredDevice := &ttnpb.EndDevice{
		Ids: &ttnpb.EndDeviceIdentifiers{
			ApplicationIds: &ttnpb.ApplicationIdentifiers{
				ApplicationId: registeredApplicationID,
			},
			DeviceId: registeredDeviceID,
			JoinEui:  registeredJoinEUI.Bytes(),
			DevEui:   registeredDevEUI.Bytes(),
		},
		RootKeys: &ttnpb.RootKeys{
			RootKeyId: registeredRootKeyID,
			NwkKey: &ttnpb.KeyEnvelope{
				Key:          registeredNwkKey.Bytes(),
				KekLabel:     registeredKEKLabel,
				EncryptedKey: registeredNwkKeyEnc,
			},
			AppKey: &ttnpb.KeyEnvelope{
				Key:          registeredAppKey.Bytes(),
				KekLabel:     registeredKEKLabel,
				EncryptedKey: registeredAppKeyEnc,
			},
		},
	}
	for _, tc := range []struct { //nolint:paralleltest
		Name            string
		ContextFunc     func(context.Context) context.Context
		GetByIDFunc     func(ctx context.Context, appID *ttnpb.ApplicationIdentifiers, devID string, paths []string) (*ttnpb.EndDevice, error)
		DeviceRequest   *ttnpb.GetEndDeviceRequest
		ErrorAssertion  func(*testing.T, error) bool
		DeviceAssertion func(*testing.T, *ttnpb.EndDevice) bool
		GetByIDCalls    uint64
	}{
		{
			Name: "Permission denied",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, &rights.Rights{
					ApplicationRights: *rights.NewMap(map[string]*ttnpb.Rights{
						unique.ID(test.Context(), &ttnpb.ApplicationIdentifiers{
							ApplicationId: registeredApplicationID,
						}): nil,
					}),
				})
			},
			GetByIDFunc: func(ctx context.Context, appID *ttnpb.ApplicationIdentifiers, devID string, paths []string) (*ttnpb.EndDevice, error) {
				test.MustTFromContext(ctx).Errorf("GetByIDFunc must not be called")
				return nil, errors.New("GetByIDFunc must not be called")
			},
			DeviceRequest: &ttnpb.GetEndDeviceRequest{
				EndDeviceIds: ttnpb.Clone(registeredDevice.Ids),
				FieldMask:    ttnpb.FieldMask("ids"),
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				a := assertions.New(t)
				return a.So(errors.IsPermissionDenied(err), should.BeTrue)
			},
		},

		{
			Name: "Not found",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, &rights.Rights{
					ApplicationRights: *rights.NewMap(map[string]*ttnpb.Rights{
						unique.ID(test.Context(), &ttnpb.ApplicationIdentifiers{ApplicationId: registeredApplicationID}): ttnpb.RightsFrom(
							ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ,
						),
					}),
				})
			},
			GetByIDFunc: func(ctx context.Context, appID *ttnpb.ApplicationIdentifiers, devID string, paths []string) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(appID, should.Resemble, &ttnpb.ApplicationIdentifiers{
					ApplicationId: registeredApplicationID,
				})
				a.So(devID, should.Equal, unregisteredDeviceID)
				a.So(paths, should.HaveSameElementsDeep, []string{"ids"})
				return nil, errNotFound.New()
			},
			DeviceRequest: &ttnpb.GetEndDeviceRequest{
				EndDeviceIds: &ttnpb.EndDeviceIdentifiers{
					ApplicationIds: &ttnpb.ApplicationIdentifiers{
						ApplicationId: registeredApplicationID,
					},
					DeviceId: unregisteredDeviceID,
					JoinEui:  unregisteredJoinEUI.Bytes(),
					DevEui:   unregisteredDevEUI.Bytes(),
				},
				FieldMask: ttnpb.FieldMask("ids"),
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
				return rights.NewContext(ctx, &rights.Rights{
					ApplicationRights: *rights.NewMap(map[string]*ttnpb.Rights{
						unique.ID(test.Context(), &ttnpb.ApplicationIdentifiers{
							ApplicationId: "bar-application",
						}): ttnpb.RightsFrom(
							ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ,
						),
					}),
				})
			},
			GetByIDFunc: func(ctx context.Context, appID *ttnpb.ApplicationIdentifiers, devID string, paths []string) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(appID, should.Resemble, &ttnpb.ApplicationIdentifiers{
					ApplicationId: "bar-application",
				})
				a.So(devID, should.Equal, "bar-device")
				a.So(paths, should.HaveSameElementsDeep, []string{"ids"})
				return ttnpb.Clone(&ttnpb.EndDevice{
					Ids: registeredDevice.Ids,
				}), nil
			},
			DeviceRequest: &ttnpb.GetEndDeviceRequest{
				EndDeviceIds: &ttnpb.EndDeviceIdentifiers{
					ApplicationIds: &ttnpb.ApplicationIdentifiers{
						ApplicationId: "bar-application",
					},
					DeviceId: "bar-device",
					JoinEui:  registeredJoinEUI.Bytes(),
					DevEui:   registeredDevEUI.Bytes(),
				},
				FieldMask: ttnpb.FieldMask("ids"),
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
				return rights.NewContext(ctx, &rights.Rights{
					ApplicationRights: *rights.NewMap(map[string]*ttnpb.Rights{
						unique.ID(test.Context(), &ttnpb.ApplicationIdentifiers{
							ApplicationId: registeredApplicationID,
						}): ttnpb.RightsFrom(
							ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ,
						),
					}),
				})
			},
			GetByIDFunc: func(ctx context.Context, appID *ttnpb.ApplicationIdentifiers, devID string, paths []string) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(appID, should.Resemble, &ttnpb.ApplicationIdentifiers{
					ApplicationId: registeredApplicationID,
				})
				a.So(devID, should.Equal, registeredDeviceID)
				a.So(paths, should.HaveSameElementsDeep, []string{"ids"})
				return ttnpb.Clone(&ttnpb.EndDevice{
					Ids: registeredDevice.Ids,
				}), nil
			},
			DeviceRequest: &ttnpb.GetEndDeviceRequest{
				EndDeviceIds: ttnpb.Clone(registeredDevice.Ids),
				FieldMask:    ttnpb.FieldMask("ids"),
			},
			DeviceAssertion: func(t *testing.T, dev *ttnpb.EndDevice) bool {
				return assertions.New(t).So(dev, should.Resemble, &ttnpb.EndDevice{
					Ids: registeredDevice.Ids,
				})
			},
			GetByIDCalls: 1,
		},

		{
			Name: "Get keys without permission",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, &rights.Rights{
					ApplicationRights: *rights.NewMap(map[string]*ttnpb.Rights{
						unique.ID(test.Context(), &ttnpb.ApplicationIdentifiers{
							ApplicationId: registeredApplicationID,
						}): ttnpb.RightsFrom(
							ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ,
						),
					}),
				})
			},
			GetByIDFunc: func(ctx context.Context, appID *ttnpb.ApplicationIdentifiers, devID string, paths []string) (*ttnpb.EndDevice, error) {
				test.MustTFromContext(ctx).Errorf("GetByIDFunc must not be called")
				return nil, errors.New("GetByIDFunc must not be called")
			},
			DeviceRequest: &ttnpb.GetEndDeviceRequest{
				EndDeviceIds: ttnpb.Clone(registeredDevice.Ids),
				FieldMask:    ttnpb.FieldMask("ids", "root_keys.app_key.key", "root_keys.nwk_key.key"),
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				a := assertions.New(t)
				return a.So(errors.IsPermissionDenied(err), should.BeTrue)
			},
		},

		{
			Name: "Get keys/AppKey encrypted/NwkKey plaintext",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, &rights.Rights{
					ApplicationRights: *rights.NewMap(map[string]*ttnpb.Rights{
						unique.ID(test.Context(), &ttnpb.ApplicationIdentifiers{
							ApplicationId: registeredApplicationID,
						}): ttnpb.RightsFrom(
							ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ,
							ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ_KEYS,
						),
					}),
				})
			},
			GetByIDFunc: func(ctx context.Context, appID *ttnpb.ApplicationIdentifiers, devID string, paths []string) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(appID, should.Resemble, &ttnpb.ApplicationIdentifiers{
					ApplicationId: registeredApplicationID,
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
				ret := ttnpb.Clone(registeredDevice)
				ret.RootKeys.AppKey = &ttnpb.KeyEnvelope{
					EncryptedKey: ret.RootKeys.AppKey.EncryptedKey,
					KekLabel:     ret.RootKeys.AppKey.KekLabel,
				}
				ret.RootKeys.NwkKey = &ttnpb.KeyEnvelope{
					Key: ret.RootKeys.NwkKey.Key,
				}
				return ret, nil
			},
			DeviceRequest: &ttnpb.GetEndDeviceRequest{
				EndDeviceIds: ttnpb.Clone(registeredDevice.Ids),
				FieldMask:    ttnpb.FieldMask("ids", "root_keys.app_key.key", "root_keys.nwk_key.key"),
			},
			DeviceAssertion: func(t *testing.T, dev *ttnpb.EndDevice) bool {
				a := assertions.New(t)
				expected := ttnpb.Clone(registeredDevice)
				expected.RootKeys = &ttnpb.RootKeys{
					NwkKey: &ttnpb.KeyEnvelope{
						Key: registeredNwkKey.Bytes(),
					},
					AppKey: &ttnpb.KeyEnvelope{
						Key: registeredAppKey.Bytes(),
					},
				}
				return a.So(dev, should.Resemble, expected)
			},
			GetByIDCalls: 1,
		},

		{
			Name: "Get keys/AppKey plaintext/NwkKey encrypted",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, &rights.Rights{
					ApplicationRights: *rights.NewMap(map[string]*ttnpb.Rights{
						unique.ID(test.Context(), &ttnpb.ApplicationIdentifiers{
							ApplicationId: registeredApplicationID,
						}): ttnpb.RightsFrom(
							ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ,
							ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ_KEYS,
						),
					}),
				})
			},
			GetByIDFunc: func(ctx context.Context, appID *ttnpb.ApplicationIdentifiers, devID string, paths []string) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(appID, should.Resemble, &ttnpb.ApplicationIdentifiers{
					ApplicationId: registeredApplicationID,
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
				ret := ttnpb.Clone(registeredDevice)
				ret.RootKeys.AppKey = &ttnpb.KeyEnvelope{
					Key: ret.RootKeys.AppKey.Key,
				}
				ret.RootKeys.NwkKey = &ttnpb.KeyEnvelope{
					EncryptedKey: ret.RootKeys.NwkKey.EncryptedKey,
					KekLabel:     ret.RootKeys.NwkKey.KekLabel,
				}
				return ret, nil
			},
			DeviceRequest: &ttnpb.GetEndDeviceRequest{
				EndDeviceIds: ttnpb.Clone(registeredDevice.Ids),
				FieldMask:    ttnpb.FieldMask("ids", "root_keys.app_key.key", "root_keys.nwk_key.key"),
			},
			DeviceAssertion: func(t *testing.T, dev *ttnpb.EndDevice) bool {
				a := assertions.New(t)
				expected := ttnpb.Clone(registeredDevice)
				expected.RootKeys = &ttnpb.RootKeys{
					NwkKey: &ttnpb.KeyEnvelope{
						Key: registeredNwkKey.Bytes(),
					},
					AppKey: &ttnpb.KeyEnvelope{
						Key: registeredAppKey.Bytes(),
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
				componenttest.NewComponent(t, &component.Config{
					ServiceBase: config.ServiceBase{
						KeyVault: config.KeyVault{
							Provider: "static",
							Static:   keyVault,
						},
					},
				}),
				&Config{
					Devices: &MockDeviceRegistry{
						GetByIDFunc: func(ctx context.Context, appID *ttnpb.ApplicationIdentifiers, devID string, paths []string) (*ttnpb.EndDevice, error) {
							atomic.AddUint64(&getByIDCalls, 1)
							return tc.GetByIDFunc(ctx, appID, devID, paths)
						},
					},
					DevNonceLimit: defaultDevNonceLimit,
				},
			))

			js.AddContextFiller(tc.ContextFunc)
			js.AddContextFiller(func(ctx context.Context) context.Context {
				ctx, cancel := context.WithDeadline(ctx, time.Now().Add(Timeout))
				_ = cancel
				return ctx
			})
			js.AddContextFiller(func(ctx context.Context) context.Context {
				return test.ContextWithTB(ctx, t)
			})
			componenttest.StartComponent(t, js.Component)
			defer js.Close()

			ctx := js.FillContext(test.Context())
			req := ttnpb.Clone(tc.DeviceRequest)

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

func TestDeviceRegistrySet(t *testing.T) { //nolint:paralleltest
	registeredApplicationID := "foo-application"
	registeredDeviceID := "foo-device"
	registeredRootKeyID := "testKey"
	registeredKEKLabel := "test"
	registeredJoinEUI := types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	registeredDevEUI := types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	registeredNwkKey := &types.AES128Key{
		0xff, 0xee, 0xdd, 0xcc, 0xbb, 0xaa, 0x99, 0x88, 0x77, 0x66, 0x55, 0x44, 0x33, 0x22, 0x11, 0x0,
	}
	registeredAppKey := &types.AES128Key{
		0x0, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff,
	}
	unregisteredJoinEUI := types.EUI64{0x42, 0x42, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	unregisteredDevEUI := types.EUI64{0x42, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	keyVault := map[string][]byte{
		"test": {0x0, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf},
	}
	registeredDevice := &ttnpb.EndDevice{
		Ids: &ttnpb.EndDeviceIdentifiers{
			ApplicationIds: &ttnpb.ApplicationIdentifiers{
				ApplicationId: registeredApplicationID,
			},
			DeviceId: registeredDeviceID,
			JoinEui:  registeredJoinEUI.Bytes(),
			DevEui:   registeredDevEUI.Bytes(),
		},
		RootKeys: &ttnpb.RootKeys{
			RootKeyId: registeredRootKeyID,
			NwkKey: &ttnpb.KeyEnvelope{
				Key:      registeredNwkKey.Bytes(),
				KekLabel: registeredKEKLabel,
			},
			AppKey: &ttnpb.KeyEnvelope{
				Key:      registeredAppKey.Bytes(),
				KekLabel: registeredKEKLabel,
			},
		},
	}
	for _, tc := range []struct { //nolint:paralleltest
		Name            string
		ContextFunc     func(context.Context) context.Context
		SetByIDFunc     func(ctx context.Context, appID *ttnpb.ApplicationIdentifiers, devID string, paths []string, cb func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error)
		DeviceRequest   *ttnpb.SetEndDeviceRequest
		ErrorAssertion  func(*testing.T, error) bool
		DeviceAssertion func(*testing.T, *ttnpb.EndDevice) bool
		SetByIDCalls    uint64
	}{
		{
			Name: "Permission denied",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, &rights.Rights{
					ApplicationRights: *rights.NewMap(map[string]*ttnpb.Rights{
						unique.ID(test.Context(), &ttnpb.ApplicationIdentifiers{
							ApplicationId: registeredApplicationID,
						}): nil,
					}),
				})
			},
			DeviceRequest: &ttnpb.SetEndDeviceRequest{
				EndDevice: ttnpb.Clone(registeredDevice),
			},
			SetByIDFunc: func(ctx context.Context, appID *ttnpb.ApplicationIdentifiers, devID string, paths []string, cb func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
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
				return rights.NewContext(ctx, &rights.Rights{
					ApplicationRights: *rights.NewMap(map[string]*ttnpb.Rights{
						unique.ID(test.Context(), ttnpb.Clone(registeredDevice.Ids.ApplicationIds)): ttnpb.RightsFrom(
							ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE,
						),
					}),
				})
			},
			DeviceRequest: &ttnpb.SetEndDeviceRequest{
				EndDevice: &ttnpb.EndDevice{
					Ids: &ttnpb.EndDeviceIdentifiers{
						ApplicationIds: &ttnpb.ApplicationIdentifiers{
							ApplicationId: registeredApplicationID,
						},
						DeviceId: registeredDeviceID,
						DevEui:   registeredDevEUI.Bytes(),
					},
				},
			},
			SetByIDFunc: func(ctx context.Context, appID *ttnpb.ApplicationIdentifiers, devID string, paths []string, cb func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
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
				return rights.NewContext(ctx, &rights.Rights{
					ApplicationRights: *rights.NewMap(map[string]*ttnpb.Rights{
						unique.ID(test.Context(), ttnpb.Clone(registeredDevice.Ids.ApplicationIds)): ttnpb.RightsFrom(
							ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE,
						),
					}),
				})
			},
			DeviceRequest: &ttnpb.SetEndDeviceRequest{
				EndDevice: &ttnpb.EndDevice{
					Ids: &ttnpb.EndDeviceIdentifiers{
						ApplicationIds: &ttnpb.ApplicationIdentifiers{
							ApplicationId: registeredApplicationID,
						},
						DeviceId: registeredDeviceID,
						JoinEui:  registeredJoinEUI.Bytes(),
					},
				},
			},
			SetByIDFunc: func(ctx context.Context, appID *ttnpb.ApplicationIdentifiers, devID string, paths []string, cb func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
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
				return rights.NewContext(ctx, &rights.Rights{
					ApplicationRights: *rights.NewMap(map[string]*ttnpb.Rights{
						unique.ID(test.Context(), &ttnpb.ApplicationIdentifiers{
							ApplicationId: registeredApplicationID,
						}): ttnpb.RightsFrom(
							ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE,
						),
					}),
				})
			},
			DeviceRequest: &ttnpb.SetEndDeviceRequest{
				EndDevice: &ttnpb.EndDevice{
					Ids: &ttnpb.EndDeviceIdentifiers{
						ApplicationIds: &ttnpb.ApplicationIdentifiers{
							ApplicationId: registeredApplicationID,
						},
						DeviceId: "new-device",
						JoinEui:  unregisteredJoinEUI.Bytes(),
						DevEui:   unregisteredDevEUI.Bytes(),
					},
				},
			},
			SetByIDFunc: func(ctx context.Context, appID *ttnpb.ApplicationIdentifiers, devID string, gets []string, cb func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(appID, should.Resemble, &ttnpb.ApplicationIdentifiers{
					ApplicationId: registeredApplicationID,
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
				a.So(dev, should.Resemble, &ttnpb.EndDevice{
					Ids: &ttnpb.EndDeviceIdentifiers{
						ApplicationIds: &ttnpb.ApplicationIdentifiers{
							ApplicationId: registeredApplicationID,
						},
						DeviceId: "new-device",
						JoinEui:  unregisteredJoinEUI.Bytes(),
						DevEui:   unregisteredDevEUI.Bytes(),
					},
				})
				return dev, err
			},
			DeviceAssertion: func(t *testing.T, dev *ttnpb.EndDevice) bool {
				a := assertions.New(t)
				return a.So(dev, should.Resemble, &ttnpb.EndDevice{
					Ids: &ttnpb.EndDeviceIdentifiers{
						ApplicationIds: &ttnpb.ApplicationIdentifiers{
							ApplicationId: registeredApplicationID,
						},
						DeviceId: "new-device",
						JoinEui:  unregisteredJoinEUI.Bytes(),
						DevEui:   unregisteredDevEUI.Bytes(),
					},
				})
			},
			SetByIDCalls: 1,
		},

		{
			Name: "Set without keys",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, &rights.Rights{
					ApplicationRights: *rights.NewMap(map[string]*ttnpb.Rights{
						unique.ID(test.Context(), &ttnpb.ApplicationIdentifiers{
							ApplicationId: registeredApplicationID,
						}): ttnpb.RightsFrom(
							ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE,
						),
					}),
				})
			},
			DeviceRequest: &ttnpb.SetEndDeviceRequest{
				EndDevice: ttnpb.Clone(&ttnpb.EndDevice{
					Ids:   registeredDevice.Ids,
					NetId: types.NetID{0x42, 0x00, 0x00}.Bytes(),
				}),
				FieldMask: ttnpb.FieldMask("net_id"),
			},
			SetByIDFunc: func(ctx context.Context, appID *ttnpb.ApplicationIdentifiers, devID string, gets []string, cb func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(appID, should.Resemble, &ttnpb.ApplicationIdentifiers{
					ApplicationId: registeredApplicationID,
				})
				a.So(devID, should.Equal, registeredDeviceID)
				a.So(gets, should.HaveSameElementsDeep, []string{
					"net_id",
				})
				dev, sets, err := cb(ttnpb.Clone(&ttnpb.EndDevice{
					Ids:   registeredDevice.Ids,
					NetId: registeredDevice.NetId,
				}))
				a.So(sets, should.HaveSameElementsDeep, []string{
					"net_id",
				})
				a.So(dev, should.Resemble, &ttnpb.EndDevice{
					Ids:   registeredDevice.Ids,
					NetId: types.NetID{0x42, 0x00, 0x00}.Bytes(),
				})
				return dev, err
			},
			DeviceAssertion: func(t *testing.T, dev *ttnpb.EndDevice) bool {
				return assertions.New(t).So(dev, should.Resemble, &ttnpb.EndDevice{
					Ids:   registeredDevice.Ids,
					NetId: types.NetID{0x42, 0x00, 0x00}.Bytes(),
				})
			},
			SetByIDCalls: 1,
		},

		{
			Name: "Set keys without permission",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, &rights.Rights{
					ApplicationRights: *rights.NewMap(map[string]*ttnpb.Rights{
						unique.ID(test.Context(), &ttnpb.ApplicationIdentifiers{
							ApplicationId: registeredApplicationID,
						}): ttnpb.RightsFrom(
							ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE,
						),
					}),
				})
			},
			DeviceRequest: &ttnpb.SetEndDeviceRequest{
				EndDevice: ttnpb.Clone(registeredDevice),
				FieldMask: ttnpb.FieldMask("root_keys.app_key.key", "root_keys.nwk_key.key"),
			},
			SetByIDFunc: func(ctx context.Context, appID *ttnpb.ApplicationIdentifiers, devID string, paths []string, cb func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
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
				return rights.NewContext(ctx, &rights.Rights{
					ApplicationRights: *rights.NewMap(map[string]*ttnpb.Rights{
						unique.ID(test.Context(), &ttnpb.ApplicationIdentifiers{
							ApplicationId: registeredApplicationID,
						}): ttnpb.RightsFrom(
							ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE,
							ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE_KEYS,
						),
					}),
				})
			},
			DeviceRequest: &ttnpb.SetEndDeviceRequest{
				EndDevice: ttnpb.Clone(&ttnpb.EndDevice{
					Ids: registeredDevice.Ids,
					RootKeys: &ttnpb.RootKeys{
						AppKey: &ttnpb.KeyEnvelope{
							Key: types.AES128Key{0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}.Bytes(),
						},
						NwkKey: &ttnpb.KeyEnvelope{
							Key: types.AES128Key{0xff, 0xee, 0xdd, 0xcc, 0xbb, 0xaa, 0x99, 0x88, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42}.Bytes(),
						},
					},
				}),
				FieldMask: ttnpb.FieldMask("root_keys.app_key.key", "root_keys.nwk_key.key"),
			},
			SetByIDFunc: func(ctx context.Context, appID *ttnpb.ApplicationIdentifiers, devID string, gets []string, cb func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(appID, should.Resemble, &ttnpb.ApplicationIdentifiers{
					ApplicationId: registeredApplicationID,
				})
				a.So(devID, should.Equal, registeredDeviceID)
				a.So(gets, should.HaveSameElementsDeep, []string{
					"root_keys.app_key.key",
					"root_keys.nwk_key.key",
				})
				dev, sets, err := cb(ttnpb.Clone(registeredDevice))
				a.So(sets, should.HaveSameElementsDeep, []string{
					"root_keys.app_key.encrypted_key",
					"root_keys.app_key.kek_label",
					"root_keys.app_key.key",
					"root_keys.nwk_key.encrypted_key",
					"root_keys.nwk_key.kek_label",
					"root_keys.nwk_key.key",
				})
				a.So(dev, should.Resemble, &ttnpb.EndDevice{
					Ids: registeredDevice.Ids,
					RootKeys: &ttnpb.RootKeys{
						AppKey: &ttnpb.KeyEnvelope{
							EncryptedKey: []byte{0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff},
						},
						NwkKey: &ttnpb.KeyEnvelope{
							EncryptedKey: []byte{0xff, 0xee, 0xdd, 0xcc, 0xbb, 0xaa, 0x99, 0x88, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42},
						},
					},
				})
				return dev, err
			},
			DeviceAssertion: func(t *testing.T, dev *ttnpb.EndDevice) bool {
				return assertions.New(t).So(dev, should.Resemble, &ttnpb.EndDevice{
					Ids: registeredDevice.Ids,
					RootKeys: &ttnpb.RootKeys{
						AppKey: &ttnpb.KeyEnvelope{
							Key: types.AES128Key{0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}.Bytes(),
						},
						NwkKey: &ttnpb.KeyEnvelope{
							Key: types.AES128Key{0xff, 0xee, 0xdd, 0xcc, 0xbb, 0xaa, 0x99, 0x88, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42}.Bytes(),
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
				componenttest.NewComponent(t, &component.Config{
					ServiceBase: config.ServiceBase{
						KeyVault: config.KeyVault{
							Provider: "static",
							Static:   keyVault,
						},
					},
				}),
				&Config{
					Devices: &MockDeviceRegistry{
						SetByIDFunc: func(ctx context.Context, appID *ttnpb.ApplicationIdentifiers, devID string, paths []string, cb func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
							atomic.AddUint64(&setByIDCalls, 1)
							return tc.SetByIDFunc(ctx, appID, devID, paths, cb)
						},
					},
					DevNonceLimit: defaultDevNonceLimit,
				},
			))

			js.AddContextFiller(tc.ContextFunc)
			js.AddContextFiller(func(ctx context.Context) context.Context {
				ctx, cancel := context.WithDeadline(ctx, time.Now().Add(Timeout))
				_ = cancel
				return ctx
			})
			js.AddContextFiller(func(ctx context.Context) context.Context {
				return test.ContextWithTB(ctx, t)
			})
			componenttest.StartComponent(t, js.Component)
			defer js.Close()

			ctx := js.FillContext(test.Context())
			req := ttnpb.Clone(tc.DeviceRequest)

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

func TestDeviceRegistryDelete(t *testing.T) { //nolint:paralleltest
	registeredApplicationID := "foo-application"
	registeredDeviceID := "foo-device"
	registeredRootKeyID := "testKey"
	registeredKEKLabel := "test"
	unregisteredDeviceID := "bar-device"
	registeredJoinEUI := types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	registeredDevEUI := types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	registeredNwkKey := &types.AES128Key{
		0xff, 0xee, 0xdd, 0xcc, 0xbb, 0xaa, 0x99, 0x88, 0x77, 0x66, 0x55, 0x44, 0x33, 0x22, 0x11, 0x0,
	}
	registeredAppKey := &types.AES128Key{
		0x0, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff,
	}
	unregisteredJoinEUI := types.EUI64{0x42, 0x42, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	unregisteredDevEUI := types.EUI64{0x42, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	keyVault := map[string][]byte{
		"test": {0x0, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf},
	}
	registeredDevice := &ttnpb.EndDevice{
		Ids: &ttnpb.EndDeviceIdentifiers{
			ApplicationIds: &ttnpb.ApplicationIdentifiers{
				ApplicationId: registeredApplicationID,
			},
			DeviceId: registeredDeviceID,
			JoinEui:  registeredJoinEUI.Bytes(),
			DevEui:   registeredDevEUI.Bytes(),
		},
		RootKeys: &ttnpb.RootKeys{
			RootKeyId: registeredRootKeyID,
			NwkKey: &ttnpb.KeyEnvelope{
				Key:      registeredNwkKey.Bytes(),
				KekLabel: registeredKEKLabel,
			},
			AppKey: &ttnpb.KeyEnvelope{
				Key:      registeredAppKey.Bytes(),
				KekLabel: registeredKEKLabel,
			},
		},
	}
	for _, tc := range []struct { //nolint:paralleltest
		Name            string
		ContextFunc     func(context.Context) context.Context
		SetByIDFunc     func(ctx context.Context, appID *ttnpb.ApplicationIdentifiers, devID string, paths []string, cb func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error)
		DeviceRequest   *ttnpb.EndDeviceIdentifiers
		ErrorAssertion  func(*testing.T, error) bool
		SetByIDCalls    uint64
		DeleteKeysCalls uint64
	}{
		{
			Name: "Permission denied",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, &rights.Rights{
					ApplicationRights: *rights.NewMap(map[string]*ttnpb.Rights{
						unique.ID(test.Context(), &ttnpb.ApplicationIdentifiers{
							ApplicationId: registeredApplicationID,
						}): nil,
					}),
				})
			},
			DeviceRequest: ttnpb.Clone(registeredDevice.Ids),
			SetByIDFunc: func(ctx context.Context, appID *ttnpb.ApplicationIdentifiers, devID string, paths []string, cb func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				test.MustTFromContext(ctx).Errorf("SetByIDFunc must not be called")
				return nil, errors.New("SetByIDFunc must not be called")
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				a := assertions.New(t)
				return a.So(errors.IsPermissionDenied(err), should.BeTrue)
			},
		},

		{
			Name: "Not found",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, &rights.Rights{
					ApplicationRights: *rights.NewMap(map[string]*ttnpb.Rights{
						unique.ID(test.Context(), &ttnpb.ApplicationIdentifiers{
							ApplicationId: registeredApplicationID,
						}): ttnpb.RightsFrom(
							ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE,
						),
					}),
				})
			},
			DeviceRequest: &ttnpb.EndDeviceIdentifiers{
				ApplicationIds: &ttnpb.ApplicationIdentifiers{
					ApplicationId: registeredApplicationID,
				},
				DeviceId: unregisteredDeviceID,
				JoinEui:  unregisteredJoinEUI.Bytes(),
				DevEui:   unregisteredDevEUI.Bytes(),
			},
			SetByIDFunc: func(ctx context.Context, appID *ttnpb.ApplicationIdentifiers, devID string, paths []string, cb func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(appID, should.Resemble, &ttnpb.ApplicationIdentifiers{
					ApplicationId: registeredApplicationID,
				})
				a.So(devID, should.Equal, unregisteredDeviceID)
				a.So(paths, should.BeNil)
				return nil, errNotFound.New()
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				a := assertions.New(t)
				return a.So(errors.IsNotFound(err), should.BeTrue)
			},
			SetByIDCalls:    1,
			DeleteKeysCalls: 0,
		},

		{
			Name: "Delete",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, &rights.Rights{
					ApplicationRights: *rights.NewMap(map[string]*ttnpb.Rights{
						unique.ID(test.Context(), &ttnpb.ApplicationIdentifiers{
							ApplicationId: registeredApplicationID,
						}): ttnpb.RightsFrom(
							ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE,
						),
					}),
				})
			},
			DeviceRequest: ttnpb.Clone(registeredDevice.Ids),
			SetByIDFunc: func(ctx context.Context, appID *ttnpb.ApplicationIdentifiers, devID string, paths []string, cb func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(appID, should.Resemble, &ttnpb.ApplicationIdentifiers{
					ApplicationId: registeredApplicationID,
				})
				a.So(devID, should.Equal, registeredDeviceID)
				a.So(paths, should.BeNil)
				dev, _, err := cb(ttnpb.Clone(registeredDevice))
				return dev, err
			},
			SetByIDCalls:    1,
			DeleteKeysCalls: 1,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			var (
				setByIDCalls    uint64
				deleteKeysCalls uint64
			)

			js := test.Must(New(
				componenttest.NewComponent(t, &component.Config{
					ServiceBase: config.ServiceBase{
						KeyVault: config.KeyVault{
							Provider: "static",
							Static:   keyVault,
						},
					},
				}),
				&Config{
					Devices: &MockDeviceRegistry{
						SetByIDFunc: func(ctx context.Context, appID *ttnpb.ApplicationIdentifiers, devID string, paths []string, cb func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
							atomic.AddUint64(&setByIDCalls, 1)
							return tc.SetByIDFunc(ctx, appID, devID, paths, cb)
						},
					},
					Keys: &MockKeyRegistry{
						DeleteFunc: func(c context.Context, e1, e2 types.EUI64) error {
							atomic.AddUint64(&deleteKeysCalls, 1)
							return nil
						},
					},
					DevNonceLimit: defaultDevNonceLimit,
				},
			))

			js.AddContextFiller(tc.ContextFunc)
			js.AddContextFiller(func(ctx context.Context) context.Context {
				ctx, cancel := context.WithDeadline(ctx, time.Now().Add(Timeout))
				_ = cancel
				return ctx
			})
			js.AddContextFiller(func(ctx context.Context) context.Context {
				return test.ContextWithTB(ctx, t)
			})
			componenttest.StartComponent(t, js.Component)
			defer js.Close()

			ctx := js.FillContext(test.Context())
			req := ttnpb.Clone(tc.DeviceRequest)

			_, err := ttnpb.NewJsEndDeviceRegistryClient(js.LoopbackConn()).Delete(ctx, req)
			a.So(setByIDCalls, should.Equal, tc.SetByIDCalls)
			a.So(deleteKeysCalls, should.Equal, tc.DeleteKeysCalls)
			if tc.ErrorAssertion != nil {
				a.So(tc.ErrorAssertion(t, err), should.BeTrue)
			} else {
				a.So(err, should.BeNil)
			}
			a.So(req, should.Resemble, tc.DeviceRequest)
		})
	}
}

func TestDeviceRegistryBatchDelete(t *testing.T) { // nolint:paralleltest
	registeredApplicationID := "test-app"
	registeredApplicationIDs := &ttnpb.ApplicationIdentifiers{
		ApplicationId: registeredApplicationID,
	}
	dev1 := &ttnpb.EndDevice{
		Ids: &ttnpb.EndDeviceIdentifiers{
			ApplicationIds: &ttnpb.ApplicationIdentifiers{
				ApplicationId: registeredApplicationID,
			},
			DeviceId: "test-device-1",
			JoinEui:  types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}.Bytes(),
			DevEui:   types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}.Bytes(),
		},
		RootKeys: &ttnpb.RootKeys{
			RootKeyId: "testKey",
			NwkKey: &ttnpb.KeyEnvelope{
				Key: types.AES128Key{
					0xff, 0xee, 0xdd, 0xcc, 0xbb, 0xaa, 0x99, 0x88, 0x77, 0x66, 0x55, 0x44, 0x33, 0x22, 0x11, 0x0,
				}.Bytes(),
				KekLabel: "test",
			},
			AppKey: &ttnpb.KeyEnvelope{
				Key: types.AES128Key{
					0x0, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff,
				}.Bytes(),
				KekLabel: "test",
			},
		},
	}
	dev2 := ttnpb.Clone(dev1)
	dev2.Ids.DeviceId = "test-device-2"
	dev2.Ids.JoinEui = types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}.Bytes()
	dev2.Ids.DevEui = types.EUI64{0x42, 0x43, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}.Bytes()

	dev3 := ttnpb.Clone(dev1)
	dev3.Ids.DeviceId = "test-device-3"
	dev3.Ids.JoinEui = types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}.Bytes()
	dev3.Ids.DevEui = types.EUI64{0x42, 0x44, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}.Bytes()

	for _, tc := range []struct {
		Name            string
		ContextFunc     func(context.Context) context.Context
		BatchDeleteFunc func(
			ctx context.Context,
			appIDs *ttnpb.ApplicationIdentifiers,
			deviceIDs []string,
		) ([]*ttnpb.EndDeviceIdentifiers, error)
		BatchDeleteKeysFunc  func(ctx context.Context, devIDs []*ttnpb.EndDeviceIdentifiers) error
		Request              *ttnpb.BatchDeleteEndDevicesRequest
		ErrorAssertion       func(*testing.T, error) bool
		BatchDeleteCalls     uint64
		BatchDeleteKeysCalls uint64
	}{
		{
			Name: "No device write rights",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, &rights.Rights{
					ApplicationRights: *rights.NewMap(map[string]*ttnpb.Rights{
						unique.ID(test.Context(), registeredApplicationIDs): {
							Rights: []ttnpb.Right{
								ttnpb.Right_RIGHT_GATEWAY_SETTINGS_BASIC,
							},
						},
					}),
				})
			},
			BatchDeleteFunc: func(
				ctx context.Context,
				appIDs *ttnpb.ApplicationIdentifiers,
				deviceIDs []string,
			) ([]*ttnpb.EndDeviceIdentifiers, error) {
				err := errors.New("BatchDeleteFunc must not be called")
				test.MustTFromContext(ctx).Error(err)
				return nil, err
			},
			BatchDeleteKeysFunc: func(ctx context.Context, devIDs []*ttnpb.EndDeviceIdentifiers) error {
				err := errors.New("BatchDeleteKeysFunc must not be called")
				test.MustTFromContext(ctx).Error(err)
				return err
			},
			Request: &ttnpb.BatchDeleteEndDevicesRequest{
				ApplicationIds: registeredApplicationIDs,
				DeviceIds: []string{
					dev1.Ids.DeviceId,
					dev2.Ids.DeviceId,
					dev3.Ids.DeviceId,
				},
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				t.Helper()
				if !assertions.New(t).So(errors.IsPermissionDenied(err), should.BeTrue) {
					t.Errorf("Received error: %s", err)
					return false
				}
				return true
			},
			BatchDeleteCalls:     0,
			BatchDeleteKeysCalls: 0,
		},
		{
			Name: "Non-existing device",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, &rights.Rights{
					ApplicationRights: *rights.NewMap(map[string]*ttnpb.Rights{
						unique.ID(test.Context(), registeredApplicationIDs): {
							Rights: []ttnpb.Right{
								ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE,
							},
						},
					}),
				})
			},
			BatchDeleteFunc: func(
				ctx context.Context,
				appIDs *ttnpb.ApplicationIdentifiers,
				deviceIDs []string,
			) ([]*ttnpb.EndDeviceIdentifiers, error) {
				// Devices not found are skipped.
				return nil, nil
			},
			BatchDeleteKeysFunc: func(ctx context.Context, devIDs []*ttnpb.EndDeviceIdentifiers) error {
				// Devices not found are skipped.
				return nil
			},
			Request: &ttnpb.BatchDeleteEndDevicesRequest{
				ApplicationIds: registeredApplicationIDs,
				DeviceIds: []string{
					dev1.Ids.DeviceId,
					dev2.Ids.DeviceId,
					dev3.Ids.DeviceId,
				},
			},
			BatchDeleteCalls:     1,
			BatchDeleteKeysCalls: 1,
		},
		{
			Name: "Wrong application ID",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, &rights.Rights{
					ApplicationRights: *rights.NewMap(map[string]*ttnpb.Rights{
						unique.ID(test.Context(), registeredApplicationIDs): {
							Rights: []ttnpb.Right{
								ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE,
							},
						},
					}),
				})
			},
			Request: &ttnpb.BatchDeleteEndDevicesRequest{
				ApplicationIds: &ttnpb.ApplicationIdentifiers{ApplicationId: "test-unknown-app-id"},
				DeviceIds: []string{
					dev1.Ids.DeviceId,
					dev2.Ids.DeviceId,
					dev3.Ids.DeviceId,
				},
			},
			BatchDeleteFunc: func(
				ctx context.Context,
				appIDs *ttnpb.ApplicationIdentifiers,
				deviceIDs []string,
			) ([]*ttnpb.EndDeviceIdentifiers, error) {
				err := errors.New("BatchDeleteFunc must not be called")
				test.MustTFromContext(ctx).Error(err)
				return nil, err
			},
			BatchDeleteKeysFunc: func(ctx context.Context, devIDs []*ttnpb.EndDeviceIdentifiers) error {
				err := errors.New("BatchDeleteKeysFunc must not be called")
				test.MustTFromContext(ctx).Error(err)
				return err
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				t.Helper()
				if !assertions.New(t).So(errors.IsPermissionDenied(err), should.BeTrue) {
					t.Errorf("Received error: %s", err)
					return false
				}
				return true
			},
			BatchDeleteCalls:     0,
			BatchDeleteKeysCalls: 0,
		},
		{
			Name: "Invalid Device",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, &rights.Rights{
					ApplicationRights: *rights.NewMap(map[string]*ttnpb.Rights{
						unique.ID(test.Context(), registeredApplicationIDs): {
							Rights: []ttnpb.Right{
								ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE,
							},
						},
					}),
				})
			},
			Request: &ttnpb.BatchDeleteEndDevicesRequest{
				ApplicationIds: registeredApplicationIDs,
				DeviceIds: []string{
					"test-dev-&*@(#)",
				},
			},
			BatchDeleteFunc: func(
				ctx context.Context,
				appIDs *ttnpb.ApplicationIdentifiers,
				deviceIDs []string,
			) ([]*ttnpb.EndDeviceIdentifiers, error) {
				err := errors.New("BatchDeleteFunc must not be called")
				test.MustTFromContext(ctx).Error(err)
				return nil, err
			},
			BatchDeleteKeysFunc: func(ctx context.Context, devIDs []*ttnpb.EndDeviceIdentifiers) error {
				err := errors.New("BatchDeleteKeysFunc must not be called")
				test.MustTFromContext(ctx).Error(err)
				return err
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				t.Helper()
				if !assertions.New(t).So(errors.IsInvalidArgument(err), should.BeTrue) {
					t.Errorf("Received error: %s", err)
					return false
				}
				return true
			},
			BatchDeleteCalls:     0,
			BatchDeleteKeysCalls: 0,
		},
		{
			Name: "Existing device",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, &rights.Rights{
					ApplicationRights: *rights.NewMap(map[string]*ttnpb.Rights{
						unique.ID(test.Context(), registeredApplicationIDs): {
							Rights: []ttnpb.Right{
								ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE,
							},
						},
					}),
				})
			},
			BatchDeleteFunc: func(
				ctx context.Context,
				appIDs *ttnpb.ApplicationIdentifiers,
				deviceIDs []string,
			) ([]*ttnpb.EndDeviceIdentifiers, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(deviceIDs, should.HaveLength, 1)
				a.So(appIDs, should.Resemble, registeredApplicationIDs)
				a.So(deviceIDs[0], should.Equal, dev1.GetIds().DeviceId)
				return []*ttnpb.EndDeviceIdentifiers{
					dev1.Ids,
				}, nil
			},
			BatchDeleteKeysFunc: func(ctx context.Context, devIDs []*ttnpb.EndDeviceIdentifiers) error {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(devIDs, should.HaveLength, 1) {
					return fmt.Errorf("Invalid number of devices for BatchDeleteKeysFunc: %d", len(devIDs))
				}
				a.So(devIDs[0], should.Resemble, dev1.Ids)
				return nil
			},
			Request: &ttnpb.BatchDeleteEndDevicesRequest{
				ApplicationIds: registeredApplicationIDs,
				DeviceIds: []string{
					dev1.Ids.DeviceId,
				},
			},
			BatchDeleteCalls:     1,
			BatchDeleteKeysCalls: 1,
		},
		{
			Name: "One invalid device in batch",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, &rights.Rights{
					ApplicationRights: *rights.NewMap(map[string]*ttnpb.Rights{
						unique.ID(test.Context(), registeredApplicationIDs): {
							Rights: []ttnpb.Right{
								ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE,
							},
						},
					}),
				})
			},
			BatchDeleteFunc: func(
				ctx context.Context,
				appIDs *ttnpb.ApplicationIdentifiers,
				deviceIDs []string,
			) ([]*ttnpb.EndDeviceIdentifiers, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(deviceIDs, should.HaveLength, 3)
				a.So(appIDs, should.Resemble, registeredApplicationIDs)
				for _, devID := range deviceIDs {
					switch devID {
					case dev1.GetIds().DeviceId:
					case dev2.GetIds().DeviceId:
						t.Log("Known device ID")
					case "test-dev-unknown-id":
						t.Log("Ignore expected unknown device ID")
					default:
						t.Log("Unexpected device ID")
					}
				}
				return []*ttnpb.EndDeviceIdentifiers{
					dev1.Ids,
					dev2.Ids,
				}, nil
			},
			BatchDeleteKeysFunc: func(ctx context.Context, devIDs []*ttnpb.EndDeviceIdentifiers) error {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(devIDs, should.HaveLength, 2)
				if !a.So(devIDs, should.HaveLength, 2) {
					return fmt.Errorf("Invalid number of devices for BatchDeleteKeysFunc: %d", len(devIDs))
				}
				a.So(devIDs[0], should.Resemble, dev1.Ids)
				a.So(devIDs[1], should.Resemble, dev2.Ids)
				return nil
			},
			Request: &ttnpb.BatchDeleteEndDevicesRequest{
				ApplicationIds: registeredApplicationIDs,
				DeviceIds: []string{
					dev1.Ids.DeviceId,
					dev2.Ids.DeviceId,
					"test-dev-unknown-id",
				},
			},
			BatchDeleteCalls:     1,
			BatchDeleteKeysCalls: 1,
		},
		{
			Name: "Valid Batch",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, &rights.Rights{
					ApplicationRights: *rights.NewMap(map[string]*ttnpb.Rights{
						unique.ID(test.Context(), registeredApplicationIDs): {
							Rights: []ttnpb.Right{
								ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE,
							},
						},
					}),
				})
			},
			BatchDeleteFunc: func(
				ctx context.Context,
				appIDs *ttnpb.ApplicationIdentifiers,
				deviceIDs []string,
			) ([]*ttnpb.EndDeviceIdentifiers, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(deviceIDs, should.HaveLength, 3)
				a.So(appIDs, should.Resemble, registeredApplicationIDs)
				for _, devID := range deviceIDs {
					switch devID {
					case dev1.GetIds().DeviceId:
					case dev2.GetIds().DeviceId:
					case dev3.GetIds().DeviceId:
						// Known device ID
					default:
						t.Error("Unknown device ID: ", devID)
					}
				}
				return []*ttnpb.EndDeviceIdentifiers{
					dev1.Ids,
					dev2.Ids,
					dev3.Ids,
				}, nil
			},
			BatchDeleteKeysFunc: func(ctx context.Context, devIDs []*ttnpb.EndDeviceIdentifiers) error {
				a := assertions.New(test.MustTFromContext(ctx))
				if !a.So(devIDs, should.HaveLength, 3) {
					return fmt.Errorf("Invalid number of devices for BatchDeleteKeysFunc: %d", len(devIDs))
				}
				a.So(devIDs[0], should.Resemble, dev1.Ids)
				a.So(devIDs[1], should.Resemble, dev2.Ids)
				a.So(devIDs[2], should.Resemble, dev3.Ids)
				return nil
			},
			Request: &ttnpb.BatchDeleteEndDevicesRequest{
				ApplicationIds: registeredApplicationIDs,
				DeviceIds: []string{
					dev1.Ids.DeviceId,
					dev2.Ids.DeviceId,
					dev3.Ids.DeviceId,
				},
			},
			BatchDeleteCalls:     1,
			BatchDeleteKeysCalls: 1,
		},
	} {
		tc := tc
		test.RunSubtest(t, test.SubtestConfig{
			Name:     tc.Name,
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				t.Helper()
				var (
					batchDeleteCalls     uint64
					batchDeleteKeysCalls uint64
				)
				js := test.Must(New(
					componenttest.NewComponent(t, &component.Config{
						ServiceBase: config.ServiceBase{
							KeyVault: config.KeyVault{
								Provider: "static",
								Static: map[string][]byte{
									"test": {0x0, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf},
								},
							},
						},
					}),
					&Config{
						Devices: &MockDeviceRegistry{
							BatchDeleteFunc: func(
								ctx context.Context,
								appIDs *ttnpb.ApplicationIdentifiers,
								deviceIDs []string,
							) ([]*ttnpb.EndDeviceIdentifiers, error) {
								atomic.AddUint64(&batchDeleteCalls, 1)
								return tc.BatchDeleteFunc(ctx, appIDs, deviceIDs)
							},
						},
						Keys: &MockKeyRegistry{
							BatchDeleteFunc: func(ctx context.Context, devIDs []*ttnpb.EndDeviceIdentifiers) error {
								atomic.AddUint64(&batchDeleteKeysCalls, 1)
								return tc.BatchDeleteKeysFunc(ctx, devIDs)
							},
						},
						DevNonceLimit: defaultDevNonceLimit,
					},
				))
				js.AddContextFiller(tc.ContextFunc)
				js.AddContextFiller(func(ctx context.Context) context.Context {
					ctx, cancel := context.WithDeadline(ctx, time.Now().Add(Timeout))
					_ = cancel
					return ctx
				})
				js.AddContextFiller(func(ctx context.Context) context.Context {
					return test.ContextWithTB(ctx, t)
				})
				componenttest.StartComponent(t, js.Component)
				defer js.Close()
				ctx = js.FillContext(ctx)
				req := ttnpb.Clone(tc.Request)
				_, err := ttnpb.NewJsEndDeviceBatchRegistryClient(js.LoopbackConn()).Delete(ctx, req)
				a.So(batchDeleteCalls, should.Equal, tc.BatchDeleteCalls)
				a.So(batchDeleteKeysCalls, should.Equal, tc.BatchDeleteKeysCalls)
				if tc.ErrorAssertion != nil {
					a.So(tc.ErrorAssertion(t, err), should.BeTrue)
				} else {
					a.So(err, should.BeNil)
				}
			},
		})
	}
}
