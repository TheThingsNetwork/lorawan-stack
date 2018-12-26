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

// Package networkserver provides a LoRaWAN 1.1-compliant Network Server implementation.
package networkserver

import (
	"context"
	"fmt"
	"hash/fnv"
	"sync"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"go.thethings.network/lorawan-stack/pkg/cluster"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/rpcmiddleware/hooks"
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

	// classCTimeout represents the time interval, within which class C
	// device should acknowledge the downlink or answer MAC command.
	classCTimeout = 5 * time.Minute
)

func timePtr(t time.Time) *time.Time {
	return &t
}

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

// PeerGetter is the interface, which wraps GetPeer method.
type PeerGetter interface {
	GetPeer(ctx context.Context, role ttnpb.PeerInfo_Role, ids ttnpb.Identifiers) cluster.Peer
}

// NsGsClientFunc is the function used to get Gateway Server.
type NsGsClientFunc func(ctx context.Context, id ttnpb.GatewayIdentifiers) (ttnpb.NsGsClient, error)

// NewGatewayServerPeerGetterFunc returns a NsGsClientFunc, which uses g to retrieve Gateway Server clients.
func NewGatewayServerPeerGetterFunc(g PeerGetter) NsGsClientFunc {
	return func(ctx context.Context, ids ttnpb.GatewayIdentifiers) (ttnpb.NsGsClient, error) {
		p := g.GetPeer(ctx, ttnpb.PeerInfo_GATEWAY_SERVER, ids)
		if p == nil {
			return nil, errGatewayServerNotFound
		}
		return ttnpb.NewNsGsClient(p.Conn()), nil
	}
}

// NsJsClientFunc is the function used to get Join Server.
type NsJsClientFunc func(ctx context.Context, id ttnpb.EndDeviceIdentifiers) (ttnpb.NsJsClient, error)

