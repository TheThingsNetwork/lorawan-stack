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
	"bytes"
	"context"
	"encoding"
	"io"
	"math/rand"
	"runtime/trace"
	"strconv"
	"sync"
	"time"

	"github.com/go-redis/redis/v7"
	pbtypes "github.com/gogo/protobuf/types"
	"github.com/oklog/ulid/v2"
	"github.com/vmihailenco/msgpack/v5"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver"
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

const (
	fieldKey = "fields"

	shortFCntKey = "16bit"
	longFCntKey  = "32bit"
	pendingKey   = "pending"
)

// DeviceRegistry is an implementation of networkserver.DeviceRegistry.
type DeviceRegistry struct {
	Redis   *ttnredis.Client
	LockTTL time.Duration

	entropyMu *sync.Mutex
	entropy   io.Reader
}

func (r *DeviceRegistry) Init() error {
	if err := ttnredis.InitMutex(r.Redis); err != nil {
		return err
	}
	r.entropyMu = &sync.Mutex{}
	r.entropy = ulid.Monotonic(rand.New(rand.NewSource(time.Now().UnixNano())), 1000)
	return nil
}

func (r *DeviceRegistry) uidKey(uid string) string {
	return deviceUIDKey(r.Redis, uid)
}

func (r *DeviceRegistry) addrKey(addr types.DevAddr) string {
	return r.Redis.Key("addr", addr.String())
}

func (r *DeviceRegistry) euiKey(joinEUI, devEUI types.EUI64) string {
	return r.Redis.Key("eui", joinEUI.String(), devEUI.String())
}

