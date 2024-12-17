// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

// Package networkserver implements the LoRaWAN Network Server.
package networkserver_test

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver" // nolint: revive
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

var errNotFound = errors.DefineNotFound("not_found", "not found")

func TestMACSettingsProfileRegistryGet(t *testing.T) {
	t.Parallel()
	nilProfileAssertion := func(t *testing.T, profile *ttnpb.GetMACSettingsProfileResponse) bool {
		t.Helper()
		return assertions.New(t).So(profile, should.BeNil)
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

	registeredProfileIDs := &ttnpb.MACSettingsProfileIdentifiers{
		ApplicationIds: &ttnpb.ApplicationIdentifiers{ApplicationId: "test-app-id"},
		ProfileId:      "test-profile-id",
	}

	for _, tc := range []struct {
		Name             string
		ContextFunc      func(context.Context) context.Context
		GetFunc          func(context.Context, *ttnpb.MACSettingsProfileIdentifiers, []string) (*ttnpb.MACSettingsProfile, error) // nolint: lll
		ProfileRequest   *ttnpb.GetMACSettingsProfileRequest
		ProfileAssertion func(*testing.T, *ttnpb.GetMACSettingsProfileResponse) bool
		ErrorAssertion   func(*testing.T, error) bool
		GetCalls         uint64
	}{
		{
			Name: "Permission denied",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, &rights.Rights{
					ApplicationRights: *rights.NewMap(map[string]*ttnpb.Rights{
						unique.ID(test.Context(), &ttnpb.ApplicationIdentifiers{ApplicationId: "test-app-id"}): nil,
					}),
				})
			},
			GetFunc: func(
				ctx context.Context,
				_ *ttnpb.MACSettingsProfileIdentifiers,
				_ []string,
			) (*ttnpb.MACSettingsProfile, error) {
				err := errors.New("GetFunc must not be called")
				test.MustTFromContext(ctx).Error(err)
				return nil, err
			},
			ProfileRequest: &ttnpb.GetMACSettingsProfileRequest{
				MacSettingsProfileIds: registeredProfileIDs,
				FieldMask:             ttnpb.FieldMask("mac_settings"),
			},
			ProfileAssertion: nilProfileAssertion,
			ErrorAssertion:   permissionDeniedErrorAssertion,
			GetCalls:         0,
		},
		{
			Name: "Invalid application ID",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, &rights.Rights{
					ApplicationRights: *rights.NewMap(map[string]*ttnpb.Rights{
						unique.ID(test.Context(), &ttnpb.ApplicationIdentifiers{
							ApplicationId: "invalid-application",
						}): ttnpb.RightsFrom(
							ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ,
						),
					}),
				})
			},
			GetFunc: func(
				ctx context.Context,
				_ *ttnpb.MACSettingsProfileIdentifiers,
				_ []string,
			) (*ttnpb.MACSettingsProfile, error) {
				err := errors.New("GetFunc must not be called")
				test.MustTFromContext(ctx).Error(err)
				return nil, err
			},
			ProfileRequest: &ttnpb.GetMACSettingsProfileRequest{
				MacSettingsProfileIds: registeredProfileIDs,
				FieldMask:             ttnpb.FieldMask("mac_settings"),
			},
			ProfileAssertion: nilProfileAssertion,
			ErrorAssertion:   permissionDeniedErrorAssertion,
			GetCalls:         0,
		},
		{
			Name: "Not found",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, &rights.Rights{
					ApplicationRights: *rights.NewMap(map[string]*ttnpb.Rights{
						unique.ID(test.Context(), &ttnpb.ApplicationIdentifiers{
							ApplicationId: "test-app-id",
						}): ttnpb.RightsFrom(
							ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ,
						),
					}),
				})
			},
			GetFunc: func(
				ctx context.Context,
				ids *ttnpb.MACSettingsProfileIdentifiers,
				paths []string,
			) (*ttnpb.MACSettingsProfile, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(ids, should.Resemble, ids)
				a.So(paths, should.HaveSameElementsDeep, []string{
					"mac_settings",
				})
				return nil, errNotFound.New()
			},
			ProfileRequest: &ttnpb.GetMACSettingsProfileRequest{
				MacSettingsProfileIds: registeredProfileIDs,
				FieldMask:             ttnpb.FieldMask("mac_settings"),
			},
			ProfileAssertion: nilProfileAssertion,
			ErrorAssertion:   notFoundErrorAssertion,
			GetCalls:         1,
		},
		{
			Name: "Found",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, &rights.Rights{
					ApplicationRights: *rights.NewMap(map[string]*ttnpb.Rights{
						unique.ID(test.Context(), &ttnpb.ApplicationIdentifiers{
							ApplicationId: "test-app-id",
						}): ttnpb.RightsFrom(
							ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ,
						),
					}),
				})
			},
			GetFunc: func(
				ctx context.Context,
				ids *ttnpb.MACSettingsProfileIdentifiers,
				paths []string,
			) (*ttnpb.MACSettingsProfile, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(ids, should.Resemble, ids)
				a.So(paths, should.HaveSameElementsDeep, []string{
					"ids",
					"mac_settings",
				})
				return ttnpb.Clone(&ttnpb.MACSettingsProfile{
					Ids: ids,
					MacSettings: &ttnpb.MACSettings{
						ResetsFCnt: &ttnpb.BoolValue{Value: true},
					},
				}), nil
			},
			ProfileRequest: &ttnpb.GetMACSettingsProfileRequest{
				MacSettingsProfileIds: registeredProfileIDs,
				FieldMask:             ttnpb.FieldMask("ids", "mac_settings"),
			},
			ProfileAssertion: func(t *testing.T, profile *ttnpb.GetMACSettingsProfileResponse) bool {
				t.Helper()
				return assertions.New(t).So(profile.MacSettingsProfile, should.Resemble, &ttnpb.MACSettingsProfile{
					Ids: &ttnpb.MACSettingsProfileIdentifiers{
						ApplicationIds: &ttnpb.ApplicationIdentifiers{
							ApplicationId: "test-app-id",
						},
						ProfileId: "test-profile-id",
					},
					MacSettings: &ttnpb.MACSettings{
						ResetsFCnt: &ttnpb.BoolValue{Value: true},
					},
				})
			},
			ErrorAssertion: nilErrorAssertion,
			GetCalls:       1,
		},
	} {
		tc := tc
		test.RunSubtest(t, test.SubtestConfig{
			Name:     tc.Name,
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				t.Helper()
				var getCalls uint64

				ns, ctx, _, stop := StartTest(
					ctx,
					TestConfig{
						NetworkServer: Config{
							MACSettingsProfileRegistry: &MockMACSettingsProfileRegistry{
								GetFunc: func(
									ctx context.Context,
									ids *ttnpb.MACSettingsProfileIdentifiers,
									paths []string,
								) (*ttnpb.MACSettingsProfile, error) {
									atomic.AddUint64(&getCalls, 1)
									return tc.GetFunc(ctx, ids, paths)
								},
							},
						},
						TaskStarter: StartTaskExclude(
							DownlinkProcessTaskName,
							DownlinkDispatchTaskName,
						),
					},
				)
				defer stop()

				ns.AddContextFiller(tc.ContextFunc)
				ns.AddContextFiller(func(ctx context.Context) context.Context {
					return test.ContextWithTB(ctx, t)
				})

				req := ttnpb.Clone(tc.ProfileRequest)

				profile, err := ttnpb.NewNsMACSettingsProfileRegistryClient(ns.LoopbackConn()).Get(ctx, req)
				if a.So(tc.ErrorAssertion(t, err), should.BeTrue) {
					a.So(tc.ProfileAssertion(t, profile), should.BeTrue)
				}
				a.So(req, should.Resemble, tc.ProfileRequest)
				a.So(getCalls, should.Equal, tc.GetCalls)
			},
		})
	}
}

