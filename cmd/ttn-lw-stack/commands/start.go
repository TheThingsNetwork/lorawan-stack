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

package commands

import (
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"go.thethings.network/lorawan-stack/v3/cmd/internal/shared"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver"
	asdistribredis "go.thethings.network/lorawan-stack/v3/pkg/applicationserver/distribution/redis"
	asioapredis "go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/packages/redis"
	asiopsredis "go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/pubsub/redis"
	asiowebredis "go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/web/redis"
	asmetaredis "go.thethings.network/lorawan-stack/v3/pkg/applicationserver/metadata/redis"
	asredis "go.thethings.network/lorawan-stack/v3/pkg/applicationserver/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	"go.thethings.network/lorawan-stack/v3/pkg/console"
	"go.thethings.network/lorawan-stack/v3/pkg/deviceclaimingserver"
	"go.thethings.network/lorawan-stack/v3/pkg/devicerepository"
	"go.thethings.network/lorawan-stack/v3/pkg/devicetemplateconverter"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	events_grpc "go.thethings.network/lorawan-stack/v3/pkg/events/grpc"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayconfigurationserver"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver"
	gsredis "go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver"
	"go.thethings.network/lorawan-stack/v3/pkg/joinserver"
	jsredis "go.thethings.network/lorawan-stack/v3/pkg/joinserver/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver"
	nsredis "go.thethings.network/lorawan-stack/v3/pkg/networkserver/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/packetbrokeragent"
	"go.thethings.network/lorawan-stack/v3/pkg/qrcodegenerator"
	"go.thethings.network/lorawan-stack/v3/pkg/random"
	"go.thethings.network/lorawan-stack/v3/pkg/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/web"
)

const defaultLockTTL = 10 * time.Second

func NewComponentDeviceRegistryRedis(conf Config, name string) *redis.Client {
	return redis.New(conf.Redis.WithNamespace(name, "devices"))
}

func NewNetworkServerDeviceRegistryRedis(conf Config) *redis.Client {
	return NewComponentDeviceRegistryRedis(conf, "ns")
}

func NewNetworkServerApplicationUplinkQueueRedis(conf Config) *redis.Client {
	return redis.New(conf.Redis.WithNamespace("ns", "application-uplinks"))
}

func NewNetworkServerDownlinkTaskRedis(conf Config) *redis.Client {
	return redis.New(conf.Redis.WithNamespace("ns", "tasks"))
}

func NewApplicationServerDeviceRegistryRedis(conf Config) *redis.Client {
	return NewComponentDeviceRegistryRedis(conf, "as")
}

var errUnknownComponent = errors.DefineInvalidArgument("unknown_component", "unknown component `{component}`")

