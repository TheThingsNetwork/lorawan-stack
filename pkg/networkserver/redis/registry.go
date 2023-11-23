// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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
	"bytes"
	"context"
	"runtime/trace"

	"github.com/redis/go-redis/v9"
	"github.com/vmihailenco/msgpack/v5"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/internal/registry"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver"
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal/time"
	ttnredis "go.thethings.network/lorawan-stack/v3/pkg/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"golang.org/x/exp/maps"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	errInvalidFieldmask   = errors.DefineInvalidArgument("invalid_fieldmask", "invalid fieldmask")
	errInvalidIdentifiers = errors.DefineInvalidArgument("invalid_identifiers", "invalid identifiers")
	errReadOnlyField      = errors.DefineInvalidArgument("read_only_field", "read-only field `{field}`")

	errRelayServed = errors.DefineAlreadyExists("relay_served", "`{served}` is already served by `{serving}`")
)

// DeviceSchemaVersion is the Network Server database schema version regarding the devices namespace.
// Bump when a migration is required to the devices namespace.
const DeviceSchemaVersion = 1

// UplinkSchemaVersion is the Network Server database schema version regarding the uplink namespace.
// Bump when a migration is required to the uplink namespace.
const UplinkSchemaVersion = 1

// UnsupportedDeviceMigrationVersionBreakpoint indicates the breakpoint for versions that
// cannot be auto-migrated to latest. Use v3.24.0 of The Things Stack to migrate
// to a supported SchemaVersion before migrating to latest.
const UnsupportedDeviceMigrationVersionBreakpoint = 1

// DeviceRegistry is an implementation of networkserver.DeviceRegistry.
type DeviceRegistry struct {
	Redis   *ttnredis.Client
	LockTTL time.Duration
}

func (r *DeviceRegistry) Init(ctx context.Context) error {
	if err := ttnredis.InitMutex(ctx, r.Redis); err != nil {
		return err
	}
	return nil
}

func (r *DeviceRegistry) uidKey(uid string) string {
	return UIDKey(r.Redis, uid)
}

func (r *DeviceRegistry) addrKey(addr types.DevAddr) string {
	return r.Redis.Key("addr", addr.String())
}

func (r *DeviceRegistry) euiKey(joinEUI, devEUI types.EUI64) string {
	return r.Redis.Key("eui", joinEUI.String(), devEUI.String())
}

func (r *DeviceRegistry) relayRulesMapping(ctx context.Context, dev *ttnpb.EndDevice) map[string]string {
	m := make(map[string]string)
	add := func(rules []*ttnpb.RelayUplinkForwardingRule) {
		for _, rule := range rules {
			if rule.GetDeviceId() == "" {
				continue
			}
			servedUID := unique.ID(ctx, &ttnpb.EndDeviceIdentifiers{
				ApplicationIds: dev.Ids.ApplicationIds,
				DeviceId:       rule.DeviceId,
			})
			path := r.Redis.Key("relay", "rules", servedUID)
			m[path] = servedUID
		}
	}
	for _, rules := range [][]*ttnpb.RelayUplinkForwardingRule{
		dev.GetMacSettings().GetRelay().GetServing().GetUplinkForwardingRules(),
		dev.GetMacSettings().GetDesiredRelay().GetServing().GetUplinkForwardingRules(),
		dev.GetMacState().GetCurrentParameters().GetRelay().GetServing().GetUplinkForwardingRules(),
		dev.GetMacState().GetDesiredParameters().GetRelay().GetServing().GetUplinkForwardingRules(),
		dev.GetPendingMacState().GetCurrentParameters().GetRelay().GetServing().GetUplinkForwardingRules(),
		dev.GetPendingMacState().GetDesiredParameters().GetRelay().GetServing().GetUplinkForwardingRules(),
	} {
		add(rules)
	}
	return m
}

// GetByID gets device by appID, devID.
func (r *DeviceRegistry) GetByID(ctx context.Context, appID *ttnpb.ApplicationIdentifiers, devID string, paths []string) (*ttnpb.EndDevice, context.Context, error) {
	defer trace.StartRegion(ctx, "get end device by id").End()

	ids := &ttnpb.EndDeviceIdentifiers{
		ApplicationIds: appID,
		DeviceId:       devID,
	}
	if err := ids.ValidateContext(ctx); err != nil {
		return nil, ctx, err
	}

	pb := &ttnpb.EndDevice{}
	if err := ttnredis.GetProto(ctx, r.Redis, r.uidKey(unique.ID(ctx, ids))).ScanProto(pb); err != nil {
		return nil, ctx, err
	}
	pb, err := ttnpb.FilterGetEndDevice(pb, paths...)
	if err != nil {
		return nil, ctx, err
	}
	return pb, ctx, nil
}

