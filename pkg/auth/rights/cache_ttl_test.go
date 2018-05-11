// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

package rights

import (
	"context"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
)

var _ cache = new(ttlCache)

func TestCacheTTL(t *testing.T) {
	a := assertions.New(t)
	ctx, cancel := context.WithCancel(log.NewContextWithField(log.NewContext(context.Background(), test.GetLogger(t)), "hook", "rights"))
	cache := newTTLCache(ctx, time.Duration(time.Millisecond*200))

	key := "foo"

	rights, err := cache.GetOrFetch(key, func() ([]ttnpb.Right, error) {
		return []ttnpb.Right{ttnpb.Right(1)}, nil
	})
	a.So(err, should.BeNil)
	a.So(rights, should.Resemble, []ttnpb.Right{ttnpb.Right(1)})
	a.So(cache.entries, should.HaveLength, 1)

	// Although fetch function is different the previous response is still cached.
	rights, err = cache.GetOrFetch(key, func() ([]ttnpb.Right, error) {
		return []ttnpb.Right{ttnpb.Right(2)}, nil
	})
	a.So(err, should.BeNil)
	a.So(rights, should.Resemble, []ttnpb.Right{ttnpb.Right(1)})
	a.So(cache.entries, should.HaveLength, 1)

	// Sleep for 250 milliseconds so the entry expires.
	time.Sleep(time.Millisecond * 250)

	// Entry has been garbage collected.
	a.So(cache.entries, should.HaveLength, 0)

	// Stop the garbage collector.
	cancel()

	// Fetch again with different response.
	rights, err = cache.GetOrFetch(key, func() ([]ttnpb.Right, error) {
		return []ttnpb.Right{ttnpb.Right(2)}, nil
	})
	a.So(err, should.BeNil)
	a.So(rights, should.Resemble, []ttnpb.Right{ttnpb.Right(2)})
	a.So(cache.entries, should.HaveLength, 1)

	// Sleep for 250 milliseconds again.
	time.Sleep(time.Millisecond * 250)

	// Check that the entry has not been garbage collected.
	a.So(cache.entries, should.HaveLength, 1)
	a.So(cache.entries, should.ContainKey, key)
}
