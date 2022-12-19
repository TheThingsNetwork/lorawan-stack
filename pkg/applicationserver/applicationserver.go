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
	"fmt"
	"net"
	"runtime/trace"
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
	ioweb "go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/web"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/lastseen"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/metadata"
	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto/cryptoutil"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/interop"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/messageprocessors"
	"go.thethings.network/lorawan-stack/v3/pkg/messageprocessors/cayennelpp"
	"go.thethings.network/lorawan-stack/v3/pkg/messageprocessors/devicerepository"
	"go.thethings.network/lorawan-stack/v3/pkg/messageprocessors/javascript"
	"go.thethings.network/lorawan-stack/v3/pkg/task"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"go.thethings.network/lorawan-stack/v3/pkg/web"
	"go.thethings.network/lorawan-stack/v3/pkg/workerpool"
	"google.golang.org/grpc"
)

// ApplicationServer implements the Application Server component.
//
// The Application Server exposes the As, AppAs and AsEndDeviceRegistry services.
type ApplicationServer struct {
	ttnpb.UnimplementedAsServer
	ttnpb.UnimplementedNsAsServer

	*component.Component
	ctx context.Context

	config *Config

	linkRegistry           LinkRegistry
	deviceRegistry         DeviceRegistry
	appUpsRegistry         ApplicationUplinkRegistry
	locationRegistry       metadata.EndDeviceLocationRegistry
	formatters             messageprocessors.MapPayloadProcessor
	webhooks               ioweb.Webhooks
	webhookTemplates       ioweb.TemplateStore
	pubsub                 *pubsub.PubSub
	appPackages            packages.Server
	deviceLastSeenProvider lastseen.LastSeenProvider

	clusterDistributor distribution.Distributor
	localDistributor   distribution.Distributor

	grpc struct {
		asDevices asEndDeviceRegistryServer
		appAs     ttnpb.AppAsServer
	}

	interopClient InteropClient
	interopID     string

	activationPool     workerpool.WorkerPool[*ttnpb.EndDeviceIdentifiers]
	processingPool     workerpool.WorkerPool[*ttnpb.ApplicationUp]
	deviceLastSeenPool workerpool.WorkerPool[lastSeenAtInfo]
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
		interopConf.BlobConfig = baseConf.Blob

		interopCl, err = interop.NewClient(ctx, interopConf, c, interop.SelectorApplicationServer)
		if err != nil {
			return nil, err
		}
	}

	as = &ApplicationServer{
		Component:        c,
		ctx:              ctx,
		config:           conf,
		linkRegistry:     conf.Links,
		deviceRegistry:   wrapEndDeviceRegistryWithReplacedFields(conf.Devices, replacedEndDeviceFields...),
		appUpsRegistry:   conf.UplinkStorage.Registry,
		locationRegistry: conf.EndDeviceMetadataStorage.Location.Registry,
		formatters:       make(messageprocessors.MapPayloadProcessor),
		clusterDistributor: distribution.NewPubSubDistributor(
			ctx,
			c,
			conf.Distribution.Timeout,
			conf.Distribution.Global.PubSub,
			conf.Distribution.Global.Individual.SubscriptionOptions(),
		),
		localDistributor: distribution.NewLocalDistributor(
			ctx,
			c,
			conf.Distribution.Timeout,
			conf.Distribution.Local.Broadcast.SubscriptionOptions(),
			conf.Distribution.Local.Individual.SubscriptionOptions(),
		),
		interopClient: interopCl,
		interopID:     conf.Interop.ID,
	}

	as.formatters[ttnpb.PayloadFormatter_FORMATTER_JAVASCRIPT] = javascript.New()
	as.formatters[ttnpb.PayloadFormatter_FORMATTER_CAYENNELPP] = cayennelpp.New()
	as.formatters[ttnpb.PayloadFormatter_FORMATTER_REPOSITORY] = devicerepository.New(as.formatters, as)

	as.activationPool = workerpool.NewWorkerPool(workerpool.Config[*ttnpb.EndDeviceIdentifiers]{
		Component: c,
		Context:   ctx,
		Name:      "set_activated",
		Handler:   as.setActivated,
	})
	as.processingPool = workerpool.NewWorkerPool(workerpool.Config[*ttnpb.ApplicationUp]{
		Component: c,
		Context:   ctx,
		Name:      "process_application_uplinks",
		Handler:   as.processUpAsync,
	})
	as.deviceLastSeenPool = workerpool.NewWorkerPool(workerpool.Config[lastSeenAtInfo]{
		Component: c,
		Context:   ctx,
		Name:      "store_device_last_seen_from_uplink",
		Handler:   as.storeDeviceLastSeen,
	})

	as.grpc.asDevices = asEndDeviceRegistryServer{
		AS:       as,
		kekLabel: conf.DeviceKEKLabel,
	}
	as.grpc.appAs = iogrpc.New(as,
		iogrpc.WithMQTTConfigProvider(as),
		iogrpc.WithGetEndDeviceIdentifiers(func(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers) (*ttnpb.EndDeviceIdentifiers, error) {
			dev, err := as.deviceRegistry.Get(ctx, ids, []string{"ids"})
			if err != nil {
				return nil, err
			}
			return dev.Ids, nil
		}),
		iogrpc.WithPayloadProcessor(as.formatters),
		iogrpc.WithSkipPayloadCrypto(func(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers) (bool, error) {
			link, err := as.getLink(ctx, ids.ApplicationIds, []string{"skip_payload_crypto"})
			if err != nil {
				return false, err
			}
			dev, err := as.deviceRegistry.Get(ctx, ids, []string{"skip_payload_crypto_override"})
			if err != nil {
				return false, err
			}
			return as.skipPayloadCrypto(ctx, link, dev, nil), nil
		}),
	)

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
			as.RegisterTask(&task.Config{
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
				Restart: task.RestartOnFailure,
				Backoff: task.DefaultBackoffConfig,
			})
		}
	}

	if as.webhooks, err = conf.Webhooks.NewWebhooks(ctx, as); err != nil {
		return nil, err
	}

	if as.webhookTemplates, err = conf.Webhooks.Templates.NewTemplateStore(ctx, as); err != nil {
		return nil, err
	}

	if as.pubsub, err = conf.PubSub.NewPubSub(c, as); err != nil {
		return nil, err
	}

	if as.appPackages, err = conf.Packages.NewApplicationPackages(ctx, as); err != nil {
		return nil, err
	}

	if as.deviceLastSeenProvider, err = conf.DeviceLastSeen.NewLastSeen(ctx, c); err != nil {
		return nil, err
	}

	c.GRPC.RegisterUnaryHook("/ttn.lorawan.v3.NsAs", cluster.HookName, c.ClusterAuthUnaryHook())

	c.RegisterGRPC(as)
	c.RegisterWeb(as)
	return as, nil
}