// BatchGetByID gets devices by appID, deviceIDs.
func (r *DeviceRegistry) BatchGetByID(
	ctx context.Context, appID *ttnpb.ApplicationIdentifiers, deviceIDs []string, paths []string,
) ([]*ttnpb.EndDevice, error) {
	defer trace.StartRegion(ctx, "batch get end device by id").End()

	keys := make([]string, len(deviceIDs))
	for i, devID := range deviceIDs {
		ids := &ttnpb.EndDeviceIdentifiers{
			ApplicationIds: appID,
			DeviceId:       devID,
		}
		if err := ids.ValidateContext(ctx); err != nil {
			return nil, err
		}
		keys[i] = r.uidKey(unique.ID(ctx, ids))
	}

	results, err := r.Redis.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, ttnredis.ConvertError(err)
	}

	protos := make([]*ttnpb.EndDevice, len(deviceIDs))
	for i, result := range results {
		switch cmd := result.(type) {
		case nil:
		case string:
			dev := &ttnpb.EndDevice{}
			if err := ttnredis.UnmarshalProto(cmd, dev); err != nil {
				return nil, err
			}
			dev, err = ttnpb.FilterGetEndDevice(dev, paths...)
			if err != nil {
				return nil, err
			}
			protos[i] = dev
		}
	}
	return protos, nil
}

// GetByEUI gets device by joinEUI, devEUI.
func (r *DeviceRegistry) GetByEUI(ctx context.Context, joinEUI, devEUI types.EUI64, paths []string) (*ttnpb.EndDevice, context.Context, error) {
	defer trace.StartRegion(ctx, "get end device by eui").End()

	pb := &ttnpb.EndDevice{}
	if err := ttnredis.FindProto(ctx, r.Redis, r.euiKey(joinEUI, devEUI), func(uid string) (string, error) {
		return r.uidKey(uid), nil
	}).ScanProto(pb); err != nil {
		return nil, ctx, err
	}
	pb, err := ttnpb.FilterGetEndDevice(pb, paths...)
	if err != nil {
		return nil, ctx, err
	}
	return pb, ctx, nil
}

type UplinkMatchSession struct {
	FNwkSIntKey       *ttnpb.KeyEnvelope
	ResetsFCnt        *ttnpb.BoolValue
	Supports32BitFCnt *ttnpb.BoolValue
	LoRaWANVersion    ttnpb.MACVersion
	LastFCnt          uint32
}

type UplinkMatchPendingSession struct {
	FNwkSIntKey    *ttnpb.KeyEnvelope
	LoRaWANVersion ttnpb.MACVersion
}

func encodeStruct(enc *msgpack.Encoder, fs ...func(enc *msgpack.Encoder) error) error {
	if err := enc.EncodeMapLen(len(fs)); err != nil {
		return err
	}
	for _, f := range fs {
		if err := f(enc); err != nil {
			return err
		}
	}
	return nil
}

func makeEncodeCustomEncoderField(name string, v msgpack.CustomEncoder) func(enc *msgpack.Encoder) error {
	return func(enc *msgpack.Encoder) error {
		if err := enc.EncodeString(name); err != nil {
			return err
		}
		return v.EncodeMsgpack(enc)
	}
}

func makeEncodeFNwkSIntField(v *ttnpb.KeyEnvelope) func(enc *msgpack.Encoder) error {
	return makeEncodeCustomEncoderField("f_nwk_s_int_key", v)
}

func makeEncodeLoRaWANVersionField(v ttnpb.MACVersion) func(enc *msgpack.Encoder) error {
	return makeEncodeCustomEncoderField("lorawan_version", v)
}

func makeEncodeBoolValueField(name string, v *ttnpb.BoolValue) func(enc *msgpack.Encoder) error {
	return func(enc *msgpack.Encoder) error {
		if err := enc.EncodeString(name); err != nil {
			return err
		}
		if err := enc.EncodeMapLen(1); err != nil {
			return err
		}
		if err := enc.EncodeString("value"); err != nil {
			return err
		}
		return enc.EncodeBool(v.Value)
	}
}

func makeEncodeResetsFCntField(v *ttnpb.BoolValue) func(enc *msgpack.Encoder) error {
	return makeEncodeBoolValueField("resets_f_cnt", v)
}

func makeEncodeSupports32BitFCntField(v *ttnpb.BoolValue) func(enc *msgpack.Encoder) error {
	return makeEncodeBoolValueField("supports_32_bit_f_cnt", v)
}

func makeEncodeLastFCntField(v uint32) func(enc *msgpack.Encoder) error {
	return func(enc *msgpack.Encoder) error {
		if err := enc.EncodeString("last_f_cnt"); err != nil {
			return err
		}
		return enc.EncodeUint32(v)
	}
}

var errInvalidFieldCount = errors.DefineCorruption("field_count", "invalid field count '{count}'")

func decodeBoolValue(dec *msgpack.Decoder) (*ttnpb.BoolValue, error) {
	n, err := dec.DecodeMapLen()
	if err != nil {
		return nil, err
	}
	if n != 1 {
		return nil, errInvalidFieldCount.WithAttributes("count", n)
	}

	s, err := dec.DecodeString()
	if err != nil {
		return nil, err
	}
	if s != "value" {
		return nil, errInvalidField.WithAttributes("field", s)
	}

	v, err := dec.DecodeBool()
	if err != nil {
		return nil, err
	}
	return &ttnpb.BoolValue{
		Value: v,
	}, nil
}

