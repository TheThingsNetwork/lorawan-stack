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

package applicationserver

import (
	"context"
	"fmt"
	"time"

	"go.thethings.network/lorawan-stack/pkg/applicationserver/io"
	"go.thethings.network/lorawan-stack/pkg/component"
	errors "go.thethings.network/lorawan-stack/pkg/errorsv3"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
	"google.golang.org/grpc"
)

func (as *ApplicationServer) linkAll(ctx context.Context) error {
	return as.linkRegistry.Range(ctx, func(ids ttnpb.ApplicationIdentifiers, target *ttnpb.ApplicationLink) bool {
		as.startLinkTask(ctx, ids, target)
		return true
	})
}

var linkBackoff = []time.Duration{100 * time.Millisecond, 1 * time.Second, 10 * time.Second}

func (as *ApplicationServer) startLinkTask(ctx context.Context, ids ttnpb.ApplicationIdentifiers, target *ttnpb.ApplicationLink) {
	// TODO: Add jitter to the backoff (https://github.com/TheThingsIndustries/lorawan-stack/issues/1227)
	as.StartTask(ctx, func(ctx context.Context) error {
		err := as.link(ctx, ids, target)
		switch {
		case errors.IsFailedPrecondition(err), errors.IsUnauthenticated(err), errors.IsPermissionDenied(err):
			log.FromContext(ctx).WithError(err).Warn("Failed to link")
			return nil
		case errors.IsCanceled(err), errors.IsAlreadyExists(err):
			return nil
		default:
			return err
		}
	}, component.TaskRestartOnFailure, linkBackoff...)
}

type link struct {
	ctx           context.Context
	cancel        func()
	subscribeCh   chan *io.Connection
	unsubscribeCh chan *io.Connection
	upCh          chan *ttnpb.ApplicationUp
}

const linkBufferSize = 10

var (
	errAlreadyLinked = errors.DefineAlreadyExists("already_linked", "already linked to `{application_uid}`")
	errNSNotFound    = errors.DefineNotFound("network_server_not_found", "Network Server not found for `{application_uid}`")
	errExternalNS    = errors.DefineFailedPrecondition("external_network_server", "link to external Network Server not supported")
)

func (as *ApplicationServer) link(ctx context.Context, ids ttnpb.ApplicationIdentifiers, target *ttnpb.ApplicationLink) error {
	uid := unique.ID(ctx, ids)
	ctx = log.NewContextWithField(ctx, "application_uid", uid)
	ctx, cancelCtx := context.WithCancel(ctx)
	l := &link{
		ctx:           ctx,
		cancel:        cancelCtx,
		subscribeCh:   make(chan *io.Connection, 1),
		unsubscribeCh: make(chan *io.Connection, 1),
		upCh:          make(chan *ttnpb.ApplicationUp, linkBufferSize),
	}
	if _, loaded := as.links.LoadOrStore(uid, l); loaded {
		cancelCtx()
		return errAlreadyLinked.WithAttributes("application_uid", uid)
	}
	defer func() {
		cancelCtx()
		as.links.Delete(uid)
	}()
	var nsName string
	var conn *grpc.ClientConn
	var callOpt grpc.CallOption
	if target.NetworkServerAddress != "" {
		// TODO: Dial to external Network Server.
		// nsName = target.NetworkServerAddress
		// conn = grpc.Dial(...)
		// callOpt = grpc.PerRPCCredentials(rpcmetadata.MD{
		// 	AuthType:  "Key",
		// 	AuthValue: target.APIKey,
		// })
		return errExternalNS
	} else {
		ns := as.GetPeer(ctx, ttnpb.PeerInfo_NETWORK_SERVER, ids)
		if ns == nil {
			return errNSNotFound.WithAttributes("application_uid", unique.ID(ctx, ids))
		}
		nsName, conn, callOpt = ns.Name(), ns.Conn(), as.WithClusterAuth()
	}
	client := ttnpb.NewAsNsClient(conn)
	logger := log.FromContext(ctx)
	logger.Debug("Linking")
	stream, err := client.LinkApplication(ctx, &ids, callOpt)
	if err != nil {
		logger.WithError(err).Warn("Linking failed")
		return err
	}
	logger.Info("Linked")

	go l.run()
	for {
		up, err := stream.Recv()
		if err != nil {
			if errors.IsCanceled(err) {
				logger.Debug("Unlinked")
			} else {
				logger.WithError(err).Warn("Link failed")
			}
			return err
		}
		ctx := events.ContextWithCorrelationID(ctx, fmt.Sprintf("uplink:%s", events.NewCorrelationID()))
		registerReceiveUplink(ctx, up, nsName)
		if err := as.processUp(ctx, up); err != nil {
			logger.WithError(err).Warn("Failed to process upstream message")
			registerDropUplink(ctx, up, err)
			continue
		}
		l.upCh <- up
		registerForwardUplink(ctx, up)
	}
}

var errNotLinked = errors.DefineNotFound("not_linked", "not linked to `{application_uid}`")

func (as *ApplicationServer) cancelLink(ctx context.Context, ids ttnpb.ApplicationIdentifiers) error {
	uid := unique.ID(ctx, ids)
	val, ok := as.links.Load(uid)
	if !ok {
		return errNotLinked.WithAttributes("application_uid", uid)
	}
	l := val.(*link)
	log.FromContext(ctx).WithField("application_uid", uid).Debug("Unlinking")
	l.cancel()
	<-l.ctx.Done()
	as.links.Delete(uid)
	return nil
}

func (l *link) run() {
	subscribers := make(map[*io.Connection]string)
	for {
		select {
		case <-l.ctx.Done():
			return
		case conn := <-l.subscribeCh:
			correlationID := fmt.Sprintf("subscriber:%s", events.NewCorrelationID())
			subscribers[conn] = correlationID
			registerSubscribe(events.ContextWithCorrelationID(l.ctx, correlationID), conn)
			log.FromContext(conn.Context()).Debug("Subscribed")
		case conn := <-l.unsubscribeCh:
			if correlationID, ok := subscribers[conn]; ok {
				delete(subscribers, conn)
				registerUnsubscribe(events.ContextWithCorrelationID(l.ctx, correlationID), conn)
				log.FromContext(conn.Context()).Debug("Unsubscribed")
			}
		case up := <-l.upCh:
			for conn := range subscribers {
				if err := conn.SendUp(up); err != nil {
					log.FromContext(conn.Context()).WithError(err).Warn("Send upstream message failed")
				}
			}
		}
	}
}
