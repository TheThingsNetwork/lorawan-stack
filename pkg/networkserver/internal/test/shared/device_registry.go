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

package test

import (
	"bytes"
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/kr/pretty"
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

func isNil(c interface{}) bool {
	if c == nil {
		return true
	}
	if val := reflect.ValueOf(c); val.Kind() == reflect.Ptr {
		return val.IsNil()
	}
	return false
}

func nilEquality(a interface{}, b interface{}) (bool, bool) {
	if isNil(a) != isNil(b) {
		return false, true
	}
	if isNil(a) {
		return true, true
	}
	return false, false
}

func protoEquals(a proto.Message, b proto.Message) bool {
	if eq, ok := nilEquality(a, b); ok {
		return eq
	}

	ab, err := proto.Marshal(a)
	if err != nil {
		panic(err)
	}
	bb, err := proto.Marshal(b)
	if err != nil {
		panic(err)
	}

	return bytes.Equal(ab, bb)
}

func uplinkMatchEquals(a *UplinkMatch, b *UplinkMatch) bool {
	if m, ok := nilEquality(a, b); ok {
		return m
	}

	return protoEquals(a.ApplicationIdentifiers, b.ApplicationIdentifiers) &&
		a.DeviceID == b.DeviceID &&
		a.LoRaWANVersion == b.LoRaWANVersion &&
		protoEquals(a.FNwkSIntKey, b.FNwkSIntKey) &&
		a.LastFCnt == b.LastFCnt &&
		protoEquals(a.ResetsFCnt, b.ResetsFCnt) &&
		protoEquals(a.Supports32BitFCnt, b.Supports32BitFCnt) &&
		a.IsPending == b.IsPending
}

func handleDeviceRegistryTest(ctx context.Context, reg DeviceRegistry) {
	type uplinkMatch struct {
		*ttnpb.EndDevice
		IsPending bool
	}
	assertUplinkMatch := func(ctx context.Context, up *ttnpb.UplinkMessage, maxAttempts uint, expectedMatch uplinkMatch) bool {
		t, a := test.MustNewTFromContext(ctx)
		t.Helper()

		expectedSession, expectedMACState := expectedMatch.Session, expectedMatch.MacState
		if expectedMatch.IsPending {
			expectedSession, expectedMACState = expectedMatch.PendingSession, expectedMatch.PendingMacState
		}
		var matched bool
		var attempts []*UplinkMatch
		err := reg.RangeByUplinkMatches(ctx, up, CacheTTL, func(storedCtx context.Context, match *UplinkMatch) (bool, error) {
			attempts = append(attempts, match)
			a.So(matched, should.BeFalse)
			a.So(storedCtx, should.HaveParentContextOrEqual, ctx)
			matched = uplinkMatchEquals(match, &UplinkMatch{
				ApplicationIdentifiers: expectedMatch.Ids.ApplicationIds,
				DeviceID:               expectedMatch.Ids.DeviceId,
				LoRaWANVersion:         expectedMACState.LorawanVersion,
				FNwkSIntKey:            expectedSession.Keys.FNwkSIntKey,
				LastFCnt:               expectedSession.LastFCntUp,
				ResetsFCnt:             expectedMatch.GetMacSettings().GetResetsFCnt(),
				Supports32BitFCnt:      expectedMatch.GetMacSettings().GetSupports_32BitFCnt(),
				IsPending:              expectedMatch.IsPending,
			})
			return matched, nil
		})
		if !a.So(err, should.BeNil) {
			t.Errorf("Expected nil error, got: %v\n", errors.Stack(err))
		}
		if !a.So(matched, should.BeTrue) {
			t.Errorf("Device did not match after %d attempts", len(attempts))
		}
		if !a.So(len(attempts), should.BeLessThanOrEqualTo, maxAttempts) {
			t.Errorf("Attempted matches: %s", pretty.Sprint(attempts))
		}
		return !a.Failed()
	}
	assertNoUplinkMatch := func(ctx context.Context, up *ttnpb.UplinkMessage, maxAttempts uint) bool {
		t, a := test.MustNewTFromContext(ctx)
		t.Helper()

		var attempts []*UplinkMatch
		err := reg.RangeByUplinkMatches(ctx, up, CacheTTL, func(storedCtx context.Context, match *UplinkMatch) (bool, error) {
			attempts = append(attempts, match)
			a.So(storedCtx, should.HaveParentContextOrEqual, ctx)
			return false, nil
		})
		if !a.So(len(attempts), should.BeLessThanOrEqualTo, maxAttempts) {
			t.Errorf("Attempted matches: %s", pretty.Sprint(attempts))
		}
		if !a.So(err, should.NotBeNil) {
			return false
		}
		if !a.So(errors.IsNotFound(err), should.BeTrue) {
			t.Errorf("Expected 'Not found' error, got: %v", errors.Stack(err))
		}
		return !a.Failed()
	}
	assertNoDevice := func(ctx context.Context, pb *ttnpb.EndDevice) bool {
		t, a := test.MustNewTFromContext(ctx)
		t.Helper()

		stored, storedCtx, err := reg.GetByID(ctx, pb.Ids.ApplicationIds, pb.Ids.DeviceId, ttnpb.EndDeviceFieldPathsTopLevel)
		if !test.AllTrue(
			a.So(err, should.NotBeNil),
			a.So(errors.IsNotFound(err), should.BeTrue),
			a.So(storedCtx, should.HaveParentContextOrEqual, ctx),
			a.So(stored, should.BeNil),
		) {
			t.Error("GetByID assertion failed with empty registry")
			return false
		}

		stored, storedCtx, err = reg.GetByEUI(ctx, *pb.Ids.JoinEui, *pb.Ids.DevEui, ttnpb.EndDeviceFieldPathsTopLevel)
		if !test.AllTrue(
			a.So(err, should.NotBeNil),
			a.So(errors.IsNotFound(err), should.BeTrue),
			a.So(storedCtx, should.HaveParentContextOrEqual, ctx),
			a.So(stored, should.BeNil),
		) {
			t.Error("GetByEUI assertion failed with empty registry")
			return false
		}

		stored, storedCtx, err = reg.SetByID(ctx, pb.Ids.ApplicationIds, pb.Ids.DeviceId, ttnpb.EndDeviceFieldPathsTopLevel,
			func(storedCtx context.Context, stored *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
				a.So(storedCtx, should.HaveParentContextOrEqual, ctx)
				if !a.So(stored, should.BeNil) {
					t.Error("Registry is not empty")
				}
				return nil, nil, nil
			},
		)
		if !test.AllTrue(
			a.So(err, should.BeNil),
			a.So(storedCtx, should.HaveParentContextOrEqual, ctx),
			a.So(stored, should.BeNil),
		) {
			t.Error("Read-only SetByID assertion failed with empty registry")
			return false
		}
		return !a.Failed()
	}
	assertCreateDevice := func(ctx context.Context, pb *ttnpb.EndDevice, fields ...string) bool {
		t, a := test.MustNewTFromContext(ctx)
		t.Helper()

		if !a.So(assertNoDevice(ctx, pb), should.BeTrue) {
			t.Error("Registry not empty")
			return false
		}

		start := time.Now()

		stored, storedCtx, err := CreateDevice(ctx, reg, pb, fields...)
		pb.CreatedAt = stored.GetCreatedAt()
		pb.UpdatedAt = stored.GetUpdatedAt()
		if !test.AllTrue(
			a.So(err, should.BeNil) || a.So(errors.Stack(err), should.BeEmpty),
			a.So(storedCtx, should.HaveParentContextOrEqual, ctx),
			a.So(*ttnpb.StdTime(stored.GetCreatedAt()), should.HappenAfter, start),
			a.So(stored.GetUpdatedAt(), should.Equal, stored.GetCreatedAt()),
			a.So(stored, should.Resemble, pb),
		) {
			t.Error("Device creation assertion failed")
			return false
		}
		ctx = storedCtx

		stored, storedCtx, err = reg.GetByID(ctx, pb.Ids.ApplicationIds, pb.Ids.DeviceId, ttnpb.EndDeviceFieldPathsTopLevel)
		if !test.AllTrue(
			a.So(err, should.BeNil) || a.So(errors.Stack(err), should.BeEmpty),
			a.So(storedCtx, should.HaveParentContextOrEqual, ctx),
			a.So(stored, should.Resemble, pb),
		) {
			t.Error("GetByID assertion failed with non-empty registry")
			return false
		}
		ctx = storedCtx

		stored, storedCtx, err = reg.GetByEUI(ctx, *pb.Ids.JoinEui, *pb.Ids.DevEui, ttnpb.EndDeviceFieldPathsTopLevel)
		if !test.AllTrue(
			a.So(err, should.BeNil) || a.So(errors.Stack(err), should.BeEmpty),
			a.So(storedCtx, should.HaveParentContextOrEqual, ctx),
			a.So(stored, should.Resemble, pb),
		) {
			t.Error("GetByEUI assertion failed with non-empty registry")
			return false
		}
		ctx = storedCtx

		stored, storedCtx, err = reg.SetByID(ctx, pb.Ids.ApplicationIds, pb.Ids.DeviceId, fields,
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
			t.Error("Read-only SetByID assertion failed with non-empty registry")
			return false
		}
		return true
	}

	t, a := test.MustNewTFromContext(ctx)

	pb := &ttnpb.EndDevice{
		FrequencyPlanId:   test.EUFrequencyPlanID,
		LorawanVersion:    ttnpb.MACVersion_MAC_V1_0_3,
		LorawanPhyVersion: ttnpb.PHYVersion_RP001_V1_0_3_REV_A,
		Ids: &ttnpb.EndDeviceIdentifiers{
			JoinEui:        &types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			DevEui:         &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			ApplicationIds: &ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"},
			DeviceId:       "test-dev",
		},
		Session: &ttnpb.Session{
			DevAddr:    types.DevAddr{0x42, 0xff, 0xff, 0xff},
			LastFCntUp: 41,
			Keys: &ttnpb.SessionKeys{
				FNwkSIntKey: &ttnpb.KeyEnvelope{
					Key: &types.AES128Key{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				},
			},
		},
		MacState: MakeDefaultEU868MACState(ttnpb.Class_CLASS_A, ttnpb.MACVersion_MAC_V1_0_3, ttnpb.PHYVersion_RP001_V1_0_3_REV_A),
		PendingSession: &ttnpb.Session{
			DevAddr: types.DevAddr{0x43, 0xff, 0xff, 0xff},
			Keys: &ttnpb.SessionKeys{
				FNwkSIntKey: &ttnpb.KeyEnvelope{
					EncryptedKey: []byte{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xfe},
					KekLabel:     "kek-label",
				},
				SNwkSIntKey: &ttnpb.KeyEnvelope{
					EncryptedKey: []byte{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xfd},
					KekLabel:     "kek-label",
				},
				NwkSEncKey: &ttnpb.KeyEnvelope{
					EncryptedKey: []byte{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xfc},
					KekLabel:     "kek-label",
				},
			},
		},
		PendingMacState: MakeDefaultEU868MACState(ttnpb.Class_CLASS_A, ttnpb.MACVersion_MAC_V1_1, ttnpb.PHYVersion_RP001_V1_1_REV_B),
	}
	pbFields := []string{
		"frequency_plan_id",
		"ids.application_ids",
		"ids.dev_eui",
		"ids.device_id",
		"ids.join_eui",
		"lorawan_phy_version",
		"lorawan_version",
		"mac_state",
		"pending_mac_state",
		"pending_session",
		"session",
	}

	pbCurrentUp := MakeDataUplink(WithDeviceDataUplinkConfig(pb, false, ttnpb.DataRateIndex_DATA_RATE_2, 1, 1)(DataUplinkConfig{
		DecodePayload: true,
		FPort:         0x01,
	}))
	pbPendingUp := MakeDataUplink(WithDeviceDataUplinkConfig(pb, true, ttnpb.DataRateIndex_DATA_RATE_1, 2, 42)(DataUplinkConfig{
		DecodePayload: true,
		FPort:         0x42,
	}))

	if !a.So(assertNoUplinkMatch(ctx, pbCurrentUp, 0), should.BeTrue) {
		t.Fatal("pb current session uplink matching assertion failed with empty registry")
	}
	if !a.So(assertNoUplinkMatch(ctx, pbPendingUp, 0), should.BeTrue) {
		t.Fatal("pb pending session uplink matching assertion failed with empty registry")
	}

	if !a.So(assertCreateDevice(ctx, pb, pbFields...), should.BeTrue) {
		t.Fatal("pb creation assertion failed")
	}

	if !a.So(assertUplinkMatch(ctx, pbCurrentUp, 1,
		uplinkMatch{
			EndDevice: pb,
		},
	), should.BeTrue) {
		t.Fatal("pb current session uplink matching assertion failed with non-empty registry")
	}
	if !a.So(assertUplinkMatch(ctx, pbPendingUp, 1,
		uplinkMatch{
			EndDevice: pb,
			IsPending: true,
		},
	), should.BeTrue) {
		t.Fatal("pb pending session uplink matching assertion failed with non-empty registry")
	}

	pbOther := CopyEndDevice(pb)
	pbOther.Session.LastFCntUp = pbCurrentUp.Payload.GetMacPayload().FHdr.FCnt
	pbOther.Session.Keys.FNwkSIntKey.Key = &types.AES128Key{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xfe}
	pbOther.PendingSession = nil
	pbOther.Ids.DeviceId = "test-dev-other"
	pbOther.Ids.DevEui = &types.EUI64{0x43, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

	pbOtherCurrentUp := MakeDataUplink(WithDeviceDataUplinkConfig(pbOther, false, ttnpb.DataRateIndex_DATA_RATE_2, 1, 0)(DataUplinkConfig{
		DecodePayload: true,
		FPort:         0x01,
	}))

	if !a.So(assertCreateDevice(ctx, pbOther, pbFields...), should.BeTrue) {
		t.Fatal("pbOther creation assertion failed")
	}

	if !a.So(assertUplinkMatch(ctx, pbOtherCurrentUp, 1,
		uplinkMatch{
			EndDevice: pbOther,
		},
	), should.BeTrue) {
		t.Fatal("pbOther current session uplink matching assertion failed")
	}
	if !a.So(assertUplinkMatch(ctx, pbCurrentUp, 1,
		uplinkMatch{
			EndDevice: pb,
		},
	), should.BeTrue) {
		t.Fatal("cached pb current session uplink matching assertion failed")
	}
	if !a.So(assertUplinkMatch(ctx, pbPendingUp, 1,
		uplinkMatch{
			EndDevice: pb,
			IsPending: true,
		},
	), should.BeTrue) {
		t.Fatal("cached pb pending session uplink matching assertion failed")
	}
	if !a.So(assertUplinkMatch(ctx, pbOtherCurrentUp, 1,
		uplinkMatch{
			EndDevice: pbOther,
		},
	), should.BeTrue) {
		t.Fatal("cached pbOther current session uplink matching assertion failed")
	}

	err := DeleteDevice(ctx, reg, pb.Ids.ApplicationIds, pb.Ids.DeviceId)
	if !a.So(err, should.BeNil) {
		t.Fatalf("pb deletion failed with: %s", errors.Stack(err))
	}

	if !a.So(assertNoDevice(ctx, pb), should.BeTrue) {
		t.Fatalf("Failed to assert registry emptiness after pb deletion")
	}

	if !a.So(assertUplinkMatch(ctx, pbOtherCurrentUp, 1,
		uplinkMatch{
			EndDevice: pbOther,
		},
	), should.BeTrue) {
		t.Fatal("pbOther current session uplink matching assertion failed")
	}

	err = DeleteDevice(ctx, reg, pbOther.Ids.ApplicationIds, pbOther.Ids.DeviceId)
	if !a.So(err, should.BeNil) {
		t.Fatalf("pbOther deletion failed with: %s", errors.Stack(err))
	}

	if !a.So(assertNoDevice(ctx, pbOther), should.BeTrue) {
		t.Fatalf("Failed to assert registry emptiness after pbOther deletion")
	}

	// NOTE: Even though devices are deleted, the matching results are still cached (within TTL) and matched here. That is expected.
	if !a.So(assertUplinkMatch(ctx, pbCurrentUp, 1,
		uplinkMatch{
			EndDevice: pb,
		},
	), should.BeTrue) {
		t.Fatal("cached pb current session uplink matching assertion failed after device deletion")
	}
	if !a.So(assertUplinkMatch(ctx, pbPendingUp, 1,
		uplinkMatch{
			EndDevice: pb,
			IsPending: true,
		},
	), should.BeTrue) {
		t.Fatal("cached pb pending session uplink matching assertion failed after device deletion")
	}
	if !a.So(assertUplinkMatch(ctx, pbOtherCurrentUp, 1,
		uplinkMatch{
			EndDevice: pbOther,
		},
	), should.BeTrue) {
		t.Fatal("pbOther current session uplink matching assertion failed after device deletion")
	}
}

// HandleDeviceRegistryTest runs a DeviceRegistry test suite on reg.
func HandleDeviceRegistryTest(t *testing.T, reg DeviceRegistry) {
	t.Helper()
	test.RunTest(t, test.TestConfig{
		Parallel: true,
		Func: func(ctx context.Context, a *assertions.Assertion) {
			t.Helper()
			if !test.RunSubtestFromContext(ctx, test.SubtestConfig{
				Name: "1st run",
				Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
					handleDeviceRegistryTest(ctx, reg)
				},
			}) {
				t.Skip("Skipping 2nd run")
			}
			sleepFor := 2 * CacheTTL
			t.Logf("Sleeping for %v for cached values to get cleaned up...", sleepFor)
			time.Sleep(sleepFor)
			test.RunSubtestFromContext(ctx, test.SubtestConfig{
				Name: "2nd run",
				Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
					handleDeviceRegistryTest(ctx, reg)
				},
			})
		},
	})
}
