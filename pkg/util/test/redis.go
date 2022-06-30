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
	"context"
	"crypto/rand"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"testing"

	ulid "github.com/oklog/ulid/v2"
	"github.com/redis/go-redis/v9"
	ttnredis "go.thethings.network/lorawan-stack/v3/pkg/redis"
)

const (
	defaultDatabase = 1
	defaultAddress  = "localhost:6379"
)

var defaultNamespace = [...]string{
	"redistest",
}

type redisHook struct {
	testing.TB
}

func (redisHook) formatCommand(cmd redis.Cmder) string {
	ss := make([]string, 0, len(cmd.Args()))
	for _, arg := range cmd.Args() {
		ss = append(ss, fmt.Sprint(arg))
	}
	return strings.Join(ss, " ")
}

func (h redisHook) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	GetLogger(h.TB).Debugf("Executing `%s`", h.formatCommand(cmd))
	return ctx, nil
}

func (h redisHook) AfterProcess(context.Context, redis.Cmder) error {
	return nil
}

func (h redisHook) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	printLog := GetLogger(h.TB).Debug
	if len(cmds) == 0 {
		printLog("Executing empty pipeline")
	} else {
		s := fmt.Sprintf("Executing %d commands in pipeline:", len(cmds))
		for _, cmd := range cmds {
			s += fmt.Sprintf("\n   %s", h.formatCommand(cmd))
		}
		printLog(s)
	}
	return ctx, nil
}

func (h redisHook) AfterProcessPipeline(context.Context, []redis.Cmder) error {
	return nil
}

// NewRedis returns a new namespaced *redis.Client ready to use
// and a flush function, which should be called after the client is not needed anymore to clean the namespace.
// NewRedis respects TEST_REDIS, REDIS_ADDRESS and REDIS_DB environment variables.
// Client returned logs commands executed.
func NewRedis(ctx context.Context, namespace ...string) (*ttnredis.Client, func()) {
	t := MustTBFromContext(ctx)
	if os.Getenv("TEST_REDIS") != "1" {
		t.Skip("TEST_REDIS is not set to `1`, skipping Redis tests")
	}

	conf := ttnredis.Config{
		Address:       defaultAddress,
		Database:      defaultDatabase,
		RootNamespace: defaultNamespace[:],
		// CI has at most 1 virtual CPU available, resulting in a default pool size of 10.
		// Tests that require more than 10 concurrent connections, such as the ones which
		// subscribe to messages, will fail since no connection will be available.
		PoolSize: 32 * runtime.NumCPU(),
	}.WithNamespace(append(append([]string{ulid.MustNew(ulid.Now(), rand.Reader).String()}, namespace...), t.Name())...)

	if addr := os.Getenv("REDIS_ADDRESS"); addr != "" {
		conf.Address = addr
	}
	if db := os.Getenv("REDIS_DB"); db != "" {
		var err error
		conf.Database, err = strconv.Atoi(db)
		if err != nil {
			t.Fatalf("Expected REDIS_DB to be an integer, got `%s`", db)
			return nil, nil
		}
	}

	cl := ttnredis.New(conf)
	if err := cl.Ping(ctx).Err(); err != nil {
		t.Fatalf("Failed to ping Redis: `%s`", err)
	}

	cl.Client.AddHook(redisHook{
		TB: t,
	})

	flushNamespace := func() {
		logger := GetLogger(t)

		cl := ttnredis.New(conf)
		defer cl.Close()

		q := cl.Key("*")
		keys, err := cl.Client.Keys(ctx, q).Result()
		if err != nil {
			logger.WithError(err).WithField("query", q).Fatal("Failed to query Redis for keys")
			return
		}

		if len(keys) > 0 {
			n, err := cl.Client.Del(ctx, keys...).Result()
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
