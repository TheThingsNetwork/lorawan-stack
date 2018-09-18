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

package applicationserver

import (
	"context"
	"fmt"
	"sync"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"go.thethings.network/lorawan-stack/pkg/applicationserver/io"
	iogrpc "go.thethings.network/lorawan-stack/pkg/applicationserver/io/grpc"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/component"
	errors "go.thethings.network/lorawan-stack/pkg/errorsv3"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
	"google.golang.org/grpc"
)

// ApplicationServer implements the Application Server component.
//
// The Application Server exposes the As, AppAs and AsEndDeviceRegistry services.
type ApplicationServer struct {
	*component.Component

	linkMode       LinkMode
	linkRegistry   LinkRegistry
	deviceRegistry DeviceRegistry

	links sync.Map
}

// New returns new *ApplicationServer.
func New(c *component.Component, conf *Config) (*ApplicationServer, error) {
	as := &ApplicationServer{
		Component:      c,
		linkMode:       conf.LinkMode,
		linkRegistry:   conf.Links,
		deviceRegistry: conf.Devices,
	}

	c.RegisterGRPC(as)
	if conf.LinkMode == LinkAll {
		c.RegisterTask(as.linkAll, component.TaskRestartOnFailure)
	}
	return as, nil
}

// RegisterServices registers services provided by as at s.
func (as *ApplicationServer) RegisterServices(s *grpc.Server) {
	ttnpb.RegisterAsServer(s, as)
	// TODO: Register AsEndDeviceRegistryServer (https://github.com/TheThingsIndustries/lorawan-stack/issues/1117)
	ttnpb.RegisterAppAsServer(s, iogrpc.New(as))
}

// RegisterHandlers registers gRPC handlers.
func (as *ApplicationServer) RegisterHandlers(s *runtime.ServeMux, conn *grpc.ClientConn) {
	ttnpb.RegisterAsHandler(as.Context(), s, conn)
	// TODO: Register AsEndDeviceRegistryHandler (https://github.com/TheThingsIndustries/lorawan-stack/issues/1117)
}

// Roles returns the roles that the Application Server fulfills.
func (as *ApplicationServer) Roles() []ttnpb.PeerInfo_Role {
	return []ttnpb.PeerInfo_Role{ttnpb.PeerInfo_APPLICATION_SERVER}
}

// Connect connects an application or integration by its identifiers to the Application Server, and returns a
// io.Connection for traffic and control.
func (as *ApplicationServer) Connect(ctx context.Context, protocol string, ids ttnpb.ApplicationIdentifiers) (*io.Connection, error) {
	if err := rights.RequireApplication(ctx, ids, ttnpb.RIGHT_APPLICATION_TRAFFIC_READ); err != nil {
		return nil, err
	}

	uid := unique.ID(ctx, ids)
	logger := log.FromContext(ctx).WithField("application_uid", uid)
	ctx = events.ContextWithCorrelationID(ctx, fmt.Sprintf("application_conn:%s", events.NewCorrelationID()))

	val, ok := as.links.Load(uid)
	if !ok {
		return nil, errNotLinked.WithAttributes("application_uid", uid)
	}
	l := val.(*link)
	conn := io.NewConnection(ctx, protocol, ids)
	l.subscribeCh <- conn
	go func() {
		<-ctx.Done()
		l.unsubscribeCh <- conn
	}()
	logger.Info("Application connected")
	return conn, nil
}

var (
	errDeviceNotFound = errors.DefineNotFound("device_not_found", "device `{device_uid}` not found")
)

func (as *ApplicationServer) processUp(ctx context.Context, up *ttnpb.ApplicationUp) error {
	return as.deviceRegistry.Set(ctx, up.EndDeviceIdentifiers, func(ed *ttnpb.EndDevice) (*ttnpb.EndDevice, error) {
		if ed == nil {
			return nil, errDeviceNotFound.WithAttributes("device_uid", unique.ID(ctx, up.EndDeviceIdentifiers))
		}

		// TODO:
		// - Handle join accept; update session
		// - Recompute downlink queue on join accept and invalidation
		// - Decrypt uplink messages
		// - Report events on downlink queue changes

		return ed, nil
	})
}
