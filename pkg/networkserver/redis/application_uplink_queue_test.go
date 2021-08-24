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

package redis_test

import (
	"fmt"
	"testing"

	"go.thethings.network/lorawan-stack/v3/pkg/networkserver"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal/test/shared"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
)

var _ networkserver.ApplicationUplinkQueue = &ApplicationUplinkQueue{}

func TestApplicationUplinkQueue(t *testing.T) {
	for _, consumers := range []int{1, 2, 4, 8} {
		t.Run(fmt.Sprintf("Consumers=%d", consumers), func(t *testing.T) {
			_, ctx := test.New(t)
			consumerIDs := make([]string, 0, consumers)
			for i := 0; i < consumers; i++ {
				consumerIDs = append(consumerIDs, fmt.Sprintf("consumer-%d-%d", consumers, i))
			}
			q, closeFn := NewRedisApplicationUplinkQueue(ctx)
			t.Cleanup(closeFn)
			HandleApplicationUplinkQueueTest(t, q, consumerIDs)
		})
	}
}
