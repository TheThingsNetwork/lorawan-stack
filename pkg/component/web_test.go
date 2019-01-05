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

package component

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/config"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

const (
	httpAddress     = "0.0.0.0:8097"
	metricsPassword = "secret-metrics-test-password"
	pprofPassword   = "secret-pprof-test-password"
)

func TestPProf(t *testing.T) {
	a := assertions.New(t)

	config := &Config{
		ServiceBase: config.ServiceBase{
			HTTP: config.HTTP{
				Listen: httpAddress,
				Metrics: config.Metrics{
					Enable:   true,
					Password: metricsPassword,
				},
				PProf: config.PProf{
					Enable:   true,
					Password: pprofPassword,
				},
			},
		},
	}
	c, err := New(test.GetLogger(t), config)
	a.So(err, should.BeNil)

	err = c.listenWeb()
	a.So(err, should.BeNil)

	client := &http.Client{}

	for _, tc := range []struct {
		path               string
		username, password string
	}{
		{
			path:     "/debug/pprof",
			username: pprofUsername,
			password: pprofPassword,
		},
		{
			path:     "metrics",
			username: metricsUsername,
			password: metricsPassword,
		},
	} {
		t.Run(fmt.Sprintf("%s endpoint", tc.path), func(t *testing.T) {
			a := assertions.New(t)

			url := fmt.Sprintf("http://%s/%s", httpAddress, tc.path)
			res, err := client.Get(url)
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}
			a.So(res.StatusCode, should.BeIn, []int{401, 403})

			req, err := http.NewRequest("GET", url, nil)
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}
			req.SetBasicAuth(tc.username, tc.password)
			res, err = client.Do(req)
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}
			a.So(res.StatusCode, should.BeBetweenOrEqual, 200, 299)
		})
	}
}
