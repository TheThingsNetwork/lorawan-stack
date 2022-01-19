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
	"strings"
	"sync/atomic"
	"testing"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/mohae/deepcopy"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal/test"
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver/mac"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestDeviceRegistryGet(t *testing.T) {
	for _, tc := range []struct {
		Name           string
		ContextFunc    func(context.Context) context.Context
		GetByIDFunc    func(context.Context, *ttnpb.ApplicationIdentifiers, string, []string) (*ttnpb.EndDevice, context.Context, error)
		KeyVault       map[string][]byte
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
						unique.ID(test.Context(), ttnpb.ApplicationIdentifiers{ApplicationId: "test-app-id"}): {
							Rights: []ttnpb.Right{
								ttnpb.Right_RIGHT_GATEWAY_SETTINGS_BASIC,
							},
						},
					},
				})
			},
			GetByIDFunc: func(ctx context.Context, appID *ttnpb.ApplicationIdentifiers, devID string, gets []string) (*ttnpb.EndDevice, context.Context, error) {
				err := errors.New("GetByIDFunc must not be called")
				test.MustTFromContext(ctx).Error(err)
				return nil, ctx, err
			},
			Request: &ttnpb.GetEndDeviceRequest{
				EndDeviceIds: &ttnpb.EndDeviceIdentifiers{
					DeviceId:       "test-dev-id",
					ApplicationIds: &ttnpb.ApplicationIdentifiers{ApplicationId: "test-app-id"},
				},
				FieldMask: &pbtypes.FieldMask{
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
			Name: "no keys",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(test.Context(), ttnpb.ApplicationIdentifiers{ApplicationId: "test-app-id"}): {
							Rights: []ttnpb.Right{
								ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ,
							},
						},
					},
				})
			},
			GetByIDFunc: func(ctx context.Context, appID *ttnpb.ApplicationIdentifiers, devID string, gets []string) (*ttnpb.EndDevice, context.Context, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(appID, should.Resemble, ttnpb.ApplicationIdentifiers{ApplicationId: "test-app-id"})
				a.So(devID, should.Equal, "test-dev-id")
				a.So(gets, should.HaveSameElementsDeep, []string{
					"frequency_plan_id",
				})
				return &ttnpb.EndDevice{
					Ids: &ttnpb.EndDeviceIdentifiers{
						DeviceId:       "test-dev-id",
						ApplicationIds: &ttnpb.ApplicationIdentifiers{ApplicationId: "test-app-id"},
					},
					FrequencyPlanId: test.EUFrequencyPlanID,
				}, ctx, nil
			},
			Request: &ttnpb.GetEndDeviceRequest{
				EndDeviceIds: &ttnpb.EndDeviceIdentifiers{
					DeviceId:       "test-dev-id",
					ApplicationIds: &ttnpb.ApplicationIdentifiers{ApplicationId: "test-app-id"},
				},
				FieldMask: &pbtypes.FieldMask{
					Paths: []string{
						"frequency_plan_id",
					},
				},
			},
			Device: &ttnpb.EndDevice{
				Ids: &ttnpb.EndDeviceIdentifiers{
					DeviceId:       "test-dev-id",
					ApplicationIds: &ttnpb.ApplicationIdentifiers{ApplicationId: "test-app-id"},
				},
				FrequencyPlanId: test.EUFrequencyPlanID,
			},
			GetByIDCalls: 1,
		},

		{
			Name: "with keys",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(test.Context(), ttnpb.ApplicationIdentifiers{ApplicationId: "test-app-id"}): {
							Rights: []ttnpb.Right{
								ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ,
								ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ_KEYS,
								ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ,
							},
						},
					},
				})
			},
			GetByIDFunc: func(ctx context.Context, appID *ttnpb.ApplicationIdentifiers, devID string, gets []string) (*ttnpb.EndDevice, context.Context, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(appID, should.Resemble, ttnpb.ApplicationIdentifiers{ApplicationId: "test-app-id"})
				a.So(devID, should.Equal, "test-dev-id")
				a.So(gets, should.HaveSameElementsDeep, []string{
					"frequency_plan_id",
					"session",
					"queued_application_downlinks",
				})
				return &ttnpb.EndDevice{
					Ids: &ttnpb.EndDeviceIdentifiers{
						DeviceId:       "test-dev-id",
						ApplicationIds: &ttnpb.ApplicationIdentifiers{ApplicationId: "test-app-id"},
					},
					FrequencyPlanId: test.EUFrequencyPlanID,
					Session: &ttnpb.Session{
						Keys: &ttnpb.SessionKeys{
							FNwkSIntKey: &ttnpb.KeyEnvelope{
								KekLabel:     "test",
								EncryptedKey: []byte{0x96, 0x77, 0x8b, 0x25, 0xae, 0x6c, 0xa4, 0x35, 0xf9, 0x2b, 0x5b, 0x97, 0xc0, 0x50, 0xae, 0xd2, 0x46, 0x8a, 0xb8, 0xa1, 0x7a, 0xd8, 0x4e, 0x5d},
							},
						},
					},
				}, ctx, nil
			},
			KeyVault: map[string][]byte{
				"test": {0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17},
			},
			Request: &ttnpb.GetEndDeviceRequest{
				EndDeviceIds: &ttnpb.EndDeviceIdentifiers{
					DeviceId:       "test-dev-id",
					ApplicationIds: &ttnpb.ApplicationIdentifiers{ApplicationId: "test-app-id"},
				},
				FieldMask: &pbtypes.FieldMask{
					Paths: []string{
						"frequency_plan_id",
						"session",
					},
				},
			},
			Device: &ttnpb.EndDevice{
				Ids: &ttnpb.EndDeviceIdentifiers{
					DeviceId:       "test-dev-id",
					ApplicationIds: &ttnpb.ApplicationIdentifiers{ApplicationId: "test-app-id"},
				},
				FrequencyPlanId: test.EUFrequencyPlanID,
				Session: &ttnpb.Session{
					Keys: &ttnpb.SessionKeys{
						FNwkSIntKey: &ttnpb.KeyEnvelope{
							Key: AES128KeyPtr(types.AES128Key{0x0, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}),
						},
					},
				},
			},
			GetByIDCalls: 1,
		},

		{
			Name: "with specific key envelope",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(test.Context(), ttnpb.ApplicationIdentifiers{ApplicationId: "test-app-id"}): {
							Rights: []ttnpb.Right{
								ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ,
								ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ_KEYS,
							},
						},
					},
				})
			},
			GetByIDFunc: func(ctx context.Context, appID *ttnpb.ApplicationIdentifiers, devID string, gets []string) (*ttnpb.EndDevice, context.Context, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(appID, should.Resemble, ttnpb.ApplicationIdentifiers{ApplicationId: "test-app-id"})
				a.So(devID, should.Equal, "test-dev-id")
				a.So(gets, should.HaveSameElementsDeep, []string{
					"pending_session.keys.f_nwk_s_int_key",
				})
				return &ttnpb.EndDevice{
					Ids: &ttnpb.EndDeviceIdentifiers{
						DeviceId:       "test-dev-id",
						ApplicationIds: &ttnpb.ApplicationIdentifiers{ApplicationId: "test-app-id"},
					},
					PendingSession: &ttnpb.Session{
						Keys: &ttnpb.SessionKeys{
							FNwkSIntKey: &ttnpb.KeyEnvelope{
								KekLabel:     "test",
								EncryptedKey: []byte{0x96, 0x77, 0x8b, 0x25, 0xae, 0x6c, 0xa4, 0x35, 0xf9, 0x2b, 0x5b, 0x97, 0xc0, 0x50, 0xae, 0xd2, 0x46, 0x8a, 0xb8, 0xa1, 0x7a, 0xd8, 0x4e, 0x5d},
							},
						},
					},
				}, ctx, nil
			},
			KeyVault: map[string][]byte{
				"test": {0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17},
			},
			Request: &ttnpb.GetEndDeviceRequest{
				EndDeviceIds: &ttnpb.EndDeviceIdentifiers{
					DeviceId:       "test-dev-id",
					ApplicationIds: &ttnpb.ApplicationIdentifiers{ApplicationId: "test-app-id"},
				},
				FieldMask: &pbtypes.FieldMask{
					Paths: []string{
						"pending_session.keys.f_nwk_s_int_key",
					},
				},
			},
			Device: &ttnpb.EndDevice{
				Ids: &ttnpb.EndDeviceIdentifiers{
					DeviceId:       "test-dev-id",
					ApplicationIds: &ttnpb.ApplicationIdentifiers{ApplicationId: "test-app-id"},
				},
				PendingSession: &ttnpb.Session{
					Keys: &ttnpb.SessionKeys{
						FNwkSIntKey: &ttnpb.KeyEnvelope{
							Key: AES128KeyPtr(types.AES128Key{0x0, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}),
						},
					},
				},
			},
			GetByIDCalls: 1,
		},

		{
			Name: "with specific key",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(test.Context(), ttnpb.ApplicationIdentifiers{ApplicationId: "test-app-id"}): {
							Rights: []ttnpb.Right{
								ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ,
								ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ_KEYS,
							},
						},
					},
				})
			},
			GetByIDFunc: func(ctx context.Context, appID *ttnpb.ApplicationIdentifiers, devID string, gets []string) (*ttnpb.EndDevice, context.Context, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(appID, should.Resemble, ttnpb.ApplicationIdentifiers{ApplicationId: "test-app-id"})
				a.So(devID, should.Equal, "test-dev-id")
				a.So(gets, should.HaveSameElementsDeep, []string{
					"pending_session.keys.f_nwk_s_int_key.encrypted_key",
					"pending_session.keys.f_nwk_s_int_key.kek_label",
					"pending_session.keys.f_nwk_s_int_key.key",
				})
				return &ttnpb.EndDevice{
					Ids: &ttnpb.EndDeviceIdentifiers{
						DeviceId:       "test-dev-id",
						ApplicationIds: &ttnpb.ApplicationIdentifiers{ApplicationId: "test-app-id"},
					},
					PendingSession: &ttnpb.Session{
						Keys: &ttnpb.SessionKeys{
							FNwkSIntKey: &ttnpb.KeyEnvelope{
								KekLabel:     "test",
								EncryptedKey: []byte{0x96, 0x77, 0x8b, 0x25, 0xae, 0x6c, 0xa4, 0x35, 0xf9, 0x2b, 0x5b, 0x97, 0xc0, 0x50, 0xae, 0xd2, 0x46, 0x8a, 0xb8, 0xa1, 0x7a, 0xd8, 0x4e, 0x5d},
							},
						},
					},
				}, ctx, nil
			},
			KeyVault: map[string][]byte{
				"test": {0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17},
			},
			Request: &ttnpb.GetEndDeviceRequest{
				EndDeviceIds: &ttnpb.EndDeviceIdentifiers{
					DeviceId:       "test-dev-id",
					ApplicationIds: &ttnpb.ApplicationIdentifiers{ApplicationId: "test-app-id"},
				},
				FieldMask: &pbtypes.FieldMask{
					Paths: []string{
						"pending_session.keys.f_nwk_s_int_key.key",
					},
				},
			},
			Device: &ttnpb.EndDevice{
				Ids: &ttnpb.EndDeviceIdentifiers{
					DeviceId:       "test-dev-id",
					ApplicationIds: &ttnpb.ApplicationIdentifiers{ApplicationId: "test-app-id"},
				},
				PendingSession: &ttnpb.Session{
					Keys: &ttnpb.SessionKeys{
						FNwkSIntKey: &ttnpb.KeyEnvelope{
							Key: AES128KeyPtr(types.AES128Key{0x0, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}),
						},
					},
				},
			},
			GetByIDCalls: 1,
		},
	} {
		tc := tc
		test.RunSubtest(t, test.SubtestConfig{
			Name:     tc.Name,
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				var getByIDCalls uint64

				ns, ctx, _, stop := StartTest(
					ctx,
					TestConfig{
						Component: component.Config{
							ServiceBase: config.ServiceBase{
								KeyVault: config.KeyVault{
									Provider: "static",
									Static:   tc.KeyVault,
								},
							},
						},
						NetworkServer: Config{
							Devices: &MockDeviceRegistry{
								GetByIDFunc: func(ctx context.Context, appID *ttnpb.ApplicationIdentifiers, devID string, gets []string) (*ttnpb.EndDevice, context.Context, error) {
									atomic.AddUint64(&getByIDCalls, 1)
									return tc.GetByIDFunc(ctx, appID, devID, gets)
								},
							},
						},
						TaskStarter: StartTaskExclude(
							DownlinkProcessTaskName,
						),
					},
				)
				defer stop()

				ns.AddContextFiller(tc.ContextFunc)
				ns.AddContextFiller(func(ctx context.Context) context.Context {
					return test.ContextWithTB(ctx, t)
				})

				req := deepcopy.Copy(tc.Request).(*ttnpb.GetEndDeviceRequest)
				dev, err := ttnpb.NewNsEndDeviceRegistryClient(ns.LoopbackConn()).Get(ctx, req)
				if tc.ErrorAssertion != nil && a.So(tc.ErrorAssertion(t, err), should.BeTrue) {
					a.So(dev, should.BeNil)
				} else if a.So(err, should.BeNil) {
					a.So(dev, should.Resemble, tc.Device)
				}
				a.So(req, should.Resemble, tc.Request)
				a.So(getByIDCalls, should.Equal, tc.GetByIDCalls)
			},
		})
	}
}

