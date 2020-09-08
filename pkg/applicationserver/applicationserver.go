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
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/distribution"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io"
	iogrpc "go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/grpc"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/mqtt"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/packages"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/pubsub"
	_ "go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/pubsub/provider/mqtt" // The MQTT integration provider
	_ "go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/pubsub/provider/nats" // The NATS integration provider
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/web"
	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto/cryptoutil"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/interop"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/messageprocessors"
	"go.thethings.network/lorawan-stack/v3/pkg/messageprocessors/cayennelpp"
	"go.thethings.network/lorawan-stack/v3/pkg/messageprocessors/javascript"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmiddleware/hooks"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"google.golang.org/grpc"
)

// ApplicationServer implements the Application Server component.
//
// The Application Server exposes the As, AppAs and AsEndDeviceRegistry services.
type ApplicationServer struct {
	*component.Component
	ctx context.Context

	config *Config

	linkRegistry     LinkRegistry
	deviceRegistry   DeviceRegistry
	formatters       payloadFormatters
	webhooks         web.Webhooks
	webhookTemplates web.TemplateStore
	pubsub           *pubsub.PubSub
	appPackages      packages.Server

	clusterDistributor distribution.Distributor
	localDistributor   distribution.Distributor

	grpc struct {
		asDevices asEndDeviceRegistryServer
		appAs     ttnpb.AppAsServer
	}

	interopClient InteropClient
	interopID     string
}

// Context returns the context of the Application Server.
func (as *ApplicationServer) Context() context.Context {
	return as.ctx
}

var errListenFrontend = errors.DefineFailedPrecondition("listen_frontend", "failed to start frontend listener `{protocol}` on address `{address}`")

// New returns new *ApplicationServer.
func New(c *component.Component, conf *Config) (as *ApplicationServer, err error) {
	ctx := log.NewContextWithField(c.Context(), "namespace", "applicationserver")

	baseConf := c.GetBaseConfig(ctx)

	var interopCl InteropClient
	if !conf.Interop.IsZero() {
		interopConf := conf.Interop.InteropClient
		interopConf.GetFallbackTLSConfig = func(ctx context.Context) (*tls.Config, error) {
			return c.GetTLSClientConfig(ctx)
		}
		interopConf.BlobConfig = baseConf.Blob

		interopCl, err = interop.NewClient(ctx, interopConf)
		if err != nil {
			return nil, err
		}
	}

	as = &ApplicationServer{
		Component:      c,
		ctx:            ctx,
		config:         conf,
		linkRegistry:   conf.Links,
		deviceRegistry: wrapEndDeviceRegistryWithReplacedFields(conf.Devices, replacedEndDeviceFields...),
		formatters: payloadFormatters(map[ttnpb.PayloadFormatter]messageprocessors.PayloadEncodeDecoder{
			ttnpb.PayloadFormatter_FORMATTER_JAVASCRIPT: javascript.New(),
			ttnpb.PayloadFormatter_FORMATTER_CAYENNELPP: cayennelpp.New(),
		}),
		clusterDistributor: distribution.NewPubSubDistributor(ctx, conf.Distribution.Timeout, conf.Distribution.PubSub),
		localDistributor:   distribution.NewLocalDistributor(ctx, conf.Distribution.Timeout),
		interopClient:      interopCl,
		interopID:          conf.Interop.ID,
	}

	as.grpc.asDevices = asEndDeviceRegistryServer{
		AS:       as,
		kekLabel: conf.DeviceKEKLabel,
	}
	as.grpc.appAs = iogrpc.New(as, iogrpc.WithMQTTConfigProvider(as))

	ctx, cancel := context.WithCancel(as.Context())
	defer func() {
		if err != nil {
			cancel()
		}
	}()

	for _, version := range []struct {
		Format mqtt.Format
		Config config.MQTT
	}{
		{
			Format: mqtt.JSON,
			Config: conf.MQTT,
		},
	} {
		for _, endpoint := range []component.Endpoint{
			component.NewTCPEndpoint(version.Config.Listen, "MQTT"),
			component.NewTLSEndpoint(version.Config.ListenTLS, "MQTT"),
		} {
			version := version
			endpoint := endpoint
			if endpoint.Address() == "" {
				continue
			}
			as.RegisterTask(&component.TaskConfig{
				Context: as.Context(),
				ID:      fmt.Sprintf("serve_mqtt/%s", endpoint.Address()),
				Func: func(ctx context.Context) error {
					l, err := as.ListenTCP(endpoint.Address())
					var lis net.Listener
					if err == nil {
						lis, err = endpoint.Listen(l)
					}
					if err != nil {
						return errListenFrontend.WithCause(err).WithAttributes(
							"address", endpoint.Address(),
							"protocol", endpoint.Protocol(),
						)
					}
					defer lis.Close()
					return mqtt.Serve(ctx, as, lis, version.Format, endpoint.Protocol())
				},
				Restart: component.TaskRestartOnFailure,
				Backoff: component.DefaultTaskBackoffConfig,
			})
		}
	}

	if webhooks, err := conf.Webhooks.NewWebhooks(ctx, as); err != nil {
		return nil, err
	} else if webhooks != nil {
		as.webhooks = webhooks
		c.RegisterWeb(webhooks)
	}

	if as.webhookTemplates, err = conf.Webhooks.Templates.NewTemplateStore(); err != nil {
		return nil, err
	}

	if as.pubsub, err = conf.PubSub.NewPubSub(c, as); err != nil {
		return nil, err
	}

	if as.appPackages, err = conf.Packages.NewApplicationPackages(ctx, as); err != nil {
		return nil, err
	} else if as.appPackages != nil {
		c.RegisterGRPC(as.appPackages)
	}

	hooks.RegisterUnaryHook("/ttn.lorawan.v3.NsAs", cluster.HookName, c.ClusterAuthUnaryHook())

	c.RegisterGRPC(as)
	return as, nil
}

