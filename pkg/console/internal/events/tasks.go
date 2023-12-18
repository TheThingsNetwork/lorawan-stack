// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

package events

import (
	"context"
	"errors"
	"io"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/console/internal/events/eventsmux"
	"go.thethings.network/lorawan-stack/v3/pkg/console/internal/events/protocol"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

func makeMuxTask(m eventsmux.Interface, cancel func(error)) func(context.Context) error {
	return func(ctx context.Context) (err error) {
		defer func() { cancel(err) }()
		return m.Run(ctx)
	}
}

func makeReadTask(
	conn *websocket.Conn, m eventsmux.Interface, rateLimit func() error, cancel func(error),
) func(context.Context) error {
	return func(ctx context.Context) (err error) {
		defer func() { cancel(err) }()
		defer func() {
			if closeErr := (websocket.CloseError{}); errors.As(err, &closeErr) {
				log.FromContext(ctx).WithFields(log.Fields(
					"code", closeErr.Code,
					"reason", closeErr.Reason,
				)).Debug("WebSocket closed")
				err = io.EOF
			}
		}()
		for {
			var request protocol.RequestWrapper
			if err := wsjson.Read(ctx, conn, &request); err != nil {
				return err
			}
			if err := rateLimit(); err != nil {
				return err
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case m.Requests() <- request.Contents:
			}
		}
	}
}

func makeWriteTask(conn *websocket.Conn, m eventsmux.Interface, cancel func(error)) func(context.Context) error {
	return func(ctx context.Context) (err error) {
		defer func() { cancel(err) }()
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case response := <-m.Responses():
				if err := wsjson.Write(ctx, conn, response); err != nil {
					return err
				}
			}
		}
	}
}

func makePingTask(conn *websocket.Conn, cancel func(error), period time.Duration) func(context.Context) error {
	return func(ctx context.Context) (err error) {
		ticker := time.NewTicker(period)
		defer ticker.Stop()
		defer func() { cancel(err) }()
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-ticker.C:
				if err := conn.Ping(ctx); err != nil {
					return err
				}
			}
		}
	}
}
