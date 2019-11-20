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

package applicationserver_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/mohae/deepcopy"
	"github.com/smartystreets/assertions"
	. "go.thethings.network/lorawan-stack/pkg/applicationserver"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/pkg/component/test"
	"go.thethings.network/lorawan-stack/pkg/config"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/unique"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestDeviceRegistryGet(t *testing.T) {
	registeredApplicationID := "foo-application"
	registeredDeviceID := "foo-device"
	registeredDevice := &ttnpb.EndDevice{
		EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
			ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
				ApplicationID: registeredApplicationID,
			},
			DeviceID: registeredDeviceID,
		},
		Formatters: &ttnpb.MessagePayloadFormatters{
			UpFormatter:   ttnpb.PayloadFormatter_FORMATTER_REPOSITORY,
			DownFormatter: ttnpb.PayloadFormatter_FORMATTER_REPOSITORY,
		},
		Session: &ttnpb.Session{
			SessionKeys: ttnpb.SessionKeys{
				AppSKey: &ttnpb.KeyEnvelope{
					KEKLabel:     "test",
					EncryptedKey: []byte{0x96, 0x77, 0x8b, 0x25, 0xae, 0x6c, 0xa4, 0x35, 0xf9, 0x2b, 0x5b, 0x97, 0xc0, 0x50, 0xae, 0xd2, 0x46, 0x8a, 0xb8, 0xa1, 0x7a, 0xd8, 0x4e, 0x5d},
				},
			},
		},
	}
	registeredKEKs := map[string][]byte{
		"test": {0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17},
	}

	expectedSession := &ttnpb.Session{
		SessionKeys: ttnpb.SessionKeys{
			AppSKey: &ttnpb.KeyEnvelope{
				Key: aes128KeyPtr(types.AES128Key{0x0, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}),
			},
		},
	}

	nilDeviceAssertion := func(t *testing.T, dev *ttnpb.EndDevice) bool {
		t.Helper()
		return assertions.New(t).So(dev, should.BeNil)
	}

	nilErrorAssertion := func(t *testing.T, err error) bool {
		t.Helper()
		return assertions.New(t).So(err, should.BeNil)
	}
	permissionDeniedErrorAssertion := func(t *testing.T, err error) bool {
		t.Helper()
		return assertions.New(t).So(errors.IsPermissionDenied(err), should.BeTrue)
	}
	notFoundErrorAssertion := func(t *testing.T, err error) bool {
		t.Helper()
		return assertions.New(t).So(errors.IsNotFound(err), should.BeTrue)
	}

	for _, tc := range []struct {
		Name            string
		ContextFunc     func(context.Context) context.Context
		GetFunc         func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, paths []string) (*ttnpb.EndDevice, error)
		DeviceRequest   *ttnpb.GetEndDeviceRequest
		DeviceAssertion func(*testing.T, *ttnpb.EndDevice) bool
		ErrorAssertion  func(*testing.T, error) bool
		GetCalls        uint64
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
			GetFunc: func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, paths []string) (*ttnpb.EndDevice, error) {
				test.MustTFromContext(ctx).Errorf("GetFunc must not be called")
				return nil, errors.New("GetFunc must not be called")
			},
			DeviceRequest: &ttnpb.GetEndDeviceRequest{
				EndDeviceIdentifiers: registeredDevice.EndDeviceIdentifiers,
				FieldMask: pbtypes.FieldMask{
					Paths: []string{"formatters"},
				},
			},
			DeviceAssertion: nilDeviceAssertion,
			ErrorAssertion:  permissionDeniedErrorAssertion,
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
			GetFunc: func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, paths []string) (*ttnpb.EndDevice, error) {
				test.MustTFromContext(ctx).Errorf("GetFunc must not be called")
				return nil, errors.New("GetFunc must not be called")
			},
			DeviceRequest: &ttnpb.GetEndDeviceRequest{
				EndDeviceIdentifiers: registeredDevice.EndDeviceIdentifiers,
				FieldMask: pbtypes.FieldMask{
					Paths: []string{"formatters"},
				},
			},
			DeviceAssertion: nilDeviceAssertion,
			ErrorAssertion:  permissionDeniedErrorAssertion,
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
			GetFunc: func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, paths []string) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(ids, should.Resemble, ids)
				a.So(paths, should.HaveSameElementsDeep, []string{
					"formatters",
				})
				return nil, errNotFound
			},
			DeviceRequest: &ttnpb.GetEndDeviceRequest{
				EndDeviceIdentifiers: registeredDevice.EndDeviceIdentifiers,
				FieldMask: pbtypes.FieldMask{
					Paths: []string{"formatters"},
				},
			},
			DeviceAssertion: nilDeviceAssertion,
			ErrorAssertion:  notFoundErrorAssertion,
			GetCalls:        1,
		},

		{
			Name: "Get formatters",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(test.Context(), ttnpb.ApplicationIdentifiers{ApplicationID: registeredApplicationID}): ttnpb.RightsFrom(
							ttnpb.RIGHT_APPLICATION_DEVICES_READ,
						),
					},
				})
			},
			GetFunc: func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, paths []string) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(ids, should.Resemble, ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
						ApplicationID: registeredApplicationID,
					},
					DeviceID: registeredDeviceID,
				})
				a.So(paths, should.HaveSameElementsDeep, []string{
					"formatters",
				})
				return deepcopy.Copy(&ttnpb.EndDevice{
					EndDeviceIdentifiers: registeredDevice.EndDeviceIdentifiers,
					Formatters:           registeredDevice.Formatters,
				}).(*ttnpb.EndDevice), nil
			},
			DeviceRequest: &ttnpb.GetEndDeviceRequest{
				EndDeviceIdentifiers: registeredDevice.EndDeviceIdentifiers,
				FieldMask: pbtypes.FieldMask{
					Paths: []string{"formatters"},
				},
			},
			DeviceAssertion: func(t *testing.T, dev *ttnpb.EndDevice) bool {
				return assertions.New(t).So(dev, should.Resemble, &ttnpb.EndDevice{
					EndDeviceIdentifiers: registeredDevice.EndDeviceIdentifiers,
					Formatters:           registeredDevice.Formatters,
				})
			},
			ErrorAssertion: nilErrorAssertion,
			GetCalls:       1,
		},

		{
			Name: "Get formatters, session/no key rights",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(test.Context(), ttnpb.ApplicationIdentifiers{ApplicationID: registeredApplicationID}): ttnpb.RightsFrom(
							ttnpb.RIGHT_APPLICATION_DEVICES_READ,
						),
					},
				})
			},
			GetFunc: func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, paths []string) (*ttnpb.EndDevice, error) {
				test.MustTFromContext(ctx).Errorf("GetFunc must not be called")
				return nil, errors.New("GetFunc must not be called")
			},
			DeviceRequest: &ttnpb.GetEndDeviceRequest{
				EndDeviceIdentifiers: registeredDevice.EndDeviceIdentifiers,
				FieldMask: pbtypes.FieldMask{
					Paths: []string{"formatters", "session"},
				},
			},
			DeviceAssertion: nilDeviceAssertion,
			ErrorAssertion:  permissionDeniedErrorAssertion,
		},

		{
			Name: "Get formatters,session/has rights",
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
			GetFunc: func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, paths []string) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(ids, should.Resemble, ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
						ApplicationID: registeredApplicationID,
					},
					DeviceID: registeredDeviceID,
				})
				a.So(paths, should.HaveSameElementsDeep, []string{
					"formatters",
					"session",
				})
				return deepcopy.Copy(registeredDevice).(*ttnpb.EndDevice), nil
			},
			DeviceRequest: &ttnpb.GetEndDeviceRequest{
				EndDeviceIdentifiers: registeredDevice.EndDeviceIdentifiers,
				FieldMask: pbtypes.FieldMask{
					Paths: []string{"formatters", "session"},
				},
			},
			DeviceAssertion: func(t *testing.T, dev *ttnpb.EndDevice) bool {
				return assertions.New(t).So(dev, should.Resemble, &ttnpb.EndDevice{
					EndDeviceIdentifiers: registeredDevice.EndDeviceIdentifiers,
					Formatters:           registeredDevice.Formatters,
					Session:              expectedSession,
				})
			},
			ErrorAssertion: nilErrorAssertion,
			GetCalls:       1,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			var getCalls uint64

			as := test.Must(New(
				componenttest.NewComponent(t, &component.Config{
					ServiceBase: config.ServiceBase{
						KeyVault: config.KeyVault{
							Provider: "static",
							Static:   registeredKEKs,
						},
					},
				}),
				&Config{
					LinkMode: "explicit",
					Devices: &MockDeviceRegistry{
						GetFunc: func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, paths []string) (*ttnpb.EndDevice, error) {
							atomic.AddUint64(&getCalls, 1)
							return tc.GetFunc(ctx, ids, paths)
						},
					},
				})).(*ApplicationServer)

			as.AddContextFiller(tc.ContextFunc)
			as.AddContextFiller(func(ctx context.Context) context.Context {
				ctx, cancel := context.WithDeadline(ctx, time.Now().Add(Timeout))
				_ = cancel
				return ctx
			})
			as.AddContextFiller(func(ctx context.Context) context.Context {
				return test.ContextWithT(ctx, t)
			})
			componenttest.StartComponent(t, as.Component)
			defer as.Close()

			ctx := as.FillContext(test.Context())
			req := deepcopy.Copy(tc.DeviceRequest).(*ttnpb.GetEndDeviceRequest)

			dev, err := ttnpb.NewAsEndDeviceRegistryClient(as.LoopbackConn()).Get(ctx, req)
			a.So(req, should.Resemble, tc.DeviceRequest)
			a.So(getCalls, should.Equal, tc.GetCalls)
			if a.So(tc.ErrorAssertion(t, err), should.BeTrue) {
				a.So(tc.DeviceAssertion(t, dev), should.BeTrue)
			}
		})
	}
}

