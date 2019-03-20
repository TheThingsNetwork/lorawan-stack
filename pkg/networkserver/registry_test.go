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
	"fmt"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/errors"
	. "go.thethings.network/lorawan-stack/pkg/networkserver"
	"go.thethings.network/lorawan-stack/pkg/networkserver/redis"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

// handleRegistryTest runs a test suite on reg.
func handleRegistryTest(t *testing.T, reg DeviceRegistry) {
	a := assertions.New(t)

	ctx := test.Context()

	pb := &ttnpb.EndDevice{
		EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
			JoinEUI:                &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			DevEUI:                 &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app"},
			DeviceID:               "test-dev",
		},
		Session: &ttnpb.Session{
			DevAddr: types.DevAddr{0x42, 0xff, 0xff, 0xff},
		},
		PendingSession: &ttnpb.Session{
			DevAddr: types.DevAddr{0x43, 0xff, 0xff, 0xff},
		},
	}

	ret, err := reg.GetByID(ctx, pb.EndDeviceIdentifiers.ApplicationIdentifiers, pb.EndDeviceIdentifiers.DeviceID, ttnpb.EndDeviceFieldPathsTopLevel)
	if !a.So(err, should.NotBeNil) || !a.So(errors.IsNotFound(err), should.BeTrue) {
		t.Fatalf("Error received: %v", err)
	}
	a.So(ret, should.BeNil)

	ret, err = reg.GetByEUI(ctx, *pb.EndDeviceIdentifiers.JoinEUI, *pb.EndDeviceIdentifiers.DevEUI, ttnpb.EndDeviceFieldPathsTopLevel)
	if !a.So(err, should.NotBeNil) || !a.So(errors.IsNotFound(err), should.BeTrue) {
		t.Fatalf("Error received: %v", err)
	}
	a.So(ret, should.BeNil)

	var rets []*ttnpb.EndDevice
	rets = nil
	err = reg.RangeByAddr(ctx, pb.Session.DevAddr, ttnpb.EndDeviceFieldPathsTopLevel, func(dev *ttnpb.EndDevice) bool {
		rets = append(rets, dev)
		return true
	})
	a.So(err, should.BeNil)
	a.So(rets, should.BeNil)

	rets = nil
	err = reg.RangeByAddr(ctx, pb.PendingSession.DevAddr, ttnpb.EndDeviceFieldPathsTopLevel, func(dev *ttnpb.EndDevice) bool {
		rets = append(rets, dev)
		return true
	})
	a.So(err, should.BeNil)
	a.So(rets, should.BeNil)

	start := time.Now()

	ret, err = reg.SetByID(ctx, pb.ApplicationIdentifiers, pb.DeviceID,
		[]string{
			"ids.dev_eui",
			"ids.join_eui",
			"pending_session",
			"session",
		},
		func(stored *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
			if !a.So(stored, should.BeNil) {
				t.Fatal("Registry is not empty")
			}
			return pb, []string{
				"ids.dev_eui",
				"ids.join_eui",
				"pending_session",
				"session",
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

	ret, err = reg.GetByID(ctx, pb.EndDeviceIdentifiers.ApplicationIdentifiers, pb.EndDeviceIdentifiers.DeviceID, ttnpb.EndDeviceFieldPathsTopLevel)
	a.So(err, should.BeNil)
	a.So(ret, should.HaveEmptyDiff, pb)

	ret, err = reg.GetByEUI(ctx, *pb.EndDeviceIdentifiers.JoinEUI, *pb.EndDeviceIdentifiers.DevEUI, ttnpb.EndDeviceFieldPathsTopLevel)
	a.So(err, should.BeNil)
	a.So(ret, should.HaveEmptyDiff, pb)

	err = reg.RangeByAddr(ctx, pb.Session.DevAddr, ttnpb.EndDeviceFieldPathsTopLevel, func(dev *ttnpb.EndDevice) bool {
		rets = append(rets, dev)
		return true
	})
	a.So(err, should.BeNil)
	a.So(rets, should.HaveSameElementsDiff, []*ttnpb.EndDevice{pb})

	rets = nil
	err = reg.RangeByAddr(ctx, pb.PendingSession.DevAddr, ttnpb.EndDeviceFieldPathsTopLevel, func(dev *ttnpb.EndDevice) bool {
		rets = append(rets, dev)
		return true
	})
	a.So(err, should.BeNil)
	a.So(rets, should.HaveSameElementsDiff, []*ttnpb.EndDevice{pb})

	pbOther := CopyEndDevice(pb)
	pbOther.EndDeviceIdentifiers.DeviceID = "test-dev-other"
	pbOther.EndDeviceIdentifiers.DevEUI = &types.EUI64{0x43, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

	ret, err = reg.GetByID(ctx, pbOther.EndDeviceIdentifiers.ApplicationIdentifiers, pbOther.EndDeviceIdentifiers.DeviceID, ttnpb.EndDeviceFieldPathsTopLevel)
	if !a.So(err, should.NotBeNil) || !a.So(errors.IsNotFound(err), should.BeTrue) {
		t.Fatalf("Error received: %v", err)
	}
	a.So(ret, should.BeNil)

	ret, err = reg.GetByEUI(ctx, *pbOther.EndDeviceIdentifiers.JoinEUI, *pbOther.EndDeviceIdentifiers.DevEUI, ttnpb.EndDeviceFieldPathsTopLevel)
	if !a.So(err, should.NotBeNil) || !a.So(errors.IsNotFound(err), should.BeTrue) {
		t.Fatalf("Error received: %v", err)
	}
	a.So(ret, should.BeNil)

	ret, err = reg.SetByID(ctx, pbOther.ApplicationIdentifiers, pbOther.DeviceID,
		[]string{
			"ids.dev_eui",
			"ids.join_eui",
			"pending_session",
			"session",
		},
		func(stored *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
			if !a.So(stored, should.BeNil) {
				t.Fatal("Registry is not empty")
			}
			return pbOther, []string{
				"ids.dev_eui",
				"ids.join_eui",
				"pending_session",
				"session",
			}, nil
		},
	)
	if !a.So(err, should.BeNil) || !a.So(ret, should.NotBeNil) {
		t.Fatalf("Failed to create device: %s", err)
	}
	a.So(ret.CreatedAt, should.HappenAfter, pb.CreatedAt)
	a.So(ret.UpdatedAt, should.HappenAfter, pb.UpdatedAt)
	a.So(ret.UpdatedAt, should.Equal, ret.CreatedAt)
	pbOther.CreatedAt = ret.CreatedAt
	pbOther.UpdatedAt = ret.UpdatedAt
	a.So(ret, should.HaveEmptyDiff, pbOther)

	ret, err = reg.GetByID(ctx, pbOther.EndDeviceIdentifiers.ApplicationIdentifiers, pbOther.EndDeviceIdentifiers.DeviceID, ttnpb.EndDeviceFieldPathsTopLevel)
	a.So(err, should.BeNil)
	a.So(ret, should.Resemble, pbOther)

	ret, err = reg.GetByEUI(ctx, *pbOther.EndDeviceIdentifiers.JoinEUI, *pbOther.EndDeviceIdentifiers.DevEUI, ttnpb.EndDeviceFieldPathsTopLevel)
	a.So(err, should.BeNil)
	a.So(ret, should.Resemble, pbOther)

	rets = nil
	err = reg.RangeByAddr(ctx, pbOther.Session.DevAddr, ttnpb.EndDeviceFieldPathsTopLevel, func(dev *ttnpb.EndDevice) bool {
		rets = append(rets, dev)
		return true
	})
	a.So(err, should.BeNil)
	a.So(rets, should.HaveSameElementsDiff, []*ttnpb.EndDevice{pb, pbOther})

	rets = nil
	err = reg.RangeByAddr(ctx, pbOther.PendingSession.DevAddr, ttnpb.EndDeviceFieldPathsTopLevel, func(dev *ttnpb.EndDevice) bool {
		rets = append(rets, dev)
		return true
	})
	a.So(err, should.BeNil)
	a.So(rets, should.HaveSameElementsDiff, []*ttnpb.EndDevice{pb, pbOther})

	err = DeleteDevice(ctx, reg, pb.EndDeviceIdentifiers.ApplicationIdentifiers, pb.EndDeviceIdentifiers.DeviceID)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	ret, err = reg.GetByEUI(ctx, *pb.EndDeviceIdentifiers.JoinEUI, *pb.EndDeviceIdentifiers.DevEUI, ttnpb.EndDeviceFieldPathsTopLevel)
	if !a.So(err, should.NotBeNil) || !a.So(errors.IsNotFound(err), should.BeTrue) {
		t.Fatalf("Error received: %v", err)
	}
	a.So(ret, should.BeNil)

	ret, err = reg.GetByID(ctx, pb.EndDeviceIdentifiers.ApplicationIdentifiers, pb.EndDeviceIdentifiers.DeviceID, ttnpb.EndDeviceFieldPathsTopLevel)
	if !a.So(err, should.NotBeNil) || !a.So(errors.IsNotFound(err), should.BeTrue) {
		t.Fatalf("Error received: %v", err)
	}
	a.So(ret, should.BeNil)

	rets = nil
	err = reg.RangeByAddr(ctx, pb.Session.DevAddr, ttnpb.EndDeviceFieldPathsTopLevel, func(dev *ttnpb.EndDevice) bool {
		rets = append(rets, dev)
		return true
	})
	a.So(err, should.BeNil)
	a.So(rets, should.HaveSameElementsDiff, []*ttnpb.EndDevice{pbOther})

	rets = nil
	err = reg.RangeByAddr(ctx, pb.PendingSession.DevAddr, ttnpb.EndDeviceFieldPathsTopLevel, func(dev *ttnpb.EndDevice) bool {
		rets = append(rets, dev)
		return true
	})
	a.So(err, should.BeNil)
	a.So(rets, should.HaveSameElementsDiff, []*ttnpb.EndDevice{pbOther})

	err = DeleteDevice(ctx, reg, pbOther.EndDeviceIdentifiers.ApplicationIdentifiers, pbOther.EndDeviceIdentifiers.DeviceID)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	ret, err = reg.GetByID(ctx, pbOther.EndDeviceIdentifiers.ApplicationIdentifiers, pbOther.EndDeviceIdentifiers.DeviceID, ttnpb.EndDeviceFieldPathsTopLevel)
	if !a.So(err, should.NotBeNil) || !a.So(errors.IsNotFound(err), should.BeTrue) {
		t.Fatalf("Error received: %v", err)
	}
	a.So(ret, should.BeNil)

	ret, err = reg.GetByEUI(ctx, *pbOther.EndDeviceIdentifiers.JoinEUI, *pbOther.EndDeviceIdentifiers.DevEUI, ttnpb.EndDeviceFieldPathsTopLevel)
	if !a.So(err, should.NotBeNil) || !a.So(errors.IsNotFound(err), should.BeTrue) {
		t.Fatalf("Error received: %v", err)
	}
	a.So(ret, should.BeNil)

	rets = nil
	err = reg.RangeByAddr(ctx, pbOther.Session.DevAddr, ttnpb.EndDeviceFieldPathsTopLevel, func(dev *ttnpb.EndDevice) bool {
		rets = append(rets, dev)
		return true
	})
	a.So(err, should.BeNil)
	a.So(rets, should.BeNil)

	rets = nil
	err = reg.RangeByAddr(ctx, pbOther.PendingSession.DevAddr, ttnpb.EndDeviceFieldPathsTopLevel, func(dev *ttnpb.EndDevice) bool {
		rets = append(rets, dev)
		return true
	})
	a.So(err, should.BeNil)
	a.So(rets, should.BeNil)
}

func TestRegistries(t *testing.T) {
	t.Parallel()

	namespace := [...]string{
		"networkserver_test",
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
				t.Run("1st run", func(t *testing.T) { handleRegistryTest(t, reg) })
				if t.Failed() {
					t.Skip("Skipping 2nd run")
				}
				t.Run("2nd run", func(t *testing.T) { handleRegistryTest(t, reg) })
			})
		}
	}
}