// RegisterServices registers services provided by as at s.
func (as *ApplicationServer) RegisterServices(s *grpc.Server) {
	ttnpb.RegisterAsServer(s, as)
	ttnpb.RegisterNsAsServer(s, as)
	ttnpb.RegisterAsEndDeviceRegistryServer(s, as.grpc.asDevices)
	ttnpb.RegisterAppAsServer(s, as.grpc.appAs)
	if as.webhooks != nil {
		ttnpb.RegisterApplicationWebhookRegistryServer(s, web.NewWebhookRegistryRPC(as.webhooks.Registry(), as.webhookTemplates))
	}
	if as.pubsub != nil {
		ttnpb.RegisterApplicationPubSubRegistryServer(s, as.pubsub)
	}
}

// RegisterHandlers registers gRPC handlers.
func (as *ApplicationServer) RegisterHandlers(s *runtime.ServeMux, conn *grpc.ClientConn) {
	ttnpb.RegisterAsHandler(as.Context(), s, conn)
	ttnpb.RegisterAsEndDeviceRegistryHandler(as.Context(), s, conn)
	ttnpb.RegisterAppAsHandler(as.Context(), s, conn)
	if as.webhooks != nil {
		ttnpb.RegisterApplicationWebhookRegistryHandler(as.Context(), s, conn)
	}
	if as.pubsub != nil {
		ttnpb.RegisterApplicationPubSubRegistryHandler(as.Context(), s, conn)
	}
}

// Roles returns the roles that the Application Server fulfills.
func (as *ApplicationServer) Roles() []ttnpb.ClusterRole {
	return []ttnpb.ClusterRole{ttnpb.ClusterRole_APPLICATION_SERVER}
}

// Subscribe subscribes an application or integration by its identifiers to the Application Server, and returns a
// Subscription for traffic and control. If the cluster parameter is true, the subscription receives all of the
// traffic of the application. Otherwise, only traffic that was processed locally is sent.
func (as *ApplicationServer) Subscribe(ctx context.Context, protocol string, ids *ttnpb.ApplicationIdentifiers, cluster bool) (*io.Subscription, error) {
	ctx = events.ContextWithCorrelationID(ctx, fmt.Sprintf("as:conn:%s", events.NewCorrelationID()))
	if ids != nil {
		uid := unique.ID(ctx, ids)
		ctx = log.NewContextWithField(ctx, "application_uid", uid)
	}
	if cluster {
		return as.clusterDistributor.Subscribe(ctx, protocol, ids)
	} else {
		return as.localDistributor.Subscribe(ctx, protocol, ids)
	}
}

// Publish processes the given upstream message and then publishes it to the application frontends.
func (as *ApplicationServer) Publish(ctx context.Context, up *ttnpb.ApplicationUp) error {
	link, err := as.linkRegistry.Get(ctx, up.ApplicationIdentifiers, []string{
		"default_formatters",
		"skip_payload_crypto",
	})
	if err != nil {
		return err
	}
	return as.processUp(ctx, up, link)
}

func (as *ApplicationServer) processUp(ctx context.Context, up *ttnpb.ApplicationUp, link *ttnpb.ApplicationLink) error {
	ctx = events.ContextWithCorrelationID(ctx, append(up.CorrelationIDs, fmt.Sprintf("as:up:%s", events.NewCorrelationID()))...)
	up.CorrelationIDs = events.CorrelationIDsFromContext(ctx)
	registerReceiveUp(ctx, up)

	now := time.Now().UTC()
	up.ReceivedAt = &now

	pass, err := as.handleUp(ctx, up, link)
	if err != nil {
		log.FromContext(ctx).WithError(err).Warn("Failed to process upstream message")
		registerDropUp(ctx, up, err)
		return nil
	}
	if !pass {
		return nil
	}

	if err := as.publishUp(ctx, up); err != nil {
		log.FromContext(ctx).WithError(err).Warn("Failed to broadcast upstream message")
		registerDropUp(ctx, up, err)
		return nil
	}
	registerForwardUp(ctx, up)

	return nil
}

func (as *ApplicationServer) publishUp(ctx context.Context, up *ttnpb.ApplicationUp) error {
	if err := as.localDistributor.Publish(ctx, up); err != nil {
		return err
	}
	return as.clusterDistributor.Publish(ctx, up)
}

func skipPayloadCrypto(link *ttnpb.ApplicationLink, dev *ttnpb.EndDevice) bool {
	if dev.SkipPayloadCryptoOverride != nil {
		return dev.SkipPayloadCryptoOverride.Value
	}
	return link.SkipPayloadCrypto.GetValue()
}

var (
	errDeviceNotFound  = errors.DefineNotFound("device_not_found", "device `{device_uid}` not found")
	errNoDeviceSession = errors.DefineFailedPrecondition("no_device_session", "no device session; check device activation")
)

