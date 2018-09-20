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
	"sync"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"go.thethings.network/lorawan-stack/pkg/applicationserver/io"
	iogrpc "go.thethings.network/lorawan-stack/pkg/applicationserver/io/grpc"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/crypto"
	"go.thethings.network/lorawan-stack/pkg/crypto/cryptoutil"
	errors "go.thethings.network/lorawan-stack/pkg/errorsv3"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/messageprocessors"
	"go.thethings.network/lorawan-stack/pkg/messageprocessors/cayennelpp"
	"go.thethings.network/lorawan-stack/pkg/messageprocessors/javascript"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
	"google.golang.org/grpc"
)

// ApplicationServer implements the Application Server component.
//
// The Application Server exposes the As, AppAs and AsEndDeviceRegistry services.
type ApplicationServer struct {
	*component.Component

	linkMode          LinkMode
	linkRegistry      LinkRegistry
	deviceRegistry    DeviceRegistry
	keyVault          crypto.KeyVault
	payloadFormatters map[ttnpb.PayloadFormatter]messageprocessors.PayloadEncodeDecoder

	links sync.Map
}

// New returns new *ApplicationServer.
func New(c *component.Component, conf *Config) (*ApplicationServer, error) {
	as := &ApplicationServer{
		Component:      c,
		linkMode:       conf.LinkMode,
		linkRegistry:   conf.Links,
		deviceRegistry: conf.Devices,
		keyVault:       conf.KeyVault,
		payloadFormatters: map[ttnpb.PayloadFormatter]messageprocessors.PayloadEncodeDecoder{
			ttnpb.PayloadFormatter_FORMATTER_JAVASCRIPT: javascript.New(),
			ttnpb.PayloadFormatter_FORMATTER_CAYENNELPP: cayennelpp.New(),
		},
	}

	c.RegisterGRPC(as)
	if conf.LinkMode == LinkAll {
		c.RegisterTask(as.linkAll, component.TaskRestartOnFailure)
	}
	return as, nil
}

// RegisterServices registers services provided by as at s.
func (as *ApplicationServer) RegisterServices(s *grpc.Server) {
	ttnpb.RegisterAsServer(s, as)
	// TODO: Register AsEndDeviceRegistryServer (https://github.com/TheThingsIndustries/lorawan-stack/issues/1117)
	ttnpb.RegisterAppAsServer(s, iogrpc.New(as))
}

// RegisterHandlers registers gRPC handlers.
func (as *ApplicationServer) RegisterHandlers(s *runtime.ServeMux, conn *grpc.ClientConn) {
	ttnpb.RegisterAsHandler(as.Context(), s, conn)
	// TODO: Register AsEndDeviceRegistryHandler (https://github.com/TheThingsIndustries/lorawan-stack/issues/1117)
}

// Roles returns the roles that the Application Server fulfills.
func (as *ApplicationServer) Roles() []ttnpb.PeerInfo_Role {
	return []ttnpb.PeerInfo_Role{ttnpb.PeerInfo_APPLICATION_SERVER}
}

// Connect connects an application or integration by its identifiers to the Application Server, and returns a
// io.Connection for traffic and control.
func (as *ApplicationServer) Connect(ctx context.Context, protocol string, ids ttnpb.ApplicationIdentifiers) (*io.Connection, error) {
	if err := rights.RequireApplication(ctx, ids, ttnpb.RIGHT_APPLICATION_TRAFFIC_READ); err != nil {
		return nil, err
	}

	uid := unique.ID(ctx, ids)
	logger := log.FromContext(ctx).WithField("application_uid", uid)
	ctx = events.ContextWithCorrelationID(ctx, fmt.Sprintf("application_conn:%s", events.NewCorrelationID()))

	val, ok := as.links.Load(uid)
	if !ok {
		return nil, errNotLinked.WithAttributes("application_uid", uid)
	}
	l := val.(*link)
	conn := io.NewConnection(ctx, protocol, ids)
	l.subscribeCh <- conn
	go func() {
		<-ctx.Done()
		l.unsubscribeCh <- conn
	}()
	logger.Info("Application connected")
	return conn, nil
}

var errJSNotFound = errors.DefineNotFound("join_server_not_found", "Join Server not found for JoinEUI `{join_eui}`")

func (as *ApplicationServer) getAppSKey(ctx context.Context, sessionKeyID string, ids ttnpb.EndDeviceIdentifiers) (ttnpb.KeyEnvelope, error) {
	js := as.GetPeer(ctx, ttnpb.PeerInfo_JOIN_SERVER, ids)
	if js == nil {
		return ttnpb.KeyEnvelope{}, errJSNotFound.WithAttributes("join_eui", *ids.JoinEUI)
	}
	client := ttnpb.NewAsJsClient(js.Conn())
	req := &ttnpb.SessionKeyRequest{
		SessionKeyID: sessionKeyID,
		DevEUI:       *ids.DevEUI,
	}
	res, err := client.GetAppSKey(ctx, req, as.WithClusterAuth())
	if err != nil {
		return ttnpb.KeyEnvelope{}, err
	}
	return res.AppSKey, nil
}

