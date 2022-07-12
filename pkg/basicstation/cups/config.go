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

package cups

import (
	"context"

	"go.thethings.network/lorawan-stack/v3/pkg/component"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/grpc"
)

// ServerConfig is the configuration of the CUPS server.
type ServerConfig struct {
	ExplicitEnable  bool `name:"require-explicit-enable" description:"Require gateways to explicitly enable CUPS. This option is ineffective"`
	RegisterUnknown struct {
		Type   string `name:"account-type" description:"Type of account to register unknown gateways to (user|organization)"`
		ID     string `name:"id" description:"ID of the account to register unknown gateways to"`
		APIKey string `name:"api-key" description:"API Key to use for unknown gateway registration"`
	} `name:"owner-for-unknown"`
	Default struct {
		LNSURI string `name:"lns-uri" description:"The default LNS URI that the gateways should use"`
	} `name:"default" description:"Default gateway settings"`
	AllowCUPSURIUpdate bool `name:"allow-cups-uri-update" description:"Allow CUPS URI updates"`
}

// NewServer returns a new CUPS server from this config on top of the component.
func (conf ServerConfig) NewServer(c *component.Component, customOpts ...Option) *Server {
	opts := []Option{
		WithAllowCUPSURIUpdate(conf.AllowCUPSURIUpdate),
		WithDefaultLNSURI(conf.Default.LNSURI),
	}
	var registerUnknownTo *ttnpb.OrganizationOrUserIdentifiers
	switch conf.RegisterUnknown.Type {
	case "user":
		registerUnknownTo = (&ttnpb.UserIdentifiers{
			UserId: conf.RegisterUnknown.ID,
		}).GetOrganizationOrUserIdentifiers()
	case "organization":
		registerUnknownTo = (&ttnpb.OrganizationIdentifiers{
			OrganizationId: conf.RegisterUnknown.ID,
		}).GetOrganizationOrUserIdentifiers()
	}
	if registerUnknownTo != nil && conf.RegisterUnknown.APIKey != "" {
		opts = append(opts,
			WithRegisterUnknown(registerUnknownTo, func(ctx context.Context) grpc.CallOption {
				return grpc.PerRPCCredentials(rpcmetadata.MD{
					AuthType:      "bearer",
					AuthValue:     conf.RegisterUnknown.APIKey,
					AllowInsecure: c.AllowInsecureForCredentials(),
				})
			}),
		)
	}
	// The Server.tlsConfig is used when dialing a CUPS or an LNS server to query its certificate chain.
	// When dialing servers with self-signed certs, the Root CA of target server must either be trusted by the system or added explicitly via the `--tls.root-ca` option.
	if tlsConfig, err := c.GetTLSClientConfig(c.Context()); err == nil {
		opts = append(opts, WithTLSConfig(tlsConfig))
	}
	s := NewServer(c, append(opts, customOpts...)...)
	c.RegisterWeb(s)
	return s
}
