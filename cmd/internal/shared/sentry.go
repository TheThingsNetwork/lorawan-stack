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
	raven "github.com/getsentry/raven-go"
	"go.thethings.network/lorawan-stack/pkg/config"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/log/middleware/sentry"
)

// SentryMiddleware generates a log.Middleware sending errors logs to Sentry from a config.
//
// If no Sentry config was found, the function returns nil.
func SentryMiddleware(c config.ServiceBase) (log.Middleware, error) {
	if c.Sentry.DSN == "" {
		return nil, nil
	}

	s, err := raven.New(c.Sentry.DSN)
	if err != nil {
		return nil, err
	}
	s.SetIncludePaths([]string{"go.thethings.network/lorawan-stack"})

	return sentry.New(s), nil
}
