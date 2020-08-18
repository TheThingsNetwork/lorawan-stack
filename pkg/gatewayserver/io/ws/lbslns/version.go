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

	pbtypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/v3/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	pfconfig "go.thethings.network/lorawan-stack/v3/pkg/pfconfig/lbslns"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
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
	if v.Features == "" {
		return false
	}
	if strings.Contains(v.Features, "prod") {
		return true
	}
	return false
}

// GetRouterConfig gets router config for the particular version message.
func (f *lbsLNS) GetRouterConfig(ctx context.Context, msg []byte, bandID string, fps map[string]*frequencyplans.FrequencyPlan, receivedAt time.Time) (context.Context, []byte, *ttnpb.GatewayStatus, error) {
	var version Version
	if err := json.Unmarshal(msg, &version); err != nil {
		return nil, nil, nil, err
	}
	cfg, err := pfconfig.GetRouterConfig(bandID, fps, version.IsProduction(), time.Now())
	if err != nil {
		return nil, nil, nil, err
	}
	routerCfg, err := cfg.MarshalJSON()
	if err != nil {
		return nil, nil, nil, err
	}
	// TODO: Revisit these fields for v3 events (https://github.com/TheThingsNetwork/lorawan-stack/issues/2629)
	stat := &ttnpb.GatewayStatus{
		Time: receivedAt,
		Versions: map[string]string{
			"station":  version.Station,
			"firmware": version.Firmware,
			"package":  version.Package,
			"platform": fmt.Sprintf("%s - Firmware %s - Protocol %d", version.Model, version.Firmware, version.Protocol),
		},
		Advanced: &pbtypes.Struct{
			Fields: map[string]*pbtypes.Value{
				"model": {
					Kind: &pbtypes.Value_StringValue{
						StringValue: version.Model,
					},
				},
				"features": {
					Kind: &pbtypes.Value_StringValue{
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