func (as *ApplicationServer) downlinkQueueOp(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, items []*ttnpb.ApplicationDownlink, op func(ttnpb.AsNsClient, context.Context, *ttnpb.DownlinkQueueRequest, ...grpc.CallOption) (*pbtypes.Empty, error)) error {
	ctx = events.ContextWithCorrelationID(ctx, fmt.Sprintf("as:downlink:%s", events.NewCorrelationID()))
	for _, item := range items {
		item.CorrelationIDs = append(item.CorrelationIDs, events.CorrelationIDsFromContext(ctx)...)
	}
	logger := log.FromContext(ctx)
	for _, item := range items {
		registerReceiveDownlink(ctx, ids, item)
	}
	peer, err := as.GetPeer(ctx, ttnpb.ClusterRole_NETWORK_SERVER, ids)
	if err != nil {
		return err
	}
	pc, err := peer.Conn()
	if err != nil {
		return err
	}
	link, err := as.linkRegistry.Get(ctx, ids.ApplicationIdentifiers, []string{
		"default_formatters",
		"skip_payload_crypto",
	})
	if err != nil {
		return err
	}
	_, err = as.deviceRegistry.Set(ctx, ids,
		[]string{
			"formatters",
			"pending_session",
			"session",
			"skip_payload_crypto_override",
			"version_ids",
		},
		func(dev *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
			var mask []string
			if dev == nil {
				return nil, nil, errDeviceNotFound.WithAttributes("device_uid", unique.ID(ctx, ids))
			}
			sessions := make([]*ttnpb.Session, 0, 2)
			if dev.Session != nil {
				sessions = append(sessions, dev.Session)
				mask = append(mask, "session.last_a_f_cnt_down")
			}
			if dev.PendingSession != nil {
				// Downlink can be encrypted with the pending session while the device first joined but not confirmed the
				// session by sending an uplink.
				sessions = append(sessions, dev.PendingSession)
				mask = append(mask, "pending_session.last_a_f_cnt_down")
			}
			if len(sessions) == 0 {
				return nil, nil, errNoDeviceSession.New()
			}
			var encryptedItems []*ttnpb.ApplicationDownlink
			for _, session := range sessions {
				for _, item := range items {
					fCnt := session.LastAFCntDown + 1
					if skipPayloadCrypto(link, dev) {
						fCnt = item.FCnt
					}
					encryptedItem := &ttnpb.ApplicationDownlink{
						SessionKeyID:   session.SessionKeyID,
						FPort:          item.FPort,
						FCnt:           fCnt,
						FRMPayload:     item.FRMPayload,
						DecodedPayload: item.DecodedPayload,
						Confirmed:      item.Confirmed,
						ClassBC:        item.ClassBC,
						Priority:       item.Priority,
						CorrelationIDs: item.CorrelationIDs,
					}
					if !skipPayloadCrypto(link, dev) {
						if err := as.encodeAndEncryptDownlink(ctx, dev, session, encryptedItem, link.DefaultFormatters); err != nil {
							logger.WithError(err).Warn("Encoding and encryption of downlink message failed; drop item")
							return nil, nil, err
						}
					}
					encryptedItem.DecodedPayload = nil
					session.LastAFCntDown = encryptedItem.FCnt
					encryptedItems = append(encryptedItems, encryptedItem)
				}
			}
			client := ttnpb.NewAsNsClient(pc)
			req := &ttnpb.DownlinkQueueRequest{
				EndDeviceIdentifiers: ids,
				Downlinks:            encryptedItems,
			}
			_, err = op(client, ctx, req, as.WithClusterAuth())
			if err != nil {
				return nil, nil, err
			}
			return dev, mask, nil
		},
	)
	if err != nil {
		var errorDetails ttnpb.ErrorDetails
		if ttnErr, ok := err.(errors.ErrorDetails); ok {
			errorDetails = *ttnpb.ErrorDetailsToProto(ttnErr)
		}
		for _, item := range items {
			if err := as.publishUp(ctx, &ttnpb.ApplicationUp{
				EndDeviceIdentifiers: ids,
				CorrelationIDs:       item.CorrelationIDs,
				Up: &ttnpb.ApplicationUp_DownlinkFailed{
					DownlinkFailed: &ttnpb.ApplicationDownlinkFailed{
						ApplicationDownlink: *item,
						Error:               errorDetails,
					},
				},
			}); err != nil {
				logger.WithError(err).Warn("Failed to send upstream message")
			}
			registerDropDownlink(ctx, ids, item, err)
		}
		return err
	}
	for _, item := range items {
		if err := as.publishUp(ctx, &ttnpb.ApplicationUp{
			EndDeviceIdentifiers: ids,
			CorrelationIDs:       item.CorrelationIDs,
			Up: &ttnpb.ApplicationUp_DownlinkQueued{
				DownlinkQueued: item,
			},
		}); err != nil {
			logger.WithError(err).Warn("Failed to send upstream message")
		}
		registerForwardDownlink(ctx, ids, item, peer.Name())
	}
	return nil
}

// DownlinkQueuePush pushes the given downlink messages to the end device's application downlink queue.
// This operation changes FRMPayload in the given items.
func (as *ApplicationServer) DownlinkQueuePush(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, items []*ttnpb.ApplicationDownlink) error {
	return as.downlinkQueueOp(ctx, ids, io.CleanDownlinks(items), ttnpb.AsNsClient.DownlinkQueuePush)
}

// DownlinkQueueReplace replaces the end device's application downlink queue with the given downlink messages.
// This operation changes FRMPayload in the given items.
func (as *ApplicationServer) DownlinkQueueReplace(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, items []*ttnpb.ApplicationDownlink) error {
	return as.downlinkQueueOp(ctx, ids, io.CleanDownlinks(items), ttnpb.AsNsClient.DownlinkQueueReplace)
}

var errNoAppSKey = errors.DefineCorruption("no_app_s_key", "no AppSKey")