// RegisterServices registers services provided by as at s.
func (as *ApplicationServer) RegisterServices(s *grpc.Server) {
	ttnpb.RegisterAsServer(s, as)
	ttnpb.RegisterNsAsServer(s, as)
	ttnpb.RegisterAsEndDeviceRegistryServer(s, as.grpc.asDevices)
	ttnpb.RegisterAppAsServer(s, as.grpc.appAs)
	if wh := as.webhooks; wh != nil {
		ttnpb.RegisterApplicationWebhookRegistryServer(s, ioweb.NewWebhookRegistryRPC(wh.Registry(), as.webhookTemplates))
	}
	if ps := as.pubsub; ps != nil {
		ttnpb.RegisterApplicationPubSubRegistryServer(s, ps)
	}
	if pkgs := as.appPackages; pkgs != nil {
		pkgs.RegisterServices(s)
	}
}

// RegisterHandlers registers gRPC handlers.
func (as *ApplicationServer) RegisterHandlers(s *runtime.ServeMux, conn *grpc.ClientConn) {
	ttnpb.RegisterAsHandler(as.Context(), s, conn)                  //nolint:errcheck
	ttnpb.RegisterAsEndDeviceRegistryHandler(as.Context(), s, conn) //nolint:errcheck
	ttnpb.RegisterAppAsHandler(as.Context(), s, conn)               //nolint:errcheck
	if as.webhooks != nil {
		ttnpb.RegisterApplicationWebhookRegistryHandler(as.Context(), s, conn) //nolint:errcheck
	}
	if as.pubsub != nil {
		ttnpb.RegisterApplicationPubSubRegistryHandler(as.Context(), s, conn) //nolint:errcheck
	}
	if pkgs := as.appPackages; pkgs != nil {
		pkgs.RegisterHandlers(s, conn)
	}
}

// RegisterRoutes registers HTTP routes.
func (as *ApplicationServer) RegisterRoutes(s *web.Server) {
	if wh := as.webhooks; wh != nil {
		wh.RegisterRoutes(s)
	}
	if pkgs := as.appPackages; pkgs != nil {
		pkgs.RegisterRoutes(s)
	}
}

// Roles returns the roles that the Application Server fulfills.
func (*ApplicationServer) Roles() []ttnpb.ClusterRole {
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
	}
	return as.localDistributor.Subscribe(ctx, protocol, ids)
}

// Publish processes the given upstream message and then publishes it to the application frontends.
func (as *ApplicationServer) Publish(ctx context.Context, up *ttnpb.ApplicationUp) error {
	return as.processingPool.Publish(ctx, up)
}

func (as *ApplicationServer) processUpAsync(ctx context.Context, up *ttnpb.ApplicationUp) {
	link, err := as.getLink(ctx, up.EndDeviceIds.ApplicationIds, []string{
		"default_formatters",
		"skip_payload_crypto",
	})
	if err != nil {
		log.FromContext(ctx).WithError(err).Warn("Failed to retrieve application link")
		return
	}
	if err := as.processUp(ctx, up, link); err != nil {
		log.FromContext(ctx).WithError(err).Warn("Failed to process application uplink")
		return
	}
}

// lastSeenAtInfo holds the information needed for a worker to store entry in the LastSeen map.
type lastSeenAtInfo struct {
	ids        *ttnpb.EndDeviceIdentifiers
	lastSeenAt *pbtypes.Timestamp
}

func (as *ApplicationServer) storeDeviceLastSeen(ctx context.Context, lastSeenEntry lastSeenAtInfo) {
	if err := as.deviceLastSeenProvider.PushLastSeenFromUplink(ctx, lastSeenEntry.ids, lastSeenEntry.lastSeenAt); err != nil {
		log.FromContext(ctx).WithError(err).Warn("Failed to set device last seen timestamp")
	}
}

func (as *ApplicationServer) processUp(ctx context.Context, up *ttnpb.ApplicationUp, link *ttnpb.ApplicationLink) error {
	defer trace.StartRegion(ctx, "process up").End()

	ctx = log.NewContextWithField(ctx, "device_uid", unique.ID(ctx, up.EndDeviceIds))
	ctx = events.ContextWithCorrelationID(ctx, append(up.CorrelationIds, fmt.Sprintf("as:up:%s", events.NewCorrelationID()))...)
	up.CorrelationIds = events.CorrelationIDsFromContext(ctx)
	registerReceiveUp(ctx, up)

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
	defer trace.StartRegion(ctx, "publish up").End()

	if err := as.localDistributor.Publish(ctx, up); err != nil {
		return err
	}
	return as.clusterDistributor.Publish(ctx, up)
}

// skipPayloadCrypto indicates whether LoRaWAN FRMPayload encryption and decryption should be skipped.
// This method returns true if the AppSKey of the given session is wrapped and cannot be unwrapped by the Application
// Server, and if the end device's skip_payload_crypto_override is true or if the link's skip_payload_crypto is true.
func (as *ApplicationServer) skipPayloadCrypto(ctx context.Context, link *ttnpb.ApplicationLink, dev *ttnpb.EndDevice, session *ttnpb.Session) bool {
	if appSKey := session.GetKeys().GetAppSKey(); appSKey != nil {
		if _, err := cryptoutil.UnwrapAES128Key(ctx, appSKey, as.KeyVault); err == nil {
			return false
		}
	}
	if dev.SkipPayloadCryptoOverride != nil {
		return dev.SkipPayloadCryptoOverride.Value
	}
	return link.SkipPayloadCrypto.GetValue()
}

// lastAFCntDownFromMinFCnt computes the last application frame counter based on the
// minimum frame counter provided by the Network Server.
// The Network Server may report this minimum as being zero, thus the last application
// frame counter would be -1. As the frame counters are unsigned integers, this would
// lead to an overflow.
func lastAFCntDownFromMinFCnt(min uint32) uint32 {
	if min == 0 {
		return 0
	}
	return min - 1
}

