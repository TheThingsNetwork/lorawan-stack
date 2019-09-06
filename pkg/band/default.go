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

package band

import (
	"time"

	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

const (
	defaultReceiveDelay1 time.Duration = time.Second
	defaultReceiveDelay2 time.Duration = defaultReceiveDelay1 + time.Second

	defaultJoinAcceptDelay1 time.Duration = 5 * time.Second
	defaultJoinAcceptDelay2 time.Duration = defaultJoinAcceptDelay1 + time.Second

	defaultMaxFCntGap uint = 16384

	defaultADRAckLimit = ttnpb.ADR_ACK_LIMIT_64
	defaultADRAckDelay = ttnpb.ADR_ACK_DELAY_32

	// Random delay between 1 and 3 seconds
	defaultAckTimeout       time.Duration = 2 * time.Second
	defaultAckTimeoutMargin time.Duration = 1 * time.Second
)
