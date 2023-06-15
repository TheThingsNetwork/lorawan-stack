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

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	. "go.thethings.network/lorawan-stack/v3/pkg/joinserver"
	"go.thethings.network/lorawan-stack/v3/pkg/joinserver/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/protobuf/types/known/structpb"
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
		ProvisioningData: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"serial_number": {
					Kind: &structpb.Value_NumberValue{
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

	// Batch Operations
	pb1 := ttnpb.Clone(pb)
	pb1.Ids.DeviceId = "test-dev-1"
	pb1.Ids.DevEui = types.EUI64{0x42, 0x43, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}.Bytes()
	pb1.ProvisionerId = "mock-provisioner-1"

	pb2 := ttnpb.Clone(pb)
	pb2.Ids.DeviceId = "test-dev-2"
	pb2.Ids.DevEui = types.EUI64{0x42, 0x44, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}.Bytes()
	pb1.ProvisionerId = "mock-provisioner-2"

	pb3 := ttnpb.Clone(pb)
	pb3.Ids.DeviceId = "test-dev-3"
	pb3.Ids.DevEui = types.EUI64{0x42, 0x45, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}.Bytes()
	pb3.ProvisionerId = "mock-provisioner-3"

	// Create the devices
	for _, dev := range []*ttnpb.EndDevice{pb1, pb2, pb3} {
		devEUI := types.EUI64(dev.GetIds().DevEui)
		retCtx, err := reg.GetByEUI(ctx, joinEUI, devEUI, ttnpb.EndDeviceFieldPathsTopLevel)
		if !a.So(err, should.NotBeNil) || !a.So(errors.IsNotFound(err), should.BeTrue) {
			t.Fatalf("Error received: %v", err)
		}
		a.So(retCtx, should.BeNil)
		ret, err := reg.SetByID(ctx, dev.Ids.ApplicationIds, dev.Ids.DeviceId,
			[]string{
				"provisioner_id",
				"provisioning_data",
			},
			func(stored *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
				if !a.So(stored, should.BeNil) {
					t.Fatal("Registry is not empty")
				}
				return ttnpb.Clone(dev), []string{
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
		dev.CreatedAt = ret.CreatedAt
		dev.UpdatedAt = ret.UpdatedAt
		a.So(ret, should.HaveEmptyDiff, dev)

		retCtx, err = reg.GetByEUI(ctx, joinEUI, devEUI, ttnpb.EndDeviceFieldPathsTopLevel)
		if !a.So(err, should.BeNil) {
			t.Fatalf("Failed to get device: %s", err)
		}
		a.So(retCtx.EndDevice, should.HaveEmptyDiff, dev)
	}

	// Batch Delete
	deleted, err := reg.BatchDelete(
		ctx,
		pb.Ids.ApplicationIds,
		[]string{
			pb1.Ids.DeviceId,
			pb2.Ids.DeviceId,
			pb3.Ids.DeviceId,
			// This unknown device will be ignored.
			"test-dev-4",
		},
	)
	if !a.So(err, should.BeNil) {
		t.Fatalf("BatchDelete failed with: %s", errors.Stack(err))
	}
	if !a.So(deleted, should.HaveLength, 3) {
		t.Fatalf("BatchDelete returned wrong number of devices: %d", len(deleted))
	}
	if !a.So(deleted, should.Resemble, []*ttnpb.EndDeviceIdentifiers{pb1.Ids, pb2.Ids, pb3.Ids}) {
		t.Fatalf("Unexpected response from BatchDelete: %s", deleted)
	}

	// Check that the devices are deleted
	for _, pb := range []*ttnpb.EndDevice{pb1, pb2, pb3} {
		devEUI := types.EUI64(pb.GetIds().DevEui)
		retCtx, err := reg.GetByEUI(ctx, joinEUI, devEUI, ttnpb.EndDeviceFieldPathsTopLevel)
		if !a.So(err, should.NotBeNil) || !a.So(errors.IsNotFound(err), should.BeTrue) {
			t.Fatalf("Error received: %v", err)
		}
		a.So(retCtx, should.BeNil)
	}
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

	// Batch Operations
	noOfKeysPerDevice := uint8(10)
	devEUI1 := types.EUI64{0x44, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	devEUI2 := types.EUI64{0x45, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	devEUI3 := types.EUI64{0x46, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

	// Create keys for each device
	for _, devEUI := range []types.EUI64{devEUI1, devEUI2, devEUI3} {
		for i := byte(0); i < noOfKeysPerDevice; i++ {
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
			if err != nil {
				t.Fatalf("Error creating session key with ID %v for devEUI %v: %v", sid, devEUI, err)
			}
			// Read the keys back
			_, err = reg.GetByID(ctx, joinEUI, devEUI, sid, ttnpb.SessionKeysFieldPathsTopLevel)
			if err != nil {
				t.Fatalf("Error reading session key with ID %v for devEUI %v: %v", sid, devEUI, err)
			}
		}
	}

	// Batch Delete
	err = reg.BatchDelete(ctx, []*ttnpb.EndDeviceIdentifiers{
		{
			JoinEui: joinEUI.Bytes(),
			DevEui:  devEUI1.Bytes(),
		},
		{
			JoinEui: joinEUI.Bytes(),
			DevEui:  devEUI2.Bytes(),
		},
		{
			JoinEui: joinEUI.Bytes(),
			DevEui:  devEUI3.Bytes(),
		},
	})
	if !a.So(err, should.BeNil) {
		t.Fatalf("Could not BatchDelete keys: %v", err)
	}

	// Check if all keys are deleted
	for _, devEUI := range []types.EUI64{devEUI1, devEUI2, devEUI3} {
		for i := byte(0); i < noOfKeysPerDevice; i++ {
			sid := bytes.Repeat([]byte{i}, 4)
			_, err := reg.GetByID(ctx, joinEUI, devEUI, sid, ttnpb.SessionKeysFieldPathsTopLevel)
			if !a.So(err, should.NotBeNil) || !a.So(errors.IsNotFound(err), should.BeTrue) {
				t.Fatalf("Error received: %v", err)
			}
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
