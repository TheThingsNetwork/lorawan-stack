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
	"fmt"
	"testing"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/mohae/deepcopy"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/errors"
	. "go.thethings.network/lorawan-stack/pkg/joinserver"
	"go.thethings.network/lorawan-stack/pkg/joinserver/redis"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func CopyEndDevice(pb *ttnpb.EndDevice) *ttnpb.EndDevice {
	return deepcopy.Copy(pb).(*ttnpb.EndDevice)
}

// handleDeviceRegistryTest runs a test suite on reg.
func handleDeviceRegistryTest(t *testing.T, reg DeviceRegistry) {
	a := assertions.New(t)

	ctx := test.Context()

	pb := &ttnpb.EndDevice{
		EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
			JoinEUI:                &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			DevEUI:                 &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app"},
			DeviceID:               "test-dev",
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

	retCtx, err := reg.GetByEUI(ctx, *pb.EndDeviceIdentifiers.JoinEUI, *pb.EndDeviceIdentifiers.DevEUI, ttnpb.EndDeviceFieldPathsTopLevel)
	if !a.So(err, should.NotBeNil) || !a.So(errors.IsNotFound(err), should.BeTrue) {
		t.Fatalf("Error received: %v", err)
	}
	a.So(retCtx, should.BeNil)

	start := time.Now()

	ret, err := reg.SetByID(ctx, pb.ApplicationIdentifiers, pb.DeviceID,
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

	retCtx, err = reg.GetByEUI(ctx, *pb.EndDeviceIdentifiers.JoinEUI, *pb.EndDeviceIdentifiers.DevEUI, ttnpb.EndDeviceFieldPathsTopLevel)
	if !a.So(err, should.BeNil) {
		t.Fatalf("Failed to get device: %s", err)
	}
	a.So(retCtx.EndDevice, should.HaveEmptyDiff, pb)

	pbOther := CopyEndDevice(pb)
	pbOther.DeviceID = "other-device"
	pbOther.DevEUI = &types.EUI64{0x43, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

	retCtx, err = reg.GetByEUI(ctx, *pbOther.EndDeviceIdentifiers.JoinEUI, *pbOther.EndDeviceIdentifiers.DevEUI, ttnpb.EndDeviceFieldPathsTopLevel)
	if !a.So(err, should.NotBeNil) || !a.So(errors.IsNotFound(err), should.BeTrue) {
		t.Fatalf("Error received: %v", err)
	}
	a.So(retCtx, should.BeNil)

	ret, err = reg.SetByID(ctx, pbOther.ApplicationIdentifiers, pbOther.DeviceID,
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

	err = DeleteDevice(ctx, reg, pb.ApplicationIdentifiers, pb.DeviceID)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	retCtx, err = reg.GetByEUI(ctx, *pb.EndDeviceIdentifiers.JoinEUI, *pb.EndDeviceIdentifiers.DevEUI, ttnpb.EndDeviceFieldPathsTopLevel)
	if !a.So(err, should.NotBeNil) || !a.So(errors.IsNotFound(err), should.BeTrue) {
		t.Fatalf("Error received: %v", err)
	}
	a.So(retCtx, should.BeNil)

	ret, err = reg.SetByID(ctx, pbOther.ApplicationIdentifiers, pbOther.DeviceID,
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

	retCtx, err = reg.GetByEUI(ctx, *pbOther.EndDeviceIdentifiers.JoinEUI, *pbOther.EndDeviceIdentifiers.DevEUI, ttnpb.EndDeviceFieldPathsTopLevel)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	a.So(retCtx.EndDevice, should.HaveEmptyDiff, pbOther)

	err = DeleteDevice(ctx, reg, pbOther.ApplicationIdentifiers, pbOther.DeviceID)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	retCtx, err = reg.GetByEUI(ctx, *pbOther.EndDeviceIdentifiers.JoinEUI, *pbOther.EndDeviceIdentifiers.DevEUI, ttnpb.EndDeviceFieldPathsTopLevel)
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
		New  func(t testing.TB) (reg DeviceRegistry, closeFn func() error)
		N    uint16
	}{
		{
			Name: "Redis",
			New: func(t testing.TB) (DeviceRegistry, func() error) {
				cl, flush := test.NewRedis(t, namespace[:]...)
				reg := &redis.DeviceRegistry{Redis: cl}
				return reg, func() error {
					flush()
					return cl.Close()
				}
			},
			N: 8,
		},
	} {
		for i := 0; i < int(tc.N); i++ {
			t.Run(fmt.Sprintf("%s/%d", tc.Name, i), func(t *testing.T) {
				t.Parallel()
				reg, closeFn := tc.New(t)
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
	pb := ttnpb.NewPopulatedSessionKeys(test.Randy, false)
	pb.SessionKeyID = []byte{0x11, 0x22, 0x33, 0x44}

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
		New  func(t testing.TB) (reg KeyRegistry, closeFn func() error)
		N    uint16
	}{
		{
			Name: "Redis",
			New: func(t testing.TB) (KeyRegistry, func() error) {
				cl, flush := test.NewRedis(t, namespace[:]...)
				reg := &redis.KeyRegistry{Redis: cl}
				return reg, func() error {
					flush()
					return cl.Close()
				}
			},
			N: 8,
		},
	} {
		for i := 0; i < int(tc.N); i++ {
			t.Run(fmt.Sprintf("%s/%d", tc.Name, i), func(t *testing.T) {
				t.Parallel()
				reg, closeFn := tc.New(t)
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
			})
		}
	}
}
