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
	"testing"
	"time"

	"github.com/bluele/gcache"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

type funcFetcher func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, fieldMaskPaths ...string) (*ttnpb.EndDevice, error)

func (f funcFetcher) Get(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, fieldMaskPaths ...string) (*ttnpb.EndDevice, error) {
	return f(ctx, ids, fieldMaskPaths...)
}

func TestEndDeviceFetcher(t *testing.T) {
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

		_, err := f.Get(test.Context(), ttnpb.EndDeviceIdentifiers{}, "locations")
		a.So(err, should.BeNil)
		a.So(numCalls, should.Equal, 1)

		fakeClock := gcache.NewFakeClock()
		cache := gcache.New(-1).Clock(fakeClock).Expiration(time.Second).Build()

		dev1 := ttnpb.EndDeviceIdentifiers{
			DeviceID: "dev1",
			ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
				ApplicationID: "app1",
			},
		}
		dev2 := ttnpb.EndDeviceIdentifiers{
			DeviceID: "dev2",
			ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
				ApplicationID: "app2",
			},
		}

		cf := applicationserver.NewCachedEndDeviceFetcher(f, cache)

		t.Run("Cold", func(t *testing.T) {
			a := assertions.New(t)
			cf.Get(test.Context(), dev1, "locations")
			a.So(numCalls, should.Equal, 2)
			cf.Get(test.Context(), dev1, "locations")
			a.So(numCalls, should.Equal, 2)
		})

		t.Run("Expire", func(t *testing.T) {
			a := assertions.New(t)
			fakeClock.Advance(2 * time.Second)
			cf.Get(test.Context(), dev1, "locations")
			a.So(numCalls, should.Equal, 3)
			cf.Get(test.Context(), dev1, "locations")
			a.So(numCalls, should.Equal, 3)
		})

		t.Run("OtherDevice", func(t *testing.T) {
			a := assertions.New(t)
			cf.Get(test.Context(), dev1, "locations")
			a.So(numCalls, should.Equal, 3)
			cf.Get(test.Context(), dev2, "locations")
			a.So(numCalls, should.Equal, 4)
		})

		t.Run("OtherFieldMask", func(t *testing.T) {
			a := assertions.New(t)
			cf.Get(test.Context(), dev1, "attributes")
			a.So(numCalls, should.Equal, 5)
			cf.Get(test.Context(), dev1, "attributes")
			a.So(numCalls, should.Equal, 5)
		})

		t.Run("CacheError", func(t *testing.T) {
			a := assertions.New(t)
			mockErr = fmt.Errorf("foobar")
			fakeClock.Advance(2 * time.Second)
			_, err := cf.Get(test.Context(), dev1, "locations")
			a.So(numCalls, should.Equal, 6)
			a.So(err, should.Resemble, mockErr)
			_, err = cf.Get(test.Context(), dev1, "locations")
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
				a.So(deadline.Sub(time.Now()), should.AlmostEqual, timeout, margin.Nanoseconds())
				return nil, nil
			},
		)
		cf := applicationserver.NewTimeoutEndDeviceFetcher(f, timeout)

		dev := ttnpb.EndDeviceIdentifiers{
			DeviceID: "dev1",
			ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
				ApplicationID: "app1",
			},
		}

		cf.Get(test.Context(), dev, "locations")
	})
}
