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

package cryptoutil

import (
	"context"
	"crypto/tls"
	"crypto/x509"

	"go.thethings.network/lorawan-stack/pkg/crypto"
)

type emptyKeyVault struct {
	ComponentPrefixKEKLabeler
}

// EmptyKeyVault is an empty key vault.
var EmptyKeyVault crypto.KeyVault = emptyKeyVault{}

func (emptyKeyVault) Wrap(ctx context.Context, plaintext []byte, kekLabel string) ([]byte, error) {
	return nil, errKEKNotFound
}

func (emptyKeyVault) Unwrap(ctx context.Context, ciphertext []byte, kekLabel string) ([]byte, error) {
	return nil, errKEKNotFound
}

func (emptyKeyVault) GetCertificate(ctx context.Context, id string) (*x509.Certificate, error) {
	return nil, errCertificateNotFound
}

func (emptyKeyVault) ExportCertificate(ctx context.Context, id string) (*tls.Certificate, error) {
	return nil, errCertificateNotFound
}
