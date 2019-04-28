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

// Package joinserver provides a LoRaWAN-compliant Join Server implementation.
package joinserver

import (
	"io"
	"math/rand"
	"sync"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/oklog/ulid"
	"go.thethings.network/lorawan-stack/pkg/cluster"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/rpcmiddleware/hooks"
	"go.thethings.network/lorawan-stack/pkg/rpcmiddleware/rpclog"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"google.golang.org/grpc"
)

// Config represents the JoinServer configuration.
type Config struct {
	Devices         DeviceRegistry       `name:"-"`
	Keys            KeyRegistry          `name:"-"`
	JoinEUIPrefixes []*types.EUI64Prefix `name:"join-eui-prefix" description:"JoinEUI prefixes handled by this JS"`
}

// JoinServer implements the Join Server component.
//
// The Join Server exposes the NsJs and DeviceRegistry services.
type JoinServer struct {
	*component.Component

	devices DeviceRegistry
	keys    KeyRegistry

	euiPrefixes []*types.EUI64Prefix

	entropyMu *sync.Mutex
	entropy   io.Reader

	grpc struct {
		nsJs      nsJsServer
		asJs      asJsServer
		jsDevices jsEndDeviceRegistryServer
	}
}

// New returns new *JoinServer.
func New(c *component.Component, conf *Config) (*JoinServer, error) {
	js := &JoinServer{
		Component: c,

		devices: conf.Devices,
		keys:    conf.Keys,

		euiPrefixes: conf.JoinEUIPrefixes,

		entropyMu: &sync.Mutex{},
		entropy:   ulid.Monotonic(rand.New(rand.NewSource(time.Now().UnixNano())), 0),
	}

	js.grpc.jsDevices = jsEndDeviceRegistryServer{JS: js}
	js.grpc.asJs = asJsServer{JS: js}
	js.grpc.nsJs = nsJsServer{JS: js}

	// TODO: Support authentication from non-cluster-local NS and AS (https://github.com/TheThingsNetwork/lorawan-stack/issues/4).
	hooks.RegisterUnaryHook("/ttn.lorawan.v3.NsJs", rpclog.NamespaceHook, rpclog.UnaryNamespaceHook("joinserver"))
	hooks.RegisterUnaryHook("/ttn.lorawan.v3.AsJs", rpclog.NamespaceHook, rpclog.UnaryNamespaceHook("joinserver"))
	hooks.RegisterUnaryHook("/ttn.lorawan.v3.NsJs", cluster.HookName, c.ClusterAuthUnaryHook())
	hooks.RegisterUnaryHook("/ttn.lorawan.v3.AsJs", cluster.HookName, c.ClusterAuthUnaryHook())

	c.RegisterGRPC(js)
	return js, nil
}

// Roles of the gRPC service.
func (js *JoinServer) Roles() []ttnpb.PeerInfo_Role {
	return []ttnpb.PeerInfo_Role{ttnpb.PeerInfo_JOIN_SERVER}
}

// RegisterServices registers services provided by js at s.
func (js *JoinServer) RegisterServices(s *grpc.Server) {
	ttnpb.RegisterAsJsServer(s, js.grpc.asJs)
	ttnpb.RegisterNsJsServer(s, js.grpc.nsJs)
	ttnpb.RegisterJsEndDeviceRegistryServer(s, js.grpc.jsDevices)
}

// RegisterHandlers registers gRPC handlers.
func (js *JoinServer) RegisterHandlers(s *runtime.ServeMux, conn *grpc.ClientConn) {
	ttnpb.RegisterJsEndDeviceRegistryHandler(js.Context(), s, conn)
}
