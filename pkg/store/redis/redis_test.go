// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package redis_test

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/cmd/shared"
	"github.com/TheThingsNetwork/ttn/pkg/config"
	. "github.com/TheThingsNetwork/ttn/pkg/store/redis"
	"github.com/TheThingsNetwork/ttn/pkg/store/storetest"
)

func TestStore(t *testing.T) {
	s := New(&Config{
		Redis: config.Redis{
			Address:  shared.DefaultRedisConfig.Address,
			Database: 9,
			Prefix:   "test",
		},
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
	storetest.TestByteStore(t, s)
}
