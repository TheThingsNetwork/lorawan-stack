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
	root string
}

// FromFilesystem returns an interface that fetches files from the local filesystem.
func FromFilesystem(rootElements ...string) Interface {
	root := filepath.Join(rootElements...)
	return fsFetcher{
		baseFetcher: baseFetcher{
			latency: fetchLatency.WithLabelValues("fs", root),
		},
		root: root,
	}
}

func (f fsFetcher) File(pathElements ...string) ([]byte, error) {
	if len(pathElements) == 0 {
		return nil, errFilenameNotSpecified
	}

	start := time.Now()

	p := filepath.Join(pathElements...)
	rp, err := realOSPath(f.root, p)
	if err != nil {
		return nil, err
	}
	content, err := ioutil.ReadFile(rp)
	if err == nil {
		f.observeLatency(time.Since(start))
		return content, nil
	}

	if os.IsNotExist(err) {
		return nil, errFileNotFound.WithAttributes("filename", p)
	}
	return nil, errCouldNotReadFile.WithCause(err).WithAttributes("filename", p)
}
