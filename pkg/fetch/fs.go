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
	"os"
	"path/filepath"
	"syscall"
	"time"

	"go.thethings.network/lorawan-stack/pkg/errors"
)

type fsFetcher struct {
	baseFetcher
}

// FromFilesystem returns an interface that fetches files from the local filesystem
func FromFilesystem(basePath string) Interface {
	basePath = filepath.Clean(basePath)
	return fsFetcher{
		baseFetcher{
			base:    basePath,
			latency: fetchLatency.WithLabelValues("fs", basePath),
		},
	}
}

func (f fsFetcher) File(pathElements ...string) ([]byte, error) {
	start := time.Now()
	content, err := ioutil.ReadFile(filepath.Join(append([]string{f.base}, pathElements...)...))
	if err == nil {
		f.observeLatency(time.Since(start))
		return content, nil
	}

	attributes := errors.Attributes{
		"filename": filepath.Join(pathElements...),
	}
	switch err := err.(type) {
	case *os.PathError:
		if errno, ok := err.Err.(syscall.Errno); ok && errno == syscall.ENOENT {
			return nil, ErrFileNotFound.New(attributes)
		}
		return nil, ErrFileFailedToOpen.NewWithCause(attributes, err)
	default:
		return nil, err
	}
}
