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
	"go.thethings.network/lorawan-stack/pkg/errors"
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

	linkMode       LinkMode
	linkRegistry   LinkRegistry
	deviceRegistry DeviceRegistry
	formatter      payloadFormatter

	links sync.Map
}

// New returns new *ApplicationServer.
func New(c *component.Component, conf *Config) (*ApplicationServer, error) {
	linkMode, err := conf.GetLinkMode()
	if err != nil {
		return nil, err
	}
	as := &ApplicationServer{
		Component:      c,
		linkMode:       linkMode,
		linkRegistry:   conf.Links,
		deviceRegistry: conf.Devices,
		formatter: payloadFormatter{
			repository: conf.DeviceRepository.Client(),
			upFormatters: map[ttnpb.PayloadFormatter]messageprocessors.PayloadDecoder{
				ttnpb.PayloadFormatter_FORMATTER_JAVASCRIPT: javascript.New(),
				ttnpb.PayloadFormatter_FORMATTER_CAYENNELPP: cayennelpp.New(),
			},
			downFormatters: map[ttnpb.PayloadFormatter]messageprocessors.PayloadEncoder{
				ttnpb.PayloadFormatter_FORMATTER_JAVASCRIPT: javascript.New(),
				ttnpb.PayloadFormatter_FORMATTER_CAYENNELPP: cayennelpp.New(),
			},
		},
	}

	c.RegisterGRPC(as)
	if as.linkMode == LinkAll {
		c.RegisterTask(as.linkAll, component.TaskRestartOnFailure)
	}
	return as, nil
}

// RegisterServices registers services provided by as at s.
func (as *ApplicationServer) RegisterServices(s *grpc.Server) {
	ttnpb.RegisterAsServer(s, as)
	ttnpb.RegisterAsEndDeviceRegistryServer(s, as)
	ttnpb.RegisterAppAsServer(s, iogrpc.New(as))
}

