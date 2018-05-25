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

package grpcerrors

import (
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
	errshould "go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

func TestGRPC(t *testing.T) {
	a := assertions.New(t)
	d := &errors.ErrDescriptor{
		MessageFormat: "You do not have access to application `{app_id}`",
		Code:          77,
		Type:          errors.PermissionDenied,
		Namespace:     "pkg/foo",
		SafeAttributes: []string{
			"app_id",
			"count",
		},
	}
	d.Register()

	attributes := errors.Attributes{
		"app_id": "foo",
		"count":  42,
		"unsafe": "secret",
	}

	err := d.New(attributes)

	code := GRPCCode(err)
	a.So(code, should.Equal, codes.PermissionDenied)

	// other errors should be unknown
	other := fmt.Errorf("Foo")
	code = GRPCCode(other)
	a.So(code, should.Equal, codes.Unknown)

	grpcErr := ToGRPC(err)

	got := FromGRPC(grpcErr)
	a.So(got.Code(), should.Equal, d.Code)
	a.So(got.Type(), should.Equal, d.Type)
	a.So(got.Message(), should.Equal, "You do not have access to application `foo`")
	a.So(got.Error(), should.Equal, "pkg/foo[77]: You do not have access to application `foo`")
	a.So(got.ID(), should.Equal, err.ID())

	a.So(got.Attributes(), should.NotBeEmpty)
	a.So(got.Attributes()["app_id"], should.Resemble, attributes["app_id"])
	a.So(got.Attributes()["count"], should.AlmostEqual, attributes["count"])
	a.So(got.Attributes(), should.NotContainKey, "unsafe")
}

func TestFromUnspecifiedGRPC(t *testing.T) {
	a := assertions.New(t)

	err := grpc.Errorf(codes.DeadlineExceeded, "This is an error")

	got := FromGRPC(err)
	a.So(got.Code(), should.Equal, errors.NoCode)
	a.So(got.Type(), should.Equal, errors.Timeout)
	a.So(got.Error(), should.Equal, "This is an error")
	a.So(got.Attributes(), should.BeNil)
	a.So(got.ID(), should.NotBeEmpty)
}

func TestWellKnown(t *testing.T) {
	a := assertions.New(t)

	a.So(FromGRPC(ToGRPC(io.EOF)), errshould.Describe, errors.ErrEOF)
	a.So(FromGRPC(ToGRPC(context.Canceled)), errshould.Describe, errors.ErrContextCanceled)
	a.So(FromGRPC(ToGRPC(context.DeadlineExceeded)), errshould.Describe, errors.ErrContextDeadlineExceeded)
	a.So(FromGRPC(ToGRPC(grpc.ErrClientConnClosing)), errshould.Describe, ErrClientConnClosing)
	a.So(FromGRPC(ToGRPC(grpc.ErrClientConnTimeout)), errshould.Describe, ErrClientConnTimeout)
	a.So(FromGRPC(ToGRPC(grpc.ErrServerStopped)), errshould.Describe, ErrServerStopped)

	a.So(FromGRPC(io.EOF), errshould.Describe, errors.ErrEOF)
	a.So(FromGRPC(context.Canceled), errshould.Describe, errors.ErrContextCanceled)
	a.So(FromGRPC(context.DeadlineExceeded), errshould.Describe, errors.ErrContextDeadlineExceeded)
	a.So(FromGRPC(grpc.ErrClientConnClosing), errshould.Describe, ErrClientConnClosing)
	a.So(FromGRPC(grpc.ErrClientConnTimeout), errshould.Describe, ErrClientConnTimeout)
	a.So(FromGRPC(grpc.ErrServerStopped), errshould.Describe, ErrServerStopped)
}
