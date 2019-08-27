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
	"crypto"
	"crypto/tls"
	"crypto/x509"
	"net"
	"net/url"
	"strings"

	echo "github.com/labstack/echo/v4"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/web"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// ServerConfig is the configuration of the CUPS server.
type ServerConfig struct {
	ExplicitEnable  bool `name:"require-explicit-enable" description:"Require gateways to explicitly enable CUPS"`
	RegisterUnknown struct {
		Type   string `name:"account-type" description:"Type of account to register unknown gateways to (user|organization)"`
		ID     string `name:"id" description:"ID of the account to register unknown gateways to"`
		APIKey string `name:"api-key" description:"API Key to use for unknown gateway registration"`
	} `name:"owner-for-unknown"`
	AllowCUPSURIUpdate bool `name:"allow-cups-uri-update" description:"Allow CUPS URI updates"`
}

// NewServer returns a new CUPS server from this config on top of the component.
func (conf ServerConfig) NewServer(c *component.Component, customOpts ...Option) *Server {
	var registerUnknownTo *ttnpb.OrganizationOrUserIdentifiers
	switch conf.RegisterUnknown.Type {
	case "user":
		registerUnknownTo = ttnpb.UserIdentifiers{UserID: conf.RegisterUnknown.ID}.OrganizationOrUserIdentifiers()
	case "organization":
		registerUnknownTo = ttnpb.OrganizationIdentifiers{OrganizationID: conf.RegisterUnknown.ID}.OrganizationOrUserIdentifiers()
	}
	opts := []Option{
		WithExplicitEnable(conf.ExplicitEnable),
		WithRegisterUnknown(registerUnknownTo),
		WithAllowCUPSURIUpdate(conf.AllowCUPSURIUpdate),
	}
	if conf.RegisterUnknown.APIKey != "" {
		opts = append(opts, WithAuth(func(ctx context.Context, gatewayEUI types.EUI64, auth string) grpc.CallOption {
			return grpc.PerRPCCredentials(rpcmetadata.MD{
				AuthType:      "bearer",
				AuthValue:     conf.RegisterUnknown.APIKey,
				AllowInsecure: c.AllowInsecureForCredentials(),
			})
		}))
	}
	if tlsConfig, err := c.GetTLSConfig(c.Context()); err == nil {
		opts = append(opts, WithRootCAs(tlsConfig.RootCAs))
	}
	s := NewServer(c, append(opts, customOpts...)...)
	c.RegisterWeb(s)
	return s
}

// Server implements the Basic Station Configuration and Update Server.
type Server struct {
	component *component.Component

	// registry and access can be used to override the default behavior of getting
	// clients from the appropriate cluster peer.
	registry ttnpb.GatewayRegistryClient
	access   ttnpb.GatewayAccessClient

	auth func(context.Context, types.EUI64, string) grpc.CallOption

	requireExplicitEnable bool
	registerUnknown       bool
	defaultOwner          ttnpb.OrganizationOrUserIdentifiers

	allowCUPSURIUpdate bool

	rootCAs *x509.CertPool
	trust   *x509.Certificate

	signers map[uint32]crypto.Signer
}

func (s *Server) getAuth(ctx context.Context, eui types.EUI64, auth string) grpc.CallOption {
	if s.auth != nil {
		return s.auth(ctx, eui, auth)
	}
	return s.component.WithClusterAuth()
}

var errESUnavailable = errors.DefineUnavailable("entity_registry_unavailable", "Entity Registry unavailable for gateway_id `{gateway_id}`")

func (s *Server) getRegistry(ctx context.Context, ids *ttnpb.GatewayIdentifiers) (ttnpb.GatewayRegistryClient, error) {
	if s.registry != nil {
		return s.registry, nil
	}
	if peer := s.component.GetPeer(ctx, ttnpb.ClusterRole_ENTITY_REGISTRY, ids); peer != nil {
		return ttnpb.NewGatewayRegistryClient(peer.Conn()), nil
	}
	return nil, errESUnavailable.WithAttributes("gateway_id", ids.GetGatewayID())
}

var errAUnavailable = errors.DefineUnavailable("access_unavailable", "Access unavailable for gateway_id `{gateway_id}`")

func (s *Server) getAccess(ctx context.Context, ids *ttnpb.GatewayIdentifiers) (ttnpb.GatewayAccessClient, error) {
	if s.access != nil {
		return s.access, nil
	}
	if peer := s.component.GetPeer(ctx, ttnpb.ClusterRole_ACCESS, ids); peer != nil {
		return ttnpb.NewGatewayAccessClient(peer.Conn()), nil
	}
	return nil, errAUnavailable.WithAttributes("gateway_id", ids.GetGatewayID())
}

// Option configures the CUPSServer.
type Option func(s *Server)