// DownlinkQueueList lists the application downlink queue of the given end device.
func (as *ApplicationServer) DownlinkQueueList(ctx context.Context, ids ttnpb.EndDeviceIdentifiers) ([]*ttnpb.ApplicationDownlink, error) {
	dev, err := as.deviceRegistry.Get(ctx, ids, []string{
		"formatters",
		"session",
		"skip_payload_crypto",
		"pending_session",
	})
	if err != nil {
		return nil, err
	}
	link, err := as.linkRegistry.Get(ctx, ids.ApplicationIdentifiers, []string{
		"default_formatters",
		"skip_payload_crypto",
	})
	if err != nil {
		return nil, err
	}
	pc, err := as.GetPeerConn(ctx, ttnpb.ClusterRole_NETWORK_SERVER, ids)
	if err != nil {
		return nil, err
	}
	client := ttnpb.NewAsNsClient(pc)
	res, err := client.DownlinkQueueList(ctx, &ids, as.WithClusterAuth())
	if err != nil {
		return nil, err
	}
	if len(res.Downlinks) == 0 {
		return nil, nil
	}
	var queue []*ttnpb.ApplicationDownlink
	var session *ttnpb.Session
	switch {
	case dev.Session != nil:
		session = dev.Session
	case dev.PendingSession != nil:
		session = dev.PendingSession
	default:
		return nil, errNoDeviceSession.New()
	}
	if session.AppSKey == nil {
		return nil, errNoAppSKey.New()
	}
	queue, _ = ttnpb.PartitionDownlinksBySessionKeyIDEquality(session.SessionKeyID, res.Downlinks...)
	if skipPayloadCrypto(link, dev) {
		return queue, nil
	}
	for _, item := range queue {
		if err := as.decryptAndDecodeDownlink(ctx, dev, item, link.DefaultFormatters); err != nil {
			return nil, err
		}
	}
	return queue, nil
}

var errJSUnavailable = errors.DefineUnavailable("join_server_unavailable", "Join Server unavailable for JoinEUI `{join_eui}`")

func (as *ApplicationServer) fetchAppSKey(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, sessionKeyID []byte) (ttnpb.KeyEnvelope, error) {
	req := &ttnpb.SessionKeyRequest{
		SessionKeyID: sessionKeyID,
		DevEUI:       *ids.DevEUI,
	}
	if js, err := as.GetPeer(ctx, ttnpb.ClusterRole_JOIN_SERVER, ids); err == nil {
		cc, err := js.Conn()
		if err != nil {
			return ttnpb.KeyEnvelope{}, err
		}
		res, err := ttnpb.NewAsJsClient(cc).GetAppSKey(ctx, req, as.WithClusterAuth())
		if err == nil {
			return res.AppSKey, nil
		}
		if !errors.IsNotFound(err) {
			return ttnpb.KeyEnvelope{}, err
		}
	}
	if as.interopClient != nil && !interop.GeneratedSessionKeyID(sessionKeyID) {
		res, err := as.interopClient.GetAppSKey(ctx, as.interopID, req)
		if err == nil {
			return res.AppSKey, nil
		}
		if !errors.IsNotFound(err) {
			return ttnpb.KeyEnvelope{}, err
		}
	}
	return ttnpb.KeyEnvelope{}, errJSUnavailable.WithAttributes("join_eui", *ids.JoinEUI)
}

func (as *ApplicationServer) handleUp(ctx context.Context, up *ttnpb.ApplicationUp, link *ttnpb.ApplicationLink) (pass bool, err error) {
	ctx = log.NewContextWithField(ctx, "device_uid", unique.ID(ctx, up.EndDeviceIdentifiers))
	if up.Simulated {
		return true, as.handleSimulatedUp(ctx, up, link)
	}
	switch p := up.Up.(type) {
	case *ttnpb.ApplicationUp_JoinAccept:
		return true, as.handleJoinAccept(ctx, up.EndDeviceIdentifiers, p.JoinAccept, link)
	case *ttnpb.ApplicationUp_UplinkMessage:
		return true, as.handleUplink(ctx, up.EndDeviceIdentifiers, p.UplinkMessage, link)
	case *ttnpb.ApplicationUp_DownlinkQueueInvalidated:
		return as.handleDownlinkQueueInvalidated(ctx, up.EndDeviceIdentifiers, p.DownlinkQueueInvalidated, link)
	case *ttnpb.ApplicationUp_DownlinkQueued:
		return true, as.decryptDownlinkMessage(ctx, up.EndDeviceIdentifiers, p.DownlinkQueued, link)
	case *ttnpb.ApplicationUp_DownlinkSent:
		return true, as.decryptDownlinkMessage(ctx, up.EndDeviceIdentifiers, p.DownlinkSent, link)
	case *ttnpb.ApplicationUp_DownlinkFailed:
		return true, as.decryptDownlinkMessage(ctx, up.EndDeviceIdentifiers, &p.DownlinkFailed.ApplicationDownlink, link)
	case *ttnpb.ApplicationUp_DownlinkAck:
		return true, as.decryptDownlinkMessage(ctx, up.EndDeviceIdentifiers, p.DownlinkAck, link)
	case *ttnpb.ApplicationUp_DownlinkNack:
		return true, as.handleDownlinkNack(ctx, up.EndDeviceIdentifiers, p.DownlinkNack, link)
	case *ttnpb.ApplicationUp_ServiceData:
		return true, nil
	default:
		return false, nil
	}
}

func (as *ApplicationServer) handleSimulatedUp(ctx context.Context, up *ttnpb.ApplicationUp, link *ttnpb.ApplicationLink) error {
	switch p := up.Up.(type) {
	case *ttnpb.ApplicationUp_UplinkMessage:
		return as.handleSimulatedUplink(ctx, up.EndDeviceIdentifiers, p.UplinkMessage, link)
	default:
		return nil
	}
}

var errFetchAppSKey = errors.Define("app_s_key", "failed to get AppSKey")