var (
	errDeviceNotFound  = errors.DefineNotFound("device_not_found", "device `{device_uid}` not found")
	errNoDeviceSession = errors.DefineFailedPrecondition("no_device_session", "no device session; check device activation")
	errRebuild         = errors.DefineAborted("rebuild", "could not rebuild device session; check device address")
)

// buildSessionsFromError attempts to rebuild the end device session and pending session based on the error
// details found in the provided error. This may mutate the session, pending session and device address.
// If the sessions cannot be rebuilt from the provided error, the error itself is returned.
func (as *ApplicationServer) buildSessionsFromError(ctx context.Context, dev *ttnpb.EndDevice, err error) ([]string, error) {
	defer trace.StartRegion(ctx, "rebuild sessions from error").End()

	reconstructSession := func(sessionKeyID []byte, devAddr *types.DevAddr, minFCntDown uint32) (*ttnpb.Session, error) {
		appSKey, err := as.fetchAppSKey(ctx, dev.Ids, sessionKeyID)
		if err != nil {
			return nil, errFetchAppSKey.WithCause(err)
		}
		return &ttnpb.Session{
			DevAddr: devAddr.Bytes(),
			Keys: &ttnpb.SessionKeys{
				SessionKeyId: sessionKeyID,
				AppSKey:      appSKey,
			},
			LastAFCntDown: lastAFCntDownFromMinFCnt(minFCntDown),
		}, nil
	}

	var ttnErr errors.ErrorDetails
	if !errors.As(err, &ttnErr) {
		return nil, err
	}
	details := ttnErr.Details()
	if len(details) == 0 {
		return nil, err
	}
	var diagnostics *ttnpb.DownlinkQueueOperationErrorDetails
	for _, detail := range details {
		var ok bool
		diagnostics, ok = detail.(*ttnpb.DownlinkQueueOperationErrorDetails)
		if ok {
			break
		}
	}
	if diagnostics == nil {
		return nil, err
	}

	var mask []string
	if diagnostics.DevAddr != nil {
		switch {
		// If the SessionKeyID and DevAddr did not change, just update the LastAFCntDown.
		case dev.Session != nil &&
			bytes.Equal(diagnostics.SessionKeyId, dev.Session.Keys.SessionKeyId) &&
			types.MustDevAddr(dev.Session.DevAddr).OrZero().Equal(types.MustDevAddr(diagnostics.DevAddr).OrZero()):
			dev.Session.LastAFCntDown = lastAFCntDownFromMinFCnt(diagnostics.MinFCntDown)
		// If there is a SessionKeyID on the Network Server side, rebuild the session.
		case len(diagnostics.SessionKeyId) > 0:
			session, err := reconstructSession(diagnostics.SessionKeyId,
				types.MustDevAddr(diagnostics.DevAddr),
				diagnostics.MinFCntDown,
			)
			if err != nil {
				return nil, err
			}
			dev.Session = session
			dev.Ids.DevAddr = session.DevAddr
		default:
			return nil, errRebuild.New()
		}
	} else {
		dev.Session = nil
		dev.Ids.DevAddr = nil
	}
	mask = ttnpb.AddFields(mask, "session", "ids.dev_addr")

	if diagnostics.PendingDevAddr != nil {
		switch {
		// If the SessionKeyID did not change, just update the LastAFcntDown.
		case dev.PendingSession != nil &&
			bytes.Equal(diagnostics.PendingSessionKeyId, dev.PendingSession.Keys.SessionKeyId) &&
			types.MustDevAddr(
				dev.PendingSession.DevAddr).OrZero().Equal(types.MustDevAddr(diagnostics.PendingDevAddr).OrZero()):
			dev.PendingSession.LastAFCntDown = lastAFCntDownFromMinFCnt(diagnostics.PendingMinFCntDown)
		// If there is a SessionKeyID on the Network Server side, rebuild the session.
		case len(diagnostics.PendingSessionKeyId) > 0:
			session, err := reconstructSession(diagnostics.PendingSessionKeyId,
				types.MustDevAddr(diagnostics.PendingDevAddr),
				diagnostics.PendingMinFCntDown,
			)
			if err != nil {
				return nil, err
			}
			dev.PendingSession = session
		default:
			return nil, errRebuild.New()
		}
	} else {
		dev.PendingSession = nil
	}
	mask = ttnpb.AddFields(mask, "pending_session")

	return mask, nil
}

type downlinkQueueOperation struct {
	Items             []*ttnpb.ApplicationDownlink
	Operation         func(ttnpb.AsNsClient, context.Context, *ttnpb.DownlinkQueueRequest, ...grpc.CallOption) (*pbtypes.Empty, error)
	SkipSessionKeyIDs [][]byte
	ResultFunc        func(decrypted, encrypted []*ttnpb.ApplicationDownlink, err error)
}

func (d downlinkQueueOperation) shouldSkip(sessionKeyID []byte) bool {
	for _, id := range d.SkipSessionKeyIDs {
		if bytes.Equal(id, sessionKeyID) {
			return true
		}
	}
	return false
}

const maxDownlinkQueueOperationAttempts = 50

func (as *ApplicationServer) attemptDownlinkQueueOp(ctx context.Context, dev *ttnpb.EndDevice, link *ttnpb.ApplicationLink, peer cluster.Peer, op downlinkQueueOperation) (mask []string, err error) {
	defer trace.StartRegion(ctx, "attempt downlink queue operation").End()

	var encryptedItems []*ttnpb.ApplicationDownlink
	if op.ResultFunc != nil {
		defer func() {
			op.ResultFunc(op.Items, encryptedItems, err)
		}()
	}

	for _, down := range op.Items {
		if down.FPort == 0 {
			return nil, errNoFPort.New()
		}
	}

	pc, err := peer.Conn()
	if err != nil {
		return nil, err
	}
	mask = make([]string, 0, 2)
	attempt := 1
	for {
		ctx := log.NewContextWithField(ctx, "attempt", attempt)

		sessions := make([]*ttnpb.Session, 0, 2)
		if dev.Session != nil && !op.shouldSkip(dev.Session.Keys.SessionKeyId) {
			sessions = append(sessions, dev.Session)
			mask = ttnpb.AddFields(mask, "session.last_a_f_cnt_down")
		}
		if dev.PendingSession != nil && !op.shouldSkip(dev.PendingSession.Keys.SessionKeyId) {
			// Downlink can be encrypted with the pending session while the device first joined but not confirmed the
			// session by sending an uplink.
			sessions = append(sessions, dev.PendingSession)
			mask = ttnpb.AddFields(mask, "pending_session.last_a_f_cnt_down")
		}
		if len(sessions) == 0 {
			return nil, errNoDeviceSession.New()
		}
		encryptedItems, err = as.encryptDownlinks(ctx, dev, link, op.Items, sessions)
		if err != nil {
			return nil, err
		}

		_, err = op.Operation(ttnpb.NewAsNsClient(pc), ctx, &ttnpb.DownlinkQueueRequest{
			EndDeviceIds: dev.Ids,
			Downlinks:    encryptedItems,
		}, as.WithClusterAuth())
		if err == nil {
			return mask, nil
		}
		if attempt >= maxDownlinkQueueOperationAttempts || as.skipPayloadCrypto(ctx, link, dev, nil) {
			return nil, err
		}
		mask, err = as.buildSessionsFromError(ctx, dev, err)
		if err != nil {
			return nil, err
		}
		attempt++
	}
}

