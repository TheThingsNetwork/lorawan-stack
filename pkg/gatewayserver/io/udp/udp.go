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

package udp

import (
	"context"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/scheduling"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ratelimit"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	encoding "go.thethings.network/lorawan-stack/v3/pkg/ttnpb/udp"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"go.thethings.network/lorawan-stack/v3/pkg/workerpool"
)

type srv struct {
	ctx    context.Context
	config Config

	server      io.Server
	conn        *net.UDPConn
	connections sync.Map
	firewall    Firewall

	limitLogs ratelimit.Interface
}

func (*srv) Protocol() string            { return "udp" }
func (*srv) SupportsDownlinkClaim() bool { return true }

var (
	limitLogsConfig      = config.RateLimitingProfile{MaxPerMin: 1}
	limitLogsSize   uint = 1 << 13
)

// Serve serves the UDP frontend.
func Serve(ctx context.Context, server io.Server, conn *net.UDPConn, config Config) error {
	ctx = log.NewContextWithField(ctx, "namespace", "gatewayserver/io/udp")
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	var firewall Firewall
	if config.AddrChangeBlock > 0 {
		firewall = NewMemoryFirewall(ctx, config.AddrChangeBlock)
	}
	if config.RateLimiting.Enable {
		firewall = NewRateLimitingFirewall(firewall, config.RateLimiting.Messages, config.RateLimiting.Threshold)
	}
	limitLogs, err := ratelimit.NewProfile(ctx, limitLogsConfig, limitLogsSize)
	if err != nil {
		return err
	}
	s := &srv{
		ctx:      ctx,
		config:   config,
		server:   server,
		conn:     conn,
		firewall: firewall,

		limitLogs: limitLogs,
	}
	wp := workerpool.NewWorkerPool(workerpool.Config{
		Component:  server,
		Context:    ctx,
		Name:       "udp",
		Handler:    s.handlePacket,
		MaxWorkers: config.PacketHandlers,
		QueueSize:  config.PacketBuffer,
	})
	go s.gc()
	go func() {
		<-ctx.Done()
		s.conn.Close()
	}()
	return s.read(wp)
}

func (s *srv) read(wp workerpool.WorkerPool) error {
	var buf [65507]byte
	for {
		n, addr, err := s.conn.ReadFromUDP(buf[:])
		if err != nil {
			if s.ctx.Err() == nil {
				log.FromContext(s.ctx).WithError(err).Warn("Read failed")
			}
			return err
		}
		now := time.Now()
		ctx := log.NewContextWithField(s.ctx, "remote_addr", addr.String())
		logger := log.FromContext(ctx)

		if err := ratelimit.Require(s.server.RateLimiter(), ratelimit.GatewayUDPTrafficResource(addr)); err != nil {
			if ratelimit.Require(s.limitLogs, ratelimit.NewCustomResource(addr.IP.String())) == nil {
				logger.WithError(err).Warn("Drop packet")
			}
			continue
		}

		packetBuf := make([]byte, n)
		copy(packetBuf, buf[:])

		packet := encoding.Packet{
			GatewayAddr: addr,
			ReceivedAt:  now,
		}
		if err := packet.UnmarshalBinary(packetBuf); err != nil {
			logger.WithError(err).Debug("Failed to unmarshal packet")
			continue
		}
		switch packet.PacketType {
		case encoding.PullData, encoding.PushData, encoding.TxAck:
		default:
			logger.WithField("packet_type", packet.PacketType).Debug("Invalid packet type for uplink")
			continue
		}
		if packet.GatewayEUI == nil {
			logger.Debug("No gateway EUI in uplink message")
			continue
		}

		if err := wp.Publish(ctx, packet); err != nil {
			logger.WithError(err).Warn("UDP packet publishing failed")
		}
	}
}

