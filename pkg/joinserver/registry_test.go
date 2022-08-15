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
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	. "go.thethings.network/lorawan-stack/v3/pkg/joinserver"
	"go.thethings.network/lorawan-stack/v3/pkg/joinserver/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

// handleDeviceRegistryTest runs a test suite on reg.
func handleDeviceRegistryTest(t *testing.T, reg DeviceRegistry) {
	a, ctx := test.New(t)

	joinEUI := types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	devEUI := types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

	pb := &ttnpb.EndDevice{
		Ids: &ttnpb.EndDeviceIdentifiers{
			JoinEui:        joinEUI.Bytes(),
			DevEui:         devEUI.Bytes(),
			ApplicationIds: &ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"},
			DeviceId:       "test-dev",
		},
		ProvisionerId: "mock",
		ProvisioningData: &pbtypes.Struct{
			Fields: map[string]*pbtypes.Value{
				"serial_number": {
					Kind: &pbtypes.Value_NumberValue{
						NumberValue: 42,
					},
				},
			},
		},
	}

	retCtx, err := reg.GetByEUI(ctx, joinEUI, devEUI, ttnpb.EndDeviceFieldPathsTopLevel)
	if !a.So(err, should.NotBeNil) || !a.So(errors.IsNotFound(err), should.BeTrue) {
		t.Fatalf("Error received: %v", err)
	}
	a.So(retCtx, should.BeNil)

	start := time.Now()

	ret, err := reg.SetByID(ctx, pb.Ids.ApplicationIds, pb.Ids.DeviceId,
		[]string{
			"provisioner_id",
			"provisioning_data",
		},
		func(stored *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
			if !a.So(stored, should.BeNil) {
				t.Fatal("Registry is not empty")
			}
			return ttnpb.Clone(pb), []string{
				"ids.application_ids",
				"ids.dev_eui",
				"ids.device_id",
				"ids.join_eui",
				"provisioner_id",
				"provisioning_data",
			}, nil
		},
	)
	if !a.So(err, should.BeNil) || !a.So(ret, should.NotBeNil) {
		t.Fatalf("Failed to create device: %s", err)
	}
	a.So(*ttnpb.StdTime(ret.CreatedAt), should.HappenAfter, start)
	a.So(*ttnpb.StdTime(ret.UpdatedAt), should.HappenAfter, start)
	a.So(ret.UpdatedAt, should.Equal, ret.CreatedAt)
	pb.CreatedAt = ret.CreatedAt
	pb.UpdatedAt = ret.UpdatedAt
	a.So(ret, should.HaveEmptyDiff, pb)

	retCtx, err = reg.GetByEUI(ctx, joinEUI, devEUI, ttnpb.EndDeviceFieldPathsTopLevel)
	if !a.So(err, should.BeNil) {
		t.Fatalf("Failed to get device: %s", err)
	}
	a.So(retCtx.EndDevice, should.HaveEmptyDiff, pb)

	pbOther := ttnpb.Clone(pb)
	pbOther.Ids.DeviceId = "other-device"
	pbOther.Ids.DevEui = types.EUI64{0x43, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}.Bytes()

	retCtx, err = reg.GetByEUI(ctx,
		types.MustEUI64(pbOther.Ids.JoinEui).OrZero(),
		types.MustEUI64(pbOther.Ids.DevEui).OrZero(),
		ttnpb.EndDeviceFieldPathsTopLevel)
	if !a.So(err, should.NotBeNil) || !a.So(errors.IsNotFound(err), should.BeTrue) {
		t.Fatalf("Error received: %v", err)
	}
	a.So(retCtx, should.BeNil)

	_, err = reg.SetByID(ctx, pbOther.Ids.ApplicationIds, pbOther.Ids.DeviceId,
		[]string{
			"provisioner_id",
			"provisioning_data",
		},
		func(stored *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
			if !a.So(stored, should.BeNil) {
				t.Fatal("Registry is not empty")
			}
			return ttnpb.Clone(pbOther), []string{
				"ids.application_ids",
				"ids.dev_eui",
				"ids.device_id",
				"ids.join_eui",
				"provisioner_id",
				"provisioning_data",
			}, nil
		},
	)
	if !a.So(errors.IsAlreadyExists(err), should.BeTrue) {
		t.Fatal("Device with conflicting provisioner unique ID created")
	}

	err = DeleteDevice(ctx, reg, pb.Ids.ApplicationIds, pb.Ids.DeviceId)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	retCtx, err = reg.GetByEUI(ctx, joinEUI, devEUI, ttnpb.EndDeviceFieldPathsTopLevel)
	if !a.So(err, should.NotBeNil) || !a.So(errors.IsNotFound(err), should.BeTrue) {
		t.Fatalf("Error received: %v", err)
	}
	a.So(retCtx, should.BeNil)

	ret, err = reg.SetByID(ctx, pbOther.Ids.ApplicationIds, pbOther.Ids.DeviceId,
		[]string{
			"provisioner_id",
			"provisioning_data",
		},
		func(stored *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
			if !a.So(stored, should.BeNil) {
				t.Fatal("Registry is not empty")
			}
			return ttnpb.Clone(pbOther), []string{
				"ids.application_ids",
				"ids.dev_eui",
				"ids.device_id",
				"ids.join_eui",
				"provisioner_id",
				"provisioning_data",
			}, nil
		},
	)
	if !a.So(err, should.BeNil) || !a.So(ret, should.NotBeNil) { // No more conflicts.
		t.Fatalf("Failed to create device: %s", err)
	}

	a.So(*ttnpb.StdTime(ret.CreatedAt), should.HappenAfter, *ttnpb.StdTime(pb.CreatedAt))
	a.So(*ttnpb.StdTime(ret.UpdatedAt), should.HappenAfter, *ttnpb.StdTime(pb.UpdatedAt))
	a.So(ret.UpdatedAt, should.Equal, ret.CreatedAt)
	pbOther.CreatedAt = ret.CreatedAt
	pbOther.UpdatedAt = ret.UpdatedAt
	a.So(ret, should.HaveEmptyDiff, pbOther)

	retCtx, err = reg.GetByEUI(
		ctx,
		types.MustEUI64(pbOther.Ids.JoinEui).OrZero(),
		types.MustEUI64(pbOther.Ids.DevEui).OrZero(),
		ttnpb.EndDeviceFieldPathsTopLevel,
	)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	a.So(retCtx.EndDevice, should.HaveEmptyDiff, pbOther)

	err = DeleteDevice(ctx, reg, pbOther.Ids.ApplicationIds, pbOther.Ids.DeviceId)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	retCtx, err = reg.GetByEUI(
		ctx,
		types.MustEUI64(pbOther.Ids.JoinEui).OrZero(),
		types.MustEUI64(pbOther.Ids.DevEui).OrZero(),
		ttnpb.EndDeviceFieldPathsTopLevel,
	)
	if !a.So(err, should.NotBeNil) || !a.So(errors.IsNotFound(err), should.BeTrue) {
		t.Fatalf("Error received: %v", err)
	}
	a.So(retCtx, should.BeNil)
}

