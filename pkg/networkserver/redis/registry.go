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

package redis

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"runtime/trace"
	"sync"

	"github.com/go-redis/redis/v8"
	pbtypes "github.com/gogo/protobuf/types"
	"github.com/oklog/ulid/v2"
	"github.com/vmihailenco/msgpack/v5"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver"
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal/time"
	ttnredis "go.thethings.network/lorawan-stack/v3/pkg/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
)

var (
	errInvalidFieldmask     = errors.DefineInvalidArgument("invalid_fieldmask", "invalid fieldmask")
	errInvalidIdentifiers   = errors.DefineInvalidArgument("invalid_identifiers", "invalid identifiers")
	errDuplicateIdentifiers = errors.DefineAlreadyExists("duplicate_identifiers", "duplicate identifiers")
	errReadOnlyField        = errors.DefineInvalidArgument("read_only_field", "read-only field `{field}`")
)

// DeviceRegistry is an implementation of networkserver.DeviceRegistry.
type DeviceRegistry struct {
	Redis   *ttnredis.Client
	LockTTL time.Duration

	entropyMu *sync.Mutex
	entropy   io.Reader
}

func (r *DeviceRegistry) Init(ctx context.Context) error {
	if err := ttnredis.InitMutex(ctx, r.Redis); err != nil {
		return err
	}
	r.entropyMu = &sync.Mutex{}
	r.entropy = ulid.Monotonic(rand.New(rand.NewSource(time.Now().UnixNano())), 1000)
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

// GetByID gets device by appID, devID.
func (r *DeviceRegistry) GetByID(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string) (*ttnpb.EndDevice, context.Context, error) {
	ids := ttnpb.EndDeviceIdentifiers{
		ApplicationIdentifiers: appID,
		DeviceID:               devID,
	}
	if err := ids.ValidateContext(ctx); err != nil {
		return nil, ctx, err
	}

	defer trace.StartRegion(ctx, "get end device by id").End()

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
	ResetsFCnt        *pbtypes.BoolValue
	Supports32BitFCnt *pbtypes.BoolValue
	LoRaWANVersion    ttnpb.MACVersion
	LastFCnt          uint32
}

type UplinkMatchPendingSession struct {
	FNwkSIntKey    *ttnpb.KeyEnvelope
	LoRaWANVersion ttnpb.MACVersion
}

type UplinkMatchResult struct {
	FNwkSIntKey       *ttnpb.KeyEnvelope
	ResetsFCnt        *pbtypes.BoolValue
	Supports32BitFCnt *pbtypes.BoolValue
	UID               string
	LoRaWANVersion    ttnpb.MACVersion
	LastFCnt          uint32
	IsPending         bool
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

func makeEncodeBoolValueField(name string, v *pbtypes.BoolValue) func(enc *msgpack.Encoder) error {
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

func makeEncodeResetsFCntField(v *pbtypes.BoolValue) func(enc *msgpack.Encoder) error {
	return makeEncodeBoolValueField("resets_f_cnt", v)
}

func makeEncodeSupports32BitFCntField(v *pbtypes.BoolValue) func(enc *msgpack.Encoder) error {
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

func decodeBoolValue(dec *msgpack.Decoder) (*pbtypes.BoolValue, error) {
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
	return &pbtypes.BoolValue{
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

// EncodeMsgpack implements msgpack.CustomEncoder interface.
func (v UplinkMatchResult) EncodeMsgpack(enc *msgpack.Encoder) error {
	fs := []func(enc *msgpack.Encoder) error{
		makeEncodeFNwkSIntField(v.FNwkSIntKey),
		makeEncodeLoRaWANVersionField(v.LoRaWANVersion),
		func(enc *msgpack.Encoder) error {
			if err := enc.EncodeString("uid"); err != nil {
				return err
			}
			return enc.EncodeString(v.UID)
		},
	}
	if v.LastFCnt > 0 {
		fs = append(fs, makeEncodeLastFCntField(v.LastFCnt))
	}
	if v.IsPending {
		fs = append(fs, func(enc *msgpack.Encoder) error {
			if err := enc.EncodeString("is_pending"); err != nil {
				return err
			}
			return enc.EncodeBool(v.IsPending)
		})
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
func (v *UplinkMatchResult) DecodeMsgpack(dec *msgpack.Decoder) error {
	n, err := dec.DecodeMapLen()
	if err != nil {
		return err
	}
	if n > 7 {
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

		case "uid":
			fv, err := dec.DecodeString()
			if err != nil {
				return err
			}
			v.UID = fv

		case "is_pending":
			fv, err := dec.DecodeBool()
			if err != nil {
				return err
			}
			v.IsPending = fv

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

const noUplinkMatchMarker = '-'

var errNoUplinkMatch = errors.DefineNotFound("no_uplink_match", "no device matches uplink")

// RangeByUplinkMatches ranges over devices matching the uplink.
func (r *DeviceRegistry) RangeByUplinkMatches(ctx context.Context, up *ttnpb.UplinkMessage, cacheTTL time.Duration, f func(context.Context, *networkserver.UplinkMatch) (bool, error)) error {
	defer trace.StartRegion(ctx, "range end devices by dev_addr").End()

	pld := up.Payload.GetMACPayload()
	lsb := uint16(pld.FCnt)

	addrKey := r.addrKey(pld.DevAddr)
	addrKeyCurrent := CurrentAddrKey(addrKey)
	addrKeyPending := PendingAddrKey(addrKey)
	fieldKeyCurrent := FieldKey(addrKeyCurrent)
	fieldKeyPending := FieldKey(addrKeyPending)

	payloadHash := uplinkPayloadHash(up.RawPayload)

	matchResultKey := ttnredis.Key(addrKey, "up", payloadHash, "result")
	matchUIDKeyCurrentLE := ttnredis.Key(addrKeyCurrent, "up", payloadHash, "le")
	matchUIDKeyCurrentGT := ttnredis.Key(addrKeyCurrent, "up", payloadHash, "gt")
	matchUIDKeyPending := ttnredis.Key(addrKeyPending, "up", payloadHash)
	matchFieldKeyCurrent := ttnredis.Key(fieldKeyCurrent, "up", payloadHash)
	matchFieldKeyPending := ttnredis.Key(fieldKeyPending, "up", payloadHash)

	var matchKeys []string
	if pld.Ack {
		matchKeys = []string{
			matchResultKey,

			addrKeyCurrent,
			fieldKeyCurrent,
			matchUIDKeyCurrentLE,
			matchUIDKeyCurrentGT,
			matchFieldKeyCurrent,
		}
	} else {
		matchKeys = []string{
			matchResultKey,

			addrKeyCurrent,
			fieldKeyCurrent,
			matchUIDKeyCurrentLE,
			matchUIDKeyCurrentGT,
			matchFieldKeyCurrent,

			addrKeyPending,
			fieldKeyPending,
			matchUIDKeyPending,
			matchFieldKeyPending,
		}
	}
	vs, err := ttnredis.RunInterfaceSliceScript(ctx, r.Redis, deviceMatchScript, matchKeys, lsb, cacheTTL.Milliseconds()).Result()
	if err != nil {
		if err == redis.Nil {
			return errNoUplinkMatch.New()
		}
		return ttnredis.ConvertError(err)
	}
	if len(vs) < 1 {
		panic("empty match script return value")
	}
	matchType, ok := vs[0].(string)
	if !ok {
		panic(fmt.Sprintf("expected first match script return value element to be a string, got '%v'(%T)", vs[0], vs[0]))
	}
	processResult := func(ctx context.Context, s string) error {
		if s == string(noUplinkMatchMarker) {
			return errNoUplinkMatch.New()
		}
		ctx = log.NewContextWithField(ctx, "match_key", matchResultKey)
		res := &UplinkMatchResult{}
		if err := msgpack.Unmarshal([]byte(s), res); err != nil {
			log.FromContext(ctx).WithError(err).Error("Failed to unmarshal match result")
			return errDatabaseCorruption.WithCause(err)
		}
		ctx = log.NewContextWithField(ctx, "device_uid", res.UID)
		ids, err := unique.ToDeviceID(res.UID)
		if err != nil {
			log.FromContext(ctx).WithError(err).Error("Failed to parse match result UID as device identifiers")
			return errDatabaseCorruption.WithCause(err)
		}
		ok, err = f(ctx, &networkserver.UplinkMatch{
			ApplicationIdentifiers: ids.ApplicationIdentifiers,
			DeviceID:               ids.DeviceID,
			LoRaWANVersion:         res.LoRaWANVersion,
			FNwkSIntKey:            res.FNwkSIntKey,
			LastFCnt:               res.LastFCnt,
			IsPending:              res.IsPending,
			ResetsFCnt:             res.ResetsFCnt,
			Supports32BitFCnt:      res.Supports32BitFCnt,
		})
		if err != nil {
			return errNoUplinkMatch.WithCause(err)
		}
		if !ok {
			if err := r.Redis.Set(ctx, matchResultKey, []byte{noUplinkMatchMarker}, cacheTTL).Err(); err != nil {
				return ttnredis.ConvertError(err)
			}
			return errNoUplinkMatch.New()
		}
		if err = r.Redis.Expire(ctx, matchResultKey, cacheTTL).Err(); err != nil {
			return ttnredis.ConvertError(err)
		}
		return nil
	}
	if matchType == "result" {
		if len(vs) != 2 {
			panic(fmt.Sprintf("expected match script return value of `result` type to contain 2 elements, got %d", len(vs)))
		}
		s, ok := vs[1].(string)
		if !ok {
			panic(fmt.Sprintf("expected second element of match script return value of `result` type to be a string, got '%v'(%T)", vs[1], vs[1]))
		}
		return processResult(ctx, s)
	}

	// NOTE(1): Indexes must be consistent with lua/deviceMatch.lua.
	// NOTE(2): Lua indexing starts from 1.
	const (
		currentLEIdx = 4
		currentGTIdx = 5
		pendingIdx   = 9
	)
	for _, v := range vs[1:] {
		idx, ok := v.(int64)
		if !ok {
			panic(fmt.Sprintf("expected match script `continue` type return value to be int64, got '%v'(%T)", v, v))
		}
		var (
			matchUIDKey   string
			matchFieldKey string
		)
		switch idx {
		case currentLEIdx:
			matchUIDKey = matchUIDKeyCurrentLE
			matchFieldKey = matchFieldKeyCurrent
		case currentGTIdx:
			matchUIDKey = matchUIDKeyCurrentGT
			matchFieldKey = matchFieldKeyCurrent
		case pendingIdx:
			matchUIDKey = matchUIDKeyPending
			matchFieldKey = matchFieldKeyPending
		default:
			panic(fmt.Sprintf("invalid index returned by match script with `continue` type: %d", idx))
		}
		var uid string
		for {
			var s string
			switch {
			case idx == currentGTIdx:
				uid, s, err = func() (string, string, error) {
					var ackArg uint8
					if pld.Ack {
						ackArg = 1
					}
					var args []interface{}
					if uid != "" {
						args = []interface{}{ackArg, uid}
					} else {
						args = []interface{}{ackArg}
					}
					vs, err := ttnredis.RunInterfaceSliceScript(ctx, r.Redis, deviceMatchScanGTScript, []string{matchUIDKey, matchFieldKey}, args...).Result()
					if err != nil {
						return "", "", err
					}
					if len(vs) < 1 {
						panic("empty match scan script return value")
					}
					uid, ok := vs[0].(string)
					if !ok {
						panic(fmt.Sprintf("expected first match scan script return value to be a string, got '%v'(%T)", vs[0], vs[0]))
					}
					s, ok := vs[1].(string)
					if !ok {
						panic(fmt.Sprintf("expected second match scan script return value to be a string, got '%v'(%T)", vs[1], vs[1]))
					}
					return uid, s, nil
				}()
			case uid == "":
				uid, err = r.Redis.LIndex(ctx, matchUIDKey, -1).Result()
			default:
				uid, err = deviceMatchScanScript.Run(ctx, r.Redis, []string{matchUIDKey, matchFieldKey}, uid).Text()
			}
			if err != nil {
				if err == redis.Nil {
					break
				}
				log.FromContext(ctx).WithField("key", matchUIDKey).WithError(err).Error("Failed to scan UID")
				return ttnredis.ConvertError(err)
			}
			ctx := log.NewContextWithFields(ctx, log.Fields(
				"device_uid", uid,
				"match_key", matchUIDKey,
			))
			ids, err := unique.ToDeviceID(uid)
			if err != nil {
				log.FromContext(ctx).WithError(err).Error("Failed to parse UID as device identifiers")
				return errDatabaseCorruption.WithCause(err)
			}

			if s == "" {
				s, err = r.Redis.HGet(ctx, matchFieldKey, uid).Result()
				if err != nil {
					if err == redis.Nil {
						// Another client already processed this entry
						uid = ""
						log.FromContext(ctx).Debug("Another client has already processed this UID")
						continue
					}
					log.FromContext(ctx).WithField("key", matchFieldKey).WithError(err).Error("Failed to get device session")
					return ttnredis.ConvertError(err)
				}
			}
			var m *networkserver.UplinkMatch
			if idx == pendingIdx {
				ses := &UplinkMatchPendingSession{}
				err = msgpack.Unmarshal([]byte(s), ses)
				if err == nil {
					m = &networkserver.UplinkMatch{
						ApplicationIdentifiers: ids.ApplicationIdentifiers,
						DeviceID:               ids.DeviceID,
						LoRaWANVersion:         ses.LoRaWANVersion,
						FNwkSIntKey:            ses.FNwkSIntKey,
						IsPending:              true,
					}
				}
			} else {
				ses := &UplinkMatchSession{}
				err = msgpack.Unmarshal([]byte(s), ses)
				if err == nil {
					m = &networkserver.UplinkMatch{
						ApplicationIdentifiers: ids.ApplicationIdentifiers,
						DeviceID:               ids.DeviceID,
						LoRaWANVersion:         ses.LoRaWANVersion,
						FNwkSIntKey:            ses.FNwkSIntKey,
						LastFCnt:               ses.LastFCnt,
						ResetsFCnt:             ses.ResetsFCnt,
						Supports32BitFCnt:      ses.Supports32BitFCnt,
					}
				}
			}
			if err != nil {
				log.FromContext(ctx).WithError(err).Error("Failed to unmarshal device session")
				return err
			}
			ok, err := f(ctx, m)
			if err != nil {
				return errNoUplinkMatch.WithCause(err)
			}
			if ok {
				b, err := msgpack.Marshal(UplinkMatchResult{
					UID:               uid,
					LoRaWANVersion:    m.LoRaWANVersion,
					FNwkSIntKey:       m.FNwkSIntKey,
					LastFCnt:          m.LastFCnt,
					ResetsFCnt:        m.ResetsFCnt,
					Supports32BitFCnt: m.Supports32BitFCnt,
					IsPending:         m.IsPending,
				})
				if err != nil {
					return err
				}
				_, err = r.Redis.Pipelined(ctx, func(p redis.Pipeliner) error {
					p.Set(ctx, matchResultKey, b, cacheTTL)
					p.Del(ctx,
						matchUIDKeyCurrentLE,
						matchUIDKeyCurrentGT,
						matchUIDKeyPending,
						matchFieldKeyCurrent,
						matchFieldKeyPending,
					)
					return nil
				})
				if err != nil {
					return ttnredis.ConvertError(err)
				}
				return nil
			}
		}
	}
	ok, err = r.Redis.SetNX(ctx, matchResultKey, []byte{noUplinkMatchMarker}, cacheTTL).Result()
	if err != nil {
		return ttnredis.ConvertError(err)
	}
	if !ok {
		// Another instance set the result, while this goroutine was busy with processing.
		// Ideally, this should be done via a Lua script to avoid (highely unlikely) race conditions.
		// TODO: Redis 6.2.0 introduces `GET` argument to `SET`, which can be used above instead.
		// https://github.com/TheThingsNetwork/lorawan-stack/issues/3592
		s, err := r.Redis.Get(ctx, matchResultKey).Result()
		if err != nil {
			return ttnredis.ConvertError(err)
		}
		return processResult(ctx, s)
	}
	return errNoUplinkMatch.New()
}

func equalEUI64(x, y *types.EUI64) bool {
	if x == nil || y == nil {
		return x == y
	}
	return x.Equal(*y)
}

func removeAddrMapping(ctx context.Context, r redis.Cmdable, addrKey, uid string) (*redis.IntCmd, *redis.IntCmd) {
	return r.ZRem(ctx, addrKey, uid), r.HDel(ctx, FieldKey(addrKey), uid)
}

func MarshalDeviceCurrentSession(dev *ttnpb.EndDevice) ([]byte, error) {
	return msgpack.Marshal(UplinkMatchSession{
		LoRaWANVersion:    dev.GetMACState().GetLoRaWANVersion(),
		FNwkSIntKey:       dev.GetSession().GetFNwkSIntKey(),
		LastFCnt:          dev.GetSession().GetLastFCntUp(),
		ResetsFCnt:        dev.GetMACSettings().GetResetsFCnt(),
		Supports32BitFCnt: dev.GetMACSettings().GetSupports32BitFCnt(),
	})
}

func MarshalDevicePendingSession(dev *ttnpb.EndDevice) ([]byte, error) {
	return msgpack.Marshal(UplinkMatchSession{
		LoRaWANVersion: dev.GetPendingMACState().GetLoRaWANVersion(),
		FNwkSIntKey:    dev.GetPendingSession().GetFNwkSIntKey(),
	})
}

var errInvalidDevice = errors.DefineInvalidArgument("invalid_device", "device is invalid")

// SetByID sets device by appID, devID.
func (r *DeviceRegistry) SetByID(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, gets []string, f func(ctx context.Context, pb *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, context.Context, error) {
	ids := ttnpb.EndDeviceIdentifiers{
		ApplicationIdentifiers: appID,
		DeviceID:               devID,
	}
	if err := ids.ValidateContext(ctx); err != nil {
		return nil, ctx, err
	}
	uid := unique.ID(ctx, ids)
	uk := r.uidKey(uid)

	defer trace.StartRegion(ctx, "set end device by id").End()

	var pb *ttnpb.EndDevice
	r.entropyMu.Lock()
	lockID, err := ulid.New(ulid.Timestamp(time.Now()), r.entropy)
	r.entropyMu.Unlock()
	if err != nil {
		return nil, ctx, err
	}
	lockIDStr := lockID.String()
	if err = ttnredis.LockedWatch(ctx, r.Redis, uk, lockIDStr, r.LockTTL, func(tx *redis.Tx) error {
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
				p.Del(ctx, uk)
				p.Del(ctx, uidLastInvalidationKey(r.Redis, uid))
				if stored.JoinEUI != nil && stored.DevEUI != nil {
					p.Del(ctx, r.euiKey(*stored.JoinEUI, *stored.DevEUI))
				}
				if stored.PendingSession != nil {
					removeAddrMapping(ctx, p, PendingAddrKey(r.addrKey(stored.PendingSession.DevAddr)), uid)
				}
				if stored.Session != nil {
					removeAddrMapping(ctx, p, CurrentAddrKey(r.addrKey(stored.Session.DevAddr)), uid)
				}
				return nil
			}

			if err = pb.ValidateFields(sets...); err != nil {
				return err
			}
			if stored == nil {
				if err := ttnpb.RequireFields(sets,
					"ids.application_ids",
					"ids.device_id",
				); err != nil {
					return errInvalidFieldmask.WithCause(err)
				}
				if pb.ApplicationIdentifiers != appID || pb.DeviceID != devID {
					return errInvalidIdentifiers.New()
				}
				if pb.JoinEUI != nil && pb.DevEUI != nil {
					ek := r.euiKey(*pb.JoinEUI, *pb.DevEUI)

					if err := ttnredis.LockMutex(ctx, tx, ek, lockIDStr, r.LockTTL); err != nil {
						return err
					}
					if err := tx.Watch(ctx, ek).Err(); err != nil {
						return err
					}
					i, err := tx.Exists(ctx, ek).Result()
					if err != nil {
						return err
					}
					if i != 0 {
						return errDuplicateIdentifiers.New()
					}
					p.Set(ctx, ek, uid, 0)
					ttnredis.UnlockMutex(ctx, p, ek, lockIDStr, r.LockTTL)
				}
			} else {
				if ttnpb.HasAnyField(sets, "ids.application_ids.application_id") && pb.ApplicationID != stored.ApplicationID {
					return errReadOnlyField.WithAttributes("field", "ids.application_ids.application_id")
				}
				if ttnpb.HasAnyField(sets, "ids.device_id") && pb.DeviceID != stored.DeviceID {
					return errReadOnlyField.WithAttributes("field", "ids.device_id")
				}
				if ttnpb.HasAnyField(sets, "ids.join_eui") && !equalEUI64(pb.JoinEUI, stored.JoinEUI) {
					return errReadOnlyField.WithAttributes("field", "ids.join_eui")
				}
				if ttnpb.HasAnyField(sets, "ids.dev_eui") && !equalEUI64(pb.DevEUI, stored.DevEUI) {
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
			updated.UpdatedAt = time.Now().UTC()
			if stored == nil {
				updated.CreatedAt = updated.UpdatedAt
			}

			if updated.Session != nil && updated.MACState == nil ||
				updated.PendingSession != nil && updated.PendingMACState == nil {
				return errInvalidDevice.New()
			}

			storedPendingSession := stored.GetPendingSession()
			if updated.PendingSession != nil || storedPendingSession != nil {
				removeStored, setAddr, setFields := func() (bool, bool, bool) {
					switch {
					case updated.PendingSession == nil:
						return true, false, false
					case storedPendingSession == nil:
						return false, true, true
					case !updated.PendingSession.DevAddr.Equal(storedPendingSession.DevAddr):
						return true, true, true
					}
					storedPendingMACState := stored.GetPendingMACState()
					return false, false, storedPendingMACState == nil ||
						updated.PendingMACState.LoRaWANVersion != storedPendingMACState.LoRaWANVersion ||
						!updated.PendingSession.FNwkSIntKey.Equal(storedPendingSession.FNwkSIntKey)
				}()
				if removeStored {
					removeAddrMapping(ctx, p, PendingAddrKey(r.addrKey(storedPendingSession.DevAddr)), uid)
				}
				if setAddr {
					p.ZAdd(ctx, PendingAddrKey(r.addrKey(updated.PendingSession.DevAddr)), &redis.Z{
						Score:  float64(time.Now().UnixNano()),
						Member: uid,
					})
				}
				if setFields {
					b, err := MarshalDevicePendingSession(updated)
					if err != nil {
						return err
					}
					p.HSet(ctx, FieldKey(PendingAddrKey(r.addrKey(updated.PendingSession.DevAddr))), uid, b)
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
					case !updated.Session.DevAddr.Equal(storedSession.DevAddr):
						return true, true, true
					case updated.Session.LastFCntUp != storedSession.LastFCntUp:
						return false, true, true
					}
					storedMACState := stored.GetMACState()
					storedMACSettings := stored.GetMACSettings()
					return false, false, storedMACState == nil ||
						updated.MACState.LoRaWANVersion != storedMACState.LoRaWANVersion ||
						!updated.Session.FNwkSIntKey.Equal(storedSession.FNwkSIntKey) ||
						!updated.MACSettings.GetResetsFCnt().Equal(storedMACSettings.GetResetsFCnt()) ||
						!updated.MACSettings.GetSupports32BitFCnt().Equal(storedMACSettings.GetSupports32BitFCnt())
				}()
				if removeStored {
					removeAddrMapping(ctx, p, CurrentAddrKey(r.addrKey(storedSession.DevAddr)), uid)
				}
				if setAddr {
					p.ZAdd(ctx, CurrentAddrKey(r.addrKey(updated.Session.DevAddr)), &redis.Z{
						Score:  float64(updated.Session.LastFCntUp & 0xffff),
						Member: uid,
					})
				}
				if setFields {
					b, err := MarshalDeviceCurrentSession(updated)
					if err != nil {
						return err
					}
					p.HSet(ctx, FieldKey(CurrentAddrKey(r.addrKey(updated.Session.DevAddr))), uid, b)
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
