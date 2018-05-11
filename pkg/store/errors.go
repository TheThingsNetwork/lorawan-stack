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

package store

import "go.thethings.network/lorawan-stack/pkg/errors"

func init() {
	ErrInvalidData.Register()
	ErrEmptyFilter.Register()
	ErrNilKey.Register()
}

// ErrInvalidData represents an error returned, when data specified is not valid.
var ErrInvalidData = &errors.ErrDescriptor{
	MessageFormat: "Invalid data",
	Type:          errors.InvalidArgument,
	Code:          1,
}

// ErrEmptyFilter represents an error returned, when filter specified is empty.
var ErrEmptyFilter = &errors.ErrDescriptor{
	MessageFormat: "Filter is empty",
	Type:          errors.InvalidArgument,
	Code:          2,
}

// ErrNilKey represents an error returned, when key specified is nil.
var ErrNilKey = &errors.ErrDescriptor{
	MessageFormat: "Nil key specified",
	Type:          errors.InvalidArgument,
	Code:          3,
}
