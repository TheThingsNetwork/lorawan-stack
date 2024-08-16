// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

// Package client implements a managed gateway client.
package client

import (
	"context"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/workerpool"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

// Component is the interface to the component.
type Component interface {
	workerpool.Component
	AllowInsecureForCredentials() bool
	GetPeerConn(ctx context.Context, role ttnpb.ClusterRole, ids cluster.EntityIdentifiers) (*grpc.ClientConn, error)
}

type client struct {
	component Component
}

// NewEvents initializes a new events subscriber.
func NewEvents(component Component) events.Subscriber {
	return &client{component}
}

type eventData struct {
	*ttnpb.GatewayIdentifiers
	*ttnpb.ManagedGatewayEventData
}

func (c *client) subscribeEventData(
	ctx context.Context, ch chan *eventData, ids ...*ttnpb.GatewayIdentifiers,
) (wait func() error, err error) {
	conn, err := c.component.GetPeerConn(ctx, ttnpb.ClusterRole_GATEWAY_CONFIGURATION_SERVER, nil)
	if err != nil {
		return nil, err
	}
	callOpt, err := rpcmetadata.WithForwardedAuth(ctx, c.component.AllowInsecureForCredentials())
	if err != nil {
		return nil, err
	}
	client := ttnpb.NewManagedGatewayConfigurationServiceClient(conn)

	ctx, cancel := context.WithCancel(ctx)
	wg, ctx := errgroup.WithContext(ctx)
	for _, gtwID := range ids {
		gtwID := gtwID
		stream, err := client.StreamEvents(ctx, gtwID, callOpt)
		if err != nil {
			cancel()
			return nil, err
		}
		wg.Go(func() error {
			for {
				data, err := stream.Recv()
				if err != nil {
					return err
				}
				select {
				case <-ctx.Done():
					return ctx.Err()
				case ch <- &eventData{
					GatewayIdentifiers:      gtwID,
					ManagedGatewayEventData: data,
				}:
				}
			}
		})
	}
	return func() error {
		defer cancel()
		return wg.Wait()
	}, nil
}

func mapEvent(ctx context.Context, eventData *eventData) []events.Event {
	switch data := eventData.Data.(type) {
	case *ttnpb.ManagedGatewayEventData_Entity:
		return []events.Event{
			evtUpdateManagedGateway.NewWithIdentifiersAndData(
				ctx, eventData.GatewayIdentifiers, data.Entity,
			),
		}
	case *ttnpb.ManagedGatewayEventData_Location:
		return []events.Event{
			evtUpdateManagedGatewayLocation.NewWithIdentifiersAndData(
				ctx, eventData.GatewayIdentifiers, data.Location,
			),
		}
	case *ttnpb.ManagedGatewayEventData_SystemStatus:
		return []events.Event{
			evtReceiveManagedGatewaySystemStatus.NewWithIdentifiersAndData(
				ctx, eventData.GatewayIdentifiers, data.SystemStatus,
			),
		}
	case *ttnpb.ManagedGatewayEventData_ControllerConnection:
		if data.ControllerConnection.NetworkInterfaceType !=
			ttnpb.ManagedGatewayNetworkInterfaceType_MANAGED_GATEWAY_NETWORK_INTERFACE_TYPE_UNSPECIFIED {
			return []events.Event{
				evtManagedGatewayControllerUp.NewWithIdentifiersAndData(
					ctx, eventData.GatewayIdentifiers, data.ControllerConnection,
				),
			}
		}
		return []events.Event{
			// If the Controller connection is down, report that everything is down because it cannot be determined.
			evtManagedGatewayControllerDown.New(ctx, events.WithIdentifiers(eventData.GatewayIdentifiers)),
			evtManagedGatewayGatewayServerDown.New(ctx, events.WithIdentifiers(eventData.GatewayIdentifiers)),
			evtManagedGatewayCellularDown.New(ctx, events.WithIdentifiers(eventData.GatewayIdentifiers)),
			evtManagedGatewayWiFiDown.New(ctx, events.WithIdentifiers(eventData.GatewayIdentifiers)),
			evtManagedGatewayEthernetDown.New(ctx, events.WithIdentifiers(eventData.GatewayIdentifiers)),
		}
	case *ttnpb.ManagedGatewayEventData_GatewayServerConnection:
		if data.GatewayServerConnection.NetworkInterfaceType !=
			ttnpb.ManagedGatewayNetworkInterfaceType_MANAGED_GATEWAY_NETWORK_INTERFACE_TYPE_UNSPECIFIED {
			return []events.Event{
				evtManagedGatewayGatewayServerUp.NewWithIdentifiersAndData(
					ctx, eventData.GatewayIdentifiers, data.GatewayServerConnection,
				),
			}
		}
		return []events.Event{
			evtManagedGatewayGatewayServerDown.New(ctx, events.WithIdentifiers(eventData.GatewayIdentifiers)),
		}
	case *ttnpb.ManagedGatewayEventData_CellularBackhaul:
		switch data.CellularBackhaul.NetworkInterface.GetStatus() {
		case ttnpb.ManagedGatewayNetworkInterfaceStatus_MANAGED_GATEWAY_NETWORK_INTERFACE_STATUS_UP:
			return []events.Event{
				evtManagedGatewayCellularUp.NewWithIdentifiersAndData(
					ctx, eventData.GatewayIdentifiers, data.CellularBackhaul,
				),
			}
		default:
			return []events.Event{
				evtManagedGatewayCellularDown.New(ctx, events.WithIdentifiers(eventData.GatewayIdentifiers)),
			}
		}
	case *ttnpb.ManagedGatewayEventData_WifiBackhaul:
		switch data.WifiBackhaul.NetworkInterface.GetStatus() {
		case ttnpb.ManagedGatewayNetworkInterfaceStatus_MANAGED_GATEWAY_NETWORK_INTERFACE_STATUS_UP:
			return []events.Event{
				evtManagedGatewayWiFiUp.NewWithIdentifiersAndData(
					ctx, eventData.GatewayIdentifiers, data.WifiBackhaul,
				),
			}
		case ttnpb.ManagedGatewayNetworkInterfaceStatus_MANAGED_GATEWAY_NETWORK_INTERFACE_STATUS_FAILED:
			return []events.Event{
				evtManagedGatewayWiFiFail.New(ctx, events.WithIdentifiers(eventData.GatewayIdentifiers)),
			}
		default:
			return []events.Event{
				evtManagedGatewayWiFiDown.New(ctx, events.WithIdentifiers(eventData.GatewayIdentifiers)),
			}
		}
	case *ttnpb.ManagedGatewayEventData_EthernetBackhaul:
		switch data.EthernetBackhaul.NetworkInterface.GetStatus() {
		case ttnpb.ManagedGatewayNetworkInterfaceStatus_MANAGED_GATEWAY_NETWORK_INTERFACE_STATUS_UP:
			return []events.Event{
				evtManagedGatewayEthernetUp.NewWithIdentifiersAndData(
					ctx, eventData.GatewayIdentifiers, data.EthernetBackhaul,
				),
			}
		default:
			return []events.Event{
				evtManagedGatewayEthernetDown.New(ctx, events.WithIdentifiers(eventData.GatewayIdentifiers)),
			}
		}
	}
	return nil
}

func matcherByNames(names []string) func(name string) bool {
	if len(names) == 0 {
		return func(string) bool { return true }
	}
	m := make(map[string]struct{}, len(names))
	for _, name := range names {
		m[name] = struct{}{}
	}
	return func(name string) bool {
		_, ok := m[name]
		return ok
	}
}

// Subscribe implements events.Subscriber.
// This method does not block once the subscription is established.
func (c *client) Subscribe(
	ctx context.Context, names []string, identifiers []*ttnpb.EntityIdentifiers, hdl events.Handler,
) error {
	ids := make([]*ttnpb.GatewayIdentifiers, 0, len(identifiers))
	for _, id := range identifiers {
		if id.GetGatewayIds() != nil {
			ids = append(ids, id.GetGatewayIds())
		}
	}
	ch := make(chan *eventData)
	wait, err := c.subscribeEventData(ctx, ch, ids...)
	if err != nil {
		return err
	}
	go func() {
		defer close(ch)
		if err := wait(); err != nil && !errors.IsCanceled(err) {
			log.FromContext(ctx).WithError(err).Warn("Failed to subscribe to managed gateway events")
		}
	}()
	match := matcherByNames(names)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case eventData, ok := <-ch:
				if !ok {
					return
				}
				evts := mapEvent(ctx, eventData)
				for _, evt := range evts {
					if !match(evt.Name()) {
						continue
					}
					hdl.Notify(evt)
				}
			}
		}
	}()
	return nil
}

