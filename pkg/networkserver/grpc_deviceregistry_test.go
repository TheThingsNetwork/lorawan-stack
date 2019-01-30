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
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/mohae/deepcopy"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/errors"
	. "go.thethings.network/lorawan-stack/pkg/networkserver"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestDeviceRegistryGet(t *testing.T) {
	type getByIDCallKey struct{}

	ids := ttnpb.EndDeviceIdentifiers{
		DeviceID: DeviceID,
		ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
			ApplicationID: ApplicationID,
		},
	}

	for _, tc := range []struct {
		Name             string
		ContextFunc      func(context.Context) context.Context
		GetByIDFunc      func(context.Context, ttnpb.ApplicationIdentifiers, string, []string) (*ttnpb.EndDevice, error)
		Request          *ttnpb.GetEndDeviceRequest
		Device           *ttnpb.EndDevice
		ErrorAssertion   func(*testing.T, error) bool
		ContextAssertion func(context.Context) bool
	}{
		{
			Name: "No device read rights",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, ids.ApplicationIdentifiers): {
							Rights: []ttnpb.Right{
								ttnpb.RIGHT_GATEWAY_SETTINGS_BASIC,
							},
						},
					},
				})
			},
			GetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string) (*ttnpb.EndDevice, error) {
				test.MustTFromContext(ctx).Fatal("GetByIDFunc must not be called")
				panic("Unreachable")
			},
			Request: &ttnpb.GetEndDeviceRequest{
				EndDeviceIdentifiers: ids,
				FieldMask: pbtypes.FieldMask{
					Paths: []string{
						"test",
					},
				},
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				if !assertions.New(t).So(errors.IsPermissionDenied(err), should.BeTrue) {
					t.Errorf("Received error: %s", err)
					return false
				}
				return true
			},
			ContextAssertion: func(ctx context.Context) bool {
				a := assertions.New(test.MustTFromContext(ctx))
				return a.So(test.MustCounterFromContext(ctx, getByIDCallKey{}), should.Equal, 0)
			},
		},

		{
			Name: "Valid request",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, ids.ApplicationIdentifiers): {
							Rights: []ttnpb.Right{
								ttnpb.RIGHT_APPLICATION_DEVICES_READ,
							},
						},
					},
				})
			},
			GetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string) (*ttnpb.EndDevice, error) {
				defer test.MustIncrementContextCounter(ctx, getByIDCallKey{}, 1)
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(appID, should.Resemble, ids.ApplicationIdentifiers)
				a.So(devID, should.Equal, DeviceID)
				a.So(paths, should.HaveSameElementsDeep, []string{
					"test",
				})
				return &ttnpb.EndDevice{
					EndDeviceIdentifiers: ids,
				}, nil
			},
			Request: &ttnpb.GetEndDeviceRequest{
				EndDeviceIdentifiers: ids,
				FieldMask: pbtypes.FieldMask{
					Paths: []string{
						"test",
					},
				},
			},
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ids,
			},
			ContextAssertion: func(ctx context.Context) bool {
				a := assertions.New(test.MustTFromContext(ctx))
				return a.So(test.MustCounterFromContext(ctx, getByIDCallKey{}), should.Equal, 1)
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			ns := test.Must(New(
				component.MustNew(test.GetLogger(t), &component.Config{}),
				&Config{
					Devices: &MockDeviceRegistry{
						GetByIDFunc: tc.GetByIDFunc,
					},
					DeduplicationWindow: 42,
					CooldownWindow:      42,
					DownlinkTasks:       &MockDownlinkTaskQueue{},
				})).(*NetworkServer)

			ns.AddContextFiller(tc.ContextFunc)
			ns.AddContextFiller(func(ctx context.Context) context.Context {
				return test.ContextWithCounter(ctx, getByIDCallKey{})
			})
			ns.AddContextFiller(func(ctx context.Context) context.Context {
				ctx, cancel := context.WithDeadline(ctx, time.Now().Add(Timeout))
				_ = cancel
				return ctx
			})
			ns.AddContextFiller(func(ctx context.Context) context.Context {
				return test.ContextWithT(ctx, t)
			})
			test.Must(nil, ns.Start())
			defer ns.Close()

			req := deepcopy.Copy(tc.Request).(*ttnpb.GetEndDeviceRequest)

			dev, err := ttnpb.NewNsEndDeviceRegistryClient(ns.LoopbackConn()).Get(test.Context(), req)
			if tc.ErrorAssertion != nil {
				a.So(tc.ErrorAssertion(t, err), should.BeTrue)
				a.So(dev, should.BeNil)
			} else {
				a.So(err, should.BeNil)
				a.So(dev, should.Resemble, tc.Device)
			}
			a.So(req, should.Resemble, tc.Request)
		})
	}
}

