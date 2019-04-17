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

// Package cloud implements an events.PubSub implementation that uses Go Cloud PubSub.
package cloud

import (
	"context"
	"encoding/json"

	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"gocloud.dev/pubsub"
)

// WrapPubSub wraps an existing PubSub and publishes all events received from Go Cloud to that PubSub.
// If the subURL is an empty string, this PubSub will only publish to Go Cloud.
func WrapPubSub(ctx context.Context, wrapped events.PubSub, pubURL, subURL string) (ps *PubSub, err error) {
	ps = &PubSub{
		PubSub:      wrapped,
		ctx:         ctx,
		contentType: "application/protobuf",
	}
	ps.topic, err = pubsub.OpenTopic(ctx, pubURL)
	if err != nil {
		return nil, err
	}
	if subURL != "" {
		ps.subscription, err = pubsub.OpenSubscription(ctx, subURL)
		if err != nil {
			return nil, err
		}
		go func() {
			for ctx.Err() == nil {
				msg, err := ps.subscription.Receive(ctx)
				if err != nil {
					return
				}
				switch msg.Metadata["content-type"] {
				case "application/protobuf":
					var evtpb ttnpb.Event
					if err = evtpb.Unmarshal(msg.Body); err == nil {
						if evt, err := events.FromProto(&evtpb); err == nil {
							ps.PubSub.Publish(evt)
						}
					}
				case "application/json":
					if evt, err := events.UnmarshalJSON(msg.Body); err == nil {
						ps.PubSub.Publish(evt)
					}
				}
				msg.Ack()
			}
		}()
	}
	return ps, nil
}

// NewPubSub creates a new PubSub that publishes and subscribes to Go Cloud.
// If the subURL is an empty string, this PubSub will only publish to Go Cloud.
func NewPubSub(ctx context.Context, pubURL, subURL string) (*PubSub, error) {
	return WrapPubSub(ctx, events.NewPubSub(events.DefaultBufferSize), pubURL, subURL)
}

// PubSub with Go Cloud backend.
type PubSub struct {
	events.PubSub

	ctx          context.Context
	contentType  string
	topic        *pubsub.Topic
	subscription *pubsub.Subscription
}

// Close the Go Cloud publisher.
func (ps *PubSub) Close() (err error) {
	if ps.subscription != nil {
		err = ps.subscription.Shutdown(ps.ctx)
	}
	pubShutdownErr := ps.topic.Shutdown(ps.ctx)
	if err == nil {
		err = pubShutdownErr
	}
	return err
}

// Publish an event to Go Cloud.
func (ps *PubSub) Publish(evt events.Event) {
	var body []byte
	switch ps.contentType {
	case "application/protobuf":
		evtpb, err := events.Proto(evt)
		if err != nil {
			return
		}
		body, err = evtpb.Marshal()
		if err != nil {
			return
		}
	case "application/json":
		var err error
		body, err = json.Marshal(evt)
		if err != nil {
			return
		}
	}
	ps.topic.Send(evt.Context(), &pubsub.Message{
		Metadata: map[string]string{
			"content-type": ps.contentType,
		},
		Body: body,
	})
}
