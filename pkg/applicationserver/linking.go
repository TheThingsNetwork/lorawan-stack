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

package applicationserver

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"go.thethings.network/lorawan-stack/pkg/applicationserver/io"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/errorcontext"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/rpcclient"
	"go.thethings.network/lorawan-stack/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
	"google.golang.org/grpc"
)

func (as *ApplicationServer) linkAll(ctx context.Context) error {
	return as.linkRegistry.Range(ctx, nil,
		func(ctx context.Context, ids ttnpb.ApplicationIdentifiers, _ *ttnpb.ApplicationLink) bool {
			as.startLinkTask(ctx, ids)
			return true
		},
	)
}

var linkBackoff = []time.Duration{100 * time.Millisecond, 1 * time.Second, 10 * time.Second}

func (as *ApplicationServer) startLinkTask(ctx context.Context, ids ttnpb.ApplicationIdentifiers) {
	ctx = log.NewContextWithField(ctx, "application_uid", unique.ID(ctx, ids))
	as.StartTask(ctx, "link", func(ctx context.Context) error {
		target, err := as.linkRegistry.Get(ctx, ids, []string{
			"network_server_address",
			"api_key",
			"default_formatters",
		})
		if err != nil {
			if !errors.IsNotFound(err) {
				log.FromContext(ctx).WithError(err).Error("Failed to get link")
			}
			return nil
		}

		err = as.link(ctx, ids, target)
		switch {
		case errors.IsFailedPrecondition(err),
			errors.IsUnauthenticated(err),
			errors.IsPermissionDenied(err),
			errors.IsInvalidArgument(err):
			log.FromContext(ctx).WithError(err).Warn("Failed to link")
			return nil
		case errors.IsCanceled(err),
			errors.IsAlreadyExists(err):
			return nil
		default:
			return err
		}
	}, component.TaskRestartOnFailure, 0.1, linkBackoff...)
}

type link struct {
	// Align for sync/atomic.
	ups,
	downlinks uint64
	linkTime,
	lastUpTime,
	lastDownlinkTime int64

	ttnpb.ApplicationIdentifiers
	ttnpb.ApplicationLink
	ctx    context.Context
	cancel errorcontext.CancelFunc

	conn      *grpc.ClientConn
	connName  string
	connReady chan struct{}
	callOpts  []grpc.CallOption

	subscribeCh   chan *io.Subscription
	unsubscribeCh chan *io.Subscription
	upCh          chan *io.ContextualApplicationUp
}

const linkBufferSize = 10

var (
	errAlreadyLinked  = errors.DefineAlreadyExists("already_linked", "already linked to `{application_uid}`")
	errNSPeerNotFound = errors.DefineNotFound("network_server_not_found", "Network Server not found for `{application_uid}`")
)

func (as *ApplicationServer) connectLink(ctx context.Context, link *link) error {
	var allowInsecure bool
	if link.NetworkServerAddress != "" {
		options := rpcclient.DefaultDialOptions(ctx)
		if allowInsecure = as.AllowInsecureForCredentials(); allowInsecure {
			options = append(options, grpc.WithInsecure())
		}
		conn, err := grpc.DialContext(ctx, link.NetworkServerAddress, options...)
		if err != nil {
			return err
		}
		link.conn = conn
		link.connName = link.NetworkServerAddress
		go func() {
			<-ctx.Done()
			conn.Close()
		}()
	} else {
		allowInsecure = !as.ClusterTLS()
		ns, err := as.GetPeer(ctx, ttnpb.ClusterRole_NETWORK_SERVER, link.ApplicationIdentifiers)
		if err != nil {
			return errNSPeerNotFound.WithCause(err).WithAttributes("application_uid", unique.ID(ctx, link.ApplicationIdentifiers))
		}
		link.conn = ns.Conn()
		link.connName = ns.Name()
	}
	link.callOpts = []grpc.CallOption{
		grpc.PerRPCCredentials(rpcmetadata.MD{
			ID:            link.ApplicationID,
			AuthType:      "Bearer",
			AuthValue:     link.APIKey,
			AllowInsecure: allowInsecure,
		}),
	}
	link.linkTime = time.Now().UnixNano()
	close(link.connReady)
	return nil
}

