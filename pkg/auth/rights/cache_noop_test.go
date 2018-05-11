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

package rights

import (
	"errors"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

var _ cache = new(noopCache)

func TestNoopCache(t *testing.T) {
	a := assertions.New(t)
	cache := new(noopCache)

	rights, err := cache.GetOrFetch("key", func() ([]ttnpb.Right, error) {
		return []ttnpb.Right{ttnpb.Right(0)}, nil
	})
	a.So(err, should.BeNil)
	a.So(rights, should.HaveLength, 1)
	a.So(rights, should.Contain, ttnpb.Right(0))

	time.Sleep(200 * time.Millisecond)

	rights, err = cache.GetOrFetch("key", func() ([]ttnpb.Right, error) {
		return []ttnpb.Right{ttnpb.Right(1)}, nil
	})
	a.So(err, should.BeNil)
	a.So(rights, should.HaveLength, 1)
	a.So(rights, should.Contain, ttnpb.Right(1))

	var ErrFetch = errors.New("error")
	rights, err = cache.GetOrFetch("key", func() ([]ttnpb.Right, error) {
		return nil, ErrFetch
	})
	a.So(err, should.NotBeNil)
	a.So(err, should.Equal, ErrFetch)
	a.So(rights, should.BeNil)
}
