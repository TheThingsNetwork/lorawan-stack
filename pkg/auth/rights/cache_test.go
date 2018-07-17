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

package rights

import (
	"context"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	errors "go.thethings.network/lorawan-stack/pkg/errorsv3"
)

var currentTime time.Time

func timeTravel(d time.Duration) {
	currentTime = currentTime.Add(d)
}

func TestCache(t *testing.T) {
	now = func() time.Time {
		return currentTime
	}

	a := assertions.New(t)

	mockErr := errors.New("mock")

	mockFetcher := &mockFetcher{}

	c := NewInMemoryCache(mockFetcher, 5*time.Minute, time.Minute).(*inMemoryCache)

	mockFetcher.applicationError, mockFetcher.gatewayError, mockFetcher.organizationError = mockErr, mockErr, mockErr

	ctxA := context.WithValue(context.Background(), "ctx", "A")
	res := fetchRights(ctxA, "foo", c)

	a.So(mockFetcher.applicationCtx, should.Equal, ctxA)
	a.So(mockFetcher.gatewayCtx, should.Equal, ctxA)
	a.So(mockFetcher.organizationCtx, should.Equal, ctxA)

	a.So(res.AppErr, should.Resemble, mockFetcher.applicationError)
	a.So(res.GtwErr, should.Resemble, mockFetcher.gatewayError)
	a.So(res.OrgErr, should.Resemble, mockFetcher.organizationError)

	timeTravel(31 * time.Second) // Error responses should be cached for 1 minute.

	ctxB := context.WithValue(context.Background(), "ctx", "B")
	res = fetchRights(ctxB, "foo", c)

	a.So(mockFetcher.applicationCtx, should.Equal, ctxA)
	a.So(mockFetcher.gatewayCtx, should.Equal, ctxA)
	a.So(mockFetcher.organizationCtx, should.Equal, ctxA)

	timeTravel(31 * time.Second) // Error responses should be expired after 1 minute.

	res = fetchRights(ctxB, "foo", c)

	a.So(mockFetcher.applicationCtx, should.Equal, ctxB)
	a.So(mockFetcher.gatewayCtx, should.Equal, ctxB)
	a.So(mockFetcher.organizationCtx, should.Equal, ctxB)

	timeTravel(61 * time.Second)

	mockFetcher.applicationError, mockFetcher.gatewayError, mockFetcher.organizationError = nil, nil, nil

	res = fetchRights(ctxA, "foo", c)

	a.So(mockFetcher.applicationCtx, should.Equal, ctxA)
	a.So(mockFetcher.gatewayCtx, should.Equal, ctxA)
	a.So(mockFetcher.organizationCtx, should.Equal, ctxA)

	timeTravel(3 * time.Minute) // Success responses should be cached for 5 minutes.

	res = fetchRights(ctxB, "foo", c)

	a.So(mockFetcher.applicationCtx, should.Equal, ctxA)
	a.So(mockFetcher.gatewayCtx, should.Equal, ctxA)
	a.So(mockFetcher.organizationCtx, should.Equal, ctxA)

	timeTravel(3 * time.Minute) // Success responses should be expired after 5 minutes.

	res = fetchRights(ctxB, "foo", c)

	a.So(mockFetcher.applicationCtx, should.Equal, ctxB)
	a.So(mockFetcher.gatewayCtx, should.Equal, ctxB)
	a.So(mockFetcher.organizationCtx, should.Equal, ctxB)

	timeTravel(time.Hour)

	c.maybeCleanup()

	a.So(c.applicationRights, should.BeEmpty)
	a.So(c.gatewayRights, should.BeEmpty)
	a.So(c.organizationRights, should.BeEmpty)
}
