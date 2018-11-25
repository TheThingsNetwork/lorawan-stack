// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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
			JoinEUI: &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			DevEUI:  &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		},
	}

	ret, err := reg.GetByEUI(ctx, *pb.EndDeviceIdentifiers.JoinEUI, *pb.EndDeviceIdentifiers.DevEUI, pb.FieldMaskPaths())
	if !a.So(err, should.NotBeNil) || !a.So(errors.IsNotFound(err), should.BeTrue) {
		t.Fatalf("Error received: %v", err)
	}
	a.So(ret, should.BeNil)

	start := time.Now()

	ret, err = CreateDevice(ctx, reg, pb)
	if !a.So(err, should.BeNil) || !a.So(ret, should.NotBeNil) {
		t.FailNow()
	}
	a.So(ret.CreatedAt, should.HappenAfter, start)
	a.So(ret.UpdatedAt, should.HappenAfter, start)
	a.So(ret.UpdatedAt, should.Equal, ret.CreatedAt)
	pb.CreatedAt = ret.CreatedAt
	pb.UpdatedAt = ret.UpdatedAt
	a.So(ret, should.HaveEmptyDiff, pb)

	ret, err = reg.GetByEUI(ctx, *pb.EndDeviceIdentifiers.JoinEUI, *pb.EndDeviceIdentifiers.DevEUI, pb.FieldMaskPaths())
	a.So(err, should.BeNil)
	a.So(ret, should.HaveEmptyDiff, pb)

	pbOther := CopyEndDevice(pb)
	pbOther.EndDeviceIdentifiers.DevEUI = &types.EUI64{0x43, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

	ret, err = reg.GetByEUI(ctx, *pbOther.EndDeviceIdentifiers.JoinEUI, *pbOther.EndDeviceIdentifiers.DevEUI, pb.FieldMaskPaths())
	if !a.So(err, should.NotBeNil) || !a.So(errors.IsNotFound(err), should.BeTrue) {
		t.Fatalf("Error received: %v", err)
	}
	a.So(ret, should.BeNil)

	ret, err = CreateDevice(ctx, reg, pbOther)
	if !a.So(err, should.BeNil) || !a.So(ret, should.NotBeNil) {
		t.FailNow()
	}
	a.So(ret.CreatedAt, should.HappenAfter, pb.CreatedAt)
	a.So(ret.UpdatedAt, should.HappenAfter, pb.UpdatedAt)
	a.So(ret.UpdatedAt, should.Equal, ret.CreatedAt)
	pbOther.CreatedAt = ret.CreatedAt
	pbOther.UpdatedAt = ret.UpdatedAt
	a.So(ret, should.HaveEmptyDiff, pbOther)

	ret, err = reg.GetByEUI(ctx, *pbOther.EndDeviceIdentifiers.JoinEUI, *pbOther.EndDeviceIdentifiers.DevEUI, pb.FieldMaskPaths())
	a.So(err, should.BeNil)
	a.So(ret, should.HaveEmptyDiff, pbOther)

	err = DeleteDevice(ctx, reg, *pb.EndDeviceIdentifiers.JoinEUI, *pb.EndDeviceIdentifiers.DevEUI)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	ret, err = reg.GetByEUI(ctx, *pb.EndDeviceIdentifiers.JoinEUI, *pb.EndDeviceIdentifiers.DevEUI, pb.FieldMaskPaths())
	if !a.So(err, should.NotBeNil) || !a.So(errors.IsNotFound(err), should.BeTrue) {
		t.Fatalf("Error received: %v", err)
	}
	a.So(ret, should.BeNil)

	err = DeleteDevice(ctx, reg, *pbOther.EndDeviceIdentifiers.JoinEUI, *pbOther.EndDeviceIdentifiers.DevEUI)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	ret, err = reg.GetByEUI(ctx, *pbOther.EndDeviceIdentifiers.JoinEUI, *pbOther.EndDeviceIdentifiers.DevEUI, pb.FieldMaskPaths())
	if !a.So(err, should.NotBeNil) || !a.So(errors.IsNotFound(err), should.BeTrue) {
		t.Fatalf("Error received: %v", err)
	}
	a.So(ret, should.BeNil)

	// TODO: Test field mask application once implemented (https://github.com/TheThingsIndustries/lorawan-stack/issues/1212)
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

	devEUI := types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	pb := ttnpb.NewPopulatedSessionKeys(test.Randy, false)
	pb.SessionKeyID = "test-keys"

	ret, err := reg.GetByID(ctx, devEUI, pb.SessionKeyID, pb.FieldMaskPaths())
	if !a.So(err, should.NotBeNil) || !a.So(errors.IsNotFound(err), should.BeTrue) {
		t.Fatalf("Error received: %v", err)
	}
	a.So(ret, should.BeNil)

	ret, err = CreateKeys(ctx, reg, devEUI, pb)
	if !a.So(err, should.BeNil) || !a.So(ret, should.NotBeNil) {
		t.FailNow()
	}
	a.So(ret, should.HaveEmptyDiff, pb)

	ret, err = reg.GetByID(ctx, devEUI, pb.SessionKeyID, pb.FieldMaskPaths())
	a.So(err, should.BeNil)
	a.So(ret, should.HaveEmptyDiff, pb)

	pbOther := CopySessionKeys(pb)
	devEUIOther := types.EUI64{0x43, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

	ret, err = reg.GetByID(ctx, devEUIOther, pbOther.SessionKeyID, pbOther.FieldMaskPaths())
	if !a.So(err, should.NotBeNil) || !a.So(errors.IsNotFound(err), should.BeTrue) {
		t.Fatalf("Error received: %v", err)
	}
	a.So(ret, should.BeNil)

	ret, err = CreateKeys(ctx, reg, devEUIOther, pbOther)
	if !a.So(err, should.BeNil) || !a.So(ret, should.NotBeNil) {
		t.FailNow()
	}
	a.So(ret, should.HaveEmptyDiff, pbOther)

	ret, err = reg.GetByID(ctx, devEUIOther, pbOther.SessionKeyID, pb.FieldMaskPaths())
	a.So(err, should.BeNil)
	a.So(ret, should.HaveEmptyDiff, pbOther)

	err = DeleteKeys(ctx, reg, devEUI, pb.SessionKeyID)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	ret, err = reg.GetByID(ctx, devEUI, pb.SessionKeyID, pb.FieldMaskPaths())
	if !a.So(err, should.NotBeNil) || !a.So(errors.IsNotFound(err), should.BeTrue) {
		t.Fatalf("Error received: %v", err)
	}
	a.So(ret, should.BeNil)

	err = DeleteKeys(ctx, reg, devEUIOther, pbOther.SessionKeyID)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	ret, err = reg.GetByID(ctx, devEUIOther, pbOther.SessionKeyID, pbOther.FieldMaskPaths())
	if !a.So(err, should.NotBeNil) || !a.So(errors.IsNotFound(err), should.BeTrue) {
		t.Fatalf("Error received: %v", err)
	}
	a.So(ret, should.BeNil)

	// TODO: Test field mask application once implemented (https://github.com/TheThingsIndustries/lorawan-stack/issues/1212)
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