// SubscribeWithHistory implements events.SubscriberWithHistory.
// This method blocks until the context is done or an error occurs.
func (c *client) SubscribeWithHistory(
	ctx context.Context,
	names []string,
	identifiers []*ttnpb.EntityIdentifiers,
	_ *time.Time,
	_ int,
	hdl events.Handler,
) error {
	ids := make([]*ttnpb.GatewayIdentifiers, 0, len(identifiers))
	for _, id := range identifiers {
		if id.GetGatewayIds() != nil {
			ids = append(ids, id.GetGatewayIds())
		}
	}
	ch := make(chan *eventData)
	wait, err := c.subscribeEventData(ctx, ch, ids...)
	if err != nil {
		return err
	}
	match := matcherByNames(names)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case eventData := <-ch:
				evts := mapEvent(ctx, eventData)
				for _, evt := range evts {
					if !match(evt.Name()) {
						continue
					}
					hdl.Notify(evt)
				}
			}
		}
	}()
	return wait()
}

// FindRelated implements events.SubscriberWithHistory.
func (*client) FindRelated(context.Context, string) ([]events.Event, error) {
	return nil, nil
}

// FetchHistory implements events.SubscriberWithHistory.
func (*client) FetchHistory(
	context.Context, []string, []*ttnpb.EntityIdentifiers, *time.Time, int,
) ([]events.Event, error) {
	return nil, nil
}
