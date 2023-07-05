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

package blocklist_test

import (
	"testing"

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/blocklist"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestBlocklist(t *testing.T) {
	t.Parallel()

	a := assertions.New(t)

	b := blocklist.New("foo", "bar")
	bs := blocklist.Blocklists{b}

	a.So(bs.Contains("foo"), should.BeTrue)
	a.So(bs.Contains("baz"), should.BeFalse)
}

func TestGlobalBlocklist(t *testing.T) {
	t.Parallel()

	a := assertions.New(t)

	b := blocklist.New("foo", "bar")

	ctx := blocklist.NewContext(test.Context(), b)

	a.So(blocklist.Check(ctx, "root"), should.NotBeNil)
	a.So(blocklist.Check(ctx, "foo"), should.NotBeNil)
	a.So(blocklist.Check(ctx, "foobar"), should.BeNil)
}
