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
	"sync"
	"time"

	"go.thethings.network/lorawan-stack/pkg/applicationserver/io"
	"go.thethings.network/lorawan-stack/pkg/applicationserver/io/pubsub/provider"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/errorcontext"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
	"gocloud.dev/pubsub"
)

// PubSub is an PubSub frontend that exposes ttnpb.ApplicationPubSubRegistryServer.
type PubSub struct {
	ttnpb.ApplicationPubSubRegistryServer

	*component.Component
	ctx      context.Context
	server   io.Server
	registry Registry

	integrations      sync.Map
	integrationErrors sync.Map
}

// New creates a new pusub frontend.
func New(c *component.Component, server io.Server, registry Registry) (*PubSub, error) {
	ctx := log.NewContextWithField(c.FillContext(c.Context()), "namespace", "applicationserver/io/pubsub")
	ps := &PubSub{
		Component: c,
		ctx:       ctx,
		server:    server,
		registry:  registry,
	}
	ps.RegisterTask(ctx, "pubsubs_start_all", ps.startAll, component.TaskRestartOnFailure)
	return ps, nil
}

func (ps *PubSub) startAll(ctx context.Context) error {
	return ps.registry.Range(ctx, []string{"ids"},
		func(ctx context.Context, _ ttnpb.ApplicationIdentifiers, pb *ttnpb.ApplicationPubSub) bool {
			ps.startTask(ctx, pb.ApplicationPubSubIdentifiers)
			return true
		},
	)
}

var startBackoff = []time.Duration{100 * time.Millisecond, 1 * time.Second, 10 * time.Second}

func (ps *PubSub) startTask(ctx context.Context, ids ttnpb.ApplicationPubSubIdentifiers) {
	ctx = log.NewContextWithFields(ctx, log.Fields(
		"application_uid", unique.ID(ctx, ids.ApplicationIdentifiers),
		"pub_sub_id", ids.PubSubID,
	))
	ps.StartTask(ctx, "pubsub", func(ctx context.Context) error {
		target, err := ps.registry.Get(ctx, ids, ttnpb.ApplicationPubSubFieldPathsNested)
		if err != nil {
			if !errors.IsNotFound(err) {
				log.FromContext(ctx).WithError(err).Error("Failed to stop pubsub")
			}
			return nil
		}

		err = ps.start(ctx, target)
		switch {
		case errors.IsFailedPrecondition(err),
			errors.IsUnauthenticated(err),
			errors.IsPermissionDenied(err),
			errors.IsInvalidArgument(err):
			log.FromContext(ctx).WithError(err).Warn("Failed to start")
			return nil
		case errors.IsCanceled(err),
			errors.IsAlreadyExists(err):
			return nil
		default:
			return err
		}
	}, component.TaskRestartOnFailure, 0.1, startBackoff...)
}

type integration struct {
	ttnpb.ApplicationPubSub
	ctx    context.Context
	cancel errorcontext.CancelFunc

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
			case *ttnpb.ApplicationUp_LocationSolved:
				topic = i.conn.Topics.LocationSolved
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
		if err := operation.EndDeviceIdentifiers.ValidateContext(ctx); err != nil {
			logger.WithError(err).Warn("Failed to validate downlink queue operation")
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

var errAlreadyConfigured = errors.DefineAlreadyExists("already_configured", "already configured to `{application_uid}` `{pub_sub_id}`")

func (ps *PubSub) start(ctx context.Context, pb *ttnpb.ApplicationPubSub) (err error) {
	appUID := unique.ID(ctx, pb.ApplicationIdentifiers)
	psUID := PubSubUID(appUID, pb.PubSubID)
	ctx = log.NewContextWithFields(ctx, log.Fields(
		"application_uid", appUID,
		"pub_sub_id", pb.PubSubID,
	))
	ctx = rights.NewContext(ctx, rights.Rights{
		ApplicationRights: map[string]*ttnpb.Rights{
			appUID: {
				Rights: []ttnpb.Right{ttnpb.RIGHT_APPLICATION_TRAFFIC_READ}, // Required by io.Subscribe.
			},
		},
	})
	ctx, cancel := errorcontext.New(ctx)
	defer func() {
		cancel(err)
	}()
	i := &integration{
		ApplicationPubSub: *pb,
		ctx:               ctx,
		cancel:            cancel,
		server:            ps.server,
	}
	if _, loaded := ps.integrations.LoadOrStore(psUID, i); loaded {
		return errAlreadyConfigured.WithAttributes("application_uid", appUID, "pub_sub_id", pb.PubSubID)
	}
	go func() {
		<-ctx.Done()
		ps.integrationErrors.Store(psUID, ctx.Err())
		ps.integrations.Delete(psUID)
	}()
	provider, err := provider.GetProvider(pb.Provider)
	if err != nil {
		return err
	}
	i.conn, err = provider.OpenConnection(ctx, pb)
	if err != nil {
		return err
	}
	ctx = log.NewContextWithField(ctx, "provider", fmt.Sprintf("%T", pb.Provider))
	logger := log.FromContext(ctx)
	i.sub, err = ps.server.Subscribe(ctx, "pubsub", pb.ApplicationIdentifiers)
	if err != nil {
		return err
	}
	format, ok := formats[pb.Format]
	if !ok {
		return errFormatNotFound.WithAttributes("format", pb.Format)
	}
	i.format = format
	go i.handleUp(ctx)
	i.startHandleDown(ctx)
	logger.Info("Started")
	<-ctx.Done()
	if err := ctx.Err(); errors.IsCanceled(err) {
		logger.Info("Integration canceled")
		i.conn.Shutdown(ctx)
	}
	return
}

func (ps *PubSub) stop(ctx context.Context, ids ttnpb.ApplicationPubSubIdentifiers) error {
	appUID := unique.ID(ctx, ids.ApplicationIdentifiers)
	psUID := PubSubUID(appUID, ids.PubSubID)
	if val, ok := ps.integrations.Load(psUID); ok {
		i := val.(*integration)
		log.FromContext(ctx).WithFields(log.Fields(
			"application_uid", appUID,
			"pub_sub_id", ids.PubSubID,
		)).Debug("Integration canceled")
		i.cancel(context.Canceled)
	} else {
		ps.integrationErrors.Delete(psUID)
	}
	return nil
}