// WithAuth sets the auth function for gateways that don't provide TTN auth.
// When this auth method is used, the CUPS server will look up the cups-credentials
// attribute in the gateway registry.
func WithAuth(auth func(ctx context.Context, gatewayEUI types.EUI64, auth string) grpc.CallOption) Option {
	return func(s *Server) {
		s.auth = auth
	}
}

// WithExplicitEnable requires CUPS to be explicitly enabled with a cups attribute
// in the gateway registry.
func WithExplicitEnable(enable bool) Option {
	return func(s *Server) {
		s.requireExplicitEnable = enable
	}
}

// WithRegisterUnknown configures the CUPS server to register gateways if they
// do not already exist in the registry. The gateways will be registered under the
// given owner.
func WithRegisterUnknown(owner *ttnpb.OrganizationOrUserIdentifiers) Option {
	return func(s *Server) {
		if owner != nil {
			s.registerUnknown, s.defaultOwner = true, *owner
		} else {
			s.registerUnknown, s.defaultOwner = false, ttnpb.OrganizationOrUserIdentifiers{}
		}
	}
}

// WithAllowCUPSURIUpdate configures the CUPS server to allow updates of the CUPS
// Server URI.
func WithAllowCUPSURIUpdate(allow bool) Option {
	return func(s *Server) {
		s.allowCUPSURIUpdate = allow
	}
}

// WithTrust configures the CUPS server to return the given certificate to gateways
// as trusted certificate for the CUPS server. This should typically be the certificate
// of the Root CA in the chain of the CUPS server's TLS certificate.
func WithTrust(cert *x509.Certificate) Option {
	return func(s *Server) {
		s.trust = cert
	}
}

// WithRootCAs configures the CUPS server with the given Root CAs that will be used
// to lookup CUPS/LNS Root CAs.
func WithRootCAs(pool *x509.CertPool) Option {
	return func(s *Server) {
		s.rootCAs = pool
	}
}

// WithSigner configures the CUPS server with a firmware signer.
func WithSigner(keyCRC uint32, signer crypto.Signer) Option {
	return func(s *Server) {
		s.signers[keyCRC] = signer
	}
}

// WithRegistries overrides the CUPS server's gateway registries.
func WithRegistries(registry ttnpb.GatewayRegistryClient, access ttnpb.GatewayAccessClient) Option {
	return func(s *Server) {
		s.registry, s.access = registry, access
	}
}

// NewServer returns a new CUPS server on top of the given gateway registry
// and gateway access clients.
func NewServer(c *component.Component, options ...Option) *Server {
	s := &Server{
		component: c,
		signers:   make(map[uint32]crypto.Signer),
	}
	for _, opt := range options {
		opt(s)
	}
	return s
}

// RegisterRoutes implements web.Registerer
func (s *Server) RegisterRoutes(web *web.Server) {
	web.POST("/update-info", s.UpdateInfo)
}

func getContext(c echo.Context) context.Context {
	ctx := c.Request().Context()
	md := metadata.New(map[string]string{
		"authorization": c.Request().Header.Get(echo.HeaderAuthorization),
	})
	if ctxMd, ok := metadata.FromIncomingContext(ctx); ok {
		md = metadata.Join(ctxMd, md)
	}
	return metadata.NewIncomingContext(ctx, md)
}

var errNoTrust = errors.DefineInternal("no_trust", "no trusted certificate configured")

// parseAddress parses a CUPS or LNS address.
//
// It supports the typical format "host:port" (port being optional).
// It allows schemes "http://host:port" to be present.
// If schemes http/https/ws/wss are used, the port is inferred if not present.
func parseAddress(address string) (scheme, host, port string, err error) {
	if address == "" {
		return
	}
	if strings.Contains(address, "://") {
		url, err := url.Parse(address)
		if err != nil {
			return "", "", "", err
		}
		scheme, address = url.Scheme, url.Host
	}
	if strings.Contains(address, ":") {
		host, port, err = net.SplitHostPort(address)
		if err != nil {
			host = address
			err = nil
		}
	} else {
		host = address
	}
	if port == "" {
		switch scheme {
		case "http", "ws":
			port = "80"
		case "https", "wss":
			port = "443"
		}
	}
	return
}

func (s *Server) getTrust(address string) (*x509.Certificate, error) {
	if address == "" {
		if s.trust == nil {
			return nil, errNoTrust
		}
		return s.trust, nil
	}
	_, host, port, err := parseAddress(address)
	if err != nil {
		return nil, err
	}
	if port == "" {
		port = "443"
	}
	conn, err := tls.Dial("tcp", net.JoinHostPort(host, port), &tls.Config{RootCAs: s.rootCAs})
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	certChain := conn.ConnectionState().VerifiedChains[0]
	return certChain[len(certChain)-1], nil
}
