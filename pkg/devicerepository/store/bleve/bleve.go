// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

package bleve

import (
	"context"
	"path"
	"time"

	"github.com/blevesearch/bleve"
	"github.com/bluele/gcache"
	"go.thethings.network/lorawan-stack/v3/pkg/devicerepository/store"
	"go.thethings.network/lorawan-stack/v3/pkg/devicerepository/store/remote"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/fetch"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
)

const (
	// defaultTimeout is the timeout when trying to open the index. This is to avoid
	// blocking on the index open call, which hangs indefinitely if the index is
	// already in use by a different process.
	defaultTimeout = 5 * time.Second

	// cacheSize is the size of the cache for brands and models.
	cacheSize = 1024
)

// bleveStore wraps a store.Store adding support for searching/sorting results using a bleve index.
type bleveStore struct {
	ctx context.Context

	store store.Store
	index bleve.Index
	cache gcache.Cache
}

// NewStore returns a new Device Repository store with indexing capabilities (using bleve).
func (c Config) NewStore(ctx context.Context) (store.Store, error) {
	wd, err := getWorkingDirectory(c.SearchPaths)
	if err != nil {
		return nil, err
	}
	s := &bleveStore{
		ctx:   ctx,
		store: remote.NewRemoteStore(fetch.FromFilesystem(wd)),

		cache: gcache.New(cacheSize).LFU().Build(),
	}

	ctx, cancel := context.WithTimeout(s.ctx, defaultTimeout)
	defer cancel()
	s.index, err = openIndex(ctx, path.Join(wd, indexPath))
	if err != nil {
		return nil, err
	}
	go func() {
		<-s.ctx.Done()
		if err := s.Close(); err != nil {
			log.FromContext(ctx).WithError(err).Warn("Failed to close index")
		}
	}()

	return s, nil
}

var errCannotOpenIndex = errors.DefineNotFound("cannot_open_index", "cannot open index")

func openIndex(ctx context.Context, path string) (bleve.Index, error) {
	log.FromContext(ctx).WithField("path", path).Debug("Loading index")
	index, err := bleve.OpenUsing(path, map[string]any{"read_only": true, "bolt_timeout": "60s"})
	if err != nil {
		return nil, errCannotOpenIndex.WithAttributes("path", path).WithCause(err)
	}
	return index, nil
}