// NewJoinServerPeerGetterFunc returns a NsJsClientFunc, which uses g to retrieve Join Server clients.
func NewJoinServerPeerGetterFunc(g PeerGetter) NsJsClientFunc {
	return func(ctx context.Context, ids ttnpb.EndDeviceIdentifiers) (ttnpb.NsJsClient, error) {
		p := g.GetPeer(ctx, ttnpb.PeerInfo_JOIN_SERVER, ids)
		if p == nil {
			return nil, errJoinServerNotFound
		}
		return ttnpb.NewNsJsClient(p.Conn()), nil
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

// NetworkServer implements the Network Server component.
//
// The Network Server exposes the GsNs, AsNs, DeviceRegistry and ApplicationDownlinkQueue services.
type NetworkServer struct {
	*component.Component

	devices DeviceRegistry

	NetID types.NetID

	applicationServersMu *sync.RWMutex
	applicationServers   map[string]*applicationUpStream

	metadataAccumulators *sync.Map // uint64 -> *metadataAccumulator

	metadataAccumulatorPool *sync.Pool
	hashPool                *sync.Pool

	macHandlers        *sync.Map // ttnpb.MACCommandIdentifier -> MACHandler
	downlinkTasks      DownlinkTaskQueue
	downlinkPriorities DownlinkPriorities

	deduplicationDone WindowEndFunc
	collectionDone    WindowEndFunc

	gsClient NsGsClientFunc
	jsClient NsJsClientFunc

	handleASUplink func(ctx context.Context, ids ttnpb.ApplicationIdentifiers, up *ttnpb.ApplicationUp) (bool, error)
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

// WithNsGsClientFunc overrides the default NsGsClientFunc, which
// is used to get the Gateway Server by gateway identifiers.
func WithNsGsClientFunc(f NsGsClientFunc) Option {
	return func(ns *NetworkServer) {
		ns.gsClient = f
	}
}

// WithNsJsClientFunc overrides the default NsJsClientFunc, which
// is used to get the Join Server by end device identifiers.
func WithNsJsClientFunc(f NsJsClientFunc) Option {
	return func(ns *NetworkServer) {
		ns.jsClient = f
	}
}

// WithASUplinkHandler overrides the default function called, for sending the uplink to AS.
func WithASUplinkHandler(f func(context.Context, ttnpb.ApplicationIdentifiers, *ttnpb.ApplicationUp) (bool, error)) Option {
	return func(ns *NetworkServer) {
		ns.handleASUplink = f
	}
}

// WithMACHandler registers a MACHandler for specified CID in 0x80-0xFF range.
// WithMACHandler panics if a MACHandler for the CID is already registered, or if
// the CID is out of range.
func WithMACHandler(cid ttnpb.MACCommandIdentifier, fn MACHandler) Option {
	if cid < 0x80 || cid > 0xFF {
		panic(errCIDOutOfRange.WithAttributes(
			"min", fmt.Sprintf("0x%X", 0x80),
			"max", fmt.Sprintf("0x%X", 0xFF),
		))
	}

	return func(ns *NetworkServer) {
		_, ok := ns.macHandlers.LoadOrStore(cid, fn)
		if ok {
			panic(errDuplicateCIDHandler.WithAttributes(
				"cid", fmt.Sprintf("0x%X", int32(cid)),
			))
		}
	}
}

// New returns new NetworkServer.
func New(c *component.Component, conf *Config, opts ...Option) (*NetworkServer, error) {
	downlinkPriorities, err := conf.DownlinkPriorities.Parse()
	if err != nil {
		return nil, err
	}
	ns := &NetworkServer{
		Component:               c,
		devices:                 conf.Devices,
		downlinkTasks:           conf.DownlinkTasks,
		downlinkPriorities:      downlinkPriorities,
		applicationServersMu:    &sync.RWMutex{},
		applicationServers:      make(map[string]*applicationUpStream),
		metadataAccumulators:    &sync.Map{},
		metadataAccumulatorPool: &sync.Pool{},
		hashPool:                &sync.Pool{},
		macHandlers:             &sync.Map{},
	}
	ns.hashPool.New = func() interface{} {
		return fnv.New64a()
	}
	ns.metadataAccumulatorPool.New = func() interface{} {
		return &metadataAccumulator{}
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

	if ns.gsClient == nil {
		ns.gsClient = NewGatewayServerPeerGetterFunc(ns.Component)
	}
	if ns.jsClient == nil {
		ns.jsClient = NewJoinServerPeerGetterFunc(ns.Component)
	}

	if ns.handleASUplink == nil {
		ns.handleASUplink = func(ctx context.Context, ids ttnpb.ApplicationIdentifiers, up *ttnpb.ApplicationUp) (ok bool, err error) {
			ns.applicationServersMu.RLock()
			as, ok := ns.applicationServers[unique.ID(ctx, ids)]
			if ok {
				err = as.Send(up)
			}
			ns.applicationServersMu.RUnlock()
			return ok, err
		}
	}

	hooks.RegisterUnaryHook("/ttn.lorawan.v3.GsNs", cluster.HookName, c.ClusterAuthUnaryHook())
	hooks.RegisterStreamHook("/ttn.lorawan.v3.AsNs", cluster.HookName, c.ClusterAuthStreamHook())
	hooks.RegisterUnaryHook("/ttn.lorawan.v3.AsApplicationDownlinkQueue", cluster.HookName, c.ClusterAuthUnaryHook())

	ns.Component.RegisterTask(func(ctx context.Context) error {
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

// RegisterServices registers services provided by ns at s.
func (ns *NetworkServer) RegisterServices(s *grpc.Server) {
	ttnpb.RegisterGsNsServer(s, ns)
	ttnpb.RegisterAsNsServer(s, ns)
	ttnpb.RegisterNsEndDeviceRegistryServer(s, ns)
}

// RegisterHandlers registers gRPC handlers.
func (ns *NetworkServer) RegisterHandlers(s *runtime.ServeMux, conn *grpc.ClientConn) {
	ttnpb.RegisterNsEndDeviceRegistryHandler(ns.Context(), s, conn)
}

// Roles returns the roles that the Network Server fulfills.
func (ns *NetworkServer) Roles() []ttnpb.PeerInfo_Role {
	return []ttnpb.PeerInfo_Role{ttnpb.PeerInfo_NETWORK_SERVER}
}
