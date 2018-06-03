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

package marshaling

import (
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestDiff(t *testing.T) {
	a := assertions.New(t)
	old := map[string]interface{}{
		"foo": "foo",
		"bar": "bar",
		"baz": "baz",
	}
	new := map[string]interface{}{
		"foo": "baz",
		"bar": "bar",
		"qux": "qux",
	}
	a.So(Diff(new, old), should.Resemble, map[string]interface{}{
		"foo": "baz", // new value updated
		"qux": "qux", // new value added
		// bar unchanged
		"baz": nil, // old value removed
	})
}
