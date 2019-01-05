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
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Interface is an abstraction for file retrieval.
type Interface interface {
	File(pathElements ...string) ([]byte, error)
}

type baseFetcher struct {
	base    string
	latency prometheus.Observer
}

func (f baseFetcher) observeLatency(d time.Duration) {
	if f.latency != nil {
		f.latency.Observe(d.Seconds())
	}
}
