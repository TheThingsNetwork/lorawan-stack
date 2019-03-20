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

package mock

import (
	"context"
	"time"

	"go.thethings.network/lorawan-stack/pkg/gatewayserver/io"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// Frontend is a mock front-end.
type Frontend struct {
	Up     chan *ttnpb.UplinkMessage
	Status chan *ttnpb.GatewayStatus
	TxAck  chan *ttnpb.TxAcknowledgment
	Down   chan *ttnpb.DownlinkMessage
}

func (*Frontend) Protocol() string { return "mock" }

// ConnectFrontend connects a new mock front-end to the given server.
// The gateway time starts at Unix epoch.
func ConnectFrontend(ctx context.Context, ids ttnpb.GatewayIdentifiers, server io.Server) (*Frontend, error) {
	f := &Frontend{
		Up:     make(chan *ttnpb.UplinkMessage, 1),
		Status: make(chan *ttnpb.GatewayStatus, 1),
		TxAck:  make(chan *ttnpb.TxAcknowledgment, 1),
		Down:   make(chan *ttnpb.DownlinkMessage, 1),
	}
	conn, err := server.Connect(ctx, f, ids)
	if err != nil {
		return nil, err
	}
	started := time.Now()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case up := <-f.Up:
				gatewayTime := time.Unix(0, 0).Add(time.Since(started))
				up.ReceivedAt = time.Now()
				up.Settings.Time = &gatewayTime
				conn.HandleUp(up)
			case status := <-f.Status:
				conn.HandleStatus(status)
			case txAck := <-f.TxAck:
				conn.HandleTxAck(txAck)
			}
		}
	}()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case down := <-conn.Down():
				f.Down <- down
			}
		}
	}()
	return f, nil
}
