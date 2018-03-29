// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package log

import (
	"context"
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestContext(t *testing.T) {
	a := assertions.New(t)
	ctx := context.Background()

	logger, err := NewLogger()
	a.So(err, should.BeNil)

	a.So(FromContext(ctx), should.NotBeNil)
	a.So(FromContext(ctx), should.Equal, Noop)
	a.So(func() { Must(FromContext(ctx)) }, should.NotPanic)

	ctx = NewContext(ctx, logger)

	a.So(FromContext(ctx), should.Equal, logger)
	a.So(func() { Must(FromContext(ctx)) }, should.NotPanic)

	t.Run("NewContextWithField", func(t *testing.T) {
		a := assertions.New(t)
		withKV := FromContext(NewContextWithField(ctx, "key", "value")).(*entry)
		v, ok := withKV.fields.Get("key")
		a.So(ok, should.BeTrue)
		a.So(v, should.Equal, "value")
	})

	t.Run("NewContextWithFields", func(t *testing.T) {
		a := assertions.New(t)
		withKV := FromContext(NewContextWithFields(ctx, Fields("key", "value"))).(*entry)
		v, ok := withKV.fields.Get("key")
		a.So(ok, should.BeTrue)
		a.So(v, should.Equal, "value")
	})

}
