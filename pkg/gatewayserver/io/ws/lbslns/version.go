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

package lbslns

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io/ws"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	pfconfig "go.thethings.network/lorawan-stack/v3/pkg/pfconfig/lbslns"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Version contains version information.
// This message is sent by the gateway.
type Version struct {
	Station  string `json:"station"`
	Firmware string `json:"firmware"`
	Package  string `json:"package"`
	Model    string `json:"model"`
	Protocol int    `json:"protocol"`
	Features string `json:"features,omitempty"`
}

// MarshalJSON implements json.Marshaler.
func (v Version) MarshalJSON() ([]byte, error) {
	type Alias Version
	return json.Marshal(struct {
		Type string `json:"msgtype"`
		Alias
	}{
		Type:  TypeUpstreamVersion,
		Alias: Alias(v),
	})
}

// IsProduction checks the features field for "prod" and returns true if found.
// This is then used to set debug options in the router config.
func (v Version) IsProduction() bool {
	return strings.Contains(v.Features, "prod")
}

// GetRouterConfig gets router config for the particular version message.
func (*lbsLNS) GetRouterConfig(
	ctx context.Context,
	msg []byte,
	bandID string,
	fps []*frequencyplans.FrequencyPlan,
	antennaGain int,
	receivedAt time.Time,
) (context.Context, []byte, *ttnpb.GatewayStatus, error) {
	var version Version
	if err := json.Unmarshal(msg, &version); err != nil {
		return ctx, nil, nil, err
	}
	// We attempt to transfer time to all gateways by default.
	// In the future, we should disable time transfers permanently
	// to gateways that signal the presence of a PPS.
	// References https://github.com/lorabasics/basicstation/issues/135.
	ws.UpdateSessionTimeSync(ctx, true)
	cfg, err := pfconfig.GetRouterConfig(ctx, bandID, fps, version, time.Now(), antennaGain)
	if err != nil {
		return ctx, nil, nil, err
	}
	// The SX1301 configuration object should not specify a bandwidth field for the FSK channel.
	// See https://doc.sm.tc/station/tcproto.html#router-config-message under the SX1301CONF section.
	for _, sx1301 := range cfg.SX1301Config {
		if ch := sx1301.FSKChannel; ch != nil {
			ch.Bandwidth = 0
		}
	}
	routerCfg, err := cfg.MarshalJSON()
	if err != nil {
		return ctx, nil, nil, err
	}
	// TODO: Revisit these fields for v3 events (https://github.com/TheThingsNetwork/lorawan-stack/issues/2629)
	stat := &ttnpb.GatewayStatus{
		Time: timestamppb.New(receivedAt),
		Versions: map[string]string{
			"station":  version.Station,
			"firmware": version.Firmware,
			"package":  version.Package,
			"platform": fmt.Sprintf("%s - Firmware %s - Protocol %d", version.Model, version.Firmware, version.Protocol),
		},
		Advanced: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"model": {
					Kind: &structpb.Value_StringValue{
						StringValue: version.Model,
					},
				},
				"features": {
					Kind: &structpb.Value_StringValue{
						StringValue: version.Features,
					},
				},
			},
		},
	}

	return log.NewContextWithFields(ctx, log.Fields(
		"station", version.Station,
		"firmware", version.Firmware,
		"model", version.Model,
	)), routerCfg, stat, nil
}
