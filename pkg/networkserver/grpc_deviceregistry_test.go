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
	"go.thethings.network/lorawan-stack/pkg/frequencyplans"
	. "go.thethings.network/lorawan-stack/pkg/networkserver"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
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
				},
				FieldMask: pbtypes.FieldMask{
					Paths: []string{
						"frequency_plan_id",
						"lorawan_phy_version",
						"lorawan_version",
						"mac_settings.supports_32_bit_f_cnt",
						"mac_settings.use_adr",
						"resets_f_cnt",
						"resets_join_nonces",
						"supports_class_b",
						"supports_class_c",
						"supports_join",
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
					"mac_settings.adr_margin",
					"resets_f_cnt",
					"resets_join_nonces",
					"supports_class_b",
					"supports_class_c",
					"supports_join",
				})

				dev, sets, err := f(nil)
				a.So(err, should.BeNil)
				a.So(sets, should.HaveSameElementsDeep, []string{
					"frequency_plan_id",
					"lorawan_phy_version",
					"lorawan_version",
					"mac_settings.adr_margin",
					"resets_f_cnt",
					"resets_join_nonces",
					"supports_class_b",
					"supports_class_c",
					"supports_join",
				})
				a.So(dev, should.Resemble, &ttnpb.EndDevice{
					EndDeviceIdentifiers: ids,
					FrequencyPlanID:      test.EUFrequencyPlanID,
					LoRaWANPHYVersion:    ttnpb.PHY_V1_0,
					LoRaWANVersion:       ttnpb.MAC_V1_0,
					SupportsJoin:         true,
					MACSettings: &ttnpb.MACSettings{
						ADRMargin: &pbtypes.FloatValue{Value: 4},
					},
				})
				return dev, nil
			},
			Request: &ttnpb.SetEndDeviceRequest{
				Device: ttnpb.EndDevice{
					EndDeviceIdentifiers: ids,
					FrequencyPlanID:      test.EUFrequencyPlanID,
					LoRaWANPHYVersion:    ttnpb.PHY_V1_0,
					LoRaWANVersion:       ttnpb.MAC_V1_0,
					SupportsJoin:         true,
					MACSettings: &ttnpb.MACSettings{
						ADRMargin: &pbtypes.FloatValue{Value: 4},
					},
				},
				FieldMask: pbtypes.FieldMask{
					Paths: []string{
						"frequency_plan_id",
						"lorawan_phy_version",
						"lorawan_version",
						"mac_settings.adr_margin",
						"resets_f_cnt",
						"resets_join_nonces",
						"supports_class_b",
						"supports_class_c",
						"supports_join",
					},
				},
			},
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ids,
				FrequencyPlanID:      test.EUFrequencyPlanID,
				LoRaWANPHYVersion:    ttnpb.PHY_V1_0,
				LoRaWANVersion:       ttnpb.MAC_V1_0,
				SupportsJoin:         true,
				MACSettings: &ttnpb.MACSettings{
					ADRMargin: &pbtypes.FloatValue{Value: 4},
				},
			},
			ContextAssertion: func(ctx context.Context) bool {
				a := assertions.New(test.MustTFromContext(ctx))
				return a.So(test.MustCounterFromContext(ctx, setByIDCallKey{}), should.Equal, 1)
			},
		},

		{
			// https://github.com/TheThingsNetwork/lorawan-stack/issues/104#issuecomment-465074076
			Name: "Create OTAA device with existing session",
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
					"mac_settings.supports_32_bit_f_cnt",
					"resets_f_cnt",
					"resets_join_nonces",
					"root_keys.app_key.key",
					"session.dev_addr",
					"session.keys.app_s_key.key",
					"session.keys.f_nwk_s_int_key.key",
					"session.last_f_cnt_up",
					"session.last_n_f_cnt_down",
					"session.started_at",
					"supports_join",
				})

				dev, sets, err := f(nil)
				a.So(err, should.BeNil)
				a.So(sets, should.HaveSameElementsDeep, []string{
					"frequency_plan_id",
					"lorawan_phy_version",
					"lorawan_version",
					"mac_settings.supports_32_bit_f_cnt",
					"mac_settings.use_adr",
					"mac_state",
					"resets_f_cnt",
					"resets_join_nonces",
					"root_keys.app_key.key",
					"session.dev_addr",
					"session.keys.app_s_key.key",
					"session.keys.f_nwk_s_int_key.key",
					"session.keys.nwk_s_enc_key.kek_label",
					"session.keys.nwk_s_enc_key.key",
					"session.keys.s_nwk_s_int_key.kek_label",
					"session.keys.s_nwk_s_int_key.key",
					"session.last_f_cnt_up",
					"session.last_n_f_cnt_down",
					"session.started_at",
					"supports_join",
				})

				expected := &ttnpb.EndDevice{
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						DeviceID:               ids.DeviceID,
						ApplicationIdentifiers: ids.ApplicationIdentifiers,
						DevAddr:                &types.DevAddr{0x01, 0x0b, 0x60, 0x0c},
					},
					FrequencyPlanID:   test.EUFrequencyPlanID,
					LoRaWANPHYVersion: ttnpb.PHY_V1_0_2_REV_B,
					LoRaWANVersion:    ttnpb.MAC_V1_0_2,
					SupportsJoin:      true,
					MACSettings: &ttnpb.MACSettings{
						Supports32BitFCnt: &pbtypes.BoolValue{Value: true},
						UseADR:            &pbtypes.BoolValue{Value: true},
					},
					RootKeys: &ttnpb.RootKeys{
						AppKey: &ttnpb.KeyEnvelope{
							Key: []byte{0x4f, 0x93, 0x95, 0x95, 0xce, 0x83, 0x28, 0x8a, 0xfe, 0xf8, 0x1b, 0xd8, 0x81, 0xc3, 0xc3, 0x6e},
						},
					},
					Session: &ttnpb.Session{
						StartedAt:     time.Unix(0, 42).UTC(),
						DevAddr:       types.DevAddr{0x01, 0x0b, 0x60, 0x0c},
						LastFCntUp:    45872,
						LastNFCntDown: 1880,
						SessionKeys: ttnpb.SessionKeys{
							AppSKey: &ttnpb.KeyEnvelope{
								Key: []byte{0xa1, 0x5b, 0xef, 0x4a, 0x32, 0x33, 0x27, 0x4a, 0xe9, 0x17, 0xe4, 0xaf, 0xb1, 0x90, 0x55, 0xf2},
							},
							FNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: []byte{0x9e, 0x2f, 0xb6, 0x1d, 0x73, 0x10, 0xc9, 0x27, 0x98, 0x86, 0xdb, 0x79, 0xfa, 0x52, 0xf9, 0xf4},
							},
							NwkSEncKey: &ttnpb.KeyEnvelope{
								Key: []byte{0x9e, 0x2f, 0xb6, 0x1d, 0x73, 0x10, 0xc9, 0x27, 0x98, 0x86, 0xdb, 0x79, 0xfa, 0x52, 0xf9, 0xf4},
							},
							SNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: []byte{0x9e, 0x2f, 0xb6, 0x1d, 0x73, 0x10, 0xc9, 0x27, 0x98, 0x86, 0xdb, 0x79, 0xfa, 0x52, 0xf9, 0xf4},
							},
						},
					},
				}
				err = ResetMACState(expected, frequencyplans.NewStore(test.FrequencyPlansFetcher), ttnpb.MACSettings{})
				if !a.So(err, should.BeNil) {
					t.Fatalf("Failed to reset MAC state: %s", err)
				}
				a.So(dev, should.Resemble, expected)
				return dev, nil
			},
			Request: &ttnpb.SetEndDeviceRequest{
				Device: ttnpb.EndDevice{
					EndDeviceIdentifiers: ids,
					FrequencyPlanID:      test.EUFrequencyPlanID,
					LoRaWANPHYVersion:    ttnpb.PHY_V1_0_2_REV_B,
					LoRaWANVersion:       ttnpb.MAC_V1_0_2,
					SupportsJoin:         true,
					MACSettings: &ttnpb.MACSettings{
						Supports32BitFCnt: &pbtypes.BoolValue{Value: true},
						UseADR:            &pbtypes.BoolValue{Value: true},
					},
					RootKeys: &ttnpb.RootKeys{
						AppKey: &ttnpb.KeyEnvelope{
							Key: []byte{0x4f, 0x93, 0x95, 0x95, 0xce, 0x83, 0x28, 0x8a, 0xfe, 0xf8, 0x1b, 0xd8, 0x81, 0xc3, 0xc3, 0x6e},
						},
					},
					Session: &ttnpb.Session{
						StartedAt:     time.Unix(0, 42).UTC(),
						DevAddr:       types.DevAddr{0x01, 0x0b, 0x60, 0x0c},
						LastFCntUp:    45872,
						LastNFCntDown: 1880,
						SessionKeys: ttnpb.SessionKeys{
							AppSKey: &ttnpb.KeyEnvelope{
								Key: []byte{0xa1, 0x5b, 0xef, 0x4a, 0x32, 0x33, 0x27, 0x4a, 0xe9, 0x17, 0xe4, 0xaf, 0xb1, 0x90, 0x55, 0xf2},
							},
							FNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: []byte{0x9e, 0x2f, 0xb6, 0x1d, 0x73, 0x10, 0xc9, 0x27, 0x98, 0x86, 0xdb, 0x79, 0xfa, 0x52, 0xf9, 0xf4},
							},
						},
					},
				},
				FieldMask: pbtypes.FieldMask{
					Paths: []string{
						"frequency_plan_id",
						"lorawan_phy_version",
						"lorawan_version",
						"mac_settings.use_adr",
						"mac_settings.supports_32_bit_f_cnt",
						"resets_f_cnt",
						"resets_join_nonces",
						"root_keys.app_key.key",
						"session.dev_addr",
						"session.keys.app_s_key.key",
						"session.keys.f_nwk_s_int_key.key",
						"session.last_f_cnt_up",
						"session.last_n_f_cnt_down",
						"session.started_at",
						"supports_join",
					},
				},
			},
			Device: func() *ttnpb.EndDevice {
				expected := &ttnpb.EndDevice{
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						DeviceID:               ids.DeviceID,
						ApplicationIdentifiers: ids.ApplicationIdentifiers,
						DevAddr:                &types.DevAddr{0x01, 0x0b, 0x60, 0x0c},
					},
					FrequencyPlanID:   test.EUFrequencyPlanID,
					LoRaWANPHYVersion: ttnpb.PHY_V1_0_2_REV_B,
					LoRaWANVersion:    ttnpb.MAC_V1_0_2,
					SupportsJoin:      true,
					MACSettings: &ttnpb.MACSettings{
						Supports32BitFCnt: &pbtypes.BoolValue{Value: true},
						UseADR:            &pbtypes.BoolValue{Value: true},
					},
					RootKeys: &ttnpb.RootKeys{
						AppKey: &ttnpb.KeyEnvelope{
							Key: []byte{0x4f, 0x93, 0x95, 0x95, 0xce, 0x83, 0x28, 0x8a, 0xfe, 0xf8, 0x1b, 0xd8, 0x81, 0xc3, 0xc3, 0x6e},
						},
					},
					Session: &ttnpb.Session{
						StartedAt:     time.Unix(0, 42).UTC(),
						DevAddr:       types.DevAddr{0x01, 0x0b, 0x60, 0x0c},
						LastFCntUp:    45872,
						LastNFCntDown: 1880,
						SessionKeys: ttnpb.SessionKeys{
							AppSKey: &ttnpb.KeyEnvelope{
								Key: []byte{0xa1, 0x5b, 0xef, 0x4a, 0x32, 0x33, 0x27, 0x4a, 0xe9, 0x17, 0xe4, 0xaf, 0xb1, 0x90, 0x55, 0xf2},
							},
							FNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: []byte{0x9e, 0x2f, 0xb6, 0x1d, 0x73, 0x10, 0xc9, 0x27, 0x98, 0x86, 0xdb, 0x79, 0xfa, 0x52, 0xf9, 0xf4},
							},
							NwkSEncKey: &ttnpb.KeyEnvelope{
								Key: []byte{0x9e, 0x2f, 0xb6, 0x1d, 0x73, 0x10, 0xc9, 0x27, 0x98, 0x86, 0xdb, 0x79, 0xfa, 0x52, 0xf9, 0xf4},
							},
							SNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: []byte{0x9e, 0x2f, 0xb6, 0x1d, 0x73, 0x10, 0xc9, 0x27, 0x98, 0x86, 0xdb, 0x79, 0xfa, 0x52, 0xf9, 0xf4},
							},
						},
					},
				}
				if err := ResetMACState(expected, frequencyplans.NewStore(test.FrequencyPlansFetcher), ttnpb.MACSettings{}); err != nil {
					t.Fatalf("Failed to reset MAC state: %s", err)
				}
				return expected
			}(),
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
			ns.FrequencyPlans = frequencyplans.NewStore(test.FrequencyPlansFetcher)

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
