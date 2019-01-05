// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

package errorcontext_test

import (
	"context"
	"errors"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/errorcontext"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

var err error

func ExampleErrorContext() {
	ctx, cancel := errorcontext.New(test.Context())
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
		ctx, cancel := errorcontext.New(test.Context())
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
		ctx, cancel := context.WithCancel(test.Context())
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
