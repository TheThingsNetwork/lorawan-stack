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

// Package networkserver provides a LoRaWAN-compliant Network Server implementation.
package networkserver

import (
	"context"
	"crypto/tls"
	"fmt"
	"sync"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/interop"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal/time"
	"go.thethings.network/lorawan-stack/v3/pkg/random"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmiddleware/hooks"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmiddleware/rpclog"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"google.golang.org/grpc"
)

const (
	// recentDownlinkCount is the maximum amount of recent downlinks stored per device.
	recentDownlinkCount = 20

	// fOptsCapacity is the maximum length of FOpts in bytes.
	fOptsCapacity = 15

	// infrastructureDelay represents a time interval Network Server uses as a buffer to account for infrastructure delay.
	infrastructureDelay = time.Second

	// peeringScheduleDelay is the schedule delay used for scheduling downlink via peering.
	// The schedule delay is used to estimate the transmission time, which is used as the minimum time for a subsequent transmission.
	//
	// When scheduling downlink to a cluster Gateway Server, the schedule delay is reported by the Gateway Server and is accurate.
	// When scheduling downlink via peering, the schedule delay is unknown, and should be sufficiently high to avoid conflicts.
	peeringScheduleDelay = infrastructureDelay + 4*time.Second

	// networkInitiatedDownlinkInterval is the minimum time.Duration passed before a network-initiated(e.g. Class B or C) downlink following an arbitrary downlink.
	networkInitiatedDownlinkInterval = time.Second
)

// windowDurationFunc is a function, which is used by Network Server to determine the duration of deduplication and cooldown windows.
type windowDurationFunc func(ctx context.Context) time.Duration

// makeWindowEndAfterFunc returns a windowDurationFunc, which always returns d.
func makeWindowDurationFunc(d time.Duration) windowDurationFunc {
	return func(ctx context.Context) time.Duration { return d }
}

// newDevAddrFunc is a function, which is used by Network Server to derive new DevAddrs.
type newDevAddrFunc func(ctx context.Context, dev *ttnpb.EndDevice) types.DevAddr

// makeNewDevAddrFunc returns a newDevAddrFunc, which derives DevAddrs using specified prefixes.
func makeNewDevAddrFunc(ps ...types.DevAddrPrefix) newDevAddrFunc {
	return func(ctx context.Context, dev *ttnpb.EndDevice) types.DevAddr {
		var devAddr types.DevAddr
		random.Read(devAddr[:])
		p := ps[random.Intn(len(ps))]
		return devAddr.WithPrefix(p)
	}
}

// DownlinkPriorities define the schedule priorities for the different types of downlink.
type DownlinkPriorities struct {
	// JoinAccept is the downlink priority for join-accept messages.
	JoinAccept,
	// MACCommands is the downlink priority for downlink messages with MAC commands as FRMPayload (FPort = 0) or as FOpts.
	// If the MAC commands are carried in FOpts, the highest priority of this value and the concerning application
	// downlink message's priority is used.
	MACCommands,
	// MaxApplicationDownlink is the highest priority permitted by the Network Server for application downlink.
	MaxApplicationDownlink ttnpb.TxSchedulePriority
}

// InteropClient is a client, which Network Server can use for interoperability.
type InteropClient interface {
	HandleJoinRequest(context.Context, types.NetID, *ttnpb.JoinRequest) (*ttnpb.JoinResponse, error)
}

// NetworkServer implements the Network Server component.
//
// The Network Server exposes the GsNs, AsNs, DeviceRegistry and ApplicationDownlinkQueue services.
type NetworkServer struct {
	*component.Component
	ctx context.Context

	devices DeviceRegistry

	netID      types.NetID
	newDevAddr newDevAddrFunc

	applicationServers *sync.Map // string -> *applicationUpStream
	applicationUplinks ApplicationUplinkQueue

	downlinkTasks      DownlinkTaskQueue
	downlinkPriorities DownlinkPriorities

	deduplicationWindow windowDurationFunc
	collectionWindow    windowDurationFunc

	defaultMACSettings ttnpb.MACSettings

	interopClient InteropClient

	uplinkDeduplicator UplinkDeduplicator

	deviceKEKLabel        string
	downlinkQueueCapacity int

	scheduledDownlinkMatcher ScheduledDownlinkMatcher
}

