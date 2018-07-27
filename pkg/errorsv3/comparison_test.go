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

package errors_test

import (
	"testing"

	"github.com/smartystreets/assertions"
	errors "go.thethings.network/lorawan-stack/pkg/errorsv3"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestResembles(t *testing.T) {
	a := assertions.New(t)

	a.So(errors.Resemble(nil, nil), should.BeTrue)

	defInvalidArgument := errors.DefineInvalidArgument("test_resembles_invalid_argument", "invalid argument")

	// Nil errors never resemble.
	a.So(errors.Resemble(defInvalidArgument, nil), should.BeFalse)
	a.So(errors.Resemble(nil, defInvalidArgument), should.BeFalse)

	// Typed nil is invalid.
	a.So(errors.Resemble(defInvalidArgument, (*errors.Definition)(nil)), should.BeFalse)
	a.So(errors.Resemble(defInvalidArgument, (*errors.Error)(nil)), should.BeFalse)

	errInvalidArgument := defInvalidArgument.WithAttributes("foo", "bar")
	grpcErrInvalidArgument := errInvalidArgument.GRPCStatus().Err()

	// Errors and definitions resemble all pointer/non-pointer combinations:
	a.So(errors.Resemble(errInvalidArgument, defInvalidArgument), should.BeTrue)
	a.So(errors.Resemble(errInvalidArgument, &defInvalidArgument), should.BeTrue)
	a.So(errors.Resemble(&errInvalidArgument, defInvalidArgument), should.BeTrue)
	a.So(errors.Resemble(&errInvalidArgument, &defInvalidArgument), should.BeTrue)

	// Should resemble gRPC error:
	a.So(errors.Resemble(grpcErrInvalidArgument, defInvalidArgument), should.BeTrue)
	a.So(errors.Resemble(grpcErrInvalidArgument, errInvalidArgument), should.BeTrue)

	defPermissionDenied := errors.DefinePermissionDenied("test_resembles_permission_denied", "permission denied")

	a.So(errors.Resemble(defInvalidArgument, defPermissionDenied), should.BeFalse)

}
