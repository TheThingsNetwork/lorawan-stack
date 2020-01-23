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
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"go.thethings.network/lorawan-stack/cmd/internal/shared"
	"go.thethings.network/lorawan-stack/pkg/applicationserver"
	asioapredis "go.thethings.network/lorawan-stack/pkg/applicationserver/io/packages/redis"
	asiopsredis "go.thethings.network/lorawan-stack/pkg/applicationserver/io/pubsub/redis"
	asiowebredis "go.thethings.network/lorawan-stack/pkg/applicationserver/io/web/redis"
	asredis "go.thethings.network/lorawan-stack/pkg/applicationserver/redis"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/console"
	"go.thethings.network/lorawan-stack/pkg/devicetemplateconverter"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/events"
	events_grpc "go.thethings.network/lorawan-stack/pkg/events/grpc"
	"go.thethings.network/lorawan-stack/pkg/gatewayconfigurationserver"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver"
	"go.thethings.network/lorawan-stack/pkg/identityserver"
	"go.thethings.network/lorawan-stack/pkg/joinserver"
	jsredis "go.thethings.network/lorawan-stack/pkg/joinserver/redis"
	"go.thethings.network/lorawan-stack/pkg/networkserver"
	nsredis "go.thethings.network/lorawan-stack/pkg/networkserver/redis"
	"go.thethings.network/lorawan-stack/pkg/qrcodegenerator"
	"go.thethings.network/lorawan-stack/pkg/redis"
	"go.thethings.network/lorawan-stack/pkg/web"
)

var errUnknownComponent = errors.DefineInvalidArgument("unknown_component", "unknown component `{component}`")

var startCommand = &cobra.Command{
	Use:   "start [is|gs|ns|as|js|console|gcs|dtc|qrg|all]... [flags]",
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
			default:
				return errUnknownComponent.WithAttributes("component", arg)
			}
		}

		logger.Info("Setting up core component")

		var rootRedirect web.Registerer

		var componentOptions []component.Option

		c, err := component.New(logger, &component.Config{ServiceBase: config.ServiceBase}, componentOptions...)
		if err != nil {
			return shared.ErrInitializeBaseComponent.WithCause(err)
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
			is, err := identityserver.New(c, &config.IS)
			if err != nil {
				return shared.ErrInitializeIdentityServer.WithCause(err)
			}
			if config.Cache.Service == "redis" {
				is.SetRedisCache(redis.New(&redis.Config{
					Redis:     config.Cache.Redis,
					Namespace: []string{"is", "cache"},
				}))
			}
			if oauthMount := config.IS.OAuth.UI.MountPath(); oauthMount != "/" {
				rootRedirect = web.Redirect("/", http.StatusFound, oauthMount)
			}
		}

		if start.GatewayServer || startDefault {
			logger.Info("Setting up Gateway Server")
			gs, err := gatewayserver.New(c, &config.GS)
			if err != nil {
				return shared.ErrInitializeGatewayServer.WithCause(err)
			}
			_ = gs
		}

		if start.NetworkServer || startDefault {
			redisConsumerGroup := "ns"

			logger.Info("Setting up Network Server")
			config.NS.ApplicationUplinks = nsredis.NewApplicationUplinkQueue(redis.New(&redis.Config{
				Redis:     config.Redis,
				Namespace: []string{"ns", "application-uplinks"},
			}), 100, redisConsumerGroup, redisConsumerID)
			config.NS.Devices = &nsredis.DeviceRegistry{Redis: redis.New(&redis.Config{
				Redis:     config.Redis,
				Namespace: []string{"ns", "devices"},
			})}
			nsDownlinkTasks := nsredis.NewDownlinkTaskQueue(redis.New(&redis.Config{
				Redis:     config.Redis,
				Namespace: []string{"ns", "tasks"},
			}), 100000, redisConsumerGroup, redisConsumerID)
			if err := nsDownlinkTasks.Init(); err != nil {
				return shared.ErrInitializeNetworkServer.WithCause(err)
			}
			config.NS.DownlinkTasks = nsDownlinkTasks
			ns, err := networkserver.New(c, &config.NS)
			if err != nil {
				return shared.ErrInitializeNetworkServer.WithCause(err)
			}
			ns.Component.RegisterTask(ns.Context(), "queue_downlink", nsDownlinkTasks.Run, component.TaskRestartOnFailure)
		}

		if start.ApplicationServer || startDefault {
			logger.Info("Setting up Application Server")
			config.AS.Links = &asredis.LinkRegistry{Redis: redis.New(&redis.Config{
				Redis:     config.Redis,
				Namespace: []string{"as", "links"},
			})}
			config.AS.Devices = &asredis.DeviceRegistry{Redis: redis.New(&redis.Config{
				Redis:     config.Redis,
				Namespace: []string{"as", "devices"},
			})}
			config.AS.PubSub.Registry = &asiopsredis.PubSubRegistry{Redis: redis.New(&redis.Config{
				Redis:     config.Redis,
				Namespace: []string{"as", "io", "pubsub"},
			})}
			config.AS.ApplicationPackages.Registry = &asioapredis.ApplicationPackagesRegistry{Redis: redis.New(&redis.Config{
				Redis:     config.Redis,
				Namespace: []string{"as", "io", "applicationpackages"},
			})}
			if config.AS.Webhooks.Target != "" {
				config.AS.Webhooks.Registry = &asiowebredis.WebhookRegistry{Redis: redis.New(&redis.Config{
					Redis:     config.Redis,
					Namespace: []string{"as", "io", "webhooks"},
				})}
			}
			as, err := applicationserver.New(c, &config.AS)
			if err != nil {
				return shared.ErrInitializeApplicationServer.WithCause(err)
			}
			_ = as
		}

		if start.JoinServer || startDefault {
			logger.Info("Setting up Join Server")
			config.JS.Devices = &jsredis.DeviceRegistry{Redis: redis.New(&redis.Config{
				Redis:     config.Redis,
				Namespace: []string{"js", "devices"},
			})}
			config.JS.Keys = &jsredis.KeyRegistry{Redis: redis.New(&redis.Config{
				Redis:     config.Redis,
				Namespace: []string{"js", "keys"},
			})}
			js, err := joinserver.New(c, &config.JS)
			if err != nil {
				return shared.ErrInitializeJoinServer.WithCause(err)
			}
			_ = js
		}

		if start.Console || startDefault {
			logger.Info("Setting up Console")
			console, err := console.New(c, config.Console)
			if err != nil {
				return shared.ErrInitializeConsole.WithCause(err)
			}
			_ = console
			if consoleMount := config.Console.UI.MountPath(); consoleMount != "/" {
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

		if rootRedirect != nil {
			c.RegisterWeb(rootRedirect)
		}

		logger.Info("Starting...")

		return c.Run()
	},
}

func init() {
	Root.AddCommand(startCommand)
}
