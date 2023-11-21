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
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/component"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ratelimit"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/web"
	"golang.org/x/sync/singleflight"
	"google.golang.org/grpc"
)

// Server implements the Basic Station Configuration and Update Server.
type Server struct {
	component *component.Component

	// registry and access can be used to override the default behavior of getting
	// clients from the appropriate cluster peer.
	registry ttnpb.GatewayRegistryClient
	access   ttnpb.GatewayAccessClient
	auth     func(context.Context) grpc.CallOption

	requireExplicitEnable bool
	registerUnknown       bool
	defaultOwner          *ttnpb.OrganizationOrUserIdentifiers
	defaultOwnerAuth      func(context.Context) grpc.CallOption
	defaultLNSURI         string

	allowCUPSURIUpdate bool

	tlsConfig *tls.Config
	trust     *x509.Certificate

	getTrustOnce singleflight.Group
	trustCacheMu sync.RWMutex
	trustCache   map[string]*x509.Certificate

	signers map[uint32]crypto.Signer
}

func (s *Server) getServerAuth(ctx context.Context) grpc.CallOption {
	if s.auth != nil {
		return s.auth(ctx)
	}
	return s.component.WithClusterAuth()
}

func (s *Server) getRegistry(ctx context.Context, ids *ttnpb.GatewayIdentifiers) (ttnpb.GatewayRegistryClient, error) {
	if s.registry != nil {
		return s.registry, nil
	}
	cc, err := s.component.GetPeerConn(ctx, ttnpb.ClusterRole_ENTITY_REGISTRY, nil)
	if err != nil {
		return nil, err
	}
	return ttnpb.NewGatewayRegistryClient(cc), nil
}

func (s *Server) getAccess(ctx context.Context, ids *ttnpb.GatewayIdentifiers) (ttnpb.GatewayAccessClient, error) {
	if s.access != nil {
		return s.access, nil
	}
	cc, err := s.component.GetPeerConn(ctx, ttnpb.ClusterRole_ACCESS, nil)
	if err != nil {
		return nil, err
	}
	return ttnpb.NewGatewayAccessClient(cc), nil
}

// Option configures the CUPSServer.
type Option func(s *Server)

// WithRegisterUnknown configures the CUPS server to register gateways if they
// do not already exist in the registry. The gateways will be registered under the
// given owner.
func WithRegisterUnknown(owner *ttnpb.OrganizationOrUserIdentifiers, auth func(context.Context) grpc.CallOption) Option {
	return func(s *Server) {
		if owner != nil {
			s.registerUnknown, s.defaultOwner, s.defaultOwnerAuth = true, owner, auth
		} else {
			s.registerUnknown, s.defaultOwner, s.defaultOwnerAuth = false, nil, nil
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

// WithDefaultLNSURI configures the CUPS server with a default LNS URI to use when
// no Gateway Server address is registered for a gateway.
func WithDefaultLNSURI(uri string) Option {
	return func(s *Server) {
		s.defaultLNSURI = uri
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

// WithTLSConfig configures the CUPS server with the given TLS config that will
// be used to lookup CUPS/LNS Root CAs.
func WithTLSConfig(tlsConfig *tls.Config) Option {
	return func(s *Server) {
		s.tlsConfig = tlsConfig
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

// WithAuth overrides the CUPS server's server auth func.
func WithAuth(auth func(ctx context.Context) grpc.CallOption) Option {
	return func(s *Server) {
		s.auth = auth
	}
}

// NewServer returns a new CUPS server on top of the given gateway registry
// and gateway access clients.
func NewServer(c *component.Component, options ...Option) *Server {
	s := &Server{
		component:  c,
		signers:    make(map[uint32]crypto.Signer),
		trustCache: make(map[string]*x509.Certificate),
	}
	for _, opt := range options {
		opt(s)
	}
	return s
}

// RegisterRoutes implements web.Registerer
func (s *Server) RegisterRoutes(web *web.Server) {
	router := web.Router().NewRoute().Subrouter()
	router.Use(ratelimit.HTTPMiddleware(s.component.RateLimiter(), "http:gcs:cups"))
	router.Path("/update-info").HandlerFunc(s.UpdateInfo).Methods(http.MethodPost)
}

var errNoTrust = errors.DefineInternal("no_trust", "no trusted certificate found")

// parseAddress parses a CUPS or LNS address.
//
// It supports the typical format "host:port" (port being optional).
// It allows schemes "http://host:port" to be present.
// If schemes http/https/ws/wss are used, the port is inferred if not present.
func parseAddress(defaultScheme, address string) (scheme, host, port string, err error) {
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
	if scheme == "" {
		scheme = defaultScheme
	}
	if port == "" {
		switch scheme {
		case "http":
			port = "80"
		case "ws":
			port = "1887"
		case "https":
			port = "443"
		case "wss":
			port = "8887"
		}
	}
	return
}

func (s *Server) getTrust(address string) (*x509.Certificate, error) {
	if address == "" {
		if s.trust != nil {
			return s.trust, nil
		}
		return nil, errNoTrust.New()
	}
	_, host, port, err := parseAddress("https", address)
	if err != nil {
		return nil, err
	}
	address = net.JoinHostPort(host, port)

	trustI, err, _ := s.getTrustOnce.Do(address, func() (any, error) {
		s.trustCacheMu.RLock()
		trust, ok := s.trustCache[address]
		s.trustCacheMu.RUnlock()
		if ok {
			return trust, nil
		}

		conn, err := tls.DialWithDialer(&net.Dialer{
			Timeout: 5 * time.Second,
		}, "tcp", address, s.tlsConfig)
		if err != nil {
			return nil, err
		}
		defer conn.Close()
		if verifiedChains := conn.ConnectionState().VerifiedChains; len(verifiedChains) > 0 {
			chain := verifiedChains[0]
			trust = chain[len(chain)-1]
		}
		if s.tlsConfig != nil && s.tlsConfig.InsecureSkipVerify {
			chain := conn.ConnectionState().PeerCertificates
			trust = chain[len(chain)-1]
		}

		if trust != nil {
			s.trustCacheMu.Lock()
			s.trustCache[address] = trust
			s.trustCacheMu.Unlock()
			return trust, nil
		}

		return nil, errNoTrust.New()
	})
	if err != nil {
		return nil, err
	}
	return trustI.(*x509.Certificate), nil
}
