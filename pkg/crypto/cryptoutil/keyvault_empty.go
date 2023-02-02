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

package cryptoutil

import (
	"context"
	"crypto/tls"

	"go.thethings.network/lorawan-stack/v3/pkg/crypto"
)

type emptyKeyVault struct{}

// Key implements crypto.KeyVault.
func (emptyKeyVault) Key(_ context.Context, label string) ([]byte, error) {
	return nil, errKeyNotFound.WithAttributes("label", label)
}

// ServerCertificate implements crypto.KeyVault.
func (emptyKeyVault) ServerCertificate(_ context.Context, label string) (tls.Certificate, error) {
	return tls.Certificate{}, errCertificateNotFound.WithAttributes("label", label)
}

// ClientCertificate implements crypto.KeyVault.
func (emptyKeyVault) ClientCertificate(_ context.Context, label string) (tls.Certificate, error) {
	return tls.Certificate{}, errCertificateNotFound.WithAttributes("label", label)
}

// EmptyKeyVault is an empty key vault.
var EmptyKeyVault crypto.KeyVault = emptyKeyVault{}
