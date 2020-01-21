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

package url_test

import (
	"net/url"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
	urlutil "go.thethings.network/lorawan-stack/pkg/util/url"
)

func TestURLClone(t *testing.T) {
	a := assertions.New(t)

	u := &url.URL{
		Scheme: "http",
		Host:   "localhost:1885",
		User:   url.UserPassword("foo", "bar"),
	}
	clone := urlutil.CloneURL(u)
	u.Scheme = "https"
	u.Host = "localhost:8885"
	u.User = url.UserPassword("bar", "foo")

	a.So(u.User, should.NotResemble, clone.User)
	a.So(u, should.NotResemble, clone)
}
