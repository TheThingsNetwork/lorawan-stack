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

package nats

import (
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestCombineSubjects(t *testing.T) {
	a := assertions.New(t)

	for _, tc := range []struct {
		name     string
		subject1 string
		subject2 string
		expected string
	}{
		{
			name:     "EmptySubject1",
			subject1: "",
			subject2: "bar.bar2",
			expected: "bar.bar2",
		},
		{
			name:     "EmptySubject2",
			subject1: "foo.foo2",
			subject2: "",
			expected: "foo.foo2",
		},
		{
			name:     "BothProvided",
			subject1: "foo.foo2",
			subject2: "bar.bar2",
			expected: "foo.foo2.bar.bar2",
		},
		{
			name:     "NoneProvided",
			subject1: "",
			subject2: "",
			expected: "",
		},
		{
			name:     "Trailing1",
			subject1: "foo.",
			subject2: "",
			expected: "foo",
		},
		{
			name:     "Trailing2",
			subject1: "foo.",
			subject2: ".bar",
			expected: "foo.bar",
		},
		{
			name:     "Trailing3",
			subject1: ".foo.test.",
			subject2: ".bar.",
			expected: "foo.test.bar",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a.So(combineSubjects(tc.subject1, tc.subject2), should.Equal, tc.expected)
		})
	}
}
