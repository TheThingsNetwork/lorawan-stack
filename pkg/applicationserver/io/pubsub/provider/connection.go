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

package provider

import (
	"context"
	"reflect"

	"github.com/golang/protobuf/proto"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"gocloud.dev/pubsub"
	"google.golang.org/grpc/codes"
)

// DownlinkSubscriptions contains the subscriptions for the push and replace queue operations.
type DownlinkSubscriptions struct {
	Push    *pubsub.Subscription
	Replace *pubsub.Subscription
}

// Shutdown shutdowns the active subscriptions.
func (ds *DownlinkSubscriptions) Shutdown(ctx context.Context) error {
	return shutdown(ctx,
		ds.Push,
		ds.Replace,
	)
}

// UplinkTopics contains the topics for the uplink messages.
type UplinkTopics struct {
	UplinkMessage            *pubsub.Topic
	JoinAccept               *pubsub.Topic
	DownlinkAck              *pubsub.Topic
	DownlinkNack             *pubsub.Topic
	DownlinkSent             *pubsub.Topic
	DownlinkFailed           *pubsub.Topic
	DownlinkQueued           *pubsub.Topic
	DownlinkQueueInvalidated *pubsub.Topic
	LocationSolved           *pubsub.Topic
	ServiceData              *pubsub.Topic
}

// Shutdown shutdowns the active topics.
func (ut *UplinkTopics) Shutdown(ctx context.Context) error {
	return shutdown(ctx,
		ut.UplinkMessage,
		ut.JoinAccept,
		ut.DownlinkAck,
		ut.DownlinkNack,
		ut.DownlinkSent,
		ut.DownlinkFailed,
		ut.DownlinkQueued,
		ut.DownlinkQueueInvalidated,
		ut.LocationSolved,
		ut.ServiceData,
	)
}

// Shutdowner is an interface that contains a contextual shutdown method.
type Shutdowner interface {
	Shutdown(ctx context.Context) error
}

// ProviderConnection is an interface that represents a provider specific connection.
type ProviderConnection interface { //nolint:revive
	Shutdowner
}

// Connection is a wrapper that wraps the topics and subscriptions with a ProviderConnection.
type Connection struct {
	Topics             UplinkTopics
	Subscriptions      DownlinkSubscriptions
	ProviderConnection ProviderConnection
}

// Shutdown shuts down the topics, subscriptions and the connections if required.
func (c *Connection) Shutdown(ctx context.Context) error {
	return shutdown(ctx,
		&c.Topics,
		&c.Subscriptions,
		c.ProviderConnection,
	)
}

var errShutdown = errors.DefineInternal("shutdown", "shutdown")

func shutdown(ctx context.Context, shutdowners ...Shutdowner) error {
	details := make([]proto.Message, 0, len(shutdowners))
	for _, s := range shutdowners {
		if isNil(s) {
			continue
		}
		if err := s.Shutdown(ctx); err != nil {
			details = append(details, toProtoMessage(err))
		}
	}
	if len(details) > 0 {
		return errShutdown.WithDetails(details...)
	}
	return nil
}

func toProtoMessage(err error) proto.Message {
	if ttnErr, ok := errors.From(err); ok {
		return ttnpb.ErrorDetailsToProto(ttnErr)
	}
	return &ttnpb.ErrorDetails{
		Code:          uint32(codes.Unknown),
		MessageFormat: err.Error(),
	}
}

func isNil(c interface{}) bool {
	if c == nil {
		return true
	}
	if val := reflect.ValueOf(c); val.Kind() == reflect.Ptr {
		return val.IsNil()
	}
	return false
}
