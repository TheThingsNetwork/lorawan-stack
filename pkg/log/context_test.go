// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

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

	a.So(FromContext(ctx), should.BeNil)
	a.So(Ensure(FromContext(ctx)), should.Equal, Noop)
	a.So(func() { Must(FromContext(ctx)) }, should.Panic)

	ctx = WithLogger(ctx, logger)

	a.So(FromContext(ctx), should.Equal, logger)
	a.So(Ensure(FromContext(ctx)), should.Equal, logger)
	a.So(func() { Must(FromContext(ctx)) }, should.NotPanic)
}
