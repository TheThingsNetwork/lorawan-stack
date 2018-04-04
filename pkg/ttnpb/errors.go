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

package ttnpb

import "github.com/TheThingsNetwork/ttn/pkg/errors"

func init() {
	ErrEmptyUpdateMask.Register()
	ErrInvalidPathUpdateMask.Register()
	ErrEmptyIdentifiers.Register()
}

// ErrEmptyUpdateMask is returned when the update mask is specified but empty.
var ErrEmptyUpdateMask = &errors.ErrDescriptor{
	MessageFormat: "update_mask must be non-empty",
	Code:          1,
	Type:          errors.InvalidArgument,
}

// ErrInvalidPathUpdateMask is returned when the update mask includes a wrong field path.
var ErrInvalidPathUpdateMask = &errors.ErrDescriptor{
	MessageFormat: "Invalid update_mask: `{path}` is not a valid path",
	Code:          2,
	Type:          errors.InvalidArgument,
}

// ErrEmptyIdentifiers is returned when the XXXIdentifiers are empty.
var ErrEmptyIdentifiers = &errors.ErrDescriptor{
	MessageFormat: "Identifiers must be non-empty",
	Code:          3,
	Type:          errors.InvalidArgument,
}
