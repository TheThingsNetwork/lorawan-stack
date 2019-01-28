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

package basicstation

import (
	"context"
	"crypto"
	"crypto/tls"
	"crypto/x509"
	"net"
	"net/url"
	"strings"

	"github.com/labstack/echo"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/web"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// CUPSServer implements the Basic Station Configuration and Update Server.
type CUPSServer struct {
	gatewayRegistry ttnpb.GatewayRegistryClient
	gatewayAccess   ttnpb.GatewayAccessClient
	fallbackAuth    grpc.CallOption

	requireExplicitEnable bool
	registerUnknown       bool
	defaultOwner          ttnpb.OrganizationOrUserIdentifiers

	rootCAs *x509.CertPool
	trust   *x509.Certificate

	signers map[uint32]crypto.Signer
}

// Option configures the CUPSServer.
type Option func(s *CUPSServer)

// WithFallbackAuth sets fallback auth for gateways that don't provide TTN auth.
// When this auth method is used, the CUPS server will look up the _cups_credentials
// attribute in the gateway registry.
func WithFallbackAuth(auth grpc.CallOption) Option {
	return func(s *CUPSServer) {
		s.fallbackAuth = auth
	}
}

// WithExplicitEnable requires CUPS to be explicitly enabled with a _cups attribute
// in the gateway registry.
func WithExplicitEnable(enable bool) Option {
	return func(s *CUPSServer) {
		s.requireExplicitEnable = enable
	}
}

// WithRegisterUnknown configures the CUPS server to register gateways if they
// do not already exist in the registry. The gateways will be registered under the
// given owner.
func WithRegisterUnknown(owner *ttnpb.OrganizationOrUserIdentifiers) Option {
	return func(s *CUPSServer) {
		if owner != nil {
			s.registerUnknown, s.defaultOwner = true, *owner
		} else {
			s.registerUnknown, s.defaultOwner = false, ttnpb.OrganizationOrUserIdentifiers{}
		}
	}
}

// WithTrust configures the CUPS server to return the given certificate to gateways
// as trusted certificate for the CUPS server. This should typically be the certificate
// of the Root CA in the chain of the CUPS server's TLS certificate.
func WithTrust(cert *x509.Certificate) Option {
	return func(s *CUPSServer) {
		s.trust = cert
	}
}

// WithRootCAs configures the CUPS server with the given Root CAs that will be used
// to lookup CUPS/LNS Root CAs.
func WithRootCAs(pool *x509.CertPool) Option {
	return func(s *CUPSServer) {
		s.rootCAs = pool
	}
}

// WithSigner configures the CUPS server with a firmware signer.
func WithSigner(keyCRC uint32, signer crypto.Signer) Option {
	return func(s *CUPSServer) {
		s.signers[keyCRC] = signer
	}
}

// NewCUPSServer returns a new CUPS server on top of the given gateway registry
// and gateway access clients.
func NewCUPSServer(gr ttnpb.GatewayRegistryClient, ga ttnpb.GatewayAccessClient, options ...Option) *CUPSServer {
	s := &CUPSServer{
		gatewayRegistry: gr,
		gatewayAccess:   ga,
		signers:         make(map[uint32]crypto.Signer),
	}
	for _, opt := range options {
		opt(s)
	}
	return s
}

// RegisterRoutes implements web.Registerer
func (s *CUPSServer) RegisterRoutes(web *web.Server) {
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

func (s *CUPSServer) getTrust(address string) (*x509.Certificate, error) {
	if address == "" {
		if s.trust == nil {
			return nil, errNoTrust
		}
		return s.trust, nil
	}
	if strings.Contains(address, "//") {
		url, err := url.Parse(address)
		if err != nil {
			return nil, err
		}
		address = url.Host
	}
	if _, _, err := net.SplitHostPort(address); err != nil {
		address = net.JoinHostPort(address, "443")
	}
	conn, err := tls.Dial("tcp", address, &tls.Config{RootCAs: s.rootCAs})
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	certChain := conn.ConnectionState().VerifiedChains[0]
	return certChain[len(certChain)-1], nil
}
