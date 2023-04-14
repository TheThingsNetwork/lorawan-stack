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

package shared

import (
	"strings"

	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
)

// InitializeFallbacks initializes configuration fallbacks.
func InitializeFallbacks(conf *config.ServiceBase) error {
	// Fallback to the default Redis configuration for the cache system
	if conf.Cache.Redis.IsZero() {
		conf.Cache.Redis = conf.Redis
	}
	// Fallback to the default Redis configuration for the events system
	if conf.Events.Redis.Config.IsZero() {
		conf.Events.Redis.Config = conf.Redis
	}
	if !conf.Redis.Equals(DefaultRedisConfig) {
		// Fallback to the default Redis configuration for the cache system
		if conf.Cache.Redis.Equals(DefaultRedisConfig) {
			conf.Cache.Redis = conf.Redis
		}
		// Fallback to the default Redis configuration for the events system
		if conf.Events.Redis.Config.Equals(DefaultRedisConfig) {
			conf.Events.Redis.Config = conf.Redis
		}
	}
	return nil
}

// InitializeLogger initializes the logger.
func InitializeLogger(conf *config.Log) (log.Stack, error) {
	var (
		logHandler log.Handler
		err        error
	)
	format := strings.ToLower(conf.Format)
	switch format {
	case "json", "console":
		logHandler, err = log.NewZap(format)
		if err != nil {
			return nil, ErrInitializeLogger.WithCause(err)
		}
	default:
		return nil, ErrInvalidLogFormat.WithAttributes("format", format)
	}
	return log.NewLogger(
		logHandler,
		log.WithLevel(conf.Level),
	), nil
}