var errInvalidField = errors.DefineInvalidArgument("field", "invalid field `{field}`")

// EncodeMsgpack implements msgpack.CustomEncoder interface.
func (v UplinkMatchSession) EncodeMsgpack(enc *msgpack.Encoder) error {
	fs := []func(enc *msgpack.Encoder) error{
		makeEncodeFNwkSIntField(v.FNwkSIntKey),
		makeEncodeLoRaWANVersionField(v.LoRaWANVersion),
	}
	if v.LastFCnt > 0 {
		fs = append(fs, makeEncodeLastFCntField(v.LastFCnt))
	}
	if v.ResetsFCnt != nil {
		fs = append(fs, makeEncodeResetsFCntField(v.ResetsFCnt))
	}
	if v.Supports32BitFCnt != nil {
		fs = append(fs, makeEncodeSupports32BitFCntField(v.Supports32BitFCnt))
	}
	return encodeStruct(enc, fs...)
}

// DecodeMsgpack implements msgpack.CustomDecoder interface.
func (v *UplinkMatchSession) DecodeMsgpack(dec *msgpack.Decoder) error {
	n, err := dec.DecodeMapLen()
	if err != nil {
		return err
	}
	if n > 5 {
		return errInvalidFieldCount.WithAttributes("count", n)
	}
	for i := 0; i < n; i++ {
		s, err := dec.DecodeString()
		if err != nil {
			return err
		}
		switch s {
		case "f_nwk_s_int_key":
			fv := &ttnpb.KeyEnvelope{}
			if err := fv.DecodeMsgpack(dec); err != nil {
				return err
			}
			v.FNwkSIntKey = fv

		case "lorawan_version":
			var fv ttnpb.MACVersion
			if err := fv.DecodeMsgpack(dec); err != nil {
				return err
			}
			v.LoRaWANVersion = fv

		case "resets_f_cnt":
			fv, err := decodeBoolValue(dec)
			if err != nil {
				return err
			}
			v.ResetsFCnt = fv

		case "supports_32_bit_f_cnt":
			fv, err := decodeBoolValue(dec)
			if err != nil {
				return err
			}
			v.Supports32BitFCnt = fv

		case "last_f_cnt":
			fv, err := dec.DecodeUint32()
			if err != nil {
				return err
			}
			v.LastFCnt = fv

		default:
			return errInvalidField.WithAttributes("field", s)
		}
	}
	return nil
}

// EncodeMsgpack implements msgpack.CustomEncoder interface.
func (v UplinkMatchPendingSession) EncodeMsgpack(enc *msgpack.Encoder) error {
	return encodeStruct(enc,
		makeEncodeFNwkSIntField(v.FNwkSIntKey),
		makeEncodeLoRaWANVersionField(v.LoRaWANVersion),
	)
}

// DecodeMsgpack implements msgpack.CustomDecoder interface.
func (v *UplinkMatchPendingSession) DecodeMsgpack(dec *msgpack.Decoder) error {
	n, err := dec.DecodeMapLen()
	if err != nil {
		return err
	}
	if n > 2 {
		return errInvalidFieldCount.WithAttributes("count", n)
	}
	for i := 0; i < n; i++ {
		s, err := dec.DecodeString()
		if err != nil {
			return err
		}
		switch s {
		case "f_nwk_s_int_key":
			fv := &ttnpb.KeyEnvelope{}
			if err := fv.DecodeMsgpack(dec); err != nil {
				return err
			}
			v.FNwkSIntKey = fv

		case "lorawan_version":
			var fv ttnpb.MACVersion
			if err := fv.DecodeMsgpack(dec); err != nil {
				return err
			}
			v.LoRaWANVersion = fv

		default:
			return errInvalidField.WithAttributes("field", s)
		}
	}
	return nil
}

func CurrentAddrKey(addrKey string) string {
	return ttnredis.Key(addrKey, "current")
}

func PendingAddrKey(addrKey string) string {
	return ttnredis.Key(addrKey, "pending")
}

func FieldKey(addrKey string) string {
	return ttnredis.Key(addrKey, "fields")
}

var (
	errNoUplinkMatch = errors.DefineNotFound("no_uplink_match", "no device matches uplink")

	errInvalidMemberType  = errors.DefineInvalidArgument("invalid_member_type", "invalid member type", "uid")
	errMissingSessionData = errors.DefineInvalidArgument("missing_session_data", "missing session data", "uid")
)

