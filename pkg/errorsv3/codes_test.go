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

package errors

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestCodes(t *testing.T) {
	a := assertions.New(t)

	errStdLib := errors.New("go stdlib error")
	errUnknown := Define("test_codes_unknown", "")

	a.So(IsUnknown(errUnknown), should.BeTrue)
	a.So(IsUnknown(errUnknown.GRPCStatus().Err()), should.BeTrue)
	a.So(IsUnknown(errStdLib), should.BeFalse)
	a.So(IsUnknown(DefineInternal("test_codes_not_unknown", "")), should.BeFalse)

	a.So(IsCanceled(context.Canceled), should.BeTrue)
	a.So(IsDeadlineExceeded(context.DeadlineExceeded), should.BeTrue)
	a.So(IsInvalidArgument(DefineInvalidArgument("test_codes_invalid_argument", "")), should.BeTrue)
	a.So(IsNotFound(DefineNotFound("test_codes_not_found", "")), should.BeTrue)
	a.So(IsAlreadyExists(DefineAlreadyExists("test_codes_already_exists", "")), should.BeTrue)
	a.So(IsPermissionDenied(DefinePermissionDenied("test_codes_permission_denied", "")), should.BeTrue)
	a.So(IsResourceExhausted(DefineResourceExhausted("test_codes_resource_exhausted", "")), should.BeTrue)
	a.So(IsFailedPrecondition(DefineFailedPrecondition("test_codes_failed_precondition", "")), should.BeTrue)
	a.So(IsAborted(DefineAborted("test_codes_aborted", "")), should.BeTrue)
	errInternal := DefineInternal("test_codes_internal", "")
	a.So(IsInternal(errInternal), should.BeTrue)
	a.So(IsUnavailable(DefineUnavailable("test_codes_unavailable", "")), should.BeTrue)
	a.So(IsDataLoss(DefineDataLoss("test_codes_data_loss", "")), should.BeTrue)
	a.So(IsDataLoss(DefineCorruption("test_codes_corruption", "")), should.BeTrue)
	a.So(IsUnauthenticated(DefineUnauthenticated("test_codes_unauthenticated", "")), should.BeTrue)

	// Unknown errors with a non-unknown cause take the cause's code
	a.So(IsInternal(errUnknown.WithCause(errInternal)), should.BeTrue)

	a.So(HTTPStatusCode(errInternal), should.Equal, http.StatusInternalServerError)
	a.So(HTTPStatusCode(errStdLib), should.Equal, http.StatusInternalServerError)
}