func (as *ApplicationServer) downlinkQueueOp(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, items []*ttnpb.ApplicationDownlink, op func(ttnpb.AsNsClient, context.Context, *ttnpb.DownlinkQueueRequest, ...grpc.CallOption) (*pbtypes.Empty, error)) error {
	ctx = events.ContextWithCorrelationID(ctx, fmt.Sprintf("as:downlink:%s", events.NewCorrelationID()))
	link, err := as.getLink(ctx, ids.ApplicationIds, []string{
		"default_formatters",
		"skip_payload_crypto",
	})
	if err != nil {
		return err
	}
	peer, err := as.GetPeer(ctx, ttnpb.ClusterRole_NETWORK_SERVER, nil)
	if err != nil {
		return err
	}
	for _, item := range items {
		ctx := events.ContextWithCorrelationID(ctx, item.CorrelationIds...)
		item.CorrelationIds = events.CorrelationIDsFromContext(ctx)
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
			if dev == nil {
				return nil, nil, errDeviceNotFound.WithAttributes("device_uid", unique.ID(ctx, ids))
			}
			for _, item := range items {
				var err error
				if item.FrmPayload != nil && item.DecodedPayload == nil {
					err = as.decodeDownlink(ctx, dev, item, link.DefaultFormatters)
				} else {
					err = as.encodeDownlink(ctx, dev, item, link.DefaultFormatters)
				}
				if err != nil {
					log.FromContext(ctx).WithError(err).Warn("Encoding or decoding downlink message failed")
					return nil, nil, err
				}
			}
			registerReceiveDownlinks(ctx, ids, items)
			mask, err := as.attemptDownlinkQueueOp(ctx, dev, link, peer, downlinkQueueOperation{
				Items:     items,
				Operation: op,
				ResultFunc: func(decrypted, encrypted []*ttnpb.ApplicationDownlink, err error) {
					if err != nil {
						as.registerDropDownlinks(ctx, ids, decrypted, err)
					} else {
						as.registerForwardDownlinks(ctx, ids, decrypted, encrypted, peer.Name())
					}
				},
			})
			if err != nil {
				return nil, nil, err
			}
			return dev, mask, nil
		},
	)
	return err
}

// DownlinkQueuePush pushes the given downlink messages to the end device's application downlink queue.
// This operation changes FRMPayload in the given items.
func (as *ApplicationServer) DownlinkQueuePush(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, items []*ttnpb.ApplicationDownlink) error {
	return as.downlinkQueueOp(ctx, ids, io.CleanDownlinks(items), ttnpb.AsNsClient.DownlinkQueuePush)
}

// DownlinkQueueReplace replaces the end device's application downlink queue with the given downlink messages.
// This operation changes FRMPayload in the given items.
func (as *ApplicationServer) DownlinkQueueReplace(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, items []*ttnpb.ApplicationDownlink) error {
	return as.downlinkQueueOp(ctx, ids, io.CleanDownlinks(items), ttnpb.AsNsClient.DownlinkQueueReplace)
}

var errNoAppSKey = errors.DefineCorruption("no_app_s_key", "no AppSKey")

// DownlinkQueueList lists the application downlink queue of the given end device.
func (as *ApplicationServer) DownlinkQueueList(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers) ([]*ttnpb.ApplicationDownlink, error) {
	dev, err := as.deviceRegistry.Get(ctx, ids, []string{
		"formatters",
		"session",
		"skip_payload_crypto",
		"pending_session",
	})
	if err != nil {
		return nil, err
	}
	link, err := as.getLink(ctx, ids.ApplicationIds, []string{
		"default_formatters",
		"skip_payload_crypto",
	})
	if err != nil {
		return nil, err
	}
	pc, err := as.GetPeerConn(ctx, ttnpb.ClusterRole_NETWORK_SERVER, nil)
	if err != nil {
		return nil, err
	}
	client := ttnpb.NewAsNsClient(pc)
	res, err := client.DownlinkQueueList(ctx, ids, as.WithClusterAuth())
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
	if session.GetKeys().GetAppSKey() == nil {
		return nil, errNoAppSKey.New()
	}
	queue, _ = ttnpb.PartitionDownlinksBySessionKeyIDEquality(session.Keys.SessionKeyId, res.Downlinks...)
	if as.skipPayloadCrypto(ctx, link, dev, session) {
		return queue, nil
	}
	for _, item := range queue {
		if err := as.decryptAndDecodeDownlink(ctx, dev, item, link.DefaultFormatters); err != nil {
			return nil, err
		}
	}
	return queue, nil
}

var (
	errJSUnavailable = errors.DefineUnavailable("join_server_unavailable", "Join Server unavailable for JoinEUI `{join_eui}`")
	errNoDevEUI      = errors.DefineInvalidArgument("no_dev_eui", "no device EUI provided")
	errNoJoinEUI     = errors.DefineInvalidArgument("no_join_eui", "no join EUI provided")
)

