// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

package networkserver

import (
	"bytes"
	"context"

	"go.thethings.network/lorawan-stack/v3/pkg/band"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto/cryptoutil"
	"go.thethings.network/lorawan-stack/v3/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal/time"
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver/mac"
	"go.thethings.network/lorawan-stack/v3/pkg/specification/relayspec"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type relayKeyService struct {
	devices DeviceRegistry
	keys    crypto.KeyService
}

var _ mac.RelayKeyService = (*relayKeyService)(nil)

// BatchDeriveRootWorSKey implements mac.RelayKeyService.
func (r *relayKeyService) BatchDeriveRootWorSKey(
	ctx context.Context, appID *ttnpb.ApplicationIdentifiers, deviceIDs []string, sessionKeyIDs [][]byte,
) (devAddrs []*types.DevAddr, keys []*types.AES128Key, err error) {
	if len(deviceIDs) != len(sessionKeyIDs) {
		panic("device IDs and session key IDs must have the same length")
	}
	if len(deviceIDs) == 0 {
		return nil, nil, nil
	}
	devices, err := r.devices.BatchGetByID(
		ctx,
		appID,
		deviceIDs,
		[]string{
			"pending_session.dev_addr",
			"pending_session.keys.nwk_s_enc_key.encrypted_key",
			"pending_session.keys.nwk_s_enc_key.kek_label",
			"pending_session.keys.nwk_s_enc_key.key",
			"pending_session.keys.session_key_id",
			"session.dev_addr",
			"session.keys.nwk_s_enc_key.encrypted_key",
			"session.keys.nwk_s_enc_key.kek_label",
			"session.keys.nwk_s_enc_key.key",
			"session.keys.session_key_id",
		},
	)
	if err != nil {
		return nil, nil, err
	}
	devAddrs, keys = make([]*types.DevAddr, len(deviceIDs)), make([]*types.AES128Key, len(deviceIDs))
	for i, dev := range devices {
		var devAddr types.DevAddr
		var keyEnvelope *ttnpb.KeyEnvelope
		switch {
		case dev.GetPendingSession().GetKeys().GetNwkSEncKey() != nil &&
			bytes.Equal(dev.PendingSession.Keys.SessionKeyId, sessionKeyIDs[i]):
			copy(devAddr[:], dev.PendingSession.DevAddr)
			keyEnvelope = dev.PendingSession.Keys.NwkSEncKey
		case dev.GetSession().GetKeys().GetNwkSEncKey() != nil &&
			bytes.Equal(dev.Session.Keys.SessionKeyId, sessionKeyIDs[i]):
			copy(devAddr[:], dev.Session.DevAddr)
			keyEnvelope = dev.Session.Keys.NwkSEncKey
		default:
			continue
		}
		key, err := cryptoutil.UnwrapAES128Key(ctx, keyEnvelope, r.keys)
		if err != nil {
			return nil, nil, err
		}
		key = crypto.DeriveRootWorSKey(key)
		devAddrs[i], keys[i] = &devAddr, &key
	}
	return devAddrs, keys, nil
}

func (ns *NetworkServer) relayKeyService() mac.RelayKeyService {
	return &relayKeyService{ns.devices, ns.KeyService()}
}

func relayUplinkToken(ids *ttnpb.EndDeviceIdentifiers, sessionKeyID []byte, fullFCnt uint32) ([]byte, error) {
	token := &ttnpb.RelayUplinkToken{
		Ids:          ids,
		SessionKeyId: sessionKeyID,
		FullFCnt:     fullFCnt,
	}
	if err := token.ValidateFields(); err != nil {
		return nil, err
	}
	return proto.Marshal(token)
}

func handleRelayForwardingProtocol(
	ctx context.Context,
	dev *ttnpb.EndDevice,
	fullFCnt uint32,
	phy *band.Band,
	up *ttnpb.UplinkMessage,
	keyService crypto.KeyService,
) (_ *ttnpb.UplinkMessage, queuedEvents []events.Event, err error) {
	defer func() {
		if err != nil {
			queuedEvents = append(queuedEvents, evtDropRelayUplink.NewWithIdentifiersAndData(ctx, dev.Ids, err))
		}
	}()
	session := dev.Session
	nwkSEncKey := session.Keys.NwkSEncKey
	key, err := cryptoutil.UnwrapAES128Key(ctx, nwkSEncKey, keyService)
	if err != nil {
		return nil, queuedEvents, err
	}
	rawPayload, err := crypto.DecryptUplink(
		key, types.DevAddr(session.DevAddr), fullFCnt, up.Payload.GetMacPayload().FrmPayload,
	)
	if err != nil {
		return nil, queuedEvents, err
	}
	req := &ttnpb.RelayForwardUplinkReq{}
	if err := lorawan.UnmarshalRelayForwardUplinkReq(phy, rawPayload, req); err != nil {
		return nil, queuedEvents, err
	}
	uplinkToken, err := relayUplinkToken(dev.Ids, session.Keys.SessionKeyId, fullFCnt)
	if err != nil {
		return nil, queuedEvents, err
	}
	mdTime, mdReceivedAt := up.ReceivedAt, up.ReceivedAt
	var mdGPSTime *timestamppb.Timestamp
	for _, md := range up.RxMetadata {
		if md.GpsTime != nil {
			mdTime, mdGPSTime, mdReceivedAt = md.Time, md.GpsTime, md.ReceivedAt
			break
		}
		if mdReceivedAt == nil {
			mdTime, mdGPSTime, mdReceivedAt = md.Time, md.GpsTime, md.ReceivedAt
			continue
		}
		if md.ReceivedAt != nil && md.ReceivedAt.AsTime().Before(mdReceivedAt.AsTime()) {
			mdTime, mdGPSTime, mdReceivedAt = md.Time, md.GpsTime, md.ReceivedAt
		}
	}
	adjustTime := func(ts *timestamppb.Timestamp) *timestamppb.Timestamp {
		if ts == nil {
			return nil
		}
		t := ts.AsTime().Add(-(up.ConsumedAirtime.AsDuration() + phy.RelayForwardDelay))
		return timestamppb.New(t)
	}
	up = &ttnpb.UplinkMessage{
		RawPayload: req.RawPayload,
		Settings: &ttnpb.TxSettings{
			DataRate:  req.DataRate,
			Frequency: req.Frequency,
			Time:      adjustTime(up.Settings.Time),
		},
		RxMetadata: []*ttnpb.RxMetadata{
			{
				GatewayIds: relayspec.GatewayIdentifiers,
				Relay: &ttnpb.RelayMetadata{
					DeviceId:   dev.Ids.DeviceId,
					WorChannel: req.WorChannel,
				},
				Time:        adjustTime(mdTime),
				Rssi:        float32(req.Rssi),
				ChannelRssi: float32(req.Rssi),
				Snr:         float32(req.Snr),
				UplinkToken: uplinkToken,
				GpsTime:     adjustTime(mdGPSTime),
				ReceivedAt:  adjustTime(mdReceivedAt),
			},
		},
		ReceivedAt:     up.ReceivedAt,
		CorrelationIds: events.CorrelationIDsFromContext(ctx),
	}
	if err := up.ValidateFields(); err != nil {
		return nil, queuedEvents, err
	}
	up.Payload = &ttnpb.Message{}
	if err := lorawan.UnmarshalMessage(up.RawPayload, up.Payload); err != nil {
		return nil, queuedEvents, err
	}
	if err := up.Payload.ValidateFields(); err != nil {
		return nil, queuedEvents, err
	}
	queuedEvents = append(queuedEvents, evtProcessRelayUplink.NewWithIdentifiersAndData(ctx, dev.Ids, up))
	return up, queuedEvents, nil
}

func relayLoopbackFunc(
	conn *grpc.ClientConn,
	up *ttnpb.UplinkMessage,
	callOpts ...grpc.CallOption,
) func(context.Context) error {
	client := ttnpb.NewGsNsClient(conn)
	return func(ctx context.Context) error {
		switch _, err := client.HandleUplink(ctx, up, callOpts...); {
		case err == nil, errors.IsNotFound(err), errors.IsAlreadyExists(err):
			return nil
		default:
			return err
		}
	}
}

func relayUpdateRules(
	deviceID string, sessionKeyID []byte, rules []*ttnpb.ServingRelayParameters_UplinkForwardingRule,
) bool {
	for _, rule := range rules {
		if rule.DeviceId != deviceID || bytes.Equal(rule.SessionKeyId, sessionKeyID) {
			continue
		}
		rule.LastWFCnt = 0
		rule.SessionKeyId = sessionKeyID
		return true
	}
	return false
}

var relayDeliverSessionKeysPaths = ttnpb.AddFields(
	deviceDownlinkFullPaths[:],
	"mac_settings.desired_relay.mode.serving.uplink_forwarding_rules",
	"mac_state.desired_parameters.relay.mode.serving.uplink_forwarding_rules",
	"pending_mac_state.desired_parameters.relay.mode.serving.uplink_forwarding_rules",
)

func (ns *NetworkServer) deliverRelaySessionKeys(ctx context.Context, dev *ttnpb.EndDevice, sessionKeyID []byte) error {
	for _, served := range []*ttnpb.ServedRelayParameters{
		dev.MacSettings.GetRelay().GetServed(),
		dev.MacSettings.GetDesiredRelay().GetServed(),
	} {
		if served == nil {
			continue
		}
		serving, ctx, err := ns.devices.SetByID(
			ctx,
			dev.Ids.ApplicationIds,
			served.ServingDeviceId,
			relayDeliverSessionKeysPaths,
			func(ctx context.Context, serving *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
				if serving == nil {
					return nil, nil, nil
				}
				var paths []string
				for path, rules := range map[string][]*ttnpb.ServingRelayParameters_UplinkForwardingRule{
					"mac_settings.desired_relay.mode.serving.uplink_forwarding_rules":                 serving.MacSettings.GetDesiredRelay().GetServing().GetUplinkForwardingRules(),                     // nolint:lll
					"mac_state.desired_parameters.relay.mode.serving.uplink_forwarding_rules":         serving.MacState.GetDesiredParameters().GetRelay().GetServing().GetUplinkForwardingRules(),        // nolint:lll
					"pending_mac_state.desired_parameters.relay.mode.serving.uplink_forwarding_rules": serving.PendingMacState.GetDesiredParameters().GetRelay().GetServing().GetUplinkForwardingRules(), // nolint:lll
				} {
					if relayUpdateRules(dev.Ids.DeviceId, sessionKeyID, rules) {
						paths = ttnpb.AddFields(paths, path)
					}
				}
				return serving, paths, nil
			},
		)
		if err != nil {
			return err
		}
		if err := ns.updateDataDownlinkTask(ctx, serving, time.Time{}); err != nil {
			log.FromContext(ctx).WithError(err).Error("Failed to update downlink task queue after session key delivery")
		}
	}
	return nil
}
