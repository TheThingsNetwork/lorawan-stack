// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package errors

import (
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestNew(t *testing.T) {
	a := assertions.New(t)

	err := New("Something went wrong")

	a.So(err.Namespace(), should.Equal, "errors")
}

func TestRegistry(t *testing.T) {
	a := assertions.New(t)

	reg := &registry{
		byNamespaceAndCode: make(map[string]map[Code]*ErrDescriptor),
	}

	a.So(reg.GetAll(), should.BeEmpty)

	desc := &ErrDescriptor{
		MessageFormat: "You made a mistake cost us {price, plural, =0 {no money} =1 {one dollar} other {{price} dollars}}",
		Type:          InvalidArgument,
		Code:          391,
		Namespace:     "pkg/foo",
	}
	reg.Register(desc)

	all := reg.GetAll()
	a.So(all, should.HaveLength, 1)
	a.So(all[0], should.Resemble, desc)

	// duplicate ns-code combination, panic
	a.So(func() {
		reg.Register(&ErrDescriptor{
			MessageFormat: "You made a mistake cost us {price, plural, =0 {no money} =1 {one dollar} other {{price} dollars}}",
			Type:          InvalidArgument,
			Code:          391,
			Namespace:     "pkg/foo",
		})
	}, should.Panic)

	// missing code, panic
	a.So(func() {
		reg.Register(&ErrDescriptor{
			MessageFormat: "You made a mistake cost us {price, plural, =0 {no money} =1 {one dollar} other {{price} dollars}}",
			Type:          InvalidArgument,
			Namespace:     "pkg/foo",
		})
	}, should.Panic)

	// missing namespace, panic
	a.So(func() {
		reg.Register(&ErrDescriptor{
			MessageFormat: "You made a mistake cost us {price, plural, =0 {no money} =1 {one dollar} other {{price} dollars}}",
			Type:          InvalidArgument,
			Code:          392,
		})
	}, should.Panic)

	reg.Register(&ErrDescriptor{
		MessageFormat: "You made a mistake cost us {price, plural, =0 {no money} =1 {one dollar} other {{price} dollars}}",
		Type:          InvalidArgument,
		Code:          392,
		Namespace:     "pkg/foo",
	})

	a.So(reg.GetAll(), should.HaveLength, 2)
}
