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

package healthcheck

import (
	"net/http"

	"github.com/hellofresh/health-go/v5"
	healthHttp "github.com/hellofresh/health-go/v5/checks/http"
	healthPg "github.com/hellofresh/health-go/v5/checks/postgres"
)

type defaultHealthChecker struct {
	h *health.Health
}

// AddHTTPCheck implements HealthChecker.
func (hghc *defaultHealthChecker) AddHTTPCheck(name string, url string) error {
	check := healthHttp.New(healthHttp.Config{
		URL: url,
	})
	return hghc.AddCheck(name, check)
}

// AddPgCheck implements HealthChecker.
func (hghc *defaultHealthChecker) AddPgCheck(name string, dsn string) error {
	check := healthPg.New(healthPg.Config{
		DSN: dsn,
	})
	return hghc.AddCheck(name, check)
}

// AddCheck implements HealthChecker.
func (hghc *defaultHealthChecker) AddCheck(name string, check Check) error {
	return hghc.h.Register(health.Config{
		Name:  name,
		Check: health.CheckFunc(check),
	})
}

// GetHandler implements HealthChecker.
func (hghc *defaultHealthChecker) GetHandler() http.Handler {
	return hghc.h.Handler()
}

// NewDefaultHealthChecker creates a new HealthCheker implementation using hellofresh/health-go.
func NewDefaultHealthChecker() (HealthChecker, error) {
	h, err := health.New(health.WithSystemInfo())
	if err != nil {
		return nil, err
	}
	return &defaultHealthChecker{h: h}, nil
}