// handleJoinAccept handles a join-accept message.
// If the application or device is not configured to skip application crypto, the InvalidatedDownlinks and the AppSKey
// in the given join-accept message is reset.
func (as *ApplicationServer) handleJoinAccept(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, joinAccept *ttnpb.ApplicationJoinAccept, link *ttnpb.ApplicationLink) error {
	logger := log.FromContext(ctx).WithFields(log.Fields(
		"join_eui", ids.JoinEUI,
		"dev_eui", ids.DevEUI,
		"session_key_id", joinAccept.SessionKeyID,
	))
	_, err := as.deviceRegistry.Set(ctx, ids,
		[]string{
			"pending_session",
			"session",
			"skip_payload_crypto_override",
		},
		func(dev *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
			var mask []string
			if dev == nil {
				return nil, nil, errDeviceNotFound.WithAttributes("device_uid", unique.ID(ctx, ids))
			}
			if joinAccept.AppSKey != nil {
				logger.Debug("Received AppSKey from Network Server")
			} else {
				logger.Debug("Fetch AppSKey from Join Server")
				key, err := as.fetchAppSKey(ctx, ids, joinAccept.SessionKeyID)
				if err != nil {
					return nil, nil, errFetchAppSKey.WithCause(err)
				}
				joinAccept.AppSKey = &key
				logger.Debug("Fetched AppSKey from Join Server")
			}
			previousSession := dev.PendingSession
			dev.PendingSession = &ttnpb.Session{
				DevAddr: *ids.DevAddr,
				SessionKeys: ttnpb.SessionKeys{
					SessionKeyID: joinAccept.SessionKeyID,
					AppSKey:      joinAccept.AppSKey,
				},
				StartedAt: time.Now().UTC(),
			}
			mask = append(mask, "pending_session")
			if !skipPayloadCrypto(link, dev) {
				if len(joinAccept.InvalidatedDownlinks) > 0 {
					// The Network Server does not reset the downlink queues as the new security session is established,
					// but rather when the session is confirmed on the first uplink. The downlink queue of the current
					// session is passed as part of the join-accept in order to allow the Application Server to compute
					// the downlink queue of this new pending session.
					logger := logger.WithField("count", len(joinAccept.InvalidatedDownlinks))
					logger.Debug("Recalculating downlink queue to restore downlink queue on join")
					if err := as.recalculatePendingDownlinkQueue(ctx, dev, link, previousSession, joinAccept.InvalidatedDownlinks); err != nil {
						logger.WithError(err).Warn("Failed to recalculate downlink queue; items lost")
					}
					joinAccept.InvalidatedDownlinks = nil
				}
				joinAccept.AppSKey = nil
			}
			return dev, mask, nil
		},
	)
	if err != nil {
		return err
	}
	return nil
}

// resetInvalidDownlinkQueue clears the invalid downlink queue of the provided device and publishes the appropriate events.
func (as *ApplicationServer) resetInvalidDownlinkQueue(ctx context.Context, ids ttnpb.EndDeviceIdentifiers) error {
	logger := log.FromContext(ctx)
	pc, err := as.GetPeerConn(ctx, ttnpb.ClusterRole_NETWORK_SERVER, ids)
	if err != nil {
		return err
	}
	client := ttnpb.NewAsNsClient(pc)
	req := &ttnpb.DownlinkQueueRequest{
		EndDeviceIdentifiers: ids,
	}
	_, err = client.DownlinkQueueReplace(ctx, req, as.WithClusterAuth())
	if err != nil {
		logger.WithError(err).Warn("Failed to clear the downlink queue; any queued items in the Network Server are invalid")
		events.Publish(evtInvalidQueueDataDown.NewWithIdentifiersAndData(ctx, ids, err))
	} else {
		events.Publish(evtLostQueueDataDown.NewWithIdentifiersAndData(ctx, ids, err))
	}
	return err
}

// downlinkQueueTransaction represents a transaction to be run on the downlink queues of the provided device.
type downlinkQueueTransaction func(context.Context, *ttnpb.EndDevice) error

// runDownlinkQueueTransaction runs the provided downlink queue transaction on the device. If the transaction
// fails, the LastAFCntDown session fields are restored to their previous values and the downlink queue is reset.
func (as *ApplicationServer) runDownlinkQueueTransaction(ctx context.Context, dev *ttnpb.EndDevice, link *ttnpb.ApplicationLink, t downlinkQueueTransaction) error {
	if skipPayloadCrypto(link, dev) {
		return errPayloadCryptoDisabled.New()
	}
	logger := log.FromContext(ctx)
	pendingSession := dev.PendingSession
	if pendingSession != nil {
		logger = logger.WithField("pending_session_key_id", pendingSession.SessionKeyID)
	}
	session := dev.Session
	if session != nil {
		logger = logger.WithField("session_key_id", session.SessionKeyID)
	}
	ctx = log.NewContext(ctx, logger)
	oldPendingLastAFCntDown := pendingSession.GetLastAFCntDown()
	oldLastAFCntDown := session.GetLastAFCntDown()
	if err := t(ctx, dev); err != nil {
		// If something fails, clear the downlink queue as an empty downlink queue is better than a downlink queue
		// with items that are encrypted with the wrong AppSKey.
		if pendingSession != nil {
			pendingSession.LastAFCntDown = oldPendingLastAFCntDown
		}
		if session != nil {
			session.LastAFCntDown = oldLastAFCntDown
		}
		logger.WithError(err).Warn("Failed to recalculate downlink queue; clear the downlink queue")
		as.resetInvalidDownlinkQueue(ctx, dev.EndDeviceIdentifiers)
	}
	return t(ctx, dev)
}

var errUnknownSession = errors.DefineNotFound("unknown_session", "unknown session")