func (as *ApplicationServer) fetchAppSKey(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, sessionKeyID []byte) (*ttnpb.KeyEnvelope, error) {
	defer trace.StartRegion(ctx, "fetch AppSKey").End()

	if ids == nil || ids.DevEui == nil {
		return nil, errNoDevEUI.New()
	}
	if ids.JoinEui == nil {
		return nil, errNoJoinEUI.New()
	}
	req := &ttnpb.SessionKeyRequest{
		SessionKeyId: sessionKeyID,
		DevEui:       ids.DevEui,
		JoinEui:      ids.JoinEui,
	}
	if js, err := as.GetPeer(ctx, ttnpb.ClusterRole_JOIN_SERVER, nil); err == nil {
		cc, err := js.Conn()
		if err != nil {
			return nil, err
		}
		res, err := ttnpb.NewAsJsClient(cc).GetAppSKey(ctx, req, as.WithClusterAuth())
		if err == nil && res.AppSKey != nil {
			return res.AppSKey, nil
		}
		if !errors.IsNotFound(err) {
			return nil, err
		}
	}
	if as.interopClient != nil && !interop.GeneratedSessionKeyID(sessionKeyID) {
		res, err := as.interopClient.GetAppSKey(ctx, as.interopID, req)
		if err == nil && res.AppSKey != nil {
			return res.AppSKey, nil
		}
		if !errors.IsNotFound(err) {
			return nil, err
		}
	}
	return nil, errJSUnavailable.WithAttributes("join_eui", types.MustEUI64(ids.JoinEui).OrZero())
}

func (as *ApplicationServer) handleUp(ctx context.Context, up *ttnpb.ApplicationUp, link *ttnpb.ApplicationLink) (pass bool, err error) {
	if up.Simulated {
		return true, as.handleSimulatedUp(ctx, up, link)
	}
	switch p := up.Up.(type) {
	case *ttnpb.ApplicationUp_JoinAccept:
		return true, as.handleJoinAccept(ctx, up.EndDeviceIds, p.JoinAccept, link)
	case *ttnpb.ApplicationUp_UplinkMessage:
		return true, as.handleUplink(ctx, uplinkInfo{up.EndDeviceIds, up.ReceivedAt, p.UplinkMessage, false, link})
	case *ttnpb.ApplicationUp_UplinkNormalized:
		return true, nil
	case *ttnpb.ApplicationUp_DownlinkQueueInvalidated:
		return as.handleDownlinkQueueInvalidated(ctx, up.EndDeviceIds, p.DownlinkQueueInvalidated, link)
	case *ttnpb.ApplicationUp_DownlinkSent:
		return true, as.decryptDownlinkMessage(ctx, up.EndDeviceIds, p.DownlinkSent, link)
	case *ttnpb.ApplicationUp_DownlinkFailed:
		return true, as.decryptDownlinkMessage(ctx, up.EndDeviceIds, p.DownlinkFailed.Downlink, link)
	case *ttnpb.ApplicationUp_DownlinkAck:
		return true, as.decryptDownlinkMessage(ctx, up.EndDeviceIds, p.DownlinkAck, link)
	case *ttnpb.ApplicationUp_DownlinkNack:
		return true, as.handleDownlinkNack(ctx, up.EndDeviceIds, p.DownlinkNack, link)
	case *ttnpb.ApplicationUp_LocationSolved:
		return true, as.handleLocationSolved(ctx, up.EndDeviceIds, p.LocationSolved, link)
	case *ttnpb.ApplicationUp_ServiceData:
		return true, nil
	default:
		return false, nil
	}
}

func (as *ApplicationServer) handleSimulatedUp(ctx context.Context, up *ttnpb.ApplicationUp, link *ttnpb.ApplicationLink) error {
	switch p := up.Up.(type) {
	case *ttnpb.ApplicationUp_UplinkMessage:
		return as.handleSimulatedUplink(ctx, uplinkInfo{up.EndDeviceIds, up.ReceivedAt, p.UplinkMessage, true, link})
	default:
		return nil
	}
}

var errFetchAppSKey = errors.Define("app_s_key", "failed to get AppSKey")

// handleJoinAccept handles a join-accept message.
// If the application or device is not configured to skip application crypto, the InvalidatedDownlinks and the AppSKey
// in the given join-accept message is reset.
func (as *ApplicationServer) handleJoinAccept(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, joinAccept *ttnpb.ApplicationJoinAccept, link *ttnpb.ApplicationLink) error {
	defer trace.StartRegion(ctx, "handle join accept").End()

	logger := log.FromContext(ctx).WithFields(log.Fields(
		"join_eui", ids.JoinEui,
		"dev_eui", ids.DevEui,
		"session_key_id", joinAccept.SessionKeyId,
	))
	peer, err := as.GetPeer(ctx, ttnpb.ClusterRole_NETWORK_SERVER, nil)
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
			if joinAccept.AppSKey != nil {
				logger.Debug("Received AppSKey from Network Server")
			} else {
				logger.Debug("Fetch AppSKey from Join Server")
				key, err := as.fetchAppSKey(ctx, ids, joinAccept.SessionKeyId)
				if err != nil {
					return nil, nil, errFetchAppSKey.WithCause(err)
				}
				joinAccept.AppSKey = key
				logger.Debug("Fetched AppSKey from Join Server")
			}
			previousSession := dev.PendingSession
			dev.PendingSession = &ttnpb.Session{
				DevAddr: ids.DevAddr,
				Keys: &ttnpb.SessionKeys{
					SessionKeyId: joinAccept.SessionKeyId,
					AppSKey:      joinAccept.AppSKey,
				},
			}
			mask = ttnpb.AddFields(mask, "pending_session")
			if as.skipPayloadCrypto(ctx, link, dev, dev.PendingSession) {
				return dev, mask, nil
			}
			joinAccept.AppSKey = nil
			if len(joinAccept.InvalidatedDownlinks) == 0 {
				return dev, mask, nil
			}

			// The Network Server does not reset the downlink queues as the new security session is established,
			// but rather when the session is confirmed on the first uplink. The downlink queue of the current
			// session is passed as part of the join-accept in order to allow the Application Server to compute
			// the downlink queue of this new pending session.
			logger := logger.WithField("count", len(joinAccept.InvalidatedDownlinks))
			logger.Debug("Recalculating downlink queue to restore downlink queue on join")

			items := make([]*ttnpb.ApplicationDownlink, 0, len(joinAccept.InvalidatedDownlinks))
			for _, msg := range joinAccept.InvalidatedDownlinks {
				if err := as.decryptDownlink(ctx, dev, msg, previousSession); err != nil {
					logger.WithError(err).Warn("Failed to decrypt downlink message; drop item")
					registerDropDownlink(ctx, ids, msg, err)
					continue
				}
				items = append(items, msg)
			}
			joinAccept.InvalidatedDownlinks = nil
			if len(items) == 0 {
				return dev, mask, nil
			}

			pushMask, err := as.attemptDownlinkQueueOp(ctx, dev, link, peer, downlinkQueueOperation{
				Items:     items,
				Operation: ttnpb.AsNsClient.DownlinkQueuePush,
				// The session from which the downlinks originate already contains them. As such
				// we don't need to push them there.
				SkipSessionKeyIDs: [][]byte{items[0].SessionKeyId},
				ResultFunc: func(decrypted, _ []*ttnpb.ApplicationDownlink, err error) {
					if err != nil {
						as.registerDropDownlinks(ctx, ids, decrypted, err)
					}
				},
			})
			if err != nil {
				return nil, nil, err
			}
			mask = ttnpb.AddFields(mask, pushMask...)

			return dev, mask, nil
		},
	)
	if err != nil {
		return err
	}

	// Publish last seen event.
	if err := as.deviceLastSeenPool.Publish(ctx, lastSeenAtInfo{
		ids:        ids,
		lastSeenAt: joinAccept.ReceivedAt,
	}); err != nil {
		logger.WithError(err).Warn("Failed to publish last seen event")
	}

	return nil
}