func deviceSupports32BitFCnt(pb *ttnpb.EndDevice) bool {
	if pb.GetMACSettings().GetSupports32BitFCnt() != nil {
		return pb.MACSettings.Supports32BitFCnt.Value
	}
	return true
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
	if err := ttnredis.GetProto(r.Redis, r.uidKey(unique.ID(ctx, ids))).ScanProto(pb); err != nil {
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
	if err := ttnredis.FindProto(r.Redis, r.euiKey(joinEUI, devEUI), func(uid string) (string, error) {
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

type uplinkMatch struct {
	appID                   ttnpb.ApplicationIdentifiers
	devID                   string
	loRaWANVersion          ttnpb.MACVersion
	fNwkSIntKeyKey          *types.AES128Key
	fNwkSIntKeyKEKLabel     string
	fNwkSIntKeyEncryptedKey []byte
	resetsFCnt              *bool

	fCnt      uint32
	lastFCnt  uint32
	isPending bool
}

func (m uplinkMatch) ApplicationIdentifiers() ttnpb.ApplicationIdentifiers {
	return m.appID
}

func (m uplinkMatch) DeviceID() string {
	return m.devID
}

func (m uplinkMatch) LoRaWANVersion() ttnpb.MACVersion {
	return m.loRaWANVersion
}

func (m uplinkMatch) FNwkSIntKey() *ttnpb.KeyEnvelope {
	return &ttnpb.KeyEnvelope{
		Key:          m.fNwkSIntKeyKey,
		KEKLabel:     m.fNwkSIntKeyKEKLabel,
		EncryptedKey: m.fNwkSIntKeyEncryptedKey,
	}
}

func (m uplinkMatch) FCnt() uint32 {
	return m.fCnt
}

func (m uplinkMatch) LastFCnt() uint32 {
	return m.lastFCnt
}

func (m uplinkMatch) ResetsFCnt() *pbtypes.BoolValue {
	if m.resetsFCnt == nil {
		return nil
	}
	return &pbtypes.BoolValue{
		Value: *m.resetsFCnt,
	}
}

func (m uplinkMatch) IsPending() bool {
	return m.isPending
}

type matchKeySet struct {
	ShortFCnt          string
	LongFCntNoRollover string
	LongFCntRollover   string
	Pending            string
	Legacy             string
}

func boolPtr(v bool) *bool {
	return &v
}

func encodeBool(v bool) uint8 {
	if v {
		return 1
	}
	return 0
}

var errInvalidFormat = errors.DefineInvalidArgument("invalid_format", "invalid value format")

func decodeString(v interface{}) (string, error) {
	s, ok := v.(string)
	if !ok {
		return "", errInvalidFormat.New()
	}
	return s, nil
}

func decodeBool(v interface{}) (bool, error) {
	s, err := decodeString(v)
	if err != nil {
		return false, err
	}
	switch s {
	case "0":
		return false, nil
	case "1":
		return true, nil
	default:
		return false, errInvalidFormat.New()
	}
}

func decodeBytes(v interface{}) ([]byte, error) {
	s, err := decodeString(v)
	if err != nil {
		return nil, err
	}
	return []byte(s), nil
}

func decodeBinary(src interface{}, dst encoding.BinaryUnmarshaler) error {
	b, err := decodeBytes(src)
	if err != nil {
		return err
	}
	if err = dst.UnmarshalBinary(b); err != nil {
		return err
	}
	return nil
}

func decodeMACVersion(v interface{}) (ttnpb.MACVersion, error) {
	var ver ttnpb.MACVersion
	if err := decodeBinary(v, &ver); err != nil {
		return -1, err
	}
	if err := ver.Validate(); err != nil {
		return -1, err
	}
	return ver, nil
}

func decodeAES128Key(v interface{}) (*types.AES128Key, error) {
	key := &types.AES128Key{}
	if err := decodeBinary(v, key); err != nil {
		return nil, err
	}
	return key, nil
}

func getUplinkMatch(ctx context.Context, r redis.Cmdable, inputKeys, processingKeys matchKeySet, appID ttnpb.ApplicationIdentifiers, devID string, devAddr types.DevAddr, lsb uint16, matchKey, uidKey string) ([]*uplinkMatch, error) {
	var isPending bool
	switch matchKey {
	case inputKeys.ShortFCnt, processingKeys.ShortFCnt,
		inputKeys.LongFCntNoRollover, processingKeys.LongFCntNoRollover,
		inputKeys.LongFCntRollover, processingKeys.LongFCntRollover:
	case inputKeys.Pending, processingKeys.Pending:
		// NOTE: While the legacy key may point to a pending session, we can safely ignore that
		// and let device rejoin and perform migration to new format.
		isPending = true
	case inputKeys.Legacy, processingKeys.Legacy:
		pb := &ttnpb.EndDevice{}
		if err := ttnredis.GetProto(r, uidKey).ScanProto(pb); err != nil {
			return nil, err
		}
		ms := make([]*uplinkMatch, 0, 2)
		if pb.GetMACState() != nil &&
			pb.GetSession() != nil &&
			pb.Session.DevAddr.Equal(devAddr) &&
			pb.Session.FNwkSIntKey != nil {
			var resetsFCnt *bool
			if pb.GetMACSettings().GetResetsFCnt() != nil {
				resetsFCnt = &pb.MACSettings.ResetsFCnt.Value
			}
			fCnt := uint32(lsb)
			switch {
			case !deviceSupports32BitFCnt(pb), fCnt >= pb.Session.LastFCntUp, fCnt == 0 && pb.Session.LastFCntUp == 0:
			case fCnt > pb.Session.LastFCntUp&0xffff:
				fCnt |= pb.Session.LastFCntUp &^ 0xffff
			case pb.Session.LastFCntUp < 0xffff0000:
				fCnt |= (pb.Session.LastFCntUp + 0x10000) &^ 0xffff
			}
			if pb.ApplicationIdentifiers != appID || pb.DeviceID != devID {
				return nil, errDatabaseCorruption.New()
			}
			ms = append(ms, &uplinkMatch{
				appID:                   appID,
				devID:                   devID,
				loRaWANVersion:          pb.MACState.LoRaWANVersion,
				fNwkSIntKeyKey:          pb.Session.FNwkSIntKey.Key,
				fNwkSIntKeyKEKLabel:     pb.Session.FNwkSIntKey.KEKLabel,
				fNwkSIntKeyEncryptedKey: pb.Session.FNwkSIntKey.EncryptedKey,
				resetsFCnt:              resetsFCnt,
				fCnt:                    fCnt,
				lastFCnt:                pb.Session.LastFCntUp,
			})
		}
		if pb.GetPendingMACState() != nil &&
			pb.GetPendingSession() != nil &&
			pb.PendingSession.DevAddr.Equal(devAddr) &&
			pb.PendingSession.FNwkSIntKey != nil {
			ms = append(ms, &uplinkMatch{
				appID:                   appID,
				devID:                   devID,
				loRaWANVersion:          pb.PendingMACState.LoRaWANVersion,
				fNwkSIntKeyKey:          pb.PendingSession.FNwkSIntKey.Key,
				fNwkSIntKeyKEKLabel:     pb.PendingSession.FNwkSIntKey.KEKLabel,
				fNwkSIntKeyEncryptedKey: pb.PendingSession.FNwkSIntKey.EncryptedKey,
				fCnt:                    uint32(lsb),
				isPending:               true,
			})
		}
		return ms, nil
	default:
		return nil, errDatabaseCorruption.New()
	}

	var fields []string
	if isPending {
		fields = []string{
			"pending_mac_state.lorawan_version",
			"pending_session.keys.f_nwk_s_int_key.encrypted_key",
			"pending_session.keys.f_nwk_s_int_key.kek_label",
			"pending_session.keys.f_nwk_s_int_key.key",
		}
	} else {
		fields = []string{
			"mac_settings.resets_f_cnt",
			"mac_state.lorawan_version",
			"session.keys.f_nwk_s_int_key.encrypted_key",
			"session.keys.f_nwk_s_int_key.kek_label",
			"session.keys.f_nwk_s_int_key.key",
			"session.last_f_cnt_up",
		}
	}

	vs, err := r.HMGet(ttnredis.Key(uidKey, fieldKey), fields...).Result()
	if err != nil {
		return nil, ttnredis.ConvertError(err)
	}
	if len(vs) != len(fields) {
		return nil, errDatabaseCorruption.New()
	}
	m := &uplinkMatch{
		appID:     appID,
		devID:     devID,
		fCnt:      uint32(lsb),
		isPending: isPending,
	}
	for i, v := range vs {
		if v == nil {
			continue
		}
		if isPending {
			switch fields[i] {
			case "pending_mac_state.lorawan_version":
				v, err := decodeMACVersion(v)
				if err != nil {
					return nil, errDatabaseCorruption.WithCause(err)
				}
				m.loRaWANVersion = v
			case "pending_session.keys.f_nwk_s_int_key.encrypted_key":
				v, err := decodeBytes(v)
				if err != nil {
					return nil, errDatabaseCorruption.WithCause(err)
				}
				m.fNwkSIntKeyEncryptedKey = v
			case "pending_session.keys.f_nwk_s_int_key.kek_label":
				v, err := decodeString(v)
				if err != nil {
					return nil, errDatabaseCorruption.WithCause(err)
				}
				m.fNwkSIntKeyKEKLabel = v
			case "pending_session.keys.f_nwk_s_int_key.key":
				v, err := decodeAES128Key(v)
				if err != nil {
					return nil, errDatabaseCorruption.WithCause(err)
				}
				m.fNwkSIntKeyKey = v
			default:
				return nil, errDatabaseCorruption.New()
			}
		} else {
			switch fields[i] {
			case "mac_settings.resets_f_cnt.value":
				v, err := decodeBool(v)
				if err != nil {
					return nil, errDatabaseCorruption.WithCause(err)
				}
				m.resetsFCnt = &v
			case "mac_state.lorawan_version":
				v, err := decodeMACVersion(v)
				if err != nil {
					return nil, errDatabaseCorruption.WithCause(err)
				}
				m.loRaWANVersion = v
			case "session.keys.f_nwk_s_int_key.encrypted_key":
				v, err := decodeBytes(v)
				if err != nil {
					return nil, errDatabaseCorruption.WithCause(err)
				}
				m.fNwkSIntKeyEncryptedKey = v
			case "session.keys.f_nwk_s_int_key.kek_label":
				v, err := decodeString(v)
				if err != nil {
					return nil, errDatabaseCorruption.WithCause(err)
				}
				m.fNwkSIntKeyKEKLabel = v
			case "session.keys.f_nwk_s_int_key.key":
				v, err := decodeAES128Key(v)
				if err != nil {
					return nil, errDatabaseCorruption.WithCause(err)
				}
				m.fNwkSIntKeyKey = v
			case "session.last_f_cnt_up":
				s, err := decodeString(v)
				if err != nil {
					return nil, errDatabaseCorruption.WithCause(err)
				}
				n, err := strconv.ParseUint(s, 10, 32)
				if err != nil {
					return nil, errDatabaseCorruption.WithCause(errInvalidFormat.WithCause(err))
				}
				m.lastFCnt = uint32(n)
				switch matchKey {
				case inputKeys.ShortFCnt, processingKeys.ShortFCnt:
				case inputKeys.LongFCntNoRollover, processingKeys.LongFCntNoRollover:
					m.fCnt |= m.lastFCnt &^ 0xffff
				case inputKeys.LongFCntRollover, processingKeys.LongFCntRollover:
					if n > 0xffff_0000 {
						return nil, errDatabaseCorruption.New()
					}
					m.fCnt |= (m.lastFCnt + 0x1_0000) &^ 0xffff
				default:
					return nil, errDatabaseCorruption.New()
				}

			default:
				return nil, errDatabaseCorruption.New()
			}
		}
	}
	return []*uplinkMatch{m}, nil
}

var errNoUplinkMatch = errors.DefineNotFound("no_uplink_match", "no device matches uplink")

// RangeByUplinkMatches ranges over devices matching the uplink.
func (r *DeviceRegistry) RangeByUplinkMatches(ctx context.Context, up *ttnpb.UplinkMessage, cacheTTL time.Duration, f func(context.Context, networkserver.UplinkMatch) bool) error {
	defer trace.StartRegion(ctx, "range end devices by dev_addr").End()
	if cacheTTL < time.Millisecond {
		// TODO: Remove once https://github.com/TheThingsNetwork/lorawan-stack/issues/2698 is closed.
		cacheTTL = time.Millisecond
	}
	pld := up.Payload.GetMACPayload()
	addrKey := r.addrKey(pld.DevAddr)

	addrKeys := struct {
		ShortFCnt string
		LongFCnt  string
		Pending   string
		Legacy    string
	}{
		ShortFCnt: ttnredis.Key(addrKey, shortFCntKey),
		LongFCnt:  ttnredis.Key(addrKey, longFCntKey),
		Pending:   ttnredis.Key(addrKey, pendingKey),
		Legacy:    addrKey,
	}

	payloadHash := uplinkPayloadHash(up.RawPayload)

	matchKeys := struct {
		Match      string
		Input      matchKeySet
		Processing matchKeySet
	}{
		Match: ttnredis.Key(addrKey, "match", payloadHash),
		Input: matchKeySet{
			ShortFCnt:          ttnredis.Key(addrKeys.ShortFCnt, payloadHash),
			LongFCntNoRollover: ttnredis.Key(addrKeys.LongFCnt, payloadHash, "no-rollover"),
			LongFCntRollover:   ttnredis.Key(addrKeys.LongFCnt, payloadHash, "rollover"),
			Pending:            ttnredis.Key(addrKeys.Pending, payloadHash),
			Legacy:             ttnredis.Key(addrKeys.Legacy, payloadHash),
		},
	}
	matchKeys.Processing = matchKeySet{
		ShortFCnt:          ttnredis.Key(matchKeys.Input.ShortFCnt, "processing"),
		LongFCntNoRollover: ttnredis.Key(matchKeys.Input.LongFCntNoRollover, "processing"),
		LongFCntRollover:   ttnredis.Key(matchKeys.Input.LongFCntRollover, "processing"),
		Pending:            ttnredis.Key(matchKeys.Input.Pending, "processing"),
		Legacy:             ttnredis.Key(matchKeys.Input.Legacy, "processing"),
	}

	type MatchResult struct {
		Key string
		UID string
	}
	lsb := uint16(pld.FCnt)
	v, err := deviceMatchScript.Run(r.Redis, []string{
		matchKeys.Match,

		addrKeys.ShortFCnt,
		addrKeys.LongFCnt,
		addrKeys.Pending,
		addrKeys.Legacy,

		matchKeys.Input.ShortFCnt,
		matchKeys.Processing.ShortFCnt,

		matchKeys.Input.LongFCntNoRollover,
		matchKeys.Processing.LongFCntNoRollover,

		matchKeys.Input.LongFCntRollover,
		matchKeys.Processing.LongFCntRollover,

		matchKeys.Input.Pending,
		matchKeys.Processing.Pending,

		matchKeys.Input.Legacy,
		matchKeys.Processing.Legacy,
	}, lsb, cacheTTL.Milliseconds()).Result()
	if err != nil {
		if err == redis.Nil {
			return errNoUplinkMatch
		}
		return ttnredis.ConvertError(err)
	}
	// NOTE: Lua indexing starts from 1.
	var scanKeys []string
	switch v := v.(type) {
	case []interface{}:
		keyIndexes := make([]uint8, 0, len(v))
		for _, iface := range v {
			idx, ok := iface.(int64)
			if !ok {
				log.FromContext(ctx).Error("Failed to process match script return value as index")
				return errDatabaseCorruption.New()
			}
			if idx < 6 || idx > 15 {
				log.FromContext(ctx).WithField("index", idx).Error("Index out of bounds returned by match script")
				return errDatabaseCorruption.New()
			}
			keyIndexes = append(keyIndexes, uint8(idx))
		}
		scanKeys = make([]string, 0, 10)
		for i := 0; i < len(keyIndexes); i++ {
			switch keyIndexes[i] {
			case 6:
				scanKeys = append(scanKeys, matchKeys.Input.ShortFCnt, matchKeys.Processing.ShortFCnt)
			case 7:
				scanKeys = append(scanKeys, matchKeys.Processing.ShortFCnt)
				continue
			case 8:
				scanKeys = append(scanKeys, matchKeys.Input.LongFCntNoRollover, matchKeys.Processing.LongFCntNoRollover)
			case 9:
				scanKeys = append(scanKeys, matchKeys.Processing.LongFCntNoRollover)
				continue
			case 10:
				scanKeys = append(scanKeys, matchKeys.Input.LongFCntRollover, matchKeys.Processing.LongFCntRollover)
			case 11:
				scanKeys = append(scanKeys, matchKeys.Processing.LongFCntRollover)
				continue
			case 12:
				scanKeys = append(scanKeys, matchKeys.Input.Pending, matchKeys.Processing.Pending)
			case 13:
				scanKeys = append(scanKeys, matchKeys.Processing.Pending)
				continue
			case 14:
				scanKeys = append(scanKeys, matchKeys.Input.Legacy, matchKeys.Processing.Legacy)
			case 15:
				scanKeys = append(scanKeys, matchKeys.Processing.Legacy)
				continue
			default:
				panic("invalid index")
			}
			if len(keyIndexes) > i+1 && keyIndexes[i+1] == keyIndexes[i]+1 {
				// Next key is "processing" key, which is already added above - skip
				i++
			}
		}

	case string:
		res := &MatchResult{}
		if err := msgpack.Unmarshal([]byte(v), res); err != nil {
			return err
		}

		ids, err := unique.ToDeviceID(res.UID)
		if err != nil {
			log.FromContext(ctx).WithError(err).Error("Failed to parse match uid value recorded")
			return errDatabaseCorruption.WithCause(err)
		}
		ms, err := getUplinkMatch(ctx, r.Redis, matchKeys.Input, matchKeys.Processing, ids.ApplicationIdentifiers, ids.DeviceID, pld.DevAddr, lsb, res.Key, r.uidKey(res.UID))
		if err != nil {
			return err
		}
		for _, m := range ms {
			if f(ctx, m) {
				return nil
			}
		}
		return errNoUplinkMatch

	default:
		log.FromContext(ctx).WithField("value", v).WithError(err).Error("Failed to process matching result")
		return errDatabaseCorruption.New()
	}

	args := make([]interface{}, 1, 2)
	args[0] = cacheTTL.Milliseconds()
	for len(scanKeys) > 0 {
		v, err := deviceMatchScanScript.Run(r.Redis, scanKeys, args...).Result()
		if err != nil && err != redis.Nil {
			return ttnredis.ConvertError(err)
		}
		if err == redis.Nil {
			return errNoUplinkMatch
		}
		vs, ok := v.([]interface{})
		if !ok || len(vs) != 2 || vs[0] == nil {
			log.FromContext(ctx).WithField("value", v).Error("Invalid value returned by device match scan script")
			return errDatabaseCorruption.New()
		}
		i, ok := vs[0].(int64)
		switch {
		case !ok:
			log.FromContext(ctx).WithField("index", vs[0]).Error("Invalid index returned by device match scan script")
			return errDatabaseCorruption.New()
		case i < 1, int64(len(scanKeys)) < i:
			log.FromContext(ctx).WithFields(log.Fields(
				"index", i,
				"scan_key_count", len(scanKeys),
			)).Error("Invalid index returned by device match scan script")
			return errDatabaseCorruption.New()
		case i > 1:
			scanKeys = scanKeys[i-1:]
		}

		vsUID := vs[1]
		if vsUID == nil {
			return errDatabaseCorruption.New()
		}
		uid, err := decodeString(vsUID)
		if err != nil {
			log.FromContext(ctx).WithError(err).Error("Failed to parse uid returned by device match scan script as a string")
			return errDatabaseCorruption.WithCause(err)
		}

		ids, err := unique.ToDeviceID(uid)
		if err != nil {
			log.FromContext(ctx).WithError(err).Error("Failed to parse uid returned by device match scan script as device identifiers")
			return errDatabaseCorruption.WithCause(err)
		}
		ms, err := getUplinkMatch(ctx, r.Redis, matchKeys.Input, matchKeys.Processing, ids.ApplicationIdentifiers, ids.DeviceID, pld.DevAddr, lsb, scanKeys[0], r.uidKey(uid))
		if err != nil {
			return err
		}
		for _, m := range ms {
			if f(ctx, m) {
				b, err := msgpack.Marshal(MatchResult{
					Key: scanKeys[0],
					UID: uid,
				})
				if err != nil {
					return err
				}
				_, err = r.Redis.Pipelined(func(p redis.Pipeliner) error {
					p.Set(matchKeys.Match, string(b), cacheTTL)
					p.Del(scanKeys...)
					return nil
				})
				if err != nil {
					return ttnredis.ConvertError(err)
				}
				return nil
			}
		}
		if len(args) > 1 {
			args[1] = vsUID
		} else {
			args = append(args, vsUID)
		}
	}
	return errNoUplinkMatch
}

func equalEUI64(x, y *types.EUI64) bool {
	if x == nil || y == nil {
		return x == y
	}
	return x.Equal(*y)
}

func removeLegacyDevAddrMapping(r redis.Cmdable, addrKey, uid string) {
	r.SRem(addrKey, uid)
}

func removeCurrentDevAddrMapping(r redis.Cmdable, addrKey, uid string, supports32Bit bool) {
	removeLegacyDevAddrMapping(r, addrKey, uid)
	if !supports32Bit {
		r.ZRem(ttnredis.Key(addrKey, shortFCntKey), uid)
	} else {
		r.ZRem(ttnredis.Key(addrKey, longFCntKey), uid)
	}
}

func removePendingDevAddrMapping(r redis.Cmdable, addrKey, uid string) {
	removeLegacyDevAddrMapping(r, addrKey, uid)
	r.SRem(ttnredis.Key(addrKey, pendingKey), uid)
}

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
		cmd := ttnredis.GetProto(tx, uk)
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
		_, err = tx.TxPipelined(func(p redis.Pipeliner) error {
			if pb == nil && len(sets) == 0 {
				p.Del(uk)
				p.Del(deviceUIDLastInvalidationKey(r.Redis, uid))
				if stored.JoinEUI != nil && stored.DevEUI != nil {
					p.Del(r.euiKey(*stored.JoinEUI, *stored.DevEUI))
				}
				if stored.PendingSession != nil {
					removePendingDevAddrMapping(p, r.addrKey(stored.PendingSession.DevAddr), uid)
				}
				if stored.Session != nil {
					removeCurrentDevAddrMapping(p, r.addrKey(stored.Session.DevAddr), uid, deviceSupports32BitFCnt(stored))
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
					if err := tx.Watch(ek).Err(); err != nil {
						return err
					}
					i, err := tx.Exists(ek).Result()
					if err != nil {
						return err
					}
					if i != 0 {
						return errDuplicateIdentifiers.New()
					}
					p.Set(ek, uid, 0)
					ttnredis.UnlockMutex(p, ek, lockIDStr, r.LockTTL)
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

			var delFields []string
			var setFields []interface{}

			// NOTE: The following sequence of switches use concept of "container" - a container is the pointer type "containing" the field value we're interested in.

			switch storedCont, updatedCont := stored.GetMACSettings().GetResetsFCnt(), updated.MACSettings.GetResetsFCnt(); {
			case storedCont == nil && updatedCont == nil:
			case updatedCont == nil:
				delFields = append(delFields, "mac_settings.resets_f_cnt")
			case storedCont == nil, storedCont.Value != updatedCont.Value:
				setFields = append(setFields, "mac_settings.resets_f_cnt", encodeBool(updatedCont.Value))
			}

			switch storedCont, updatedCont := stored.GetMACState(), updated.MACState; {
			case storedCont == nil && updatedCont == nil:
			case updatedCont == nil:
				delFields = append(delFields, "mac_state.lorawan_version")
			case storedCont == nil, storedCont.LoRaWANVersion != updatedCont.LoRaWANVersion:
				setFields = append(setFields, "mac_state.lorawan_version", updatedCont.LoRaWANVersion)
			}

			switch storedCont, updatedCont := stored.GetPendingMACState(), updated.PendingMACState; {
			case storedCont == nil && updatedCont == nil:
			case updatedCont == nil:
				delFields = append(delFields, "pending_mac_state.lorawan_version")
			case storedCont == nil, storedCont.LoRaWANVersion != updatedCont.LoRaWANVersion:
				setFields = append(setFields, "pending_mac_state.lorawan_version", updatedCont.LoRaWANVersion)
			}

			switch storedCont, updatedCont := stored.GetPendingSession().GetSessionKeys().GetFNwkSIntKey(), updated.GetPendingSession().GetSessionKeys().GetFNwkSIntKey(); {
			case storedCont == nil && updatedCont == nil:
			case updatedCont == nil:
				delFields = append(delFields, "pending_session.keys.f_nwk_s_int_key.kek_label")
				delFields = append(delFields, "pending_session.keys.f_nwk_s_int_key.encrypted_key")
			case storedCont == nil, !bytes.Equal(storedCont.EncryptedKey, updatedCont.EncryptedKey):
				setFields = append(setFields, "pending_session.keys.f_nwk_s_int_key.encrypted_key", updatedCont.EncryptedKey)
				fallthrough
			case storedCont == nil, storedCont.KEKLabel != updatedCont.KEKLabel:
				setFields = append(setFields, "pending_session.keys.f_nwk_s_int_key.kek_label", updatedCont.KEKLabel)
			}
			switch storedCont, updatedCont := stored.GetPendingSession().GetSessionKeys().GetFNwkSIntKey().GetKey(), updated.GetPendingSession().GetSessionKeys().GetFNwkSIntKey().GetKey(); {
			case storedCont == nil && updatedCont == nil:
			case updatedCont == nil:
				delFields = append(delFields, "pending_session.keys.f_nwk_s_int_key.key")
			case storedCont == nil, storedCont.Equal(*updatedCont):
				setFields = append(setFields, "pending_session.keys.f_nwk_s_int_key.key", *updatedCont)
			}

			switch storedCont, updatedCont := stored.GetSession().GetSessionKeys().GetFNwkSIntKey(), updated.GetSession().GetSessionKeys().GetFNwkSIntKey(); {
			case storedCont == nil && updatedCont == nil:
			case updatedCont == nil:
				delFields = append(delFields, "session.keys.f_nwk_s_int_key.kek_label")
				delFields = append(delFields, "session.keys.f_nwk_s_int_key.encrypted_key")
			case storedCont == nil, !bytes.Equal(storedCont.EncryptedKey, updatedCont.EncryptedKey):
				setFields = append(setFields, "session.keys.f_nwk_s_int_key.encrypted_key", updatedCont.EncryptedKey)
				fallthrough
			case storedCont == nil, storedCont.KEKLabel != updatedCont.KEKLabel:
				setFields = append(setFields, "session.keys.f_nwk_s_int_key.kek_label", updatedCont.KEKLabel)
			}
			switch storedCont, updatedCont := stored.GetSession().GetSessionKeys().GetFNwkSIntKey().GetKey(), updated.GetSession().GetSessionKeys().GetFNwkSIntKey().GetKey(); {
			case storedCont == nil && updatedCont == nil:
			case updatedCont == nil:
				delFields = append(delFields, "session.keys.f_nwk_s_int_key.key")
			case storedCont == nil, storedCont.Equal(*updatedCont):
				setFields = append(setFields, "session.keys.f_nwk_s_int_key.key", *updatedCont)
			}

			switch storedCont, updatedCont := stored.GetSession(), updated.GetSession(); {
			case storedCont == nil && updatedCont == nil:
			case updatedCont == nil:
				delFields = append(delFields, "session.last_f_cnt_up")
			case storedCont == nil, storedCont.LastFCntUp != updatedCont.LastFCntUp:
				setFields = append(setFields, "session.last_f_cnt_up", updatedCont.LastFCntUp)
			}

			fk := ttnredis.Key(uk, fieldKey)
			if len(delFields) > 0 {
				p.HDel(fk, delFields...)
			}
			if len(setFields) > 0 {
				p.HSet(fk, setFields...)
			}

			storedSupports32BitFCnt := deviceSupports32BitFCnt(stored)
			updatedSupports32BitFCnt := deviceSupports32BitFCnt(updated)

			if stored.GetPendingSession() == nil || updated.GetPendingSession() == nil ||
				!updated.PendingSession.DevAddr.Equal(stored.PendingSession.DevAddr) {
				if stored.GetPendingSession() != nil {
					removePendingDevAddrMapping(p, r.addrKey(stored.PendingSession.DevAddr), uid)
				}

				if updated.GetPendingSession() != nil {
					p.SAdd(ttnredis.Key(r.addrKey(updated.PendingSession.DevAddr), pendingKey), uid)
				}
			}

			if stored.GetSession() == nil || updated.GetSession() == nil ||
				!updated.Session.DevAddr.Equal(stored.Session.DevAddr) ||
				storedSupports32BitFCnt != updatedSupports32BitFCnt {
				if stored.GetSession() != nil {
					removeCurrentDevAddrMapping(p, r.addrKey(stored.Session.DevAddr), uid, storedSupports32BitFCnt)
				}

				if updated.GetSession() != nil {
					addrKey := r.addrKey(updated.Session.DevAddr)
					z := &redis.Z{
						Score:  float64(updated.Session.LastFCntUp),
						Member: uid,
					}
					if !updatedSupports32BitFCnt {
						p.ZAdd(ttnredis.Key(addrKey, shortFCntKey), z)
					} else {
						p.ZAdd(ttnredis.Key(addrKey, longFCntKey), z)
					}
				}
			}

			_, err := ttnredis.SetProto(p, uk, updated, 0)
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
