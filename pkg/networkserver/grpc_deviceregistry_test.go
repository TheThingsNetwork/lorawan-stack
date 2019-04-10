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
	"fmt"
	"sync/atomic"
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
	for _, tc := range []struct {
		Name           string
		ContextFunc    func(context.Context) context.Context
		GetByIDFunc    func(context.Context, ttnpb.ApplicationIdentifiers, string, []string) (*ttnpb.EndDevice, error)
		Request        *ttnpb.GetEndDeviceRequest
		Device         *ttnpb.EndDevice
		ErrorAssertion func(*testing.T, error) bool
		GetByIDCalls   uint64
	}{
		{
			Name: "No device read rights",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"}): {
							Rights: []ttnpb.Right{
								ttnpb.RIGHT_GATEWAY_SETTINGS_BASIC,
							},
						},
					},
				})
			},
			GetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string) (*ttnpb.EndDevice, error) {
				err := errors.New("GetByIDFunc must not be called")
				test.MustTFromContext(ctx).Error(err)
				return nil, err
			},
			Request: &ttnpb.GetEndDeviceRequest{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DeviceID:               "test-dev-id",
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
				},
				FieldMask: pbtypes.FieldMask{
					Paths: []string{
						"frequency_plan_id",
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
		},

		{
			Name: "Valid request",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"}): {
							Rights: []ttnpb.Right{
								ttnpb.RIGHT_APPLICATION_DEVICES_READ,
							},
						},
					},
				})
			},
			GetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(appID, should.Resemble, ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"})
				a.So(devID, should.Equal, "test-dev-id")
				a.So(gets, should.HaveSameElementsDeep, []string{
					"frequency_plan_id",
				})
				return &ttnpb.EndDevice{
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						DeviceID:               "test-dev-id",
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
					},
				}, nil
			},
			Request: &ttnpb.GetEndDeviceRequest{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DeviceID:               "test-dev-id",
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
				},
				FieldMask: pbtypes.FieldMask{
					Paths: []string{
						"frequency_plan_id",
					},
				},
			},
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DeviceID:               "test-dev-id",
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
				},
			},
			GetByIDCalls: 1,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			var getByIDCalls uint64

			ns := test.Must(New(
				component.MustNew(test.GetLogger(t), &component.Config{}),
				&Config{
					Devices: &MockDeviceRegistry{
						GetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string) (*ttnpb.EndDevice, error) {
							atomic.AddUint64(&getByIDCalls, 1)
							return tc.GetByIDFunc(ctx, appID, devID, gets)
						},
					},
					DownlinkTasks:       &MockDownlinkTaskQueue{},
					DeduplicationWindow: 42,
					CooldownWindow:      42,
				})).(*NetworkServer)

			ns.AddContextFiller(tc.ContextFunc)
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
			if tc.ErrorAssertion != nil && a.So(tc.ErrorAssertion(t, err), should.BeTrue) {
				a.So(dev, should.BeNil)
			} else if a.So(err, should.BeNil) {
				a.So(dev, should.Resemble, tc.Device)
			}
			a.So(req, should.Resemble, tc.Request)
			a.So(getByIDCalls, should.Equal, tc.GetByIDCalls)
		})
	}
}

