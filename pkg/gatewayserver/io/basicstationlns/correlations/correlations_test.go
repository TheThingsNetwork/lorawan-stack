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
	"strconv"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

var (
	GCTime   = (1 << 8) * test.Delay
	WaitTime = 2 * GCTime
)

func store(corr *DownlinkCorrelation, n int) {
	for i := 0; i < n; i++ {
		token := corr.GenerateNextToken()
		corr.Store(token, []string{strconv.Itoa(i)})
	}
}

func TestCorrelations(t *testing.T) {
	a := assertions.New(t)
	ctx := log.NewContext(test.Context(), test.GetLogger(t))
	defer ctx.Done()

	corr := New(ctx, GCTime)
	go corr.GC()

	// Simple Load Store
	for i := 0; i < 10000; i++ {
		corr.Store(corr.GenerateNextToken(), []string{strconv.Itoa(i)})
	}
	for i := 0; i < 10000; i++ {
		ids := corr.Fetch(int64(i))
		if !a.So(ids, should.Resemble, []string{strconv.Itoa(i)}) {
			t.Fatalf("Invalid Correlation IDs for token %d : %v", i, ids)
		}
	}
	ids := corr.Fetch(25)
	if !a.So(ids, should.BeEmpty) {
		t.Fatalf("Invalid Correlation IDs: %v", ids)
	}

	// Test GC
	corr.token = -1
	for i := 0; i < 10000; i++ {
		corr.Store(corr.GenerateNextToken(), []string{strconv.Itoa(i)})
	}
	time.Sleep(WaitTime)

	// Check if the map has been cleared
	for i := 0; i < 10000; i++ {
		ids := corr.Fetch(int64(i))
		if !a.So(ids, should.BeEmpty) {
			t.Fatalf("Invalid non-empty CorrelationIDs for token %d : %v", i, ids)
		}
	}
}
