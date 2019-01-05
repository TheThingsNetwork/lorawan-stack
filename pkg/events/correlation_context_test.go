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

package events_test

import (
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestCorrelationContext(t *testing.T) {
	a := assertions.New(t)

	ctx := test.Context()
	a.So(events.CorrelationIDsFromContext(ctx), should.BeEmpty)

	// Add random correlation ID:
	ctx = events.ContextWithEnsuredCorrelationID(ctx)
	correlationIDs := events.CorrelationIDsFromContext(ctx)
	a.So(correlationIDs, should.HaveLength, 1)

	// Do not add random correlation ID again:
	ctx = events.ContextWithEnsuredCorrelationID(ctx)
	correlationIDs = events.CorrelationIDsFromContext(ctx)
	a.So(correlationIDs, should.HaveLength, 1)

	// Add correlation ID:
	ctx = events.ContextWithCorrelationID(ctx, "foo")
	correlationIDs = events.CorrelationIDsFromContext(ctx)
	a.So(correlationIDs, should.HaveLength, 2)

	// Do not add correlation ID again:
	ctx = events.ContextWithCorrelationID(ctx, "foo")
	correlationIDs = events.CorrelationIDsFromContext(ctx)
	a.So(correlationIDs, should.HaveLength, 2)
}
