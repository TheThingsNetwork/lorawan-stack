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

package interop

import (
	"go.thethings.network/lorawan-stack/pkg/errors"
)

var (
	errUnknownMACVersion  = errors.DefineInvalidArgument("unknown_mac_version", "unknown MAC version")
	errInvalidLength      = errors.DefineInvalidArgument("invalid_length", "invalid length")
	errInvalidRequestType = errors.DefineInvalidArgument("invalid_request_type", "invalid request type `{type}`")
)