// recalculatePendingDownlinkQueue computes the downlink queue of the device's pending session, by decrypting the
// invalidated queue using the device session and moving them to the pending session.
// The previous pending session is used for cases in which the device has not been activated, and as such does not
// have a device session. In such cases, if there are downlinks present in the pending queue, they are decrypted
// using the previous pending session and encrypted using new pending session.
// This method mutates the LastAFCntDown of pending session. Downlinks which cannot be decrypted are dropped.
// This method uses the downlink queue transaction mechanism, so any errors that occur during recomputation will
// result in an downlink queue reset attempt.
func (as *ApplicationServer) recalculatePendingDownlinkQueue(ctx context.Context, dev *ttnpb.EndDevice, link *ttnpb.ApplicationLink, previousPendingSession *ttnpb.Session, invalidatedDownlinks []*ttnpb.ApplicationDownlink) error {
	return as.runDownlinkQueueTransaction(ctx, dev, link, func(ctx context.Context, dev *ttnpb.EndDevice) error {
		var previousPendingQueue, previousQueue []*ttnpb.ApplicationDownlink
		unmatched := invalidatedDownlinks
		if previousPendingSession != nil {
			previousPendingQueue, unmatched = ttnpb.PartitionDownlinksBySessionKeyIDEquality(previousPendingSession.SessionKeyID, unmatched...)
		}
		if dev.Session != nil {
			previousQueue, unmatched = ttnpb.PartitionDownlinksBySessionKeyIDEquality(dev.Session.SessionKeyID, unmatched...)
		}
		logger := log.FromContext(ctx)
		for _, item := range unmatched {
			logger.WithFields(log.Fields(
				"f_port", item.FPort,
				"f_cnt", item.FCnt,
				"session_key_id", item.SessionKeyID,
			)).Warn("Downlink message with unknown session key ID found; drop item")
			registerDropDownlink(ctx, dev.EndDeviceIdentifiers, item, errUnknownSession)
		}

		var sourceQueue []*ttnpb.ApplicationDownlink
		var sourceSession *ttnpb.Session
		switch {
		case previousPendingSession == nil && dev.Session == nil:
			// We cannot decode any of the downlinks if both sessions are missing.
			return errNoDeviceSession.New()
		case dev.Session != nil:
			// In the presence of a current session we copy the items from the current session to the pending session.
			sourceQueue = previousQueue
			sourceSession = dev.Session
			logger.Debug("Initializing pending downlink queue from the current session")
		case previousPendingSession != nil:
			// In the presence of a previous session and the absence of a current session we migrate the items from the
			// previous pending session to the new pending session.
			sourceQueue = previousPendingQueue
			sourceSession = previousPendingSession
			logger.Debug("Initializing pending downlink queue from the previous pending session")
		}

		newPendingQueue, err := as.migrateDownlinkQueue(ctx, dev.EndDeviceIdentifiers, sourceQueue, sourceSession, dev.PendingSession, 1)
		if err != nil {
			return err
		}

		pc, err := as.GetPeerConn(ctx, ttnpb.ClusterRole_NETWORK_SERVER, dev.ApplicationIdentifiers)
		if err != nil {
			return err
		}
		client := ttnpb.NewAsNsClient(pc)
		_, err = client.DownlinkQueueReplace(ctx, &ttnpb.DownlinkQueueRequest{
			EndDeviceIdentifiers: dev.EndDeviceIdentifiers,
			Downlinks:            append(previousQueue, newPendingQueue...),
		}, as.WithClusterAuth())
		return err
	})
}

// recalculateDownlinkQueue computes downlink queue of the end device's current session by decrypting the provided
// downlinks using the previous session, and then re-encrypting them using current session starting from the
// provided AFCntDown.
// The previous session is provided for cases in which the downlink queue has to be migrated to a new device
// session. This can occur on device re-activation, when a device sends an uplink with a previously used session key
// ID for which the session key is available.
// This method mutates the LastAFCntDown of end device's session. Downlinks which cannot be decrypted are dropped.
// The pending downlink queue of the end device is discarded.
// This method uses the downlink queue transaction mechanism, so any errors that occur during recomputation will
// result in an downlink queue reset attempt.
func (as *ApplicationServer) recalculateDownlinkQueue(ctx context.Context, dev *ttnpb.EndDevice, link *ttnpb.ApplicationLink, previousSession *ttnpb.Session, previousDownlinks []*ttnpb.ApplicationDownlink, nextAFCntDown uint32, skipEmptyReplace bool) error {
	return as.runDownlinkQueueTransaction(ctx, dev, link, func(ctx context.Context, dev *ttnpb.EndDevice) (err error) {
		downlinks, unmatched := ttnpb.PartitionDownlinksBySessionKeyIDEquality(previousSession.SessionKeyID, previousDownlinks...)
		for _, item := range unmatched {
			log.FromContext(ctx).WithFields(log.Fields(
				"f_port", item.FPort,
				"f_cnt", item.FCnt,
				"session_key_id", item.SessionKeyID,
			)).Warn("Downlink message with unknown session key ID found; drop item")
			registerDropDownlink(ctx, dev.EndDeviceIdentifiers, item, errUnknownSession)
		}
		var newQueue []*ttnpb.ApplicationDownlink
		newQueue, err = as.migrateDownlinkQueue(ctx, dev.EndDeviceIdentifiers, downlinks, previousSession, dev.Session, nextAFCntDown)
		if err != nil {
			return err
		}

		if skipEmptyReplace && len(previousDownlinks) == 0 {
			return nil
		}

		pc, err := as.GetPeerConn(ctx, ttnpb.ClusterRole_NETWORK_SERVER, dev.ApplicationIdentifiers)
		if err != nil {
			return err
		}
		client := ttnpb.NewAsNsClient(pc)
		req := &ttnpb.DownlinkQueueRequest{
			EndDeviceIdentifiers: dev.EndDeviceIdentifiers,
			Downlinks:            newQueue,
		}
		_, err = client.DownlinkQueueReplace(ctx, req, as.WithClusterAuth())
		return err
	})
}