func TestDeviceRegistrySet(t *testing.T) {
	type setByIDCallKey struct{}

	ids := ttnpb.EndDeviceIdentifiers{
		DeviceID: DeviceID,
		ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
			ApplicationID: ApplicationID,
		},
	}

	for _, tc := range []struct {
		Name             string
		ContextFunc      func(context.Context) context.Context
		SetByIDFunc      func(context.Context, ttnpb.ApplicationIdentifiers, string, []string, func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error)
		Request          *ttnpb.SetEndDeviceRequest
		Device           *ttnpb.EndDevice
		ErrorAssertion   func(*testing.T, error) bool
		ContextAssertion func(context.Context) bool
	}{
		{
			Name: "No device write rights",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, ids.ApplicationIdentifiers): {
							Rights: []ttnpb.Right{
								ttnpb.RIGHT_GATEWAY_SETTINGS_BASIC,
							},
						},
					},
				})
			},
			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				test.MustTFromContext(ctx).Fatal("SetByIDFunc must not be called")
				panic("Unreachable")
			},
			Request: &ttnpb.SetEndDeviceRequest{
				Device: ttnpb.EndDevice{
					EndDeviceIdentifiers: ids,
					SupportsJoin:         true,
					MACSettings:          &ttnpb.MACSettings{},
				},
				FieldMask: pbtypes.FieldMask{
					Paths: []string{
						"frequency_plan_id",
						"lorawan_phy_version",
						"lorawan_version",
						"mac_settings.use_adr",
						"resets_f_cnt",
						"resets_join_nonces",
						"supports_class_b",
						"supports_class_c",
						"supports_join",
						"uses_32_bit_f_cnt",
					},
				},
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				if !assertions.New(t).So(errors.IsPermissionDenied(err), should.BeTrue) {
					t.Errorf("Received error: %s", err)
					return false
				}
				return true
			},
			ContextAssertion: func(ctx context.Context) bool {
				a := assertions.New(test.MustTFromContext(ctx))
				return a.So(test.MustCounterFromContext(ctx, setByIDCallKey{}), should.Equal, 0)
			},
		},

		{
			Name: "Create OTAA device",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, ids.ApplicationIdentifiers): {
							Rights: []ttnpb.Right{
								ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
							},
						},
					},
				})
			},
			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				defer test.MustIncrementContextCounter(ctx, setByIDCallKey{}, 1)
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(appID, should.Resemble, ids.ApplicationIdentifiers)
				a.So(devID, should.Equal, DeviceID)
				a.So(gets, should.HaveSameElementsDeep, []string{
					"frequency_plan_id",
					"lorawan_phy_version",
					"lorawan_version",
					"mac_settings.use_adr",
					"resets_f_cnt",
					"resets_join_nonces",
					"supports_class_b",
					"supports_class_c",
					"supports_join",
					"uses_32_bit_f_cnt",
				})

				dev, sets, err := f(nil)
				a.So(err, should.BeNil)
				a.So(sets, should.HaveSameElementsDeep, []string{
					"frequency_plan_id",
					"lorawan_phy_version",
					"lorawan_version",
					"mac_settings.adr_margin",
					"mac_settings.class_b_timeout",
					"mac_settings.class_c_timeout",
					"mac_settings.use_adr",
					"resets_f_cnt",
					"resets_join_nonces",
					"supports_class_b",
					"supports_class_c",
					"supports_join",
					"uses_32_bit_f_cnt",
				})
				a.So(dev, should.Resemble, &ttnpb.EndDevice{
					EndDeviceIdentifiers: ids,
					SupportsJoin:         true,
					MACSettings: &ttnpb.MACSettings{
						ADRMargin:     15,
						ClassBTimeout: time.Minute,
						ClassCTimeout: 10 * time.Second,
					},
				})
				return &ttnpb.EndDevice{
					EndDeviceIdentifiers: ids,
					SupportsJoin:         true,
					MACSettings: &ttnpb.MACSettings{
						ADRMargin:     15,
						ClassBTimeout: time.Minute,
						ClassCTimeout: 10 * time.Second,
					},
				}, nil
			},
			Request: &ttnpb.SetEndDeviceRequest{
				Device: ttnpb.EndDevice{
					EndDeviceIdentifiers: ids,
					SupportsJoin:         true,
					MACSettings:          &ttnpb.MACSettings{},
				},
				FieldMask: pbtypes.FieldMask{
					Paths: []string{
						"frequency_plan_id",
						"lorawan_phy_version",
						"lorawan_version",
						"mac_settings.use_adr",
						"resets_f_cnt",
						"resets_join_nonces",
						"supports_class_b",
						"supports_class_c",
						"supports_join",
						"uses_32_bit_f_cnt",
					},
				},
			},
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ids,
				SupportsJoin:         true,
				MACSettings: &ttnpb.MACSettings{
					ADRMargin:     15,
					ClassBTimeout: time.Minute,
					ClassCTimeout: 10 * time.Second,
				},
			},
			ContextAssertion: func(ctx context.Context) bool {
				a := assertions.New(test.MustTFromContext(ctx))
				return a.So(test.MustCounterFromContext(ctx, setByIDCallKey{}), should.Equal, 1)
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			ns := test.Must(New(
				component.MustNew(test.GetLogger(t), &component.Config{}),
				&Config{
					Devices: &MockDeviceRegistry{
						SetByIDFunc: tc.SetByIDFunc,
					},
					DeduplicationWindow: 42,
					CooldownWindow:      42,
					DownlinkTasks:       &MockDownlinkTaskQueue{},
				})).(*NetworkServer)

			ns.AddContextFiller(tc.ContextFunc)
			ns.AddContextFiller(func(ctx context.Context) context.Context {
				return test.ContextWithCounter(ctx, setByIDCallKey{})
			})
			ns.AddContextFiller(func(ctx context.Context) context.Context {
				ctx, cancel := context.WithDeadline(ctx, time.Now().Add(Timeout))
				_ = cancel
				return ctx
			})
			ns.AddContextFiller(func(ctx context.Context) context.Context {
				return test.ContextWithT(ctx, t)
			})
			test.Must(nil, ns.Start())
			defer ns.Close()

			req := deepcopy.Copy(tc.Request).(*ttnpb.SetEndDeviceRequest)

			dev, err := ttnpb.NewNsEndDeviceRegistryClient(ns.LoopbackConn()).Set(test.Context(), req)
			if tc.ErrorAssertion != nil {
				a.So(tc.ErrorAssertion(t, err), should.BeTrue)
				a.So(dev, should.BeNil)
			} else {
				a.So(err, should.BeNil)
				a.So(dev, should.Resemble, tc.Device)
			}
			a.So(req, should.Resemble, tc.Request)
		})
	}
}