func TestDeviceRegistries(t *testing.T) {
	t.Parallel()

	namespace := [...]string{
		"joinserver_test",
	}

	for _, tc := range []struct {
		Name string
		New  func(ctx context.Context) (DeviceRegistry, func() error, error)
		N    uint16
	}{
		{
			Name: "Redis",
			New: func(ctx context.Context) (DeviceRegistry, func() error, error) {
				cl, flush := test.NewRedis(ctx, namespace[:]...)
				devReg := &redis.DeviceRegistry{
					Redis:   cl,
					LockTTL: test.Delay << 10,
				}
				if err := devReg.Init(ctx); err != nil {
					return nil, nil, err
				}
				return devReg, func() error {
					flush()
					return cl.Close()
				}, nil
			},
			N: 8,
		},
	} {
		tc := tc
		for i := 0; i < int(tc.N); i++ {
			test.RunSubtest(t, test.SubtestConfig{
				Name:     fmt.Sprintf("%s/%d", tc.Name, i),
				Parallel: true,
				Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
					reg, closeFn, err := tc.New(ctx)
					if !a.So(err, should.BeNil) {
						t.FailNow()
					}
					if closeFn != nil {
						defer func() {
							if err := closeFn(); err != nil {
								t.Errorf("Failed to close registry: %s", err)
							}
						}()
					}
					t.Run("1st run", func(t *testing.T) { handleDeviceRegistryTest(t, reg) })
					if t.Failed() {
						t.Skip("Skipping 2nd run")
					}
					t.Run("2nd run", func(t *testing.T) { handleDeviceRegistryTest(t, reg) })
				},
			})
		}
	}
}

