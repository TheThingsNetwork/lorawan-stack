// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package fetch

import (
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/TheThingsNetwork/ttn/pkg/errors/httperrors"
	"github.com/gregjones/httpcache"
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
