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

package test

import (
	"context"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal/test"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func handleDeviceRegistryTest(ctx context.Context, reg DeviceRegistry) {
	t, a := test.MustNewTFromContext(ctx)

	type uplinkMatch struct {
		*ttnpb.EndDevice
		IsPending bool
	}
	assertUplinkMatches := func(ctx context.Context, up *ttnpb.UplinkMessage, expectedMatches ...uplinkMatch) bool {
		t := test.MustTFromContext(ctx)
		t.Helper()
		a := assertions.New(t)

		if len(expectedMatches) == 0 {
			if !a.So(errors.IsNotFound(reg.RangeByUplinkMatches(ctx, up, time.Second, func(storedCtx context.Context, match UplinkMatch) bool {
				a.So(storedCtx, should.HaveParentContextOrEqual, ctx)
				return false
			})), should.BeTrue) {
				t.Error("Devices matched, when no match was expected")
				return false
			}
			return true
		}

		var matches []UplinkMatch
		if !test.AllTrue(
			a.So(reg.RangeByUplinkMatches(ctx, up, time.Second, func(storedCtx context.Context, match UplinkMatch) bool {
				a.So(storedCtx, should.HaveParentContextOrEqual, ctx)
				matches = append(matches, match)
				return false
			}), should.BeNil),
			a.So(matches, should.HaveLength, len(expectedMatches)),
		) {
			t.Error("No matching device found")
			return false
		}
		for i, match := range matches {
			expectedMatch := expectedMatches[i]
			session, macState, msb := expectedMatch.Session, expectedMatch.MACState, expectedMatch.Session.LastFCntUp&0xffff0000
			if expectedMatch.IsPending {
				session, macState, msb = expectedMatch.PendingSession, expectedMatch.PendingMACState, 0
			}
			if !test.AllTrue(
				a.So(match.ApplicationIdentifiers(), should.Resemble, expectedMatch.ApplicationIdentifiers),
				a.So(match.DeviceID(), should.Equal, expectedMatch.DeviceID),
				a.So(match.LoRaWANVersion(), should.Equal, macState.LoRaWANVersion),
				a.So(match.FNwkSIntKey(), should.Resemble, session.FNwkSIntKey),
				a.So(match.FCnt(), should.Equal, msb|up.Payload.GetMACPayload().FCnt),
				a.So(match.LastFCnt(), should.Equal, session.LastFCntUp),
				a.So(match.IsPending(), should.Equal, expectedMatch.IsPending),
				a.So(match.ResetsFCnt(), should.Resemble, expectedMatch.GetMACSettings().GetResetsFCnt()),
			) {
				t.Error("Invalid devices matched")
				return false
			}
		}
		return true
	}

	pb := &ttnpb.EndDevice{
		EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
			JoinEUI:                &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			DevEUI:                 &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-app"},
			DeviceID:               "test-dev",
		},
		Session: &ttnpb.Session{
			DevAddr:    types.DevAddr{0x42, 0xff, 0xff, 0xff},
			LastFCntUp: 41,
		},
		MACState: MakeDefaultEU868MACState(ttnpb.CLASS_A, ttnpb.MAC_V1_0_3, ttnpb.PHY_V1_0_3_REV_A),
		PendingSession: &ttnpb.Session{
			DevAddr: types.DevAddr{0x43, 0xff, 0xff, 0xff},
		},
		PendingMACState: MakeDefaultEU868MACState(ttnpb.CLASS_A, ttnpb.MAC_V1_1, ttnpb.PHY_V1_1_REV_B),
	}
	pbFields := []string{
		"ids.application_ids",
		"ids.dev_eui",
		"ids.device_id",
		"ids.join_eui",
		"mac_state",
		"pending_mac_state",
		"pending_session",
		"session",
	}
	currentUp := &ttnpb.UplinkMessage{
		Payload: &ttnpb.Message{
			Payload: &ttnpb.Message_MACPayload{
				MACPayload: &ttnpb.MACPayload{
					FHDR: ttnpb.FHDR{
						FCnt:    42,
						DevAddr: pb.Session.DevAddr,
					},
				},
			},
		},
	}
	pendingUp := &ttnpb.UplinkMessage{
		Payload: &ttnpb.Message{
			Payload: &ttnpb.Message_MACPayload{
				MACPayload: &ttnpb.MACPayload{
					FHDR: ttnpb.FHDR{
						FCnt:    4242,
						DevAddr: pb.PendingSession.DevAddr,
					},
				},
			},
		},
	}

	stored, storedCtx, err := reg.GetByID(ctx, pb.EndDeviceIdentifiers.ApplicationIdentifiers, pb.EndDeviceIdentifiers.DeviceID, ttnpb.EndDeviceFieldPathsTopLevel)
	if !test.AllTrue(
		a.So(err, should.NotBeNil),
		a.So(errors.IsNotFound(err), should.BeTrue),
		a.So(storedCtx, should.HaveParentContextOrEqual, ctx),
		a.So(stored, should.BeNil),
	) {
		t.Fatal("GetByID assertion failed with empty registry")
	}

	stored, storedCtx, err = reg.GetByEUI(ctx, *pb.EndDeviceIdentifiers.JoinEUI, *pb.EndDeviceIdentifiers.DevEUI, ttnpb.EndDeviceFieldPathsTopLevel)
	if !test.AllTrue(
		a.So(err, should.NotBeNil),
		a.So(errors.IsNotFound(err), should.BeTrue),
		a.So(storedCtx, should.HaveParentContextOrEqual, ctx),
		a.So(stored, should.BeNil),
	) {
		t.Fatal("GetByEUI assertion failed with empty registry")
	}

	stored, storedCtx, err = reg.SetByID(ctx, pb.ApplicationIdentifiers, pb.DeviceID, pbFields,
		func(storedCtx context.Context, stored *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
			a.So(storedCtx, should.HaveParentContextOrEqual, ctx)
			if !a.So(stored, should.BeNil) {
				t.Fatal("Registry is not empty")
			}
			return nil, nil, nil
		},
	)
	if !test.AllTrue(
		a.So(err, should.BeNil),
		a.So(storedCtx, should.HaveParentContextOrEqual, ctx),
		a.So(stored, should.BeNil),
	) {
		t.Fatal("Read-only SetByID assertion failed with empty registry")
	}

	if !a.So(assertUplinkMatches(ctx, currentUp), should.BeTrue) {
		t.Fatal("RangeByUplinkMatches assertion failed for current session uplink with empty registry")
	}
	if !a.So(assertUplinkMatches(ctx, pendingUp), should.BeTrue) {
		t.Fatal("RangeByUplinkMatches assertion failed for pending session uplink with empty registry")
	}

	start := time.Now()

	stored, storedCtx, err = CreateDevice(storedCtx, reg, pb, pbFields...)
	pb.CreatedAt = stored.GetCreatedAt()
	pb.UpdatedAt = stored.GetUpdatedAt()
	if !test.AllTrue(
		a.So(err, should.BeNil) || a.So(errors.Stack(err), should.BeEmpty),
		a.So(storedCtx, should.HaveParentContextOrEqual, ctx),
		a.So(stored.GetCreatedAt(), should.HappenAfter, start),
		a.So(stored.GetUpdatedAt(), should.Equal, stored.GetCreatedAt()),
		a.So(stored, should.Resemble, pb),
	) {
		t.Fatal("Device creation assertion failed")
	}
	ctx = storedCtx

	stored, storedCtx, err = reg.GetByID(ctx, pb.EndDeviceIdentifiers.ApplicationIdentifiers, pb.EndDeviceIdentifiers.DeviceID, ttnpb.EndDeviceFieldPathsTopLevel)
	if !test.AllTrue(
		a.So(err, should.BeNil) || a.So(errors.Stack(err), should.BeEmpty),
		a.So(storedCtx, should.HaveParentContextOrEqual, ctx),
		a.So(stored, should.Resemble, pb),
	) {
		t.Fatal("GetByID assertion failed with non-empty registry")
	}
	ctx = storedCtx

	stored, storedCtx, err = reg.GetByEUI(ctx, *pb.EndDeviceIdentifiers.JoinEUI, *pb.EndDeviceIdentifiers.DevEUI, ttnpb.EndDeviceFieldPathsTopLevel)
	if !test.AllTrue(
		a.So(err, should.BeNil) || a.So(errors.Stack(err), should.BeEmpty),
		a.So(storedCtx, should.HaveParentContextOrEqual, ctx),
		a.So(stored, should.Resemble, pb),
	) {
		t.Fatal("GetByEUI assertion failed with non-empty registry")
	}
	ctx = storedCtx

	stored, storedCtx, err = reg.SetByID(ctx, pb.ApplicationIdentifiers, pb.DeviceID, pbFields,
		func(storedCtx context.Context, stored *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
			a.So(storedCtx, should.HaveParentContextOrEqual, ctx)
			a.So(stored, should.Resemble, pb)
			return stored, nil, nil
		},
	)
	if !test.AllTrue(
		a.So(err, should.BeNil) || a.So(errors.Stack(err), should.BeEmpty),
		a.So(storedCtx, should.HaveParentContextOrEqual, ctx),
		a.So(stored, should.Resemble, pb),
	) {
		t.Fatal("Read-only SetByID assertion failed with non-empty registry")
	}
	ctx = storedCtx

	if !a.So(assertUplinkMatches(ctx, currentUp,
		uplinkMatch{
			EndDevice: pb,
		},
	), should.BeTrue) {
		t.Fatal("RangeByUplinkMatches assertion failed for current session uplink with non-empty registry")
	}
	if !a.So(assertUplinkMatches(ctx, pendingUp,
		uplinkMatch{
			EndDevice: pb,
			IsPending: true,
		},
	), should.BeTrue) {
		t.Fatal("RangeByUplinkMatches assertion failed for pending session uplink with non-empty registry")
	}

	pbOther := CopyEndDevice(pb)
	pbOther.Session.LastFCntUp = pb.Session.LastFCntUp + 1
	pbOther.EndDeviceIdentifiers.DeviceID = "test-dev-other"
	pbOther.EndDeviceIdentifiers.DevEUI = &types.EUI64{0x43, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

	stored, storedCtx, err = reg.GetByID(ctx, pbOther.EndDeviceIdentifiers.ApplicationIdentifiers, pbOther.EndDeviceIdentifiers.DeviceID, ttnpb.EndDeviceFieldPathsTopLevel)
	a.So(storedCtx, should.HaveParentContextOrEqual, ctx)
	if !a.So(err, should.NotBeNil) || !a.So(errors.IsNotFound(err), should.BeTrue) {
		t.Fatalf("Error received: %v", err)
	}
	a.So(stored, should.BeNil)

	stored, storedCtx, err = reg.GetByEUI(ctx, *pbOther.EndDeviceIdentifiers.JoinEUI, *pbOther.EndDeviceIdentifiers.DevEUI, ttnpb.EndDeviceFieldPathsTopLevel)
	a.So(storedCtx, should.HaveParentContextOrEqual, ctx)
	if !a.So(err, should.NotBeNil) || !a.So(errors.IsNotFound(err), should.BeTrue) {
		t.Fatalf("Error received: %v", err)
	}
	a.So(stored, should.BeNil)

	stored, storedCtx, err = CreateDevice(ctx, reg, pbOther,
		"ids.application_ids",
		"ids.dev_eui",
		"ids.device_id",
		"ids.join_eui",
		"pending_session",
		"session",
	)
	a.So(storedCtx, should.HaveParentContextOrEqual, ctx)
	if !a.So(err, should.BeNil) || !a.So(stored, should.NotBeNil) {
		t.Fatalf("Failed to create device: %s", err)
	}
	a.So(stored.CreatedAt, should.HappenAfter, start)
	a.So(stored.UpdatedAt, should.Equal, stored.CreatedAt)
	pbOther.CreatedAt = stored.CreatedAt
	pbOther.UpdatedAt = stored.UpdatedAt
	a.So(stored, should.HaveEmptyDiff, pbOther)

	stored, storedCtx, err = reg.GetByID(ctx, pbOther.EndDeviceIdentifiers.ApplicationIdentifiers, pbOther.EndDeviceIdentifiers.DeviceID, ttnpb.EndDeviceFieldPathsTopLevel)
	a.So(storedCtx, should.HaveParentContextOrEqual, ctx)
	a.So(err, should.BeNil)
	a.So(stored, should.Resemble, pbOther)

	stored, storedCtx, err = reg.GetByEUI(ctx, *pbOther.EndDeviceIdentifiers.JoinEUI, *pbOther.EndDeviceIdentifiers.DevEUI, ttnpb.EndDeviceFieldPathsTopLevel)
	a.So(storedCtx, should.HaveParentContextOrEqual, ctx)
	a.So(err, should.BeNil)
	a.So(stored, should.Resemble, pbOther)

	for i := 0; i < 4; i++ {
		a.So(assertUplinkMatches(ctx, currentUp,
			uplinkMatch{
				EndDevice: pbOther,
			},
			uplinkMatch{
				EndDevice: pb,
			},
		), should.BeTrue)
		a.So(assertUplinkMatches(ctx, pendingUp,
			uplinkMatch{
				EndDevice: pbOther,
				IsPending: true,
			},
			uplinkMatch{
				EndDevice: pb,
				IsPending: true,
			},
		), should.BeTrue)
	}

	err = DeleteDevice(ctx, reg, pb.EndDeviceIdentifiers.ApplicationIdentifiers, pb.EndDeviceIdentifiers.DeviceID)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	stored, storedCtx, err = reg.GetByEUI(ctx, *pb.EndDeviceIdentifiers.JoinEUI, *pb.EndDeviceIdentifiers.DevEUI, ttnpb.EndDeviceFieldPathsTopLevel)
	a.So(storedCtx, should.HaveParentContextOrEqual, ctx)
	if !a.So(err, should.NotBeNil) || !a.So(errors.IsNotFound(err), should.BeTrue) {
		t.Fatalf("Error received: %v", err)
	}
	a.So(stored, should.BeNil)

	stored, storedCtx, err = reg.GetByID(ctx, pb.EndDeviceIdentifiers.ApplicationIdentifiers, pb.EndDeviceIdentifiers.DeviceID, ttnpb.EndDeviceFieldPathsTopLevel)
	a.So(storedCtx, should.HaveParentContextOrEqual, ctx)
	if !a.So(err, should.NotBeNil) || !a.So(errors.IsNotFound(err), should.BeTrue) {
		t.Fatalf("Error received: %v", err)
	}
	a.So(stored, should.BeNil)

	for i := 0; i < 4; i++ {
		a.So(assertUplinkMatches(ctx, currentUp,
			uplinkMatch{
				EndDevice: pbOther,
			},
		), should.BeTrue)
		a.So(assertUplinkMatches(ctx, pendingUp,
			uplinkMatch{
				EndDevice: pbOther,
				IsPending: true,
			},
		), should.BeTrue)
	}

	err = DeleteDevice(ctx, reg, pbOther.EndDeviceIdentifiers.ApplicationIdentifiers, pbOther.EndDeviceIdentifiers.DeviceID)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	stored, storedCtx, err = reg.GetByID(ctx, pbOther.EndDeviceIdentifiers.ApplicationIdentifiers, pbOther.EndDeviceIdentifiers.DeviceID, ttnpb.EndDeviceFieldPathsTopLevel)
	a.So(storedCtx, should.HaveParentContextOrEqual, ctx)
	if !a.So(err, should.NotBeNil) || !a.So(errors.IsNotFound(err), should.BeTrue) {
		t.Fatalf("Error received: %v", err)
	}
	a.So(stored, should.BeNil)

	stored, storedCtx, err = reg.GetByEUI(ctx, *pbOther.EndDeviceIdentifiers.JoinEUI, *pbOther.EndDeviceIdentifiers.DevEUI, ttnpb.EndDeviceFieldPathsTopLevel)
	a.So(storedCtx, should.HaveParentContextOrEqual, ctx)
	if !a.So(err, should.NotBeNil) || !a.So(errors.IsNotFound(err), should.BeTrue) {
		t.Fatalf("Error received: %v", err)
	}
	a.So(stored, should.BeNil)

	for i := 0; i < 4; i++ {
		a.So(assertUplinkMatches(ctx, currentUp), should.BeTrue)
		a.So(assertUplinkMatches(ctx, pendingUp), should.BeTrue)
	}
}

// HandleDeviceRegistryTest runs a DeviceRegistry test suite on reg.
func HandleDeviceRegistryTest(t *testing.T, reg DeviceRegistry) {
	t.Helper()
	test.RunTest(t, test.TestConfig{
		Parallel: true,
		Func: func(ctx context.Context, a *assertions.Assertion) {
			t.Helper()
			test.RunSubtestFromContext(ctx, test.SubtestConfig{
				Name: "1st run",
				Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
					handleDeviceRegistryTest(ctx, reg)
				},
			})
			if t.Failed() {
				t.Skip("Skipping 2nd run")
			}
			test.RunSubtestFromContext(ctx, test.SubtestConfig{
				Name: "2st run",
				Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
					handleDeviceRegistryTest(ctx, reg)
				},
			})
		},
	})
}
