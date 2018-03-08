// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

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

	a.So(ErrEOF.Describes(io.EOF), should.BeTrue)
	a.So(ErrContextCanceled.Describes(context.Canceled), should.BeTrue)
	a.So(ErrContextDeadlineExceeded.Describes(context.DeadlineExceeded), should.BeTrue)
}