var errUnknownSession = errors.DefineNotFound("unknown_session", "unknown session")

// matchSession updates the currently active ttnpb.Session of a ttnpb.EndDevice, based on the provided session key ID.
// This function will mutate the provided ttnpb.EndDevice and migrate the Session field to the session that matches
// the provided session key ID.
// The following fields are expected to be part of the provided ttnpb.EndDevice:
//   - session and pending_session, used to decide which session is currently active.
//   - formatters, version_ids, used by the downlink queue encoders, in cases in which the queue must be recalculated.
//   - skip_payload_crypto_override, used by the downlink queue migration mechanism in order to avoid payload encryption.
func (as *ApplicationServer) matchSession(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, dev *ttnpb.EndDevice, link *ttnpb.ApplicationLink, sessionKeyID []byte) ([]string, error) {
	logger := log.FromContext(ctx)
	var mask []string
	switch {
	case dev.Session != nil && bytes.Equal(dev.Session.Keys.SessionKeyId, sessionKeyID):
	case dev.PendingSession != nil && bytes.Equal(dev.PendingSession.Keys.SessionKeyId, sessionKeyID):
		dev.Session = dev.PendingSession
		dev.PendingSession = nil
		mask = ttnpb.AddFields(mask, "session", "pending_session")
		logger.Debug("Switched to pending session")
	default:
		appSKey, err := as.fetchAppSKey(ctx, ids, sessionKeyID)
		if err != nil {
			return nil, errFetchAppSKey.WithCause(err)
		}
		dev.Session = &ttnpb.Session{
			DevAddr: ids.DevAddr,
			Keys: &ttnpb.SessionKeys{
				SessionKeyId: sessionKeyID,
				AppSKey:      appSKey,
			},
		}
		dev.PendingSession = nil
		dev.Ids.DevAddr = ids.DevAddr
		mask = ttnpb.AddFields(mask, "session", "pending_session", "ids.dev_addr")
		logger.Debug("Restored session")
	}
	return mask, nil
}

// storeUplink stores the provided *ttnpb.ApplicationUplink in the device uplink storage.
// Only fields which are used by integrations are stored.
// The fields which are stored are based on the following usages:
// - io/packages/loragls/v3/package.go#multiFrameQuery
// - io/packages/loragls/v3/api/objects.go#parseRxMetadata.
func (as *ApplicationServer) storeUplink(
	ctx context.Context,
	ids *ttnpb.EndDeviceIdentifiers,
	uplink *ttnpb.ApplicationUplink,
) error {
	cleanUplink := &ttnpb.ApplicationUplink{
		RxMetadata: make([]*ttnpb.RxMetadata, 0, len(uplink.RxMetadata)),
		ReceivedAt: uplink.ReceivedAt,
	}
	for _, md := range uplink.RxMetadata {
		if md.GatewayIds == nil {
			continue
		}
		cleanUplink.RxMetadata = append(cleanUplink.RxMetadata, &ttnpb.RxMetadata{
			GatewayIds: &ttnpb.GatewayIdentifiers{
				GatewayId: md.GatewayIds.GatewayId,
			},
			AntennaIndex:  md.AntennaIndex,
			FineTimestamp: md.FineTimestamp,
			Location:      md.Location,
			Rssi:          md.Rssi,
			Snr:           md.Snr,
		})
	}
	return as.appUpsRegistry.Push(ctx, ids, cleanUplink)
}

// setActivated attempts to mark the end device as activated in the Entity Registry.
// If the update succeeds, the end device will be updated in the Application Server end device registry
// in order to avoid subsequent calls.
func (as *ApplicationServer) setActivated(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers) {
	defer trace.StartRegion(ctx, "set activated").End()

	cc, err := as.GetPeerConn(ctx, ttnpb.ClusterRole_ENTITY_REGISTRY, nil)
	if err != nil {
		log.FromContext(ctx).WithError(err).Warn("Failed to get Entity Registry peer")
		return
	}
	now := time.Now().UTC()
	mask := []string{"activated_at"}
	_, err = ttnpb.NewEndDeviceRegistryClient(cc).Update(ctx, &ttnpb.UpdateEndDeviceRequest{
		EndDevice: &ttnpb.EndDevice{
			Ids:         ids,
			ActivatedAt: ttnpb.ProtoTimePtr(now),
		},
		FieldMask: ttnpb.FieldMask(mask...),
	}, as.WithClusterAuth())
	if err != nil {
		log.FromContext(ctx).WithError(err).Warn("Failed to set end device activation status in Entity Registry")
		return
	}
	if _, err = as.deviceRegistry.Set(ctx, ids, mask,
		func(dev *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
			if dev == nil {
				return nil, nil, errDeviceNotFound.WithAttributes("device_uid", unique.ID(ctx, ids))
			}
			dev.ActivatedAt = ttnpb.ProtoTimePtr(now)
			return dev, mask, nil
		},
	); err != nil {
		log.FromContext(ctx).WithError(err).Warn("Failed to set end device activation status in local registry")
		return
	}
}

