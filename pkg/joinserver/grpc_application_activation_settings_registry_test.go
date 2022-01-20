// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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
	"testing"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/v3/pkg/component/test"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto/cryptoutil"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	. "go.thethings.network/lorawan-stack/v3/pkg/joinserver"
	"go.thethings.network/lorawan-stack/v3/pkg/joinserver/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
)

func NewRedisApplicationActivationSettingRegistry(ctx context.Context) (ApplicationActivationSettingRegistry, func()) {
	tb := test.MustTBFromContext(ctx)
	cl, flush := test.NewRedis(ctx, "application-activation-settings")
	reg := &redis.ApplicationActivationSettingRegistry{
		Redis:   cl,
		LockTTL: test.Delay << 10,
	}
	if err := reg.Init(ctx); !assertions.New(tb).So(err, should.BeNil) {
		tb.FailNow()
	}
	return reg,
		func() {
			flush()
			cl.Close()
		}
}

func TestApplicationActivationSettingRegistryServer(t *testing.T) {
	_, ctx := test.New(t)

	const (
		jsKEKLabel      = "js-kek-label"
		sessionKEKLabel = "session-kek-label"

		appIDStr = "test-app"
		asID     = "test-as-id"
	)
	appID := &ttnpb.ApplicationIdentifiers{
		ApplicationId: appIDStr,
	}
	netID := types.NetID{0x0, 0x1, 0x2}
	jsKEK := types.AES128Key{0x0, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc, 0xd, 0xe, 0xe}
	sessionKEK := types.AES128Key{0x0, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc, 0xd, 0xe, 0xd}

	keyVault := cryptoutil.NewMemKeyVault(map[string][]byte{
		jsKEKLabel:      jsKEK[:],
		sessionKEKLabel: sessionKEK[:],
	})
	jsKEKEnvelopeUnwrapped := &ttnpb.KeyEnvelope{
		Key: &jsKEK,
	}

	credOpt := grpc.PerRPCCredentials(rpcmetadata.MD{
		AuthType:      "Bearer",
		AuthValue:     "key",
		AllowInsecure: true,
	})
	newJS := func(ctx context.Context, rights ...ttnpb.Right) (ttnpb.ApplicationActivationSettingRegistryClient, ApplicationActivationSettingRegistry, func()) {
		reg, closeFn := NewRedisApplicationActivationSettingRegistry(ctx)

		js := test.Must(New(
			componenttest.NewComponent(t, &component.Config{},
				component.WithClusterNew(func(context.Context, *cluster.Config, ...cluster.Option) (cluster.Cluster, error) {
					return &test.MockCluster{
						JoinFunc: test.ClusterJoinNilFunc,
						GetPeerFunc: func(reqCtx context.Context, role ttnpb.ClusterRole, ids cluster.EntityIdentifiers) (cluster.Peer, error) {
							_, a := test.MustNewTFromContext(ctx)
							a.So(role, should.Equal, ttnpb.ClusterRole_ACCESS)
							return test.Must(test.NewGRPCServerPeer(ctx, &test.MockApplicationAccessServer{
								ListRightsFunc: func(ctx context.Context, ids *ttnpb.ApplicationIdentifiers) (*ttnpb.Rights, error) {
									a.So(ids, should.Resemble, appID)
									return &ttnpb.Rights{
										Rights: rights,
									}, nil
								},
							}, ttnpb.RegisterApplicationAccessServer)).(cluster.Peer), nil
						},
					}, nil
				}),
			),
			&Config{
				ApplicationActivationSettings: reg,
				DeviceKEKLabel:                jsKEKLabel,
			},
		)).(*JoinServer)
		js.KeyVault = keyVault
		componenttest.StartComponent(t, js.Component)
		return ttnpb.NewApplicationActivationSettingRegistryClient(js.LoopbackConn()), reg, func() {
			js.Close()
			closeFn()
		}
	}

	// Get errors
	for _, tc := range []struct {
		Name           string
		Request        *ttnpb.GetApplicationActivationSettingsRequest
		Rights         []ttnpb.Right
		ErrorAssertion func(*testing.T, error) bool
	}{
		{
			Name:    "Empty request",
			Request: &ttnpb.GetApplicationActivationSettingsRequest{},
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(errors.IsInvalidArgument(err), should.BeTrue)
			},
		},
		{
			Name: "No rights",
			Request: &ttnpb.GetApplicationActivationSettingsRequest{
				ApplicationIds: appID,
				FieldMask: &pbtypes.FieldMask{
					Paths: []string{
						"kek",
						"kek_label",
					},
				},
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(errors.IsPermissionDenied(err), should.BeTrue)
			},
		},
		{
			Name: "No read right",
			Request: &ttnpb.GetApplicationActivationSettingsRequest{
				ApplicationIds: appID,
				FieldMask: &pbtypes.FieldMask{
					Paths: []string{
						"kek",
						"kek_label",
					},
				},
			},
			Rights: []ttnpb.Right{
				ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE_KEYS,
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(errors.IsPermissionDenied(err), should.BeTrue)
			},
		},
		{
			Name: "Not found/no paths",
			Request: &ttnpb.GetApplicationActivationSettingsRequest{
				ApplicationIds: appID,
			},
			Rights: []ttnpb.Right{
				ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ_KEYS,
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(errors.IsNotFound(err), should.BeTrue)
			},
		},
		{
			Name: "Not found/with paths",
			Request: &ttnpb.GetApplicationActivationSettingsRequest{
				ApplicationIds: appID,
				FieldMask: &pbtypes.FieldMask{
					Paths: []string{
						"kek",
						"kek_label",
					},
				},
			},
			Rights: []ttnpb.Right{
				ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ_KEYS,
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(errors.IsNotFound(err), should.BeTrue)
			},
		},
	} {
		tc := tc
		test.RunSubtestFromContext(ctx, test.SubtestConfig{
			Name:     fmt.Sprintf("Get errors/%s", tc.Name),
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				cl, _, stop := newJS(ctx, tc.Rights...)
				defer stop()

				sets, err := cl.Get(ctx, tc.Request, credOpt)
				if a.So(err, should.NotBeNil) {
					if !a.So(tc.ErrorAssertion(t, err), should.BeTrue) {
						t.Errorf("Error assertion failed. Error: %s", test.FormatError(err))
					}
					a.So(sets, should.BeNil)
				}
			},
		})
	}

	// Set errors
	for _, tc := range []struct {
		Name           string
		Request        *ttnpb.SetApplicationActivationSettingsRequest
		Rights         []ttnpb.Right
		ErrorAssertion func(*testing.T, error) bool
	}{
		{
			Name:    "Empty request",
			Request: &ttnpb.SetApplicationActivationSettingsRequest{},
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(errors.IsInvalidArgument(err), should.BeTrue)
			},
		},
		{
			Name: "No rights",
			Request: &ttnpb.SetApplicationActivationSettingsRequest{
				ApplicationIds: appID,
				Settings: &ttnpb.ApplicationActivationSettings{
					KekLabel: sessionKEKLabel,
					Kek:      jsKEKEnvelopeUnwrapped,
				},
				FieldMask: &pbtypes.FieldMask{
					Paths: []string{
						"kek",
						"kek_label",
					},
				},
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(errors.IsPermissionDenied(err), should.BeTrue)
			},
		},
		{
			Name: "No write right",
			Request: &ttnpb.SetApplicationActivationSettingsRequest{
				ApplicationIds: appID,
				Settings: &ttnpb.ApplicationActivationSettings{
					KekLabel: sessionKEKLabel,
					Kek:      jsKEKEnvelopeUnwrapped,
				},
				FieldMask: &pbtypes.FieldMask{
					Paths: []string{
						"kek",
						"kek_label",
					},
				},
			},
			Rights: []ttnpb.Right{
				ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ_KEYS,
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(errors.IsPermissionDenied(err), should.BeTrue)
			},
		},
		{
			Name: "No read right",
			Request: &ttnpb.SetApplicationActivationSettingsRequest{
				ApplicationIds: appID,
				Settings: &ttnpb.ApplicationActivationSettings{
					KekLabel: sessionKEKLabel,
					Kek:      jsKEKEnvelopeUnwrapped,
				},
				FieldMask: &pbtypes.FieldMask{
					Paths: []string{
						"kek",
						"kek_label",
					},
				},
			},
			Rights: []ttnpb.Right{
				ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE_KEYS,
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(errors.IsPermissionDenied(err), should.BeTrue)
			},
		},
		{
			Name: "No paths",
			Request: &ttnpb.SetApplicationActivationSettingsRequest{
				ApplicationIds: appID,
				Settings: &ttnpb.ApplicationActivationSettings{
					KekLabel: sessionKEKLabel,
					Kek:      jsKEKEnvelopeUnwrapped,
				},
			},
			Rights: []ttnpb.Right{
				ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ_KEYS,
				ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE_KEYS,
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(errors.IsInvalidArgument(err), should.BeTrue)
			},
		},
		{
			Name: "Empty KEK",
			Request: &ttnpb.SetApplicationActivationSettingsRequest{
				ApplicationIds: appID,
				Settings: &ttnpb.ApplicationActivationSettings{
					KekLabel: sessionKEKLabel,
					Kek: &ttnpb.KeyEnvelope{
						Key: &types.AES128Key{},
					},
				},
				FieldMask: &pbtypes.FieldMask{
					Paths: []string{
						"kek_label",
						"kek",
					},
				},
			},
			Rights: []ttnpb.Right{
				ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ_KEYS,
				ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE_KEYS,
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(errors.IsInvalidArgument(err), should.BeTrue)
			},
		},
		{
			Name: "KEK with empty label",
			Request: &ttnpb.SetApplicationActivationSettingsRequest{
				ApplicationIds: appID,
				Settings: &ttnpb.ApplicationActivationSettings{
					Kek: jsKEKEnvelopeUnwrapped,
				},
				FieldMask: &pbtypes.FieldMask{
					Paths: []string{
						"kek_label",
						"kek",
					},
				},
			},
			Rights: []ttnpb.Right{
				ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ_KEYS,
				ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE_KEYS,
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(errors.IsInvalidArgument(err), should.BeTrue)
			},
		},
	} {
		tc := tc
		test.RunSubtestFromContext(ctx, test.SubtestConfig{
			Name:     fmt.Sprintf("Set errors/%s", tc.Name),
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				cl, _, stop := newJS(ctx, tc.Rights...)
				defer stop()

				sets, err := cl.Set(ctx, tc.Request, credOpt)
				if a.So(err, should.NotBeNil) {
					if !a.So(tc.ErrorAssertion(t, err), should.BeTrue) {
						t.Errorf("Error assertion failed. Error: %s", test.FormatError(err))
					}
					a.So(sets, should.BeNil)
				}
			},
		})
	}

	// Delete errors
	for _, tc := range []struct {
		Name           string
		Request        *ttnpb.DeleteApplicationActivationSettingsRequest
		Rights         []ttnpb.Right
		ErrorAssertion func(*testing.T, error) bool
	}{
		{
			Name:    "Empty request",
			Request: &ttnpb.DeleteApplicationActivationSettingsRequest{},
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(errors.IsInvalidArgument(err), should.BeTrue)
			},
		},
		{
			Name: "No rights",
			Request: &ttnpb.DeleteApplicationActivationSettingsRequest{
				ApplicationIds: appID,
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(errors.IsPermissionDenied(err), should.BeTrue)
			},
		},
		{
			Name: "No write right",
			Request: &ttnpb.DeleteApplicationActivationSettingsRequest{
				ApplicationIds: appID,
			},
			Rights: []ttnpb.Right{
				ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ_KEYS,
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(errors.IsPermissionDenied(err), should.BeTrue)
			},
		},
		{
			Name: "No read right",
			Request: &ttnpb.DeleteApplicationActivationSettingsRequest{
				ApplicationIds: appID,
			},
			Rights: []ttnpb.Right{
				ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE_KEYS,
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(errors.IsPermissionDenied(err), should.BeTrue)
			},
		},
		{
			Name: "Not found",
			Request: &ttnpb.DeleteApplicationActivationSettingsRequest{
				ApplicationIds: appID,
			},
			Rights: []ttnpb.Right{
				ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ_KEYS,
				ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE_KEYS,
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(errors.IsNotFound(err), should.BeTrue)
			},
		},
	} {
		tc := tc
		test.RunSubtestFromContext(ctx, test.SubtestConfig{
			Name:     fmt.Sprintf("Delete errors/%s", tc.Name),
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				cl, _, stop := newJS(ctx, tc.Rights...)
				defer stop()

				v, err := cl.Delete(ctx, tc.Request, credOpt)
				if a.So(err, should.NotBeNil) {
					if !a.So(tc.ErrorAssertion(t, err), should.BeTrue) {
						t.Errorf("Error assertion failed. Error: %s", test.FormatError(err))
					}
					a.So(v, should.BeNil)
				}
			},
		})
	}

	for _, tc := range []struct {
		Name           string
		CreateSettings *ttnpb.ApplicationActivationSettings
		CreatePaths    []string
		GetSettings    *ttnpb.ApplicationActivationSettings
	}{
		{
			Name: "KEK sent plaintext",
			CreateSettings: &ttnpb.ApplicationActivationSettings{
				KekLabel:            sessionKEKLabel,
				Kek:                 jsKEKEnvelopeUnwrapped,
				HomeNetId:           &netID,
				ApplicationServerId: asID,
			},
			CreatePaths: []string{
				"application_server_id",
				"kek",
				"kek_label",
				"home_net_id",
			},
			GetSettings: &ttnpb.ApplicationActivationSettings{
				KekLabel:            sessionKEKLabel,
				Kek:                 jsKEKEnvelopeUnwrapped,
				HomeNetId:           &netID,
				ApplicationServerId: asID,
			},
		},
	} {
		tc := tc
		test.RunSubtestFromContext(ctx, test.SubtestConfig{
			Name:     fmt.Sprintf("Flow/%s", tc.Name),
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				cl, reg, stop := newJS(ctx,
					ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ_KEYS,
					ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE_KEYS,
				)
				defer stop()

				if !test.RunSubtestFromContext(ctx, test.SubtestConfig{
					Name: "Create",
					Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
						sets, err := cl.Set(ctx, &ttnpb.SetApplicationActivationSettingsRequest{
							ApplicationIds: appID,
							Settings:       tc.CreateSettings,
							FieldMask: &pbtypes.FieldMask{
								Paths: tc.CreatePaths,
							},
						}, credOpt)
						if !a.So(err, should.BeNil) {
							t.Fatalf("Failed to create settings: %s", test.FormatError(err))
						}
						a.So(sets, should.Resemble, tc.CreateSettings)
					},
				}) {
					t.FailNow()
				}
				test.RunSubtestFromContext(ctx, test.SubtestConfig{
					Name: "Encrypted storage at rest",
					Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
						stored, err := reg.GetByID(ctx, appID, ttnpb.ApplicationActivationSettingsFieldPathsTopLevel)
						if !a.So(err, should.BeNil) {
							t.Fatalf("Failed to get settings from registry directly: %s", test.FormatError(err))
						}
						a.So(stored.GetKek().GetKey(), should.BeNil)
					},
				})
				test.RunSubtestFromContext(ctx, test.SubtestConfig{
					Name: "Get after creation",
					Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
						sets, err := cl.Get(ctx, &ttnpb.GetApplicationActivationSettingsRequest{
							ApplicationIds: appID,
							FieldMask: &pbtypes.FieldMask{
								Paths: tc.CreatePaths,
							},
						}, credOpt)
						if !a.So(err, should.BeNil) {
							t.Fatalf("Failed to get settings: %s", test.FormatError(err))
						}
						a.So(sets, should.Resemble, tc.GetSettings)
					},
				})
				test.RunSubtestFromContext(ctx, test.SubtestConfig{
					Name: "Update",
					Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
						sets, err := cl.Set(ctx, &ttnpb.SetApplicationActivationSettingsRequest{
							ApplicationIds: appID,
							Settings:       tc.CreateSettings,
							FieldMask: &pbtypes.FieldMask{
								Paths: tc.CreatePaths,
							},
						}, credOpt)
						if !a.So(err, should.BeNil) {
							t.Fatalf("Failed to update settings: %s", test.FormatError(err))
						}
						a.So(sets, should.Resemble, tc.CreateSettings)
					},
				})
				test.RunSubtestFromContext(ctx, test.SubtestConfig{
					Name: "Remove KEK",
					Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
						sets, err := cl.Set(ctx, &ttnpb.SetApplicationActivationSettingsRequest{
							ApplicationIds: appID,
							Settings:       &ttnpb.ApplicationActivationSettings{},
							FieldMask: &pbtypes.FieldMask{
								Paths: []string{
									"kek_label",
									"kek",
								},
							},
						}, credOpt)
						if !a.So(err, should.BeNil) {
							t.Fatalf("Failed to remove KEK: %s", test.FormatError(err))
						}
						a.So(sets, should.Resemble, &ttnpb.ApplicationActivationSettings{})
					},
				})
				test.RunSubtestFromContext(ctx, test.SubtestConfig{
					Name: "Delete",
					Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
						_, err := cl.Delete(ctx, &ttnpb.DeleteApplicationActivationSettingsRequest{
							ApplicationIds: appID,
						}, credOpt)
						if !a.So(err, should.BeNil) {
							t.Fatalf("Failed to delete settings: %s", test.FormatError(err))
						}
					},
				})
				test.RunSubtestFromContext(ctx, test.SubtestConfig{
					Name: "Get after deletion",
					Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
						sets, err := cl.Get(ctx, &ttnpb.GetApplicationActivationSettingsRequest{
							ApplicationIds: appID,
							FieldMask: &pbtypes.FieldMask{
								Paths: tc.CreatePaths,
							},
						}, credOpt)
						if !a.So(err, should.NotBeNil) {
							t.Fatalf("Successful get after deletion")
						}
						if !a.So(errors.IsNotFound(err), should.BeTrue) {
							t.Errorf("Expected 'Not found' error. Got: %s", test.FormatError(err))
						}
						a.So(sets, should.BeNil)
					},
				})
				test.RunSubtestFromContext(ctx, test.SubtestConfig{
					Name: "Artifacts after deletion",
					Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
						stored, err := reg.GetByID(ctx, appID, ttnpb.ApplicationActivationSettingsFieldPathsTopLevel)
						if !a.So(err, should.NotBeNil) {
							t.Fatalf("Successful direct registry get after deletion")
						}
						if !a.So(errors.IsNotFound(err), should.BeTrue) {
							t.Errorf("Expected 'Not found' error. Got: %s", test.FormatError(err))
						}
						a.So(stored, should.BeNil)
					},
				})
			},
		})
	}
}