func TestMACSettingsProfileRegistryCreate(t *testing.T) {
	t.Parallel()
	nilProfileAssertion := func(t *testing.T, profile *ttnpb.CreateMACSettingsProfileResponse) bool {
		t.Helper()
		return assertions.New(t).So(profile, should.BeNil)
	}
	nilErrorAssertion := func(t *testing.T, err error) bool {
		t.Helper()
		return assertions.New(t).So(err, should.BeNil)
	}
	permissionDeniedErrorAssertion := func(t *testing.T, err error) bool {
		t.Helper()
		return assertions.New(t).So(errors.IsPermissionDenied(err), should.BeTrue)
	}
	alreadyExistsErrorAssertion := func(t *testing.T, err error) bool {
		t.Helper()
		return assertions.New(t).So(errors.IsAlreadyExists(err), should.BeTrue)
	}

	registeredProfile := &ttnpb.MACSettingsProfile{
		Ids: &ttnpb.MACSettingsProfileIdentifiers{
			ApplicationIds: &ttnpb.ApplicationIdentifiers{ApplicationId: "test-app-id"},
			ProfileId:      "test-profile-id",
		},
		MacSettings: &ttnpb.MACSettings{
			ResetsFCnt: &ttnpb.BoolValue{Value: true},
		},
	}

	for _, tc := range []struct {
		Name             string
		ContextFunc      func(context.Context) context.Context
		SetFunc          func(ctx context.Context, ids *ttnpb.MACSettingsProfileIdentifiers, paths []string, f func(context.Context, *ttnpb.MACSettingsProfile) (*ttnpb.MACSettingsProfile, []string, error)) (*ttnpb.MACSettingsProfile, error) // nolint: lll
		ProfileRequest   *ttnpb.CreateMACSettingsProfileRequest
		ProfileAssertion func(*testing.T, *ttnpb.CreateMACSettingsProfileResponse) bool
		ErrorAssertion   func(*testing.T, error) bool
		SetCalls         uint64
	}{
		{
			Name: "Permission denied",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, &rights.Rights{
					ApplicationRights: *rights.NewMap(map[string]*ttnpb.Rights{
						unique.ID(test.Context(), &ttnpb.ApplicationIdentifiers{
							ApplicationId: "test-app-id",
						}): nil,
					}),
				})
			},
			SetFunc: func(
				ctx context.Context,
				_ *ttnpb.MACSettingsProfileIdentifiers,
				_ []string,
				_ func(context.Context, *ttnpb.MACSettingsProfile) (*ttnpb.MACSettingsProfile, []string, error),
			) (*ttnpb.MACSettingsProfile, error) {
				err := errors.New("SetFunc must not be called")
				test.MustTFromContext(ctx).Error(err)
				return nil, err
			},
			ProfileRequest: &ttnpb.CreateMACSettingsProfileRequest{
				MacSettingsProfile: registeredProfile,
			},
			ProfileAssertion: nilProfileAssertion,
			ErrorAssertion:   permissionDeniedErrorAssertion,
			SetCalls:         0,
		},
		{
			Name: "Invalid application ID",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, &rights.Rights{
					ApplicationRights: *rights.NewMap(map[string]*ttnpb.Rights{
						unique.ID(test.Context(), &ttnpb.ApplicationIdentifiers{
							ApplicationId: "invalid-application",
						}): ttnpb.RightsFrom(
							ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE,
						),
					}),
				})
			},
			SetFunc: func(
				ctx context.Context,
				_ *ttnpb.MACSettingsProfileIdentifiers,
				_ []string,
				_ func(context.Context, *ttnpb.MACSettingsProfile) (*ttnpb.MACSettingsProfile, []string, error),
			) (*ttnpb.MACSettingsProfile, error) {
				err := errors.New("SetFunc must not be called")
				test.MustTFromContext(ctx).Error(err)
				return nil, err
			},
			ProfileRequest: &ttnpb.CreateMACSettingsProfileRequest{
				MacSettingsProfile: registeredProfile,
			},
			ProfileAssertion: nilProfileAssertion,
			ErrorAssertion:   permissionDeniedErrorAssertion,
			SetCalls:         0,
		},
		{
			Name: "Create",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, &rights.Rights{
					ApplicationRights: *rights.NewMap(map[string]*ttnpb.Rights{
						unique.ID(test.Context(), &ttnpb.ApplicationIdentifiers{
							ApplicationId: "test-app-id",
						}): ttnpb.RightsFrom(
							ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE,
						),
					}),
				})
			},
			SetFunc: func(
				ctx context.Context,
				ids *ttnpb.MACSettingsProfileIdentifiers,
				paths []string,
				f func(context.Context, *ttnpb.MACSettingsProfile) (*ttnpb.MACSettingsProfile, []string, error),
			) (*ttnpb.MACSettingsProfile, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(ids, should.Resemble, ids)
				a.So(paths, should.HaveSameElementsDeep, []string{
					"ids",
					"mac_settings",
				})
				profile, sets, err := f(ctx, nil)
				a.So(sets, should.HaveSameElementsDeep, paths)
				a.So(profile, should.Resemble, registeredProfile)
				return profile, err
			},
			ProfileRequest: &ttnpb.CreateMACSettingsProfileRequest{
				MacSettingsProfile: registeredProfile,
			},
			ProfileAssertion: func(t *testing.T, profile *ttnpb.CreateMACSettingsProfileResponse) bool {
				t.Helper()
				return assertions.New(t).So(profile.MacSettingsProfile, should.Resemble, registeredProfile)
			},
			ErrorAssertion: nilErrorAssertion,
			SetCalls:       1,
		},
		{
			Name: "Create already exists",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, &rights.Rights{
					ApplicationRights: *rights.NewMap(map[string]*ttnpb.Rights{
						unique.ID(test.Context(), &ttnpb.ApplicationIdentifiers{
							ApplicationId: "test-app-id",
						}): ttnpb.RightsFrom(
							ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE,
						),
					}),
				})
			},
			SetFunc: func(
				ctx context.Context,
				ids *ttnpb.MACSettingsProfileIdentifiers,
				paths []string,
				f func(context.Context, *ttnpb.MACSettingsProfile) (*ttnpb.MACSettingsProfile, []string, error),
			) (*ttnpb.MACSettingsProfile, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(ids, should.Resemble, ids)
				a.So(paths, should.HaveSameElementsDeep, []string{
					"ids",
					"mac_settings",
				})
				profile, sets, err := f(ctx, ttnpb.Clone(registeredProfile))
				a.So(sets, should.HaveSameElementsDeep, []string{})
				a.So(profile, should.BeNil)
				return profile, err
			},
			ProfileRequest: &ttnpb.CreateMACSettingsProfileRequest{
				MacSettingsProfile: registeredProfile,
			},
			ProfileAssertion: nilProfileAssertion,
			ErrorAssertion:   alreadyExistsErrorAssertion,
			SetCalls:         1,
		},
	} {
		tc := tc
		test.RunSubtest(t, test.SubtestConfig{
			Name:     tc.Name,
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				t.Helper()
				var setCalls uint64

				ns, ctx, _, stop := StartTest(
					ctx,
					TestConfig{
						NetworkServer: Config{
							MACSettingsProfileRegistry: &MockMACSettingsProfileRegistry{
								SetFunc: func(
									ctx context.Context,
									ids *ttnpb.MACSettingsProfileIdentifiers,
									paths []string,
									f func(context.Context, *ttnpb.MACSettingsProfile) (*ttnpb.MACSettingsProfile, []string, error),
								) (*ttnpb.MACSettingsProfile, error) {
									atomic.AddUint64(&setCalls, 1)
									return tc.SetFunc(ctx, ids, paths, f)
								},
							},
						},
						TaskStarter: StartTaskExclude(
							DownlinkProcessTaskName,
							DownlinkDispatchTaskName,
						),
					},
				)
				defer stop()

				ns.AddContextFiller(tc.ContextFunc)
				ns.AddContextFiller(func(ctx context.Context) context.Context {
					return test.ContextWithTB(ctx, t)
				})

				req := ttnpb.Clone(tc.ProfileRequest)

				profile, err := ttnpb.NewNsMACSettingsProfileRegistryClient(ns.LoopbackConn()).Create(ctx, req)
				if a.So(tc.ErrorAssertion(t, err), should.BeTrue) {
					a.So(tc.ProfileAssertion(t, profile), should.BeTrue)
				}
				a.So(req, should.Resemble, tc.ProfileRequest)
				a.So(setCalls, should.Equal, tc.SetCalls)
			},
		})
	}
}

