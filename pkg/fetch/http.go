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
	"io"
	"net/http"
	"net/url"
	"path"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
)

type httpFetcher struct {
	baseFetcher
	httpClient *http.Client
	root       *url.URL
}

func (f httpFetcher) File(pathElements ...string) ([]byte, error) {
	if len(pathElements) == 0 {
		return nil, errFilenameNotSpecified.New()
	}

	start := time.Now()

	p := path.Join(pathElements...)
	url, err := realURLPath(f.root, p)
	if err != nil {
		return nil, err
	}

	resp, err := f.httpClient.Get(url)
	if err != nil {
		return nil, errCouldNotFetchFile.WithCause(err).WithAttributes("filename", p)
	}
	if err = errors.FromHTTP(resp); err != nil {
		return nil, errCouldNotFetchFile.WithCause(err).WithAttributes("filename", p)
	}
	defer resp.Body.Close()

	result, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errCouldNotReadFile.WithCause(err).WithAttributes("filename", p)
	}
	f.observeLatency(time.Since(start))
	return result, nil
}

// FromHTTP returns an object to fetch files from a webserver.
func FromHTTP(client *http.Client, rootURL string) (Interface, error) {
	var root *url.URL
	if rootURL != "" {
		var err error
		root, err = url.Parse(rootURL)
		if err != nil {
			return nil, err
		}
		if !root.IsAbs() {
			return nil, errSchemeNotSpecified.New()
		}
	}
	return httpFetcher{
		baseFetcher: baseFetcher{
			latency: fetchLatency.WithLabelValues("http", rootURL),
		},
		httpClient: client,
		root:       root,
	}, nil
}
