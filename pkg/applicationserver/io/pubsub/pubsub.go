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
	"sync"
	"time"

	"go.thethings.network/lorawan-stack/pkg/applicationserver/io"
	"go.thethings.network/lorawan-stack/pkg/applicationserver/io/pubsub/provider"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/errorcontext"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
	"gocloud.dev/pubsub"
)

type srv struct {
	*component.Component
	ctx      context.Context
	server   io.Server
	registry Registry

	integrations      sync.Map
	integrationErrors sync.Map
}

// Start starts the pusub frontend.
func Start(c *component.Component, server io.Server, registry Registry) error {
	ctx := log.NewContextWithField(c.Context(), "namespace", "applicationserver/io/pubsub")
	s := &srv{
		Component: c,
		ctx:       ctx,
		server:    server,
		registry:  registry,
	}
	c.RegisterTask(ctx, "link_all", s.integrateAll, component.TaskRestartOnFailure)
	return nil
}

func (s *srv) integrateAll(ctx context.Context) error {
	return s.registry.Range(ctx, nil,
		func(ctx context.Context, pb *ttnpb.ApplicationPubSub) bool {
			s.startIntegrationTask(ctx, pb.ApplicationPubSubIdentifiers)
			return true
		},
	)
}

var integrationBackoff = []time.Duration{100 * time.Millisecond, 1 * time.Second, 10 * time.Second}

func (s *srv) startIntegrationTask(ctx context.Context, ids ttnpb.ApplicationPubSubIdentifiers) {
	ctx = log.NewContextWithFields(ctx, log.Fields(
		"application_uid", unique.ID(ctx, ids.ApplicationIdentifiers),
		"pubsub_id", ids.PubSubID,
	))
	s.StartTask(ctx, "integrate", func(ctx context.Context) error {
		target, err := s.registry.Get(ctx, ids, []string{
			"attributes",
			"format",
			"provider",

			"downlink_push_topic",
			"downlink_replace_topic",

			"uplink_message",
			"join_accept",
			"downlink_ack",
			"downlink_nack",
			"downlink_queued",
			"downlink_sent",
			"downlink_failed",
			"location_solved",
		})
		if err != nil {
			if !errors.IsNotFound(err) {
				log.FromContext(ctx).WithError(err).Error("Failed to get link")
			}
			return nil
		}

		err = s.integrate(ctx, target)
		switch {
		case errors.IsFailedPrecondition(err),
			errors.IsUnauthenticated(err),
			errors.IsPermissionDenied(err),
			errors.IsInvalidArgument(err):
			log.FromContext(ctx).WithError(err).Warn("Failed to integrate")
			return nil
		case errors.IsCanceled(err),
			errors.IsAlreadyExists(err):
			return nil
		default:
			return err
		}
	}, component.TaskRestartOnFailure, 0.1, integrationBackoff...)
}

type integration struct {
	ttnpb.ApplicationPubSub
	ctx    context.Context
	cancel errorcontext.CancelFunc

	subscriptions *provider.DownlinkSubscriptions
	topics        *provider.UplinkTopics

	server io.Server
	sub    *io.Subscription
	format Format
}

func (i *integration) handleUp(ctx context.Context) {
	logger := log.FromContext(ctx)
	for {
		select {
		case <-ctx.Done():
			logger.WithError(ctx.Err()).Debug("Done sending upstream messages")
			return
		case up := <-i.sub.Up():
			var topic *pubsub.Topic
			switch up.ApplicationUp.Up.(type) {
			case *ttnpb.ApplicationUp_UplinkMessage:
				topic = i.topics.UplinkMessage
			case *ttnpb.ApplicationUp_JoinAccept:
				topic = i.topics.JoinAccept
			case *ttnpb.ApplicationUp_DownlinkAck:
				topic = i.topics.DownlinkAck
			case *ttnpb.ApplicationUp_DownlinkNack:
				topic = i.topics.DownlinkNack
			case *ttnpb.ApplicationUp_DownlinkSent:
				topic = i.topics.DownlinkSent
			case *ttnpb.ApplicationUp_DownlinkFailed:
				topic = i.topics.DownlinkFailed
			case *ttnpb.ApplicationUp_DownlinkQueued:
				topic = i.topics.DownlinkQueued
			case *ttnpb.ApplicationUp_LocationSolved:
				topic = i.topics.LocationSolved
			}
			if topic == nil {
				continue
			}
			buf, err := i.format.FromUp(up.ApplicationUp)
			if err != nil {
				logger.WithError(err).Warn("Failed to marshal upstream message")
				continue
			}
			err = topic.Send(ctx, &pubsub.Message{
				Body: buf,
			})
			if err != nil {
				logger.WithError(err).Warn("Failed to publish upstream message")
				continue
			}
			logger.Debug("Publish upstream message")
		}
	}
}