func TestDeviceRegistryDelete(t *testing.T) {
	type setByIDCallKey struct{}

	ids := ttnpb.EndDeviceIdentifiers{
		DeviceID: DeviceID,
		ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
			ApplicationID: ApplicationID,
		},
	}

	for _, tc := range []struct {
		Name             string
		ContextFunc      func(context.Context) context.Context
		SetByIDFunc      func(context.Context, ttnpb.ApplicationIdentifiers, string, []string, func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error)
		Request          *ttnpb.EndDeviceIdentifiers
		ErrorAssertion   func(*testing.T, error) bool
		ContextAssertion func(context.Context) bool
	}{
		{
			Name: "No device write rights",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, ids.ApplicationIdentifiers): {
							Rights: []ttnpb.Right{
								ttnpb.RIGHT_GATEWAY_SETTINGS_BASIC,
							},
						},
					},
				})
			},
			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				test.MustTFromContext(ctx).Fatal("SetByIDFunc must not be called")
				panic("Unreachable")
			},
			Request: deepcopy.Copy(&ids).(*ttnpb.EndDeviceIdentifiers),
			ErrorAssertion: func(t *testing.T, err error) bool {
				if !assertions.New(t).So(errors.IsPermissionDenied(err), should.BeTrue) {
					t.Errorf("Received error: %s", err)
					return false
				}
				return true
			},
			ContextAssertion: func(ctx context.Context) bool {
				a := assertions.New(test.MustTFromContext(ctx))
				return a.So(test.MustCounterFromContext(ctx, setByIDCallKey{}), should.Equal, 0)
			},
		},

		{
			Name: "Non-existing device",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, ids.ApplicationIdentifiers): {
							Rights: []ttnpb.Right{
								ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
							},
						},
					},
				})
			},
			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				defer test.MustIncrementContextCounter(ctx, setByIDCallKey{}, 1)
				t := test.MustTFromContext(ctx)
				a := assertions.New(t)
				a.So(appID, should.Resemble, ids.ApplicationIdentifiers)
				a.So(devID, should.Equal, DeviceID)
				a.So(gets, should.BeNil)

				dev, sets, err := f(nil)
				a.So(err, should.BeNil)
				a.So(sets, should.BeNil)
				a.So(dev, should.BeNil)
				return nil, nil
			},
			Request: deepcopy.Copy(&ids).(*ttnpb.EndDeviceIdentifiers),
			ContextAssertion: func(ctx context.Context) bool {
				a := assertions.New(test.MustTFromContext(ctx))
				return a.So(test.MustCounterFromContext(ctx, setByIDCallKey{}), should.Equal, 1)
			},
		},

		{
			Name: "Existing device",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, ids.ApplicationIdentifiers): {
							Rights: []ttnpb.Right{
								ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
							},
						},
					},
				})
			},
			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				defer test.MustIncrementContextCounter(ctx, setByIDCallKey{}, 1)
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(appID, should.Resemble, ids.ApplicationIdentifiers)
				a.So(devID, should.Equal, DeviceID)
				a.So(paths, should.BeNil)

				dev, sets, err := f(&ttnpb.EndDevice{
					EndDeviceIdentifiers: ids,
				})
				a.So(err, should.BeNil)
				a.So(sets, should.BeNil)
				a.So(dev, should.BeNil)
				return nil, nil
			},
			Request: deepcopy.Copy(&ids).(*ttnpb.EndDeviceIdentifiers),
			ContextAssertion: func(ctx context.Context) bool {
				a := assertions.New(test.MustTFromContext(ctx))
				return a.So(test.MustCounterFromContext(ctx, setByIDCallKey{}), should.Equal, 1)
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			ns := test.Must(New(
				component.MustNew(test.GetLogger(t), &component.Config{}),
				&Config{
					Devices: &MockDeviceRegistry{
						SetByIDFunc: tc.SetByIDFunc,
					},
					DeduplicationWindow: 42,
					CooldownWindow:      42,
					DownlinkTasks:       &MockDownlinkTaskQueue{},
				})).(*NetworkServer)

			ns.AddContextFiller(tc.ContextFunc)
			ns.AddContextFiller(func(ctx context.Context) context.Context {
				return test.ContextWithCounter(ctx, setByIDCallKey{})
			})
			ns.AddContextFiller(func(ctx context.Context) context.Context {
				ctx, cancel := context.WithDeadline(ctx, time.Now().Add(Timeout))
				_ = cancel
				return ctx
			})
			ns.AddContextFiller(func(ctx context.Context) context.Context {
				return test.ContextWithT(ctx, t)
			})
			test.Must(nil, ns.Start())
			defer ns.Close()

			req := deepcopy.Copy(tc.Request).(*ttnpb.EndDeviceIdentifiers)

			res, err := ttnpb.NewNsEndDeviceRegistryClient(ns.LoopbackConn()).Delete(test.Context(), req)
			if tc.ErrorAssertion != nil {
				a.So(tc.ErrorAssertion(t, err), should.BeTrue)
				a.So(res, should.BeNil)
			} else {
				a.So(err, should.BeNil)
				a.So(res, should.Resemble, ttnpb.Empty)
			}
			a.So(req, should.Resemble, tc.Request)
		})
	}
}
