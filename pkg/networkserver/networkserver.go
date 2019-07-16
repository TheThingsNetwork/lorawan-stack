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
	"hash/fnv"
	"io"
	"sync"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"go.thethings.network/lorawan-stack/pkg/cluster"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/rpcmiddleware/hooks"
	"go.thethings.network/lorawan-stack/pkg/rpcmiddleware/rpclog"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/unique"
	"google.golang.org/grpc"
)

const (
	// recentDownlinkCount is the maximum amount of recent downlinks stored per device.
	recentDownlinkCount = 20

	// fOptsCapacity is the maximum length of FOpts in bytes.
	fOptsCapacity = 15
)

// WindowEndFunc is a function, which is used by Network Server to determine the end of deduplication and cooldown windows.
type WindowEndFunc func(ctx context.Context, up *ttnpb.UplinkMessage) <-chan time.Time

// NewWindowEndAfterFunc returns a WindowEndFunc, which closes
// the returned channel after at least duration d after up.ServerTime or if the context is done.
func NewWindowEndAfterFunc(d time.Duration) WindowEndFunc {
	return func(ctx context.Context, up *ttnpb.UplinkMessage) <-chan time.Time {
		ch := make(chan time.Time, 1)

		if up.ReceivedAt.IsZero() {
			up.ReceivedAt = time.Now()
		}

		end := up.ReceivedAt.Add(d)
		if end.Before(time.Now()) {
			ch <- end
			return ch
		}

		go func() {
			time.Sleep(time.Until(up.ReceivedAt.Add(d)))
			ch <- end
		}()
		return ch
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

	netID           types.NetID
	devAddrPrefixes []types.DevAddrPrefix

	applicationServers *sync.Map // string -> *applicationUpStream

	metadataAccumulators *sync.Map // uint64 -> *metadataAccumulator

	metadataAccumulatorPool *sync.Pool
	hashPool                *sync.Pool

	downlinkTasks      DownlinkTaskQueue
	downlinkPriorities DownlinkPriorities

	deduplicationDone WindowEndFunc
	collectionDone    WindowEndFunc

	handleASUplink func(ctx context.Context, ids ttnpb.ApplicationIdentifiers, up *ttnpb.ApplicationUp) (bool, error)

	defaultMACSettings ttnpb.MACSettings

	interopClient InteropClient
}

// Option configures the NetworkServer.
type Option func(ns *NetworkServer)

// WithDeduplicationDoneFunc overrides the default WindowEndFunc, which
// is used to determine the end of uplink metadata deduplication.
func WithDeduplicationDoneFunc(f WindowEndFunc) Option {
	return func(ns *NetworkServer) {
		ns.deduplicationDone = f
	}
}

// WithCollectionDoneFunc overrides the default WindowEndFunc, which
// is used to determine the end of uplink duplicate collection.
func WithCollectionDoneFunc(f WindowEndFunc) Option {
	return func(ns *NetworkServer) {
		ns.collectionDone = f
	}
}

// WithASUplinkHandler overrides the default function called, which is used for sending the uplink to AS.
func WithASUplinkHandler(f func(context.Context, ttnpb.ApplicationIdentifiers, *ttnpb.ApplicationUp) (bool, error)) Option {
	return func(ns *NetworkServer) {
		ns.handleASUplink = f
	}
}

// New returns new NetworkServer.
func New(c *component.Component, conf *Config, opts ...Option) (*NetworkServer, error) {
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
	ns := &NetworkServer{
		Component:               c,
		ctx:                     log.NewContextWithField(c.Context(), "namespace", "networkserver"),
		devices:                 conf.Devices,
		netID:                   conf.NetID,
		devAddrPrefixes:         devAddrPrefixes,
		applicationServers:      &sync.Map{},
		metadataAccumulators:    &sync.Map{},
		metadataAccumulatorPool: &sync.Pool{},
		hashPool:                &sync.Pool{},
		downlinkTasks:           conf.DownlinkTasks,
		downlinkPriorities:      downlinkPriorities,
		defaultMACSettings: ttnpb.MACSettings{
			ClassBTimeout:         conf.DefaultMACSettings.ClassBTimeout,
			ClassCTimeout:         conf.DefaultMACSettings.ClassCTimeout,
			StatusTimePeriodicity: conf.DefaultMACSettings.StatusTimePeriodicity,
		},
		interopClient: conf.InteropClient,
	}
	ns.hashPool.New = func() interface{} {
		return fnv.New64a()
	}
	ns.metadataAccumulatorPool.New = func() interface{} {
		return &metadataAccumulator{}
	}

	if conf.DefaultMACSettings.ADRMargin != nil {
		ns.defaultMACSettings.ADRMargin = &pbtypes.FloatValue{Value: *conf.DefaultMACSettings.ADRMargin}
	}
	if conf.DefaultMACSettings.DesiredRx1Delay != nil {
		ns.defaultMACSettings.DesiredRx1Delay = &ttnpb.MACSettings_RxDelayValue{Value: *conf.DefaultMACSettings.DesiredRx1Delay}
	}
	if conf.DefaultMACSettings.StatusCountPeriodicity != nil {
		ns.defaultMACSettings.StatusCountPeriodicity = &pbtypes.UInt32Value{Value: *conf.DefaultMACSettings.StatusCountPeriodicity}
	}

	for _, opt := range opts {
		opt(ns)
	}

	switch {
	case ns.deduplicationDone == nil && conf.DeduplicationWindow == 0:
		return nil, errInvalidConfiguration.WithCause(errors.New("DeduplicationWindow is zero and WithDeduplicationDoneFunc not specified"))
	case ns.collectionDone == nil && conf.DeduplicationWindow == 0:
		return nil, errInvalidConfiguration.WithCause(errors.New("DeduplicationWindow is zero and WithCollectionDoneFunc not specified"))
	case ns.collectionDone == nil && conf.CooldownWindow == 0:
		return nil, errInvalidConfiguration.WithCause(errors.New("CooldownWindow is zero and WithCollectionDoneFunc not specified"))
	}

	if ns.downlinkTasks == nil {
		return nil, errInvalidConfiguration.WithCause(errors.New("DownlinkTasks is not specified"))
	}

	if ns.deduplicationDone == nil {
		ns.deduplicationDone = NewWindowEndAfterFunc(conf.DeduplicationWindow)
	}
	if ns.collectionDone == nil {
		ns.collectionDone = NewWindowEndAfterFunc(conf.DeduplicationWindow + conf.CooldownWindow)
	}

	if ns.handleASUplink == nil {
		ns.handleASUplink = func(ctx context.Context, ids ttnpb.ApplicationIdentifiers, up *ttnpb.ApplicationUp) (bool, error) {
			v, ok := ns.applicationServers.Load(unique.ID(ctx, ids))
			if !ok {
				return false, nil
			}
			as := v.(ttnpb.AsNs_LinkApplicationServer)

			var err error
			if err = as.Send(up); err != nil {
				return true, err
			}
			if _, err = as.Recv(); err != nil {
				return true, err
			}
			return true, nil
		}
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
func (ns *NetworkServer) Roles() []ttnpb.PeerInfo_Role {
	return []ttnpb.PeerInfo_Role{ttnpb.PeerInfo_NETWORK_SERVER}
}

func (ns *NetworkServer) Close() {
	ns.Component.Close()

	logger := ns.Logger()
	ns.applicationServers.Range(func(k interface{}, v interface{}) bool {
		logger := logger.WithField("application_uid", k.(string))
		logger.Debug("Close Application Server link")
		if err := v.(io.Closer).Close(); err != nil {
			logger.WithError(err).Warn("Failed to close AS link")
		}
		logger.Debug("Application Server link closed")
		return true
	})
}
