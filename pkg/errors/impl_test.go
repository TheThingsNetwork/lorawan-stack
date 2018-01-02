// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package errors

import (
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

var (
	_ Error = &Impl{}
	_ Error = &SafeImpl{}
)

func TestSafeImpl(t *testing.T) {
	a := assertions.New(t)

	desc := &ErrDescriptor{
		Type:           Unauthorized,
		Code:           Code(33),
		SafeAttributes: []string{"foo"},
		Namespace:      "ns",
		registered:     true,
	}

	i := desc.New(Attributes{
		"foo": "bar",
		"quu": "qux",
	})

	a.So(i.Attributes(), should.Resemble, Attributes{
		"foo": "bar",
		"quu": "qux",
	})

	safe := Safe(i)

	a.So(safe.Attributes(), should.Resemble, Attributes{
		"foo": "bar",
	})

	a.So(safe.Code(), should.Resemble, desc.Code)
	a.So(safe.Type(), should.Resemble, desc.Type)
	a.So(safe.Namespace(), should.Resemble, desc.Namespace)
	a.So(safe.ID(), should.Resemble, i.ID())
}
