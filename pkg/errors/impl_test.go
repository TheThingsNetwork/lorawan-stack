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
