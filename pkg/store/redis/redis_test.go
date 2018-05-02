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

package redis_test

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/store"
	. "github.com/TheThingsNetwork/ttn/pkg/store/redis"
	"github.com/TheThingsNetwork/ttn/pkg/store/storetest"
	"github.com/TheThingsNetwork/ttn/pkg/util/test"
)

func newStore() *Store {
	conf := &Config{
		Redis:     test.RedisConfig(),
		IndexKeys: storetest.IndexedFields,
	}

	s := New(conf)
	keys, err := s.Redis.Keys(conf.Prefix + Separator + "*").Result()
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

func TestByteMapStore(t *testing.T) {
	storetest.TestByteMapStore(t, func() store.ByteMapStore { return newStore() })
}

func TestByteSetStore(t *testing.T) {
	storetest.TestByteSetStore(t, func() store.ByteSetStore { return newStore() })
}
