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
	"context"
	"path/filepath"
	"time"

	"gocloud.dev/blob"
	"gocloud.dev/gcerrors"
)

type bucketFetcher struct {
	baseFetcher
	bucket *blob.Bucket
}

// FromBucket returns an interface that fetches files from the given blob bucket.
func FromBucket(bucket *blob.Bucket, basePath string) Interface {
	return &bucketFetcher{
		baseFetcher: baseFetcher{
			base:    basePath,
			latency: fetchLatency.WithLabelValues("bucket", basePath),
		},
		bucket: bucket,
	}
}

func (f *bucketFetcher) File(pathElements ...string) ([]byte, error) {
	if len(pathElements) == 0 {
		return nil, errFilenameNotSpecified
	}

	start := time.Now()
	content, err := f.bucket.ReadAll(context.TODO(), filepath.Join(append([]string{f.base}, pathElements...)...))
	if err == nil {
		f.observeLatency(time.Since(start))
		return content, nil
	}

	if gcerrors.Code(err) == gcerrors.NotFound {
		return nil, errFileNotFound.WithAttributes("filename", filepath.Join(pathElements...))
	}
	return nil, errCouldNotReadFile.WithCause(err).WithAttributes("filename", filepath.Join(pathElements...))
}
