// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package tokenkey

import (
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestNilCache(t *testing.T) {
	a := assertions.New(t)

	a.So(NilCache.Set("http://foo", "kid", "ES256", "PEM"), should.BeNil)

	key, alg, err := NilCache.Get("http://foo", "kid")
	a.So(key, should.BeEmpty)
	a.So(alg, should.BeEmpty)
	a.So(err, should.BeNil)
}

func TestMemoryCache(t *testing.T) {
	a := assertions.New(t)

	cache := &MemoryCache{}

	iss := "http://foo.test"
	kid := "kid"
	alg := "ES256"
	key := "PEM"

	// set the key
	a.So(cache.Set(iss, kid, alg, key), should.BeNil)

	// get the key
	{
		key, alg, err := cache.Get(iss, kid)
		a.So(key, should.Equal, key)
		a.So(alg, should.Equal, alg)
		a.So(err, should.BeNil)
	}

	// get another key
	{
		key, alg, err := cache.Get("http://other.test", kid)
		a.So(key, should.BeEmpty)
		a.So(alg, should.BeEmpty)
		a.So(err, should.BeNil)
	}

	// get another key
	{
		key, alg, err := cache.Get(iss, "other")
		a.So(key, should.BeEmpty)
		a.So(alg, should.BeEmpty)
		a.So(err, should.BeNil)
	}

	// set another key
	{
		a.So(cache.Set(iss, "other", alg, key), should.BeNil)
		key, alg, err := cache.Get(iss, kid)
		a.So(key, should.Equal, key)
		a.So(alg, should.Equal, alg)
		a.So(err, should.BeNil)
	}

	// get the key
	{
		key, alg, err := cache.Get(iss, kid)
		a.So(key, should.Equal, key)
		a.So(alg, should.Equal, alg)
		a.So(err, should.BeNil)
	}
}