// RangeByUplinkMatches ranges over devices matching the uplink.
func (r *DeviceRegistry) RangeByUplinkMatches(ctx context.Context, up *ttnpb.UplinkMessage, f func(context.Context, *networkserver.UplinkMatch) (bool, error)) error {
	defer trace.StartRegion(ctx, "range end devices by uplink matches").End()

	type sessionEntry struct {
		UID     string
		Session string
	}
	buildSortedSessionSet := func(
		scoresCmd *redis.ZSliceCmd, mappingCmd *redis.MapStringStringCmd, usePivot bool, pivot uint16,
	) ([]sessionEntry, error) {
		scores, err := scoresCmd.Result()
		if err != nil {
			if err != redis.Nil {
				return nil, ttnredis.ConvertError(err)
			}
			scores = nil
		}
		mapping, err := mappingCmd.Result()
		if err != nil {
			if err != redis.Nil {
				return nil, ttnredis.ConvertError(err)
			}
			mapping = nil
		}

		floatPivot := float64(pivot)
		head := make([]sessionEntry, 0, len(scores)) // The elements smaller or equal to the pivot, in descending order.
		tail := make([]sessionEntry, 0, len(scores)) // The elements greater than the pivot, in descending order.
		current := tail
		for _, z := range scores {
			uid, ok := z.Member.(string)
			if !ok {
				return nil, errDatabaseCorruption.WithCause(errInvalidMemberType.WithAttributes("uid", uid))
			}
			session, ok := mapping[uid]
			if !ok {
				return nil, errDatabaseCorruption.WithCause(errMissingSessionData.WithAttributes("uid", uid))
			}

			if usePivot && z.Score <= floatPivot {
				tail = current   // Save the tail with appended items.
				current = head   // Start appending to the head slice.
				usePivot = false // Avoid switching slices from this point forward.
			}

			current = append(current, sessionEntry{
				UID:     uid,
				Session: session,
			})
		}

		return append(current, tail...), nil
	}
	ZRangeArgsWithScores := func(ctx context.Context, p redis.Pipeliner, key string) *redis.ZSliceCmd {
		return p.ZRangeArgsWithScores(ctx, redis.ZRangeArgs{
			Key:     key,
			Start:   "-inf",
			Stop:    "inf",
			ByScore: true,
			Rev:     true,
		})
	}

	pld := up.Payload.GetMacPayload()
	ackFlag := pld.FHdr.FCtrl.Ack
	lsb := uint16(pld.FHdr.FCnt)

	addrKey := r.addrKey(types.MustDevAddr(pld.FHdr.DevAddr).OrZero())
	addrKeyCurrent := CurrentAddrKey(addrKey)
	addrKeyPending := PendingAddrKey(addrKey)
	fieldKeyCurrent := FieldKey(addrKeyCurrent)
	fieldKeyPending := FieldKey(addrKeyPending)

	var (
		currentSessionScoresCmd  *redis.ZSliceCmd
		currentSessionMappingCmd *redis.MapStringStringCmd
		pendingSessionScoresCmd  *redis.ZSliceCmd
		pendingSessionMappingCmd *redis.MapStringStringCmd
	)
	if _, err := r.Redis.TxPipelined(ctx, func(p redis.Pipeliner) error {
		currentSessionScoresCmd = ZRangeArgsWithScores(ctx, p, addrKeyCurrent)
		currentSessionMappingCmd = p.HGetAll(ctx, fieldKeyCurrent)
		if !ackFlag {
			pendingSessionScoresCmd = ZRangeArgsWithScores(ctx, p, addrKeyPending)
			pendingSessionMappingCmd = p.HGetAll(ctx, fieldKeyPending)
		}
		return nil
	}); err != nil {
		return ttnredis.ConvertError(err)
	}

	var (
		currentSessionSet []sessionEntry
		pendingSessionSet []sessionEntry
		err               error
	)
	currentSessionSet, err = buildSortedSessionSet(currentSessionScoresCmd, currentSessionMappingCmd, true, lsb)
	if err != nil {
		return err
	}
	if !ackFlag {
		pendingSessionSet, err = buildSortedSessionSet(pendingSessionScoresCmd, pendingSessionMappingCmd, false, 0)
		if err != nil {
			return err
		}
	}

	fillContext := func(ctx context.Context, uid string) (context.Context, *ttnpb.EndDeviceIdentifiers, error) {
		ids, err := unique.ToDeviceID(uid)
		if err != nil {
			log.FromContext(ctx).WithError(err).Error("Failed to parse UID as device identifiers")
			return ctx, nil, errDatabaseCorruption.WithCause(err)
		}
		return ctx, ids, nil
	}

	for _, session := range currentSessionSet {
		ctx, ids, err := fillContext(ctx, session.UID)
		if err != nil {
			return err
		}
		if ids == nil {
			continue
		}

		ses := &UplinkMatchSession{}
		err = msgpack.Unmarshal([]byte(session.Session), ses)
		if err != nil {
			continue
		}

		if uint16(ses.LastFCnt) > lsb {
			if ses.Supports32BitFCnt != nil && !ses.Supports32BitFCnt.Value && (ackFlag || ses.ResetsFCnt == nil || !ses.ResetsFCnt.Value) {
				continue
			}
		}

		stop, err := f(ctx, &networkserver.UplinkMatch{
			ApplicationIdentifiers: ids.ApplicationIds,
			DeviceID:               ids.DeviceId,
			LoRaWANVersion:         ses.LoRaWANVersion,
			FNwkSIntKey:            ses.FNwkSIntKey,
			LastFCnt:               ses.LastFCnt,
			ResetsFCnt:             ses.ResetsFCnt,
			Supports32BitFCnt:      ses.Supports32BitFCnt,
		})
		if err != nil || stop {
			return err
		}
	}
	for _, session := range pendingSessionSet {
		ctx, ids, err := fillContext(ctx, session.UID)
		if err != nil {
			return err
		}
		if ids == nil {
			continue
		}

		ses := &UplinkMatchPendingSession{}
		err = msgpack.Unmarshal([]byte(session.Session), ses)
		if err != nil {
			continue
		}
		stop, err := f(ctx, &networkserver.UplinkMatch{
			ApplicationIdentifiers: ids.ApplicationIds,
			DeviceID:               ids.DeviceId,
			LoRaWANVersion:         ses.LoRaWANVersion,
			FNwkSIntKey:            ses.FNwkSIntKey,
			IsPending:              true,
		})
		if err != nil || stop {
			return err
		}
	}

	return errNoUplinkMatch.New()
}

