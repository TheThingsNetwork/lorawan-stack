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

package correlations

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// DownlinkInfo is the information associated with a particular downlink.
type DownlinkInfo struct {
	correlationIDs []string
	txTime         time.Time
}

// DownlinkCorrelation is used to maintain tokens to correlate Downlink messages with their corresponding TxConfitrmations.
type DownlinkCorrelation struct {
	// token is the unique token associated with each Downlink.
	// It's passed to the BasicStation as the `diid` field and is returned as-is in the TxConfirmation if the downlink packet was put on air.
	// This is a free-running counter that is allowed to overflow and is cleaned up periodically by the garbage collector.
	token        int64
	expiration   time.Duration
	ctx          context.Context
	correlations sync.Map
}

// New returns a new downlink correlation.
func New(ctx context.Context, expiration time.Duration) *DownlinkCorrelation {
	return &DownlinkCorrelation{
		ctx:        ctx,
		expiration: expiration,
		token:      -1,
	}
}

// GenerateNextToken atomically increments the token value and returns the incremented value.
func (c *DownlinkCorrelation) GenerateNextToken() int64 {
	return atomic.AddInt64(&c.token, 1)
}

// Store stores the downlink correlation for the given CorrelationIDs.
func (c *DownlinkCorrelation) Store(token int64, correlationIDs []string) {
	c.correlations.Store(token, DownlinkInfo{
		correlationIDs: correlationIDs,
		txTime:         time.Now(),
	})
}

// Fetch attempts to fetch the CorrelationIDs for a given token and deletes it in the map.
// If the correlation is not found, an empty string is returned.
func (c *DownlinkCorrelation) Fetch(receivedToken int64) []string {
	if value, ok := c.correlations.Load(receivedToken); ok {
		c.correlations.Delete(receivedToken)
		return value.(DownlinkInfo).correlationIDs
	}
	return nil
}

// GC is the garbage collector that removes old items from the correlations map.
func (c *DownlinkCorrelation) GC() {
	gcTicker := time.NewTicker(c.expiration)
	for {
		select {
		case <-c.ctx.Done():
			gcTicker.Stop()
			return
		case <-gcTicker.C:
			c.correlations.Range(func(key interface{}, value interface{}) bool {
				if value.(DownlinkInfo).txTime.Before(time.Now().Add(-c.expiration)) {
					c.correlations.Delete(key)
				}
				return true
			})
		}
	}
}