// migrateDownlinkQueue constructs a new downlink queue by decrypting the items of the old queue using the
// old session and encrypting them using the new session.
// This method mutates the LastAFCntDown of the new session. Downlinks which cannot be decrypted are dropped.
// This method does not change the contents of the old downlink queue.
func (as *ApplicationServer) migrateDownlinkQueue(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, oldQueue []*ttnpb.ApplicationDownlink, oldSession *ttnpb.Session, newSession *ttnpb.Session, nextAFCntDown uint32) ([]*ttnpb.ApplicationDownlink, error) {
	if oldSession.AppSKey == nil || newSession.AppSKey == nil {
		return nil, errNoAppSKey.New()
	}
	oldAppSKey, err := cryptoutil.UnwrapAES128Key(ctx, *oldSession.AppSKey, as.KeyVault)
	if err != nil {
		return nil, err
	}
	newAppSKey, err := cryptoutil.UnwrapAES128Key(ctx, *newSession.AppSKey, as.KeyVault)
	if err != nil {
		return nil, err
	}
	newSession.LastAFCntDown = nextAFCntDown - 1
	logger := log.FromContext(ctx)
	newQueue := make([]*ttnpb.ApplicationDownlink, 0, len(oldQueue))
	for _, oldItem := range oldQueue {
		logger := logger.WithFields(log.Fields(
			"f_port", oldItem.FPort,
			"f_cnt", oldItem.FCnt,
			"session_key_id", oldItem.SessionKeyID,
		))
		frmPayload, err := crypto.DecryptDownlink(oldAppSKey, oldSession.DevAddr, oldItem.FCnt, oldItem.FRMPayload, false)
		if err != nil {
			logger.WithError(err).Warn("Failed to decrypt downlink message; drop item")
			registerDropDownlink(ctx, ids, oldItem, err)
			continue
		}
		newFRMPayload, err := crypto.EncryptDownlink(newAppSKey, newSession.DevAddr, newSession.LastAFCntDown+1, frmPayload, false)
		if err != nil {
			logger.WithError(err).Warn("Failed to encrypt downlink message; drop item")
			registerDropDownlink(ctx, ids, oldItem, err)
			continue
		}
		newItem := &ttnpb.ApplicationDownlink{
			SessionKeyID:   newSession.SessionKeyID,
			FPort:          oldItem.FPort,
			FCnt:           newSession.LastAFCntDown + 1,
			FRMPayload:     newFRMPayload,
			Confirmed:      oldItem.Confirmed,
			ClassBC:        oldItem.ClassBC,
			Priority:       oldItem.Priority,
			CorrelationIDs: oldItem.CorrelationIDs,
		}
		newQueue = append(newQueue, newItem)
		newSession.LastAFCntDown = newItem.FCnt
	}
	return newQueue, nil
}

func (as *ApplicationServer) handleUplink(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, uplink *ttnpb.ApplicationUplink, link *ttnpb.ApplicationLink) error {
	ctx = log.NewContextWithField(ctx, "session_key_id", uplink.SessionKeyID)
	logger := log.FromContext(ctx)
	dev, err := as.deviceRegistry.Set(ctx, ids,
		[]string{
			"formatters",
			"pending_session",
			"session",
			"skip_payload_crypto_override",
			"version_ids",
		},
		func(dev *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
			var mask []string
			if dev == nil {
				return nil, nil, errDeviceNotFound.WithAttributes("device_uid", unique.ID(ctx, ids))
			}
		matchSession:
			switch {
			case dev.Session != nil && bytes.Equal(dev.Session.SessionKeyID, uplink.SessionKeyID):
			case dev.PendingSession != nil && bytes.Equal(dev.PendingSession.SessionKeyID, uplink.SessionKeyID):
				dev.Session = dev.PendingSession
				dev.PendingSession = nil
				mask = append(mask, "session", "pending_session")
				logger.Debug("Switched to pending session")
			default:
				appSKey, err := as.fetchAppSKey(ctx, ids, uplink.SessionKeyID)
				if err != nil {
					return nil, nil, errFetchAppSKey.WithCause(err)
				}
				previousSession := dev.Session
				dev.Session = &ttnpb.Session{
					DevAddr: *ids.DevAddr,
					SessionKeys: ttnpb.SessionKeys{
						SessionKeyID: uplink.SessionKeyID,
						AppSKey:      &appSKey,
					},
					StartedAt: time.Now().UTC(),
				}
				previousPendingSession := dev.PendingSession
				dev.PendingSession = nil
				dev.DevAddr = ids.DevAddr
				mask = append(mask, "session", "pending_session", "ids.dev_addr")
				logger.Debug("Restored session")

				switch {
				case previousSession != nil:
				case previousPendingSession != nil:
					previousSession = previousPendingSession
				default:
					break matchSession
				}

				// At this point, the application downlink queue in the Network Server is invalid; recalculation is necessary.
				// Next AFCntDown 1 is assumed. If this is a LoRaWAN 1.0.x end device and the Network Server sent MAC layer
				// downlink already, the Network Server will trigger the DownlinkQueueInvalidated event. Therefore, this
				// recalculation may result in another recalculation.
				pc, err := as.GetPeerConn(ctx, ttnpb.ClusterRole_NETWORK_SERVER, ids)
				if err != nil {
					return nil, nil, err
				}
				client := ttnpb.NewAsNsClient(pc)
				res, err := client.DownlinkQueueList(ctx, &ids, as.WithClusterAuth())
				if err != nil {
					logger.WithError(err).Warn("Failed to list downlink queue for recalculation; clear the downlink queue")
					as.resetInvalidDownlinkQueue(ctx, ids)
				} else {
					previousQueue, unmatched := ttnpb.PartitionDownlinksBySessionKeyIDEquality(previousSession.SessionKeyID, res.Downlinks...)
					if err := as.recalculateDownlinkQueue(ctx, dev, link, previousSession, previousQueue, 1, len(unmatched) == 0); err != nil {
						logger.WithError(err).Warn("Failed to recalculate downlink queue; items lost")
					}
				}
			}
			if dev.Session.AppSKey == nil {
				return nil, nil, errNoAppSKey.New()
			}
			return dev, mask, nil
		},
	)
	if err != nil {
		return err
	}
	if !skipPayloadCrypto(link, dev) {
		if err := as.decryptAndDecodeUplink(ctx, dev, uplink, link.DefaultFormatters); err != nil {
			return err
		}
	} else if dev.Session != nil && dev.Session.AppSKey != nil {
		uplink.AppSKey = dev.Session.AppSKey
		uplink.LastAFCntDown = dev.Session.LastAFCntDown
	}
	// TODO: Run uplink messages through location solvers async (https://github.com/TheThingsNetwork/lorawan-stack/issues/37)
	return nil
}