func removeAddrMapping(ctx context.Context, r redis.Cmdable, addrKey, uid string) (*redis.IntCmd, *redis.IntCmd) {
	return r.ZRem(ctx, addrKey, uid), r.HDel(ctx, FieldKey(addrKey), uid)
}

func MarshalDeviceCurrentSession(dev *ttnpb.EndDevice) ([]byte, error) {
	return msgpack.Marshal(UplinkMatchSession{
		LoRaWANVersion:    dev.GetMacState().GetLorawanVersion(),
		FNwkSIntKey:       dev.GetSession().GetKeys().GetFNwkSIntKey(),
		LastFCnt:          dev.GetSession().GetLastFCntUp(),
		ResetsFCnt:        dev.GetMacSettings().GetResetsFCnt(),
		Supports32BitFCnt: dev.GetMacSettings().GetSupports_32BitFCnt(),
	})
}

func MarshalDevicePendingSession(dev *ttnpb.EndDevice) ([]byte, error) {
	return msgpack.Marshal(UplinkMatchSession{
		LoRaWANVersion: dev.GetPendingMacState().GetLorawanVersion(),
		FNwkSIntKey:    dev.GetPendingSession().GetKeys().GetFNwkSIntKey(),
	})
}

var errInvalidDevice = errors.DefineInvalidArgument("invalid_device", "device is invalid")

