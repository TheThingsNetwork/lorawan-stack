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

package webmiddleware

import (
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strings"
)

const (
	headerForwarded = "Forwarded"

	headerXForwardedFor   = "X-Forwarded-For"
	headerXForwardedHost  = "X-Forwarded-Host"
	headerXForwardedProto = "X-Forwarded-Proto" // We don't support non-standard headers such as Front-End-Https, X-Forwarded-Ssl, X-Url-Scheme.
	headerXRealIP         = "X-Real-IP"

	headerXForwardedClientCert        = "X-Forwarded-Client-Cert"          // Envoy mTLS.
	headerXForwardedTLSClientCert     = "X-Forwarded-Tls-Client-Cert"      // Traefik mTLS.
	headerXForwardedTLSClientCertInfo = "X-Forwarded-Tls-Client-Cert-Info" // Traefik mTLS.

	schemeHTTP  = "http"
	schemeHTTPS = "https"
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

// ProxyConfiguration is the configuration for the ProxyHeaders middleware.
type ProxyConfiguration struct {
	Trusted []*net.IPNet
}

func (c ProxyConfiguration) trustedIP(ip net.IP) bool {
	for _, ipNet := range c.Trusted {
		if ipNet.Contains(ip) {
			return true
		}
	}
	return false
}

// ParseAndAddTrusted parses a list of CIDRs and adds them to the list of trusted ranges.
func (c *ProxyConfiguration) ParseAndAddTrusted(cidrs ...string) error {
	for _, cidr := range cidrs {
		_, net, err := net.ParseCIDR(cidr)
		if err != nil {
			return err
		}
		c.Trusted = append(c.Trusted, net)
	}
	return nil
}

// ProxyHeaders processes proxy headers for trusted proxies.
func ProxyHeaders(config ProxyConfiguration) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			remoteIP, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				// The *http.Server should have set r.RemoteAddr to "IP:port".
				panic(fmt.Errorf("invalid RemoteAddr %q in *http.Request: %w", r.RemoteAddr, err))
			}
			if config.trustedIP(net.ParseIP(remoteIP)) {
				// We trust the proxy, so we parse the headers if present.
				forwardedFor, forwardedScheme, forwardedHost := parseForwardedHeaders(r.Header)
				if forwardedFor != "" {
					r.Header.Set(headerXRealIP, forwardedFor)
				}
				if forwardedScheme != "" {
					r.URL.Scheme = forwardedScheme
				}
				if forwardedHost != "" {
					r.URL.Host = forwardedHost
				}
			} else {
				// We don't trust the proxy, remove its headers.
				for _, header := range proxyHeaders {
					r.Header.Del(header)
				}
				r.Header.Set(headerXRealIP, remoteIP)
			}
			next.ServeHTTP(w, r)
		})
	}
}

func parseForwardedHeaders(h http.Header) (forwardedFor, forwardedScheme, forwardedHost string) {
	if xForwardedProto := h.Get(headerXForwardedProto); xForwardedProto != "" {
		forwardedScheme = xForwardedProto
	}
	if xForwardedHost := h.Get(headerXForwardedHost); xForwardedHost != "" {
		forwardedHost = xForwardedHost
	}
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