func (s *srv) handlePacket(ctx context.Context, pkt interface{}) {
	packet := pkt.(encoding.Packet)

	eui := *packet.GatewayEUI
	ctx = log.NewContextWithField(ctx, "gateway_eui", eui)
	logger := log.FromContext(ctx)

	switch packet.PacketType {
	case encoding.PullData, encoding.PushData:
		if err := s.writeAckFor(packet); err != nil {
			logger.WithError(err).Warn("Failed to write acknowledgment")
		}
	}

	if s.firewall != nil {
		if err := s.firewall.Filter(packet); err != nil {
			if !errors.IsResourceExhausted(err) || ratelimit.Require(s.limitLogs, ratelimit.NewCustomResource(eui.String())) == nil {
				logger.WithError(err).Warn("Packet filtered")
			}
			return
		}
	}

	cs, err := s.connect(ctx, eui)
	if err != nil {
		logger.WithError(err).Warn("Failed to connect")
		return
	}

	if err := s.handleUp(cs.io.Context(), cs, packet); err != nil {
		logger.WithError(err).Warn("Failed to handle upstream packet")
	}
}

var errConnectionNotReady = errors.DefineUnavailable("connection_not_ready", "connection is not ready")

func (s *srv) connect(ctx context.Context, eui types.EUI64) (*state, error) {
	cs := &state{
		ioWait:           make(chan struct{}),
		downlinkTaskDone: &sync.WaitGroup{},
		startHandleDown:  &sync.Once{},
		lastSeenPull:     time.Now().UnixNano(),
		lastSeenPush:     time.Now().UnixNano(),
	}
	val, loaded := s.connections.LoadOrStore(eui, cs)
	cs = val.(*state)
	if !loaded {
		var io *io.Connection
		var err error
		defer func() {
			if err != nil {
				delete := func() { s.connections.Delete(eui) }
				if expiration := s.config.ConnectionErrorExpires; expiration != 0 {
					time.AfterFunc(expiration, delete)
				} else {
					delete()
				}
			}
			cs.io, cs.ioErr = io, err
			close(cs.ioWait)
		}()
		ids := ttnpb.GatewayIdentifiers{Eui: &eui}
		ctx, ids, err = s.server.FillGatewayContext(ctx, ids)
		if err != nil {
			return nil, err
		}
		uid := unique.ID(ctx, ids)
		ctx = log.NewContextWithField(ctx, "gateway_uid", uid)
		ctx = rights.NewContext(ctx, rights.Rights{
			GatewayRights: map[string]*ttnpb.Rights{
				uid: {
					Rights: []ttnpb.Right{ttnpb.Right_RIGHT_GATEWAY_LINK},
				},
			},
		})
		io, err = s.server.Connect(ctx, s, ids)
		if err != nil {
			return nil, err
		}
	} else {
		select {
		case <-cs.ioWait:
		default:
			return nil, errConnectionNotReady.New()
		}
		if cs.ioErr != nil {
			return nil, cs.ioErr
		}
		// The connection may be disconnected and is awaiting garbage collection, see gc().
		// The connection cannot be deleted from the map at this point, because before that, the downlink tasks must be
		// awaited, which is not desirable here in the hot path.
		if err := cs.io.Context().Err(); err != nil {
			return nil, err
		}
	}
	return cs, nil
}