func TestDeviceRegistrySet(t *testing.T) {
	defaultMACSettings := DefaultConfig.DefaultMACSettings.Parse()

	customMACSettings := defaultMACSettings
	customMACSettings.Rx1Delay = &ttnpb.RxDelayValue{Value: ttnpb.RxDelay_RX_DELAY_2}
	customMACSettings.Rx1DataRateOffset = nil

	customMACSettingsOpt := EndDeviceOptions.WithMacSettings(&customMACSettings)

	multicastClassBMACSettings := defaultMACSettings
	multicastClassBMACSettings.PingSlotPeriodicity = &ttnpb.PingSlotPeriodValue{
		Value: ttnpb.PingSlotPeriod_PING_EVERY_16S,
	}

	multicastClassBMACSettingsOpt := EndDeviceOptions.WithMacSettings(&multicastClassBMACSettings)

	currentMACStateOverrideOpt := func(macState ttnpb.MACState) ttnpb.MACState {
		macState.CurrentParameters.Rx1Delay = ttnpb.RxDelay_RX_DELAY_3
		macState.CurrentParameters.Rx1DataRateOffset = ttnpb.DataRateOffset_DATA_RATE_OFFSET_1
		return macState
	}
	desiredMACStateOverrideOpt := func(macState ttnpb.MACState) ttnpb.MACState {
		macState.DesiredParameters.Rx1Delay = ttnpb.RxDelay_RX_DELAY_4
		macState.DesiredParameters.Rx1DataRateOffset = ttnpb.DataRateOffset_DATA_RATE_OFFSET_2
		return macState
	}
	activeMACStateOpts := []test.MACStateOption{
		currentMACStateOverrideOpt,
		desiredMACStateOverrideOpt,
	}

	activeSessionOpts := []test.SessionOption{
		SessionOptions.WithLastNFCntDown(0x24),
	}
	activeSessionOptsWithStartedAt := append(activeSessionOpts,
		SessionOptions.WithStartedAt(ttnpb.ProtoTimePtr(time.Unix(0, 42))),
	)

	activateOpt := EndDeviceOptions.Activate(customMACSettings, false, activeSessionOpts, activeMACStateOpts...)

	macStateWithoutRX1DelayOpt := func(dev ttnpb.EndDevice) ttnpb.EndDevice {
		dev.MacState.CurrentParameters.Rx1Delay = 0
		return dev
	}

	makeUpdateDeviceRequest := func(deviceOpts []test.EndDeviceOption, paths ...string) *SetDeviceRequest {
		return &SetDeviceRequest{
			EndDevice: test.MakeEndDevice(deviceOpts...),
			Paths:     paths,
		}
	}

	for createDevice, tcs := range map[*ttnpb.EndDevice][]struct {
		SetDevice      SetDeviceRequest
		RequiredRights []ttnpb.Right

		ReturnedDevice *ttnpb.EndDevice
		StoredDevice   *ttnpb.EndDevice
	}{
		nil: {
			// OTAA Create
			{
				SetDevice: *MakeOTAASetDeviceRequest(nil),

				ReturnedDevice: MakeOTAAEndDevice(),
				StoredDevice:   MakeOTAAEndDevice(),
			},
			{
				SetDevice: *MakeOTAASetDeviceRequest([]test.EndDeviceOption{
					EndDeviceOptions.SendJoinRequest(customMACSettings, false),
				},
					"pending_mac_state",
				),
				RequiredRights: []ttnpb.Right{
					ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE_KEYS,
				},

				ReturnedDevice: MakeOTAAEndDevice(
					EndDeviceOptions.SendJoinRequest(customMACSettings, false),
				),
				StoredDevice: MakeOTAAEndDevice(
					EndDeviceOptions.SendJoinRequest(customMACSettings, true),
				),
			},
			{
				SetDevice: *MakeOTAASetDeviceRequest([]test.EndDeviceOption{
					EndDeviceOptions.SendJoinRequest(customMACSettings, false),
					EndDeviceOptions.SendJoinAccept(ttnpb.TxSchedulePriority_HIGHEST),
				},
					"pending_mac_state",
					"pending_session",
				),
				RequiredRights: []ttnpb.Right{
					ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE_KEYS,
				},

				ReturnedDevice: MakeOTAAEndDevice(
					EndDeviceOptions.SendJoinRequest(customMACSettings, false),
					EndDeviceOptions.SendJoinAccept(ttnpb.TxSchedulePriority_HIGHEST),
				),
				StoredDevice: MakeOTAAEndDevice(
					EndDeviceOptions.SendJoinRequest(customMACSettings, true),
					EndDeviceOptions.SendJoinAccept(ttnpb.TxSchedulePriority_HIGHEST),
				),
			},
			{
				SetDevice: *MakeOTAASetDeviceRequest([]test.EndDeviceOption{
					customMACSettingsOpt,
				},
					"mac_settings",
				),

				ReturnedDevice: MakeOTAAEndDevice(
					customMACSettingsOpt,
				),
				StoredDevice: MakeOTAAEndDevice(
					customMACSettingsOpt,
				),
			},

			{
				SetDevice: *MakeOTAASetDeviceRequest([]test.EndDeviceOption{
					activateOpt,
				},
					"mac_state",
					"session",
				),
				RequiredRights: []ttnpb.Right{
					ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE_KEYS,
				},

				ReturnedDevice: MakeOTAAEndDevice(
					EndDeviceOptions.Activate(customMACSettings, false, activeSessionOpts, activeMACStateOpts...),
				),
				StoredDevice: MakeOTAAEndDevice(
					EndDeviceOptions.Activate(customMACSettings, true, activeSessionOpts, activeMACStateOpts...),
				),
			},
			{
				SetDevice: *MakeOTAASetDeviceRequest([]test.EndDeviceOption{
					activateOpt,
				},
					"mac_state.current_parameters",
					"mac_state.lorawan_version",
					"session",
				),
				RequiredRights: []ttnpb.Right{
					ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE_KEYS,
				},

				ReturnedDevice: MakeOTAAEndDevice(
					EndDeviceOptions.Activate(customMACSettings, false, activeSessionOpts, currentMACStateOverrideOpt),
				),
				StoredDevice: MakeOTAAEndDevice(
					EndDeviceOptions.Activate(customMACSettings, true, activeSessionOpts, currentMACStateOverrideOpt),
					EndDeviceOptions.WithMACStateOptions(
						MACStateOptions.WithRecentUplinks(),
						MACStateOptions.WithRecentDownlinks(),
					),
				),
			},
			{
				SetDevice: *MakeOTAASetDeviceRequest([]test.EndDeviceOption{
					activateOpt,
				},
					"mac_state.desired_parameters",
					"mac_state.lorawan_version",
					"session.dev_addr",
					"session.keys.f_nwk_s_int_key.key",
					"session.keys.nwk_s_enc_key.key",
					"session.keys.s_nwk_s_int_key.key",
					"session.keys.session_key_id",
				),
				RequiredRights: []ttnpb.Right{
					ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE_KEYS,
				},

				ReturnedDevice: MakeOTAAEndDevice(
					EndDeviceOptions.Activate(defaultMACSettings, false, nil, desiredMACStateOverrideOpt),
				),
				StoredDevice: MakeOTAAEndDevice(
					EndDeviceOptions.Activate(defaultMACSettings, true, nil, desiredMACStateOverrideOpt),
					EndDeviceOptions.WithMACStateOptions(
						MACStateOptions.WithRecentUplinks(),
						MACStateOptions.WithRecentDownlinks(),
					),
				),
			},

			// OTAA Create 1.0.3
			{
				SetDevice: *MakeOTAASetDeviceRequest([]test.EndDeviceOption{
					EndDeviceOptions.WithLorawanVersion(ttnpb.MAC_V1_0_3),
					EndDeviceOptions.WithLorawanPhyVersion(ttnpb.PHYVersion_RP001_V1_0_3_REV_A),
					activateOpt,
				},
					"mac_state",
					"session",
				),
				RequiredRights: []ttnpb.Right{
					ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE_KEYS,
				},

				ReturnedDevice: MakeOTAAEndDevice(
					EndDeviceOptions.WithLorawanVersion(ttnpb.MAC_V1_0_3),
					EndDeviceOptions.WithLorawanPhyVersion(ttnpb.PHYVersion_RP001_V1_0_3_REV_A),
					activateOpt,
				),
				StoredDevice: MakeOTAAEndDevice(
					EndDeviceOptions.WithLorawanVersion(ttnpb.MAC_V1_0_3),
					EndDeviceOptions.WithLorawanPhyVersion(ttnpb.PHYVersion_RP001_V1_0_3_REV_A),
					EndDeviceOptions.Activate(customMACSettings, true, activeSessionOpts, activeMACStateOpts...),
				),
			},
			{
				SetDevice: *MakeOTAASetDeviceRequest([]test.EndDeviceOption{
					EndDeviceOptions.WithLorawanVersion(ttnpb.MAC_V1_0_3),
					EndDeviceOptions.WithLorawanPhyVersion(ttnpb.PHYVersion_RP001_V1_0_3_REV_A),
					EndDeviceOptions.SendJoinRequest(customMACSettings, false),
				},
					"pending_mac_state",
				),
				RequiredRights: []ttnpb.Right{
					ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE_KEYS,
				},

				ReturnedDevice: MakeOTAAEndDevice(
					EndDeviceOptions.WithLorawanVersion(ttnpb.MAC_V1_0_3),
					EndDeviceOptions.WithLorawanPhyVersion(ttnpb.PHYVersion_RP001_V1_0_3_REV_A),
					EndDeviceOptions.SendJoinRequest(customMACSettings, false),
				),
				StoredDevice: MakeOTAAEndDevice(
					EndDeviceOptions.WithLorawanVersion(ttnpb.MAC_V1_0_3),
					EndDeviceOptions.WithLorawanPhyVersion(ttnpb.PHYVersion_RP001_V1_0_3_REV_A),
					EndDeviceOptions.SendJoinRequest(customMACSettings, true),
				),
			},
			{
				SetDevice: *MakeOTAASetDeviceRequest([]test.EndDeviceOption{
					EndDeviceOptions.WithLorawanVersion(ttnpb.MAC_V1_0_3),
					EndDeviceOptions.WithLorawanPhyVersion(ttnpb.PHYVersion_RP001_V1_0_3_REV_A),
					EndDeviceOptions.SendJoinRequest(customMACSettings, false),
					EndDeviceOptions.SendJoinAccept(ttnpb.TxSchedulePriority_HIGHEST),
				},
					"pending_mac_state",
					"pending_session",
				),
				RequiredRights: []ttnpb.Right{
					ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE_KEYS,
				},

				ReturnedDevice: MakeOTAAEndDevice(
					EndDeviceOptions.WithLorawanVersion(ttnpb.MAC_V1_0_3),
					EndDeviceOptions.WithLorawanPhyVersion(ttnpb.PHYVersion_RP001_V1_0_3_REV_A),
					EndDeviceOptions.SendJoinRequest(customMACSettings, false),
					EndDeviceOptions.SendJoinAccept(ttnpb.TxSchedulePriority_HIGHEST),
				),
				StoredDevice: MakeOTAAEndDevice(
					EndDeviceOptions.WithLorawanVersion(ttnpb.MAC_V1_0_3),
					EndDeviceOptions.WithLorawanPhyVersion(ttnpb.PHYVersion_RP001_V1_0_3_REV_A),
					EndDeviceOptions.SendJoinRequest(customMACSettings, true),
					EndDeviceOptions.SendJoinAccept(ttnpb.TxSchedulePriority_HIGHEST),
				),
			},
			{
				SetDevice: *MakeOTAASetDeviceRequest([]test.EndDeviceOption{
					EndDeviceOptions.WithLorawanVersion(ttnpb.MAC_V1_0_3),
					EndDeviceOptions.WithLorawanPhyVersion(ttnpb.PHYVersion_RP001_V1_0_3_REV_A),
					activateOpt,
				},
					"mac_state.current_parameters",
					"mac_state.lorawan_version",
					"session",
				),
				RequiredRights: []ttnpb.Right{
					ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE_KEYS,
				},

				ReturnedDevice: MakeOTAAEndDevice(
					EndDeviceOptions.WithLorawanVersion(ttnpb.MAC_V1_0_3),
					EndDeviceOptions.WithLorawanPhyVersion(ttnpb.PHYVersion_RP001_V1_0_3_REV_A),
					EndDeviceOptions.Activate(customMACSettings, false, activeSessionOpts, currentMACStateOverrideOpt),
				),
				StoredDevice: MakeOTAAEndDevice(
					EndDeviceOptions.WithLorawanVersion(ttnpb.MAC_V1_0_3),
					EndDeviceOptions.WithLorawanPhyVersion(ttnpb.PHYVersion_RP001_V1_0_3_REV_A),
					EndDeviceOptions.Activate(customMACSettings, true, activeSessionOpts, currentMACStateOverrideOpt),
					EndDeviceOptions.WithMACStateOptions(
						MACStateOptions.WithRecentUplinks(),
						MACStateOptions.WithRecentDownlinks(),
					),
				),
			},
			{
				SetDevice: *MakeOTAASetDeviceRequest([]test.EndDeviceOption{
					EndDeviceOptions.WithLorawanVersion(ttnpb.MAC_V1_0_3),
					EndDeviceOptions.WithLorawanPhyVersion(ttnpb.PHYVersion_RP001_V1_0_3_REV_A),
					activateOpt,
				},
					"mac_state.desired_parameters",
					"mac_state.lorawan_version",
					"session.dev_addr",
					"session.keys.f_nwk_s_int_key.key",
					"session.keys.session_key_id",
				),
				RequiredRights: []ttnpb.Right{
					ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE_KEYS,
				},

				ReturnedDevice: MakeOTAAEndDevice(
					EndDeviceOptions.WithLorawanVersion(ttnpb.MAC_V1_0_3),
					EndDeviceOptions.WithLorawanPhyVersion(ttnpb.PHYVersion_RP001_V1_0_3_REV_A),
					EndDeviceOptions.Activate(defaultMACSettings, false, nil, desiredMACStateOverrideOpt),
				),
				StoredDevice: MakeOTAAEndDevice(
					EndDeviceOptions.WithLorawanVersion(ttnpb.MAC_V1_0_3),
					EndDeviceOptions.WithLorawanPhyVersion(ttnpb.PHYVersion_RP001_V1_0_3_REV_A),
					EndDeviceOptions.Activate(defaultMACSettings, true, nil, desiredMACStateOverrideOpt),
					EndDeviceOptions.WithMACStateOptions(
						MACStateOptions.WithRecentUplinks(),
						MACStateOptions.WithRecentDownlinks(),
					),
				),
			},

			// ABP Create
			{
				SetDevice: *MakeABPSetDeviceRequest(customMACSettings, activeSessionOpts, nil, nil,
					"mac_state.current_parameters.rx1_delay",
					"session",
				),
				RequiredRights: []ttnpb.Right{
					ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE_KEYS,
				},

				ReturnedDevice: MakeABPEndDevice(customMACSettings, false, activeSessionOpts, nil),
				StoredDevice:   MakeABPEndDevice(customMACSettings, true, activeSessionOpts, nil),
			},

			// Multicast Create
			{
				SetDevice: *MakeMulticastSetDeviceRequest(ttnpb.Class_CLASS_C, defaultMACSettings, activeSessionOpts, nil, nil,
					"session.last_n_f_cnt_down",
				),
				RequiredRights: []ttnpb.Right{
					ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE_KEYS,
				},

				ReturnedDevice: MakeMulticastEndDevice(ttnpb.Class_CLASS_C, defaultMACSettings, false, activeSessionOpts, nil),
				StoredDevice:   MakeMulticastEndDevice(ttnpb.Class_CLASS_C, defaultMACSettings, true, activeSessionOpts, nil),
			},
		},

		// OTAA Update
		MakeOTAAEndDevice(): {
			{
				SetDevice: *makeUpdateDeviceRequest([]test.EndDeviceOption{
					customMACSettingsOpt,
				},
					"mac_settings",
				),

				ReturnedDevice: MakeOTAAEndDevice(
					customMACSettingsOpt,
				),
				StoredDevice: MakeOTAAEndDevice(
					customMACSettingsOpt,
				),
			},
		},

		// ABP Update
		MakeABPEndDevice(defaultMACSettings, true, activeSessionOptsWithStartedAt, nil): {
			{
				SetDevice: *makeUpdateDeviceRequest([]test.EndDeviceOption{
					customMACSettingsOpt,
				},
					"mac_settings",
				),

				ReturnedDevice: EndDevicePtr(customMACSettingsOpt(*MakeABPEndDevice(defaultMACSettings, false, activeSessionOptsWithStartedAt, nil))),
				StoredDevice:   EndDevicePtr(customMACSettingsOpt(*MakeABPEndDevice(defaultMACSettings, true, activeSessionOptsWithStartedAt, nil))),
			},

			{
				SetDevice: *makeUpdateDeviceRequest(nil,
					"mac_settings.rx2_data_rate_index",
					"mac_state.current_parameters.rx1_delay",
					"pending_mac_state",
				),
				RequiredRights: []ttnpb.Right{
					ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE_KEYS, // `pending_mac_state` requires key write rights
				},

				ReturnedDevice: EndDevicePtr(macStateWithoutRX1DelayOpt(*MakeABPEndDevice(defaultMACSettings, false, activeSessionOptsWithStartedAt, nil))),
				StoredDevice:   EndDevicePtr(macStateWithoutRX1DelayOpt(*MakeABPEndDevice(defaultMACSettings, true, activeSessionOptsWithStartedAt, nil))),
			},
		},

		// Multicast Update
		MakeMulticastEndDevice(ttnpb.Class_CLASS_B, defaultMACSettings, true, activeSessionOptsWithStartedAt, nil): {
			{
				SetDevice: *makeUpdateDeviceRequest([]test.EndDeviceOption{
					multicastClassBMACSettingsOpt,
				},
					"mac_settings",
				),

				ReturnedDevice: EndDevicePtr(multicastClassBMACSettingsOpt(*MakeMulticastEndDevice(ttnpb.Class_CLASS_B, defaultMACSettings, false, activeSessionOptsWithStartedAt, nil))),
				StoredDevice:   EndDevicePtr(multicastClassBMACSettingsOpt(*MakeMulticastEndDevice(ttnpb.Class_CLASS_B, defaultMACSettings, true, activeSessionOptsWithStartedAt, nil))),
			},
		},
	} {
		for _, tc := range tcs {
			createDevice := createDevice
			tc := tc
			test.RunSubtest(t, test.SubtestConfig{
				Name: MakeTestCaseName(func() []string {
					dev := createDevice
					typ := "Update"
					if createDevice == nil {
						dev = tc.SetDevice.EndDevice
						typ = "Create"
					}
					return []string{
						typ,
						fmt.Sprintf("mode:%s", func() string {
							switch {
							case dev.SupportsJoin:
								return "OTAA"
							case dev.Multicast:
								return "Multicast"
							default:
								return "ABP"
							}
						}()),
						fmt.Sprintf("MAC:%s", dev.LorawanVersion.String()),
						fmt.Sprintf("PHY:%s", dev.LorawanPhyVersion.String()),
						fmt.Sprintf("fp:%s", dev.FrequencyPlanId),
						fmt.Sprintf("paths:[%s]", strings.Join(tc.SetDevice.Paths, ",")),
					}
				}()...),
				Parallel: true,
				Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
					nsConf := DefaultConfig
					nsConf.DeviceKEKLabel = test.DefaultKEKLabel

					_, ctx, env, stop := StartTest(ctx, TestConfig{
						Component: component.Config{
							ServiceBase: config.ServiceBase{
								GRPC: config.GRPC{
									LogIgnoreMethods: []string{
										"/ttn.lorawan.v3.ApplicationAccess/ListRights",
										"/ttn.lorawan.v3.NsEndDeviceRegistry/Set",
									},
								},
								KeyVault: test.DefaultKeyVault,
								FrequencyPlans: config.FrequencyPlansConfig{
									ConfigSource: "static",
									Static:       test.StaticFrequencyPlans,
								},
							},
						},
						NetworkServer: nsConf,
						TaskStarter: StartTaskExclude(
							DownlinkProcessTaskName,
						),
					})
					defer stop()

					clock := test.NewMockClock(time.Now())
					defer SetMockClock(clock)()

					withCreatedAt := test.EndDeviceOptions.WithCreatedAt(ttnpb.ProtoTimePtr(clock.Now()))
					if createDevice != nil {
						_, ctx = MustCreateDevice(ctx, env.Devices, createDevice)
						clock.Add(time.Nanosecond)
					}

					now := clock.Now()
					withTimestamps := withCreatedAt.Compose(
						test.EndDeviceOptions.WithUpdatedAt(ttnpb.ProtoTimePtr(now)),
						func(dev ttnpb.EndDevice) ttnpb.EndDevice {
							if dev.Session != nil && dev.Session.StartedAt == nil {
								dev.Session = CopySession(dev.Session)
								dev.Session.StartedAt = ttnpb.ProtoTimePtr(now)
							}
							return dev
						},
					)

					req := &ttnpb.SetEndDeviceRequest{
						EndDevice: *tc.SetDevice.EndDevice,
						FieldMask: &pbtypes.FieldMask{
							Paths: tc.SetDevice.Paths,
						},
					}

					dev, err, ok := env.AssertSetDevice(ctx, createDevice == nil, req)
					if !a.So(ok, should.BeTrue) || !a.So(err, should.BeError) || !a.So(errors.IsPermissionDenied(err), should.BeTrue) {
						if err != nil {
							t.Errorf("Expected 'permission denied' error, got: %s", test.FormatError(err))
						}
						return
					}
					a.So(dev, should.BeNil)
					if len(tc.RequiredRights) > 0 {
						dev, err, ok = env.AssertSetDevice(ctx, createDevice == nil, req,
							ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE,
						)
						if !a.So(ok, should.BeTrue) || !a.So(err, should.BeError) || !a.So(errors.IsPermissionDenied(err), should.BeTrue) {
							if err != nil {
								t.Errorf("Expected 'permission denied' error, got: %s", test.FormatError(err))
							}
							return
						}
						a.So(dev, should.BeNil)
					}

					rights := append([]ttnpb.Right{
						ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE,
					}, tc.RequiredRights...)
					expectedReturn := test.Must(ttnpb.ApplyEndDeviceFieldMask(nil, EndDevicePtr(withTimestamps(*tc.ReturnedDevice)), ttnpb.AddImplicitEndDeviceGetFields(tc.SetDevice.Paths...)...)).(*ttnpb.EndDevice)

					dev, err, ok = env.AssertSetDevice(ctx, createDevice == nil, req, rights...)
					if !a.So(ok, should.BeTrue) || !a.So(err, should.BeNil) || !a.So(dev, should.NotBeNil) {
						if err != nil {
							t.Errorf("Expected no error, got: %s", test.FormatError(err))
						}
						return
					}
					a.So(dev, should.Resemble, expectedReturn)

					dev, _, err = env.Devices.GetByID(ctx, tc.SetDevice.Ids.ApplicationIds, tc.SetDevice.Ids.DeviceId, ttnpb.EndDeviceFieldPathsTopLevel)
					if !a.So(err, should.BeNil) || !a.So(dev, should.NotBeNil) {
						if err != nil {
							t.Errorf("Expected no error, got: %s", test.FormatError(err))
						}
						return
					}
					a.So(dev, should.Resemble, EndDevicePtr(withTimestamps(*tc.StoredDevice)))

					now = clock.Add(time.Nanosecond)
					dev, err, ok = env.AssertSetDevice(ctx, false, &ttnpb.SetEndDeviceRequest{
						EndDevice: *expectedReturn,
						FieldMask: &pbtypes.FieldMask{
							Paths: tc.SetDevice.Paths,
						},
					}, rights...)
					if !a.So(ok, should.BeTrue) || !a.So(err, should.BeNil) || !a.So(dev, should.NotBeNil) {
						return
					}
					a.So(dev, should.Resemble, EndDevicePtr(EndDeviceOptions.WithUpdatedAt(ttnpb.ProtoTimePtr(now))(*expectedReturn)))
				},
			})
		}
	}
}

