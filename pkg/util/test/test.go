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

// Package test provides various testing utilities.
package test

import (
	"context"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto/cryptoutil"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

const (
	DefaultApplicationID = "test-app-id"
	DefaultDeviceID      = "test-dev-id"

	DefaultRootKeyID = "test-root-key-id"
)

var (
	ErrInternal = errors.DefineInternal("test_internal", "test error")
	ErrNotFound = errors.DefineNotFound("test_not_found", "test error")

	DefaultApplicationIdentifiers = ttnpb.ApplicationIdentifiers{
		ApplicationId: DefaultApplicationID,
	}

	DefaultJoinEUI = types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	DefaultDevEUI  = types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

	DefaultAppKey = types.AES128Key{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	DefaultNwkKey = types.AES128Key{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

	DefaultJoinNonce = types.JoinNonce{0x42, 0xff, 0xff}
	DefaultDevNonce  = types.DevNonce{0x42, 0xff}

	DefaultSessionKeyID = []byte("test-session-key-id")

	DefaultKEK      = types.AES128Key{0x42, 0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	DefaultKEKLabel = "test-kek-label"

	DefaultKeyVault = config.KeyVault{
		Provider: "static",
		Static: map[string][]byte{
			DefaultKEKLabel: DefaultKEK[:],
		},
	}

	DefaultAppSKey     = crypto.DeriveAppSKey(DefaultNwkKey, DefaultJoinNonce, DefaultJoinEUI, DefaultDevNonce)
	DefaultFNwkSIntKey = crypto.DeriveFNwkSIntKey(DefaultNwkKey, DefaultJoinNonce, DefaultJoinEUI, DefaultDevNonce)
	DefaultNwkSEncKey  = crypto.DeriveNwkSEncKey(DefaultNwkKey, DefaultJoinNonce, DefaultJoinEUI, DefaultDevNonce)
	DefaultSNwkSIntKey = crypto.DeriveSNwkSIntKey(DefaultNwkKey, DefaultJoinNonce, DefaultJoinEUI, DefaultDevNonce)

	DefaultAppSKeyEnvelope = &ttnpb.KeyEnvelope{
		Key: &DefaultAppSKey,
	}
	DefaultFNwkSIntKeyEnvelope = &ttnpb.KeyEnvelope{
		Key: &DefaultFNwkSIntKey,
	}
	DefaultNwkSEncKeyEnvelope = &ttnpb.KeyEnvelope{
		Key: &DefaultNwkSEncKey,
	}
	DefaultSNwkSIntKeyEnvelope = &ttnpb.KeyEnvelope{
		Key: &DefaultSNwkSIntKey,
	}

	DefaultAppSKeyEnvelopeWrapped     = Must(cryptoutil.WrapAES128KeyWithKEK(Context(), DefaultAppSKey, DefaultKEKLabel, DefaultKEK)).(*ttnpb.KeyEnvelope)
	DefaultFNwkSIntKeyEnvelopeWrapped = Must(cryptoutil.WrapAES128KeyWithKEK(Context(), DefaultFNwkSIntKey, DefaultKEKLabel, DefaultKEK)).(*ttnpb.KeyEnvelope)
	DefaultNwkSEncKeyEnvelopeWrapped  = Must(cryptoutil.WrapAES128KeyWithKEK(Context(), DefaultNwkSEncKey, DefaultKEKLabel, DefaultKEK)).(*ttnpb.KeyEnvelope)
	DefaultSNwkSIntKeyEnvelopeWrapped = Must(cryptoutil.WrapAES128KeyWithKEK(Context(), DefaultSNwkSIntKey, DefaultKEKLabel, DefaultKEK)).(*ttnpb.KeyEnvelope)

	DefaultAppSKeyWrapped     = DefaultAppSKeyEnvelopeWrapped.EncryptedKey
	DefaultFNwkSIntKeyWrapped = DefaultFNwkSIntKeyEnvelopeWrapped.EncryptedKey
	DefaultNwkSEncKeyWrapped  = DefaultNwkSEncKeyEnvelopeWrapped.EncryptedKey
	DefaultSNwkSIntKeyWrapped = DefaultSNwkSIntKeyEnvelopeWrapped.EncryptedKey

	DefaultNetID   = Must(types.NewNetID(2, []byte{0x00, 0x42, 0xff})).(types.NetID)
	DefaultDevAddr = Must(types.NewDevAddr(DefaultNetID, []byte{0x00, 0x02, 0xff, 0xff})).(types.DevAddr)

	DefaultLegacyAppSKey = crypto.DeriveLegacyAppSKey(DefaultNwkKey, DefaultJoinNonce, DefaultNetID, DefaultDevNonce)
	DefaultLegacyNwkSKey = crypto.DeriveLegacyNwkSKey(DefaultNwkKey, DefaultJoinNonce, DefaultNetID, DefaultDevNonce)

	DefaultMACVersion      = ttnpb.MACVersion_MAC_V1_1
	DefaultPHYVersion      = ttnpb.PHYVersion_RP001_V1_1_REV_B
	DefaultFrequencyPlanID = EUFrequencyPlanID
)

func NewWithContext(ctx context.Context, tb testing.TB) (*assertions.Assertion, context.Context) {
	tb.Helper()
	return assertions.New(tb), ContextWithTB(
		log.NewContext(
			ctx, GetLogger(tb),
		),
		tb,
	)
}

func New(tb testing.TB) (*assertions.Assertion, context.Context) {
	tb.Helper()
	return NewWithContext(Context(), tb)
}

func NewTBFromContext(ctx context.Context) (testing.TB, *assertions.Assertion, bool) {
	tb, ok := TBFromContext(ctx)
	if !ok {
		return nil, nil, false
	}
	tb.Helper()
	return tb, assertions.New(tb), true
}

func MustNewTBFromContext(ctx context.Context) (testing.TB, *assertions.Assertion) {
	tb := MustTBFromContext(ctx)
	tb.Helper()
	return tb, assertions.New(tb)
}

func NewTFromContext(ctx context.Context) (*testing.T, *assertions.Assertion, bool) {
	t, ok := TFromContext(ctx)
	if !ok {
		return nil, nil, false
	}
	t.Helper()
	return t, assertions.New(t), true
}

func MustNewTFromContext(ctx context.Context) (*testing.T, *assertions.Assertion) {
	t := MustTFromContext(ctx)
	t.Helper()
	return t, assertions.New(t)
}

var defaultTestTimeout = (1 << 18) * Delay

type TestConfig struct {
	Parallel bool
	Timeout  time.Duration
	Func     func(context.Context, *assertions.Assertion)
}

func runTestFromContext(ctx context.Context, conf TestConfig) {
	t := MustTFromContext(ctx)
	t.Helper()

	if conf.Parallel {
		t.Parallel()
	}
	a, ctx := New(t)
	ctx, cancel := context.WithDeadline(ctx, func() time.Time {
		timeout := conf.Timeout
		if timeout == 0 {
			timeout = defaultTestTimeout
		}
		dl := time.Now().Add(timeout)
		tDL, ok := t.Deadline()
		if ok && tDL.Before(dl) {
			return tDL
		}
		return dl
	}())
	defer cancel()

	dl, ok := ctx.Deadline()
	if !ok {
		panic("missing deadline in context")
	}
	timeout := time.Until(dl)

	start := time.Now()
	doneCh := make(chan struct{})
	defer func() {
		t.Helper()
		close(doneCh)
		if d := time.Since(start); d > timeout {
			t.Errorf("%s took too long to execute. Expected execution time below %v, ran for %v", t.Name(), timeout, d)
		}
	}()
	go func() {
		for {
			select {
			case <-doneCh:
				return
			case <-time.Tick(timeout / 4):
				t.Logf("%s is taking a long time to execute. Expected execution time below %v, running already for: %v", t.Name(), timeout, time.Since(start))
			}
		}
	}()
	conf.Func(ctx, a)
}

func RunTest(t *testing.T, conf TestConfig) {
	t.Helper()
	_, ctx := New(t)
	runTestFromContext(ctx, conf)
}

var defaultSubtestTimeout = defaultTestTimeout / 4

type SubtestConfig struct {
	Name     string
	Parallel bool
	Timeout  time.Duration
	Func     func(context.Context, *testing.T, *assertions.Assertion)
}

func RunSubtestFromContext(ctx context.Context, conf SubtestConfig) bool {
	t := MustTFromContext(ctx)
	t.Helper()
	// NOTE: When `-failfast` is specified, t.Run may not run and return true.
	// https://github.com/golang/go/blob/ae658cb19a265f3f4694cd4aec508b4565bda6aa/src/testing/testing.go#L1158-L1160
	var called bool
	ok := t.Run(conf.Name, func(t *testing.T) {
		called = true
		t.Helper()

		timeout := conf.Timeout
		if timeout == 0 {
			timeout = defaultSubtestTimeout
		}
		_, ctx = NewWithContext(ctx, t)
		runTestFromContext(ctx, TestConfig{
			Parallel: conf.Parallel,
			Timeout:  timeout,
			Func: func(ctx context.Context, a *assertions.Assertion) {
				t.Helper()
				conf.Func(ctx, t, a)
			},
		})
	})
	if ok && !called {
		t.Skip("Subtest did not execute, perhaps due to `-failfast`")
	}
	return ok && called
}

func RunSubtest(t *testing.T, conf SubtestConfig) bool {
	t.Helper()
	_, ctx := New(t)
	return RunSubtestFromContext(ctx, conf)
}
