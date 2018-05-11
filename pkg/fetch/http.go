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

package fetch

import (
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gregjones/httpcache"
	"go.thethings.network/lorawan-stack/pkg/errors/httperrors"
)

type httpFetcher struct {
	baseURL   string
	transport *http.Client
}

func (f httpFetcher) File(pathElements ...string) ([]byte, error) {
	allElements := append([]string{f.baseURL}, pathElements...)
	url := strings.Join(allElements, "/")
	resp, err := f.transport.Get(url)
	if err != nil {
		return nil, err
	}

	if err = httperrors.FromHTTP(resp); err != nil {
		return nil, err
	}

	result, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	return result, err
}

// FromHTTP returns an object to fetch files from a webserver.
func FromHTTP(baseURL string, cache bool) Interface {
	f := httpFetcher{
		baseURL:   baseURL,
		transport: http.DefaultClient,
	}

	if !cache {
		f.transport = httpcache.NewMemoryCacheTransport().Client()
	}

	return f
}
