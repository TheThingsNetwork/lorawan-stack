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

package cryptoutil_test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"testing"

	"github.com/smartystreets/assertions"
	. "go.thethings.network/lorawan-stack/v3/pkg/crypto/cryptoutil"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

type mockKeyVault struct {
	params struct {
		ctx        context.Context
		ciphertext []byte
		kekLabel   string
	}
	results struct {
		key []byte
		err error
	}
	calls struct {
		count int
	}
}

func (*mockKeyVault) NsKEKLabel(ctx context.Context, netID *types.NetID, addr string) string {
	return ""
}

func (*mockKeyVault) AsKEKLabel(ctx context.Context, addr string) string {
	return ""
}

func (*mockKeyVault) Wrap(ctx context.Context, plaintext []byte, kekLabel string) ([]byte, error) {
	return nil, errors.New("not implemented")
}

func (m *mockKeyVault) Unwrap(ctx context.Context, ciphertext []byte, kekLabel string) ([]byte, error) {
	m.params.ctx, m.params.ciphertext, m.params.kekLabel = ctx, ciphertext, kekLabel
	m.calls.count++
	return m.results.key, m.results.err
}

func (*mockKeyVault) Encrypt(ctx context.Context, plaintext []byte, id string) ([]byte, error) {
	return nil, errors.New("not implemented")
}

func (m *mockKeyVault) Decrypt(ctx context.Context, ciphertext []byte, id string) ([]byte, error) {
	return nil, errors.New("not implemented")
}

func (*mockKeyVault) GetCertificate(ctx context.Context, id string) (*x509.Certificate, error) {
	return nil, errors.New("not implemented")
}

func (*mockKeyVault) ExportCertificate(ctx context.Context, id string) (*tls.Certificate, error) {
	return nil, errors.New("not implemented")
}

func TestCacheUsed(t *testing.T) {
	a := assertions.New(t)
	m := &mockKeyVault{}

	ctx := test.Context()
	ck := NewCacheKeyVault(m, test.Delay, 1)

	// Cache is empty, expect a miss
	m.results.key = []byte{0x01, 0x02, 0x03}
	m.results.err = nil

	key, err := ck.Unwrap(ctx, []byte{0x02, 0x03, 0x04}, "foo")
	a.So(key, should.Resemble, m.results.key)
	a.So(err, should.Equal, m.results.err)
	a.So(m.params.ctx, should.HaveParentContextOrEqual, ctx)
	a.So(m.params.ciphertext, should.Resemble, []byte{0x02, 0x03, 0x04})
	a.So(m.params.kekLabel, should.Equal, "foo")
	a.So(m.calls.count, should.Equal, 1)

	// Expect to be served from cache
	key, err = ck.Unwrap(ctx, []byte{0x02, 0x03, 0x04}, "foo")
	a.So(key, should.Resemble, m.results.key)
	a.So(err, should.Equal, m.results.err)
	a.So(m.calls.count, should.Equal, 1)

	// Expect the old element to be evicted, and a cache miss to occur
	m.results.key = []byte{0x05, 0x06, 0x07}
	m.results.err = nil
	m.calls.count = 0

	key, err = ck.Unwrap(ctx, []byte{0x03, 0x04, 0x05}, "bar")
	a.So(key, should.Resemble, m.results.key)
	a.So(err, should.Equal, m.results.err)
	a.So(m.params.ctx, should.HaveParentContextOrEqual, ctx)
	a.So(m.params.ciphertext, should.Resemble, []byte{0x03, 0x04, 0x05})
	a.So(m.params.kekLabel, should.Equal, "bar")
	a.So(m.calls.count, should.Equal, 1)

	// Expect the cache miss
	m.results.key = []byte{0x01, 0x02, 0x03}
	m.results.err = nil
	m.calls.count = 0

	key, err = ck.Unwrap(ctx, []byte{0x02, 0x03, 0x04}, "foo")
	a.So(key, should.Resemble, m.results.key)
	a.So(err, should.Equal, m.results.err)
	a.So(m.params.ctx, should.HaveParentContextOrEqual, ctx)
	a.So(m.params.ciphertext, should.Resemble, []byte{0x02, 0x03, 0x04})
	a.So(m.params.kekLabel, should.Equal, "foo")
	a.So(m.calls.count, should.Equal, 1)

	// Delay based evictions are left out since they may lead to flakyness.
}
