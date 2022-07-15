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

package interop

import (
	"context"
	"crypto/tls"
)

func (s *Server) verifySenderCertificate(
	ctx context.Context, senderID string, state *tls.ConnectionState,
) (addrs []string, err error) {
	// TODO: Support reading TLS client certificate from proxy header.
	// (https://github.com/TheThingsNetwork/lorawan-stack/issues/717)
	senderClientCAs, err := s.SenderClientCAs(ctx, senderID)
	if err != nil {
		return nil, err
	}
	for _, chain := range state.VerifiedChains {
		peerCert, clientCA := chain[0], chain[len(chain)-1]
		for _, senderClientCA := range senderClientCAs {
			if clientCA.Equal(senderClientCA) {
				// If the TLS client certificate contains DNS addresses, use those.
				// Otherwise, fallback to using CommonName as address.
				if len(peerCert.DNSNames) > 0 {
					addrs = append([]string(nil), peerCert.DNSNames...)
				} else {
					addrs = []string{peerCert.Subject.CommonName}
				}
				return
			}
		}
	}
	// TODO: Verify state.PeerCertificates[0] with senderClientCAs as Roots
	// and state.PeerCertificates[1:] as Intermediates (https://github.com/TheThingsNetwork/lorawan-stack/issues/718).
	return nil, errUnauthenticated.New()
}