func (as *ApplicationServer) handleSimulatedUplink(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, uplink *ttnpb.ApplicationUplink, link *ttnpb.ApplicationLink) error {
	ctx = log.NewContextWithField(ctx, "session_key_id", uplink.SessionKeyID)
	dev, err := as.deviceRegistry.Get(ctx, ids,
		[]string{
			"formatters",
			"version_ids",
		},
	)
	if err != nil {
		return err
	}
	return as.decodeUplink(ctx, dev, uplink, link.DefaultFormatters)
}

func (as *ApplicationServer) handleDownlinkQueueInvalidated(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, invalid *ttnpb.ApplicationInvalidatedDownlinks, link *ttnpb.ApplicationLink) (pass bool, err error) {
	_, err = as.deviceRegistry.Set(ctx, ids,
		[]string{
			"session",
			"skip_payload_crypto_override",
		},
		func(dev *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
			if dev == nil {
				return nil, nil, errDeviceNotFound.WithAttributes("device_uid", unique.ID(ctx, ids))
			}
			if dev.Session == nil {
				return nil, nil, errNoDeviceSession.WithAttributes("device_uid", unique.ID(ctx, ids))
			}
			if skipPayloadCrypto(link, dev) {
				// When skipping application payload crypto, the upstream application is responsible for recalculating the
				// downlink queue. No error is returned here to pass the downlink queue invalidation message upstream.
				pass = true
				return dev, nil, nil
			}
			if err := as.recalculateDownlinkQueue(ctx, dev, link, dev.Session, invalid.Downlinks, invalid.LastFCntDown+1, true); err != nil {
				return nil, nil, err
			}
			return dev, []string{"session"}, nil
		},
	)
	return
}

func (as *ApplicationServer) handleDownlinkNack(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, msg *ttnpb.ApplicationDownlink, link *ttnpb.ApplicationLink) error {
	logger := log.FromContext(ctx)
	pc, err := as.GetPeerConn(ctx, ttnpb.ClusterRole_NETWORK_SERVER, ids)
	if err != nil {
		return err
	}
	client := ttnpb.NewAsNsClient(pc)
	res, err := client.DownlinkQueueList(ctx, &ids, as.WithClusterAuth())
	if err != nil {
		logger.WithError(err).Warn("Failed to list the downlink queue for inserting nacked downlink message")
		registerDropDownlink(ctx, ids, msg, err)
	} else {
		_, err := as.deviceRegistry.Set(ctx, ids,
			[]string{
				"session",
				"skip_payload_crypto_override",
			},
			func(dev *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
				if dev == nil {
					return nil, nil, errDeviceNotFound.WithAttributes("device_uid", unique.ID(ctx, ids))
				}
				if dev.Session == nil {
					return nil, nil, errNoDeviceSession.WithAttributes("device_uid", unique.ID(ctx, ids))
				}
				queue, _ := ttnpb.PartitionDownlinksBySessionKeyIDEquality(dev.Session.SessionKeyID, res.Downlinks...)
				queue = append([]*ttnpb.ApplicationDownlink{msg}, queue...)
				if err := as.recalculateDownlinkQueue(ctx, dev, link, dev.Session, queue, msg.FCnt+1, false); err != nil {
					return nil, nil, err
				}
				return dev, []string{"session"}, nil
			},
		)
		if err != nil {
			logger.WithError(err).Warn("Failed to recalculate downlink queue with inserted nacked downlink message")
			registerDropDownlink(ctx, ids, msg, err)
		}
	}
	// Decrypt the message as it will be sent to upstream after handling it.
	if err := as.decryptDownlinkMessage(ctx, ids, msg, link); err != nil {
		return err
	}
	return nil
}

// decryptDownlinkMessage decrypts the downlink message.
// If application payload crypto is skipped, this method returns nil.
func (as *ApplicationServer) decryptDownlinkMessage(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, msg *ttnpb.ApplicationDownlink, link *ttnpb.ApplicationLink) error {
	dev, err := as.deviceRegistry.Get(ctx, ids, []string{
		"formatters",
		"pending_session",
		"session",
		"skip_payload_crypto_override",
	})
	if err != nil {
		return err
	}
	if skipPayloadCrypto(link, dev) {
		return nil
	}
	return as.decryptAndDecodeDownlink(ctx, dev, msg, link.DefaultFormatters)
}

var errPayloadCryptoDisabled = errors.DefineAborted("payload_crypto_disabled", "payload crypto is disabled")

type ctxConfigKeyType struct{}

// GetConfig returns the Application Server config based on the context.
func (as *ApplicationServer) GetConfig(ctx context.Context) (*Config, error) {
	if val, ok := ctx.Value(&ctxConfigKeyType{}).(*Config); ok {
		return val, nil
	}
	return as.config, nil
}

// GetMQTTConfig returns the MQTT frontend configuration based on the context.
func (as *ApplicationServer) GetMQTTConfig(ctx context.Context) (*config.MQTT, error) {
	config, err := as.GetConfig(ctx)
	if err != nil {
		return nil, err
	}
	return &config.MQTT, nil
}
