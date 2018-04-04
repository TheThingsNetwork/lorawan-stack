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

package scripting

import "github.com/TheThingsNetwork/ttn/pkg/errors"

// ErrRuntime represents the ErrDescriptor of the error returned when
// there is a runtime error.
var ErrRuntime = &errors.ErrDescriptor{
	MessageFormat: "Runtime error",
	Type:          errors.External,
	Code:          1,
}

func init() {
	ErrRuntime.Register()
}
