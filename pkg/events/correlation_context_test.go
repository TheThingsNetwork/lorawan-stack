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
	correlationIDs := events.CorrelationIDsFromContext(ctx)
	a.So(correlationIDs, should.BeEmpty)

	// Add correlation ID:
	ctx = events.ContextWithCorrelationID(ctx, "foo")
	correlationIDs = events.CorrelationIDsFromContext(ctx)
	a.So(correlationIDs, should.Resemble, []string{"foo"})

	// Add different correlation ID:
	ctx = events.ContextWithCorrelationID(ctx, "baz")
	correlationIDs = events.CorrelationIDsFromContext(ctx)
	a.So(correlationIDs, should.Resemble, []string{"baz", "foo"})

	// Only add new correlation IDs:
	ctx = events.ContextWithCorrelationID(ctx, "bar", "foo")
	correlationIDs = events.CorrelationIDsFromContext(ctx)
	a.So(correlationIDs, should.Resemble, []string{"bar", "baz", "foo"})

	// Do not mutate the passed slice
	base := []string{
		"b",
		"a",
		"c",
	}
	ctx = events.ContextWithCorrelationID(ctx, base...)
	correlationIDs = events.CorrelationIDsFromContext(ctx)
	a.So(correlationIDs, should.Resemble, []string{"a", "b", "bar", "baz", "c", "foo"})
	a.So(base, should.Resemble, []string{
		"b",
		"a",
		"c",
	})
}
