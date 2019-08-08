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
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

type fsFetcher struct {
	baseFetcher
}

// FromFilesystem returns an interface that fetches files from the local filesystem
func FromFilesystem(basePath string) Interface {
	if basePath != "" {
		basePath = filepath.Clean(basePath)
	}
	return fsFetcher{
		baseFetcher{
			base:    basePath,
			latency: fetchLatency.WithLabelValues("fs", basePath),
		},
	}
}

func (f fsFetcher) File(pathElements ...string) ([]byte, error) {
	start := time.Now()
	var path string
	if f.base != "" {
		path = filepath.Join(append([]string{f.base}, pathElements...)...)
	} else {
		path = filepath.Join(pathElements...)
	}
	content, err := ioutil.ReadFile(path)
	if err == nil {
		f.observeLatency(time.Since(start))
		return content, nil
	}

	if os.IsNotExist(err) {
		return nil, errFileNotFound.WithAttributes("filename", filepath.Join(pathElements...))
	}
	return nil, errCouldNotReadFile.WithCause(err).WithAttributes("filename", filepath.Join(pathElements...))
}
