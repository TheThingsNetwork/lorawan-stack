// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package errors

import (
	"context"
	"io"
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestWellKnown(t *testing.T) {
	a := assertions.New(t)

	a.So(ErrEOF.Is(io.EOF), should.BeTrue)
	a.So(ErrContextCanceled.Is(context.Canceled), should.BeTrue)
	a.So(ErrContextDeadlineExceeded.Is(context.DeadlineExceeded), should.BeTrue)
}