// Option configures the NetworkServer.
type Option func(ns *NetworkServer)

var (
	DefaultOptions []Option

	processTaskBackoff = &component.TaskBackoffConfig{
		Jitter:       component.DefaultTaskBackoffConfig.Jitter,
		IntervalFunc: component.MakeTaskBackoffIntervalFunc(true, component.DefaultTaskBackoffResetDuration, component.DefaultTaskBackoffIntervals[:]...),
	}
)

const (
	applicationUplinkProcessTaskName = "process_application_uplink"
	downlinkProcessTaskName          = "process_downlink"

	maxInt = int(^uint(0) >> 1)
)

// New returns new NetworkServer.
func New(c *component.Component, conf *Config, opts ...Option) (*NetworkServer, error) {
	ctx := log.NewContextWithField(c.Context(), "namespace", "networkserver")

	switch {
	case conf.DeduplicationWindow == 0:
		return nil, errInvalidConfiguration.WithCause(errors.New("DeduplicationWindow must be greater than 0"))
	case conf.CooldownWindow == 0:
		return nil, errInvalidConfiguration.WithCause(errors.New("CooldownWindow must be greater than 0"))
	case conf.Devices == nil:
		panic(errInvalidConfiguration.WithCause(errors.New("Devices is not specified")))
	case conf.DownlinkTasks == nil:
		panic(errInvalidConfiguration.WithCause(errors.New("DownlinkTasks is not specified")))
	case conf.UplinkDeduplicator == nil:
		panic(errInvalidConfiguration.WithCause(errors.New("UplinkDeduplicator is not specified")))
	case conf.ScheduledDownlinkMatcher == nil:
		panic(errInvalidConfiguration.WithCause(errors.New("ScheduledDownlinkMatcher is not specified")))
	case conf.DownlinkQueueCapacity < 0:
		return nil, errInvalidConfiguration.WithCause(errors.New("Downlink queue capacity must be greater than or equal to 0"))
	case conf.DownlinkQueueCapacity > maxInt/2:
		return nil, errInvalidConfiguration.WithCause(errors.New(fmt.Sprintf("Downlink queue capacity must be below %d", maxInt/2)))
	}

	devAddrPrefixes := conf.DevAddrPrefixes
	if len(devAddrPrefixes) == 0 {
		devAddr, err := types.NewDevAddr(conf.NetID, nil)
		if err != nil {
			return nil, err
		}
		devAddrPrefixes = []types.DevAddrPrefix{
			{
				DevAddr: devAddr,
				Length:  uint8(32 - types.NwkAddrBits(conf.NetID)),
			},
		}
	}
	downlinkPriorities, err := conf.DownlinkPriorities.Parse()
	if err != nil {
		return nil, err
	}

	var interopCl InteropClient
	if !conf.Interop.IsZero() {
		interopConf := conf.Interop
		interopConf.GetFallbackTLSConfig = func(ctx context.Context) (*tls.Config, error) {
			return c.GetTLSClientConfig(ctx)
		}
		interopConf.BlobConfig = c.GetBaseConfig(ctx).Blob
		if interopConf.HTTPClient == nil {
			httpClient, err := c.HTTPClient(ctx)
			if err != nil {
				return nil, err
			}
			interopConf.HTTPClient = httpClient
		}

		interopCl, err = interop.NewClient(ctx, interopConf)
		if err != nil {
			return nil, err
		}
	}

	ns := &NetworkServer{
		Component:                c,
		ctx:                      ctx,
		netID:                    conf.NetID,
		newDevAddr:               makeNewDevAddrFunc(devAddrPrefixes...),
		applicationServers:       &sync.Map{},
		applicationUplinks:       conf.ApplicationUplinkQueue.Queue,
		deduplicationWindow:      makeWindowDurationFunc(conf.DeduplicationWindow),
		collectionWindow:         makeWindowDurationFunc(conf.DeduplicationWindow + conf.CooldownWindow),
		devices:                  wrapEndDeviceRegistryWithReplacedFields(conf.Devices, replacedEndDeviceFields...),
		downlinkTasks:            conf.DownlinkTasks,
		downlinkPriorities:       downlinkPriorities,
		defaultMACSettings:       conf.DefaultMACSettings.Parse(),
		interopClient:            interopCl,
		uplinkDeduplicator:       conf.UplinkDeduplicator,
		deviceKEKLabel:           conf.DeviceKEKLabel,
		downlinkQueueCapacity:    conf.DownlinkQueueCapacity,
		scheduledDownlinkMatcher: conf.ScheduledDownlinkMatcher,
	}
	ctx = ns.Context()

	if len(opts) == 0 {
		opts = DefaultOptions
	}
	for _, opt := range opts {
		opt(ns)
	}

	hooks.RegisterUnaryHook("/ttn.lorawan.v3.GsNs", rpclog.NamespaceHook, rpclog.UnaryNamespaceHook("networkserver"))
	hooks.RegisterStreamHook("/ttn.lorawan.v3.AsNs", rpclog.NamespaceHook, rpclog.StreamNamespaceHook("networkserver"))
	hooks.RegisterUnaryHook("/ttn.lorawan.v3.AsNs", rpclog.NamespaceHook, rpclog.UnaryNamespaceHook("networkserver"))
	hooks.RegisterUnaryHook("/ttn.lorawan.v3.Ns", rpclog.NamespaceHook, rpclog.UnaryNamespaceHook("networkserver"))
	hooks.RegisterUnaryHook("/ttn.lorawan.v3.GsNs", cluster.HookName, c.ClusterAuthUnaryHook())
	hooks.RegisterStreamHook("/ttn.lorawan.v3.AsNs", cluster.HookName, c.ClusterAuthStreamHook())
	hooks.RegisterUnaryHook("/ttn.lorawan.v3.AsNs", cluster.HookName, c.ClusterAuthUnaryHook())
	hooks.RegisterUnaryHook("/ttn.lorawan.v3.Ns", cluster.HookName, c.ClusterAuthUnaryHook())

	for id, f := range map[string]func(context.Context) error{
		applicationUplinkProcessTaskName: ns.processApplicationUplinkTask,
		downlinkProcessTaskName:          ns.processDownlinkTask,
	} {
		ns.RegisterTask(&component.TaskConfig{
			Context: ctx,
			ID:      id,
			Func:    f,
			Restart: component.TaskRestartAlways,
			Backoff: processTaskBackoff,
		})
	}
	c.RegisterGRPC(ns)
	return ns, nil
}

// Context returns the context of the Network Server.
func (ns *NetworkServer) Context() context.Context {
	return ns.ctx
}

// RegisterServices registers services provided by ns at s.
func (ns *NetworkServer) RegisterServices(s *grpc.Server) {
	ttnpb.RegisterGsNsServer(s, ns)
	ttnpb.RegisterAsNsServer(s, ns)
	ttnpb.RegisterNsEndDeviceRegistryServer(s, ns)
	ttnpb.RegisterNsServer(s, ns)
}

// RegisterHandlers registers gRPC handlers.
func (ns *NetworkServer) RegisterHandlers(s *runtime.ServeMux, conn *grpc.ClientConn) {
	ttnpb.RegisterNsEndDeviceRegistryHandler(ns.Context(), s, conn)
	ttnpb.RegisterNsHandler(ns.Context(), s, conn)
}

// Roles returns the roles that the Network Server fulfills.
func (ns *NetworkServer) Roles() []ttnpb.ClusterRole {
	return []ttnpb.ClusterRole{ttnpb.ClusterRole_NETWORK_SERVER}
}