func TestMACSettingsProfileRegistryUpdate(t *testing.T) {
	t.Parallel()
	nilProfileAssertion := func(t *testing.T, profile *ttnpb.UpdateMACSettingsProfileResponse) bool {
		t.Helper()
		return assertions.New(t).So(profile, should.BeNil)
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

	registeredProfileIDs := &ttnpb.MACSettingsProfileIdentifiers{
		ApplicationIds: &ttnpb.ApplicationIdentifiers{ApplicationId: "test-app-id"},
		ProfileId:      "test-profile-id",
	}

	registeredProfile := &ttnpb.MACSettingsProfile{
		Ids: &ttnpb.MACSettingsProfileIdentifiers{
			ApplicationIds: &ttnpb.ApplicationIdentifiers{ApplicationId: "test-app-id"},
			ProfileId:      "test-profile-id",
		},
		MacSettings: &ttnpb.MACSettings{
			ResetsFCnt: &ttnpb.BoolValue{Value: true},
		},
	}

	for _, tc := range []struct {
		Name             string
		ContextFunc      func(context.Context) context.Context
		SetFunc          func(ctx context.Context, ids *ttnpb.MACSettingsProfileIdentifiers, paths []string, f func(context.Context, *ttnpb.MACSettingsProfile) (*ttnpb.MACSettingsProfile, []string, error)) (*ttnpb.MACSettingsProfile, error) // nolint: lll
		ProfileRequest   *ttnpb.UpdateMACSettingsProfileRequest
		ProfileAssertion func(*testing.T, *ttnpb.UpdateMACSettingsProfileResponse) bool
		ErrorAssertion   func(*testing.T, error) bool
		SetCalls         uint64
	}{
		{
			Name: "Permission denied",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, &rights.Rights{
					ApplicationRights: *rights.NewMap(map[string]*ttnpb.Rights{
						unique.ID(test.Context(), &ttnpb.ApplicationIdentifiers{
							ApplicationId: "test-app-id",
						}): nil,
					}),
				})
			},
			SetFunc: func(
				ctx context.Context,
				_ *ttnpb.MACSettingsProfileIdentifiers,
				_ []string,
				_ func(context.Context, *ttnpb.MACSettingsProfile) (*ttnpb.MACSettingsProfile, []string, error),
			) (*ttnpb.MACSettingsProfile, error) {
				err := errors.New("SetFunc must not be called")
				test.MustTFromContext(ctx).Error(err)
				return nil, err
			},
			ProfileRequest: &ttnpb.UpdateMACSettingsProfileRequest{
				MacSettingsProfileIds: registeredProfileIDs,
				MacSettingsProfile:    registeredProfile,
				FieldMask:             ttnpb.FieldMask("mac_settings"),
			},
			ProfileAssertion: nilProfileAssertion,
			ErrorAssertion:   permissionDeniedErrorAssertion,
			SetCalls:         0,
		},
		{
			Name: "Invalid application ID",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, &rights.Rights{
					ApplicationRights: *rights.NewMap(map[string]*ttnpb.Rights{
						unique.ID(test.Context(), &ttnpb.ApplicationIdentifiers{
							ApplicationId: "invalid-application",
						}): ttnpb.RightsFrom(
							ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE,
						),
					}),
				})
			},
			SetFunc: func(
				ctx context.Context,
				_ *ttnpb.MACSettingsProfileIdentifiers,
				_ []string,
				_ func(context.Context, *ttnpb.MACSettingsProfile) (*ttnpb.MACSettingsProfile, []string, error),
			) (*ttnpb.MACSettingsProfile, error) {
				err := errors.New("SetFunc must not be called")
				test.MustTFromContext(ctx).Error(err)
				return nil, err
			},
			ProfileRequest: &ttnpb.UpdateMACSettingsProfileRequest{
				MacSettingsProfileIds: registeredProfileIDs,
				MacSettingsProfile:    registeredProfile,
				FieldMask:             ttnpb.FieldMask("mac_settings"),
			},
			ProfileAssertion: nilProfileAssertion,
			ErrorAssertion:   permissionDeniedErrorAssertion,
			SetCalls:         0,
		},
		{
			Name: "Update",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, &rights.Rights{
					ApplicationRights: *rights.NewMap(map[string]*ttnpb.Rights{
						unique.ID(test.Context(), &ttnpb.ApplicationIdentifiers{
							ApplicationId: "test-app-id",
						}): ttnpb.RightsFrom(
							ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE,
						),
					}),
				})
			},
			SetFunc: func(
				ctx context.Context,
				ids *ttnpb.MACSettingsProfileIdentifiers,
				paths []string,
				f func(context.Context, *ttnpb.MACSettingsProfile) (*ttnpb.MACSettingsProfile, []string, error),
			) (*ttnpb.MACSettingsProfile, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(ids, should.Resemble, ids)
				a.So(paths, should.HaveSameElementsDeep, []string{
					"mac_settings",
				})
				profile, sets, err := f(ctx, ttnpb.Clone(registeredProfile))
				a.So(sets, should.HaveSameElementsDeep, paths)
				a.So(profile, should.Resemble, registeredProfile)
				return profile, err
			},
			ProfileRequest: &ttnpb.UpdateMACSettingsProfileRequest{
				MacSettingsProfileIds: registeredProfileIDs,
				MacSettingsProfile:    registeredProfile,
				FieldMask:             ttnpb.FieldMask("mac_settings"),
			},
			ProfileAssertion: func(t *testing.T, profile *ttnpb.UpdateMACSettingsProfileResponse) bool {
				t.Helper()
				return assertions.New(t).So(profile.MacSettingsProfile, should.Resemble, registeredProfile)
			},
			ErrorAssertion: nilErrorAssertion,
			SetCalls:       1,
		},
		{
			Name: "Update not found",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, &rights.Rights{
					ApplicationRights: *rights.NewMap(map[string]*ttnpb.Rights{
						unique.ID(test.Context(), &ttnpb.ApplicationIdentifiers{
							ApplicationId: "test-app-id",
						}): ttnpb.RightsFrom(
							ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE,
						),
					}),
				})
			},
			SetFunc: func(
				ctx context.Context,
				ids *ttnpb.MACSettingsProfileIdentifiers,
				paths []string,
				f func(context.Context, *ttnpb.MACSettingsProfile) (*ttnpb.MACSettingsProfile, []string, error),
			) (*ttnpb.MACSettingsProfile, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(ids, should.Resemble, ids)
				a.So(paths, should.HaveSameElementsDeep, []string{
					"mac_settings",
				})
				profile, sets, err := f(ctx, nil)
				a.So(sets, should.HaveSameElementsDeep, []string{})
				a.So(profile, should.BeNil)
				return profile, err
			},
			ProfileRequest: &ttnpb.UpdateMACSettingsProfileRequest{
				MacSettingsProfileIds: registeredProfileIDs,
				MacSettingsProfile:    registeredProfile,
				FieldMask:             ttnpb.FieldMask("mac_settings"),
			},
			ProfileAssertion: nilProfileAssertion,
			ErrorAssertion:   notFoundErrorAssertion,
			SetCalls:         1,
		},
	} {
		tc := tc
		test.RunSubtest(t, test.SubtestConfig{
			Name:     tc.Name,
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				t.Helper()
				var setCalls uint64

				ns, ctx, _, stop := StartTest(
					ctx,
					TestConfig{
						NetworkServer: Config{
							MACSettingsProfileRegistry: &MockMACSettingsProfileRegistry{
								SetFunc: func(
									ctx context.Context,
									ids *ttnpb.MACSettingsProfileIdentifiers,
									paths []string,
									f func(context.Context, *ttnpb.MACSettingsProfile) (*ttnpb.MACSettingsProfile, []string, error),
								) (*ttnpb.MACSettingsProfile, error) {
									atomic.AddUint64(&setCalls, 1)
									return tc.SetFunc(ctx, ids, paths, f)
								},
							},
						},
						TaskStarter: StartTaskExclude(
							DownlinkProcessTaskName,
							DownlinkDispatchTaskName,
						),
					},
				)
				defer stop()

				ns.AddContextFiller(tc.ContextFunc)
				ns.AddContextFiller(func(ctx context.Context) context.Context {
					return test.ContextWithTB(ctx, t)
				})

				req := ttnpb.Clone(tc.ProfileRequest)

				profile, err := ttnpb.NewNsMACSettingsProfileRegistryClient(ns.LoopbackConn()).Update(ctx, req)
				if a.So(tc.ErrorAssertion(t, err), should.BeTrue) {
					a.So(tc.ProfileAssertion(t, profile), should.BeTrue)
				}
				a.So(req, should.Resemble, tc.ProfileRequest)
				a.So(setCalls, should.Equal, tc.SetCalls)
			},
		})
	}
}