var startCommand = &cobra.Command{
	Use:   "start [is|gs|ns|as|js|console|gcs|dtc|qrg|pba|dcs|all]... [flags]",
	Short: "Start The Things Stack",
	RunE: func(cmd *cobra.Command, args []string) error {
		var start struct {
			IdentityServer             bool
			GatewayServer              bool
			NetworkServer              bool
			ApplicationServer          bool
			JoinServer                 bool
			Console                    bool
			GatewayConfigurationServer bool
			DeviceTemplateConverter    bool
			QRCodeGenerator            bool
			PacketBrokerAgent          bool
			DeviceRepository           bool
			DeviceClaimingServer       bool
		}
		startDefault := len(args) == 0
		for _, arg := range args {
			switch strings.ToLower(arg) {
			case "is", "identityserver":
				start.IdentityServer = true
				start.DeviceTemplateConverter = true
				start.QRCodeGenerator = true
			case "gs", "gatewayserver":
				start.GatewayServer = true
			case "ns", "networkserver":
				start.NetworkServer = true
				start.DeviceTemplateConverter = true
				start.QRCodeGenerator = true
			case "as", "applicationserver":
				start.ApplicationServer = true
				start.DeviceTemplateConverter = true
				start.QRCodeGenerator = true
			case "js", "joinserver":
				start.JoinServer = true
				start.DeviceTemplateConverter = true
				start.QRCodeGenerator = true
			case "console":
				start.Console = true
			case "gcs":
				start.GatewayConfigurationServer = true
			case "dtc":
				start.DeviceTemplateConverter = true
			case "qrg":
				start.QRCodeGenerator = true
			case "pba":
				start.PacketBrokerAgent = true
			case "dr":
				start.DeviceRepository = true
			case "dcs":
				start.DeviceClaimingServer = true
			case "all":
				start.IdentityServer = true
				start.GatewayServer = true
				start.NetworkServer = true
				start.ApplicationServer = true
				start.JoinServer = true
				start.Console = true
				start.GatewayConfigurationServer = true
				start.DeviceTemplateConverter = true
				start.QRCodeGenerator = true
				start.PacketBrokerAgent = true
				start.DeviceRepository = true
				start.DeviceClaimingServer = true
			default:
				return errUnknownComponent.WithAttributes("component", arg)
			}
		}

		if startDefault {
			start.IdentityServer = true
			start.GatewayServer = true
			start.NetworkServer = true
			start.ApplicationServer = true
			start.JoinServer = true
			start.Console = true
			start.GatewayConfigurationServer = true
			start.DeviceTemplateConverter = true
			start.QRCodeGenerator = true
			start.PacketBrokerAgent = true
			start.DeviceRepository = true
		}

		logger.Info("Setting up core component")

		var rootRedirect web.Registerer

		var componentOptions []component.Option

		cookieHashKey, cookieBlockKey := config.ServiceBase.HTTP.Cookie.HashKey, config.ServiceBase.HTTP.Cookie.BlockKey

		if len(cookieHashKey) == 0 || isZeros(cookieHashKey) {
			cookieHashKey = random.Bytes(64)
			config.ServiceBase.HTTP.Cookie.HashKey = cookieHashKey
			logger.Warn("No cookie hash key configured, generated a random one")
		}

		if len(cookieBlockKey) == 0 || isZeros(cookieBlockKey) {
			cookieBlockKey = random.Bytes(32)
			config.ServiceBase.HTTP.Cookie.BlockKey = cookieBlockKey
			logger.Warn("No cookie block key configured, generated a random one")
		}

		c, err := component.New(logger, &component.Config{ServiceBase: config.ServiceBase}, componentOptions...)
		if err != nil {
			return shared.ErrInitializeBaseComponent.WithCause(err)
		}

		if err := shared.InitializeEvents(ctx, c, config.ServiceBase); err != nil {
			return err
		}

		c.RegisterGRPC(events_grpc.NewEventsServer(c.Context(), events.DefaultPubSub()))
		c.RegisterGRPC(component.NewConfigurationServer(c))

		if start.IdentityServer {
			logger.Info("Setting up Identity Server")
			if config.IS.OAuth.UI.TemplateData.SentryDSN == "" {
				config.IS.OAuth.UI.TemplateData.SentryDSN = config.Sentry.DSN
			}
			is, err := identityserver.New(c, &config.IS)
			if err != nil {
				return shared.ErrInitializeIdentityServer.WithCause(err)
			}
			if config.Cache.Service == "redis" {
				is.SetRedisCache(redis.New(config.Cache.Redis.WithNamespace("is", "cache")))
			}
			if accountAppMount := config.IS.OAuth.UI.MountPath(); accountAppMount != "/" {
				if !strings.HasSuffix(accountAppMount, "/") {
					accountAppMount += "/"
				}
				rootRedirect = web.Redirect("/", http.StatusFound, accountAppMount)
			}
		}

		if start.GatewayServer {
			logger.Info("Setting up Gateway Server")
			switch config.Cache.Service {
			case "redis":
				gatewayConnectionStatsRegistry := &gsredis.GatewayConnectionStatsRegistry{
					Redis:   redis.New(config.Cache.Redis.WithNamespace("gs", "cache", "connstats")),
					LockTTL: defaultLockTTL,
				}
				if err := gatewayConnectionStatsRegistry.Init(ctx); err != nil {
					return shared.ErrInitializeGatewayServer.WithCause(err)
				}
				config.GS.Stats = gatewayConnectionStatsRegistry
			}
			gs, err := gatewayserver.New(c, &config.GS)
			if err != nil {
				return shared.ErrInitializeGatewayServer.WithCause(err)
			}
			_ = gs
		}

		if start.NetworkServer {
			redisConsumerGroup := "ns"

			logger.Info("Setting up Network Server")

			applicationUplinkQueueSize := config.NS.ApplicationUplinkQueue.BufferSize
			if config.NS.ApplicationUplinkQueue.BufferSize > math.MaxInt64 {
				applicationUplinkQueueSize = math.MaxInt64
			}
			applicationUplinkQueue := nsredis.NewApplicationUplinkQueue(
				NewNetworkServerApplicationUplinkQueueRedis(*config),
				int64(applicationUplinkQueueSize), redisConsumerGroup, time.Minute,
			)
			if err := applicationUplinkQueue.Init(ctx); err != nil {
				return shared.ErrInitializeNetworkServer.WithCause(err)
			}
			defer applicationUplinkQueue.Close(ctx)
			config.NS.ApplicationUplinkQueue.Queue = applicationUplinkQueue
			devices := &nsredis.DeviceRegistry{
				Redis:   NewNetworkServerDeviceRegistryRedis(*config),
				LockTTL: defaultLockTTL,
			}
			if err := devices.Init(ctx); err != nil {
				return shared.ErrInitializeNetworkServer.WithCause(err)
			}
			config.NS.Devices = devices
			config.NS.UplinkDeduplicator = &nsredis.UplinkDeduplicator{
				Redis: redis.New(config.Cache.Redis.WithNamespace("ns", "uplink-deduplication")),
			}
			downlinkTasks := nsredis.NewDownlinkTaskQueue(
				NewNetworkServerDownlinkTaskRedis(*config),
				100000, redisConsumerGroup,
			)
			if err := downlinkTasks.Init(ctx); err != nil {
				return shared.ErrInitializeNetworkServer.WithCause(err)
			}
			defer downlinkTasks.Close(ctx)
			config.NS.DownlinkTaskQueue.Queue = downlinkTasks
			config.NS.ScheduledDownlinkMatcher = &nsredis.ScheduledDownlinkMatcher{
				Redis: redis.New(config.Redis.WithNamespace("ns", "scheduled-downlinks")),
			}
			ns, err := networkserver.New(c, &config.NS)
			if err != nil {
				return shared.ErrInitializeNetworkServer.WithCause(err)
			}
			_ = ns
		}

		if start.ApplicationServer {
			logger.Info("Setting up Application Server")
			linkRegistry := &asredis.LinkRegistry{
				Redis:   redis.New(config.Redis.WithNamespace("as", "links")),
				LockTTL: defaultLockTTL,
			}
			if err := linkRegistry.Init(ctx); err != nil {
				return shared.ErrInitializeApplicationServer.WithCause(err)
			}
			config.AS.Links = linkRegistry
			deviceRegistry := &asredis.DeviceRegistry{
				Redis:   NewApplicationServerDeviceRegistryRedis(*config),
				LockTTL: defaultLockTTL,
			}
			if err := deviceRegistry.Init(ctx); err != nil {
				return shared.ErrInitializeApplicationServer.WithCause(err)
			}
			config.AS.Devices = deviceRegistry
			config.AS.UplinkStorage.Registry = &asredis.ApplicationUplinkRegistry{
				Redis: redis.New(config.Redis.WithNamespace("as", "applicationups")),
				Limit: config.AS.UplinkStorage.Limit,
			}
			config.AS.Distribution.Global.PubSub = &asdistribredis.PubSub{
				Redis: redis.New(config.Cache.Redis.WithNamespace("as", "traffic")),
			}
			pubsubRegistry := &asiopsredis.PubSubRegistry{
				Redis:   redis.New(config.Redis.WithNamespace("as", "io", "pubsub")),
				LockTTL: defaultLockTTL,
			}
			if err := pubsubRegistry.Init(ctx); err != nil {
				return shared.ErrInitializeApplicationServer.WithCause(err)
			}
			config.AS.PubSub.Registry = pubsubRegistry
			applicationPackagesRegistry := &asioapredis.ApplicationPackagesRegistry{
				Redis:   redis.New(config.Redis.WithNamespace("as", "io", "applicationpackages")),
				LockTTL: defaultLockTTL,
			}
			if err := applicationPackagesRegistry.Init(ctx); err != nil {
				return shared.ErrInitializeApplicationServer.WithCause(err)
			}
			config.AS.Packages.Registry = applicationPackagesRegistry
			if config.AS.Webhooks.Target != "" {
				webhookRegistry := &asiowebredis.WebhookRegistry{
					Redis:   redis.New(config.Redis.WithNamespace("as", "io", "webhooks")),
					LockTTL: defaultLockTTL,
				}
				if err := webhookRegistry.Init(ctx); err != nil {
					return shared.ErrInitializeApplicationServer.WithCause(err)
				}
				config.AS.Webhooks.Registry = webhookRegistry
			}
			if cache := &config.AS.EndDeviceMetadataStorage.Location.Cache; cache.Enable {
				switch config.Cache.Service {
				case "redis":
					cache.Cache = &asmetaredis.EndDeviceLocationCache{
						Redis: redis.New(config.Cache.Redis.WithNamespace("as", "metadata", "locations")),
					}
				default:
					cache.Enable = false
				}
			}
			locationRegistry, err := config.AS.EndDeviceMetadataStorage.Location.NewRegistry(ctx, c)
			if err != nil {
				return shared.ErrInitializeApplicationServer.WithCause(err)
			}
			config.AS.EndDeviceMetadataStorage.Location.Registry = locationRegistry
			as, err := applicationserver.New(c, &config.AS)
			if err != nil {
				return shared.ErrInitializeApplicationServer.WithCause(err)
			}
			_ = as
		}

		if start.JoinServer {
			logger.Info("Setting up Join Server")
			deviceRegistry := &jsredis.DeviceRegistry{
				Redis:   NewComponentDeviceRegistryRedis(*config, "js"),
				LockTTL: defaultLockTTL,
			}
			if err := deviceRegistry.Init(ctx); err != nil {
				return shared.ErrInitializeJoinServer.WithCause(err)
			}
			config.JS.Devices = deviceRegistry
			keyRegistry := &jsredis.KeyRegistry{
				Redis:   redis.New(config.Redis.WithNamespace("js", "keys")),
				LockTTL: defaultLockTTL,
			}
			if err := keyRegistry.Init(ctx); err != nil {
				return shared.ErrInitializeJoinServer.WithCause(err)
			}
			config.JS.Keys = keyRegistry
			applicationActivationSettingRegistry := &jsredis.ApplicationActivationSettingRegistry{
				Redis:   redis.New(config.Redis.WithNamespace("js", "application-activation-settings")),
				LockTTL: defaultLockTTL,
			}
			if err := applicationActivationSettingRegistry.Init(ctx); err != nil {
				return shared.ErrInitializeJoinServer.WithCause(err)
			}
			config.JS.ApplicationActivationSettings = applicationActivationSettingRegistry
			js, err := joinserver.New(c, &config.JS)
			if err != nil {
				return shared.ErrInitializeJoinServer.WithCause(err)
			}
			_ = js
		}

		if start.Console {
			logger.Info("Setting up Console")
			if config.Console.UI.TemplateData.SentryDSN == "" {
				config.Console.UI.TemplateData.SentryDSN = config.Sentry.DSN
			}
			console, err := console.New(c, config.Console)
			if err != nil {
				return shared.ErrInitializeConsole.WithCause(err)
			}
			_ = console
			if consoleMount := config.Console.UI.MountPath(); consoleMount != "/" {
				if !strings.HasSuffix(consoleMount, "/") {
					consoleMount += "/"
				}
				rootRedirect = web.Redirect("/", http.StatusFound, consoleMount)
			}
		}

		if start.GatewayConfigurationServer {
			logger.Info("Setting up Gateway Configuration Server")
			gcs, err := gatewayconfigurationserver.New(c, &config.GCS)
			if err != nil {
				return shared.ErrInitializeGatewayConfigurationServer.WithCause(err)
			}
			_ = gcs
		}

		if start.DeviceTemplateConverter {
			logger.Info("Setting up Device Template Converter")
			dtc, err := devicetemplateconverter.New(c, &config.DTC)
			if err != nil {
				return shared.ErrInitializeDeviceTemplateConverter.WithCause(err)
			}
			_ = dtc
		}

		if start.QRCodeGenerator {
			logger.Info("Setting up QR Code Generator")
			qrg, err := qrcodegenerator.New(c, &config.QRG)
			if err != nil {
				return shared.ErrInitializeQRCodeGenerator.WithCause(err)
			}
			_ = qrg
		}

		if start.PacketBrokerAgent {
			logger.Info("Setting up Packet Broker Agent")
			pba, err := packetbrokeragent.New(c, &config.PBA)
			if err != nil {
				return shared.ErrInitializePacketBrokerAgent.WithCause(err)
			}
			_ = pba
		}

		if start.DeviceRepository {
			logger.Info("Setting up Device Repository")
			store, err := config.DR.NewStore(ctx, config.Blob)
			if err != nil {
				return shared.ErrInitializeDeviceRepository.WithCause(err)
			}
			config.DR.Store.Store = store
			dr, err := devicerepository.New(c, &config.DR)
			if err != nil {
				return shared.ErrInitializeDeviceRepository.WithCause(err)
			}
			_ = dr
		}

		if start.DeviceClaimingServer {
			logger.Info("Setting up Device Claiming Server")

			dcs, err := deviceclaimingserver.New(c, &config.DCS, nil)
			if err != nil {
				return shared.ErrInitializeDeviceClaimingServer.WithCause(err)
			}

			_ = dcs
		}

		if rootRedirect != nil {
			c.RegisterWeb(rootRedirect)
		}

		logger.Info("Starting...")

		return c.Run()
	},
}

func isZeros(buf []byte) bool {
	for _, b := range buf {
		if b != 0x00 {
			return false
		}
	}

	return true
}

func init() {
	Root.AddCommand(startCommand)
}
