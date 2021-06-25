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

package applicationserver_test

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/bluele/gcache"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

type funcFetcher func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, fieldMaskPaths ...string) (*ttnpb.EndDevice, error)

func (f funcFetcher) Get(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, fieldMaskPaths ...string) (*ttnpb.EndDevice, error) {
	return f(ctx, ids, fieldMaskPaths...)
}

func TestEndDeviceFetcher(t *testing.T) {
	ctx := log.NewContext(test.Context(), test.GetLogger(t))
	t.Run("Cache", func(t *testing.T) {
		a := assertions.New(t)
		numCalls := 0
		var mockErr error
		f := funcFetcher(
			func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, fieldMaskPaths ...string) (*ttnpb.EndDevice, error) {
				numCalls++
				return nil, mockErr
			},
		)

		_, err := f.Get(ctx, ttnpb.EndDeviceIdentifiers{}, "locations")
		a.So(err, should.BeNil)
		a.So(numCalls, should.Equal, 1)

		fakeClock := gcache.NewFakeClock()
		cache := gcache.New(-1).Clock(fakeClock).Expiration(time.Second).Build()

		dev1 := ttnpb.EndDeviceIdentifiers{
			DeviceId: "dev1",
			ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
				ApplicationId: "app1",
			},
		}
		dev2 := ttnpb.EndDeviceIdentifiers{
			DeviceId: "dev2",
			ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
				ApplicationId: "app2",
			},
		}

		cf := applicationserver.NewCachedEndDeviceFetcher(f, cache)

		t.Run("Cold", func(t *testing.T) {
			a := assertions.New(t)
			cf.Get(ctx, dev1, "locations")
			a.So(numCalls, should.Equal, 2)
			cf.Get(ctx, dev1, "locations")
			a.So(numCalls, should.Equal, 2)
		})

		t.Run("Expire", func(t *testing.T) {
			a := assertions.New(t)
			fakeClock.Advance(2 * time.Second)
			cf.Get(ctx, dev1, "locations")
			a.So(numCalls, should.Equal, 3)
			cf.Get(ctx, dev1, "locations")
			a.So(numCalls, should.Equal, 3)
		})

		t.Run("OtherDevice", func(t *testing.T) {
			a := assertions.New(t)
			cf.Get(ctx, dev1, "locations")
			a.So(numCalls, should.Equal, 3)
			cf.Get(ctx, dev2, "locations")
			a.So(numCalls, should.Equal, 4)
		})

		t.Run("OtherFieldMask", func(t *testing.T) {
			a := assertions.New(t)
			cf.Get(ctx, dev1, "attributes")
			a.So(numCalls, should.Equal, 5)
			cf.Get(ctx, dev1, "attributes")
			a.So(numCalls, should.Equal, 5)
		})

		t.Run("CacheError", func(t *testing.T) {
			a := assertions.New(t)
			mockErr = fmt.Errorf("foobar")
			fakeClock.Advance(2 * time.Second)
			_, err := cf.Get(ctx, dev1, "locations")
			a.So(numCalls, should.Equal, 6)
			a.So(err, should.Resemble, mockErr)
			_, err = cf.Get(ctx, dev1, "locations")
			a.So(numCalls, should.Equal, 6)
			a.So(err, should.Resemble, mockErr)
		})
	})
	t.Run("Timeout", func(t *testing.T) {
		timeout := 5 * time.Second
		margin := 100 * time.Millisecond
		a := assertions.New(t)
		f := funcFetcher(
			func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, fieldMaskPaths ...string) (*ttnpb.EndDevice, error) {
				deadline, ok := ctx.Deadline()
				a.So(ok, should.BeTrue)
				a.So(time.Until(deadline), should.AlmostEqual, timeout, margin.Nanoseconds())
				return nil, nil
			},
		)
		cf := applicationserver.NewTimeoutEndDeviceFetcher(f, timeout)

		dev := ttnpb.EndDeviceIdentifiers{
			DeviceId: "dev1",
			ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
				ApplicationId: "app1",
			},
		}

		cf.Get(ctx, dev, "locations")
	})
	t.Run("CircuitBreaker", func(t *testing.T) {
		timeout := (1 << 6) * test.Delay
		threshold := uint64(10)
		var mockErr error
		numCalls := uint64(0)
		f := funcFetcher(
			func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, fieldMaskPaths ...string) (*ttnpb.EndDevice, error) {
				atomic.AddUint64(&numCalls, 1)
				return nil, mockErr
			},
		)

		cf := applicationserver.NewCircuitBreakerEndDeviceFetcher(f, threshold, timeout)

		dev := ttnpb.EndDeviceIdentifiers{
			DeviceId: "dev1",
			ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
				ApplicationId: "app1",
			},
		}

		t.Run("InitialClosed", func(t *testing.T) {
			a := assertions.New(t)
			wg := sync.WaitGroup{}
			for i := 0; i < 10; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					_, err := cf.Get(ctx, dev, "location")
					a.So(err, should.BeNil)
				}()
			}
			wg.Wait()
			a.So(numCalls, should.Equal, 10)
		})

		t.Run("InitialBurst", func(t *testing.T) {
			a := assertions.New(t)
			mockErr = errors.DefineUnavailable("unavailable", "server unavailable").New()
			wg := sync.WaitGroup{}
			for i := 0; i < 10; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					_, err := cf.Get(ctx, dev, "location")
					a.So(err, should.NotBeNil)
					a.So(err, should.Resemble, mockErr)
				}()
			}
			wg.Wait()
			a.So(numCalls, should.Equal, 20)
		})

		t.Run("BreakerOpen", func(t *testing.T) {
			a := assertions.New(t)
			wg := sync.WaitGroup{}
			for i := 0; i < 10; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					_, err := cf.Get(ctx, dev, "location")
					a.So(err, should.NotBeNil)
				}()
			}
			wg.Wait()
			a.So(numCalls, should.Equal, 20)
		})

		t.Run("SecondBurst", func(t *testing.T) {
			a := assertions.New(t)
			wg := sync.WaitGroup{}

			time.Sleep(2 * timeout)

			for i := 0; i < 10; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					_, err := cf.Get(ctx, dev, "location")
					a.So(err, should.NotBeNil)
				}()
			}
			wg.Wait()
			// After the circuit breaker timeout expires, multiple calls
			// may be allowed to execute simultaneously in order to verify
			// if the underlying fetcher recovered. We expect at least one
			// new attempt after the timeout, but it is possible that all
			// of the goroutines actually do a call.
			a.So(numCalls, should.BeBetweenOrEqual, 21, 30)
		})

		t.Run("BreakerClosed", func(t *testing.T) {
			a := assertions.New(t)
			wg := sync.WaitGroup{}

			time.Sleep(2 * timeout)
			mockErr = nil

			initialCalls := numCalls
			for i := 0; i < 10; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					_, err := cf.Get(ctx, dev, "location")
					a.So(err, should.BeNil)
				}()
			}
			wg.Wait()
			a.So(numCalls, should.Equal, initialCalls+10)
		})
	})
}
