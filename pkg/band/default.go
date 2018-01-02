// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package band

import "time"

const (
	defaultReceiveDelay1 time.Duration = time.Second
	defaultReceiveDelay2 time.Duration = defaultReceiveDelay1 + time.Second

	defaultJoinAcceptDelay1 time.Duration = 5 * time.Second
	defaultJoinAcceptDelay2 time.Duration = defaultJoinAcceptDelay1 + time.Second

	defaultMaxFCntGap uint = 16384

	defaultAdrAckLimit uint8 = 64
	defaultAdrAckDelay uint8 = 32

	// Random delay between 1 and 3 seconds
	defaultAckTimeout       time.Duration = 2 * time.Second
	defaultAckTimeoutMargin time.Duration = 1 * time.Second
)
