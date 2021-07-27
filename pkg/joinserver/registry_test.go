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
	"context"
	"fmt"
	"testing"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/mohae/deepcopy"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	. "go.thethings.network/lorawan-stack/v3/pkg/joinserver"
	"go.thethings.network/lorawan-stack/v3/pkg/joinserver/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func CopyEndDevice(pb *ttnpb.EndDevice) *ttnpb.EndDevice {
	return deepcopy.Copy(pb).(*ttnpb.EndDevice)
}

// handleDeviceRegistryTest runs a test suite on reg.
func handleDeviceRegistryTest(t *testing.T, reg DeviceRegistry) {
	a, ctx := test.New(t)

	pb := &ttnpb.EndDevice{
		EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
			JoinEui:                &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			DevEui:                 &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"},
			DeviceId:               "test-dev",
		},
		ProvisionerID: "mock",
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

	retCtx, err := reg.GetByEUI(ctx, *pb.EndDeviceIdentifiers.JoinEui, *pb.EndDeviceIdentifiers.DevEui, ttnpb.EndDeviceFieldPathsTopLevel)
	if !a.So(err, should.NotBeNil) || !a.So(errors.IsNotFound(err), should.BeTrue) {
		t.Fatalf("Error received: %v", err)
	}
	a.So(retCtx, should.BeNil)

	start := time.Now()

	ret, err := reg.SetByID(ctx, pb.ApplicationIdentifiers, pb.DeviceId,
		[]string{
			"provisioner_id",
			"provisioning_data",
		},
		func(stored *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
			if !a.So(stored, should.BeNil) {
				t.Fatal("Registry is not empty")
			}
			return CopyEndDevice(pb), []string{
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
	a.So(ret.CreatedAt, should.HappenAfter, start)
	a.So(ret.UpdatedAt, should.HappenAfter, start)
	a.So(ret.UpdatedAt, should.Equal, ret.CreatedAt)
	pb.CreatedAt = ret.CreatedAt
	pb.UpdatedAt = ret.UpdatedAt
	a.So(ret, should.HaveEmptyDiff, pb)

	retCtx, err = reg.GetByEUI(ctx, *pb.EndDeviceIdentifiers.JoinEui, *pb.EndDeviceIdentifiers.DevEui, ttnpb.EndDeviceFieldPathsTopLevel)
	if !a.So(err, should.BeNil) {
		t.Fatalf("Failed to get device: %s", err)
	}
	a.So(retCtx.EndDevice, should.HaveEmptyDiff, pb)

	pbOther := CopyEndDevice(pb)
	pbOther.DeviceId = "other-device"
	pbOther.DevEui = &types.EUI64{0x43, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

	retCtx, err = reg.GetByEUI(ctx, *pbOther.EndDeviceIdentifiers.JoinEui, *pbOther.EndDeviceIdentifiers.DevEui, ttnpb.EndDeviceFieldPathsTopLevel)
	if !a.So(err, should.NotBeNil) || !a.So(errors.IsNotFound(err), should.BeTrue) {
		t.Fatalf("Error received: %v", err)
	}
	a.So(retCtx, should.BeNil)

	_, err = reg.SetByID(ctx, pbOther.ApplicationIdentifiers, pbOther.DeviceId,
		[]string{
			"provisioner_id",
			"provisioning_data",
		},
		func(stored *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
			if !a.So(stored, should.BeNil) {
				t.Fatal("Registry is not empty")
			}
			return CopyEndDevice(pbOther), []string{
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

	err = DeleteDevice(ctx, reg, pb.ApplicationIdentifiers, pb.DeviceId)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	retCtx, err = reg.GetByEUI(ctx, *pb.EndDeviceIdentifiers.JoinEui, *pb.EndDeviceIdentifiers.DevEui, ttnpb.EndDeviceFieldPathsTopLevel)
	if !a.So(err, should.NotBeNil) || !a.So(errors.IsNotFound(err), should.BeTrue) {
		t.Fatalf("Error received: %v", err)
	}
	a.So(retCtx, should.BeNil)

	ret, err = reg.SetByID(ctx, pbOther.ApplicationIdentifiers, pbOther.DeviceId,
		[]string{
			"provisioner_id",
			"provisioning_data",
		},
		func(stored *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
			if !a.So(stored, should.BeNil) {
				t.Fatal("Registry is not empty")
			}
			return CopyEndDevice(pbOther), []string{
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

	a.So(ret.CreatedAt, should.HappenAfter, pb.CreatedAt)
	a.So(ret.UpdatedAt, should.HappenAfter, pb.UpdatedAt)
	a.So(ret.UpdatedAt, should.Equal, ret.CreatedAt)
	pbOther.CreatedAt = ret.CreatedAt
	pbOther.UpdatedAt = ret.UpdatedAt
	a.So(ret, should.HaveEmptyDiff, pbOther)

	retCtx, err = reg.GetByEUI(ctx, *pbOther.EndDeviceIdentifiers.JoinEui, *pbOther.EndDeviceIdentifiers.DevEui, ttnpb.EndDeviceFieldPathsTopLevel)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	a.So(retCtx.EndDevice, should.HaveEmptyDiff, pbOther)

	err = DeleteDevice(ctx, reg, pbOther.ApplicationIdentifiers, pbOther.DeviceId)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	retCtx, err = reg.GetByEUI(ctx, *pbOther.EndDeviceIdentifiers.JoinEui, *pbOther.EndDeviceIdentifiers.DevEui, ttnpb.EndDeviceFieldPathsTopLevel)
	if !a.So(err, should.NotBeNil) || !a.So(errors.IsNotFound(err), should.BeTrue) {
		t.Fatalf("Error received: %v", err)
	}
	a.So(retCtx, should.BeNil)
}

func CopySessionKeys(pb *ttnpb.SessionKeys) *ttnpb.SessionKeys {
	return deepcopy.Copy(pb).(*ttnpb.SessionKeys)
}

func TestDeviceRegistries(t *testing.T) {
	t.Parallel()

	namespace := [...]string{
		"joinserver_test",
	}

	for _, tc := range []struct {
		Name string
		New  func(ctx context.Context) (reg DeviceRegistry, closeFn func() error)
		N    uint16
	}{
		{
			Name: "Redis",
			New: func(ctx context.Context) (DeviceRegistry, func() error) {
				cl, flush := test.NewRedis(ctx, namespace[:]...)
				return &redis.DeviceRegistry{
						Redis: cl,
					}, func() error {
						flush()
						return cl.Close()
					}
			},
			N: 8,
		},
	} {
		for i := 0; i < int(tc.N); i++ {
			test.RunSubtest(t, test.SubtestConfig{
				Name:     fmt.Sprintf("%s/%d", tc.Name, i),
				Parallel: true,
				Func: func(ctx context.Context, t *testing.T, _ *assertions.Assertion) {
					reg, closeFn := tc.New(ctx)
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
		SessionKeyID: []byte{0x11, 0x22, 0x33, 0x44},
		FNwkSIntKey:  test.DefaultFNwkSIntKeyEnvelope,
		SNwkSIntKey:  test.DefaultSNwkSIntKeyEnvelope,
		NwkSEncKey:   test.DefaultNwkSEncKeyEnvelope,
		AppSKey:      test.DefaultAppSKeyEnvelope,
	}

	ret, err := reg.GetByID(ctx, joinEUI, devEUI, pb.SessionKeyID, ttnpb.SessionKeysFieldPathsTopLevel)
	if !a.So(err, should.NotBeNil) || !a.So(errors.IsNotFound(err), should.BeTrue) {
		t.Fatalf("Error received: %v", err)
	}
	a.So(ret, should.BeNil)

	ret, err = reg.SetByID(ctx, joinEUI, devEUI, pb.SessionKeyID,
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
			return CopySessionKeys(pb), []string{
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

	ret, err = reg.GetByID(ctx, joinEUI, devEUI, pb.SessionKeyID, ttnpb.SessionKeysFieldPathsTopLevel)
	a.So(err, should.BeNil)
	a.So(ret, should.HaveEmptyDiff, pb)

	pbOther := CopySessionKeys(pb)
	joinEUIOther := types.EUI64{0x43, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	devEUIOther := types.EUI64{0x43, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

	ret, err = reg.GetByID(ctx, joinEUIOther, devEUIOther, pbOther.SessionKeyID, ttnpb.SessionKeysFieldPathsTopLevel)
	if !a.So(err, should.NotBeNil) || !a.So(errors.IsNotFound(err), should.BeTrue) {
		t.Fatalf("Error received: %v", err)
	}
	a.So(ret, should.BeNil)

	ret, err = reg.SetByID(ctx, joinEUIOther, devEUIOther, pbOther.SessionKeyID,
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
			return CopySessionKeys(pbOther), []string{
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

	ret, err = reg.GetByID(ctx, joinEUIOther, devEUIOther, pbOther.SessionKeyID, ttnpb.SessionKeysFieldPathsTopLevel)
	a.So(err, should.BeNil)
	a.So(ret, should.HaveEmptyDiff, pbOther)

	err = DeleteKeys(ctx, reg, joinEUI, devEUI, pb.SessionKeyID)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	ret, err = reg.GetByID(ctx, joinEUI, devEUI, pb.SessionKeyID, ttnpb.SessionKeysFieldPathsTopLevel)
	if !a.So(err, should.NotBeNil) || !a.So(errors.IsNotFound(err), should.BeTrue) {
		t.Fatalf("Error received: %v", err)
	}
	a.So(ret, should.BeNil)

	err = DeleteKeys(ctx, reg, joinEUIOther, devEUIOther, pbOther.SessionKeyID)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	ret, err = reg.GetByID(ctx, joinEUIOther, devEUIOther, pbOther.SessionKeyID, ttnpb.SessionKeysFieldPathsTopLevel)
	if !a.So(err, should.NotBeNil) || !a.So(errors.IsNotFound(err), should.BeTrue) {
		t.Fatalf("Error received: %v", err)
	}
	a.So(ret, should.BeNil)
}

func TestSessionKeyRegistries(t *testing.T) {
	t.Parallel()

	namespace := [...]string{
		"joinserver_test",
	}

	for _, tc := range []struct {
		Name string
		New  func(ctx context.Context) (reg KeyRegistry, closeFn func() error)
		N    uint16
	}{
		{
			Name: "Redis",
			New: func(ctx context.Context) (KeyRegistry, func() error) {
				cl, flush := test.NewRedis(ctx, namespace[:]...)
				return &redis.KeyRegistry{
						Redis: cl,
					}, func() error {
						flush()
						return cl.Close()
					}
			},
			N: 8,
		},
	} {
		for i := 0; i < int(tc.N); i++ {
			test.RunSubtest(t, test.SubtestConfig{
				Name:     fmt.Sprintf("%s/%d", tc.Name, i),
				Parallel: true,
				Func: func(ctx context.Context, t *testing.T, _ *assertions.Assertion) {
					reg, closeFn := tc.New(ctx)
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
