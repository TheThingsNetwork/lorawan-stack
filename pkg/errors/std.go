// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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
	stderrors "errors"
)

// Alias standard library error functions.
var (
	As     = stderrors.As
	Is     = stderrors.Is
	Unwrap = stderrors.Unwrap
)

// Unwrap makes the Error implement error unwrapping.
func (e Error) Unwrap() error {
	return e.cause
}

// Is makes the Error implement error comparison.
func (e Error) Is(target error) bool {
	return Resemble(e, target)
}

// Unwrap makes the Definition implement error unwrapping.
func (Definition) Unwrap() error {
	return nil
}

// Is makes the Definition implement error comparison.
func (d Definition) Is(target error) bool {
	return Resemble(d, target)
}
