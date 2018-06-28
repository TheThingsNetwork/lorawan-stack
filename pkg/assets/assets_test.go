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

package assets_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/smartystreets/assertions"
	. "go.thethings.network/lorawan-stack/pkg/assets"
	"go.thethings.network/lorawan-stack/pkg/assets/templates"
	"go.thethings.network/lorawan-stack/pkg/assets/testdata"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/config"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/web"
)

type registererFunc func(s *web.Server)

func (r registererFunc) RegisterRoutes(s *web.Server) {
	r(s)
}

func TestAppHandler(t *testing.T) {
	for _, tc := range []struct {
		Port     int
		Name     string
		Config   Config
		Expected string
	}{
		{
			Port: 9185,
			Name: "Local",
			Config: Config{
				Mount:      "test",
				SearchPath: []string{"testdata"},
			},
			Expected: testdata.ExpectedAppLocal,
		},
		{
			Port: 9186,
			Name: "CDN",
			Config: Config{
				Mount: "test",
				CDN:   "https://cdn.thethings.network",
				Apps: map[string]templates.AppData{
					"app.html": {
						Title:    "Test App",
						FileName: "test.123.js",
					},
				},
			},
			Expected: testdata.ExpectedAppCDN,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			httpAddress := fmt.Sprintf("0.0.0.0:%d", tc.Port)
			c := component.MustNew(test.GetLogger(t), &component.Config{
				ServiceBase: config.ServiceBase{HTTP: config.HTTP{Listen: httpAddress}},
			})

			as, err := New(c, tc.Config)
			a.So(err, should.BeNil)

			c.RegisterWeb(registererFunc(func(s *web.Server) {
				index := as.AppHandler("app.html", tc.Name)
				group := s.Group("/test")
				group.GET("", index)
			}))

			err = c.Start()
			a.So(err, should.BeNil)

			resp, err := http.Get(fmt.Sprintf("http://%s/test", httpAddress))
			a.So(err, should.BeNil)
			a.So(resp.StatusCode, should.Equal, http.StatusOK)

			buf, err := ioutil.ReadAll(resp.Body)
			a.So(err, should.BeNil)
			a.So(string(buf), should.Equal, tc.Expected)

			c.Close()
		})
	}
}
