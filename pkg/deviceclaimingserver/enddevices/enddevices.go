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

	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/deviceclaimingserver/enddevices/ttjsv2"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/fetch"
	"go.thethings.network/lorawan-stack/v3/pkg/httpclient"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
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
}

// Component abstracts the underlying *component.Component.
type Component interface {
	httpclient.Provider
	GetBaseConfig(ctx context.Context) config.ServiceBase
	GetPeerConn(ctx context.Context, role ttnpb.ClusterRole, ids cluster.EntityIdentifiers) (*grpc.ClientConn, error)
	AllowInsecureForCredentials() bool
}

const (
	ttjsV2Type = "ttjsv2"
)

// Upstream abstracts EndDeviceClaimingServer.
type Upstream struct {
	Component
	deviceRegistry ttnpb.EndDeviceRegistryClient
	servers        map[string]EndDeviceClaimer
}

// NewUpstream returns a new Upstream.
func NewUpstream(ctx context.Context, conf Config, c Component, opts ...Option) (*Upstream, error) {
	upstream := &Upstream{
		Component: c,
		servers:   make(map[string]EndDeviceClaimer),
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
			var cfg ttjsv2.Config
			if err := yaml.UnmarshalStrict(configBytes, &cfg); err != nil {
				return nil, err
			}
			cfg.NetID = conf.NetID
			cfg.JoinEUIPrefixes = js.JoinEUIs
			cfg.NetworkServer.Hostname = conf.NetworkServer.Hostname
			cfg.NetworkServer.HomeNSID = conf.NetworkServer.HomeNSID
			claimer, err = cfg.NewClient(ctx, c)
			if err != nil {
				return nil, err
			}
		default:
			log.FromContext(ctx).WithField("type", js.Type).Warn("Unknown Join Server type")
			continue
		}

		// The file for each client will be unique.
		clientName := strings.Trim(fileName, filepath.Ext(fileName))
		upstream.servers[clientName] = claimer
	}

	for _, opt := range opts {
		opt(upstream)
	}

	return upstream, nil
}

// Option configures Upstream.
type Option func(*Upstream)

// WithDeviceRegistry overrides the device registry of the Upstream.
func WithDeviceRegistry(reg ttnpb.EndDeviceRegistryClient) Option {
	return func(upstream *Upstream) {
		upstream.deviceRegistry = reg
	}
}

var (
	errNoEUI                = errors.DefineInvalidArgument("no_eui", "DevEUI/JoinEUI not found in request")
	errClaimingNotSupported = errors.DefineAborted("claiming_not_supported", "claiming not supported for JoinEUI `{eui}`")
)

func (upstream *Upstream) joinEUIClaimer(ctx context.Context, joinEUI types.EUI64) EndDeviceClaimer {
	for _, srv := range upstream.servers {
		if srv.SupportsJoinEUI(joinEUI) {
			return srv
		}
	}
	return nil
}

// Claim implements EndDeviceClaimingServer.
func (upstream *Upstream) Claim(
	ctx context.Context, joinEUI, devEUI types.EUI64, claimAuthenticationCode string,
) error {
	claimer := upstream.joinEUIClaimer(ctx, joinEUI)
	if claimer == nil {
		return errClaimingNotSupported.WithAttributes("eui", joinEUI)
	}
	return claimer.Claim(ctx, joinEUI, devEUI, claimAuthenticationCode)
}

// Unclaim implements EndDeviceClaimingServer.
func (upstream *Upstream) Unclaim(ctx context.Context, in *ttnpb.EndDeviceIdentifiers) (*emptypb.Empty, error) {
	if in.DevEui == nil || in.JoinEui == nil {
		return nil, errNoEUI.New()
	}
	err := upstream.requireRights(ctx, in, &ttnpb.Rights{
		Rights: []ttnpb.Right{
			ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE,
		},
	})
	if err != nil {
		return nil, err
	}
	claimer := upstream.joinEUIClaimer(ctx, types.MustEUI64(in.JoinEui).OrZero())
	if claimer == nil {
		return nil, errClaimingNotSupported.WithAttributes("eui", in.JoinEui)
	}
	err = claimer.Unclaim(ctx, in)
	if err != nil {
		return nil, err
	}
	return ttnpb.Empty, nil
}

// GetInfoByJoinEUI implements EndDeviceClaimingServer.
func (upstream *Upstream) GetInfoByJoinEUI(
	ctx context.Context, in *ttnpb.GetInfoByJoinEUIRequest,
) (*ttnpb.GetInfoByJoinEUIResponse, error) {
	joinEUI := types.MustEUI64(in.JoinEui).OrZero()
	claimer := upstream.joinEUIClaimer(ctx, joinEUI)
	return &ttnpb.GetInfoByJoinEUIResponse{
		JoinEui:          joinEUI.Bytes(),
		SupportsClaiming: (claimer != nil),
	}, nil
}

// GetClaimStatus implements EndDeviceClaimingServer.
func (upstream *Upstream) GetClaimStatus(
	ctx context.Context, in *ttnpb.EndDeviceIdentifiers,
) (*ttnpb.GetClaimStatusResponse, error) {
	if in.DevEui == nil || in.JoinEui == nil {
		return nil, errNoEUI.New()
	}
	err := upstream.requireRights(ctx, in, &ttnpb.Rights{
		Rights: []ttnpb.Right{
			ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ,
		},
	})
	if err != nil {
		return nil, err
	}
	claimer := upstream.joinEUIClaimer(ctx, types.MustEUI64(in.JoinEui).OrZero())
	if claimer == nil {
		return nil, errClaimingNotSupported.WithAttributes("eui", in.JoinEui)
	}
	return claimer.GetClaimStatus(ctx, in)
}

func (upstream *Upstream) requireRights(
	ctx context.Context, in *ttnpb.EndDeviceIdentifiers, appRights *ttnpb.Rights,
) error {
	// Collaborator must have the required rights on the application.
	if err := rights.RequireApplication(ctx, in.ApplicationIds,
		appRights.Rights...,
	); err != nil {
		return err
	}
	// Check that the device actually exists in the application.
	// If the EUIs are set in the request, the IS also checks that they match the stored device.
	callOpt, err := rpcmetadata.WithForwardedAuth(ctx, upstream.Component.AllowInsecureForCredentials())
	if err != nil {
		return err
	}
	er, err := upstream.getDeviceRegistry(ctx)
	if err != nil {
		return err
	}
	_, err = er.Get(ctx, &ttnpb.GetEndDeviceRequest{
		EndDeviceIds: in,
	}, callOpt)
	return err
}

func (upstream *Upstream) getDeviceRegistry(ctx context.Context) (ttnpb.EndDeviceRegistryClient, error) {
	if upstream.deviceRegistry != nil {
		return upstream.deviceRegistry, nil
	}
	conn, err := upstream.Component.GetPeerConn(ctx, ttnpb.ClusterRole_ENTITY_REGISTRY, nil)
	if err != nil {
		return nil, err
	}
	return ttnpb.NewEndDeviceRegistryClient(conn), nil
}
