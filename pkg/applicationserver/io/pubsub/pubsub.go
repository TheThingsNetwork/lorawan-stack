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

// Package pubsub implements the go-cloud pub/sub frontend.
package pubsub

import (
	"context"
	"fmt"
	"sync"

	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/pubsub/provider"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	"go.thethings.network/lorawan-stack/v3/pkg/errorcontext"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/task"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"gocloud.dev/pubsub"
)

// PubSub is an pub/sub frontend that exposes ttnpb.ApplicationPubSubRegistryServer.
type PubSub struct {
	ttnpb.ApplicationPubSubRegistryServer

	*component.Component
	ctx      context.Context
	server   io.Server
	registry Registry

	integrations sync.Map

	providerStatuses ProviderStatuses
}

// New creates a new pusub frontend.
func New(c *component.Component, server io.Server, registry Registry, providerStatuses ProviderStatuses) (*PubSub, error) {
	ctx := log.NewContextWithField(c.Context(), "namespace", "applicationserver/io/pubsub")
	ps := &PubSub{
		Component: c,
		ctx:       ctx,
		server:    server,
		registry:  registry,

		providerStatuses: providerStatuses,
	}
	ps.RegisterTask(&task.Config{
		Context: ctx,
		ID:      "pubsubs_start_all",
		Func:    ps.startAll,
		Restart: task.RestartOnFailure,
		Backoff: task.DefaultBackoffConfig,
	})
	return ps, nil
}

func (ps *PubSub) startAll(ctx context.Context) error {
	return ps.registry.Range(ctx, []string{"ids"},
		func(ctx context.Context, _ *ttnpb.ApplicationIdentifiers, pb *ttnpb.ApplicationPubSub) bool {
			ps.startTask(ctx, pb.Ids)
			return true
		},
	)
}

func (ps *PubSub) startTask(ctx context.Context, ids *ttnpb.ApplicationPubSubIdentifiers) {
	ctx = log.NewContextWithFields(ctx, log.Fields(
		"application_uid", unique.ID(ctx, ids.ApplicationIds),
		"pub_sub_id", ids.PubSubId,
	))
	ps.StartTask(&task.Config{
		Context: ctx,
		ID:      "pubsub",
		Func: func(ctx context.Context) error {
			target, err := ps.registry.Get(ctx, ids, ttnpb.ApplicationPubSubFieldPathsNested)
			if err != nil && !errors.IsNotFound(err) {
				return err
			} else if err != nil {
				log.FromContext(ctx).WithError(err).Warn("Pub/Sub not found")
				return nil
			}

			if err := ps.providerStatuses.Enabled(ctx, target.Provider); err != nil {
				log.FromContext(ctx).WithError(err).Debug("Pub/Sub not enabled")
				return nil
			}

			return ps.start(ctx, target)
		},
		Restart: task.RestartOnFailure,
		Backoff: io.DialTaskBackoffConfig,
	})
}

type integration struct {
	ttnpb.ApplicationPubSub
	ctx    context.Context
	cancel errorcontext.CancelFunc
	closed chan struct{}

	conn *provider.Connection

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
				topic = i.conn.Topics.UplinkMessage
			case *ttnpb.ApplicationUp_JoinAccept:
				topic = i.conn.Topics.JoinAccept
			case *ttnpb.ApplicationUp_DownlinkAck:
				topic = i.conn.Topics.DownlinkAck
			case *ttnpb.ApplicationUp_DownlinkNack:
				topic = i.conn.Topics.DownlinkNack
			case *ttnpb.ApplicationUp_DownlinkSent:
				topic = i.conn.Topics.DownlinkSent
			case *ttnpb.ApplicationUp_DownlinkFailed:
				topic = i.conn.Topics.DownlinkFailed
			case *ttnpb.ApplicationUp_DownlinkQueued:
				topic = i.conn.Topics.DownlinkQueued
			case *ttnpb.ApplicationUp_DownlinkQueueInvalidated:
				topic = i.conn.Topics.DownlinkQueueInvalidated
			case *ttnpb.ApplicationUp_LocationSolved:
				topic = i.conn.Topics.LocationSolved
			case *ttnpb.ApplicationUp_ServiceData:
				topic = i.conn.Topics.ServiceData
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
				i.cancel(err)
				return
			}
			logger.Debug("Publish upstream message")
		}
	}
}

