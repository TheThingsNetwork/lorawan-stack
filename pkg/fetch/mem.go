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
	"path"
	"time"
)

// MemFetcher represents the memory fetcher.
type memFetcher struct {
	baseFetcher
	store map[string][]byte
}

// NewMemFetcher initializes a new memory fetcher.
func NewMemFetcher(store map[string][]byte) Interface {
	return &memFetcher{
		store: store,
	}
}

// File gets content from memory.
func (f *memFetcher) File(pathElements ...string) ([]byte, error) {
	if len(pathElements) == 0 {
		return nil, errFilenameNotSpecified
	}

	start := time.Now()

	p := path.Join(pathElements...)
	content, ok := f.store[p]
	if !ok {
		return nil, errFileNotFound.WithAttributes("filename", p)
	}

	f.observeLatency(time.Since(start))
	return content, nil
}