func (as *ApplicationServer) publishNormalizedUplink(ctx context.Context, info uplinkInfo) error {
	for _, measurement := range info.uplink.NormalizedPayload {
		if err := as.Publish(ctx, &ttnpb.ApplicationUp{
			EndDeviceIds:   info.ids,
			CorrelationIds: events.CorrelationIDsFromContext(ctx),
			ReceivedAt:     info.receivedAt,
			Up: &ttnpb.ApplicationUp_UplinkNormalized{
				UplinkNormalized: &ttnpb.ApplicationUplinkNormalized{
					SessionKeyId:              info.uplink.SessionKeyId,
					FPort:                     info.uplink.FPort,
					FCnt:                      info.uplink.FCnt,
					FrmPayload:                info.uplink.FrmPayload,
					NormalizedPayload:         measurement,
					NormalizedPayloadWarnings: info.uplink.NormalizedPayloadWarnings,
					RxMetadata:                info.uplink.RxMetadata,
					Settings:                  info.uplink.Settings,
					ReceivedAt:                info.uplink.ReceivedAt,
					Confirmed:                 info.uplink.Confirmed,
					ConsumedAirtime:           info.uplink.ConsumedAirtime,
					Locations:                 info.uplink.Locations,
					VersionIds:                info.uplink.VersionIds,
					NetworkIds:                info.uplink.NetworkIds,
				},
			},
			Simulated: info.simulated,
		}); err != nil {
			return err
		}
	}
	return nil
}

type uplinkInfo struct {
	ids        *ttnpb.EndDeviceIdentifiers
	receivedAt *pbtypes.Timestamp
	uplink     *ttnpb.ApplicationUplink
	simulated  bool
	link       *ttnpb.ApplicationLink
}

func (as *ApplicationServer) handleUplink(ctx context.Context, info uplinkInfo) error {
	defer trace.StartRegion(ctx, "handle uplink").End()

	ctx = log.NewContextWithField(ctx, "session_key_id", info.uplink.SessionKeyId)
	dev, err := as.deviceRegistry.Set(ctx, info.ids,
		[]string{
			"activated_at",
			"formatters",
			"pending_session",
			"session",
			"skip_payload_crypto_override",
			"version_ids",
		},
		func(dev *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
			if dev == nil {
				return nil, nil, errDeviceNotFound.WithAttributes("device_uid", unique.ID(ctx, info.ids))
			}
			mask, err := as.matchSession(ctx, info.ids, dev, info.link, info.uplink.SessionKeyId)
			if err != nil {
				return nil, nil, err
			}
			if dev.Session.GetKeys().GetAppSKey() == nil {
				return nil, nil, errNoAppSKey.New()
			}
			return dev, mask, nil
		},
	)
	if err != nil {
		return err
	}

	if !as.skipPayloadCrypto(ctx, info.link, dev, dev.Session) {
		if err := as.decryptAndDecodeUplink(ctx, dev, info.uplink, info.link.DefaultFormatters); err != nil {
			return err
		}
		if err := as.publishNormalizedUplink(ctx, info); err != nil {
			return err
		}
		if err := as.storeUplink(ctx, info.ids, info.uplink); err != nil {
			return err
		}
	} else if appSKey := dev.GetSession().GetKeys().GetAppSKey(); appSKey != nil {
		info.uplink.AppSKey = appSKey
		info.uplink.LastAFCntDown = dev.Session.LastAFCntDown
	}

	registerUplinkLatency(ctx, info.uplink)

	if dev.VersionIds != nil {
		info.uplink.VersionIds = dev.VersionIds
	}

	// Set location in message and publish location solved if the payload contains location information.
	if locations, err := as.locationRegistry.Get(ctx, info.ids); err != nil {
		log.FromContext(ctx).WithError(err).Warn("Failed to retrieve end device locations")
	} else {
		info.uplink.Locations = locations
	}
	loc := as.locationFromPayload(info.uplink)
	if loc != nil {
		if info.uplink.Locations == nil {
			info.uplink.Locations = make(map[string]*ttnpb.Location, 1)
		}
		info.uplink.Locations["frm-payload"] = loc
		if err := as.Publish(ctx, &ttnpb.ApplicationUp{
			EndDeviceIds:   info.ids,
			CorrelationIds: events.CorrelationIDsFromContext(ctx),
			ReceivedAt:     info.receivedAt,
			Up: &ttnpb.ApplicationUp_LocationSolved{
				LocationSolved: &ttnpb.ApplicationLocation{
					Service:  "frm-payload",
					Location: loc,
				},
			},
		}); err != nil {
			log.FromContext(ctx).WithError(err).Warn("Failed to publish location solved message from location in payload")
		}
	}

	// If the device has not been activated before, publish the activation event.
	if dev.ActivatedAt == nil {
		if err := as.activationPool.Publish(ctx, info.ids); err != nil {
			log.FromContext(ctx).WithError(err).Warn("Failed to publish activation event")
		}
	}

	// Publish last seen event.
	if err := as.deviceLastSeenPool.Publish(ctx, lastSeenAtInfo{
		ids:        info.ids,
		lastSeenAt: info.uplink.ReceivedAt,
	}); err != nil {
		log.FromContext(ctx).WithError(err).Warn("Failed to publish last seen event")
	}

	return nil
}

func (as *ApplicationServer) handleSimulatedUplink(ctx context.Context, info uplinkInfo) error {
	defer trace.StartRegion(ctx, "handle simulated uplink").End()

	ctx = log.NewContextWithField(ctx, "session_key_id", info.uplink.SessionKeyId)
	dev, err := as.deviceRegistry.Get(ctx, info.ids,
		[]string{
			"formatters",
			"version_ids",
		},
	)
	if err != nil {
		return err
	}

	if locations, err := as.locationRegistry.Get(ctx, info.ids); err != nil {
		log.FromContext(ctx).WithError(err).Warn("Failed to retrieve end device locations")
	} else {
		info.uplink.Locations = locations
	}

	if err := as.decodeUplink(ctx, dev, info.uplink, info.link.DefaultFormatters); err != nil {
		return err
	}
	return as.publishNormalizedUplink(ctx, info)
}

