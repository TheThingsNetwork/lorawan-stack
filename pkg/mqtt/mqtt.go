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

// Package mqtt contains MQTT-related utilities.
package mqtt

import (
	"context"
	"sync"

	mqttlog "github.com/TheThingsIndustries/mystique/pkg/log"
	mqttnet "github.com/TheThingsIndustries/mystique/pkg/net"
	"github.com/TheThingsIndustries/mystique/pkg/packet"
	"github.com/TheThingsIndustries/mystique/pkg/session"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ratelimit"
	"go.thethings.network/lorawan-stack/v3/pkg/task"
)

// RunListener runs the MQTT accept connection loop.
func RunListener(
	ctx context.Context,
	lis mqttnet.Listener,
	ts task.Starter,
	createResource func(string) ratelimit.Resource,
	rateLimiter ratelimit.Interface,
	setupConnection func(context.Context, mqttnet.Conn) error,
) error {
	ctx = mqttlog.NewContext(ctx, Logger(log.FromContext(ctx)))
	for {
		mqttConn, err := lis.Accept()
		if err != nil {
			if ctx.Err() == nil {
				log.FromContext(ctx).WithError(err).Warn("Accept failed")
			}
			return err
		}

		remoteAddr := mqttConn.RemoteAddr().String()
		ctx := log.NewContextWithFields(ctx, log.Fields("remote_addr", remoteAddr))

		resource := createResource(remoteAddr)
		if err := ratelimit.Require(rateLimiter, resource); err != nil {
			log.FromContext(ctx).WithError(err).Debug("Drop connection")
			mqttConn.Close()
			continue
		}

		f := func(ctx context.Context) (err error) {
			defer func() {
				if err != nil {
					mqttConn.Close()
				}
			}()
			return setupConnection(ctx, mqttConn)
		}
		ts.StartTask(&task.Config{
			Context: ctx,
			ID:      "mqtt_setup_connection",
			Func:    f,
			Restart: task.RestartNever,
			Backoff: task.DefaultBackoffConfig,
		})
	}
}

// RunSession reads the control packets from the provided session and sends them to the connection.
func RunSession(
	ctx context.Context,
	cancel func(error),
	ts task.Starter,
	session session.Session,
	mqttConn mqttnet.Conn,
	wg *sync.WaitGroup,
) {
	wg.Add(2)
	controlCh := make(chan packet.ControlPacket)
	controlFunc := func(ctx context.Context) error {
		defer wg.Done()
		for {
			pkt, err := session.ReadPacket()
			if err != nil {
				cancel(err)
				return err
			}
			if pkt == nil {
				continue
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case controlCh <- pkt:
			}
		}
	}
	writeFunc := func(ctx context.Context) error {
		defer wg.Done()
		for {
			var pkt packet.ControlPacket
			select {
			case <-ctx.Done():
				return ctx.Err()
			case pkt = <-controlCh:
			case pkt = <-session.PublishChan():
			}
			if err := mqttConn.Send(pkt); err != nil {
				cancel(err)
				return err
			}
		}
	}
	closeFunc := func(ctx context.Context) error {
		log.FromContext(ctx).Info("Connected")
		<-ctx.Done()
		log.FromContext(ctx).WithError(ctx.Err()).Info("Disconnected")

		session.Close()
		mqttConn.Close()

		wg.Wait()

		return ctx.Err()
	}

	for name, f := range map[string]func(context.Context) error{
		"mqtt_control_packets":  controlFunc,
		"mqtt_write_packets":    writeFunc,
		"mqtt_close_connection": closeFunc,
	} {
		ts.StartTask(&task.Config{
			Context: ctx,
			ID:      name,
			Func:    f,
			Restart: task.RestartNever,
			Backoff: task.DefaultBackoffConfig,
		})
	}
}
