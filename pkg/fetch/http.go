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

package fetch

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/gregjones/httpcache"
	"go.thethings.network/lorawan-stack/pkg/errors"
)

const timeout = 10 * time.Second

type httpFetcher struct {
	baseFetcher
	httpClient *http.Client
}

func (f httpFetcher) File(pathElements ...string) ([]byte, error) {
	if len(pathElements) == 0 {
		return nil, errFilenameNotSpecified
	}

	start := time.Now()

	filename := strings.TrimLeft(path.Join(pathElements...), "/")
	url := fmt.Sprintf("%s/%s", f.base, filename)

	resp, err := f.httpClient.Get(url)
	if err != nil {
		return nil, errCouldNotFetchFile.WithCause(err).WithAttributes("filename", filename)
	}

	if err = errors.FromHTTP(resp); err != nil {
		return nil, errCouldNotFetchFile.WithCause(err).WithAttributes("filename", filename)
	}

	result, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	if err != nil {
		return nil, errCouldNotReadFile.WithCause(err).WithAttributes("filename", filename)
	}

	f.observeLatency(time.Since(start))
	return result, nil
}

// FromHTTP returns an object to fetch files from a webserver.
func FromHTTP(baseURL string, cache bool) Interface {
	baseURL = strings.TrimRight(baseURL, "/")
	transport := http.DefaultTransport
	if cache {
		transport = httpcache.NewMemoryCacheTransport()
	}
	f := httpFetcher{
		baseFetcher: baseFetcher{
			base:    baseURL,
			latency: fetchLatency.WithLabelValues("http", baseURL),
		},
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   timeout,
		},
	}
	return f
}
