// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

package middleware

import (
	"fmt"
	"net/http"
	"net/textproto"
	"strings"

	"go.thethings.network/lorawan-stack/v3/pkg/auth"
)

var (
	protocolHeader      = textproto.CanonicalMIMEHeaderKey("Sec-WebSocket-Protocol")
	authorizationHeader = textproto.CanonicalMIMEHeaderKey("Authorization")
	connectionHeader    = textproto.CanonicalMIMEHeaderKey("Connection")
	upgradeHeader       = textproto.CanonicalMIMEHeaderKey("Upgrade")
)

func isWebSocketRequest(r *http.Request) bool {
	h := r.Header
	return strings.EqualFold(h.Get(connectionHeader), "upgrade") &&
		strings.EqualFold(h.Get(upgradeHeader), "websocket")
}

// ProtocolAuthentication returns a middleware that authenticates WebSocket requests using the subprotocol.
// The subprotocol must be prefixed with the given prefix.
// The token is extracted from the subprotocol and used to authenticate the request.
// If the token is valid, the subprotocol is removed from the request.
// If the token is invalid, the request is not authenticated.
func ProtocolAuthentication(prefix string) func(http.Handler) http.Handler {
	prefixLen := len(prefix)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !isWebSocketRequest(r) {
				next.ServeHTTP(w, r)
				return
			}
			if r.Header.Get(authorizationHeader) != "" {
				next.ServeHTTP(w, r)
				return
			}
			protocols := strings.Split(strings.TrimSpace(r.Header.Get(protocolHeader)), ",")
			newProtocols := make([]string, 0, len(protocols))
			token := ""
			for _, protocol := range protocols {
				p := strings.TrimSpace(protocol)
				if len(p) >= prefixLen && strings.EqualFold(prefix, p[:prefixLen]) {
					token = p[prefixLen:]
					continue
				}
				newProtocols = append(newProtocols, p)
			}
			if _, _, _, err := auth.SplitToken(token); err == nil {
				if len(newProtocols) > 0 {
					r.Header.Set(protocolHeader, strings.Join(newProtocols, ","))
				} else {
					r.Header.Del(protocolHeader)
				}
				r.Header.Set(authorizationHeader, fmt.Sprintf("Bearer %s", token))
			}
			next.ServeHTTP(w, r)
		})
	}
}
