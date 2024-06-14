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

package rpcmiddleware

import (
	"context"
	"fmt"
	"net"
	"regexp"
	"strings"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	mtlsauth "go.thethings.network/lorawan-stack/v3/pkg/auth/mtls"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

const (
	headerForwarded = "forwarded"

	headerXForwardedFor   = "x-forwarded-for"
	headerXForwardedHost  = "x-forwarded-host"
	headerXForwardedProto = "x-forwarded-proto"
	headerXRealIP         = "x-real-ip"
	// We don't support non-standard headers such as Front-End-Https, X-Forwarded-Ssl, X-Url-Scheme.

	headerXForwardedClientCert        = "x-forwarded-client-cert"          // Envoy mTLS.
	headerXForwardedTLSClientCert     = "x-forwarded-tls-client-cert"      // Traefik mTLS.
	headerXForwardedTLSClientCertInfo = "x-forwarded-tls-client-cert-info" // Traefik mTLS.
)

var (
	proxyHeaders = []string{
		headerForwarded,
		headerXForwardedFor, headerXForwardedHost, headerXForwardedProto,
		headerXForwardedClientCert,
		headerXForwardedTLSClientCert, headerXForwardedTLSClientCertInfo,
		headerXRealIP,
	}
	forwardedForRegex   = regexp.MustCompile(`(?i)(?:for=)([^(;|,| )]+)`)
	forwardedHostRegex  = regexp.MustCompile(`(?i)(?:host=)([^(;|,| )]+)`)
	forwardedProtoRegex = regexp.MustCompile(`(?i)(?:proto=)(https|http)`)
)

// ProxyHeaders is the configuration for the ProxyHeaders middleware.
type ProxyHeaders struct {
	Trusted []*net.IPNet
}

func (h *ProxyHeaders) trustedIP(ip net.IP) bool {
	for _, ipNet := range h.Trusted {
		if ipNet.Contains(ip) {
			return true
		}
	}
	return false
}

// ParseAndAddTrusted parses a list of CIDRs and adds them to the list of trusted ranges.
func (h *ProxyHeaders) ParseAndAddTrusted(cidrs ...string) error {
	for _, cidr := range cidrs {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			return err
		}
		h.Trusted = append(h.Trusted, ipNet)
	}
	return nil
}

// UnaryServerInterceptor is the interceptor for unary RPCs.
func (h *ProxyHeaders) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		ctx, md := h.intercept(ctx)
		ctx = metadata.NewIncomingContext(ctx, md)
		return handler(ctx, req)
	}
}

// StreamServerInterceptor is the interceptor for streaming RPCs.
func (h *ProxyHeaders) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := stream.Context()
		wrapped := grpc_middleware.WrapServerStream(stream)
		ctx, md := h.intercept(ctx)
		wrapped.WrappedContext = metadata.NewIncomingContext(ctx, md)
		return handler(srv, wrapped)
	}
}

func (h *ProxyHeaders) intercept(ctx context.Context) (context.Context, metadata.MD) {
	md, _ := metadata.FromIncomingContext(ctx)

	p, ok := peer.FromContext(ctx)
	if !ok {
		// The gRPC server should always set this.
		panic(fmt.Errorf("no peer in gRPC context"))
	}
	remoteAddr := p.Addr.String()
	if remoteAddr == "pipe" {
		remoteAddr = "127.0.0.0:0"
	}
	remoteIP, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		// Either the gRPC server should have set the peer address to "IP:port"
		// or we have converted the pipe address to localhost above.
		panic(fmt.Errorf("invalid peer %q in gRPC context: %w", remoteAddr, err))
	}
	if h.trustedIP(net.ParseIP(remoteIP)) {
		// We trust the proxy, so we parse the headers if present.
		forwardedFor, _, _ := parseForwardedHeaders(getLastFromMD(md)) // ignore forwardedScheme and forwardedHost.
		if forwardedFor != "" {
			md.Set(headerXRealIP, strings.TrimSpace(strings.Split(forwardedFor, ",")[0]))
		}
		if cert, ok, err := mtlsauth.FromProxyHeaders(getLastFromMD(md)); err != nil {
			log.FromContext(ctx).WithError(err).Warn("Failed to parse client certificate from proxy headers")
		} else if ok {
			ctx = mtlsauth.NewContextWithClientCertificate(ctx, cert)
		}
	} else {
		// We don't trust the proxy, remove its headers.
		for _, header := range proxyHeaders {
			delete(md, header)
		}
		md.Set(headerXRealIP, remoteIP)
	}
	return ctx, md
}

type getLastFromMD metadata.MD

func (m getLastFromMD) Get(key string) string {
	if values, ok := m[key]; ok {
		return values[len(values)-1]
	}
	return ""
}

func parseForwardedHeaders(h getLastFromMD) (forwardedFor, forwardedScheme, forwardedHost string) {
	forwardedFor = h.Get(headerXForwardedFor)
	forwardedScheme = h.Get(headerXForwardedProto)
	forwardedHost = h.Get(headerXForwardedHost)
	if forwarded := h.Get(headerForwarded); forwarded != "" {
		if match := forwardedForRegex.FindStringSubmatch(forwarded); len(match) > 1 {
			forwardedFor = strings.ToLower(match[1])
		}
		if match := forwardedProtoRegex.FindStringSubmatch(forwarded); len(match) > 1 {
			forwardedScheme = strings.ToLower(match[1])
		}
		if match := forwardedHostRegex.FindStringSubmatch(forwarded); len(match) > 1 {
			forwardedHost = strings.ToLower(match[1])
		}
	}
	return forwardedFor, forwardedScheme, forwardedHost
}