// SetByID sets device by appID, devID.
func (r *DeviceRegistry) SetByID(ctx context.Context, appID *ttnpb.ApplicationIdentifiers, devID string, gets []string, f func(ctx context.Context, pb *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, context.Context, error) {
	ids := &ttnpb.EndDeviceIdentifiers{
		ApplicationIds: appID,
		DeviceId:       devID,
	}
	if err := ids.ValidateContext(ctx); err != nil {
		return nil, ctx, err
	}
	uid := unique.ID(ctx, ids)
	uk := r.uidKey(uid)

	lockerID, err := ttnredis.GenerateLockerID()
	if err != nil {
		return nil, ctx, err
	}

	defer trace.StartRegion(ctx, "set end device by id").End()

	var pb *ttnpb.EndDevice
	if err = ttnredis.LockedWatch(ctx, r.Redis, uk, lockerID, r.LockTTL, func(tx *redis.Tx) error {
		cmd := ttnredis.GetProto(ctx, tx, uk)
		stored := &ttnpb.EndDevice{}
		if err := cmd.ScanProto(stored); errors.IsNotFound(err) {
			stored = nil
		} else if err != nil {
			return err
		}

		var err error
		if stored != nil {
			pb = &ttnpb.EndDevice{}
			if err := cmd.ScanProto(pb); err != nil {
				return err
			}
			pb, err = ttnpb.FilterGetEndDevice(pb, gets...)
			if err != nil {
				return err
			}
		}

		var sets []string
		pb, sets, err = f(ctx, pb)
		if err != nil {
			return err
		}
		if err := ttnpb.ProhibitFields(sets,
			"created_at",
			"updated_at",
		); err != nil {
			return errInvalidFieldmask.WithCause(err)
		}

		if stored == nil && pb == nil {
			return nil
		}
		if pb != nil && len(sets) == 0 {
			pb, err = ttnpb.FilterGetEndDevice(stored, gets...)
			return err
		}
		_, err = tx.TxPipelined(ctx, func(p redis.Pipeliner) error {
			if pb == nil && len(sets) == 0 {
				trace.Log(ctx, "ns:redis", "delete end device")
				p.Del(ctx, uk)
				if stored.Ids.JoinEui != nil && stored.Ids.DevEui != nil {
					p.Del(
						ctx,
						r.euiKey(
							types.MustEUI64(stored.Ids.JoinEui).OrZero(),
							types.MustEUI64(stored.Ids.DevEui).OrZero(),
						),
					)
				}
				if stored.PendingSession != nil {
					devAddr := types.MustDevAddr(stored.PendingSession.DevAddr).OrZero()
					removeAddrMapping(
						ctx,
						p,
						PendingAddrKey(r.addrKey(
							devAddr,
						)),
						uid,
					)
				}
				if stored.Session != nil {
					devAddr := types.MustDevAddr(stored.Session.DevAddr).OrZero()
					removeAddrMapping(
						ctx,
						p,
						CurrentAddrKey(r.addrKey(
							devAddr,
						)),
						uid,
					)
				}
				if relayRulesMapping := r.relayRulesMapping(ctx, stored); len(relayRulesMapping) > 0 {
					p.Del(ctx, maps.Keys(relayRulesMapping)...)
				}
				return nil
			}

			if stored == nil {
				trace.Log(ctx, "ns:redis", "create end device")
				if err := ttnpb.RequireFields(sets,
					"ids.application_ids",
					"ids.device_id",
				); err != nil {
					return errInvalidFieldmask.WithCause(err)
				}
				if pb.Ids.ApplicationIds.ApplicationId != appID.ApplicationId || pb.Ids.DeviceId != devID {
					return errInvalidIdentifiers.New()
				}
				if pb.Ids.JoinEui != nil && pb.Ids.DevEui != nil {
					joinEUI := types.MustEUI64(pb.Ids.JoinEui).OrZero()
					devEUI := types.MustEUI64(pb.Ids.DevEui).OrZero()
					ek := r.euiKey(joinEUI, devEUI)
					if err := tx.Watch(ctx, ek).Err(); err != nil {
						return err
					}

					storedUIDStr, err := tx.Get(ctx, ek).Result()
					if errors.Is(err, redis.Nil) {
						p.SetNX(ctx, ek, uid, 0)
					} else if err != nil {
						return err
					} else {
						return registry.UniqueEUIViolationErr(ctx, joinEUI, devEUI, storedUIDStr)
					}
				}
				if relayRulesMapping := r.relayRulesMapping(ctx, pb); len(relayRulesMapping) > 0 {
					keys := maps.Keys(relayRulesMapping)
					if err := tx.Watch(ctx, keys...).Err(); err != nil {
						return err
					}
					ruleUIDs, err := tx.MGet(ctx, keys...).Result()
					if err != nil {
						return err
					}
					for i, ruleUID := range ruleUIDs {
						switch ruleUID := ruleUID.(type) {
						case nil:
						case string:
							return errRelayServed.WithAttributes(
								"served", relayRulesMapping[keys[i]],
								"serving", ruleUID,
							)
						default:
							panic("unreachable")
						}
					}
				}
			} else {
				if ttnpb.HasAnyField(sets, "ids.application_ids.application_id") && pb.Ids.ApplicationIds.ApplicationId != stored.Ids.ApplicationIds.ApplicationId {
					return errReadOnlyField.WithAttributes("field", "ids.application_ids.application_id")
				}
				if ttnpb.HasAnyField(sets, "ids.device_id") && pb.Ids.DeviceId != stored.Ids.DeviceId {
					return errReadOnlyField.WithAttributes("field", "ids.device_id")
				}
				if ttnpb.HasAnyField(sets, "ids.join_eui") && !bytes.Equal(pb.Ids.JoinEui, stored.Ids.JoinEui) {
					return errReadOnlyField.WithAttributes("field", "ids.join_eui")
				}
				if ttnpb.HasAnyField(sets, "ids.dev_eui") && !bytes.Equal(pb.Ids.DevEui, stored.Ids.DevEui) {
					return errReadOnlyField.WithAttributes("field", "ids.dev_eui")
				}
			}

			updated := &ttnpb.EndDevice{}
			if stored != nil {
				if err := cmd.ScanProto(updated); err != nil {
					return err
				}
			}
			updated, err = ttnpb.ApplyEndDeviceFieldMask(updated, pb, sets...)
			if err != nil {
				return err
			}
			updated.UpdatedAt = timestamppb.New(time.Now()) // NOTE: This is not equivalent to timestamppb.Now().
			if stored == nil {
				updated.CreatedAt = updated.UpdatedAt
			}

			if updated.Session != nil && updated.MacState == nil ||
				updated.PendingSession != nil && updated.PendingMacState == nil {
				return errInvalidDevice.New()
			}
			if err := updated.ValidateFields(); err != nil {
				return err
			}

			storedPendingSession := stored.GetPendingSession()
			if updated.PendingSession != nil || storedPendingSession != nil {
				removeStored, setAddr, setFields := func() (bool, bool, bool) {
					switch {
					case updated.PendingSession == nil:
						return true, false, false
					case storedPendingSession == nil:
						return false, true, true
					case !types.MustDevAddr(updated.PendingSession.DevAddr).OrZero().Equal(types.MustDevAddr(storedPendingSession.DevAddr).OrZero()):
						return true, true, true
					}
					storedPendingMACState := stored.GetPendingMacState()
					return false, false, storedPendingMACState == nil ||
						updated.PendingMacState.LorawanVersion != storedPendingMACState.LorawanVersion ||
						!proto.Equal(updated.PendingSession.Keys.FNwkSIntKey, storedPendingSession.Keys.FNwkSIntKey)
				}()
				if removeStored {
					devAddr := types.MustDevAddr(storedPendingSession.DevAddr).OrZero()
					removeAddrMapping(
						ctx,
						p,
						PendingAddrKey(r.addrKey(
							devAddr,
						)),
						uid,
					)
				}
				if setAddr {
					devAddr := types.MustDevAddr(updated.PendingSession.DevAddr).OrZero()
					z := redis.Z{
						Score:  float64(time.Now().UnixNano()),
						Member: uid,
					}
					p.ZAdd(
						ctx,
						PendingAddrKey(r.addrKey(
							devAddr,
						)),
						z,
					)
				}
				if setFields {
					devAddr := types.MustDevAddr(updated.PendingSession.DevAddr).OrZero()
					b, err := MarshalDevicePendingSession(updated)
					if err != nil {
						return err
					}
					p.HSet(
						ctx,
						FieldKey(PendingAddrKey(r.addrKey(
							devAddr,
						))),
						uid,
						b,
					)
				}
			}

			storedSession := stored.GetSession()
			if updated.Session != nil || storedSession != nil {
				removeStored, setAddr, setFields := func() (bool, bool, bool) {
					switch {
					case updated.Session == nil:
						return true, false, false
					case storedSession == nil:
						return false, true, true
					case !types.MustDevAddr(updated.Session.DevAddr).OrZero().Equal(types.MustDevAddr(storedSession.DevAddr).OrZero()):
						return true, true, true
					case updated.Session.LastFCntUp != storedSession.LastFCntUp:
						return false, true, true
					}
					storedMACState := stored.GetMacState()
					storedMACSettings := stored.GetMacSettings()
					return false, false, storedMACState == nil ||
						updated.MacState.LorawanVersion != storedMACState.LorawanVersion ||
						!proto.Equal(updated.Session.Keys.FNwkSIntKey, storedSession.Keys.FNwkSIntKey) ||
						!proto.Equal(updated.MacSettings.GetResetsFCnt(), storedMACSettings.GetResetsFCnt()) ||
						!proto.Equal(updated.MacSettings.GetSupports_32BitFCnt(), storedMACSettings.GetSupports_32BitFCnt())
				}()
				if removeStored {
					devAddr := types.MustDevAddr(storedSession.DevAddr).OrZero()
					removeAddrMapping(
						ctx,
						p,
						CurrentAddrKey(r.addrKey(
							devAddr,
						)),
						uid,
					)
				}
				if setAddr {
					devAddr := types.MustDevAddr(updated.Session.DevAddr).OrZero()
					z := redis.Z{
						Score:  float64(updated.Session.LastFCntUp & 0xffff),
						Member: uid,
					}
					p.ZAdd(
						ctx,
						CurrentAddrKey(r.addrKey(
							devAddr,
						)),
						z,
					)
				}
				if setFields {
					devAddr := types.MustDevAddr(updated.Session.DevAddr).OrZero()
					b, err := MarshalDeviceCurrentSession(updated)
					if err != nil {
						return err
					}
					p.HSet(
						ctx,
						FieldKey(CurrentAddrKey(r.addrKey(
							devAddr,
						))),
						uid,
						b,
					)
				}
			}

			storedRelayRulesMapping := r.relayRulesMapping(ctx, stored)
			updatedRelayRulesMapping := r.relayRulesMapping(ctx, updated)
			if !maps.Equal(storedRelayRulesMapping, updatedRelayRulesMapping) {
				storedRelayRulesKeys := maps.Keys(storedRelayRulesMapping)
				updatedRelayRulesKeys := maps.Keys(updatedRelayRulesMapping)
				if err := tx.Watch(ctx, append(storedRelayRulesKeys, updatedRelayRulesKeys...)...).Err(); err != nil {
					return err
				}

				removed := make([]string, 0, len(storedRelayRulesKeys))
				for _, storedKey := range storedRelayRulesKeys {
					if _, ok := updatedRelayRulesMapping[storedKey]; !ok {
						removed = append(removed, storedKey)
					}
				}
				if len(removed) > 0 {
					p.Del(ctx, removed...)
				}

				added := make([]string, 0, len(updatedRelayRulesKeys))
				for _, updatedKey := range updatedRelayRulesKeys {
					if _, ok := storedRelayRulesMapping[updatedKey]; !ok {
						added = append(added, updatedKey)
					}
				}
				if len(added) > 0 {
					ruleUIDs, err := tx.MGet(ctx, added...).Result()
					if err != nil {
						return err
					}
					for i, ruleUID := range ruleUIDs {
						switch ruleUID := ruleUID.(type) {
						case nil:
						case string:
							return errRelayServed.WithAttributes(
								"served", updatedRelayRulesMapping[added[i]],
								"serving", ruleUID,
							)
						default:
							panic("unreachable")
						}
					}
					addedMapping := make(map[string]string, len(added))
					for _, addedKey := range added {
						addedMapping[addedKey] = uid
					}
					p.MSet(ctx, addedMapping)
				}
			}

			_, err := ttnredis.SetProto(ctx, p, uk, updated, 0)
			if err != nil {
				return err
			}
			pb, err = ttnpb.FilterGetEndDevice(updated, gets...)
			return err
		})
		return err
	}); err != nil {
		return nil, ctx, ttnredis.ConvertError(err)
	}
	return pb, ctx, nil
}

// Range ranges over device uid keys in DeviceRegistry.
func (r *DeviceRegistry) Range(
	ctx context.Context,
	paths []string,
	f func(context.Context, *ttnpb.EndDeviceIdentifiers, *ttnpb.EndDevice) bool,
) error {
	deviceEntityRegex, err := ttnredis.EntityRegex((r.uidKey(unique.GenericID(ctx, "*"))))
	if err != nil {
		return err
	}
	return ttnredis.RangeRedisKeys(
		ctx,
		r.Redis,
		r.uidKey(unique.GenericID(ctx, "*")),
		ttnredis.DefaultRangeCount,
		func(key string) (bool, error) {
			if !deviceEntityRegex.MatchString(key) {
				return true, nil
			}
			dev := &ttnpb.EndDevice{}
			if err := ttnredis.GetProto(ctx, r.Redis, key).ScanProto(dev); err != nil {
				return false, err
			}
			dev, err := ttnpb.FilterGetEndDevice(dev, paths...)
			if err != nil {
				return false, err
			}
			if !f(ctx, dev.Ids, dev) {
				return false, nil
			}
			return true, nil
		})
}

// BatchDelete implements DeviceRegistry.
// This function deletes all the devices in a single transaction.
func (r *DeviceRegistry) BatchDelete(
	ctx context.Context,
	appIDs *ttnpb.ApplicationIdentifiers,
	deviceIDs []string,
) ([]*ttnpb.EndDeviceIdentifiers, error) {
	var (
		uidKeys = make([]string, 0, len(deviceIDs))
		ret     = make([]*ttnpb.EndDeviceIdentifiers, 0)
	)
	for _, devID := range deviceIDs {
		uidKeys = append(
			uidKeys,
			r.uidKey(
				unique.ID(
					ctx,
					&ttnpb.EndDeviceIdentifiers{
						ApplicationIds: appIDs,
						DeviceId:       devID,
					}),
			),
		)
	}

	// Delete the devices in a single transaction.
	transaction := func(tx *redis.Tx) error {
		// Read the devices to delete.
		addrMapping := make(map[string][]string)
		raw, err := tx.MGet(ctx, uidKeys...).Result()
		if err != nil {
			return err
		}
		euiKeys := make([]string, 0, len(raw))
		relayRulesKeys := make([]string, 0, len(raw))
		for _, raw := range raw {
			switch val := raw.(type) {
			case nil:
				continue
			case string:
				dev := &ttnpb.EndDevice{}
				if err := ttnredis.UnmarshalProto(val, dev); err != nil {
					log.FromContext(ctx).WithError(err).Warn("Failed to decode stored end device")
					continue
				}
				ret = append(ret, dev.Ids)
				uid := unique.ID(ctx, dev.GetIds())
				if dev.Ids.JoinEui != nil && dev.Ids.DevEui != nil {
					euiKeys = append(euiKeys, r.euiKey(
						types.MustEUI64(dev.GetIds().GetJoinEui()).OrZero(),
						types.MustEUI64(dev.GetIds().GetDevEui()).OrZero(),
					))
				}
				if dev.PendingSession != nil {
					devAddr := types.MustDevAddr(dev.PendingSession.DevAddr).OrZero()
					addrMapping[uid] = []string{
						PendingAddrKey(r.addrKey(
							devAddr,
						)),
					}
				}
				if dev.Session != nil {
					devAddr := types.MustDevAddr(dev.Session.DevAddr).OrZero()
					addrMapping[uid] = append(addrMapping[uid], []string{
						CurrentAddrKey(r.addrKey(
							devAddr,
						)),
					}...)
				}
				relayRulesKeys = append(relayRulesKeys, maps.Keys(r.relayRulesMapping(ctx, dev))...)
			}
		}
		// If the provided end device identifiers are not registered, it is possible
		// that the `euiKeys` set will be empty. `WATCH` does not allow an empty set
		// of keys to be provided, and as such must be manually skipped.
		if len(euiKeys) > 0 {
			if err := tx.Watch(ctx, euiKeys...).Err(); err != nil {
				return err
			}
		}
		if len(relayRulesKeys) > 0 {
			if err := tx.Watch(ctx, relayRulesKeys...).Err(); err != nil {
				return err
			}
		}
		if _, err := tx.TxPipelined(ctx, func(p redis.Pipeliner) error {
			p.Del(ctx, append(append(uidKeys, euiKeys...), relayRulesKeys...)...)
			for uid, keys := range addrMapping {
				for _, key := range keys {
					removeAddrMapping(ctx, p, key, uid)
				}
			}
			return nil
		}); err != nil {
			return err
		}
		return nil
	}

	defer trace.StartRegion(ctx, "batch delete end devices").End()
	err := r.Redis.Watch(ctx, transaction, uidKeys...)
	if err != nil {
		return nil, ttnredis.ConvertError(err)
	}
	return ret, nil
}
