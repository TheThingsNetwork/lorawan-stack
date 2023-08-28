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

package udp

import (
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
)

var (
	errNoEUI        = errors.DefineInvalidArgument("no_eui", "packet is not long enough to contain the EUI")
	errPayload      = errors.DefineInvalidArgument("payload", "parse binary payload")
	errEUI          = errors.DefineInvalidArgument("eui", "parse EUI")
	errTimestamp    = errors.DefineInvalidArgument("timestamp", "parse timestamp")
	errDataRate     = errors.DefineInvalidArgument("data_rate", "invalid data rate")
	errNotScheduled = errors.DefineInvalidArgument("not_scheduled", "downlink message not scheduled")
)
