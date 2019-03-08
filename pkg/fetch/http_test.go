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

package fetch_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/fetch"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
	httpmock "gopkg.in/jarcoal/httpmock.v1"
)

func TestHTTP(t *testing.T) {
	a := assertions.New(t)

	// Invalid path
	{
		fetcher := fetch.FromHTTP("", false)
		_, err := fetcher.File("test")
		a.So(err, should.NotBeNil)
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	serverHost := "http://server"
	content := "server content"

	httpmock.RegisterResponder("GET", fmt.Sprintf("%s/success", serverHost), httpmock.NewStringResponder(200, content))
	httpmock.RegisterResponder("GET", fmt.Sprintf("%s/fail", serverHost), httpmock.NewStringResponder(500, ""))

	fetcher := fetch.FromHTTP(serverHost, false)

	// Valid response code
	{
		receivedContent, err := fetcher.File("success")
		a.So(err, should.BeNil)
		a.So(string(receivedContent), should.Equal, content)
	}

	// Internal error response code
	{
		_, err := fetcher.File("fail")
		a.So(err, should.NotBeNil)
	}
}

func TestHTTPCache(t *testing.T) {
	a := assertions.New(t)

	cachedContent := "cached"
	nonCachedContent := "non-cached"

	servedContent := cachedContent
	s := &http.Server{
		Addr: "localhost:8083",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			now := time.Now()
			w.Header().Add("Date", now.Format(time.RFC1123))
			w.Header().Add("Expires", now.Add(5*time.Minute).Format(time.RFC1123))
			_, err := w.Write([]byte(servedContent))
			if err != nil {
				w.WriteHeader(500)
			}
			return
		}),
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,
	}
	go s.ListenAndServe()
	defer s.Close()

	time.Sleep(10 * test.Delay)

	fetcher := fetch.FromHTTP(fmt.Sprintf("http://%s", s.Addr), true)

	receivedContent, err := fetcher.File("cached")
	a.So(err, should.BeNil)
	a.So(string(receivedContent), should.Equal, cachedContent)

	servedContent = nonCachedContent
	receivedContent, err = fetcher.File("cached")
	a.So(err, should.BeNil)
	a.So(string(receivedContent), should.Equal, cachedContent)
	receivedContent, err = fetcher.File("noncached")
	a.So(err, should.BeNil)
	a.So(string(receivedContent), should.Equal, nonCachedContent)
}