func TestMACSettingsProfileRegistryDelete(t *testing.T) {
	t.Parallel()
	nilProfileAssertion := func(t *testing.T, profile *ttnpb.DeleteMACSettingsProfileResponse) bool {
		t.Helper()
		return assertions.New(t).So(profile, should.BeNil)
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
	emptyProfileAssertion := func(t *testing.T, profile *ttnpb.DeleteMACSettingsProfileResponse) bool {
		t.Helper()
		return assertions.New(t).So(profile, should.Resemble, &ttnpb.DeleteMACSettingsProfileResponse{})
	}

	registeredProfileIDs := &ttnpb.MACSettingsProfileIdentifiers{
		ApplicationIds: &ttnpb.ApplicationIdentifiers{ApplicationId: "test-app-id"},
		ProfileId:      "test-profile-id",
	}

	registeredProfile := &ttnpb.MACSettingsProfile{
		Ids: &ttnpb.MACSettingsProfileIdentifiers{
			ApplicationIds: &ttnpb.ApplicationIdentifiers{ApplicationId: "test-app-id"},
			ProfileId:      "test-profile-id",
		},
		MacSettings: &ttnpb.MACSettings{
			ResetsFCnt: &ttnpb.BoolValue{Value: true},
		},
	}

	for _, tc := range []struct {
		Name             string
		ContextFunc      func(context.Context) context.Context
		SetFunc          func(ctx context.Context, ids *ttnpb.MACSettingsProfileIdentifiers, paths []string, f func(context.Context, *ttnpb.MACSettingsProfile) (*ttnpb.MACSettingsProfile, []string, error)) (*ttnpb.MACSettingsProfile, error) // nolint: lll
		ProfileRequest   *ttnpb.DeleteMACSettingsProfileRequest
		ProfileAssertion func(*testing.T, *ttnpb.DeleteMACSettingsProfileResponse) bool
		ErrorAssertion   func(*testing.T, error) bool
		SetCalls         uint64
	}{
		{
			Name: "Permission denied",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, &rights.Rights{
					ApplicationRights: *rights.NewMap(map[string]*ttnpb.Rights{
						unique.ID(test.Context(), &ttnpb.ApplicationIdentifiers{
							ApplicationId: "test-app-id",
						}): nil,
					}),
				})
			},
			SetFunc: func(
				ctx context.Context,
				_ *ttnpb.MACSettingsProfileIdentifiers,
				_ []string,
				_ func(context.Context, *ttnpb.MACSettingsProfile) (*ttnpb.MACSettingsProfile, []string, error),
			) (*ttnpb.MACSettingsProfile, error) {
				err := errors.New("SetFunc must not be called")
				test.MustTFromContext(ctx).Error(err)
				return nil, err
			},
			ProfileRequest: &ttnpb.DeleteMACSettingsProfileRequest{
				MacSettingsProfileIds: registeredProfileIDs,
			},
			ProfileAssertion: nilProfileAssertion,
			ErrorAssertion:   permissionDeniedErrorAssertion,
			SetCalls:         0,
		},
		{
			Name: "Invalid application ID",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, &rights.Rights{
					ApplicationRights: *rights.NewMap(map[string]*ttnpb.Rights{
						unique.ID(test.Context(), &ttnpb.ApplicationIdentifiers{
							ApplicationId: "invalid-application",
						}): ttnpb.RightsFrom(
							ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE,
						),
					}),
				})
			},
			SetFunc: func(
				ctx context.Context,
				_ *ttnpb.MACSettingsProfileIdentifiers,
				_ []string,
				_ func(context.Context, *ttnpb.MACSettingsProfile) (*ttnpb.MACSettingsProfile, []string, error),
			) (*ttnpb.MACSettingsProfile, error) {
				err := errors.New("SetFunc must not be called")
				test.MustTFromContext(ctx).Error(err)
				return nil, err
			},
			ProfileRequest: &ttnpb.DeleteMACSettingsProfileRequest{
				MacSettingsProfileIds: registeredProfileIDs,
			},
			ProfileAssertion: nilProfileAssertion,
			ErrorAssertion:   permissionDeniedErrorAssertion,
			SetCalls:         0,
		},
		{
			Name: "Delete",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, &rights.Rights{
					ApplicationRights: *rights.NewMap(map[string]*ttnpb.Rights{
						unique.ID(test.Context(), &ttnpb.ApplicationIdentifiers{
							ApplicationId: "test-app-id",
						}): ttnpb.RightsFrom(
							ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE,
						),
					}),
				})
			},
			SetFunc: func(
				ctx context.Context,
				ids *ttnpb.MACSettingsProfileIdentifiers,
				paths []string,
				f func(context.Context, *ttnpb.MACSettingsProfile) (*ttnpb.MACSettingsProfile, []string, error),
			) (*ttnpb.MACSettingsProfile, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(ids, should.Resemble, ids)
				a.So(paths, should.HaveSameElementsDeep, []string{
					"ids",
					"mac_settings",
				})
				profile, sets, err := f(ctx, ttnpb.Clone(registeredProfile))
				a.So(sets, should.BeNil)
				a.So(profile, should.BeNil)
				return profile, err
			},
			ProfileRequest: &ttnpb.DeleteMACSettingsProfileRequest{
				MacSettingsProfileIds: registeredProfileIDs,
			},
			ProfileAssertion: emptyProfileAssertion,
			ErrorAssertion:   nilErrorAssertion,
			SetCalls:         1,
		},
		{
			Name: "Delete not found",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, &rights.Rights{
					ApplicationRights: *rights.NewMap(map[string]*ttnpb.Rights{
						unique.ID(test.Context(), &ttnpb.ApplicationIdentifiers{
							ApplicationId: "test-app-id",
						}): ttnpb.RightsFrom(
							ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE,
						),
					}),
				})
			},
			SetFunc: func(
				ctx context.Context,
				ids *ttnpb.MACSettingsProfileIdentifiers,
				paths []string,
				f func(context.Context, *ttnpb.MACSettingsProfile) (*ttnpb.MACSettingsProfile, []string, error),
			) (*ttnpb.MACSettingsProfile, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(ids, should.Resemble, ids)
				a.So(paths, should.HaveSameElementsDeep, []string{
					"ids",
					"mac_settings",
				})
				profile, sets, err := f(ctx, nil)
				a.So(sets, should.HaveSameElementsDeep, []string{})
				a.So(profile, should.BeNil)
				return profile, err
			},
			ProfileRequest: &ttnpb.DeleteMACSettingsProfileRequest{
				MacSettingsProfileIds: registeredProfileIDs,
			},
			ProfileAssertion: nilProfileAssertion,
			ErrorAssertion:   notFoundErrorAssertion,
			SetCalls:         1,
		},
	} {
		tc := tc
		test.RunSubtest(t, test.SubtestConfig{
			Name:     tc.Name,
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				t.Helper()
				var setCalls uint64

				ns, ctx, _, stop := StartTest(
					ctx,
					TestConfig{
						NetworkServer: Config{
							MACSettingsProfileRegistry: &MockMACSettingsProfileRegistry{
								SetFunc: func(
									ctx context.Context,
									ids *ttnpb.MACSettingsProfileIdentifiers,
									paths []string,
									f func(context.Context, *ttnpb.MACSettingsProfile) (*ttnpb.MACSettingsProfile, []string, error),
								) (*ttnpb.MACSettingsProfile, error) {
									atomic.AddUint64(&setCalls, 1)
									return tc.SetFunc(ctx, ids, paths, f)
								},
							},
						},
						TaskStarter: StartTaskExclude(
							DownlinkProcessTaskName,
							DownlinkDispatchTaskName,
						),
					},
				)
				defer stop()

				ns.AddContextFiller(tc.ContextFunc)
				ns.AddContextFiller(func(ctx context.Context) context.Context {
					return test.ContextWithTB(ctx, t)
				})

				req := ttnpb.Clone(tc.ProfileRequest)

				profile, err := ttnpb.NewNsMACSettingsProfileRegistryClient(ns.LoopbackConn()).Delete(ctx, req)
				if a.So(tc.ErrorAssertion(t, err), should.BeTrue) {
					a.So(tc.ProfileAssertion(t, profile), should.BeTrue)
				}
				a.So(req, should.Resemble, tc.ProfileRequest)
				a.So(setCalls, should.Equal, tc.SetCalls)
			},
		})
	}
}

