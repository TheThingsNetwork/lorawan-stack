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
		GetByIDFunc    func(context.Context, ttnpb.ApplicationIdentifiers, string, []string) (*ttnpb.EndDevice, context.Context, error)
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
						unique.ID(test.Context(), ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"}): {
							Rights: []ttnpb.Right{
								ttnpb.RIGHT_GATEWAY_SETTINGS_BASIC,
							},
						},
					},
				})
			},
			GetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string) (*ttnpb.EndDevice, context.Context, error) {
				err := errors.New("GetByIDFunc must not be called")
				test.MustTFromContext(ctx).Error(err)
				return nil, ctx, err
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
			Name: "no keys",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(test.Context(), ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"}): {
							Rights: []ttnpb.Right{
								ttnpb.RIGHT_APPLICATION_DEVICES_READ,
							},
						},
					},
				})
			},
			GetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string) (*ttnpb.EndDevice, context.Context, error) {
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
					FrequencyPlanID: test.EUFrequencyPlanID,
				}, ctx, nil
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
				FrequencyPlanID: test.EUFrequencyPlanID,
			},
			GetByIDCalls: 1,
		},

		{
			Name: "with keys",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(test.Context(), ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"}): {
							Rights: []ttnpb.Right{
								ttnpb.RIGHT_APPLICATION_DEVICES_READ,
								ttnpb.RIGHT_APPLICATION_DEVICES_READ_KEYS,
								ttnpb.RIGHT_APPLICATION_LINK,
							},
						},
					},
				})
			},
			GetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string) (*ttnpb.EndDevice, context.Context, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(appID, should.Resemble, ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"})
				a.So(devID, should.Equal, "test-dev-id")
				a.So(gets, should.HaveSameElementsDeep, []string{
					"frequency_plan_id",
					"session",
					"queued_application_downlinks",
				})
				return &ttnpb.EndDevice{
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						DeviceID:               "test-dev-id",
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
					},
					FrequencyPlanID: test.EUFrequencyPlanID,
					Session: &ttnpb.Session{
						SessionKeys: ttnpb.SessionKeys{
							FNwkSIntKey: &ttnpb.KeyEnvelope{
								KEKLabel:     "test",
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
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DeviceID:               "test-dev-id",
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
				},
				FieldMask: pbtypes.FieldMask{
					Paths: []string{
						"frequency_plan_id",
						"session",
					},
				},
			},
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DeviceID:               "test-dev-id",
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
				},
				FrequencyPlanID: test.EUFrequencyPlanID,
				Session: &ttnpb.Session{
					SessionKeys: ttnpb.SessionKeys{
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
						unique.ID(test.Context(), ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"}): {
							Rights: []ttnpb.Right{
								ttnpb.RIGHT_APPLICATION_DEVICES_READ,
								ttnpb.RIGHT_APPLICATION_DEVICES_READ_KEYS,
							},
						},
					},
				})
			},
			GetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string) (*ttnpb.EndDevice, context.Context, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(appID, should.Resemble, ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"})
				a.So(devID, should.Equal, "test-dev-id")
				a.So(gets, should.HaveSameElementsDeep, []string{
					"pending_session.keys.f_nwk_s_int_key",
				})
				return &ttnpb.EndDevice{
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						DeviceID:               "test-dev-id",
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
					},
					PendingSession: &ttnpb.Session{
						SessionKeys: ttnpb.SessionKeys{
							FNwkSIntKey: &ttnpb.KeyEnvelope{
								KEKLabel:     "test",
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
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DeviceID:               "test-dev-id",
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
				},
				FieldMask: pbtypes.FieldMask{
					Paths: []string{
						"pending_session.keys.f_nwk_s_int_key",
					},
				},
			},
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DeviceID:               "test-dev-id",
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
				},
				PendingSession: &ttnpb.Session{
					SessionKeys: ttnpb.SessionKeys{
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
						unique.ID(test.Context(), ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"}): {
							Rights: []ttnpb.Right{
								ttnpb.RIGHT_APPLICATION_DEVICES_READ,
								ttnpb.RIGHT_APPLICATION_DEVICES_READ_KEYS,
							},
						},
					},
				})
			},
			GetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string) (*ttnpb.EndDevice, context.Context, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(appID, should.Resemble, ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"})
				a.So(devID, should.Equal, "test-dev-id")
				a.So(gets, should.HaveSameElementsDeep, []string{
					"pending_session.keys.f_nwk_s_int_key.encrypted_key",
					"pending_session.keys.f_nwk_s_int_key.kek_label",
					"pending_session.keys.f_nwk_s_int_key.key",
				})
				return &ttnpb.EndDevice{
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						DeviceID:               "test-dev-id",
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
					},
					PendingSession: &ttnpb.Session{
						SessionKeys: ttnpb.SessionKeys{
							FNwkSIntKey: &ttnpb.KeyEnvelope{
								KEKLabel:     "test",
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
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DeviceID:               "test-dev-id",
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
				},
				FieldMask: pbtypes.FieldMask{
					Paths: []string{
						"pending_session.keys.f_nwk_s_int_key.key",
					},
				},
			},
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DeviceID:               "test-dev-id",
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
				},
				PendingSession: &ttnpb.Session{
					SessionKeys: ttnpb.SessionKeys{
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
								GetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string) (*ttnpb.EndDevice, context.Context, error) {
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

	customMacSettings := defaultMACSettings
	customMacSettings.Rx1Delay = &ttnpb.RxDelayValue{Value: ttnpb.RX_DELAY_5}
	customMacSettings.Rx1DataRateOffset = nil

	macSettingsOpt := EndDeviceOptions.WithMACSettings(&customMacSettings)

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
		// OTAA Update
		MakeOTAAEndDevice(): {
			{
				SetDevice: *makeUpdateDeviceRequest([]test.EndDeviceOption{
					macSettingsOpt,
				},
					"mac_settings",
				),

				ReturnedDevice: MakeOTAAEndDevice(
					macSettingsOpt,
				),
				StoredDevice: MakeOTAAEndDevice(
					macSettingsOpt,
				),
			},
		},
	} {
		for _, tc := range tcs {
			createDevice := createDevice
			tc := tc
			test.RunSubtest(t, test.SubtestConfig{
				Name: MakeTestCaseName(
					func() string {
						if createDevice != nil {
							return "Update"
						}
						return "Create"
					}(),
					func() string {
						if tc.ReturnedDevice.SupportsJoin {
							return "OTAA"
						}
						if tc.ReturnedDevice.Multicast {
							return "Multicast"
						}
						return "ABP"
					}(),
					tc.ReturnedDevice.LoRaWANVersion.String(),
					fmt.Sprintf("paths:[%s]", strings.Join(tc.SetDevice.Paths, ",")),
				),
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

					withCreatedAt := test.EndDeviceOptions.WithCreatedAt(clock.Now())
					if createDevice != nil {
						_, ctx = MustCreateDevice(ctx, env.Devices, createDevice)
						clock.Add(time.Nanosecond)
					}

					now := clock.Now()
					withTimestamps := withCreatedAt.Compose(
						test.EndDeviceOptions.WithUpdatedAt(now),
						func(dev ttnpb.EndDevice) ttnpb.EndDevice {
							if dev.Session != nil && dev.Session.StartedAt.IsZero() {
								dev.Session = CopySession(dev.Session)
								dev.Session.StartedAt = now
							}
							return dev
						},
					)

					req := &ttnpb.SetEndDeviceRequest{
						EndDevice: *tc.SetDevice.EndDevice,
						FieldMask: pbtypes.FieldMask{
							Paths: tc.SetDevice.Paths,
						},
					}

					dev, err, ok := env.AssertSetDevice(ctx, createDevice == nil, req)
					if !a.So(ok, should.BeTrue) {
						return
					}
					a.So(dev, should.BeNil)
					if !a.So(err, should.BeError) || !a.So(errors.IsPermissionDenied(err), should.BeTrue) {
						t.Errorf("Expected 'permission denied' error, got: %s", test.FormatError(err))
						return
					}
					if len(tc.RequiredRights) > 0 {
						dev, err, ok = env.AssertSetDevice(ctx, createDevice == nil, req,
							ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
						)
						if !a.So(ok, should.BeTrue) {
							return
						}
						a.So(dev, should.BeNil)
						if !a.So(err, should.BeError) || !a.So(errors.IsPermissionDenied(err), should.BeTrue) {
							t.Errorf("Expected 'permission denied' error, got: %s", test.FormatError(err))
							return
						}
					}

					rights := append([]ttnpb.Right{
						ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
					}, tc.RequiredRights...)
					expectedReturn := test.Must(ttnpb.ApplyEndDeviceFieldMask(nil, EndDevicePtr(withTimestamps(*tc.ReturnedDevice)), ttnpb.AddImplicitEndDeviceGetFields(tc.SetDevice.Paths...)...)).(*ttnpb.EndDevice)

					dev, err, ok = env.AssertSetDevice(ctx, createDevice == nil, req, rights...)
					if !a.So(ok, should.BeTrue) || !a.So(err, should.BeNil) || !a.So(dev, should.NotBeNil) {
						return
					}
					a.So(dev, should.Resemble, expectedReturn)

					dev, _, err = env.Devices.GetByID(ctx, tc.SetDevice.ApplicationIdentifiers, tc.SetDevice.DeviceID, ttnpb.EndDeviceFieldPathsTopLevel)
					if !a.So(err, should.BeNil) || !a.So(dev, should.NotBeNil) {
						return
					}
					a.So(dev, should.Resemble, EndDevicePtr(withTimestamps(*tc.StoredDevice)))

					now = clock.Add(time.Nanosecond)
					dev, err, ok = env.AssertSetDevice(ctx, false, &ttnpb.SetEndDeviceRequest{
						EndDevice: *expectedReturn,
						FieldMask: pbtypes.FieldMask{
							Paths: tc.SetDevice.Paths,
						},
					}, rights...)
					if !a.So(ok, should.BeTrue) || !a.So(err, should.BeNil) || !a.So(dev, should.NotBeNil) {
						return
					}
					a.So(dev, should.Resemble, EndDevicePtr(EndDeviceOptions.WithUpdatedAt(now)(*expectedReturn)))
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
				EndDeviceOptions.WithLoRaWANVersion(ttnpb.MAC_V1_0_3),
				EndDeviceOptions.WithLoRaWANPHYVersion(ttnpb.PHY_V1_0_3_REV_A),
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
				EndDeviceOptions.WithLoRaWANVersion(ttnpb.MAC_V1_0_3),
				EndDeviceOptions.WithLoRaWANPHYVersion(ttnpb.PHY_V1_0_3_REV_A),
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
					ttnpb.RIGHT_APPLICATION_DEVICES_READ,
				},
			},
			{
				Paths: []string{
					"battery_percentage",
					"session.last_f_cnt_up",
				},
				RequiredRights: []ttnpb.Right{
					ttnpb.RIGHT_APPLICATION_DEVICES_READ,
					ttnpb.RIGHT_APPLICATION_LINK,
				},
			},
			{
				Paths: []string{
					"session.keys",
				},
				RequiredRights: []ttnpb.Right{
					ttnpb.RIGHT_APPLICATION_DEVICES_READ,
					ttnpb.RIGHT_APPLICATION_DEVICES_READ_KEYS,
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
					ttnpb.RIGHT_APPLICATION_DEVICES_READ,
					ttnpb.RIGHT_APPLICATION_DEVICES_READ_KEYS,
					ttnpb.RIGHT_APPLICATION_LINK,
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
					ttnpb.RIGHT_APPLICATION_DEVICES_READ,
					ttnpb.RIGHT_APPLICATION_DEVICES_READ_KEYS,
					ttnpb.RIGHT_APPLICATION_LINK,
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
								tc.CreateDevice.Session.SessionKeys,
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
						EndDeviceIdentifiers: *test.MakeEndDeviceIdentifiers(),
						FieldMask: pbtypes.FieldMask{
							Paths: conf.Paths,
						},
					}

					var created *ttnpb.EndDevice
					if tc.CreateDevice != nil {
						created, ctx = MustCreateDevice(ctx, env.Devices, tc.CreateDevice.EndDevice)

						req.ApplicationIdentifiers = tc.CreateDevice.ApplicationIdentifiers
						req.DeviceID = tc.CreateDevice.DeviceID

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
						ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
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

						var newErr error
						macState, newErr = mac.NewState(created, ns.FrequencyPlans, DefaultConfig.DefaultMACSettings.Parse())
						if newErr != nil {
							a.So(err, should.NotBeNil)
							a.So(err, should.HaveSameErrorDefinitionAs, newErr)
							return
						}
						session = &ttnpb.Session{
							DevAddr:                    created.Session.DevAddr,
							QueuedApplicationDownlinks: created.Session.QueuedApplicationDownlinks,
							SessionKeys:                created.Session.SessionKeys,
							StartedAt:                  now,
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
					expected.MACState = macState
					expected.PendingMACState = nil
					expected.PendingSession = nil
					expected.PowerState = ttnpb.PowerState_POWER_UNKNOWN
					expected.Session = session
					expected.UpdatedAt = clock.Now().UTC()
					if !a.So(dev, should.Resemble, test.Must(ttnpb.ApplyEndDeviceFieldMask(nil, expected, ttnpb.AddImplicitEndDeviceGetFields(conf.Paths...)...)).(*ttnpb.EndDevice)) {
						return
					}
					updated, _, err := env.Devices.GetByID(ctx, tc.CreateDevice.ApplicationIdentifiers, tc.CreateDevice.DeviceID, ttnpb.EndDeviceFieldPathsTopLevel)
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
		SetByIDFunc    func(context.Context, ttnpb.ApplicationIdentifiers, string, []string, func(context.Context, *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, context.Context, error)
		Request        *ttnpb.EndDeviceIdentifiers
		ErrorAssertion func(*testing.T, error) bool
		SetByIDCalls   uint64
	}{
		{
			Name: "No device write rights",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, rights.Rights{
					ApplicationRights: map[string]*ttnpb.Rights{
						unique.ID(test.Context(), ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"}): {
							Rights: []ttnpb.Right{
								ttnpb.RIGHT_GATEWAY_SETTINGS_BASIC,
							},
						},
					},
				})
			},
			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string, f func(context.Context, *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, context.Context, error) {
				err := errors.New("SetByIDFunc must not be called")
				test.MustTFromContext(ctx).Error(err)
				return nil, ctx, err
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
						unique.ID(test.Context(), ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"}): {
							Rights: []ttnpb.Right{
								ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
							},
						},
					},
				})
			},
			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string, f func(context.Context, *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, context.Context, error) {
				t := test.MustTFromContext(ctx)
				a := assertions.New(t)
				a.So(appID, should.Resemble, ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"})
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
						unique.ID(test.Context(), ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"}): {
							Rights: []ttnpb.Right{
								ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
							},
						},
					},
				})
			},
			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string, f func(context.Context, *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, context.Context, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(appID, should.Resemble, ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"})
				a.So(devID, should.Equal, "test-dev-id")
				a.So(gets, should.BeNil)

				dev, sets, err := f(ctx, &ttnpb.EndDevice{
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						DeviceID:               "test-dev-id",
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
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
				DeviceID:               "test-dev-id",
				ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app-id"},
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
								SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string, f func(context.Context, *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, context.Context, error) {
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