func TestDeviceRegistryResetFactoryDefaults(t *testing.T) {
	activeSessionOpts := []test.SessionOption{
		SessionOptions.WithLastFCntUp(0x42),
		SessionOptions.WithLastNFCntDown(0x24),
		SessionOptions.WithDefaultQueuedApplicationDownlinks(),
	}
	macSettings := DefaultConfig.DefaultMACSettings.Parse()
	activateOpt := EndDeviceOptions.Activate(macSettings, true, activeSessionOpts)

	// TODO: Refactor into same structure as Set
	for _, tc := range []struct {
		CreateDevice *SetDeviceRequest
	}{
		{},

		{
			CreateDevice: MakeOTAASetDeviceRequest(nil),
		},
		{
			CreateDevice: MakeOTAASetDeviceRequest([]test.EndDeviceOption{
				activateOpt,
			},
				"mac_state",
				"session",
			),
		},
		{
			CreateDevice: MakeOTAASetDeviceRequest([]test.EndDeviceOption{
				EndDeviceOptions.WithLorawanVersion(ttnpb.MAC_V1_0_3),
				EndDeviceOptions.WithLorawanPhyVersion(ttnpb.PHYVersion_RP001_V1_0_3_REV_A),
				activateOpt,
			},
				"mac_state",
				"session",
			),
		},

		{
			CreateDevice: MakeABPSetDeviceRequest(macSettings, nil, nil, nil),
		},
		{
			CreateDevice: MakeABPSetDeviceRequest(macSettings, activeSessionOpts, nil, nil),
		},
		{
			CreateDevice: MakeABPSetDeviceRequest(macSettings, activeSessionOpts, nil, []test.EndDeviceOption{
				EndDeviceOptions.WithLorawanVersion(ttnpb.MAC_V1_0_3),
				EndDeviceOptions.WithLorawanPhyVersion(ttnpb.PHYVersion_RP001_V1_0_3_REV_A),
			}),
		},
	} {
		for _, conf := range []struct {
			Paths          []string
			RequiredRights []ttnpb.Right
		}{
			{},
			{
				Paths: []string{
					"battery_percentage",
					"downlink_margin",
					"last_dev_status_received_at",
					"mac_state.current_parameters",
					"session.last_f_cnt_up",
				},
				RequiredRights: []ttnpb.Right{
					ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ,
				},
			},
			{
				Paths: []string{
					"battery_percentage",
					"session.last_f_cnt_up",
				},
				RequiredRights: []ttnpb.Right{
					ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ,
				},
			},
			{
				Paths: []string{
					"session.keys",
				},
				RequiredRights: []ttnpb.Right{
					ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ,
					ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ_KEYS,
				},
			},
			{
				Paths: []string{
					"battery_percentage",
					"downlink_margin",
					"last_dev_status_received_at",
					"pending_mac_state",
					"pending_session",
				},
				RequiredRights: []ttnpb.Right{
					ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ,
					ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ_KEYS,
					ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ,
				},
			},
			{
				Paths: []string{
					"battery_percentage",
					"downlink_margin",
					"last_dev_status_received_at",
					"mac_state",
					"pending_mac_state",
					"pending_session",
					"session",
					"supports_join",
				},
				RequiredRights: []ttnpb.Right{
					ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ,
					ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ_KEYS,
					ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ,
				},
			},
		} {
			tc := tc
			conf := conf
			test.RunSubtest(t, test.SubtestConfig{
				Name: func() string {
					if tc.CreateDevice == nil {
						return "no device"
					}
					return MakeTestCaseName(
						fmt.Sprintf("paths:[%s]", strings.Join(conf.Paths, ",")),
						func() string {
							if tc.CreateDevice.EndDevice.SupportsJoin {
								return "OTAA"
							}
							if tc.CreateDevice.EndDevice.Session == nil {
								return MakeTestCaseName("ABP", "no session")
							}
							return fmt.Sprintf(MakeTestCaseName("ABP", "dev_addr:%s", "queue_len:%d", "session_keys:%v"),
								tc.CreateDevice.Session.DevAddr,
								len(tc.CreateDevice.EndDevice.Session.QueuedApplicationDownlinks),
								tc.CreateDevice.Session.Keys,
							)
						}(),
					)
				}(),
				Parallel: true,
				Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
					nsConf := DefaultConfig
					nsConf.DeviceKEKLabel = test.DefaultKEKLabel

					ns, ctx, env, stop := StartTest(ctx, TestConfig{
						Component: component.Config{
							ServiceBase: config.ServiceBase{
								GRPC: config.GRPC{
									LogIgnoreMethods: []string{
										"/ttn.lorawan.v3.ApplicationAccess/ListRights",
										"/ttn.lorawan.v3.NsEndDeviceRegistry/ResetFactoryDefaults",
									},
								},
								KeyVault: test.DefaultKeyVault,
								FrequencyPlans: config.FrequencyPlansConfig{
									ConfigSource: "static",
									Static:       test.StaticFrequencyPlans,
								},
							},
						},
						NetworkServer: nsConf,
						TaskStarter: StartTaskExclude(
							DownlinkProcessTaskName,
						),
					})
					defer stop()

					clock := test.NewMockClock(time.Now().UTC())
					defer SetMockClock(clock)()

					req := &ttnpb.ResetAndGetEndDeviceRequest{
						EndDeviceIds: test.MakeEndDeviceIdentifiers(),
						FieldMask: &pbtypes.FieldMask{
							Paths: conf.Paths,
						},
					}

					var created *ttnpb.EndDevice
					if tc.CreateDevice != nil {
						created, ctx = MustCreateDevice(ctx, env.Devices, tc.CreateDevice.EndDevice)

						req.EndDeviceIds.ApplicationIds = tc.CreateDevice.Ids.ApplicationIds
						req.EndDeviceIds.DeviceId = tc.CreateDevice.Ids.DeviceId

						clock.Add(time.Nanosecond)
					}

					dev, err, ok := env.AssertResetFactoryDefaults(ctx, req)
					if !a.So(ok, should.BeTrue) {
						return
					}
					a.So(dev, should.BeNil)
					if !a.So(err, should.BeError) || !a.So(errors.IsPermissionDenied(err), should.BeTrue) {
						t.Errorf("Expected 'permission denied' error, got: %s", test.FormatError(err))
						return
					}

					now := clock.Now().UTC()

					dev, err, ok = env.AssertResetFactoryDefaults(ctx, req, append([]ttnpb.Right{
						ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE,
					}, conf.RequiredRights...)...)
					if !a.So(ok, should.BeTrue) {
						return
					}
					if created == nil {
						a.So(err, should.NotBeNil)
						if !a.So(errors.IsNotFound(err), should.BeTrue) {
							t.Errorf("Expected 'not found' error, got: %s", test.FormatError(err))
						}
						return
					}

					var (
						macState *ttnpb.MACState
						session  *ttnpb.Session
					)
					if !created.SupportsJoin {
						if created.Session == nil {
							a.So(err, should.NotBeNil)
							if !a.So(errors.IsDataLoss(err), should.BeTrue) {
								t.Errorf("Expected 'data loss' error, got: %s", test.FormatError(err))
							}
							return
						}

						fps, err := ns.FrequencyPlansStore(ctx)
						if !a.So(err, should.BeNil) {
							t.Fail()
							return
						}
						var newErr error
						macState, newErr = mac.NewState(created, fps, DefaultConfig.DefaultMACSettings.Parse())
						if newErr != nil {
							a.So(err, should.NotBeNil)
							a.So(err, should.HaveSameErrorDefinitionAs, newErr)
							return
						}
						session = &ttnpb.Session{
							DevAddr:                    created.Session.DevAddr,
							QueuedApplicationDownlinks: created.Session.QueuedApplicationDownlinks,
							Keys:                       created.Session.Keys,
							StartedAt:                  ttnpb.ProtoTimePtr(now),
						}
					}
					if !a.So(err, should.BeNil) {
						t.Errorf("Expected no error, got: %s", test.FormatError(err))
						return
					}

					expected := CopyEndDevice(created)
					expected.BatteryPercentage = nil
					expected.DownlinkMargin = 0
					expected.LastDevStatusReceivedAt = nil
					expected.MacState = macState
					expected.PendingMacState = nil
					expected.PendingSession = nil
					expected.PowerState = ttnpb.PowerState_POWER_UNKNOWN
					expected.Session = session
					expected.UpdatedAt = ttnpb.ProtoTimePtr(clock.Now())
					if !a.So(dev, should.Resemble, test.Must(ttnpb.ApplyEndDeviceFieldMask(nil, expected, ttnpb.AddImplicitEndDeviceGetFields(conf.Paths...)...)).(*ttnpb.EndDevice)) {
						return
					}
					updated, _, err := env.Devices.GetByID(ctx, tc.CreateDevice.Ids.ApplicationIds, tc.CreateDevice.Ids.DeviceId, ttnpb.EndDeviceFieldPathsTopLevel)
					if a.So(err, should.BeNil) {
						a.So(updated, should.Resemble, expected)
					}
				},
			})
		}
	}
}