// handleKeyRegistryTest runs a test suite on reg.
func handleKeyRegistryTest(t *testing.T, reg KeyRegistry) {
	a := assertions.New(t)

	ctx := test.Context()

	joinEUI := types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	devEUI := types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	pb := &ttnpb.SessionKeys{
		SessionKeyId: []byte{0x11, 0x22, 0x33, 0x44},
		FNwkSIntKey:  test.DefaultFNwkSIntKeyEnvelope,
		SNwkSIntKey:  test.DefaultSNwkSIntKeyEnvelope,
		NwkSEncKey:   test.DefaultNwkSEncKeyEnvelope,
		AppSKey:      test.DefaultAppSKeyEnvelope,
	}

	ret, err := reg.GetByID(ctx, joinEUI, devEUI, pb.SessionKeyId, ttnpb.SessionKeysFieldPathsTopLevel)
	if !a.So(err, should.NotBeNil) || !a.So(errors.IsNotFound(err), should.BeTrue) {
		t.Fatalf("Error received: %v", err)
	}
	a.So(ret, should.BeNil)

	ret, err = reg.SetByID(ctx, joinEUI, devEUI, pb.SessionKeyId,
		[]string{
			"app_s_key",
			"f_nwk_s_int_key",
			"nwk_s_enc_key",
			"s_nwk_s_int_key",
		},
		func(stored *ttnpb.SessionKeys) (*ttnpb.SessionKeys, []string, error) {
			if !a.So(stored, should.BeNil) {
				t.Fatal("Registry is not empty")
			}
			return ttnpb.Clone(pb), []string{
				"app_s_key",
				"f_nwk_s_int_key",
				"nwk_s_enc_key",
				"s_nwk_s_int_key",
				"session_key_id",
			}, nil
		},
	)
	if !a.So(err, should.BeNil) || !a.So(ret, should.NotBeNil) {
		t.Fatalf("Failed to create keys: %s", err)
	}
	a.So(ret, should.Resemble, pb)

	ret, err = reg.GetByID(ctx, joinEUI, devEUI, pb.SessionKeyId, ttnpb.SessionKeysFieldPathsTopLevel)
	a.So(err, should.BeNil)
	a.So(ret, should.HaveEmptyDiff, pb)

	pbOther := ttnpb.Clone(pb)
	joinEUIOther := types.EUI64{0x43, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	devEUIOther := types.EUI64{0x43, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

	ret, err = reg.GetByID(ctx, joinEUIOther, devEUIOther, pbOther.SessionKeyId, ttnpb.SessionKeysFieldPathsTopLevel)
	if !a.So(err, should.NotBeNil) || !a.So(errors.IsNotFound(err), should.BeTrue) {
		t.Fatalf("Error received: %v", err)
	}
	a.So(ret, should.BeNil)

	ret, err = reg.SetByID(ctx, joinEUIOther, devEUIOther, pbOther.SessionKeyId,
		[]string{
			"app_s_key",
			"f_nwk_s_int_key",
			"nwk_s_enc_key",
			"s_nwk_s_int_key",
		},
		func(stored *ttnpb.SessionKeys) (*ttnpb.SessionKeys, []string, error) {
			if !a.So(stored, should.BeNil) {
				t.Fatal("Registry is not empty")
			}
			return ttnpb.Clone(pbOther), []string{
				"app_s_key",
				"f_nwk_s_int_key",
				"nwk_s_enc_key",
				"s_nwk_s_int_key",
				"session_key_id",
			}, nil
		},
	)
	if !a.So(err, should.BeNil) || !a.So(ret, should.NotBeNil) {
		t.Fatalf("Failed to create keys: %s", err)
	}
	a.So(ret, should.Resemble, pbOther)

	ret, err = reg.GetByID(ctx, joinEUIOther, devEUIOther, pbOther.SessionKeyId, ttnpb.SessionKeysFieldPathsTopLevel)
	a.So(err, should.BeNil)
	a.So(ret, should.HaveEmptyDiff, pbOther)

	err = DeleteKeys(ctx, reg, joinEUI, devEUI, pb.SessionKeyId)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	ret, err = reg.GetByID(ctx, joinEUI, devEUI, pb.SessionKeyId, ttnpb.SessionKeysFieldPathsTopLevel)
	if !a.So(err, should.NotBeNil) || !a.So(errors.IsNotFound(err), should.BeTrue) {
		t.Fatalf("Error received: %v", err)
	}
	a.So(ret, should.BeNil)

	err = DeleteKeys(ctx, reg, joinEUIOther, devEUIOther, pbOther.SessionKeyId)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	ret, err = reg.GetByID(ctx, joinEUIOther, devEUIOther, pbOther.SessionKeyId, ttnpb.SessionKeysFieldPathsTopLevel)
	if !a.So(err, should.NotBeNil) || !a.So(errors.IsNotFound(err), should.BeTrue) {
		t.Fatalf("Error received: %v", err)
	}
	a.So(ret, should.BeNil)

	// Check number of retained session keys. Only the last 10 should be kept.
	for i := byte(0); i < 20; i++ {
		sid := bytes.Repeat([]byte{i}, 4)
		_, err := reg.SetByID(ctx, joinEUI, devEUI, sid, []string{
			"app_s_key",
			"f_nwk_s_int_key",
			"nwk_s_enc_key",
			"s_nwk_s_int_key",
		},
			func(stored *ttnpb.SessionKeys) (*ttnpb.SessionKeys, []string, error) {
				if !a.So(stored, should.BeNil) {
					t.Fatal("Registry is not empty")
				}
				return &ttnpb.SessionKeys{
						SessionKeyId: sid,
						FNwkSIntKey:  test.DefaultFNwkSIntKeyEnvelope,
						SNwkSIntKey:  test.DefaultSNwkSIntKeyEnvelope,
						NwkSEncKey:   test.DefaultNwkSEncKeyEnvelope,
						AppSKey:      test.DefaultAppSKeyEnvelope,
					}, []string{
						"app_s_key",
						"f_nwk_s_int_key",
						"nwk_s_enc_key",
						"s_nwk_s_int_key",
						"session_key_id",
					}, nil
			})
		if !a.So(err, should.BeNil) {
			t.Fatalf("Error received: %v", err)
		}
	}
	for i := byte(0); i < 20; i++ {
		_, err := reg.GetByID(ctx, joinEUI, devEUI, bytes.Repeat([]byte{i}, 4), ttnpb.SessionKeysFieldPathsTopLevel)
		if i < 10 {
			if !a.So(err, should.NotBeNil) || !a.So(errors.IsNotFound(err), should.BeTrue) {
				t.Fatalf("Error received: %v", err)
			}
		} else {
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}
		}
	}

	// Delete all the session keys of the given device.
	err = reg.Delete(ctx, joinEUI, devEUI)
	if !a.So(err, should.BeNil) {
		t.Fatalf("Error received: %v", err)
	}
	for i := byte(0); i < 20; i++ {
		_, err := reg.GetByID(ctx, joinEUI, devEUI, bytes.Repeat([]byte{i}, 4), ttnpb.SessionKeysFieldPathsTopLevel)
		if !a.So(err, should.NotBeNil) || !a.So(errors.IsNotFound(err), should.BeTrue) {
			t.Fatalf("Error received: %v", err)
		}
	}
}

func TestSessionKeyRegistries(t *testing.T) {
	t.Parallel()

	namespace := [...]string{
		"joinserver_test",
	}

	for _, tc := range []struct {
		Name string
		New  func(ctx context.Context) (KeyRegistry, func() error, error)
		N    uint16
	}{
		{
			Name: "Redis",
			New: func(ctx context.Context) (KeyRegistry, func() error, error) {
				cl, flush := test.NewRedis(ctx, namespace[:]...)
				keyReg := &redis.KeyRegistry{
					Redis:   cl,
					LockTTL: test.Delay << 10,
					Limit:   10,
				}
				if err := keyReg.Init(ctx); err != nil {
					return nil, nil, err
				}
				return keyReg, func() error {
					flush()
					return cl.Close()
				}, nil
			},
			N: 8,
		},
	} {
		tc := tc
		for i := 0; i < int(tc.N); i++ {
			test.RunSubtest(t, test.SubtestConfig{
				Name:     fmt.Sprintf("%s/%d", tc.Name, i),
				Parallel: true,
				Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
					reg, closeFn, err := tc.New(ctx)
					if !a.So(err, should.BeNil) {
						t.FailNow()
					}
					if closeFn != nil {
						defer func() {
							if err := closeFn(); err != nil {
								t.Errorf("Failed to close registry: %s", err)
							}
						}()
					}
					t.Run("1st run", func(t *testing.T) { handleKeyRegistryTest(t, reg) })
					if t.Failed() {
						t.Skip("Skipping 2nd run")
					}
					t.Run("2nd run", func(t *testing.T) { handleKeyRegistryTest(t, reg) })
				},
			})
		}
	}
}
