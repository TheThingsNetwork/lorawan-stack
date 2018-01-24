// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package test

import (
	"os"
	"strconv"

	"github.com/TheThingsNetwork/ttn/pkg/config"
)

// RedisConfig returns a new redis config for testing
func RedisConfig() config.Redis {
	var err error
	config := config.Redis{
		Address:  "localhost:6379",
		Database: 1,
		Prefix:   "test",
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
		config.Prefix = prefix
	}
	return config
}
