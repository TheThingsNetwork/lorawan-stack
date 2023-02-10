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

package cryptoutil_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"math/big"
	"testing"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/crypto"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto/cryptoutil"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

type mockKeyVault struct {
	KeyFunc               func(ctx context.Context, label string) ([]byte, error)
	ServerCertificateFunc func(ctx context.Context, label string) (tls.Certificate, error)
	ClientCertificateFunc func(ctx context.Context, label string) (tls.Certificate, error)
}

func (m *mockKeyVault) Key(ctx context.Context, label string) ([]byte, error) {
	if m.KeyFunc == nil {
		panic("Key called but not set")
	}
	return m.KeyFunc(ctx, label)
}

func (m *mockKeyVault) ServerCertificate(ctx context.Context, label string) (tls.Certificate, error) {
	if m.ServerCertificateFunc == nil {
		panic("ServerCertificate called but not set")
	}
	return m.ServerCertificateFunc(ctx, label)
}

func (m *mockKeyVault) ClientCertificate(ctx context.Context, label string) (tls.Certificate, error) {
	if m.ClientCertificateFunc == nil {
		panic("ClientCertificate called but not set")
	}
	return m.ClientCertificateFunc(ctx, label)
}

var _ crypto.KeyVault = (*mockKeyVault)(nil)

type mockClock struct {
	time.Time
}

func (m mockClock) Now() time.Time {
	return m.Time
}

func TestCacheKeyVault(t *testing.T) {
	t.Parallel()
	a, ctx := test.New(t)

	var (
		keys, serverCerts []string
		ref               = time.Now()
		now               = &mockClock{ref}
		err               error
	)
	kv := cryptoutil.NewCacheKeyVault(
		&mockKeyVault{
			KeyFunc: func(ctx context.Context, label string) ([]byte, error) {
				keys = append(keys, label)
				if label == "error" {
					return nil, errors.New("error")
				}
				return []byte(label), nil
			},
			ServerCertificateFunc: func(ctx context.Context, label string) (tls.Certificate, error) {
				serverCerts = append(serverCerts, label)
				tmpl := &x509.Certificate{
					IsCA: true,
					Subject: pkix.Name{
						CommonName: label,
					},
					DNSNames: []string{label},
					NotAfter: time.Now().Add(3 * time.Hour),
					ExtKeyUsage: []x509.ExtKeyUsage{
						x509.ExtKeyUsageServerAuth,
					},
					SerialNumber: big.NewInt(1),
				}
				key, err := rsa.GenerateKey(rand.Reader, 2048)
				if err != nil {
					return tls.Certificate{}, err
				}
				cert, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
				if err != nil {
					return tls.Certificate{}, err
				}
				return tls.Certificate{Certificate: [][]byte{cert}}, nil
			},
		},
		cryptoutil.WithCacheKeyVaultClock(now),
		cryptoutil.WithCacheKeyVaultTTL(4*time.Hour, 5*time.Minute),
	)

	test.Must(kv.Key(ctx, "key1"))
	test.Must(kv.Key(ctx, "key1"))
	test.Must(kv.Key(ctx, "key2"))
	test.Must(kv.Key(ctx, "key1"))
	test.Must(kv.Key(ctx, "key2"))
	a.So(keys, should.Resemble, []string{"key1", "key2"})

	test.Must(kv.ServerCertificate(ctx, "server1.example.com"))
	test.Must(kv.ServerCertificate(ctx, "server2.example.com"))
	test.Must(kv.ServerCertificate(ctx, "server1.example.com"))
	a.So(serverCerts, should.Resemble, []string{"server1.example.com", "server2.example.com"})

	// Errors are also cached.
	_, err = kv.Key(ctx, "error")
	a.So(err, should.NotBeNil)
	_, err = kv.Key(ctx, "error")
	a.So(err, should.NotBeNil)
	a.So(keys, should.Resemble, []string{"key1", "key2", "error"})

	// 10 minutes after reference time, the error is no longer cached.
	now.Time = ref.Add(10 * time.Minute)
	_, err = kv.Key(ctx, "error")
	a.So(err, should.NotBeNil)
	a.So(keys, should.Resemble, []string{"key1", "key2", "error", "error"})

	// 1 hour after reference time, the certficates are still valid.
	now.Time = ref.Add(1 * time.Hour)
	test.Must(kv.ServerCertificate(ctx, "server1.example.com"))
	test.Must(kv.ServerCertificate(ctx, "server2.example.com"))
	a.So(serverCerts, should.Resemble, []string{"server1.example.com", "server2.example.com"})

	// 3 hours after reference time, the certificates are expired.
	now.Time = ref.Add(3 * time.Hour)
	test.Must(kv.ServerCertificate(ctx, "server1.example.com"))
	test.Must(kv.ServerCertificate(ctx, "server2.example.com"))
	a.So(serverCerts, should.Resemble, []string{
		"server1.example.com",
		"server2.example.com",
		"server1.example.com",
		"server2.example.com",
	})
}