func (as *ApplicationServer) handleDownlinkQueueInvalidated(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, invalid *ttnpb.ApplicationInvalidatedDownlinks, link *ttnpb.ApplicationLink) (pass bool, err error) {
	defer trace.StartRegion(ctx, "handle downlink queue invalidated").End()

	peer, err := as.GetPeer(ctx, ttnpb.ClusterRole_NETWORK_SERVER, nil)
	if err != nil {
		return false, err
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
			if dev == nil {
				return nil, nil, errDeviceNotFound.WithAttributes("device_uid", unique.ID(ctx, ids))
			}

			mask := []string{"session.last_a_f_cnt_down"}

			if as.skipPayloadCrypto(ctx, link, dev, dev.Session) {
				// When skipping application payload crypto, the upstream application is responsible for recalculating the
				// downlink queue. No error is returned here to pass the downlink queue invalidation message upstream.
				pass = true
				dev.Session.LastAFCntDown = invalid.LastFCntDown
				return dev, mask, nil
			}

			matchMask, err := as.matchSession(ctx, ids, dev, link, invalid.SessionKeyId)
			if err != nil {
				return nil, nil, err
			}
			mask = ttnpb.AddFields(mask, matchMask...)
			dev.Session.LastAFCntDown = invalid.LastFCntDown

			items := make([]*ttnpb.ApplicationDownlink, 0, len(invalid.Downlinks))
			for _, msg := range invalid.Downlinks {
				if err := as.decryptDownlink(ctx, dev, msg, nil); err != nil {
					log.FromContext(ctx).WithError(err).Warn("Failed to decrypt downlink message; drop item")
					registerDropDownlink(ctx, ids, msg, err)
					continue
				}
				items = append(items, msg)
			}
			if len(items) == 0 {
				return dev, mask, nil
			}

			pushMask, err := as.attemptDownlinkQueueOp(ctx, dev, link, peer, downlinkQueueOperation{
				Items:     items,
				Operation: ttnpb.AsNsClient.DownlinkQueuePush,
				ResultFunc: func(decrypted, _ []*ttnpb.ApplicationDownlink, err error) {
					if err != nil {
						as.registerDropDownlinks(ctx, ids, decrypted, err)
					}
				},
			})
			if err != nil {
				return nil, nil, err
			}
			mask = ttnpb.AddFields(mask, pushMask...)

			return dev, mask, nil
		},
	)
	return pass, err
}

func (as *ApplicationServer) handleDownlinkNack(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, msg *ttnpb.ApplicationDownlink, link *ttnpb.ApplicationLink) error {
	defer trace.StartRegion(ctx, "handle downlink nack").End()

	peer, err := as.GetPeer(ctx, ttnpb.ClusterRole_NETWORK_SERVER, nil)
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
			if dev == nil {
				return nil, nil, errDeviceNotFound.WithAttributes("device_uid", unique.ID(ctx, ids))
			}

			if as.skipPayloadCrypto(ctx, link, dev, dev.Session) {
				// When skipping application payload crypto, the upstream application is responsible for recalculating the
				// downlink queue. No error is returned here to pass the downlink nack message upstream.
				return dev, nil, nil
			}

			matchMask, err := as.matchSession(ctx, ids, dev, link, msg.SessionKeyId)
			if err != nil {
				return nil, nil, err
			}

			// Decrypt the message as it will be sent to upstream after handling it.
			if err := as.decryptAndDecodeDownlink(ctx, dev, msg, link.DefaultFormatters); err != nil {
				return nil, nil, err
			}

			items := []*ttnpb.ApplicationDownlink{msg}
			pushMask, err := as.attemptDownlinkQueueOp(ctx, dev, link, peer, downlinkQueueOperation{
				Items:     items,
				Operation: ttnpb.AsNsClient.DownlinkQueuePush,
				ResultFunc: func(decrypted, _ []*ttnpb.ApplicationDownlink, err error) {
					if err != nil {
						as.registerDropDownlinks(ctx, ids, decrypted, err)
					}
				},
			})
			if err != nil {
				return nil, nil, err
			}
			mask := ttnpb.AddFields(matchMask, pushMask...)

			return dev, mask, nil
		},
	)
	return err
}

// handleLocationSolved saves the provided *ttnpb.ApplicationLocation in the Entity Registry as part of the device locations.
// Locations provided by other services will be maintained.
func (as *ApplicationServer) handleLocationSolved(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, msg *ttnpb.ApplicationLocation, link *ttnpb.ApplicationLink) error {
	defer trace.StartRegion(ctx, "handle location solved").End()

	if _, err := as.locationRegistry.Merge(ctx, ids, map[string]*ttnpb.Location{
		msg.Service: msg.Location,
	}); err != nil {
		log.FromContext(ctx).WithError(err).Warn("Failed to merge end device locations")
	}
	return nil
}

// decryptDownlinkMessage decrypts the downlink message.
// If application payload crypto is skipped, this method returns nil.
func (as *ApplicationServer) decryptDownlinkMessage(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, msg *ttnpb.ApplicationDownlink, link *ttnpb.ApplicationLink) error {
	defer trace.StartRegion(ctx, "decrypt downlink message").End()

	dev, err := as.deviceRegistry.Get(ctx, ids, []string{
		"formatters",
		"pending_session",
		"session",
		"skip_payload_crypto_override",
		"version_ids",
	})
	if err != nil {
		return err
	}
	var session *ttnpb.Session
	switch {
	case dev.Session != nil && bytes.Equal(dev.Session.Keys.SessionKeyId, msg.SessionKeyId):
		session = dev.Session
	case dev.PendingSession != nil && bytes.Equal(dev.PendingSession.Keys.SessionKeyId, msg.SessionKeyId):
		session = dev.PendingSession
	}
	if as.skipPayloadCrypto(ctx, link, dev, session) {
		return nil
	}
	return as.decryptAndDecodeDownlink(ctx, dev, msg, link.DefaultFormatters)
}

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
	cfg, err := as.GetConfig(ctx)
	if err != nil {
		return nil, err
	}
	return &cfg.MQTT, nil
}

// RangeUplinks ranges the application uplinks and calls the callback function, until false is returned.
func (as *ApplicationServer) RangeUplinks(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, paths []string, f func(ctx context.Context, up *ttnpb.ApplicationUplink) bool) error {
	return as.appUpsRegistry.Range(ctx, ids, paths, f)
}

// GetEndDevice retrieves the end device associated with the provided identifiers from the end device registry.
func (as *ApplicationServer) GetEndDevice(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, paths []string) (*ttnpb.EndDevice, error) {
	return as.deviceRegistry.Get(ctx, ids, paths)
}
