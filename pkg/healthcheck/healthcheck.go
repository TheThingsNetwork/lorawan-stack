// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

// Package healthcheck provides configuration of startup probes.
package healthcheck

import (
	"context"
	"net/http"
)

// Check is a function that determines the health of the HealtChecker.
type Check func(ctx context.Context) error

// HealthChecker manages checks for determining the healhiness of a component.
type HealthChecker interface {
	AddHTTPCheck(name string, url string) error
	AddPgCheck(name string, dsn string) error
	AddCheck(name string, check Check) error
	GetHandler() http.Handler
}
