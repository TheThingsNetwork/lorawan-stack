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

package messageprocessors

import "go.thethings.network/lorawan-stack/pkg/errors"

// ErrInvalidInput represents the ErrDescriptor of the error returned when
// the input is not valid.
var ErrInvalidInput = &errors.ErrDescriptor{
	MessageFormat: "Invalid input",
	Type:          errors.InvalidArgument,
	Code:          1,
}

// ErrInvalidOutput represents the ErrDescriptor of the error returned when
// the output is invalid.
var ErrInvalidOutput = &errors.ErrDescriptor{
	MessageFormat: "Invalid output",
	Type:          errors.External,
	Code:          2,
}

// ErrInvalidOutputType represents the ErrDescriptor of the error returned when
// the output is not of the valid type.
var ErrInvalidOutputType = &errors.ErrDescriptor{
	MessageFormat:  "Invalid output of type `{type}`",
	Type:           errors.External,
	Code:           3,
	SafeAttributes: []string{"type"},
}

// ErrInvalidOutputRange represents the ErrDescriptor of the error returned when
// the output does not fall in a valid range.
var ErrInvalidOutputRange = &errors.ErrDescriptor{
	MessageFormat:  "Value `{value}` does not fall in range of `{low}` to `{high}`",
	Type:           errors.External,
	Code:           4,
	SafeAttributes: []string{"low", "high", "value"},
}

func init() {
	ErrInvalidInput.Register()
	ErrInvalidOutput.Register()
	ErrInvalidOutputType.Register()
	ErrInvalidOutputRange.Register()
}
