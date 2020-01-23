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
	"hash/fnv"
	"sync"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"go.thethings.network/lorawan-stack/pkg/cluster"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/interop"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/random"
	"go.thethings.network/lorawan-stack/pkg/rpcmiddleware/hooks"
	"go.thethings.network/lorawan-stack/pkg/rpcmiddleware/rpclog"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"google.golang.org/grpc"
)

const (
	// recentDownlinkCount is the maximum amount of recent downlinks stored per device.
	recentDownlinkCount = 20

	// fOptsCapacity is the maximum length of FOpts in bytes.
	fOptsCapacity = 15
)

// windowEndFunc is a function, which is used by Network Server to determine the end of deduplication and cooldown windows.
type windowEndFunc func(ctx context.Context, up *ttnpb.UplinkMessage) <-chan time.Time

// makeWindowEndAfterFunc returns a windowEndFunc, which closes
// the returned channel after at least duration d after up.ServerTime or if the context is done.
func makeWindowEndAfterFunc(d time.Duration) windowEndFunc {
	return func(ctx context.Context, up *ttnpb.UplinkMessage) <-chan time.Time {
		ch := make(chan time.Time, 1)

		now := timeNow()
		if up.ReceivedAt.IsZero() {
			up.ReceivedAt = now
		}

		end := up.ReceivedAt.Add(d)
		if end.Before(now) {
			ch <- end
			return ch
		}

		go func() {
			time.Sleep(timeUntil(up.ReceivedAt.Add(d)))
			ch <- end
		}()
		return ch
	}
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

	metadataAccumulators *sync.Map // uint64 -> *metadataAccumulator

	metadataAccumulatorPool *sync.Pool
	hashPool                *sync.Pool

	downlinkTasks      DownlinkTaskQueue
	downlinkPriorities DownlinkPriorities

	deduplicationDone windowEndFunc
	collectionDone    windowEndFunc

	defaultMACSettings ttnpb.MACSettings

	interopClient InteropClient

	deviceKEKLabel string
}

// Option configures the NetworkServer.
type Option func(ns *NetworkServer)

// New returns new NetworkServer.
func New(c *component.Component, conf *Config, opts ...Option) (*NetworkServer, error) {
	switch {
	case conf.DeduplicationWindow == 0:
		return nil, errInvalidConfiguration.WithCause(errors.New("DeduplicationWindow must be greater than 0"))
	case conf.CooldownWindow == 0:
		return nil, errInvalidConfiguration.WithCause(errors.New("CooldownWindow must be greater than 0"))
	case conf.DownlinkTasks == nil:
		return nil, errInvalidConfiguration.WithCause(errors.New("DownlinkTasks is not specified"))
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

	ctx := log.NewContextWithField(c.Context(), "namespace", "networkserver")

	var interopCl InteropClient
	if !conf.Interop.IsZero() {
		interopConf := conf.Interop
		interopConf.GetFallbackTLSConfig = func(ctx context.Context) (*tls.Config, error) {
			return c.GetTLSClientConfig(ctx)
		}
		interopConf.BlobConfig = c.GetBaseConfig(ctx).Blob

		interopCl, err = interop.NewClient(ctx, interopConf)
		if err != nil {
			return nil, err
		}
	}

	ns := &NetworkServer{
		Component:            c,
		ctx:                  ctx,
		netID:                conf.NetID,
		newDevAddr:           makeNewDevAddrFunc(devAddrPrefixes...),
		applicationServers:   &sync.Map{},
		applicationUplinks:   conf.ApplicationUplinks,
		deduplicationDone:    makeWindowEndAfterFunc(conf.DeduplicationWindow),
		collectionDone:       makeWindowEndAfterFunc(conf.DeduplicationWindow + conf.CooldownWindow),
		devices:              conf.Devices,
		downlinkTasks:        conf.DownlinkTasks,
		metadataAccumulators: &sync.Map{},
		metadataAccumulatorPool: &sync.Pool{
			New: func() interface{} {
				return &metadataAccumulator{}
			},
		},
		hashPool: &sync.Pool{
			New: func() interface{} {
				return fnv.New64a()
			},
		},
		downlinkPriorities: downlinkPriorities,
		defaultMACSettings: ttnpb.MACSettings{
			ClassBTimeout:         conf.DefaultMACSettings.ClassBTimeout,
			ClassCTimeout:         conf.DefaultMACSettings.ClassCTimeout,
			StatusTimePeriodicity: conf.DefaultMACSettings.StatusTimePeriodicity,
		},
		interopClient:  interopCl,
		deviceKEKLabel: conf.DeviceKEKLabel,
	}
	if conf.DefaultMACSettings.ADRMargin != nil {
		ns.defaultMACSettings.ADRMargin = &pbtypes.FloatValue{Value: *conf.DefaultMACSettings.ADRMargin}
	}
	if conf.DefaultMACSettings.DesiredRx1Delay != nil {
		ns.defaultMACSettings.DesiredRx1Delay = &ttnpb.RxDelayValue{Value: *conf.DefaultMACSettings.DesiredRx1Delay}
	}
	if conf.DefaultMACSettings.DesiredMaxDutyCycle != nil {
		ns.defaultMACSettings.DesiredMaxDutyCycle = &ttnpb.AggregatedDutyCycleValue{Value: *conf.DefaultMACSettings.DesiredMaxDutyCycle}
	}
	if conf.DefaultMACSettings.DesiredADRAckLimitExponent != nil {
		ns.defaultMACSettings.DesiredADRAckLimitExponent = &ttnpb.ADRAckLimitExponentValue{Value: *conf.DefaultMACSettings.DesiredADRAckLimitExponent}
	}
	if conf.DefaultMACSettings.DesiredADRAckDelayExponent != nil {
		ns.defaultMACSettings.DesiredADRAckDelayExponent = &ttnpb.ADRAckDelayExponentValue{Value: *conf.DefaultMACSettings.DesiredADRAckDelayExponent}
	}
	if conf.DefaultMACSettings.StatusCountPeriodicity != nil {
		ns.defaultMACSettings.StatusCountPeriodicity = &pbtypes.UInt32Value{Value: *conf.DefaultMACSettings.StatusCountPeriodicity}
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

	ns.RegisterTask(ns.Context(), "process_downlink", func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			if err := ns.processDownlinkTask(ctx); err != nil {
				return err
			}
		}
	}, component.TaskRestartOnFailure)

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