func TestDeviceRegistryDelete(t *testing.T) {
	for _, tc := range []struct {
		Name           string
		ContextFunc    func(context.Context) context.Context
		SetByIDFunc    func(context.Context, *ttnpb.ApplicationIdentifiers, string, []string, func(context.Context, *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, context.Context, error)
		Request        *ttnpb.EndDeviceIdentifiers
		ErrorAssertion func(*testing.T, error) bool
		SetByIDCalls   uint64
	}{
		{
			Name: "No device write rights",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(test.Context(), ttnpb.ApplicationIdentifiers{ApplicationId: "test-app-id"}): {
							Rights: []ttnpb.Right{
								ttnpb.Right_RIGHT_GATEWAY_SETTINGS_BASIC,
							},
						},
					},
				})
			},
			SetByIDFunc: func(ctx context.Context, appID *ttnpb.ApplicationIdentifiers, devID string, gets []string, f func(context.Context, *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, context.Context, error) {
				err := errors.New("SetByIDFunc must not be called")
				test.MustTFromContext(ctx).Error(err)
				return nil, ctx, err
			},
			Request: &ttnpb.EndDeviceIdentifiers{
				DeviceId:       "test-dev-id",
				ApplicationIds: &ttnpb.ApplicationIdentifiers{ApplicationId: "test-app-id"},
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
						unique.ID(test.Context(), ttnpb.ApplicationIdentifiers{ApplicationId: "test-app-id"}): {
							Rights: []ttnpb.Right{
								ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE,
							},
						},
					},
				})
			},
			SetByIDFunc: func(ctx context.Context, appID *ttnpb.ApplicationIdentifiers, devID string, gets []string, f func(context.Context, *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, context.Context, error) {
				t := test.MustTFromContext(ctx)
				a := assertions.New(t)
				a.So(appID, should.Resemble, ttnpb.ApplicationIdentifiers{ApplicationId: "test-app-id"})
				a.So(devID, should.Equal, "test-dev-id")
				a.So(gets, should.BeNil)

				dev, sets, err := f(ctx, nil)
				if !a.So(errors.IsNotFound(err), should.BeTrue) {
					return nil, ctx, err
				}
				a.So(sets, should.BeNil)
				a.So(dev, should.BeNil)
				return nil, ctx, nil
			},
			Request: &ttnpb.EndDeviceIdentifiers{
				DeviceId:       "test-dev-id",
				ApplicationIds: &ttnpb.ApplicationIdentifiers{ApplicationId: "test-app-id"},
			},
			SetByIDCalls: 1,
		},

		{
			Name: "Existing device",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(test.Context(), ttnpb.ApplicationIdentifiers{ApplicationId: "test-app-id"}): {
							Rights: []ttnpb.Right{
								ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE,
							},
						},
					},
				})
			},
			SetByIDFunc: func(ctx context.Context, appID *ttnpb.ApplicationIdentifiers, devID string, gets []string, f func(context.Context, *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, context.Context, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(appID, should.Resemble, ttnpb.ApplicationIdentifiers{ApplicationId: "test-app-id"})
				a.So(devID, should.Equal, "test-dev-id")
				a.So(gets, should.BeNil)

				dev, sets, err := f(ctx, &ttnpb.EndDevice{
					Ids: &ttnpb.EndDeviceIdentifiers{
						DeviceId:       "test-dev-id",
						ApplicationIds: &ttnpb.ApplicationIdentifiers{ApplicationId: "test-app-id"},
					},
				})
				if !a.So(err, should.BeNil) {
					return nil, ctx, err
				}
				a.So(sets, should.BeNil)
				a.So(dev, should.BeNil)
				return nil, ctx, nil
			},
			Request: &ttnpb.EndDeviceIdentifiers{
				DeviceId:       "test-dev-id",
				ApplicationIds: &ttnpb.ApplicationIdentifiers{ApplicationId: "test-app-id"},
			},
			SetByIDCalls: 1,
		},
	} {
		tc := tc
		test.RunSubtest(t, test.SubtestConfig{
			Name:     tc.Name,
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				var setByIDCalls uint64

				ns, ctx, env, stop := StartTest(
					ctx,
					TestConfig{
						NetworkServer: Config{
							Devices: &MockDeviceRegistry{
								SetByIDFunc: func(ctx context.Context, appID *ttnpb.ApplicationIdentifiers, devID string, gets []string, f func(context.Context, *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, context.Context, error) {
									atomic.AddUint64(&setByIDCalls, 1)
									return tc.SetByIDFunc(ctx, appID, devID, gets, f)
								},
							},
						},
						TaskStarter: StartTaskExclude(
							DownlinkProcessTaskName,
						),
					},
				)
				defer stop()

				go LogEvents(t, env.Events)

				ns.AddContextFiller(tc.ContextFunc)
				ns.AddContextFiller(func(ctx context.Context) context.Context {
					return test.ContextWithTB(ctx, t)
				})

				req := deepcopy.Copy(tc.Request).(*ttnpb.EndDeviceIdentifiers)
				res, err := ttnpb.NewNsEndDeviceRegistryClient(ns.LoopbackConn()).Delete(ctx, req)
				a.So(setByIDCalls, should.Equal, tc.SetByIDCalls)
				if tc.ErrorAssertion != nil && a.So(tc.ErrorAssertion(t, err), should.BeTrue) {
					a.So(res, should.BeNil)
				} else if a.So(err, should.BeNil) {
					a.So(res, should.Resemble, ttnpb.Empty)
				}
				a.So(req, should.Resemble, tc.Request)
			},
		})
	}
}
