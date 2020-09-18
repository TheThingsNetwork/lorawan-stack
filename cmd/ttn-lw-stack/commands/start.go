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
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"go.thethings.network/lorawan-stack/v3/cmd/internal/shared"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver"
	asdistribredis "go.thethings.network/lorawan-stack/v3/pkg/applicationserver/distribution/redis"
	asioapredis "go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/packages/redis"
	asiopsredis "go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/pubsub/redis"
	asiowebredis "go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/web/redis"
	asredis "go.thethings.network/lorawan-stack/v3/pkg/applicationserver/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	"go.thethings.network/lorawan-stack/v3/pkg/console"
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

var errUnknownComponent = errors.DefineInvalidArgument("unknown_component", "unknown component `{component}`")

var startCommand = &cobra.Command{
	Use:   "start [is|gs|ns|as|js|console|gcs|dtc|qrg|pba|all]... [flags]",
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
			default:
				return errUnknownComponent.WithAttributes("component", arg)
			}
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

		host, err := os.Hostname()
		if err != nil {
			return err
		}

		redisConsumerID := redis.Key(host, strconv.Itoa(os.Getpid()))

		if start.IdentityServer || startDefault {
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
			if oauthMount := config.IS.OAuth.UI.MountPath(); oauthMount != "/" {
				if !strings.HasSuffix(oauthMount, "/") {
					oauthMount += "/"
				}
				rootRedirect = web.Redirect("/", http.StatusFound, oauthMount)
			}
		}

		if start.GatewayServer || startDefault {
			logger.Info("Setting up Gateway Server")
			switch config.Cache.Service {
			case "redis":
				config.GS.Stats = &gsredis.GatewayConnectionStatsRegistry{
					Redis: redis.New(config.Cache.Redis.WithNamespace("gs", "cache", "connstats")),
				}
			}
			gs, err := gatewayserver.New(c, &config.GS)
			if err != nil {
				return shared.ErrInitializeGatewayServer.WithCause(err)
			}
			_ = gs
		}

		if start.NetworkServer || startDefault {
			redisConsumerGroup := "ns"

			logger.Info("Setting up Network Server")

			uplinkQueueSize := config.NS.ApplicationUplinkQueue.BufferSize
			if config.NS.ApplicationUplinkQueue.BufferSize > math.MaxInt64 {
				uplinkQueueSize = math.MaxInt64
			}
			config.NS.ApplicationUplinkQueue.Queue = nsredis.NewApplicationUplinkQueue(
				NewNetworkServerApplicationUplinkQueueRedis(*config),
				int64(uplinkQueueSize), redisConsumerGroup, redisConsumerID,
			)
			devices := &nsredis.DeviceRegistry{
				Redis:   NewNetworkServerDeviceRegistryRedis(*config),
				LockTTL: time.Second,
			}
			if err := devices.Init(); err != nil {
				return shared.ErrInitializeNetworkServer.WithCause(err)
			}
			config.NS.Devices = devices
			config.NS.UplinkDeduplicator = &nsredis.UplinkDeduplicator{
				Redis: redis.New(config.Cache.Redis.WithNamespace("ns", "uplink-deduplication")),
			}
			nsDownlinkTasks := nsredis.NewDownlinkTaskQueue(
				NewNetworkServerDownlinkTaskRedis(*config),
				100000, redisConsumerGroup, redisConsumerID,
			)
			if err := nsDownlinkTasks.Init(); err != nil {
				return shared.ErrInitializeNetworkServer.WithCause(err)
			}
			config.NS.DownlinkTasks = nsDownlinkTasks
			ns, err := networkserver.New(c, &config.NS)
			if err != nil {
				return shared.ErrInitializeNetworkServer.WithCause(err)
			}
			ns.Component.RegisterTask(&component.TaskConfig{
				Context: ns.Context(),
				ID:      "queue_downlink",
				Func:    nsDownlinkTasks.Run,
				Restart: component.TaskRestartOnFailure,
				Backoff: component.DefaultTaskBackoffConfig,
			})
		}

		if start.ApplicationServer || startDefault {
			logger.Info("Setting up Application Server")
			config.AS.Links = &asredis.LinkRegistry{
				Redis: redis.New(config.Redis.WithNamespace("as", "links")),
			}
			config.AS.Devices = &asredis.DeviceRegistry{
				Redis: NewComponentDeviceRegistryRedis(*config, "as"),
			}
			config.AS.Distribution.PubSub = &asdistribredis.PubSub{
				Redis: redis.New(config.Cache.Redis.WithNamespace("as", "traffic")),
			}
			config.AS.PubSub.Registry = &asiopsredis.PubSubRegistry{
				Redis: redis.New(config.Redis.WithNamespace("as", "io", "pubsub")),
			}
			config.AS.Packages.Registry = &asioapredis.ApplicationPackagesRegistry{
				Redis: redis.New(config.Redis.WithNamespace("as", "io", "applicationpackages")),
			}
			if config.AS.Webhooks.Target != "" {
				config.AS.Webhooks.Registry = &asiowebredis.WebhookRegistry{
					Redis: redis.New(config.Redis.WithNamespace("as", "io", "webhooks")),
				}
			}
			as, err := applicationserver.New(c, &config.AS)
			if err != nil {
				return shared.ErrInitializeApplicationServer.WithCause(err)
			}
			_ = as
		}

		if start.JoinServer || startDefault {
			logger.Info("Setting up Join Server")
			config.JS.Devices = &jsredis.DeviceRegistry{
				Redis: NewComponentDeviceRegistryRedis(*config, "js"),
			}
			config.JS.Keys = &jsredis.KeyRegistry{
				Redis: redis.New(config.Redis.WithNamespace("js", "keys")),
			}
			js, err := joinserver.New(c, &config.JS)
			if err != nil {
				return shared.ErrInitializeJoinServer.WithCause(err)
			}
			_ = js
		}

		if start.Console || startDefault {
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

		if start.GatewayConfigurationServer || startDefault {
			logger.Info("Setting up Gateway Configuration Server")
			gcs, err := gatewayconfigurationserver.New(c, &config.GCS)
			if err != nil {
				return shared.ErrInitializeGatewayConfigurationServer.WithCause(err)
			}
			_ = gcs
		}

		if start.DeviceTemplateConverter || startDefault {
			logger.Info("Setting up Device Template Converter")
			dtc, err := devicetemplateconverter.New(c, &config.DTC)
			if err != nil {
				return shared.ErrInitializeDeviceTemplateConverter.WithCause(err)
			}
			_ = dtc
		}

		if start.QRCodeGenerator || startDefault {
			logger.Info("Setting up QR Code Generator")
			qrg, err := qrcodegenerator.New(c, &config.QRG)
			if err != nil {
				return shared.ErrInitializeQRCodeGenerator.WithCause(err)
			}
			_ = qrg
		}

		if start.PacketBrokerAgent || startDefault {
			logger.Info("Setting up Packet Broker Agent")
			pba, err := packetbrokeragent.New(c, &config.PBA)
			if err != nil {
				return shared.ErrInitializePacketBrokerAgent.WithCause(err)
			}
			_ = pba
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
