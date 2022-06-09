// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

package events

import (
	"testing"

	"github.com/smartystreets/assertions"
)

func TestUniqueStrings(t *testing.T) {
	t.Parallel()
	a := assertions.New(t)

	a.So(uniqueStrings([]string{
		"aaa", "bbb", "ccc",
	}), assertions.ShouldResemble, []string{
		"aaa", "bbb", "ccc",
	})

	a.So(uniqueStrings([]string{
		"aaa", "aaa", "bbb", "ccc",
	}), assertions.ShouldResemble, []string{
		"aaa", "bbb", "ccc",
	})

	a.So(uniqueStrings([]string{
		"aaa", "aaa", "bbb", "bbb", "bbb", "ccc",
	}), assertions.ShouldResemble, []string{
		"aaa", "bbb", "ccc",
	})

	a.So(uniqueStrings([]string{
		"aaa", "aaa", "bbb", "bbb", "bbb", "ccc", "ccc", "ccc", "ccc",
	}), assertions.ShouldResemble, []string{
		"aaa", "bbb", "ccc",
	})
}
