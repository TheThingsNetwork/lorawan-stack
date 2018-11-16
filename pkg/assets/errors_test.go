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

	"github.com/labstack/echo"
	"github.com/smartystreets/assertions"
	. "go.thethings.network/lorawan-stack/pkg/assets"
	"go.thethings.network/lorawan-stack/pkg/assets/testdata"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/config"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/web"
)

var errTest = errors.DefineInternal("test", "test error")

func TestErrors(t *testing.T) {
	a := assertions.New(t)

	httpAddress := "0.0.0.0:9187"
	c := component.MustNew(test.GetLogger(t), &component.Config{
		ServiceBase: config.ServiceBase{HTTP: config.HTTP{Listen: httpAddress}},
	})

	as, err := New(c, Config{
		Mount:      "/test",
		SearchPath: []string{"testdata"},
	})
	a.So(err, should.BeNil)

	c.RegisterWeb(registererFunc(func(s *web.Server) {
		middleware := as.Errors("error.html", nil)
		group := s.Group("/error")
		group.GET("", func(c echo.Context) error {
			return errTest
		}, middleware)
	}))

	err = c.Start()
	a.So(err, should.BeNil)

	for _, tc := range []struct {
		Accept string
		Test   func(a *assertions.Assertion, buf []byte)
	}{
		{
			Accept: "text/html",
			Test: func(a *assertions.Assertion, buf []byte) {
				a.So(string(buf), should.Equal, testdata.ExpectedErrorTemplated)
			},
		},
		{
			Accept: "application/json",
			Test: func(a *assertions.Assertion, buf []byte) {
				var returnErr errors.Error
				err := returnErr.UnmarshalJSON(buf)
				a.So(err, should.BeNil)
				a.So(returnErr, should.HaveSameErrorDefinitionAs, errTest)
			},
		},
		{
			Accept: "text/plain",
			Test: func(a *assertions.Assertion, buf []byte) {
				a.So(string(buf), should.Equal, errTest.String())
			},
		},
	} {
		t.Run(tc.Accept, func(t *testing.T) {
			a := assertions.New(t)

			req, err := http.NewRequest("GET", fmt.Sprintf("http://%s/error", httpAddress), nil)
			a.So(err, should.BeNil)

			req.Header.Set("Accept", tc.Accept)

			resp, err := http.DefaultClient.Do(req)
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}
			a.So(resp.StatusCode, should.Equal, http.StatusInternalServerError)

			buf, err := ioutil.ReadAll(resp.Body)
			a.So(err, should.BeNil)

			tc.Test(a, buf)
		})
	}

	c.Close()
}
