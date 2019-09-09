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

// Package fetch offers abstractions to fetch a file with the same method,
// regardless of a location (filesystem, HTTP...).
package fetch

import (
	"fmt"
	"net/url"
	"path"
	"path/filepath"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// realPath replaces path root of p by r if r != "" and returns p otherwise.
// realPath assumes paths separated by forward slashes and assumes that both paths are cleaned.
func realPath(r, p string) (string, error) {
	if r == "" {
		return p, nil
	}
	if !path.IsAbs(p) {
		return path.Join(r, p), nil
	}
	return path.Join(r, p[1:]), nil
}

// realURLPath replaces path root of p by r if r != "" and returns p otherwise.
// realURLPath assumes URL paths and assumes that both paths are cleaned.
func realURLPath(rURL *url.URL, p string) (string, error) {
	pURL, err := url.Parse(p)
	if err != nil {
		return "", err
	}
	if rURL == nil {
		if !pURL.IsAbs() {
			return "", errSchemeNotSpecified
		}
		return p, nil
	}
	if pURL.IsAbs() {
		return "", errSchemeSpecified
	}
	if !rURL.IsAbs() {
		return "", errSchemeNotSpecified
	}
	return fmt.Sprintf("%s/%s", rURL, pURL.EscapedPath()), nil
}

// realOSPath replaces path root of p by r if r != "" and returns p otherwise.
// realOSPath assumes operating system paths and that both paths are cleaned.
func realOSPath(r, p string) (string, error) {
	if r == "" {
		return p, nil
	}
	if !filepath.IsAbs(p) {
		return filepath.Join(r, p), nil
	}
	if filepath.VolumeName(p) != "" {
		return "", errVolumeSpecified
	}
	return filepath.Join(r, p[1:]), nil
}

// Interface is an abstraction for file retrieval.
type Interface interface {
	File(pathElements ...string) ([]byte, error)
}

type baseFetcher struct {
	latency prometheus.Observer
}

func (f baseFetcher) observeLatency(d time.Duration) {
	if f.latency != nil {
		f.latency.Observe(d.Seconds())
	}
}