var (
	errDeviceNotFound = errors.DefineNotFound("device_not_found", "device `{device_uid}` not found")
	errGetAppSKey     = errors.Define("app_s_key", "failed to get AppSKey")
	errMissingAppSKey = errors.DefineCorruption("missing_app_s_key", "AppSKey is unknown")
)

func (as *ApplicationServer) processUp(ctx context.Context, up *ttnpb.ApplicationUp, link *ttnpb.ApplicationLink) error {
	return as.deviceRegistry.Set(ctx, up.EndDeviceIdentifiers, func(ed *ttnpb.EndDevice) (*ttnpb.EndDevice, error) {
		uid := unique.ID(ctx, up.EndDeviceIdentifiers)
		logger := log.FromContext(ctx).WithField("device_uid", uid)
		creating := false
		if ed == nil {
			logger.Info("Creating new device")
			ed = &ttnpb.EndDevice{
				EndDeviceIdentifiers: up.EndDeviceIdentifiers,
			}
			creating = true
		}
		resetDownlinkQueue := false

		switch p := up.Up.(type) {
		case *ttnpb.ApplicationUp_JoinAccept:
			logger := logger.WithFields(log.Fields(
				"join_eui", up.EndDeviceIdentifiers.JoinEUI,
				"dev_eui", up.EndDeviceIdentifiers.DevEUI,
				"session_key_id", p.JoinAccept.SessionKeyID,
			))
			logger.Debug("Handling join-accept...")
			var appSKey ttnpb.KeyEnvelope
			if p.JoinAccept.AppSKey != nil {
				logger.Debug("Received AppSKey from Network Server")
				appSKey = *p.JoinAccept.AppSKey
			} else {
				logger.Debug("Getting AppSKey from Join Server...")
				key, err := as.getAppSKey(ctx, p.JoinAccept.SessionKeyID, up.EndDeviceIdentifiers)
				if err != nil {
					return nil, errGetAppSKey.WithCause(err)
				}
				appSKey = key
				logger.Debug("Received AppSKey from Join Server")
			}
			ed.Session = &ttnpb.Session{
				SessionKeys: ttnpb.SessionKeys{
					SessionKeyID: p.JoinAccept.SessionKeyID,
					AppSKey:      &appSKey,
				},
				StartedAt: time.Now(),
			}
			p.JoinAccept.AppSKey = nil
			resetDownlinkQueue = true
			logger.Info("Handled join-accept")

		case *ttnpb.ApplicationUp_UplinkMessage:
			logger := logger.WithField("session_key_id", p.UplinkMessage.SessionKeyID)
			logger.Debug("Handling uplink data...")
			if ed.Session == nil || ed.Session.SessionKeyID != p.UplinkMessage.SessionKeyID {
				if !creating {
					logger.Warn("Session mismatch; restoring session...")
				}
				appSKey, err := as.getAppSKey(ctx, p.UplinkMessage.SessionKeyID, up.EndDeviceIdentifiers)
				if err != nil {
					return nil, errGetAppSKey.WithCause(err)
				}
				ed.Session = &ttnpb.Session{
					SessionKeys: ttnpb.SessionKeys{
						SessionKeyID: p.UplinkMessage.SessionKeyID,
						AppSKey:      &appSKey,
					},
					StartedAt: time.Now(),
				}
				logger.Debug("Session restored")
			} else if ed.Session.AppSKey == nil {
				return nil, errMissingAppSKey
			}
			appSKey, err := cryptoutil.UnwrapAES128Key(*ed.Session.AppSKey, as.keyVault)
			if err != nil {
				return nil, err
			}
			frmPayload, err := crypto.DecryptUplink(appSKey, *up.DevAddr, p.UplinkMessage.FCnt, p.UplinkMessage.FRMPayload)
			if err != nil {
				return nil, err
			}
			p.UplinkMessage.FRMPayload = frmPayload
			formatters := ed.Formatters
			if formatters == nil {
				formatters = link.DefaultFormatters
			}
			if formatters.GetUpFormatter() != ttnpb.PayloadFormatter_FORMATTER_NONE {
				logger := logger.WithField("formatter", formatters.UpFormatter)
				if formatter, ok := as.payloadFormatters[formatters.UpFormatter]; ok {
					if err := formatter.Decode(ctx, p.UplinkMessage, ed.VersionIDs, formatters.UpFormatterParameter); err != nil {
						logger.WithError(err).Warn("Failed to decode payload")
					}
				} else {
					logger.Warn("Payload formatter not supported")
				}
			}
			// TODO:
			// - Run uplink messages through location solvers async
			logger.Info("Handled uplink data")
		}

		_ = resetDownlinkQueue

		// TODO:
		// - Recompute downlink queue on join accept and invalidation

		return ed, nil
	})
}
