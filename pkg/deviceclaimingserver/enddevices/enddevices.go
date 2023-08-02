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

// Package enddevices provides functions to configure End Device claiming clients.
package enddevices

import (
	"context"
	"path/filepath"
	"strings"

	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto"
	"go.thethings.network/lorawan-stack/v3/pkg/deviceclaimingserver/enddevices/ttjsv2"
	"go.thethings.network/lorawan-stack/v3/pkg/fetch"
	"go.thethings.network/lorawan-stack/v3/pkg/httpclient"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v2"
)

// EndDeviceClaimer provides methods for Claiming End Devices on (external) Join Server.
type EndDeviceClaimer interface {
	// SupportsJoinEUI returns whether the Join Server supports this JoinEUI.
	SupportsJoinEUI(joinEUI types.EUI64) bool
	// Claim claims an End Device.
	Claim(ctx context.Context, joinEUI, devEUI types.EUI64, claimAuthenticationCode string) error
	// GetClaimStatus returns the claim status an End Device.
	GetClaimStatus(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers) (*ttnpb.GetClaimStatusResponse, error)
	// Unclaim releases the claim on an End Device.
	Unclaim(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers) (err error)

	// BatchUnclaim release the claim on a batch of End Devices.
	BatchUnclaim(ctx context.Context, ids []*ttnpb.EndDeviceIdentifiers) (*ttnpb.BatchUnclaimEndDevicesResponse, error)
}

// Component abstracts the underlying *component.Component.
type Component interface {
	httpclient.Provider
	KeyService() crypto.KeyService
	GetBaseConfig(ctx context.Context) config.ServiceBase
	GetPeerConn(ctx context.Context, role ttnpb.ClusterRole, ids cluster.EntityIdentifiers) (*grpc.ClientConn, error)
	AllowInsecureForCredentials() bool
}

const (
	ttjsV2Type = "ttjsv2"
)

// Upstream abstracts EndDeviceClaimingServer.
type Upstream struct {
	claimers map[string]EndDeviceClaimer
}

// NewUpstream returns a new Upstream.
func NewUpstream(ctx context.Context, c Component, conf Config, opts ...Option) (*Upstream, error) {
	upstream := &Upstream{
		claimers: make(map[string]EndDeviceClaimer),
	}
	for _, opt := range opts {
		opt(upstream)
	}

	fetcher, err := conf.Fetcher(ctx, c.GetBaseConfig(ctx).Blob, c)
	if err != nil {
		return nil, err
	}
	if fetcher == nil {
		return upstream, nil
	}
	baseConfigBytes, err := fetcher.File(JSClientConfigurationName)
	if err != nil {
		return nil, err
	}
	var baseConfig baseConfig
	if err := yaml.UnmarshalStrict(baseConfigBytes, &baseConfig); err != nil {
		return nil, err
	}

	nsID := conf.NSID
	// TODO: Remove fallback logic (https://github.com/TheThingsNetwork/lorawan-stack/issues/6048)
	if nsID == nil {
		nsID = conf.NetworkServer.HomeNSID
	}

	// Setup upstreams.
	for _, js := range baseConfig.JoinServers {
		// Fetch and parse configuration.
		fileParts := strings.Split(filepath.ToSlash(js.File), "/")
		fetcher := fetch.WithBasePath(fetcher, fileParts[:len(fileParts)-1]...)
		fileName := fileParts[len(fileParts)-1]
		configBytes, err := fetcher.File(fileName)
		if err != nil {
			return nil, err
		}

		var claimer EndDeviceClaimer
		switch js.Type {
		case ttjsV2Type:
			var ttjsConf ttjsv2.ConfigFile
			if err := yaml.UnmarshalStrict(configBytes, &ttjsConf); err != nil {
				return nil, err
			}
			claimer = ttjsv2.NewClient(c, fetcher, ttjsv2.Config{
				NetID:           conf.NetID,
				NSID:            nsID,
				ASID:            conf.ASID,
				JoinEUIPrefixes: js.JoinEUIs,
				ConfigFile:      ttjsConf,
			})
		default:
			log.FromContext(ctx).WithField("type", js.Type).Warn("Unknown Join Server type")
			continue
		}

		// The file for each client will be unique.
		clientName := strings.Trim(fileName, filepath.Ext(fileName))
		upstream.claimers[clientName] = claimer
	}

	return upstream, nil
}

// Option configures Upstream.
type Option func(*Upstream)

// WithClaimer adds a claimer to Upstream.
func WithClaimer(name string, claimer EndDeviceClaimer) Option {
	return func(upstream *Upstream) {
		upstream.claimers[name] = claimer
	}
}

// JoinEUIClaimer returns the EndDeviceClaimer for the given JoinEUI.
func (upstream *Upstream) JoinEUIClaimer(_ context.Context, joinEUI types.EUI64) EndDeviceClaimer {
	for _, claimer := range upstream.claimers {
		if claimer.SupportsJoinEUI(joinEUI) {
			return claimer
		}
	}
	return nil
}
