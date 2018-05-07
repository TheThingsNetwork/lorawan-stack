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

package errors

import (
	"fmt"
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestNew(t *testing.T) {
	a := assertions.New(t)

	err := New("Something went wrong")

	a.So(err.Namespace(), should.Equal, "pkg/errors")
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
	reg.Register(desc.Namespace, desc)

	all := reg.GetAll()
	a.So(all, should.HaveLength, 1)
	a.So(all[0], should.Resemble, desc)

	// duplicate ns-code combination, panic
	a.So(func() {
		reg.Register(desc.Namespace, &ErrDescriptor{
			MessageFormat: "You made a mistake cost us {price, plural, =0 {no money} =1 {one dollar} other {{price} dollars}}",
			Type:          InvalidArgument,
			Code:          391,
		})
	}, should.Panic)

	// missing code, panic
	a.So(func() {
		reg.Register(desc.Namespace, &ErrDescriptor{
			MessageFormat: "You made a mistake cost us {price, plural, =0 {no money} =1 {one dollar} other {{price} dollars}}",
			Type:          InvalidArgument,
		})
	}, should.Panic)

	// wrong namespace, panic
	a.So(func() {
		reg.Register(desc.Namespace, &ErrDescriptor{
			MessageFormat: "You made a mistake cost us {price, plural, =0 {no money} =1 {one dollar} other {{price} dollars}}",
			Type:          InvalidArgument,
			Namespace:     "foo/bar",
		})
	}, should.Panic)

	reg.Register(desc.Namespace, &ErrDescriptor{
		MessageFormat: "You made a mistake cost us {price, plural, =0 {no money} =1 {one dollar} other {{price} dollars}}",
		Type:          InvalidArgument,
		Code:          392,
	})

	a.So(reg.GetAll(), should.HaveLength, 2)
}

func ExampleSafe() {
	desc := &ErrDescriptor{
		MessageFormat: "You made a mistake cost us {price, plural, =0 {no money} =1 {one dollar} other {{price} dollars}}",
		Type:          InvalidArgument,
		Code:          391,
		Namespace:     "pkg/foo",
		SafeAttributes: []string{
			"price",
		},
	}

	desc.Register()

	err := desc.New(Attributes{
		"price": 12,
		"user":  "john-doe",
	})

	safe := Safe(err)

	fmt.Println(err.Attributes())
	fmt.Println(safe.Attributes())
}
