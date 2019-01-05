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

package blacklist_test

import (
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/identityserver/blacklist"
	"go.thethings.network/lorawan-stack/pkg/util/test"
)

func TestBlacklist(t *testing.T) {
	a := assertions.New(t)

	b := blacklist.New("foo", "bar")
	bs := blacklist.Blacklists{b}

	a.So(bs.Contains("foo"), should.BeTrue)
	a.So(bs.Contains("baz"), should.BeFalse)
}

func TestGlobalBlacklist(t *testing.T) {
	a := assertions.New(t)

	b := blacklist.New("foo", "bar")

	ctx := blacklist.NewContext(test.Context(), b)

	a.So(blacklist.Check(ctx, "root"), should.NotBeNil)
	a.So(blacklist.Check(ctx, "foo"), should.NotBeNil)
	a.So(blacklist.Check(ctx, "foobar"), should.BeNil)
}