func (i *integration) handleDown(ctx context.Context, op func(io.Server, context.Context, ttnpb.EndDeviceIdentifiers, []*ttnpb.ApplicationDownlink) error, subscription *pubsub.Subscription) {
	logger := log.FromContext(ctx)
	for ctx.Err() == nil {
		msg, err := subscription.Receive(ctx)
		if err != nil {
			logger.WithError(err).Warn("Failed to receive downlink queue operation")
			continue
		}
		msg.Ack()
		operation, err := i.format.ToDownlinkQueueRequest(msg.Body)
		if err != nil {
			logger.WithError(err).Warn("Failed to decode downlink queue operation")
			continue
		}
		logger.WithFields(log.Fields(
			"device_uid", unique.ID(ctx, operation.EndDeviceIdentifiers),
			"count", len(operation.Downlinks),
		)).Debug("Handle downlink messages")
		if err := op(i.server, ctx, operation.EndDeviceIdentifiers, operation.Downlinks); err != nil {
			logger.WithError(err).Warn("Failed to handle downlink messages")
		}
	}
}

func (i *integration) startHandleDown(ctx context.Context) {
	for _, downlink := range []struct {
		name         string
		op           func(io.Server, context.Context, ttnpb.EndDeviceIdentifiers, []*ttnpb.ApplicationDownlink) error
		subscription *pubsub.Subscription
	}{
		{
			name:         "push",
			op:           io.Server.DownlinkQueuePush,
			subscription: i.subscriptions.Push,
		},
		{
			name:         "replace",
			op:           io.Server.DownlinkQueueReplace,
			subscription: i.subscriptions.Replace,
		},
	} {
		if downlink.subscription == nil {
			continue
		}
		go i.handleDown(log.NewContextWithField(ctx, "operation", downlink.name), downlink.op, downlink.subscription)
	}
}

func (i *integration) shutdown(ctx context.Context) {
	i.subscriptions.Shutdown(ctx)
	i.topics.Shutdown(ctx)
}

var errAlreadyIntegrated = errors.DefineAlreadyExists("already_integrated", "already integrated to `{application_uid} {pubsub_id}`")

func (s *srv) integrate(ctx context.Context, pb *ttnpb.ApplicationPubSub) (err error) {
	uid := unique.ID(ctx, pb.ApplicationIdentifiers)
	ctx = log.NewContextWithFields(ctx, log.Fields(
		"application_uid", uid,
		"pubsub_id", pb.PubSubID,
	))
	ctx, cancel := errorcontext.New(ctx)
	i := &integration{
		ApplicationPubSub: *pb,
		ctx:               ctx,
		cancel:            cancel,
		server:            s.server,
	}
	if _, loaded := s.integrations.LoadOrStore(pb.ApplicationPubSubIdentifiers, i); loaded {
		return errAlreadyIntegrated.WithAttributes("application_uid", uid, "pubsub_id", pb.PubSubID)
	}
	go func() {
		<-ctx.Done()
		s.integrationErrors.Store(pb.ApplicationPubSubIdentifiers, ctx.Err())
		s.integrations.Delete(pb.ApplicationPubSubIdentifiers)
		if err := ctx.Err(); err != nil && !errors.IsCanceled(err) {
			log.FromContext(ctx).WithError(err).Warn("Integration failed")
		}
	}()
	provider, err := provider.GetProvider(pb.Provider)
	if err != nil {
		return err
	}
	i.subscriptions, err = provider.OpenSubscriptions(ctx, pb)
	if err != nil {
		return err
	}
	i.topics, err = provider.OpenTopics(ctx, pb)
	if err != nil {
		return err
	}
	ctx = log.NewContextWithField(ctx, "provider", pb.Provider)
	logger := log.FromContext(ctx)
	i.sub = io.NewSubscription(s.ctx, "pubsub", &pb.ApplicationIdentifiers)
	format, ok := formats[pb.Format]
	if !ok {
		return errFormatNotFound.WithAttributes("format", pb.Format)
	}
	i.format = format
	go i.handleUp(ctx)
	i.startHandleDown(ctx)
	logger.Info("Integrated")
	<-ctx.Done()
	if err := ctx.Err(); errors.IsCanceled(err) {
		logger.Info("Integration cancelled")
		i.shutdown(ctx)
	}
	return
}

func (s *srv) cancelIntegration(ctx context.Context, ids ttnpb.ApplicationPubSubIdentifiers) error {
	if val, ok := s.integrations.Load(ids); ok {
		i := val.(*integration)
		log.FromContext(ctx).WithFields(log.Fields(
			"application_uid", ids.ApplicationIdentifiers,
			"pubsub_id", ids.PubSubID,
		)).Debug("Integration cancelled")
		i.cancel(context.Canceled)
	} else {
		s.integrationErrors.Delete(ids)
	}
	return nil
}
