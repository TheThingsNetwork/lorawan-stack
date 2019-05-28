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

// Package pubsub implements the go-cloud PubSub frontend.
package pubsub

import (
	"context"
	"fmt"

	"go.thethings.network/lorawan-stack/pkg/applicationserver/io"
	"go.thethings.network/lorawan-stack/pkg/applicationserver/io/formatters"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
	"gocloud.dev/pubsub"
)

type srv struct {
	ctx           context.Context
	server        io.Server
	formatter     formatters.Formatter
	topics        []*pubsub.Topic
	subscriptions []*pubsub.Subscription
}

// Start starts the pusub frontend.
func Start(ctx context.Context, server io.Server, formatter formatters.Formatter, pubURLs, subURLs []string) ([]*io.Subscription, error) {
	ctx = log.NewContextWithField(ctx, "namespace", "applicationserver/io/pubsub")
	s := &srv{
		ctx:       ctx,
		server:    server,
		formatter: formatter,
	}

	subs := []*io.Subscription{}
	for _, pubURL := range pubURLs {
		topic, err := pubsub.OpenTopic(ctx, pubURL)
		if err != nil {
			return nil, err
		}
		sub := io.NewSubscription(s.ctx, "pubsub", nil)
		// Publish upstream
		go func() {
			logger := log.FromContext(s.ctx).WithField("publish-url", pubURL)
			for {
				select {
				case <-sub.Context().Done():
					logger.WithError(sub.Context().Err()).Debug("Done sending upstream messages")
					return
				case up := <-sub.Up():
					buf, err := s.formatter.FromUp(up)
					if err != nil {
						log.WithError(err).Warn("Failed to marshal upstream message")
						continue
					}
					err = topic.Send(ctx, &pubsub.Message{
						Body: buf,
					})
					if err != nil {
						log.WithError(err).Warn("Failed to publish upstream message")
						continue
					}
					logger.Debug("Publish upstream message")
				}
			}
		}()
		subs = append(subs, sub)
		s.topics = append(s.topics, topic)
	}
	for _, subURL := range subURLs {
		subscription, err := pubsub.OpenSubscription(ctx, subURL)
		if err != nil {
			return nil, err
		}
		// Subscribe downstream
		go func() {
			logger := log.FromContext(s.ctx).WithField("subscribe-url", subURL)
			for ctx.Err() == nil {
				msg, err := subscription.Receive(ctx)
				if err != nil {
					logger.WithError(err).Warn("Failed to receive downlink queue operation")
					continue
				}
				msg.Ack()
				operation, err := s.formatter.ToDownlinkQueueOperation(msg.Body)
				if err != nil {
					logger.WithError(err).Warn("Failed to decode downlink queue operation")
					continue
				}
				var op func(io.Server, context.Context, ttnpb.EndDeviceIdentifiers, []*ttnpb.ApplicationDownlink) error
				switch operation.Operation {
				case ttnpb.DownlinkQueueOperation_PUSH:
					op = io.Server.DownlinkQueuePush
				case ttnpb.DownlinkQueueOperation_REPLACE:
					op = io.Server.DownlinkQueueReplace
				default:
					panic(fmt.Errorf("invalid operation: %v", operation.Operation))
				}
				logger.WithFields(log.Fields(
					"device_uid", unique.ID(s.ctx, operation.EndDeviceIdentifiers),
					"count", len(operation.Downlinks),
					"operation", operation.Operation,
				)).Debug("Handle downlink messages")
				if err := op(s.server, s.ctx, operation.EndDeviceIdentifiers, operation.Downlinks); err != nil {
					logger.WithError(err).Warn("Failed to handle downlink messages")
				}
			}
		}()
		s.subscriptions = append(s.subscriptions, subscription)
	}

	go func() {
		<-ctx.Done()
		for _, topic := range s.topics {
			topic.Shutdown(ctx)
		}
		for _, subscription := range s.subscriptions {
			subscription.Shutdown(ctx)
		}
	}()

	return subs, nil
}