// RegisterHandlers registers gRPC handlers.
func (as *ApplicationServer) RegisterHandlers(s *runtime.ServeMux, conn *grpc.ClientConn) {
	ttnpb.RegisterAsHandler(as.Context(), s, conn)
	ttnpb.RegisterAsEndDeviceRegistryHandler(as.Context(), s, conn)
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

// DownlinkQueuePush pushes the given downlink messages to the end device's application downlink queue.
func (as *ApplicationServer) DownlinkQueuePush(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, items []*ttnpb.ApplicationDownlink) error {
	return errors.New("not implemented")
}

// DownlinkQueueReplace replaces the end device's application downlink queue with the given downlink messages.
func (as *ApplicationServer) DownlinkQueueReplace(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, items []*ttnpb.ApplicationDownlink) error {
	return errors.New("not implemented")
}

// DownlinkQueueList lists the application downlink queue of the given end device.
func (as *ApplicationServer) DownlinkQueueList(ctx context.Context, ids ttnpb.EndDeviceIdentifiers) ([]*ttnpb.ApplicationDownlink, error) {
	return nil, errors.New("not implemented")
}

var errJSUnavailable = errors.DefineUnavailable("join_server_unavailable", "Join Server unavailable for JoinEUI `{join_eui}`")

func (as *ApplicationServer) getAppSKey(ctx context.Context, sessionKeyID string, ids ttnpb.EndDeviceIdentifiers) (ttnpb.KeyEnvelope, error) {
	js := as.GetPeer(ctx, ttnpb.PeerInfo_JOIN_SERVER, ids)
	if js == nil {
		return ttnpb.KeyEnvelope{}, errJSUnavailable.WithAttributes("join_eui", *ids.JoinEUI)
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

func (as *ApplicationServer) handleUp(ctx context.Context, up *ttnpb.ApplicationUp, link *ttnpb.ApplicationLink) error {
	ctx = log.NewContextWithField(ctx, "device_uid", unique.ID(ctx, up.EndDeviceIdentifiers))
	switch p := up.Up.(type) {
	case *ttnpb.ApplicationUp_JoinAccept:
		return as.handleJoinAccept(ctx, up.EndDeviceIdentifiers, p.JoinAccept, link)
	case *ttnpb.ApplicationUp_UplinkMessage:
		return as.handleUplink(ctx, up.EndDeviceIdentifiers, p.UplinkMessage, link)
	default:
		return nil
	}
}

var (
	errGetAppSKey = errors.Define("app_s_key", "failed to get AppSKey")
	errNoAppSKey  = errors.DefineCorruption("no_app_s_key", "no AppSKey")
)

func (as *ApplicationServer) handleJoinAccept(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, joinAccept *ttnpb.ApplicationJoinAccept, link *ttnpb.ApplicationLink) error {
	logger := log.FromContext(ctx).WithFields(log.Fields(
		"join_eui", ids.JoinEUI,
		"dev_eui", ids.DevEUI,
		"session_key_id", joinAccept.SessionKeyID,
	))
	_, err := as.deviceRegistry.Set(ctx, ids, nil,
		func(dev *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
			var mask []string
			if dev == nil {
				logger.Info("Creating new device")
				dev = &ttnpb.EndDevice{
					EndDeviceIdentifiers: ids,
				}
				mask = append(mask, "ids")
			}
			var appSKey ttnpb.KeyEnvelope
			if joinAccept.AppSKey != nil {
				logger.Debug("Received AppSKey from Network Server")
				appSKey = *joinAccept.AppSKey
			} else {
				logger.Debug("Getting AppSKey from Join Server...")
				key, err := as.getAppSKey(ctx, joinAccept.SessionKeyID, ids)
				if err != nil {
					return nil, nil, errGetAppSKey.WithCause(err)
				}
				appSKey = key
				logger.Debug("Received AppSKey from Join Server")
			}
			dev.Session = &ttnpb.Session{
				DevAddr: *ids.DevAddr,
				SessionKeys: ttnpb.SessionKeys{
					SessionKeyID: joinAccept.SessionKeyID,
					AppSKey:      &appSKey,
				},
				StartedAt: time.Now(),
			}
			return dev, append(mask, "session"), nil
		},
	)
	if err != nil {
		return err
	}
	// TODO: Reset downlink queue
	joinAccept.AppSKey = nil
	return nil
}

func (as *ApplicationServer) handleUplink(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, uplink *ttnpb.ApplicationUplink, link *ttnpb.ApplicationLink) error {
	logger := log.FromContext(ctx).WithField("session_key_id", uplink.SessionKeyID)
	dev, err := as.deviceRegistry.Set(ctx, ids,
		[]string{
			"session",
			"formatters",
			"version_ids",
		},
		func(dev *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
			var mask []string
			if dev == nil {
				logger.Info("Creating new device")
				dev = &ttnpb.EndDevice{
					EndDeviceIdentifiers: ids,
				}
				mask = append(mask, "ids")
			}
			if dev.Session == nil || dev.Session.SessionKeyID != uplink.SessionKeyID {
				if dev.Session != nil {
					logger.Warn("Session mismatch; restoring session...")
				}
				appSKey, err := as.getAppSKey(ctx, uplink.SessionKeyID, ids)
				if err != nil {
					return nil, nil, errGetAppSKey.WithCause(err)
				}
				dev.Session = &ttnpb.Session{
					DevAddr: *ids.DevAddr,
					SessionKeys: ttnpb.SessionKeys{
						SessionKeyID: uplink.SessionKeyID,
						AppSKey:      &appSKey,
					},
					StartedAt: time.Now(),
				}
				mask = append(mask, "session")
				logger.Debug("Session restored")
			} else if dev.Session.AppSKey == nil {
				return nil, nil, errNoAppSKey
			}
			return dev, mask, nil
		},
	)
	if err != nil {
		return err
	}
	appSKey, err := cryptoutil.UnwrapAES128Key(*dev.Session.AppSKey, as.KeyVault)
	if err != nil {
		return err
	}
	frmPayload, err := crypto.DecryptUplink(appSKey, *ids.DevAddr, uplink.FCnt, uplink.FRMPayload)
	if err != nil {
		return err
	}
	uplink.FRMPayload = frmPayload
	var formatter ttnpb.PayloadFormatter
	var parameter string
	if dev.Formatters != nil {
		formatter, parameter = dev.Formatters.UpFormatter, dev.Formatters.UpFormatterParameter
	} else if link.DefaultFormatters != nil {
		formatter, parameter = link.DefaultFormatters.UpFormatter, link.DefaultFormatters.UpFormatterParameter
	}
	if formatter != ttnpb.PayloadFormatter_FORMATTER_NONE {
		if err := as.formatter.Decode(ctx, ids, dev.VersionIDs, uplink, formatter, parameter); err != nil {
			logger.WithError(err).Warn("Payload decoding failed")
		}
	}
	// TODO: Run uplink messages through location solvers async (https://github.com/TheThingsIndustries/lorawan-stack/issues/1221)
	return nil
}

// var errDeviceNotFound = errors.DefineNotFound("device_not_found", "device `{device_uid}` not found")
