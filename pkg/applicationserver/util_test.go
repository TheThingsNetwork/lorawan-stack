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

package applicationserver_test

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcserver"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"google.golang.org/grpc"
)

var testRights = []ttnpb.Right{
	ttnpb.Right_RIGHT_APPLICATION_LINK,
	ttnpb.Right_RIGHT_APPLICATION_SETTINGS_BASIC,
	ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ,
	ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE,
	ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ,
	ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_DOWN_WRITE,
}

func mustHavePeer(ctx context.Context, c *component.Component, role ttnpb.ClusterRole) {
	for i := 0; i < 20; i++ {
		time.Sleep(20 * time.Millisecond)
		if _, err := c.GetPeer(ctx, role, nil); err == nil {
			return
		}
	}
	panic("could not connect to peer")
}

func withDevAddr(ids *ttnpb.EndDeviceIdentifiers, devAddr types.DevAddr) *ttnpb.EndDeviceIdentifiers {
	newIds := &ttnpb.EndDeviceIdentifiers{}
	if err := newIds.SetFields(ids, ttnpb.EndDeviceIdentifiersFieldPathsNested...); err != nil {
		panic(err)
	}
	newIds.DevAddr = devAddr.Bytes()
	return newIds
}

type mockNS struct {
	ttnpb.UnimplementedAsNsServer

	linkCh          chan ttnpb.ApplicationIdentifiers
	unlinkCh        chan ttnpb.ApplicationIdentifiers
	upCh            chan *ttnpb.ApplicationUp
	downlinkQueueMu sync.RWMutex
	downlinkQueue   map[string][]*ttnpb.ApplicationDownlink
}

type mockNSASConn struct {
	cc   *grpc.ClientConn
	auth grpc.CallOption
}

func startMockNS(ctx context.Context, link chan *mockNSASConn) (*mockNS, string) {
	ns := &mockNS{
		linkCh:        make(chan ttnpb.ApplicationIdentifiers, 1),
		unlinkCh:      make(chan ttnpb.ApplicationIdentifiers, 1),
		upCh:          make(chan *ttnpb.ApplicationUp, 1),
		downlinkQueue: make(map[string][]*ttnpb.ApplicationDownlink),
	}
	go ns.sendTraffic(ctx, link)
	srv := rpcserver.New(ctx)
	ttnpb.RegisterAsNsServer(srv.Server, ns)
	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}
	go func() {
		if err := srv.Serve(lis); err != nil {
			panic(err)
		}
	}()
	return ns, lis.Addr().String()
}

func (ns *mockNS) sendTraffic(ctx context.Context, link chan *mockNSASConn) {
	var cc *grpc.ClientConn
	var auth grpc.CallOption
	select {
	case <-ctx.Done():
		return
	case l := <-link:
		cc, auth = l.cc, l.auth
	}
	client := ttnpb.NewNsAsClient(cc)
	for {
		select {
		case <-ctx.Done():
			return
		case up := <-ns.upCh:
			if _, err := client.HandleUplink(ctx, &ttnpb.NsAsHandleUplinkRequest{
				ApplicationUps: []*ttnpb.ApplicationUp{up},
			}, auth); err != nil {
				panic(err)
			}
		}
	}
}

func (ns *mockNS) reset() {
	ns.downlinkQueueMu.Lock()
	ns.downlinkQueue = make(map[string][]*ttnpb.ApplicationDownlink)
	ns.downlinkQueueMu.Unlock()
}

func (ns *mockNS) DownlinkQueueReplace(ctx context.Context, req *ttnpb.DownlinkQueueRequest) (*pbtypes.Empty, error) {
	ns.downlinkQueueMu.Lock()
	ns.downlinkQueue[unique.ID(ctx, req.EndDeviceIds)] = req.Downlinks
	ns.downlinkQueueMu.Unlock()
	return ttnpb.Empty, nil
}

func (ns *mockNS) DownlinkQueuePush(ctx context.Context, req *ttnpb.DownlinkQueueRequest) (*pbtypes.Empty, error) {
	ns.downlinkQueueMu.Lock()
	uid := unique.ID(ctx, req.EndDeviceIds)
	ns.downlinkQueue[uid] = append(ns.downlinkQueue[uid], req.Downlinks...)
	ns.downlinkQueueMu.Unlock()
	return ttnpb.Empty, nil
}

func (ns *mockNS) DownlinkQueueList(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers) (*ttnpb.ApplicationDownlinks, error) {
	ns.downlinkQueueMu.RLock()
	queue := ns.downlinkQueue[unique.ID(ctx, ids)]
	ns.downlinkQueueMu.RUnlock()
	return &ttnpb.ApplicationDownlinks{
		Downlinks: queue,
	}, nil
}

var errNotFound = errors.DefineNotFound("not_found", "not found")

type mockJS struct {
	ttnpb.UnimplementedAsJsServer

	keys map[string]*ttnpb.KeyEnvelope
}

func startMockJS(ctx context.Context) (*mockJS, string) {
	js := &mockJS{
		keys: make(map[string]*ttnpb.KeyEnvelope),
	}
	srv := rpcserver.New(ctx)
	ttnpb.RegisterAsJsServer(srv.Server, js)
	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}
	go func() {
		if err := srv.Serve(lis); err != nil {
			panic(err)
		}
	}()
	return js, lis.Addr().String()
}

func (js *mockJS) add(ctx context.Context, devEUI types.EUI64, sessionKeyID []byte, key *ttnpb.KeyEnvelope) {
	js.keys[fmt.Sprintf("%v:%v", devEUI, sessionKeyID)] = key
}

func (js *mockJS) GetAppSKey(ctx context.Context, req *ttnpb.SessionKeyRequest) (*ttnpb.AppSKeyResponse, error) {
	key, ok := js.keys[fmt.Sprintf("%v:%v", types.MustEUI64(req.DevEui).OrZero(), req.SessionKeyId)]
	if !ok {
		return nil, errNotFound.New()
	}
	return &ttnpb.AppSKeyResponse{
		AppSKey: key,
	}, nil
}
