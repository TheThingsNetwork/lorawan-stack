// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

package redis

import (
	"context"
	"encoding/hex"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/smartystreets/assertions"
	"github.com/vmihailenco/msgpack/v5"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestMsgpackCompatibility(t *testing.T) {
	_, ctx := test.New(t)
	cl, flush := test.NewRedis(ctx, "test", "devices")
	defer func() {
		flush()
		cl.Close()
	}()
	makeExpr := func(exprs ...string) string {
		if len(exprs) == 0 {
			panic("no expressions specified")
		}
		return fmt.Sprintf("%s\nand n == %d", strings.Join(exprs, "\nand "), len(exprs))
	}
	makeNumericExpr := func(name string, v interface{}) string {
		return fmt.Sprintf("x.%s == %d", name, v)
	}
	makeStringExpr := func(name string, v string) string {
		return fmt.Sprintf(`x.%s == "%s"`, name, v)
	}
	makeBoolValueExpr := func(name string, v bool) string {
		if !v {
			return fmt.Sprintf("x.%s and not x.%s.value", name, name)
		}
		return fmt.Sprintf("x.%s.value", name)
	}

	makeFNwkSIntUnwrappedKeyExpr := func(v types.AES128Key) string {
		return fmt.Sprintf(`x.f_nwk_s_int_key.key == "%s"`, hex.EncodeToString(v[:]))
	}
	makeFNwkSIntWrappedKeyExpr := func(v *ttnpb.KeyEnvelope) string {
		return fmt.Sprintf(`%s and x.f_nwk_s_int_key.encrypted_key == "%s"`,
			makeStringExpr("f_nwk_s_int_key.kek_label", v.KekLabel),
			hex.EncodeToString(v.EncryptedKey),
		)
	}
	makeLoRaWANVersionExpr := func(v ttnpb.MACVersion) string {
		return makeNumericExpr("lorawan_version", v)
	}
	makeSupports32BitFCntExpr := func(v bool) string {
		return makeBoolValueExpr("supports_32_bit_f_cnt", v)
	}
	makeResetsFCntExpr := func(v bool) string {
		return makeBoolValueExpr("resets_f_cnt", v)
	}
	makeLastFCntExpr := func(v uint32) string {
		return makeNumericExpr("last_f_cnt", v)
	}
	makeUIDExpr := func(v string) string {
		return makeStringExpr("uid", v)
	}

	defaultfNwkSIntKeyWrappedExpr := makeFNwkSIntWrappedKeyExpr(test.DefaultFNwkSIntKeyEnvelopeWrapped)
	defaultfNwkSIntKeyUnwrappedExpr := makeFNwkSIntUnwrappedKeyExpr(test.DefaultFNwkSIntKey)
	defaultLoRaWANVersionExpr := makeLoRaWANVersionExpr(test.DefaultMACVersion)
	makeExprWithDefaults := func(exprs ...string) string {
		return makeExpr(append([]string{
			defaultfNwkSIntKeyWrappedExpr,
			defaultLoRaWANVersionExpr,
		}, exprs...)...)
	}

	for _, tc := range []struct {
		Value   interface{}
		LuaExpr string
	}{
		{
			Value: UplinkMatchPendingSession{
				FNwkSIntKey:    test.DefaultFNwkSIntKeyEnvelopeWrapped,
				LoRaWANVersion: test.DefaultMACVersion,
			},
			LuaExpr: makeExprWithDefaults(),
		},
		{
			Value: UplinkMatchPendingSession{
				FNwkSIntKey:    test.DefaultFNwkSIntKeyEnvelope,
				LoRaWANVersion: test.DefaultMACVersion,
			},
			LuaExpr: makeExpr(
				defaultfNwkSIntKeyUnwrappedExpr,
				defaultLoRaWANVersionExpr,
			),
		},

		{
			Value: UplinkMatchSession{
				FNwkSIntKey:    test.DefaultFNwkSIntKeyEnvelopeWrapped,
				LoRaWANVersion: test.DefaultMACVersion,
			},
			LuaExpr: makeExprWithDefaults(),
		},
		{
			Value: UplinkMatchSession{
				FNwkSIntKey:    test.DefaultFNwkSIntKeyEnvelope,
				LoRaWANVersion: test.DefaultMACVersion,
			},
			LuaExpr: makeExpr(
				defaultfNwkSIntKeyUnwrappedExpr,
				defaultLoRaWANVersionExpr,
			),
		},
		{
			Value: UplinkMatchSession{
				FNwkSIntKey:       test.DefaultFNwkSIntKeyEnvelopeWrapped,
				LoRaWANVersion:    test.DefaultMACVersion,
				Supports32BitFCnt: &ttnpb.BoolValue{Value: false},
			},
			LuaExpr: makeExprWithDefaults(
				makeSupports32BitFCntExpr(false),
			),
		},
		{
			Value: UplinkMatchSession{
				FNwkSIntKey:       test.DefaultFNwkSIntKeyEnvelopeWrapped,
				LoRaWANVersion:    test.DefaultMACVersion,
				Supports32BitFCnt: &ttnpb.BoolValue{Value: true},
			},
			LuaExpr: makeExprWithDefaults(
				makeSupports32BitFCntExpr(true),
			),
		},
		{
			Value: UplinkMatchSession{
				FNwkSIntKey:    test.DefaultFNwkSIntKeyEnvelopeWrapped,
				LoRaWANVersion: test.DefaultMACVersion,
				ResetsFCnt:     &ttnpb.BoolValue{Value: true},
			},
			LuaExpr: makeExprWithDefaults(
				makeResetsFCntExpr(true),
			),
		},
		{
			Value: UplinkMatchSession{
				FNwkSIntKey:    test.DefaultFNwkSIntKeyEnvelopeWrapped,
				LoRaWANVersion: test.DefaultMACVersion,
				ResetsFCnt:     &ttnpb.BoolValue{Value: false},
			},
			LuaExpr: makeExprWithDefaults(
				makeResetsFCntExpr(false),
			),
		},
		{
			Value: UplinkMatchSession{
				FNwkSIntKey:    test.DefaultFNwkSIntKeyEnvelopeWrapped,
				LoRaWANVersion: test.DefaultMACVersion,
				LastFCnt:       42,
			},
			LuaExpr: makeExprWithDefaults(
				makeLastFCntExpr(42),
			),
		},
		{
			Value: UplinkMatchSession{
				FNwkSIntKey:       test.DefaultFNwkSIntKeyEnvelopeWrapped,
				LoRaWANVersion:    test.DefaultMACVersion,
				ResetsFCnt:        &ttnpb.BoolValue{Value: true},
				Supports32BitFCnt: &ttnpb.BoolValue{Value: false},
				LastFCnt:          42,
			},
			LuaExpr: makeExprWithDefaults(
				makeResetsFCntExpr(true),
				makeSupports32BitFCntExpr(false),
				makeLastFCntExpr(42),
			),
		},

		{
			Value: UplinkMatchResult{
				FNwkSIntKey:    test.DefaultFNwkSIntKeyEnvelopeWrapped,
				LoRaWANVersion: test.DefaultMACVersion,
				UID:            "test-uid",
			},
			LuaExpr: makeExprWithDefaults(
				makeUIDExpr("test-uid"),
			),
		},
		{
			Value: UplinkMatchResult{
				FNwkSIntKey:    test.DefaultFNwkSIntKeyEnvelope,
				LoRaWANVersion: test.DefaultMACVersion,
				UID:            "test-uid",
			},
			LuaExpr: makeExpr(
				defaultfNwkSIntKeyUnwrappedExpr,
				defaultLoRaWANVersionExpr,
				makeUIDExpr("test-uid"),
			),
		},
		{
			Value: UplinkMatchResult{
				FNwkSIntKey:       test.DefaultFNwkSIntKeyEnvelopeWrapped,
				LoRaWANVersion:    test.DefaultMACVersion,
				Supports32BitFCnt: &ttnpb.BoolValue{Value: false},
				UID:               "test-uid",
			},
			LuaExpr: makeExprWithDefaults(
				makeUIDExpr("test-uid"),
				makeSupports32BitFCntExpr(false),
			),
		},
		{
			Value: UplinkMatchResult{
				FNwkSIntKey:       test.DefaultFNwkSIntKeyEnvelopeWrapped,
				LoRaWANVersion:    test.DefaultMACVersion,
				Supports32BitFCnt: &ttnpb.BoolValue{Value: true},
				UID:               "test-uid",
			},
			LuaExpr: makeExprWithDefaults(
				makeUIDExpr("test-uid"),
				makeSupports32BitFCntExpr(true),
			),
		},
		{
			Value: UplinkMatchResult{
				FNwkSIntKey:    test.DefaultFNwkSIntKeyEnvelopeWrapped,
				LoRaWANVersion: test.DefaultMACVersion,
				UID:            "test-uid",
				ResetsFCnt:     &ttnpb.BoolValue{Value: true},
			},
			LuaExpr: makeExprWithDefaults(
				makeUIDExpr("test-uid"),
				makeResetsFCntExpr(true),
			),
		},
		{
			Value: UplinkMatchResult{
				FNwkSIntKey:    test.DefaultFNwkSIntKeyEnvelopeWrapped,
				LoRaWANVersion: test.DefaultMACVersion,
				UID:            "test-uid",
				ResetsFCnt:     &ttnpb.BoolValue{Value: false},
			},
			LuaExpr: makeExprWithDefaults(
				makeUIDExpr("test-uid"),
				makeResetsFCntExpr(false),
			),
		},
		{
			Value: UplinkMatchResult{
				FNwkSIntKey:    test.DefaultFNwkSIntKeyEnvelopeWrapped,
				LoRaWANVersion: test.DefaultMACVersion,
				UID:            "test-uid",
				LastFCnt:       42,
			},
			LuaExpr: makeExprWithDefaults(
				makeUIDExpr("test-uid"),
				makeLastFCntExpr(42),
			),
		},
		{
			Value: UplinkMatchResult{
				FNwkSIntKey:       test.DefaultFNwkSIntKeyEnvelopeWrapped,
				LoRaWANVersion:    test.DefaultMACVersion,
				UID:               "test-uid",
				ResetsFCnt:        &ttnpb.BoolValue{Value: true},
				Supports32BitFCnt: &ttnpb.BoolValue{Value: false},
				LastFCnt:          42,
			},
			LuaExpr: makeExprWithDefaults(
				makeUIDExpr("test-uid"),
				makeResetsFCntExpr(true),
				makeSupports32BitFCntExpr(false),
				makeLastFCntExpr(42),
			),
		},
		{
			Value: UplinkMatchResult{
				FNwkSIntKey:       test.DefaultFNwkSIntKeyEnvelopeWrapped,
				LoRaWANVersion:    test.DefaultMACVersion,
				UID:               "test-uid",
				ResetsFCnt:        &ttnpb.BoolValue{Value: true},
				Supports32BitFCnt: &ttnpb.BoolValue{Value: false},
				LastFCnt:          42,
				IsPending:         true,
			},
			LuaExpr: makeExprWithDefaults(
				makeUIDExpr("test-uid"),
				makeResetsFCntExpr(true),
				makeSupports32BitFCntExpr(false),
				makeLastFCntExpr(42),
				`x.is_pending`,
			),
		},
	} {
		tc := tc
		test.RunSubtestFromContext(ctx, test.SubtestConfig{
			Name: fmt.Sprintf("%T/%s", tc.Value, tc.LuaExpr),
			Func: func(ctx context.Context, _ *testing.T, a *assertions.Assertion) {
				b, err := msgpack.Marshal(tc.Value)
				if !a.So(err, should.BeNil) {
					return
				}

				decoded := reflect.New(reflect.ValueOf(tc.Value).Type()).Interface()
				err = msgpack.Unmarshal(b, decoded)
				if a.So(err, should.BeNil) {
					a.So(reflect.ValueOf(decoded).Elem().Interface(), should.Resemble, tc.Value)
				}

				v, err := redis.NewScript(fmt.Sprintf(`local x = cmsgpack.unpack(ARGV[1])
local n = 0
for _, _ in pairs(x) do
	n = n+1
end
return %s`,
					tc.LuaExpr)).Run(ctx, cl, nil, b).Result()
				if a.So(err, should.BeNil) {
					a.So(v, should.Equal, 1)
				}
			},
		})
	}
}
