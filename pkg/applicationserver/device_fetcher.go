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

package applicationserver

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/bluele/gcache"
	pbtypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"golang.org/x/sync/singleflight"
)

// EndDeviceFetcher fetches end device protos.
type EndDeviceFetcher interface {
	Get(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, fieldMaskPaths ...string) (*ttnpb.EndDevice, error)
}

// NoopEndDeviceFetcher is a no-op.
type NoopEndDeviceFetcher struct{}

// Get implements the EndDeviceFetcher interface.
func (f *NoopEndDeviceFetcher) Get(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, fieldMaskPaths ...string) (*ttnpb.EndDevice, error) {
	return nil, errDeviceNotFound.WithAttributes("device_uid", unique.ID(ctx, ids))
}

// endDeviceFetcher fetches end devices
type endDeviceFetcher struct {
	c *component.Component
}

// NewRegistryEndDeviceFetcher returns a new endDeviceFetcher.
func NewRegistryEndDeviceFetcher(c *component.Component) EndDeviceFetcher {
	return &endDeviceFetcher{c}
}

// Get implements the EndDeviceFetcher interface.
func (f *endDeviceFetcher) Get(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, fieldMaskPaths ...string) (*ttnpb.EndDevice, error) {
	cc, err := f.c.GetPeerConn(ctx, ttnpb.ClusterRole_ENTITY_REGISTRY, &ids)
	if err != nil {
		return nil, err
	}

	return ttnpb.NewEndDeviceRegistryClient(cc).Get(ctx, &ttnpb.GetEndDeviceRequest{
		EndDeviceIdentifiers: ids,
		FieldMask: &pbtypes.FieldMask{
			Paths: fieldMaskPaths,
		},
	}, f.c.WithClusterAuth())
}

type cachedEndDeviceFetcher struct {
	fetcher EndDeviceFetcher
	cache   gcache.Cache
}

type endDeviceFetcherCacheEntry struct {
	err error
	dev *ttnpb.EndDevice
}

// NewCachedEndDeviceFetcher wraps an EndDeviceFetcher with a local cache.
func NewCachedEndDeviceFetcher(fetcher EndDeviceFetcher, cache gcache.Cache) EndDeviceFetcher {
	return &cachedEndDeviceFetcher{fetcher, cache}
}

// Get implements the EndDeviceFetcher interface.
func (f *cachedEndDeviceFetcher) Get(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, fieldMaskPaths ...string) (*ttnpb.EndDevice, error) {
	key := endDeviceKey(ctx, ids, fieldMaskPaths...)
	e, err := f.cache.Get(key)
	if entry, ok := e.(*endDeviceFetcherCacheEntry); err == nil && ok {
		return entry.dev, entry.err
	}
	dev, err := f.fetcher.Get(ctx, ids, fieldMaskPaths...)
	f.cache.Set(key, &endDeviceFetcherCacheEntry{err, dev})
	return dev, err
}

type singleFlightEndDeviceFetcher struct {
	fetcher      EndDeviceFetcher
	singleflight singleflight.Group
}

// NewSingleFlightEndDeviceFetcher wraps an EndDeviceFetcher with a single flight mechanism.
func NewSingleFlightEndDeviceFetcher(fetcher EndDeviceFetcher) EndDeviceFetcher {
	return &singleFlightEndDeviceFetcher{fetcher: fetcher}
}

// Get implements the EndDeviceFetcher interface.
func (f *singleFlightEndDeviceFetcher) Get(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, fieldMaskPaths ...string) (*ttnpb.EndDevice, error) {
	key := endDeviceKey(ctx, ids, fieldMaskPaths...)
	dev, err, _ := f.singleflight.Do(key, func() (interface{}, error) {
		return f.fetcher.Get(ctx, ids, fieldMaskPaths...)
	})
	if err != nil {
		return nil, err
	}
	return dev.(*ttnpb.EndDevice), nil
}

type timeoutEndDeviceFetcher struct {
	fetcher EndDeviceFetcher
	timeout time.Duration
}

// NewTimeoutEndDeviceFetcher wraps an EndDeviceFetcher and limits the lifetime of the context used to retrieve the end device.
func NewTimeoutEndDeviceFetcher(fetcher EndDeviceFetcher, timeout time.Duration) EndDeviceFetcher {
	return &timeoutEndDeviceFetcher{fetcher, timeout}
}

// Get implements the EndDeviceFetcher interface.
func (f *timeoutEndDeviceFetcher) Get(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, fieldMaskPaths ...string) (*ttnpb.EndDevice, error) {
	ctx, cancel := context.WithTimeout(ctx, f.timeout)
	defer cancel()
	return f.fetcher.Get(ctx, ids, fieldMaskPaths...)
}

type circuitBreakerEndDeviceFetcher struct {
	fetcher EndDeviceFetcher

	threshold uint64
	timeout   time.Duration

	mu                sync.RWMutex
	failures          uint64
	lastFailedAttempt time.Time
}

// NewCircuitBreakerEndDeviceFetcher wraps an end device fetcher with a circuit breaking mechanism.
// The circuit breaker opens when the number of failure attempts is higher than the threshold,
// and closes after the provided timeout.
func NewCircuitBreakerEndDeviceFetcher(fetcher EndDeviceFetcher, threshold uint64, timeout time.Duration) EndDeviceFetcher {
	return &circuitBreakerEndDeviceFetcher{
		fetcher:   fetcher,
		threshold: threshold,
		timeout:   timeout,
	}
}

var errCircuitBreakerOpen = errors.DefineUnavailable("circuit_breaker_open", "circuit breaker open")

func (f *circuitBreakerEndDeviceFetcher) circuitOpen() error {
	f.mu.RLock()
	defer f.mu.RUnlock()
	// If the number of failures is still under the threshold, consider the circuit as being closed.
	if f.failures < f.threshold {
		return nil
	}
	// If the attempt timeout expired, consider the circuit as being closed.
	if time.Since(f.lastFailedAttempt) > f.timeout {
		return nil
	}
	// At this point we have a number of failures that is above the threshold, and any attempts are recent.
	// The circuit breaker is open.
	return errCircuitBreakerOpen.New()
}

func (f *circuitBreakerEndDeviceFetcher) observeError(ctx context.Context, err error) {
	logger := log.FromContext(ctx).WithField("circuit", "end_device_fetcher")
	f.mu.Lock()
	defer f.mu.Unlock()
	switch {
	case errors.IsCanceled(err),
		errors.IsDeadlineExceeded(err),
		errors.IsAborted(err),
		errors.IsInternal(err),
		errors.IsUnavailable(err):
		f.lastFailedAttempt = time.Now()
		f.failures++
		if f.failures < f.threshold {
			return
		}
		logger.WithError(err).WithField("count", f.failures).Warn("Circuit breaker open")
	case err == nil:
		n := f.failures
		f.lastFailedAttempt = time.Time{}
		f.failures = 0
		if n < f.threshold {
			return
		}
		logger.WithField("previous_count", n).Info("Circuit breaker closed")
	}
}

// Get implements the EndDeviceFetcher interface.
func (f *circuitBreakerEndDeviceFetcher) Get(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, fieldMaskPaths ...string) (*ttnpb.EndDevice, error) {
	if err := f.circuitOpen(); err != nil {
		return nil, err
	}
	dev, err := f.fetcher.Get(ctx, ids, fieldMaskPaths...)
	f.observeError(ctx, err)
	return dev, err
}

func endDeviceKey(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, fieldMaskPaths ...string) string {
	return fmt.Sprintf("%s:%s", unique.ID(ctx, ids), strings.Join(fieldMaskPaths, ","))
}