func (i *integration) handleDown(ctx context.Context, op func(io.Server, context.Context, *ttnpb.EndDeviceIdentifiers, []*ttnpb.ApplicationDownlink) error, subscription *pubsub.Subscription) {
	logger := log.FromContext(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		msg, err := subscription.Receive(ctx)
		if err != nil {
			logger.WithError(err).Warn("Failed to receive downlink queue operation")
			i.cancel(err)
			return
		}
		msg.Ack()
		operation, err := i.format.ToDownlinkQueueRequest(msg.Body)
		if err != nil {
			logger.WithError(err).Warn("Failed to decode downlink queue operation")
			continue
		}
		if err := operation.ValidateFields(); err != nil {
			logger.WithError(err).Warn("Failed to validate downlink queue operation")
			continue
		}
		if err := operation.EndDeviceIds.ValidateContext(ctx); err != nil {
			logger.WithError(err).Warn("Failed to validate downlink queue operation")
			continue
		}
		logger.WithFields(log.Fields(
			"device_uid", unique.ID(ctx, operation.EndDeviceIds),
			"count", len(operation.Downlinks),
		)).Debug("Handle downlink messages")
		if err := op(i.server, ctx, operation.EndDeviceIds, operation.Downlinks); err != nil {
			logger.WithError(err).Warn("Failed to handle downlink messages")
		}
	}
}

func (i *integration) startHandleDown(ctx context.Context) {
	for _, downlink := range []struct {
		name         string
		op           func(io.Server, context.Context, *ttnpb.EndDeviceIdentifiers, []*ttnpb.ApplicationDownlink) error
		subscription *pubsub.Subscription
	}{
		{
			name:         "push",
			op:           io.Server.DownlinkQueuePush,
			subscription: i.conn.Subscriptions.Push,
		},
		{
			name:         "replace",
			op:           io.Server.DownlinkQueueReplace,
			subscription: i.conn.Subscriptions.Replace,
		},
	} {
		if downlink.subscription == nil {
			continue
		}
		go i.handleDown(log.NewContextWithField(ctx, "operation", downlink.name), downlink.op, downlink.subscription)
	}
}

func (ps *PubSub) start(ctx context.Context, pb *ttnpb.ApplicationPubSub) (err error) {
	appUID := unique.ID(ctx, pb.Ids.ApplicationIds)
	psUID := PubSubUID(appUID, pb.Ids.PubSubId)

	ctx = log.NewContextWithFields(ctx, log.Fields(
		"application_uid", appUID,
		"pub_sub_id", pb.Ids.PubSubId,
		"provider", fmt.Sprintf("%T", pb.Provider),
	))
	logger := log.FromContext(ctx)
	ctx, cancel := errorcontext.New(ctx)
	defer func() {
		cancel(err)
	}()

	i := &integration{
		ApplicationPubSub: *pb,
		ctx:               ctx,
		cancel:            cancel,
		closed:            make(chan struct{}),
		server:            ps.server,
	}
	defer close(i.closed)
	if _, loaded := ps.integrations.LoadOrStore(psUID, i); loaded {
		logger.Debug("Pub/sub already started")
		return nil
	}
	defer ps.integrations.Delete(psUID)

	defer func() {
		if err != nil {
			logger.WithError(err).Warn("Pub/sub failed")
			registerIntegrationFail(ctx, i, err)
		}
	}()

	provider, err := provider.GetProvider(pb)
	if err != nil {
		return err
	}
	if err := ps.providerStatuses.Enabled(ctx, pb.GetProvider()); err != nil {
		return err
	}
	i.conn, err = provider.OpenConnection(ctx, pb, ps.providerStatuses)
	if err != nil {
		return err
	}
	defer func() {
		if err := i.conn.Shutdown(ctx); err != nil {
			logger.WithError(err).Warn("Failed to shutdown pub/sub connection")
		} else {
			logger.Debug("Shutdown pub/sub connection success")
		}
	}()
	i.sub, err = ps.server.Subscribe(ctx, "pubsub", pb.Ids.ApplicationIds, false)
	if err != nil {
		return err
	}
	go func() {
		// Close the integration if the subscription is canceled.
		select {
		case <-i.sub.Context().Done():
			err := i.sub.Context().Err()
			cancel(err)
		case <-ctx.Done():
		}
	}()
	var ok bool
	if i.format, ok = formats[pb.Format]; !ok {
		return errFormatNotFound.WithAttributes("format", pb.Format)
	}

	go i.handleUp(ctx)
	i.startHandleDown(ctx)
	logger.Info("Pub/sub started")
	registerIntegrationStart(ctx, i)
	defer func() {
		logger.Info("Pub/sub stopped")
		registerIntegrationStop(ctx, i)
	}()

	<-ctx.Done()
	return ctx.Err()
}

func (ps *PubSub) stop(ctx context.Context, ids *ttnpb.ApplicationPubSubIdentifiers) error {
	appUID := unique.ID(ctx, ids.ApplicationIds)
	psUID := PubSubUID(appUID, ids.PubSubId)
	if val, ok := ps.integrations.Load(psUID); ok {
		i := val.(*integration)
		i.cancel(context.Canceled)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-i.closed:
		}
	}
	return nil
}
