// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

package test

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/go-redis/redis"
	ulid "github.com/oklog/ulid/v2"
	"go.thethings.network/lorawan-stack/pkg/config"
	ttnredis "go.thethings.network/lorawan-stack/pkg/redis"
)

const (
	defaultDatabase = 1
	defaultAddress  = "localhost:6379"
)

var defaultNamespace = [...]string{
	"redistest",
}

// NewRedis returns a new namespaced *redis.Client ready to use
// and a flush function, which should be called after the client is not needed anymore to clean the namespace.
// NewRedis respects TEST_REDIS, REDIS_ADDRESS and REDIS_DB environment variables.
// Client returned logs commands executed.
func NewRedis(t testing.TB, namespace ...string) (*ttnredis.Client, func()) {
	if os.Getenv("TEST_REDIS") != "1" {
		t.Skip("TEST_REDIS is not set to `1`, skipping Redis tests")
		panic("New called outside test")
	}

	conf := &ttnredis.Config{
		Redis: config.Redis{
			Address:   defaultAddress,
			Database:  defaultDatabase,
			Namespace: defaultNamespace[:],
		},
		Namespace: append(append([]string{ulid.MustNew(ulid.Now(), Randy).String()}, namespace...), t.Name()),
	}

	if addr := os.Getenv("REDIS_ADDRESS"); addr != "" {
		conf.Redis.Address = addr
	}
	if db := os.Getenv("REDIS_DB"); db != "" {
		var err error
		conf.Redis.Database, err = strconv.Atoi(db)
		if err != nil {
			t.Fatalf("Expected REDIS_DB to be an integer, got `%s`", db)
			return nil, nil
		}
	}

	cl := ttnredis.New(conf)

	if err := cl.Ping().Err(); err != nil {
		t.Fatalf("Failed to ping Redis: `%s`", err)
	}

	formatCmd := func(cmd redis.Cmder) string {
		ss := make([]string, 0, len(cmd.Args()))
		for _, arg := range cmd.Args() {
			ss = append(ss, fmt.Sprint(arg))
		}
		return strings.Join(ss, " ")
	}

	cl.Client.WrapProcess(func(p func(redis.Cmder) error) func(redis.Cmder) error {
		logger := GetLogger(t)
		return func(cmd redis.Cmder) error {
			logger.Debugf("Executing `%s`", formatCmd(cmd))
			return p(cmd)
		}
	})
	cl.Client.WrapProcessPipeline(func(p func([]redis.Cmder) error) func([]redis.Cmder) error {
		logger := GetLogger(t)
		return func(cmds []redis.Cmder) error {
			var s string
			if len(cmds) == 0 {
				s = "Executing empty pipeline"
			} else {
				s = fmt.Sprintf("Executing %d commands in pipeline:", len(cmds))
				for _, cmd := range cmds {
					s += fmt.Sprintf("\n   %s", formatCmd(cmd))
				}
			}
			logger.Debug(s)
			return p(cmds)
		}
	})

	flushNamespace := func() {
		logger := GetLogger(t)

		cl := ttnredis.New(conf)
		defer cl.Close()

		q := cl.Key("*")
		keys, err := cl.Client.Keys(q).Result()
		if err != nil {
			logger.WithField("query", q).Fatal("Failed to query Redis for keys")
			return
		}

		if len(keys) > 0 {
			n, err := cl.Client.Del(keys...).Result()
			if err != nil {
				logger.WithError(err).Fatal("Failed to delete existing keys")
				return
			}
			logger.WithField("n", n).Debug("Deleted old keys")
		}
	}

	flushNamespace()

	return cl, flushNamespace
}
