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
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/config"
	"github.com/TheThingsNetwork/ttn/pkg/store"
	. "github.com/TheThingsNetwork/ttn/pkg/store/redis"
	"github.com/TheThingsNetwork/ttn/pkg/store/storetest"
)

var (
	_ store.ByteMapStore  = &Store{}
	_ store.ByteListStore = &Store{}
	_ store.ByteSetStore  = &Store{}
)

// redisConfig returns a new redis config for testing
func redisConfig() config.Redis {
	var err error
	config := config.Redis{
		Address:   "localhost:6379",
		Database:  1,
		Namespace: []string{"test"},
	}
	if address := os.Getenv("REDIS_ADDRESS"); address != "" {
		config.Address = address
	}
	if db := os.Getenv("REDIS_DB"); db != "" {
		config.Database, err = strconv.Atoi(db)
		if err != nil {
			panic(err)
		}
	}
	if prefix := os.Getenv("REDIS_PREFIX"); prefix != "" {
		config.Namespace = []string{prefix}
	}
	return config
}

func newStore() *Store {
	conf := &Config{
		Redis:     redisConfig(),
		IndexKeys: storetest.IndexedFields,
		Namespace: []string{"test"},
	}

	s := New(conf)
	keys, err := s.Redis.Keys(strings.Join(append(conf.Redis.Namespace, conf.Namespace...), Separator) + Separator + "*").Result()
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
