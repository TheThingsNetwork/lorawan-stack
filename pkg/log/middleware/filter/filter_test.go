// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package filter

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/log"
	"github.com/TheThingsNetwork/ttn/pkg/log/test"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
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
