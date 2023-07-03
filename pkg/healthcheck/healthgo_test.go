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

package healthcheck_test

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/smarty/assertions"

	"go.thethings.network/lorawan-stack/v3/pkg/healthcheck"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

var defaultDB = "ttn_lorawan_is_test"

func GetDSN(defaultDB string) *url.URL {
	dsn := url.URL{
		Scheme: "postgresql",
		Host:   "localhost:5432",
		Path:   defaultDB,
	}
	dsn.User = url.UserPassword("root", "root")
	query := make(url.Values)
	query.Add("sslmode", "disable")
	if dbAddress := os.Getenv("SQL_DB_ADDRESS"); dbAddress != "" {
		dsn.Host = dbAddress
	}
	if dbName := os.Getenv("TEST_DATABASE_NAME"); dbName != "" {
		dsn.Path = dbName
	}
	if dbAuth := os.Getenv("SQL_DB_AUTH"); dbAuth != "" {
		var username, password string
		idx := strings.Index(dbAuth, ":")
		if idx != -1 {
			username, password = dbAuth[:idx], dbAuth[idx+1:]
		} else {
			username = dbAuth
		}
		dsn.User = url.UserPassword(username, password)
	}
	dsn.RawQuery = query.Encode()
	return &dsn
}

func getDefaultHealthChecker(t *testing.T) healthcheck.HealthChecker {
	t.Helper()
	hc, err := healthcheck.NewDefaultHealthChecker()
	if err != nil {
		t.Fatalf("failed to create a health checker: %v", err)
	}
	return hc
}

func getHealthCheckerWithPassingCheck(t *testing.T) healthcheck.HealthChecker {
	t.Helper()
	a := assertions.New(t)
	hc := getDefaultHealthChecker(t)
	err := hc.AddCheck("test-check", func(ctx context.Context) error {
		return nil
	})
	a.So(err, should.Equal, nil)
	return hc
}

func getHealthCheckerWithFailngCheck(t *testing.T) healthcheck.HealthChecker {
	t.Helper()
	a := assertions.New(t)
	hc := getDefaultHealthChecker(t)
	err := hc.AddCheck("test-fail-check", func(ctx context.Context) error {
		return errors.New("failed")
	})
	a.So(err, should.Equal, nil)
	return hc
}

func getHealthCheckerWithPassingPgCheck(t *testing.T) healthcheck.HealthChecker {
	t.Helper()
	a := assertions.New(t)
	hc := getDefaultHealthChecker(t)
	dsn := GetDSN(defaultDB)
	err := hc.AddPgCheck("test-pg-check", dsn.String())
	a.So(err, should.Equal, nil)
	return hc
}

func getHealthCheckerWithFailingPgCheck(t *testing.T) healthcheck.HealthChecker {
	t.Helper()
	a := assertions.New(t)
	hc := getDefaultHealthChecker(t)
	dsn := url.URL{
		Scheme: "postgres",
		Host:   "localhost:5432",
		Path:   "ttn_lorawan_dev_missing_db",
	}
	dsn.User = url.UserPassword("root", "root")
	query := make(url.Values)
	query.Add("sslmode", "disable")
	dsn.RawQuery = query.Encode()
	err := hc.AddPgCheck("test-pg-fail-check", dsn.String())
	a.So(err, should.Equal, nil)
	return hc
}

func getHealthCheckerWithPassingHTTPCheck(t *testing.T) healthcheck.HealthChecker {
	t.Helper()
	a := assertions.New(t)
	hc := getDefaultHealthChecker(t)
	err := hc.AddHTTPCheck("test-http-check", "http://localhost:3324")
	a.So(err, should.Equal, nil)
	return hc
}

func getHealthCheckerWithFailingHTTPCheck(t *testing.T) healthcheck.HealthChecker {
	t.Helper()
	a := assertions.New(t)
	hc := getDefaultHealthChecker(t)
	err := hc.AddHTTPCheck("test-http-fail-check", "bad_addr")
	a.So(err, should.Equal, nil)
	return hc
}

func getServerWithHandler(t *testing.T, hc healthcheck.HealthChecker, addr string) http.Server {
	t.Helper()
	r := mux.NewRouter()
	r.Handle("/healthz", hc.GetHandler())
	return http.Server{
		Addr:              addr,
		Handler:           r,
		ReadHeaderTimeout: time.Second,
	}
}

// nolint:gosec
func assertGetStatusCode(a *assertions.Assertion, requestURL string, code int) {
	resp, err := http.Get(requestURL)
	a.So(err, should.Equal, nil)
	a.So(resp.StatusCode, should.Equal, code)
	resp.Body.Close()
}

func TestHealthCheckerWithPassingCheck(t *testing.T) {
	t.Parallel()
	a, _ := test.New(t)
	hc := getHealthCheckerWithPassingCheck(t)
	srv := getServerWithHandler(t, hc, ":3320")
	go func() {
		_ = srv.ListenAndServe()
	}()
	assertGetStatusCode(a, "http://localhost:3320/healthz", 200)
}

func TestHealthCheckerWithFailingCheck(t *testing.T) {
	t.Parallel()
	a, _ := test.New(t)
	hc := getHealthCheckerWithFailngCheck(t)
	srv := getServerWithHandler(t, hc, ":3321")
	go func() {
		_ = srv.ListenAndServe()
	}()
	assertGetStatusCode(a, "http://localhost:3321/healthz", 503)
}

func TestHealthCheckerWithPassingPgCheck(t *testing.T) {
	t.Parallel()
	a, _ := test.New(t)
	hc := getHealthCheckerWithPassingPgCheck(t)
	srv := getServerWithHandler(t, hc, ":3322")
	go func() {
		_ = srv.ListenAndServe()
	}()
	assertGetStatusCode(a, "http://localhost:3322/healthz", 200)
}

func TestHealthCheckerWithFailingPgCheck(t *testing.T) {
	t.Parallel()
	a, _ := test.New(t)
	hc := getHealthCheckerWithFailingPgCheck(t)
	srv := getServerWithHandler(t, hc, ":3323")
	go func() {
		_ = srv.ListenAndServe()
	}()
	assertGetStatusCode(a, "http://localhost:3323/healthz", 503)
}

func TestHealthCheckerWithPassingHTTPCheck(t *testing.T) {
	t.Parallel()
	a, _ := test.New(t)
	hc := getHealthCheckerWithPassingHTTPCheck(t)
	srv := getServerWithHandler(t, hc, ":3324")
	go func() {
		_ = srv.ListenAndServe()
	}()
	assertGetStatusCode(a, "http://localhost:3324/healthz", 200)
}

func TestHealthCheckerWithFailingHTTPCheck(t *testing.T) {
	t.Parallel()
	a, _ := test.New(t)
	hc := getHealthCheckerWithFailingHTTPCheck(t)
	srv := getServerWithHandler(t, hc, ":3325")
	go func() {
		_ = srv.ListenAndServe()
	}()
	assertGetStatusCode(a, "http://localhost:3325/healthz", 503)
}
