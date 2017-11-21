package errors

import (
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestSafeImpl(t *testing.T) {
	a := assertions.New(t)

	desc := &ErrDescriptor{
		SafeAttributes: []string{"foo"},
	}

	i := &Impl{
		descriptor: desc,
		attributes: Attributes{
			"foo": "bar",
			"quu": "qux",
		},
	}

	a.So(i.Attributes(), should.Resemble, Attributes{
		"foo": "bar",
		"quu": "qux",
	})

	a.So((&SafeImpl{i}).Attributes(), should.Resemble, Attributes{
		"foo": "bar",
	})
}
