// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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
	"net"
	"net/http"

	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
)

// Peer sets the remote address as a peer in the request context.
func Peer() MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			p := new(peer.Peer)
			if addr, err := net.ResolveTCPAddr("tcp", r.RemoteAddr); err == nil {
				p.Addr = addr
			}
			if r.TLS != nil {
				p.AuthInfo = credentials.TLSInfo{
					State: *r.TLS,
					CommonAuthInfo: credentials.CommonAuthInfo{
						SecurityLevel: credentials.PrivacyAndIntegrity,
					},
				}
			}
			ctx := peer.NewContext(r.Context(), p)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