func TestDeviceRegistrySet(t *testing.T) {
	for _, tc := range []struct {
		Name           string
		ContextFunc    func(context.Context) context.Context
		SetByIDFunc    func(context.Context, ttnpb.ApplicationIdentifiers, string, []string, func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error)
		Request        *ttnpb.SetEndDeviceRequest
		Device         *ttnpb.EndDevice
		ErrorAssertion func(*testing.T, error) bool
		SetByIDCalls   uint64
	}{
		{
			Name: "No device write rights",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"}): {
							Rights: []ttnpb.Right{
								ttnpb.RIGHT_GATEWAY_SETTINGS_BASIC,
							},
						},
					},
				})
			},
			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				err := errors.New("SetByIDFunc must not be called")
				test.MustTFromContext(ctx).Error(err)
				return nil, err
			},
			Request: &ttnpb.SetEndDeviceRequest{
				EndDevice: ttnpb.EndDevice{
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						DeviceID:               "test-dev-id",
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
						JoinEUI:                &types.EUI64{0x42, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
						DevEUI:                 &types.EUI64{0x42, 0x42, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
					},
					FrequencyPlanID:   test.EUFrequencyPlanID,
					LoRaWANPHYVersion: ttnpb.PHY_V1_0,
					LoRaWANVersion:    ttnpb.MAC_V1_0,
					SupportsJoin:      true,
				},
				FieldMask: pbtypes.FieldMask{
					Paths: []string{
						"frequency_plan_id",
						"lorawan_phy_version",
						"lorawan_version",
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
		},

		{
			Name: "Create OTAA device",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"}): {
							Rights: []ttnpb.Right{
								ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
							},
						},
					},
				})
			},
			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(appID, should.Resemble, ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"})
				a.So(devID, should.Equal, "test-dev-id")
				a.So(gets, should.HaveSameElementsDeep, []string{
					"frequency_plan_id",
					"lorawan_phy_version",
					"lorawan_version",
					"mac_settings.adr_margin",
					"supports_class_b",
					"supports_class_c",
					"supports_join",
				})

				dev, sets, err := f(nil)
				if !a.So(err, should.BeNil) {
					return nil, err
				}
				a.So(sets, should.HaveSameElementsDeep, []string{
					"frequency_plan_id",
					"lorawan_phy_version",
					"lorawan_version",
					"mac_settings.adr_margin",
					"supports_class_b",
					"supports_class_c",
					"supports_join",
				})
				a.So(dev, should.Resemble, &ttnpb.EndDevice{
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						DeviceID:               "test-dev-id",
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
						JoinEUI:                &types.EUI64{0x42, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
						DevEUI:                 &types.EUI64{0x42, 0x42, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
					},
					FrequencyPlanID:   test.EUFrequencyPlanID,
					LoRaWANPHYVersion: ttnpb.PHY_V1_0,
					LoRaWANVersion:    ttnpb.MAC_V1_0,
					SupportsJoin:      true,
					MACSettings: &ttnpb.MACSettings{
						ADRMargin: &pbtypes.FloatValue{Value: 4},
					},
				})
				return dev, nil
			},
			Request: &ttnpb.SetEndDeviceRequest{
				EndDevice: ttnpb.EndDevice{
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						DeviceID:               "test-dev-id",
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
						JoinEUI:                &types.EUI64{0x42, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
						DevEUI:                 &types.EUI64{0x42, 0x42, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
					},
					FrequencyPlanID:   test.EUFrequencyPlanID,
					LoRaWANPHYVersion: ttnpb.PHY_V1_0,
					LoRaWANVersion:    ttnpb.MAC_V1_0,
					SupportsJoin:      true,
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
						"supports_class_b",
						"supports_class_c",
						"supports_join",
					},
				},
			},
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DeviceID:               "test-dev-id",
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
					JoinEUI:                &types.EUI64{0x42, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
					DevEUI:                 &types.EUI64{0x42, 0x42, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
				},
				FrequencyPlanID:   test.EUFrequencyPlanID,
				LoRaWANPHYVersion: ttnpb.PHY_V1_0,
				LoRaWANVersion:    ttnpb.MAC_V1_0,
				SupportsJoin:      true,
				MACSettings: &ttnpb.MACSettings{
					ADRMargin: &pbtypes.FloatValue{Value: 4},
				},
			},
			SetByIDCalls: 1,
		},

		{
			// https://github.com/TheThingsNetwork/lorawan-stack/issues/104#issuecomment-465074076
			Name: "Create OTAA device with existing session",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"}): {
							Rights: []ttnpb.Right{
								ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
							},
						},
					},
				})
			},
			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(appID, should.Resemble, ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"})
				a.So(devID, should.Equal, "test-dev-id")
				a.So(gets, should.HaveSameElementsDeep, []string{
					"frequency_plan_id",
					"lorawan_phy_version",
					"lorawan_version",
					"mac_settings.supports_32_bit_f_cnt",
					"mac_settings.use_adr",
					"session.dev_addr",
					"session.keys.f_nwk_s_int_key.key",
					"session.last_f_cnt_up",
					"session.last_n_f_cnt_down",
					"session.started_at",
					"supports_join",
				})

				dev, sets, err := f(nil)
				if !a.So(err, should.BeNil) {
					return nil, err
				}
				a.So(sets, should.HaveSameElementsDeep, []string{
					"frequency_plan_id",
					"lorawan_phy_version",
					"lorawan_version",
					"mac_settings.supports_32_bit_f_cnt",
					"mac_settings.use_adr",
					"mac_state",
					"session.dev_addr",
					"session.keys.f_nwk_s_int_key.key",
					"session.keys.nwk_s_enc_key.key",
					"session.keys.s_nwk_s_int_key.key",
					"session.last_f_cnt_up",
					"session.last_n_f_cnt_down",
					"session.started_at",
					"supports_join",
				})

				expected := &ttnpb.EndDevice{
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						DeviceID:               "test-dev-id",
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
						JoinEUI:                &types.EUI64{0x70, 0xB3, 0xD5, 0x95, 0x20, 0x00, 0x00, 0x00},
						DevEUI:                 &types.EUI64{0xA8, 0x17, 0x58, 0xFF, 0xFE, 0x03, 0x22, 0x77},
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
					Session: &ttnpb.Session{
						StartedAt:     time.Unix(0, 42).UTC(),
						DevAddr:       types.DevAddr{0x01, 0x0b, 0x60, 0x0c},
						LastFCntUp:    45872,
						LastNFCntDown: 1880,
						SessionKeys: ttnpb.SessionKeys{
							FNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: &types.AES128Key{0x9e, 0x2f, 0xb6, 0x1d, 0x73, 0x10, 0xc9, 0x27, 0x98, 0x86, 0xdb, 0x79, 0xfa, 0x52, 0xf9, 0xf4},
							},
							NwkSEncKey: &ttnpb.KeyEnvelope{
								Key: &types.AES128Key{0x9e, 0x2f, 0xb6, 0x1d, 0x73, 0x10, 0xc9, 0x27, 0x98, 0x86, 0xdb, 0x79, 0xfa, 0x52, 0xf9, 0xf4},
							},
							SNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: &types.AES128Key{0x9e, 0x2f, 0xb6, 0x1d, 0x73, 0x10, 0xc9, 0x27, 0x98, 0x86, 0xdb, 0x79, 0xfa, 0x52, 0xf9, 0xf4},
							},
						},
					},
				}
				err = ResetMACState(expected, frequencyplans.NewStore(test.FrequencyPlansFetcher), ttnpb.MACSettings{})
				if !a.So(err, should.BeNil) {
					panic(fmt.Sprintf("Failed to reset MAC state: %s", err))
				}
				a.So(dev, should.Resemble, expected)
				return dev, nil
			},
			Request: &ttnpb.SetEndDeviceRequest{
				EndDevice: ttnpb.EndDevice{
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						DeviceID:               "test-dev-id",
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
						JoinEUI:                &types.EUI64{0x70, 0xB3, 0xD5, 0x95, 0x20, 0x00, 0x00, 0x00},
						DevEUI:                 &types.EUI64{0xA8, 0x17, 0x58, 0xFF, 0xFE, 0x03, 0x22, 0x77},
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
					Session: &ttnpb.Session{
						StartedAt:     time.Unix(0, 42).UTC(),
						DevAddr:       types.DevAddr{0x01, 0x0b, 0x60, 0x0c},
						LastFCntUp:    45872,
						LastNFCntDown: 1880,
						SessionKeys: ttnpb.SessionKeys{
							FNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: &types.AES128Key{0x9e, 0x2f, 0xb6, 0x1d, 0x73, 0x10, 0xc9, 0x27, 0x98, 0x86, 0xdb, 0x79, 0xfa, 0x52, 0xf9, 0xf4},
							},
						},
					},
				},
				FieldMask: pbtypes.FieldMask{
					Paths: []string{
						"frequency_plan_id",
						"lorawan_phy_version",
						"lorawan_version",
						"mac_settings.supports_32_bit_f_cnt",
						"mac_settings.use_adr",
						"session.dev_addr",
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
						DeviceID:               "test-dev-id",
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
						JoinEUI:                &types.EUI64{0x70, 0xB3, 0xD5, 0x95, 0x20, 0x00, 0x00, 0x00},
						DevEUI:                 &types.EUI64{0xA8, 0x17, 0x58, 0xFF, 0xFE, 0x03, 0x22, 0x77},
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
					Session: &ttnpb.Session{
						StartedAt:     time.Unix(0, 42).UTC(),
						DevAddr:       types.DevAddr{0x01, 0x0b, 0x60, 0x0c},
						LastFCntUp:    45872,
						LastNFCntDown: 1880,
						SessionKeys: ttnpb.SessionKeys{
							FNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: &types.AES128Key{0x9e, 0x2f, 0xb6, 0x1d, 0x73, 0x10, 0xc9, 0x27, 0x98, 0x86, 0xdb, 0x79, 0xfa, 0x52, 0xf9, 0xf4},
							},
							NwkSEncKey: &ttnpb.KeyEnvelope{
								Key: &types.AES128Key{0x9e, 0x2f, 0xb6, 0x1d, 0x73, 0x10, 0xc9, 0x27, 0x98, 0x86, 0xdb, 0x79, 0xfa, 0x52, 0xf9, 0xf4},
							},
							SNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: &types.AES128Key{0x9e, 0x2f, 0xb6, 0x1d, 0x73, 0x10, 0xc9, 0x27, 0x98, 0x86, 0xdb, 0x79, 0xfa, 0x52, 0xf9, 0xf4},
							},
						},
					},
				}
				if err := ResetMACState(expected, frequencyplans.NewStore(test.FrequencyPlansFetcher), ttnpb.MACSettings{}); err != nil {
					panic(fmt.Sprintf("Failed to reset MAC state: %s", err))
				}
				return expected
			}(),
			SetByIDCalls: 1,
		},

		{
			Name: "Create ABP device",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"}): {
							Rights: []ttnpb.Right{
								ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
							},
						},
					},
				})
			},
			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(appID, should.Resemble, ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"})
				a.So(devID, should.Equal, "test-dev-id")
				a.So(gets, should.HaveSameElementsDeep, []string{
					"frequency_plan_id",
					"lorawan_phy_version",
					"lorawan_version",
					"mac_settings.supports_32_bit_f_cnt",
					"mac_settings.use_adr",
					"session.dev_addr",
					"session.keys.f_nwk_s_int_key.key",
					"session.last_f_cnt_up",
					"session.last_n_f_cnt_down",
					"session.started_at",
					"supports_join",
				})

				dev, sets, err := f(nil)
				if !a.So(err, should.BeNil) {
					return nil, err
				}
				a.So(sets, should.HaveSameElementsDeep, []string{
					"frequency_plan_id",
					"lorawan_phy_version",
					"lorawan_version",
					"mac_settings.supports_32_bit_f_cnt",
					"mac_settings.use_adr",
					"mac_state",
					"session.dev_addr",
					"session.keys.f_nwk_s_int_key.key",
					"session.keys.nwk_s_enc_key.key",
					"session.keys.s_nwk_s_int_key.key",
					"session.last_f_cnt_up",
					"session.last_n_f_cnt_down",
					"session.started_at",
					"supports_join",
				})

				expected := &ttnpb.EndDevice{
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						DeviceID:               "test-dev-id",
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
						DevAddr:                &types.DevAddr{0x42, 0x00, 0x00, 0x00},
					},
					FrequencyPlanID:   test.EUFrequencyPlanID,
					LoRaWANPHYVersion: ttnpb.PHY_V1_0_2_REV_B,
					LoRaWANVersion:    ttnpb.MAC_V1_0_2,
					MACSettings: &ttnpb.MACSettings{
						Supports32BitFCnt: &pbtypes.BoolValue{Value: true},
						UseADR:            &pbtypes.BoolValue{Value: true},
					},
					Session: &ttnpb.Session{
						StartedAt:     time.Unix(0, 42).UTC(),
						DevAddr:       types.DevAddr{0x42, 0x00, 0x00, 0x00},
						LastFCntUp:    42,
						LastNFCntDown: 4242,
						SessionKeys: ttnpb.SessionKeys{
							FNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: &types.AES128Key{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
							},
							NwkSEncKey: &ttnpb.KeyEnvelope{
								Key: &types.AES128Key{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
							},
							SNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: &types.AES128Key{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
							},
						},
					},
				}
				err = ResetMACState(expected, frequencyplans.NewStore(test.FrequencyPlansFetcher), ttnpb.MACSettings{})
				if !a.So(err, should.BeNil) {
					panic(fmt.Sprintf("Failed to reset MAC state: %s", err))
				}
				a.So(dev, should.Resemble, expected)
				return dev, nil
			},
			Request: &ttnpb.SetEndDeviceRequest{
				EndDevice: ttnpb.EndDevice{
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						DeviceID:               "test-dev-id",
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
						DevAddr:                &types.DevAddr{0x42, 0x00, 0x00, 0x00},
					},
					FrequencyPlanID:   test.EUFrequencyPlanID,
					LoRaWANPHYVersion: ttnpb.PHY_V1_0_2_REV_B,
					LoRaWANVersion:    ttnpb.MAC_V1_0_2,
					MACSettings: &ttnpb.MACSettings{
						Supports32BitFCnt: &pbtypes.BoolValue{Value: true},
						UseADR:            &pbtypes.BoolValue{Value: true},
					},
					Session: &ttnpb.Session{
						StartedAt:     time.Unix(0, 42).UTC(),
						DevAddr:       types.DevAddr{0x42, 0x00, 0x00, 0x00},
						LastFCntUp:    42,
						LastNFCntDown: 4242,
						SessionKeys: ttnpb.SessionKeys{
							FNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: &types.AES128Key{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
							},
						},
					},
				},
				FieldMask: pbtypes.FieldMask{
					Paths: []string{
						"frequency_plan_id",
						"lorawan_phy_version",
						"lorawan_version",
						"mac_settings.supports_32_bit_f_cnt",
						"mac_settings.use_adr",
						"session.dev_addr",
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
						DeviceID:               "test-dev-id",
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
						DevAddr:                &types.DevAddr{0x42, 0x00, 0x00, 0x00},
					},
					FrequencyPlanID:   test.EUFrequencyPlanID,
					LoRaWANPHYVersion: ttnpb.PHY_V1_0_2_REV_B,
					LoRaWANVersion:    ttnpb.MAC_V1_0_2,
					MACSettings: &ttnpb.MACSettings{
						Supports32BitFCnt: &pbtypes.BoolValue{Value: true},
						UseADR:            &pbtypes.BoolValue{Value: true},
					},
					Session: &ttnpb.Session{
						StartedAt:     time.Unix(0, 42).UTC(),
						DevAddr:       types.DevAddr{0x42, 0x00, 0x00, 0x00},
						LastFCntUp:    42,
						LastNFCntDown: 4242,
						SessionKeys: ttnpb.SessionKeys{
							FNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: &types.AES128Key{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
							},
							NwkSEncKey: &ttnpb.KeyEnvelope{
								Key: &types.AES128Key{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
							},
							SNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: &types.AES128Key{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
							},
						},
					},
				}
				if err := ResetMACState(expected, frequencyplans.NewStore(test.FrequencyPlansFetcher), ttnpb.MACSettings{}); err != nil {
					panic(fmt.Sprintf("Failed to reset MAC state: %s", err))
				}
				return expected
			}(),
			SetByIDCalls: 1,
		},

		{
			// https://github.com/TheThingsNetwork/lorawan-stack/issues/159#issue-411803325
			Name: "Create ABP device with existing session",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"}): {
							Rights: []ttnpb.Right{
								ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
							},
						},
					},
				})
			},
			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(appID, should.Resemble, ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"})
				a.So(devID, should.Equal, "test-dev-id")
				a.So(gets, should.HaveSameElementsDeep, []string{
					"frequency_plan_id",
					"lorawan_phy_version",
					"lorawan_version",
					"mac_settings.supports_32_bit_f_cnt",
					"mac_settings.use_adr",
					"session.dev_addr",
					"session.keys.f_nwk_s_int_key.key",
					"session.last_f_cnt_up",
					"session.last_n_f_cnt_down",
					"session.started_at",
					"supports_join",
				})

				dev, sets, err := f(nil)
				if !a.So(err, should.BeNil) {
					return nil, err
				}
				a.So(sets, should.HaveSameElementsDeep, []string{
					"frequency_plan_id",
					"lorawan_phy_version",
					"lorawan_version",
					"mac_settings.supports_32_bit_f_cnt",
					"mac_settings.use_adr",
					"mac_state",
					"session.dev_addr",
					"session.keys.f_nwk_s_int_key.key",
					"session.keys.nwk_s_enc_key.key",
					"session.keys.s_nwk_s_int_key.key",
					"session.last_f_cnt_up",
					"session.last_n_f_cnt_down",
					"session.started_at",
					"supports_join",
				})

				expected := &ttnpb.EndDevice{
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						DeviceID:               "test-dev-id",
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
						JoinEUI:                &types.EUI64{0x70, 0xB3, 0xD5, 0x95, 0x20, 0x00, 0x00, 0x00},
						DevEUI:                 &types.EUI64{0xA8, 0x17, 0x58, 0xFF, 0xFE, 0x03, 0x22, 0x77},
						DevAddr:                &types.DevAddr{0x01, 0x0b, 0x60, 0x0c},
					},
					FrequencyPlanID:   test.EUFrequencyPlanID,
					LoRaWANPHYVersion: ttnpb.PHY_V1_0_2_REV_B,
					LoRaWANVersion:    ttnpb.MAC_V1_0_2,
					MACSettings: &ttnpb.MACSettings{
						Supports32BitFCnt: &pbtypes.BoolValue{Value: true},
						UseADR:            &pbtypes.BoolValue{Value: true},
					},
					Session: &ttnpb.Session{
						StartedAt:     time.Unix(0, 42).UTC(),
						DevAddr:       types.DevAddr{0x01, 0x0b, 0x60, 0x0c},
						LastFCntUp:    45872,
						LastNFCntDown: 1880,
						SessionKeys: ttnpb.SessionKeys{
							FNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: &types.AES128Key{0x9e, 0x2f, 0xb6, 0x1d, 0x73, 0x10, 0xc9, 0x27, 0x98, 0x86, 0xdb, 0x79, 0xfa, 0x52, 0xf9, 0xf4},
							},
							NwkSEncKey: &ttnpb.KeyEnvelope{
								Key: &types.AES128Key{0x9e, 0x2f, 0xb6, 0x1d, 0x73, 0x10, 0xc9, 0x27, 0x98, 0x86, 0xdb, 0x79, 0xfa, 0x52, 0xf9, 0xf4},
							},
							SNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: &types.AES128Key{0x9e, 0x2f, 0xb6, 0x1d, 0x73, 0x10, 0xc9, 0x27, 0x98, 0x86, 0xdb, 0x79, 0xfa, 0x52, 0xf9, 0xf4},
							},
						},
					},
				}
				err = ResetMACState(expected, frequencyplans.NewStore(test.FrequencyPlansFetcher), ttnpb.MACSettings{})
				if !a.So(err, should.BeNil) {
					panic(fmt.Sprintf("Failed to reset MAC state: %s", err))
				}
				a.So(dev, should.Resemble, expected)
				return dev, nil
			},
			Request: &ttnpb.SetEndDeviceRequest{
				EndDevice: ttnpb.EndDevice{
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						DeviceID:               "test-dev-id",
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
						JoinEUI:                &types.EUI64{0x70, 0xB3, 0xD5, 0x95, 0x20, 0x00, 0x00, 0x00},
						DevEUI:                 &types.EUI64{0xA8, 0x17, 0x58, 0xFF, 0xFE, 0x03, 0x22, 0x77},
						DevAddr:                &types.DevAddr{0x01, 0x0b, 0x60, 0x0c},
					},
					FrequencyPlanID:   test.EUFrequencyPlanID,
					LoRaWANPHYVersion: ttnpb.PHY_V1_0_2_REV_B,
					LoRaWANVersion:    ttnpb.MAC_V1_0_2,
					MACSettings: &ttnpb.MACSettings{
						Supports32BitFCnt: &pbtypes.BoolValue{Value: true},
						UseADR:            &pbtypes.BoolValue{Value: true},
					},
					Session: &ttnpb.Session{
						StartedAt:     time.Unix(0, 42).UTC(),
						DevAddr:       types.DevAddr{0x01, 0x0b, 0x60, 0x0c},
						LastFCntUp:    45872,
						LastNFCntDown: 1880,
						SessionKeys: ttnpb.SessionKeys{
							FNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: &types.AES128Key{0x9e, 0x2f, 0xb6, 0x1d, 0x73, 0x10, 0xc9, 0x27, 0x98, 0x86, 0xdb, 0x79, 0xfa, 0x52, 0xf9, 0xf4},
							},
						},
					},
				},
				FieldMask: pbtypes.FieldMask{
					Paths: []string{
						"frequency_plan_id",
						"lorawan_phy_version",
						"lorawan_version",
						"mac_settings.supports_32_bit_f_cnt",
						"mac_settings.use_adr",
						"session.dev_addr",
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
						DeviceID:               "test-dev-id",
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
						JoinEUI:                &types.EUI64{0x70, 0xB3, 0xD5, 0x95, 0x20, 0x00, 0x00, 0x00},
						DevEUI:                 &types.EUI64{0xA8, 0x17, 0x58, 0xFF, 0xFE, 0x03, 0x22, 0x77},
						DevAddr:                &types.DevAddr{0x01, 0x0b, 0x60, 0x0c},
					},
					FrequencyPlanID:   test.EUFrequencyPlanID,
					LoRaWANPHYVersion: ttnpb.PHY_V1_0_2_REV_B,
					LoRaWANVersion:    ttnpb.MAC_V1_0_2,
					MACSettings: &ttnpb.MACSettings{
						Supports32BitFCnt: &pbtypes.BoolValue{Value: true},
						UseADR:            &pbtypes.BoolValue{Value: true},
					},
					Session: &ttnpb.Session{
						StartedAt:     time.Unix(0, 42).UTC(),
						DevAddr:       types.DevAddr{0x01, 0x0b, 0x60, 0x0c},
						LastFCntUp:    45872,
						LastNFCntDown: 1880,
						SessionKeys: ttnpb.SessionKeys{
							FNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: &types.AES128Key{0x9e, 0x2f, 0xb6, 0x1d, 0x73, 0x10, 0xc9, 0x27, 0x98, 0x86, 0xdb, 0x79, 0xfa, 0x52, 0xf9, 0xf4},
							},
							NwkSEncKey: &ttnpb.KeyEnvelope{
								Key: &types.AES128Key{0x9e, 0x2f, 0xb6, 0x1d, 0x73, 0x10, 0xc9, 0x27, 0x98, 0x86, 0xdb, 0x79, 0xfa, 0x52, 0xf9, 0xf4},
							},
							SNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: &types.AES128Key{0x9e, 0x2f, 0xb6, 0x1d, 0x73, 0x10, 0xc9, 0x27, 0x98, 0x86, 0xdb, 0x79, 0xfa, 0x52, 0xf9, 0xf4},
							},
						},
					},
				}
				if err := ResetMACState(expected, frequencyplans.NewStore(test.FrequencyPlansFetcher), ttnpb.MACSettings{}); err != nil {
					panic(fmt.Sprintf("Failed to reset MAC state: %s", err))
				}
				return expected
			}(),
			SetByIDCalls: 1,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			var setByIDCalls uint64

			ns := test.Must(New(
				component.MustNew(test.GetLogger(t), &component.Config{}),
				&Config{
					Devices: &MockDeviceRegistry{
						SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
							atomic.AddUint64(&setByIDCalls, 1)
							return tc.SetByIDFunc(ctx, appID, devID, gets, f)
						},
					},
					DownlinkTasks:       &MockDownlinkTaskQueue{},
					DeduplicationWindow: 42,
					CooldownWindow:      42,
				})).(*NetworkServer)
			ns.FrequencyPlans = frequencyplans.NewStore(test.FrequencyPlansFetcher)

			ns.AddContextFiller(tc.ContextFunc)
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

			ctx := ns.FillContext(test.Context())
			req := deepcopy.Copy(tc.Request).(*ttnpb.SetEndDeviceRequest)

			dev, err := ttnpb.NewNsEndDeviceRegistryClient(ns.LoopbackConn()).Set(ctx, req)
			a.So(setByIDCalls, should.Equal, tc.SetByIDCalls)
			if tc.ErrorAssertion != nil && a.So(tc.ErrorAssertion(t, err), should.BeTrue) {
				a.So(dev, should.BeNil)
			} else if a.So(err, should.BeNil) {
				a.So(dev, should.Resemble, tc.Device)
			}
			a.So(req, should.Resemble, tc.Request)
		})
	}
}