func TestDeviceRegistrySet(t *testing.T) {
	registeredApplicationID := "foo-application"
	registeredDeviceID := "foo-device"
	registeredDevice := &ttnpb.EndDevice{
		EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
			ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
				ApplicationID: registeredApplicationID,
			},
			DeviceID: registeredDeviceID,
		},
		Formatters: &ttnpb.MessagePayloadFormatters{
			UpFormatter:   ttnpb.PayloadFormatter_FORMATTER_REPOSITORY,
			DownFormatter: ttnpb.PayloadFormatter_FORMATTER_REPOSITORY,
		},
	}
	for _, tc := range []struct {
		Name            string
		ContextFunc     func(context.Context) context.Context
		SetFunc         func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, paths []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error)
		DeviceRequest   *ttnpb.SetEndDeviceRequest
		ErrorAssertion  func(*testing.T, error) bool
		DeviceAssertion func(*testing.T, *ttnpb.EndDevice) bool
		SetCalls        uint64
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
			SetFunc: func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, paths []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				test.MustTFromContext(ctx).Errorf("SetFunc must not be called")
				return nil, errors.New("SetFunc must not be called")
			},
			DeviceRequest: &ttnpb.SetEndDeviceRequest{
				EndDevice: deepcopy.Copy(*registeredDevice).(ttnpb.EndDevice),
				FieldMask: pbtypes.FieldMask{
					Paths: []string{"formatters"},
				},
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
			SetFunc: func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, paths []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				test.MustTFromContext(ctx).Errorf("SetFunc must not be called")
				return nil, errors.New("SetFunc must not be called")
			},
			DeviceRequest: &ttnpb.SetEndDeviceRequest{
				EndDevice: deepcopy.Copy(*registeredDevice).(ttnpb.EndDevice),
				FieldMask: pbtypes.FieldMask{
					Paths: []string{"formatters"},
				},
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				a := assertions.New(t)
				return a.So(errors.IsPermissionDenied(err), should.BeTrue)
			},
		},

		{
			Name: "Create",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(test.Context(), ttnpb.ApplicationIdentifiers{ApplicationID: registeredApplicationID}): {
							Rights: []ttnpb.Right{
								ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
							},
						},
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
					},
					Formatters: &ttnpb.MessagePayloadFormatters{
						UpFormatter:   ttnpb.PayloadFormatter_FORMATTER_CAYENNELPP,
						DownFormatter: ttnpb.PayloadFormatter_FORMATTER_CAYENNELPP,
					},
				},
				FieldMask: pbtypes.FieldMask{
					Paths: []string{"formatters"},
				},
			},
			SetFunc: func(ctx context.Context, deviceIds ttnpb.EndDeviceIdentifiers, gets []string, cb func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(deviceIds, should.Resemble, ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
						ApplicationID: registeredApplicationID,
					},
					DeviceID: "new-device",
				})
				a.So(gets, should.HaveSameElementsDeep, []string{
					"formatters",
				})
				dev, sets, err := cb(nil)
				a.So(sets, should.HaveSameElementsDeep, []string{
					"formatters",
					"ids.application_ids",
					"ids.device_id",
				})
				a.So(dev, should.Resemble, &ttnpb.EndDevice{
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
							ApplicationID: registeredApplicationID,
						},
						DeviceID: "new-device",
					},
					Formatters: &ttnpb.MessagePayloadFormatters{
						UpFormatter:   ttnpb.PayloadFormatter_FORMATTER_CAYENNELPP,
						DownFormatter: ttnpb.PayloadFormatter_FORMATTER_CAYENNELPP,
					},
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
					},
					Formatters: &ttnpb.MessagePayloadFormatters{
						UpFormatter:   ttnpb.PayloadFormatter_FORMATTER_CAYENNELPP,
						DownFormatter: ttnpb.PayloadFormatter_FORMATTER_CAYENNELPP,
					},
				})
			},
			SetCalls: 1,
		},

		{
			Name: "Set",
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
				EndDevice: deepcopy.Copy(*registeredDevice).(ttnpb.EndDevice),
				FieldMask: pbtypes.FieldMask{
					Paths: []string{"formatters"},
				},
			},
			SetFunc: func(ctx context.Context, deviceIds ttnpb.EndDeviceIdentifiers, gets []string, cb func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(deviceIds, should.Resemble, registeredDevice.EndDeviceIdentifiers)
				a.So(gets, should.HaveSameElementsDeep, []string{
					"formatters",
				})
				dev, sets, err := cb(deepcopy.Copy(registeredDevice).(*ttnpb.EndDevice))
				a.So(sets, should.HaveSameElementsDeep, []string{
					"formatters",
				})
				a.So(dev, should.Resemble, registeredDevice)
				return dev, err
			},
			SetCalls: 1,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			var setCalls uint64

			as := test.Must(New(componenttest.NewComponent(t, &component.Config{}),
				&Config{
					LinkMode: "explicit",
					Devices: &MockDeviceRegistry{
						SetFunc: func(ctx context.Context, deviceIds ttnpb.EndDeviceIdentifiers, paths []string, cb func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
							atomic.AddUint64(&setCalls, 1)
							return tc.SetFunc(ctx, deviceIds, paths, cb)
						},
					},
				})).(*ApplicationServer)

			as.AddContextFiller(tc.ContextFunc)
			as.AddContextFiller(func(ctx context.Context) context.Context {
				ctx, cancel := context.WithDeadline(ctx, time.Now().Add(Timeout))
				_ = cancel
				return ctx
			})
			as.AddContextFiller(func(ctx context.Context) context.Context {
				return test.ContextWithT(ctx, t)
			})
			componenttest.StartComponent(t, as.Component)
			defer as.Close()

			ctx := as.FillContext(test.Context())
			req := deepcopy.Copy(tc.DeviceRequest).(*ttnpb.SetEndDeviceRequest)

			dev, err := ttnpb.NewAsEndDeviceRegistryClient(as.LoopbackConn()).Set(ctx, req)
			a.So(setCalls, should.Equal, tc.SetCalls)
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
	registeredDevice := &ttnpb.EndDevice{
		EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
			ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
				ApplicationID: registeredApplicationID,
			},
			DeviceID: registeredDeviceID,
		},
	}
	for _, tc := range []struct {
		Name           string
		ContextFunc    func(context.Context) context.Context
		SetFunc        func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, paths []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error)
		DeviceRequest  *ttnpb.EndDeviceIdentifiers
		ErrorAssertion func(*testing.T, error) bool
		SetCalls       uint64
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
			SetFunc: func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, paths []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				test.MustTFromContext(ctx).Errorf("SetFunc must not be called")
				return nil, errors.New("SetFunc must not be called")
			},
			DeviceRequest: deepcopy.Copy(&registeredDevice.EndDeviceIdentifiers).(*ttnpb.EndDeviceIdentifiers),
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
			SetFunc: func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, paths []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				test.MustTFromContext(ctx).Errorf("SetFunc must not be called")
				return nil, errors.New("SetFunc must not be called")
			},
			DeviceRequest: deepcopy.Copy(&registeredDevice.EndDeviceIdentifiers).(*ttnpb.EndDeviceIdentifiers),
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
							ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
						),
					},
				})
			},
			SetFunc: func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, paths []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				t := test.MustTFromContext(ctx)
				a := assertions.New(t)
				a.So(ids, should.Resemble, registeredDevice.EndDeviceIdentifiers)
				dev, sets, err := f(nil)
				a.So(err, should.BeNil)
				a.So(sets, should.BeNil)
				a.So(dev, should.BeNil)
				return nil, nil
			},
			DeviceRequest: deepcopy.Copy(&registeredDevice.EndDeviceIdentifiers).(*ttnpb.EndDeviceIdentifiers),
			SetCalls:      1,
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
			SetFunc: func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, paths []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				t := test.MustTFromContext(ctx)
				a := assertions.New(t)
				a.So(ids, should.Resemble, registeredDevice.EndDeviceIdentifiers)
				dev, sets, err := f(registeredDevice)
				a.So(err, should.BeNil)
				a.So(sets, should.BeNil)
				a.So(dev, should.BeNil)
				return nil, nil
			},
			DeviceRequest: deepcopy.Copy(&registeredDevice.EndDeviceIdentifiers).(*ttnpb.EndDeviceIdentifiers),
			SetCalls:      1,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			var setCalls uint64

			as := test.Must(New(componenttest.NewComponent(t, &component.Config{}),
				&Config{
					LinkMode: "explicit",
					Devices: &MockDeviceRegistry{
						SetFunc: func(ctx context.Context, deviceIds ttnpb.EndDeviceIdentifiers, paths []string, cb func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
							atomic.AddUint64(&setCalls, 1)
							return tc.SetFunc(ctx, deviceIds, paths, cb)
						},
					},
				})).(*ApplicationServer)

			as.AddContextFiller(tc.ContextFunc)
			as.AddContextFiller(func(ctx context.Context) context.Context {
				ctx, cancel := context.WithDeadline(ctx, time.Now().Add(Timeout))
				_ = cancel
				return ctx
			})
			as.AddContextFiller(func(ctx context.Context) context.Context {
				return test.ContextWithT(ctx, t)
			})
			componenttest.StartComponent(t, as.Component)
			defer as.Close()

			ctx := as.FillContext(test.Context())
			req := deepcopy.Copy(tc.DeviceRequest).(*ttnpb.EndDeviceIdentifiers)

			_, err := ttnpb.NewAsEndDeviceRegistryClient(as.LoopbackConn()).Delete(ctx, req)
			a.So(setCalls, should.Equal, tc.SetCalls)
			if tc.ErrorAssertion != nil {
				a.So(tc.ErrorAssertion(t, err), should.BeTrue)
			} else {
				a.So(err, should.BeNil)
			}
			a.So(req, should.Resemble, tc.DeviceRequest)
		})
	}
}
