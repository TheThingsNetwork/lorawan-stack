// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package errorcontext_test

import (
	"context"
	"errors"
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/errorcontext"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

var err error

func ExampleErrorContext() {
	ctx, cancel := errorcontext.New(context.Background())
	defer cancel(nil)

	go func() {
		for {
			// do work
			if err != nil {
				cancel(err)
			}
		}
	}()

	for {
		select {
		// case data := <-dataChan:
		case <-ctx.Done():
			return
		}
	}
}

func TestErrorContext(t *testing.T) {
	a := assertions.New(t)

	{
		err := errors.New("foo")
		ctx, cancel := errorcontext.New(context.Background())
		cancel(err)
		select {
		case <-ctx.Done():
			a.So(ctx.Err(), should.Equal, err)
		default:
			t.Error("Context was not done")
		}

		cancel(errors.New("other"))
		<-ctx.Done()
		a.So(ctx.Err(), should.Equal, err)
	}

	{
		ctx, cancel := context.WithCancel(context.Background())
		ctx, _ = errorcontext.New(ctx)
		cancel()
		select {
		case <-ctx.Done():
			a.So(ctx.Err(), should.Equal, context.Canceled)
		default:
			t.Error("Context was not done")
		}
	}
}