func TestDeviceRegistryDelete(t *testing.T) {
	for _, tc := range []struct {
		Name           string
		ContextFunc    func(context.Context) context.Context
		SetByIDFunc    func(context.Context, ttnpb.ApplicationIdentifiers, string, []string, func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error)
		Request        *ttnpb.EndDeviceIdentifiers
		ErrorAssertion func(*testing.T, error) bool
		SetByIDCalls   uint64
	}{
		{
			Name: "No device write rights",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"}): {
							Rights: []ttnpb.Right{
								ttnpb.RIGHT_GATEWAY_SETTINGS_BASIC,
							},
						},
					},
				})
			},
			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				err := errors.New("SetByIDFunc must not be called")
				test.MustTFromContext(ctx).Error(err)
				return nil, err
			},
			Request: &ttnpb.EndDeviceIdentifiers{
				DeviceID:               "test-dev-id",
				ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				if !assertions.New(t).So(errors.IsPermissionDenied(err), should.BeTrue) {
					t.Errorf("Received error: %s", err)
					return false
				}
				return true
			},
		},

		{
			Name: "Non-existing device",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"}): {
							Rights: []ttnpb.Right{
								ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
							},
						},
					},
				})
			},
			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				t := test.MustTFromContext(ctx)
				a := assertions.New(t)
				a.So(appID, should.Resemble, ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"})
				a.So(devID, should.Equal, "test-dev-id")
				a.So(gets, should.BeNil)

				dev, sets, err := f(nil)
				if !a.So(err, should.BeNil) {
					return nil, err
				}
				a.So(sets, should.BeNil)
				a.So(dev, should.BeNil)
				return nil, nil
			},
			Request: &ttnpb.EndDeviceIdentifiers{
				DeviceID:               "test-dev-id",
				ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
			},
			SetByIDCalls: 1,
		},

		{
			Name: "Existing device",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(ctx, ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"}): {
							Rights: []ttnpb.Right{
								ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
							},
						},
					},
				})
			},
			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(appID, should.Resemble, ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"})
				a.So(devID, should.Equal, "test-dev-id")
				a.So(gets, should.BeNil)

				dev, sets, err := f(&ttnpb.EndDevice{
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						DeviceID:               "test-dev-id",
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
					},
				})
				if !a.So(err, should.BeNil) {
					return nil, err
				}
				a.So(sets, should.BeNil)
				a.So(dev, should.BeNil)
				return nil, nil
			},
			Request: &ttnpb.EndDeviceIdentifiers{
				DeviceID:               "test-dev-id",
				ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
			},
			SetByIDCalls: 1,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			var setByIDCalls uint64

			ns := test.Must(New(
				component.MustNew(test.GetLogger(t), &component.Config{}),
				&Config{
					Devices: &MockDeviceRegistry{
						SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
							atomic.AddUint64(&setByIDCalls, 1)
							return tc.SetByIDFunc(ctx, appID, devID, gets, f)
						},
					},
					DownlinkTasks:       &MockDownlinkTaskQueue{},
					DeduplicationWindow: 42,
					CooldownWindow:      42,
				})).(*NetworkServer)

			ns.AddContextFiller(tc.ContextFunc)
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
			a.So(setByIDCalls, should.Equal, tc.SetByIDCalls)
			if tc.ErrorAssertion != nil && a.So(tc.ErrorAssertion(t, err), should.BeTrue) {
				a.So(res, should.BeNil)
			} else if a.So(err, should.BeNil) {
				a.So(res, should.Resemble, ttnpb.Empty)
			}
			a.So(req, should.Resemble, tc.Request)
		})
	}
}
