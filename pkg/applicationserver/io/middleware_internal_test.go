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

package io_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/applicationserver/io"
	"go.thethings.network/lorawan-stack/pkg/applicationserver/io/mock"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

var (
	registeredApplicationID = ttnpb.ApplicationIdentifiers{ApplicationID: "foo-app"}
	registeredDeviceID      = ttnpb.EndDeviceIdentifiers{
		ApplicationIdentifiers: registeredApplicationID,
		DeviceID:               "foo-device",
	}
	registeredApplicationUp = &ttnpb.ApplicationUp{
		EndDeviceIdentifiers: registeredDeviceID,
		Up: &ttnpb.ApplicationUp_JoinAccept{
			JoinAccept: &ttnpb.ApplicationJoinAccept{
				SessionKeyID: []byte{0x11},
			},
		},
	}
	timeout     = (1 << 6) * test.Delay
	backoff     = []time.Duration{(1 << 4) * test.Delay}
	errConnLost = errors.New("connection lost")
)

func TestRetryServer(t *testing.T) {
	a := assertions.New(t)
	ctx := newContextWithRightsFetcher(test.Context())
	server := mock.NewServer(nil)
	retryServer := io.NewRetryServer(server, io.WithBackoff(backoff))

	downstreamSub, err := retryServer.Subscribe(ctx, "foo", registeredApplicationID)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	a.So(downstreamSub, should.NotBeNil)

	var upstreamSub *io.Subscription
	select {
	case <-time.After(timeout):
		t.Fatal("Upstream subscription did not occur")
	case upstreamSub = <-server.Subscriptions():
	}

	// Send uplink traffic from upstream to downstream.
	err = upstreamSub.SendUp(ctx, registeredApplicationUp)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	downstreamUp := <-downstreamSub.Up()
	a.So(downstreamUp.Context, should.HaveParentContextOrEqual, ctx)
	a.So(downstreamUp.ApplicationUp, should.Resemble, registeredApplicationUp)

	// Cancel upstream subscription gracefully.
	upstreamSub.Disconnect(nil)
	select {
	case <-downstreamSub.Context().Done():
		t.Fatal("Downstream context has been cancelled")
	case <-time.After(timeout):
	}

	// Wait for reconnection.
	select {
	case <-time.After(timeout):
		t.Fatal("Upstream subscription did not occur")
	case upstreamSub = <-server.Subscriptions():
	}

	// Send uplink traffic over the reestablished link.
	err = upstreamSub.SendUp(ctx, registeredApplicationUp)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	downstreamUp = <-downstreamSub.Up()
	a.So(downstreamUp.Context, should.HaveParentContextOrEqual, ctx)
	a.So(downstreamUp.ApplicationUp, should.Resemble, registeredApplicationUp)

	// Shutdown the link completely.
	server.SetSubscribeError(errConnLost)
	upstreamSub.Disconnect(nil)

	// Check that the downstream connection failed.
	select {
	case <-time.After(timeout):
		t.Fatal("Waiting for downstream failure timed out")
	case <-downstreamSub.Context().Done():
	}
	err = downstreamSub.Context().Err()
	a.So(err, should.Equal, errConnLost)
}

func newContextWithRightsFetcher(ctx context.Context) context.Context {
	return rights.NewContextWithFetcher(
		ctx,
		rights.FetcherFunc(func(ctx context.Context, ids ttnpb.Identifiers) (*ttnpb.Rights, error) {
			return ttnpb.RightsFrom(
				ttnpb.RIGHT_APPLICATION_TRAFFIC_READ,
			), nil
		}),
	)
}
