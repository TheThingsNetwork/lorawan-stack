// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package redis_test

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/store"
	. "github.com/TheThingsNetwork/ttn/pkg/store/redis"
	"github.com/TheThingsNetwork/ttn/pkg/store/storetest"
	"github.com/TheThingsNetwork/ttn/pkg/util/test"
)

func newStore() *Store {
	s := New(&Config{
		Redis:     test.RedisConfig(),
		IndexKeys: []string{"foo", "bar"},
	})
	keys, err := s.Redis.Keys("test:*").Result()
	if err != nil {
		panic(err)
	}
	if len(keys) > 0 {
		_, err = s.Redis.Del(keys...).Result()
		if err != nil {
			panic(err)
		}
	}
	return s
}

func TestByteStore(t *testing.T) {
	storetest.TestByteStore(t, func() store.ByteStore { return newStore() })
}

func TestByteSetStore(t *testing.T) {
	storetest.TestByteSetStore(t, func() store.ByteSetStore { return newStore() })
}
