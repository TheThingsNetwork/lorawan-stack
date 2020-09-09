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
	"strings"
	"sync"

	"go.thethings.network/lorawan-stack/v3/pkg/component"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"gocloud.dev/pubsub"
)

// WrapPubSub wraps an existing PubSub and publishes all events received from Go Cloud to that PubSub.
// If the subURL is an empty string, this PubSub will only publish to Go Cloud.
func WrapPubSub(ctx context.Context, wrapped events.PubSub, taskStarter component.TaskStarter, pubURL, subURL string) (ps *PubSub, err error) {
	ctx = log.NewContextWithField(ctx, "namespace", "events/cloud")
	ctx, cancel := context.WithCancel(ctx)
	ps = &PubSub{
		PubSub:      wrapped,
		taskStarter: taskStarter,
		ctx:         ctx,
		cancel:      cancel,
		contentType: "application/protobuf",
		subURL:      subURL,
	}
	ps.topic, err = pubsub.OpenTopic(ctx, pubURL)
	if err != nil {
		return nil, err
	}
	return ps, nil
}

// NewPubSub creates a new PubSub that publishes and subscribes to Go Cloud.
// If the subURL is an empty string, this PubSub will only publish to Go Cloud.
func NewPubSub(ctx context.Context, taskStarter component.TaskStarter, pubURL, subURL string) (*PubSub, error) {
	return WrapPubSub(ctx, events.NewPubSub(events.DefaultBufferSize), taskStarter, pubURL, subURL)
}

// PubSub with Go Cloud backend.
type PubSub struct {
	events.PubSub

	taskStarter component.TaskStarter
	ctx         context.Context
	cancel      context.CancelFunc
	contentType string
	subURL      string
	topic       *pubsub.Topic
	subOnce     sync.Once
}

// Close the Go Cloud publisher.
func (ps *PubSub) Close(ctx context.Context) error {
	if err := ps.topic.Shutdown(ps.ctx); err != nil {
		return err
	}
	ps.cancel()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-ps.ctx.Done():
		return ps.ctx.Err()
	}
}

func (ps *PubSub) subscribeTask(ctx context.Context) error {
	if ps.subURL == "" {
		return nil
	}
	logger := log.FromContext(ctx)
	subscription, err := pubsub.OpenSubscription(ctx, ps.subURL)
	if err != nil {
		return err
	}
	logger.Info("Subscribed")
	defer func() {
		if err := subscription.Shutdown(ctx); err != nil {
			logger.WithError(err).Warn("Failed to close Go Cloud subscription")
		} else {
			logger.Info("Unsubscribed")
		}
	}()
	for {
		msg, err := subscription.Receive(ctx)
		if err != nil {
			return err
		}
		msg.Ack()
		m := msg.Metadata["content-type"]
		var evt events.Event
		switch m {
		case "application/protobuf":
			var e ttnpb.Event
			if err = e.Unmarshal(msg.Body); err != nil {
				logger.WithError(err).Warn("Failed to unmarshal event from binary")
				continue
			}
			if evt, err = events.FromProto(&e); err != nil {
				logger.WithError(err).Warn("Failed to unmarshal event from protobuf")
				continue
			}
		case "application/json":
			if evt, err = events.UnmarshalJSON(msg.Body); err != nil {
				logger.WithError(err).Warn("Failed to unmarshal event from JSON")
				continue
			}
		default:
			logger.WithField("content_type", m).Warn("Received event with unknown content type")
			continue
		}
		ps.PubSub.Publish(evt)
	}
}

// Subscribe to events from Go Cloud.
func (ps *PubSub) Subscribe(name string, hdl events.Handler) error {
	ps.subOnce.Do(func() {
		ps.taskStarter.StartTask(&component.TaskConfig{
			Context: ps.ctx,
			ID:      "events_cloud_subscribe",
			Func:    ps.subscribeTask,
			Restart: component.TaskRestartOnFailure,
			Backoff: component.DefaultTaskBackoffConfig,
		})
	})
	return ps.PubSub.Subscribe(name, hdl)
}

func (ps *PubSub) getMetadata(evt events.Event) map[string]string {
	ids := make(map[string][]string, 10)
	for _, id := range evt.Identifiers() {
		k := id.EntityType() + "_id"
		ids[k] = append(ids[k], id.IDString())
		if gtwID := id.GetGatewayIDs(); gtwID != nil {
			ids["gateway_eui"] = append(ids["gateway_eui"], gtwID.EUI.String())
		}
		if devID := id.GetDeviceIDs(); devID != nil {
			ids["application_id"] = append(ids["application_id"], devID.ApplicationID)
			if devID.DevEUI != nil {
				ids["dev_eui"] = append(ids["dev_eui"], devID.DevEUI.String())
			}
			if devID.JoinEUI != nil {
				ids["join_eui"] = append(ids["join_eui"], devID.JoinEUI.String())
			}
			if devID.DevAddr != nil {
				ids["dev_addr"] = append(ids["dev_addr"], devID.DevAddr.String())
			}
		}
	}
	md := make(map[string]string, len(ids)+3)
	md["content-type"] = ps.contentType
	md["event"] = evt.Name()
	md["correlation_ids"] = strings.Join(evt.CorrelationIDs(), ",")
	for k, v := range ids {
		md[k] = strings.Join(v, ",")
	}
	return md
}

// Publish an event to Go Cloud.
func (ps *PubSub) Publish(evt events.Event) {
	logger := log.FromContext(ps.ctx)
	var body []byte
	switch ps.contentType {
	case "application/protobuf":
		evtpb, err := events.Proto(evt)
		if err != nil {
			logger.WithError(err).Warn("Failed to marshal event to protobuf")
			return
		}
		body, err = evtpb.Marshal()
		if err != nil {
			logger.WithError(err).Warn("Failed to marshal event to binary")
			return
		}
	case "application/json":
		var err error
		body, err = json.Marshal(evt)
		if err != nil {
			logger.WithError(err).Warn("Failed to marshal event to JSON")
			return
		}
	}
	ps.topic.Send(evt.Context(), &pubsub.Message{
		Metadata: ps.getMetadata(evt),
		Body:     body,
	})
}
