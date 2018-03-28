// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package rights

import (
	"context"
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/log"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/TheThingsNetwork/ttn/pkg/util/test"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

// Make sure in compile time that noopCache and ttlCache implements cache.
var (
	_ cache = new(noopCache)
	_ cache = new(ttlCache)
)

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
