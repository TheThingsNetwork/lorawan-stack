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

package errors_test

import (
	"context"
	gerrors "errors"
	"net/http"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestCodes(t *testing.T) {
	a := assertions.New(t)

	errStdLib := gerrors.New("go stdlib error")
	errUnknown := errors.Define("test_codes_unknown", "")

	a.So(errors.IsUnknown(errUnknown), should.BeTrue)
	a.So(errors.IsUnknown(errUnknown.GRPCStatus().Err()), should.BeTrue)
	a.So(errors.IsUnknown(errStdLib), should.BeFalse)
	a.So(errors.IsUnknown(errors.DefineInternal("test_codes_not_unknown", "")), should.BeFalse)

	a.So(errors.IsCanceled(context.Canceled), should.BeTrue)
	a.So(errors.IsDeadlineExceeded(context.DeadlineExceeded), should.BeTrue)
	a.So(errors.IsInvalidArgument(errors.DefineInvalidArgument("test_codes_invalid_argument", "")), should.BeTrue)
	a.So(errors.IsNotFound(errors.DefineNotFound("test_codes_not_found", "")), should.BeTrue)
	a.So(errors.IsAlreadyExists(errors.DefineAlreadyExists("test_codes_already_exists", "")), should.BeTrue)
	a.So(errors.IsPermissionDenied(errors.DefinePermissionDenied("test_codes_permission_denied", "")), should.BeTrue)
	a.So(errors.IsResourceExhausted(errors.DefineResourceExhausted("test_codes_resource_exhausted", "")), should.BeTrue)
	a.So(errors.IsFailedPrecondition(errors.DefineFailedPrecondition("test_codes_failed_precondition", "")), should.BeTrue)
	a.So(errors.IsAborted(errors.DefineAborted("test_codes_aborted", "")), should.BeTrue)
	errInternal := errors.DefineInternal("test_codes_internal", "")
	a.So(errors.IsInternal(errInternal), should.BeTrue)
	a.So(errors.IsUnavailable(errors.DefineUnavailable("test_codes_unavailable", "")), should.BeTrue)
	a.So(errors.IsDataLoss(errors.DefineDataLoss("test_codes_data_loss", "")), should.BeTrue)
	a.So(errors.IsDataLoss(errors.DefineCorruption("test_codes_corruption", "")), should.BeTrue)
	a.So(errors.IsUnauthenticated(errors.DefineUnauthenticated("test_codes_unauthenticated", "")), should.BeTrue)

	// Unknown errors with a non-unknown cause take the cause's code
	a.So(errors.IsInternal(errUnknown.WithCause(errInternal)), should.BeTrue)

	a.So(errors.ToHTTPStatusCode(errInternal), should.Equal, http.StatusInternalServerError)
	a.So(errors.ToHTTPStatusCode(errStdLib), should.Equal, http.StatusInternalServerError)
}