func TestMACSettingsProfileRegistryList(t *testing.T) {
	t.Parallel()
	nilProfileAssertion := func(t *testing.T, profile *ttnpb.ListMACSettingsProfilesResponse) bool {
		t.Helper()
		return assertions.New(t).So(profile, should.BeNil)
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

	registeredProfileIDs := &ttnpb.MACSettingsProfileIdentifiers{
		ApplicationIds: &ttnpb.ApplicationIdentifiers{ApplicationId: "test-app-id"},
		ProfileId:      "test-profile-id",
	}

	for _, tc := range []struct {
		Name             string
		ContextFunc      func(context.Context) context.Context
		ListFunc         func(context.Context, *ttnpb.ApplicationIdentifiers, []string) ([]*ttnpb.MACSettingsProfile, error) // nolint: lll
		PaginationFunc   func(context.Context, uint32, uint32, *int64) context.Context
		ProfileRequest   *ttnpb.ListMACSettingsProfilesRequest
		ProfileAssertion func(*testing.T, *ttnpb.ListMACSettingsProfilesResponse) bool
		ErrorAssertion   func(*testing.T, error) bool
		ListCalls        uint64
		PaginationCalls  uint64
	}{
		{
			Name: "Permission denied",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, &rights.Rights{
					ApplicationRights: *rights.NewMap(map[string]*ttnpb.Rights{
						unique.ID(test.Context(), &ttnpb.ApplicationIdentifiers{ApplicationId: "test-app-id"}): nil,
					}),
				})
			},
			ListFunc: func(
				ctx context.Context,
				_ *ttnpb.ApplicationIdentifiers,
				_ []string,
			) ([]*ttnpb.MACSettingsProfile, error) {
				err := errors.New("ListFunc must not be called")
				test.MustTFromContext(ctx).Error(err)
				return nil, err
			},
			PaginationFunc: func(
				ctx context.Context,
				_ uint32,
				_ uint32,
				_ *int64,
			) context.Context {
				err := errors.New("PaginationFunc must not be called")
				test.MustTFromContext(ctx).Error(err)
				return ctx
			},
			ProfileRequest: &ttnpb.ListMACSettingsProfilesRequest{
				ApplicationIds: registeredProfileIDs.ApplicationIds,
				FieldMask:      ttnpb.FieldMask("mac_settings"),
			},
			ProfileAssertion: nilProfileAssertion,
			ErrorAssertion:   permissionDeniedErrorAssertion,
			ListCalls:        0,
			PaginationCalls:  0,
		},
		{
			Name: "Invalid application ID",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, &rights.Rights{
					ApplicationRights: *rights.NewMap(map[string]*ttnpb.Rights{
						unique.ID(test.Context(), &ttnpb.ApplicationIdentifiers{
							ApplicationId: "invalid-application",
						}): ttnpb.RightsFrom(
							ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ,
						),
					}),
				})
			},
			ListFunc: func(
				ctx context.Context,
				_ *ttnpb.ApplicationIdentifiers,
				_ []string,
			) ([]*ttnpb.MACSettingsProfile, error) {
				err := errors.New("ListFunc must not be called")
				test.MustTFromContext(ctx).Error(err)
				return nil, err
			},
			PaginationFunc: func(
				ctx context.Context,
				_ uint32,
				_ uint32,
				_ *int64,
			) context.Context {
				err := errors.New("PaginationFunc must not be called")
				test.MustTFromContext(ctx).Error(err)
				return ctx
			},
			ProfileRequest: &ttnpb.ListMACSettingsProfilesRequest{
				ApplicationIds: registeredProfileIDs.ApplicationIds,
				FieldMask:      ttnpb.FieldMask("mac_settings"),
			},
			ProfileAssertion: nilProfileAssertion,
			ErrorAssertion:   permissionDeniedErrorAssertion,
			ListCalls:        0,
			PaginationCalls:  0,
		},
		{
			Name: "Not found",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, &rights.Rights{
					ApplicationRights: *rights.NewMap(map[string]*ttnpb.Rights{
						unique.ID(test.Context(), &ttnpb.ApplicationIdentifiers{
							ApplicationId: "test-app-id",
						}): ttnpb.RightsFrom(
							ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ,
						),
					}),
				})
			},
			ListFunc: func(
				ctx context.Context,
				ids *ttnpb.ApplicationIdentifiers,
				paths []string,
			) ([]*ttnpb.MACSettingsProfile, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(ids, should.Resemble, ids)
				a.So(paths, should.HaveSameElementsDeep, []string{
					"mac_settings",
				})
				return nil, errNotFound.New()
			},
			PaginationFunc: func(
				ctx context.Context,
				limit uint32,
				page uint32,
				total *int64,
			) context.Context {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(limit, should.Equal, 0)
				a.So(page, should.Equal, 0)
				a.So(total, should.NotBeNil)
				return ctx
			},
			ProfileRequest: &ttnpb.ListMACSettingsProfilesRequest{
				ApplicationIds: registeredProfileIDs.ApplicationIds,
				FieldMask:      ttnpb.FieldMask("mac_settings"),
			},
			ProfileAssertion: nilProfileAssertion,
			ErrorAssertion:   notFoundErrorAssertion,
			ListCalls:        1,
			PaginationCalls:  1,
		},
		{
			Name: "Found",
			ContextFunc: func(ctx context.Context) context.Context {
				return rights.NewContext(ctx, &rights.Rights{
					ApplicationRights: *rights.NewMap(map[string]*ttnpb.Rights{
						unique.ID(test.Context(), &ttnpb.ApplicationIdentifiers{
							ApplicationId: "test-app-id",
						}): ttnpb.RightsFrom(
							ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ,
						),
					}),
				})
			},
			ListFunc: func(
				ctx context.Context,
				ids *ttnpb.ApplicationIdentifiers,
				paths []string,
			) ([]*ttnpb.MACSettingsProfile, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(ids, should.Resemble, ids)
				a.So(paths, should.HaveSameElementsDeep, []string{
					"ids",
					"mac_settings",
				})
				return []*ttnpb.MACSettingsProfile{ttnpb.Clone(&ttnpb.MACSettingsProfile{
					Ids: registeredProfileIDs,
					MacSettings: &ttnpb.MACSettings{
						ResetsFCnt: &ttnpb.BoolValue{Value: true},
					},
				})}, nil
			},
			PaginationFunc: func(
				ctx context.Context,
				limit uint32,
				page uint32,
				total *int64,
			) context.Context {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(limit, should.Equal, 2)
				a.So(page, should.Equal, 1)
				a.So(total, should.NotBeNil)
				return ctx
			},
			ProfileRequest: &ttnpb.ListMACSettingsProfilesRequest{
				ApplicationIds: registeredProfileIDs.ApplicationIds,
				FieldMask:      ttnpb.FieldMask("ids", "mac_settings"),
				Limit:          2,
				Page:           1,
			},
			ProfileAssertion: func(t *testing.T, profile *ttnpb.ListMACSettingsProfilesResponse) bool {
				t.Helper()
				a := assertions.New(t)
				a.So(profile, should.NotBeNil)
				a.So(profile.MacSettingsProfiles, should.HaveLength, 1)
				return a.So(profile.MacSettingsProfiles, should.Resemble, []*ttnpb.MACSettingsProfile{{
					Ids: &ttnpb.MACSettingsProfileIdentifiers{
						ApplicationIds: &ttnpb.ApplicationIdentifiers{
							ApplicationId: "test-app-id",
						},
						ProfileId: "test-profile-id",
					},
					MacSettings: &ttnpb.MACSettings{
						ResetsFCnt: &ttnpb.BoolValue{Value: true},
					},
				}})
			},
			ErrorAssertion:  nilErrorAssertion,
			ListCalls:       1,
			PaginationCalls: 1,
		},
	} {
		tc := tc
		test.RunSubtest(t, test.SubtestConfig{
			Name:     tc.Name,
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				t.Helper()
				var listCalls, paginationCalls uint64

				ns, ctx, _, stop := StartTest(
					ctx,
					TestConfig{
						NetworkServer: Config{
							MACSettingsProfileRegistry: &MockMACSettingsProfileRegistry{
								ListFunc: func(
									ctx context.Context,
									ids *ttnpb.ApplicationIdentifiers,
									paths []string,
								) ([]*ttnpb.MACSettingsProfile, error) {
									atomic.AddUint64(&listCalls, 1)
									return tc.ListFunc(ctx, ids, paths)
								},
								WithPaginationFunc: func(
									ctx context.Context,
									limit uint32,
									page uint32,
									total *int64,
								) context.Context {
									atomic.AddUint64(&paginationCalls, 1)
									return tc.PaginationFunc(ctx, limit, page, total)
								},
							},
						},
						TaskStarter: StartTaskExclude(
							DownlinkProcessTaskName,
							DownlinkDispatchTaskName,
						),
					},
				)
				defer stop()

				ns.AddContextFiller(tc.ContextFunc)
				ns.AddContextFiller(func(ctx context.Context) context.Context {
					return test.ContextWithTB(ctx, t)
				})

				req := ttnpb.Clone(tc.ProfileRequest)

				profile, err := ttnpb.NewNsMACSettingsProfileRegistryClient(ns.LoopbackConn()).List(ctx, req)
				if a.So(tc.ErrorAssertion(t, err), should.BeTrue) {
					a.So(tc.ProfileAssertion(t, profile), should.BeTrue)
				}
				a.So(req, should.Resemble, tc.ProfileRequest)
				a.So(listCalls, should.Equal, tc.ListCalls)
				a.So(paginationCalls, should.Equal, tc.PaginationCalls)
			},
		})
	}
}