func (as *ApplicationServer) link(ctx context.Context, ids ttnpb.ApplicationIdentifiers, target *ttnpb.ApplicationLink) (err error) {
	uid := unique.ID(ctx, ids)
	ctx = log.NewContextWithField(ctx, "application_uid", uid)
	ctx, cancel := errorcontext.New(ctx)
	defer func() {
		cancel(err)
	}()
	l := &link{
		ApplicationIdentifiers: ids,
		ApplicationLink:        *target,
		ctx:                    ctx,
		cancel:                 cancel,
		connReady:              make(chan struct{}),
		subscribeCh:            make(chan *io.Subscription, 1),
		unsubscribeCh:          make(chan *io.Subscription, 1),
		upCh:                   make(chan *io.ContextualApplicationUp, linkBufferSize),
	}
	if _, loaded := as.links.LoadOrStore(uid, l); loaded {
		return errAlreadyLinked.WithAttributes("application_uid", uid)
	}
	go func() {
		<-ctx.Done()
		as.linkErrors.Store(uid, ctx.Err())
		as.links.Delete(uid)
		if err := ctx.Err(); err != nil && !errors.IsCanceled(err) {
			log.FromContext(ctx).WithError(err).Warn("Link failed")
			registerLinkFail(ctx, l, err)
		}
	}()
	if err := as.connectLink(ctx, l); err != nil {
		return err
	}
	client := ttnpb.NewAsNsClient(l.conn)
	ctx = log.NewContextWithField(ctx, "network_server", l.connName)
	logger := log.FromContext(ctx)
	logger.Debug("Link")
	stream, err := client.LinkApplication(ctx, l.callOpts...)
	if err != nil {
		logger.WithError(err).Warn("Link setup failed")
		return err
	}
	logger.Info("Linked")
	registerLinkStart(ctx, l)
	go func() {
		<-ctx.Done()
		if err := ctx.Err(); errors.IsCanceled(err) {
			logger.Info("Unlinked")
			registerLinkStop(ctx, l)
		}
	}()

	go l.run()
	for _, sub := range as.defaultSubscribers {
		sub := sub
		l.subscribeCh <- sub
		go func() {
			<-sub.Context().Done()
			l.unsubscribeCh <- sub
		}()
	}
	for {
		up, err := stream.Recv()
		if err != nil {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				return err
			}
		}
		atomic.AddUint64(&l.ups, 1)
		atomic.StoreInt64(&l.lastUpTime, time.Now().UnixNano())

		ctx := events.ContextWithCorrelationID(ctx, append(up.CorrelationIDs, fmt.Sprintf("as:up:%s", events.NewCorrelationID()))...)
		up.CorrelationIDs = events.CorrelationIDsFromContext(ctx)
		registerReceiveUp(ctx, up, l.connName)

		handleUpErr := as.handleUp(ctx, up, l)
		if err := stream.Send(ttnpb.Empty); err != nil {
			return err
		}

		switch p := up.Up.(type) {
		case *ttnpb.ApplicationUp_JoinAccept:
			p.JoinAccept.AppSKey = nil
			p.JoinAccept.InvalidatedDownlinks = nil
		case *ttnpb.ApplicationUp_DownlinkQueueInvalidated:
			continue
		}

		if handleUpErr != nil {
			logger.WithError(handleUpErr).Warn("Failed to process upstream message")
			registerDropUp(ctx, up, handleUpErr)
			continue
		}

		l.upCh <- &io.ContextualApplicationUp{
			Context:       ctx,
			ApplicationUp: up,
		}
		registerForwardUp(ctx, up)
	}
}

var (
	errNotLinked  = errors.DefineNotFound("not_linked", "not linked to `{application_uid}`")
	errLinkFailed = errors.DefineAborted("link", "link failed")
)

func (as *ApplicationServer) cancelLink(ctx context.Context, ids ttnpb.ApplicationIdentifiers) error {
	uid := unique.ID(ctx, ids)
	if val, ok := as.links.Load(uid); ok {
		l := val.(*link)
		log.FromContext(ctx).WithField("application_uid", uid).Debug("Unlink")
		l.cancel(context.Canceled)
	} else {
		as.linkErrors.Delete(uid)
	}
	return nil
}

func (as *ApplicationServer) getLink(ctx context.Context, ids ttnpb.ApplicationIdentifiers) (*link, error) {
	uid := unique.ID(ctx, ids)
	val, ok := as.links.Load(uid)
	if !ok {
		if val, ok := as.linkErrors.Load(uid); ok {
			if err := val.(error); !errors.IsCanceled(err) {
				return nil, errLinkFailed.WithCause(err)
			}
		}
		return nil, errNotLinked.WithAttributes("application_uid", uid)
	}
	return val.(*link), nil
}

func (l *link) run() {
	subscribers := make(map[*io.Subscription]string)
	for {
		select {
		case <-l.ctx.Done():
			return
		case sub := <-l.subscribeCh:
			correlationID := fmt.Sprintf("as:subscriber:%s", events.NewCorrelationID())
			subscribers[sub] = correlationID
			registerSubscribe(events.ContextWithCorrelationID(l.ctx, correlationID), sub)
			log.FromContext(sub.Context()).Debug("Subscribed")
		case sub := <-l.unsubscribeCh:
			if correlationID, ok := subscribers[sub]; ok {
				delete(subscribers, sub)
				registerUnsubscribe(events.ContextWithCorrelationID(l.ctx, correlationID), sub)
				log.FromContext(sub.Context()).Debug("Unsubscribed")
			}
		case up := <-l.upCh:
			for sub := range subscribers {
				if err := sub.SendUp(up.Context, up.ApplicationUp); err != nil {
					log.FromContext(sub.Context()).WithError(err).Warn("Send upstream message failed")
				}
			}
		}
	}
}

// GetLinkTime returns the timestamp when the link got established.
func (l *link) GetLinkTime() time.Time { return time.Unix(0, l.linkTime) }

// GetUpStats returns the upstream statistics.
func (l *link) GetUpStats() (total uint64, t time.Time, ok bool) {
	total = atomic.LoadUint64(&l.ups)
	if ok = total > 0; ok {
		t = time.Unix(0, atomic.LoadInt64(&l.lastUpTime))
	}
	return
}

// GetDownlinkStats returns the downlink statistics.
func (l *link) GetDownlinkStats() (total uint64, t time.Time, ok bool) {
	total = atomic.LoadUint64(&l.downlinks)
	if ok = total > 0; ok {
		t = time.Unix(0, atomic.LoadInt64(&l.lastDownlinkTime))
	}
	return
}