func (s *srv) handleUp(ctx context.Context, state *state, packet encoding.Packet) error {
	logger := log.FromContext(ctx)
	md := encoding.UpstreamMetadata{
		ID: *state.io.Gateway().GetIds(),
		IP: packet.GatewayAddr.IP.String(),
	}

	now := time.Now()
	switch packet.PacketType {
	case encoding.PullData:
		atomic.StoreInt64(&state.lastSeenPull, now.UnixNano())
		state.lastDownlinkPath.Store(downlinkPath{
			addr:    *packet.GatewayAddr,
			version: packet.ProtocolVersion,
		})
		state.startHandleDownMu.RLock()
		state.startHandleDown.Do(func() {
			state.downlinkTaskDone.Add(1)
			go func() {
				defer state.downlinkTaskDone.Done()
				if err := s.handleDown(ctx, state); err != nil && !errors.Is(err, errDownlinkPathExpired) {
					logger.WithError(err).Warn("Failed to handle downstream packet")
				}
			}()
		})
		state.startHandleDownMu.RUnlock()

	case encoding.PushData:
		atomic.StoreInt64(&state.lastSeenPush, now.UnixNano())
		if len(packet.Data.RxPacket) > 0 {
			var timestamp uint32
			for _, pkt := range packet.Data.RxPacket {
				if pkt.Tmst > timestamp {
					timestamp = pkt.Tmst
				}
			}
			state.clockMu.Lock()
			state.clock.Sync(timestamp, now)
			state.clockMu.Unlock()
		}
		msg, err := encoding.ToGatewayUp(*packet.Data, md)
		if err != nil {
			logger.WithError(err).Warn("Failed to unmarshal packet")
			return err
		}
		for _, up := range io.UniqueUplinkMessagesByRSSI(msg.UplinkMessages) {
			up.ReceivedAt = ttnpb.ProtoTimePtr(packet.ReceivedAt)
			if err := state.io.HandleUp(up, nil); err != nil {
				logger.WithError(err).Warn("Failed to handle uplink message")
			}
		}
		if msg.GatewayStatus != nil {
			if err := state.io.HandleStatus(msg.GatewayStatus); err != nil {
				logger.WithError(err).Warn("Failed to handle status message")
			}
		}

	case encoding.TxAck:
		atomic.StoreInt64(&state.lastSeenPull, now.UnixNano())
		if atomic.CompareAndSwapUint32(&state.receivedTxAck, 0, 1) {
			logger.Debug("Received Tx acknowledgment, JIT queue supported")
		}
		var msg *ttnpb.GatewayUp
		if packet.Data.TxPacketAck != nil {
			var err error
			msg, err = encoding.ToGatewayUp(*packet.Data, md)
			if err != nil {
				logger.WithError(err).Warn("Failed to unmarshal packet")
				return err
			}
		} else {
			msg = &ttnpb.GatewayUp{
				TxAcknowledgment: &ttnpb.TxAcknowledgment{
					Result: ttnpb.TxAcknowledgment_SUCCESS,
				},
			}
		}
		var rtt *time.Duration
		if downlink, delta, ok := state.tokens.Get(uint16(packet.Token[0])<<8|uint16(packet.Token[1]), packet.ReceivedAt); ok {
			msg.TxAcknowledgment.DownlinkMessage = downlink
			msg.TxAcknowledgment.CorrelationIds = downlink.CorrelationIds
			rtt = &delta
		}
		if err := state.io.HandleTxAck(msg.TxAcknowledgment); err != nil {
			logger.WithError(err).Warn("Failed to handle Tx acknowledgment")
		}
		if rtt != nil {
			state.io.RecordRTT(*rtt, packet.ReceivedAt)
		}
		// TODO: Send event to NS (https://github.com/TheThingsNetwork/lorawan-stack/issues/76)
	}

	return nil
}

var (
	errClaimDownlinkFailed = errors.DefineUnavailable("downlink_claim", "failed to claim downlink")
	errDownlinkPathExpired = errors.DefineAborted("downlink_path_expired", "downlink path expired")
)

func (s *srv) handleDown(ctx context.Context, state *state) error {
	defer func() {
		state.lastDownlinkPath.Store(downlinkPath{})
		state.startHandleDownMu.Lock()
		state.startHandleDown = &sync.Once{}
		state.startHandleDownMu.Unlock()
	}()
	logger := log.FromContext(ctx)
	if err := s.server.ClaimDownlink(ctx, *state.io.Gateway().GetIds()); err != nil {
		logger.WithError(err).Error("Failed to claim downlink path")
		return errClaimDownlinkFailed.WithCause(err)
	}
	logger.Info("Downlink path claimed")
	defer func() {
		ctx := s.server.FromRequestContext(ctx)
		if err := s.server.UnclaimDownlink(ctx, *state.io.Gateway().GetIds()); err != nil {
			logger.WithError(err).Error("Failed to unclaim downlink path")
			return
		}
		logger.Info("Downlink path unclaimed")
	}()
	healthCheck := time.NewTicker(s.config.DownlinkPathExpires / 2)
	defer healthCheck.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-state.io.Context().Done():
			return state.io.Context().Err()
		case down := <-state.io.Down():
			tx, err := encoding.FromDownlinkMessage(down)
			if err != nil {
				logger.WithError(err).Warn("Failed to marshal downlink message")
				// TODO: Report to Network Server: https://github.com/TheThingsNetwork/lorawan-stack/issues/76
				break
			}
			downlinkPath := state.lastDownlinkPath.Load().(downlinkPath)
			logger := logger.WithField("remote_addr", downlinkPath.addr.String())
			packet := encoding.Packet{
				GatewayAddr:     &downlinkPath.addr,
				ProtocolVersion: downlinkPath.version,
				PacketType:      encoding.PullResp,
				Data: &encoding.Data{
					TxPacket: tx,
				},
			}
			write := func() {
				logger.Debug("Write downlink message")
				token := state.tokens.Next(down, time.Now())
				packet.Token = [2]byte{byte(token >> 8), byte(token)}
				if err := s.write(packet); err != nil {
					logger.WithError(err).Warn("Failed to write downlink message")
					// TODO: Report to Network Server: https://github.com/TheThingsNetwork/lorawan-stack/issues/76
				}
			}
			canImmediate := atomic.LoadUint32(&state.receivedTxAck) == 1
			forceLate := state.io.Gateway().ScheduleDownlinkLate
			if canImmediate && !forceLate {
				write()
				break
			}
			state.clockMu.RLock()
			if !state.clock.IsSynced() {
				state.clockMu.RUnlock()
				logger.Warn("Schedule late forced but no gateway clock available")
				write()
				break
			}
			serverTime := state.clock.ToServerTime(state.clock.FromTimestampTime(tx.Tmst))
			state.clockMu.RUnlock()
			d := time.Until(serverTime.Add(-s.config.ScheduleLateTime))
			logger.WithField("duration", d).Debug("Wait to schedule downlink message late")
			time.AfterFunc(d, write)
		case <-healthCheck.C:
			lastSeenPull := time.Unix(0, atomic.LoadInt64(&state.lastSeenPull))
			if time.Since(lastSeenPull) > s.config.DownlinkPathExpires {
				logger.Debug("Downlink path expired")
				return errDownlinkPathExpired.New()
			}
		}
	}
}

