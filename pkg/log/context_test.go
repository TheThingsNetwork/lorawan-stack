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

	a.So(FromContextMaybe(ctx), should.BeNil)
	a.So(func() {
		_ = FromContext(ctx)
	}, should.Panic)

	ctx = WithLogger(ctx, logger)

	a.So(FromContextMaybe(ctx), should.NotBeNil)
	a.So(func() {
		_ = FromContext(ctx)
	}, should.NotPanic)
}
