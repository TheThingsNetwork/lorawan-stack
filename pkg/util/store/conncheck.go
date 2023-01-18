// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

package store

import (
	"database/sql/driver"
	"fmt"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
)

type errConnCheckFailed struct {
	inner error
}

func (e *errConnCheckFailed) Error() string {
	if e.inner != nil {
		return fmt.Sprintf("connection check failed: %v", e.inner)
	}
	return "connection check failed"
}

func (e *errConnCheckFailed) Unwrap() error {
	return e.inner
}

func (*errConnCheckFailed) Is(other error) bool {
	switch {
	case other == driver.ErrBadConn: //nolint:errorlint
		return true
	default:
		return false
	}
}

var errUnexpectedRead = &errConnCheckFailed{errors.New("unexpected read")}