func (s *srv) write(packet encoding.Packet) error {
	buf, err := packet.MarshalBinary()
	if err != nil {
		return err
	}
	_, err = s.conn.WriteToUDP(buf, packet.GatewayAddr)
	return err
}

func (s *srv) writeAckFor(packet encoding.Packet) error {
	ack, err := packet.BuildAck()
	if err != nil {
		return err
	}
	buf, err := ack.MarshalBinary()
	if err != nil {
		return err
	}
	_, err = s.conn.WriteToUDP(buf, packet.GatewayAddr)
	return err
}

var errConnectionExpired = errors.Define("connection_expired", "connection expired")

func (s *srv) gc() {
	logger := log.FromContext(s.ctx)
	connectionsTicker := time.NewTicker(s.config.ConnectionExpires / 2)
	for {
		select {
		case <-s.ctx.Done():
			connectionsTicker.Stop()
			return
		case <-connectionsTicker.C:
			s.connections.Range(func(k, v interface{}) bool {
				logger := logger.WithField("gateway_eui", k.(types.EUI64))
				state := v.(*state)
				select {
				case <-state.ioWait:
				default:
					return true
				}
				if state.ioErr != nil {
					return true
				}
				select {
				case <-state.io.Context().Done():
					logger.Debug("Connection context done")
					state.downlinkTaskDone.Wait()
					s.connections.Delete(k)
				default:
					lastSeenPull := time.Unix(0, atomic.LoadInt64(&state.lastSeenPull))
					if time.Since(lastSeenPull) > s.config.ConnectionExpires {
						lastSeenPush := time.Unix(0, atomic.LoadInt64(&state.lastSeenPush))
						if time.Since(lastSeenPush) > s.config.ConnectionExpires {
							logger.Debug("Connection expired")
							state.io.Disconnect(errConnectionExpired.New())
							state.downlinkTaskDone.Wait()
							s.connections.Delete(k)
						}
					}
				}
				return true
			})
		}
	}
}

type downlinkPath struct {
	addr    net.UDPAddr
	version encoding.ProtocolVersion
}

type state struct {
	// Align for sync/atomic, timestamps are Unix ns.
	lastSeenPull  int64
	lastSeenPush  int64
	receivedTxAck uint32

	ioWait chan struct{}
	io     *io.Connection
	ioErr  error

	clock   scheduling.RolloverClock
	clockMu sync.RWMutex

	downlinkTaskDone  *sync.WaitGroup
	lastDownlinkPath  atomic.Value // downlinkPath
	startHandleDown   *sync.Once
	startHandleDownMu sync.RWMutex

	tokens io.DownlinkTokens
}
