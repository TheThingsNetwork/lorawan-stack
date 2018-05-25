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

package filter

import (
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/log/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestFieldFilter(t *testing.T) {
	a := assertions.New(t)

	filter := FieldString("foo", "bar")

	a.So(filter.Filter(&test.Entry{
		F: log.Fields("foo", "bar"),
	}), should.BeTrue)

	a.So(filter.Filter(&test.Entry{
		F: log.Fields("foo", "baz"),
	}), should.BeFalse)

	a.So(filter.Filter(&test.Entry{
		F: log.Fields(),
	}), should.BeFalse)
}

func TestAndFilter(t *testing.T) {
	a := assertions.New(t)

	A := Field("foo", func(v interface{}) bool {
		if str, ok := v.(string); ok {
			return len(str) < 5
		}
		return false
	})

	B := Field("foo", func(v interface{}) bool {
		if str, ok := v.(string); ok {
			return len(str) > 2
		}
		return false
	})

	filter := And(A, B)

	a.So(A.Filter(&test.Entry{F: log.Fields("foo", "bar")}), should.BeTrue)
	a.So(B.Filter(&test.Entry{F: log.Fields("foo", "bar")}), should.BeTrue)
	a.So(filter.Filter(&test.Entry{F: log.Fields("foo", "bar")}), should.BeTrue)

	a.So(A.Filter(&test.Entry{F: log.Fields("foo", "b")}), should.BeTrue)
	a.So(B.Filter(&test.Entry{F: log.Fields("foo", "b")}), should.BeFalse)
	a.So(filter.Filter(&test.Entry{F: log.Fields("foo", "b")}), should.BeFalse)

	a.So(A.Filter(&test.Entry{F: log.Fields("foo", "barbarbar")}), should.BeFalse)
	a.So(B.Filter(&test.Entry{F: log.Fields("foo", "barbarbar")}), should.BeTrue)
	a.So(filter.Filter(&test.Entry{F: log.Fields("foo", "barbarbar")}), should.BeFalse)

	a.So(A.Filter(&test.Entry{F: log.Fields("foo", 10)}), should.BeFalse)
	a.So(B.Filter(&test.Entry{F: log.Fields("foo", 10)}), should.BeFalse)
	a.So(filter.Filter(&test.Entry{F: log.Fields("foo", 10)}), should.BeFalse)
}

func TestOrFilter(t *testing.T) {
	a := assertions.New(t)

	A := Field("foo", func(v interface{}) bool {
		if str, ok := v.(string); ok {
			return len(str) < 5
		}
		return false
	})

	B := Field("foo", func(v interface{}) bool {
		if str, ok := v.(string); ok {
			return len(str) > 2
		}
		return false
	})

	filter := Or(A, B)

	a.So(A.Filter(&test.Entry{F: log.Fields("foo", "bar")}), should.BeTrue)
	a.So(B.Filter(&test.Entry{F: log.Fields("foo", "bar")}), should.BeTrue)
	a.So(filter.Filter(&test.Entry{F: log.Fields("foo", "bar")}), should.BeTrue)

	a.So(A.Filter(&test.Entry{F: log.Fields("foo", "b")}), should.BeTrue)
	a.So(B.Filter(&test.Entry{F: log.Fields("foo", "b")}), should.BeFalse)
	a.So(filter.Filter(&test.Entry{F: log.Fields("foo", "b")}), should.BeTrue)

	a.So(A.Filter(&test.Entry{F: log.Fields("foo", "barbarbar")}), should.BeFalse)
	a.So(B.Filter(&test.Entry{F: log.Fields("foo", "barbarbar")}), should.BeTrue)
	a.So(filter.Filter(&test.Entry{F: log.Fields("foo", "barbarbar")}), should.BeTrue)

	a.So(A.Filter(&test.Entry{F: log.Fields("foo", 10)}), should.BeFalse)
	a.So(B.Filter(&test.Entry{F: log.Fields("foo", 10)}), should.BeFalse)
	a.So(filter.Filter(&test.Entry{F: log.Fields("foo", 10)}), should.BeFalse)
}
