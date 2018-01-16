// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package grpc

import (
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestCacheTTL(t *testing.T) {
	a := assertions.New(t)
	cache := newTTLCache(time.Duration(time.Second * 3))

	auth := "Bearer eyJjSjsjsjjs.dsj"
	entityID := "foo-app"

	rights, err := cache.GetOrFetch(auth, entityID, func() ([]ttnpb.Right, error) {
		return []ttnpb.Right{ttnpb.Right(1)}, nil
	})
	a.So(err, should.BeNil)
	a.So(rights, should.Resemble, []ttnpb.Right{ttnpb.Right(1)})
	a.So(cache.entries, should.HaveLength, 1)

	// although fetch return's is different the previous response is still cached
	rights, err = cache.GetOrFetch(auth, entityID, func() ([]ttnpb.Right, error) {
		return []ttnpb.Right{ttnpb.Right(2)}, nil
	})
	a.So(err, should.BeNil)
	a.So(rights, should.Resemble, []ttnpb.Right{ttnpb.Right(1)})
	a.So(cache.entries, should.HaveLength, 1)

	// sleep for 5 seconds so the entry expires
	time.Sleep(5 * time.Second)

	// entry has been garbage collected
	a.So(cache.entries, should.HaveLength, 0)

	// fetch again with different response
	rights, err = cache.GetOrFetch(auth, entityID, func() ([]ttnpb.Right, error) {
		return []ttnpb.Right{ttnpb.Right(2)}, nil
	})
	a.So(err, should.BeNil)
	a.So(rights, should.Resemble, []ttnpb.Right{ttnpb.Right(2)})
	a.So(cache.entries, should.HaveLength, 1)
}
