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
	"crypto/rand"
	"fmt"
	"os"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/interop"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal/time"
	"go.thethings.network/lorawan-stack/v3/pkg/random"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmiddleware/hooks"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmiddleware/rpclog"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmiddleware/rpctracer"
	"go.thethings.network/lorawan-stack/v3/pkg/task"
	"go.thethings.network/lorawan-stack/v3/pkg/telemetry/tracing/tracer"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/workerpool"
	"google.golang.org/grpc"
)

const (
	// recentDownlinkCount is the maximum amount of recent downlinks stored per device.
	recentDownlinkCount = 5

	// fOptsCapacity is the maximum length of FOpts in bytes.
	fOptsCapacity = 15

	// infrastructureDelay represents a time interval that the Network Server uses as a buffer to account for infrastructure delay.
	infrastructureDelay = time.Second

	// absoluteTimeSchedulingDelay represents a time interval that the Network Server uses as a buffer to account for the transmission
	// time while scheduling absolute time downlinks, since absolute time scheduling considers the absolute time to be the timestamp
	// for the arrival, not start, of the transmission.
	absoluteTimeSchedulingDelay = 5 * time.Second

	// peeringScheduleDelay is the schedule delay used for scheduling downlink via peering.
	// The schedule delay is used to estimate the transmission time, which is used as the minimum time for a subsequent transmission.
	//
	// When scheduling downlink to a cluster Gateway Server, the schedule delay is reported by the Gateway Server and is accurate.
	// When scheduling downlink via peering, the schedule delay is unknown, and should be sufficiently high to avoid conflicts.
	peeringScheduleDelay = infrastructureDelay + 4*time.Second
)

// windowDurationFunc is a function, which is used by Network Server to determine the duration of deduplication and cooldown windows.
type windowDurationFunc func(ctx context.Context) time.Duration

// makeWindowEndAfterFunc returns a windowDurationFunc, which always returns d.
func makeWindowDurationFunc(d time.Duration) windowDurationFunc {
	return func(ctx context.Context) time.Duration { return d }
}

// netIDFunc is a function, which is used by the Network Server to determine its NetID.
type netIDFunc func(ctx context.Context) types.NetID

// makeNetIDFunc returns a netIDFunc, which always returns netID.
func makeNetIDFunc(netID types.NetID) netIDFunc {
	return func(ctx context.Context) types.NetID {
		return netID
	}
}

// nsIDFunc is a function, which is used by the Network Server to determine its NSID.
type nsIDFunc func(ctx context.Context) *types.EUI64

// makeNSIDFunc returns a nsIDFunc, which always returns nsID.
func makeNSIDFunc(nsID *types.EUI64) nsIDFunc {
	return func(ctx context.Context) *types.EUI64 {
		return nsID
	}
}

// newDevAddrFunc is a function, which is used by Network Server to derive new DevAddrs.
type newDevAddrFunc func(ctx context.Context) types.DevAddr

// makeNewDevAddrFunc returns a newDevAddrFunc, which derives DevAddrs using specified prefixes.
func makeNewDevAddrFunc(ps ...types.DevAddrPrefix) newDevAddrFunc {
	weights := make([]int64, len(ps))
	totalWeight := int64(0)
	for i, p := range ps {
		weights[i] = int64(1 << (32 - p.Length))
		totalWeight += weights[i]
	}
	return func(ctx context.Context) types.DevAddr {
		var devAddr types.DevAddr
		_, _ = rand.Read(devAddr[:])
		r := random.Int63n(totalWeight)
		for i, weight := range weights {
			r -= weight
			if r < 0 {
				return devAddr.WithPrefix(ps[i])
			}
		}
		panic("unreachable")
	}
}

// devAddrPrefixesFunc is a function, which is used by the Network Server to list it's DevAddrPrefixes.
type devAddrPrefixesFunc func(ctx context.Context) []types.DevAddrPrefix

// makeDevAddrPrefixesFunc returns a devAddrPrefixesFunc, which always returns ps.
func makeDevAddrPrefixesFunc(ps ...types.DevAddrPrefix) devAddrPrefixesFunc {
	return func(ctx context.Context) []types.DevAddrPrefix {
		return ps
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
	HandleJoinRequest(
		ctx context.Context, netID types.NetID, nsID *types.EUI64, req *ttnpb.JoinRequest,
	) (*ttnpb.JoinResponse, error)
}

// NetworkServer implements the Network Server component.
//
// The Network Server exposes the GsNs, AsNs, DeviceRegistry and ApplicationDownlinkQueue services.
type NetworkServer struct {
	ttnpb.UnimplementedAsNsServer
	ttnpb.UnimplementedGsNsServer
	ttnpb.UnimplementedNsEndDeviceRegistryServer
	ttnpb.UnimplementedNsServer

	*component.Component
	ctx context.Context

	devices DeviceRegistry

	batchDevices       ttnpb.NsEndDeviceBatchRegistryServer
	relayConfiguration ttnpb.NsRelayConfigurationServiceServer
	macSettingsProfile ttnpb.NsMACSettingsProfileRegistryServer

	netID           netIDFunc
	nsID            nsIDFunc
	clusterID       string
	newDevAddr      newDevAddrFunc
	devAddrPrefixes devAddrPrefixesFunc

	applicationUplinks ApplicationUplinkQueue

	downlinkTasks      DownlinkTaskQueue
	downlinkPriorities DownlinkPriorities

	deduplicationWindow windowDurationFunc
	collectionWindow    windowDurationFunc

	defaultMACSettings *ttnpb.MACSettings

	interopClient InteropClient

	uplinkDeduplicator UplinkDeduplicator

	deviceKEKLabel        string
	downlinkQueueCapacity int

	scheduledDownlinkMatcher ScheduledDownlinkMatcher

	uplinkSubmissionPool workerpool.WorkerPool[[]*ttnpb.ApplicationUp]
}

// Option configures the NetworkServer.
type Option func(ns *NetworkServer)

var (
	DefaultOptions []Option

	processTaskBackoff = &task.BackoffConfig{
		Jitter:       task.DefaultBackoffConfig.Jitter,
		IntervalFunc: task.MakeBackoffIntervalFunc(true, task.DefaultBackoffResetDuration, task.DefaultBackoffIntervals[:]...),
	}
)

const (
	applicationUplinkProcessTaskName  = "process_application_uplink"
	downlinkProcessTaskName           = "process_downlink"
	applicationUplinkDispatchTaskName = "dispatch_application_uplink"
	downlinkDispatchTaskName          = "dispatch_downlink"

	maxInt = int(^uint(0) >> 1)
)

// New returns new NetworkServer.
func New(c *component.Component, conf *Config, opts ...Option) (*NetworkServer, error) {
	ctx := tracer.NewContextWithTracer(c.Context(), tracerNamespace)

	ctx = log.NewContextWithField(ctx, "namespace", logNamespace)

	switch {
	case conf.DeduplicationWindow == 0:
		return nil, errInvalidConfiguration.WithCause(errors.New("DeduplicationWindow must be greater than 0"))
	case conf.CooldownWindow == 0:
		return nil, errInvalidConfiguration.WithCause(errors.New("CooldownWindow must be greater than 0"))
	case conf.Devices == nil:
		panic(errInvalidConfiguration.WithCause(errors.New("Devices is not specified")))
	case conf.DownlinkTaskQueue.NumConsumers == 0:
		return nil, errInvalidConfiguration.WithCause(errors.New("DownlinkTaskQueue.NumConsumers must be greater than 0"))
	case conf.ApplicationUplinkQueue.NumConsumers == 0:
		return nil, errInvalidConfiguration.WithCause(errors.New("ApplicationUplinkQueue.NumConsumers must be greater than 0"))
	case conf.DownlinkTaskQueue.Queue == nil:
		panic(errInvalidConfiguration.WithCause(errors.New("DownlinkTaskQueue is not specified")))
	case conf.UplinkDeduplicator == nil:
		panic(errInvalidConfiguration.WithCause(errors.New("UplinkDeduplicator is not specified")))
	case conf.ScheduledDownlinkMatcher == nil:
		panic(errInvalidConfiguration.WithCause(errors.New("ScheduledDownlinkMatcher is not specified")))
	case conf.DownlinkQueueCapacity < 0:
		return nil, errInvalidConfiguration.WithCause(errors.New("Downlink queue capacity must be greater than or equal to 0"))
	case conf.DownlinkQueueCapacity > maxInt/2:
		return nil, errInvalidConfiguration.WithCause(errors.New(fmt.Sprintf("Downlink queue capacity must be below %d", maxInt/2)))
	case conf.MACSettingsProfileRegistry == nil:
		panic(errInvalidConfiguration.WithCause(errors.New("MACSettingsProfileRegistry is not specified")))
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
		interopConf := conf.Interop.InteropClient
		interopConf.BlobConfig = c.GetBaseConfig(ctx).Blob

		interopCl, err = interop.NewClient(ctx, interopConf, c, interop.SelectorNetworkServer)
		if err != nil {
			return nil, err
		}
	}

	defaultMACSettings, err := conf.DefaultMACSettings.Parse()
	if err != nil {
		return nil, err
	}

	ns := &NetworkServer{
		Component:                c,
		ctx:                      ctx,
		netID:                    makeNetIDFunc(conf.NetID),
		nsID:                     makeNSIDFunc(conf.Interop.ID),
		clusterID:                conf.ClusterID,
		newDevAddr:               makeNewDevAddrFunc(devAddrPrefixes...),
		devAddrPrefixes:          makeDevAddrPrefixesFunc(devAddrPrefixes...),
		applicationUplinks:       conf.ApplicationUplinkQueue.Queue,
		deduplicationWindow:      makeWindowDurationFunc(conf.DeduplicationWindow),
		collectionWindow:         makeWindowDurationFunc(conf.DeduplicationWindow + conf.CooldownWindow),
		devices:                  wrapEndDeviceRegistryWithReplacedFields(conf.Devices, replacedEndDeviceFields...),
		batchDevices:             &nsEndDeviceBatchRegistry{devices: conf.Devices},
		relayConfiguration:       &nsRelayConfigurationService{devices: conf.Devices, frequencyPlans: c.FrequencyPlansStore},
		macSettingsProfile:       &NsMACSettingsProfileRegistry{registry: conf.MACSettingsProfileRegistry},
		downlinkTasks:            conf.DownlinkTaskQueue.Queue,
		downlinkPriorities:       downlinkPriorities,
		defaultMACSettings:       defaultMACSettings,
		interopClient:            interopCl,
		uplinkDeduplicator:       conf.UplinkDeduplicator,
		deviceKEKLabel:           conf.DeviceKEKLabel,
		downlinkQueueCapacity:    conf.DownlinkQueueCapacity,
		scheduledDownlinkMatcher: conf.ScheduledDownlinkMatcher,
	}
	ns.uplinkSubmissionPool = workerpool.NewWorkerPool(workerpool.Config[[]*ttnpb.ApplicationUp]{
		Component:  c,
		Context:    ctx,
		Name:       "uplink_submission",
		Handler:    ns.handleUplinkSubmission,
		QueueSize:  int(conf.ApplicationUplinkQueue.FastBufferSize),
		MaxWorkers: int(conf.ApplicationUplinkQueue.FastNumConsumers),
	})
	ctx = ns.Context()

	if len(opts) == 0 {
		opts = DefaultOptions
	}
	for _, opt := range opts {
		opt(ns)
	}

	for _, hook := range []struct {
		name       string
		middleware hooks.UnaryHandlerMiddleware
	}{
		{rpctracer.TracerHook, rpctracer.UnaryTracerHook(tracerNamespace)},
		{rpclog.NamespaceHook, rpclog.UnaryNamespaceHook(logNamespace)},
	} {
		for _, filter := range []string{
			"/ttn.lorawan.v3.GsNs",
			"/ttn.lorawan.v3.AsNs",
			"/ttn.lorawan.v3.NsEndDeviceRegistry",
			"/ttn.lorawan.v3.NsEndDeviceBatchRegistry",
			"/ttn.lorawan.v3.Ns",
			"/ttn.lorawan.v3.RelayConfigurationService",
			"/ttn.lorawan.v3.NsMACSettingsProfileRegistry",
		} {
			c.GRPC.RegisterUnaryHook(filter, hook.name, hook.middleware)
		}
	}
	c.GRPC.RegisterUnaryHook("/ttn.lorawan.v3.GsNs", cluster.HookName, c.ClusterAuthUnaryHook())
	c.GRPC.RegisterUnaryHook("/ttn.lorawan.v3.AsNs", cluster.HookName, c.ClusterAuthUnaryHook())
	c.GRPC.RegisterUnaryHook("/ttn.lorawan.v3.Ns", cluster.HookName, c.ClusterAuthUnaryHook())

	c.GRPC.RegisterStreamHook("/ttn.lorawan.v3.AsNs", rpctracer.TracerHook, rpctracer.StreamTracerHook(tracerNamespace))
	c.GRPC.RegisterStreamHook("/ttn.lorawan.v3.AsNs", rpclog.NamespaceHook, rpclog.StreamNamespaceHook(logNamespace))
	c.GRPC.RegisterStreamHook("/ttn.lorawan.v3.AsNs", cluster.HookName, c.ClusterAuthStreamHook())

	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	consumerIDPrefix := fmt.Sprintf("%s:%d", hostname, os.Getpid())
	for id, dispatcher := range map[string]interface {
		Dispatch(context.Context, string) error
	}{
		downlinkDispatchTaskName: ns.downlinkTasks,
	} {
		dispatcher := dispatcher
		ns.RegisterTask(&task.Config{
			Context: ctx,
			ID:      id,
			Func: func(ctx context.Context) error {
				return dispatcher.Dispatch(ctx, consumerIDPrefix)
			},
			Restart: task.RestartAlways,
			Backoff: processTaskBackoff,
		})
	}
	for i := uint64(0); i < conf.ApplicationUplinkQueue.NumConsumers; i++ {
		consumerID := fmt.Sprintf("%s:%d", consumerIDPrefix, i)
		ns.RegisterTask(&task.Config{
			Context: ctx,
			ID:      fmt.Sprintf("%s_%d", applicationUplinkProcessTaskName, i),
			Func:    ns.createProcessApplicationUplinkTask(consumerID),
			Restart: task.RestartAlways,
			Backoff: processTaskBackoff,
		})
	}
	for i := uint64(0); i < conf.DownlinkTaskQueue.NumConsumers; i++ {
		consumerID := fmt.Sprintf("%s:%d", consumerIDPrefix, i)
		ns.RegisterTask(&task.Config{
			Context: ctx,
			ID:      fmt.Sprintf("%s_%d", downlinkProcessTaskName, i),
			Func:    ns.createProcessDownlinkTask(consumerID),
			Restart: task.RestartAlways,
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
	ttnpb.RegisterNsEndDeviceBatchRegistryServer(s, ns.batchDevices)
	ttnpb.RegisterNsServer(s, ns)
	ttnpb.RegisterNsRelayConfigurationServiceServer(s, ns.relayConfiguration)
	ttnpb.RegisterNsMACSettingsProfileRegistryServer(s, ns.macSettingsProfile)
}

// RegisterHandlers registers gRPC handlers.
func (ns *NetworkServer) RegisterHandlers(s *runtime.ServeMux, conn *grpc.ClientConn) {
	ttnpb.RegisterNsEndDeviceRegistryHandler(ns.Context(), s, conn)
	ttnpb.RegisterNsEndDeviceBatchRegistryHandler(ns.Context(), s, conn) // nolint:errcheck
	ttnpb.RegisterNsHandler(ns.Context(), s, conn)
	ttnpb.RegisterNsRelayConfigurationServiceHandler(ns.Context(), s, conn)  // nolint:errcheck
	ttnpb.RegisterNsMACSettingsProfileRegistryHandler(ns.Context(), s, conn) // nolint:errcheck
}

// Roles returns the roles that the Network Server fulfills.
func (ns *NetworkServer) Roles() []ttnpb.ClusterRole {
	return []ttnpb.ClusterRole{ttnpb.ClusterRole_NETWORK_SERVER}
}
