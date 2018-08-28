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

package udp

import (
	"context"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	errors "go.thethings.network/lorawan-stack/pkg/errorsv3"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/io"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/io/udp/encoding"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/unique"
)

// Config contains configuration settings for the UDP gateway frontend.
// Use DefaultConfig for recommended settings.
type Config struct {
	// PacketHandlers defines the number of concurrent packet handlers.
	PacketHandlers int `name:"packet-handlers" description:"Number of concurrent packet handlers"`
	// PacketBuffer defines how many packets are buffered to handlers before it overflows.
	PacketBuffer int `name:"packet-buffer" description:"Buffer size of unhandled packets"`
	// DownlinkPathExpires defines for how long a downlink path is valid. A downlink path is renewed on each pull data and
	// TX acknowledgement packet.
	// Gateways typically pull data every 5 seconds.
	DownlinkPathExpires time.Duration `name:"downlink-path-expires" description:"Time after which a downlink path to a gateway expires"`
	// ConnectionExpires defines for how long a connection remains valid while no pull data, push data or TX
	// acknowledgement is received.
	ConnectionExpires time.Duration `name:"connection-expires" description:"Time after which a connection of a gateway expires"`
	// ScheduleLateTime defines the time in advance to the actual transmission the downlink message should be scheduled to
	// the gateway.
	ScheduleLateTime time.Duration `name:"schedule-late-time" description:"Time in advance to send downlink to the gateway when scheduling late"`
	// AddrChangeBlock defines the time to block traffic when the address changes.
	AddrChangeBlock time.Duration `name:"addr-change-block" description:"Time to block traffic when a gateway's address changes"`
}

// DefaultConfig contains the default configuration.
var DefaultConfig = Config{
	PacketHandlers:      10,
	PacketBuffer:        50,
	DownlinkPathExpires: 30 * time.Second,
	ConnectionExpires:   5 * time.Minute,
	ScheduleLateTime:    800 * time.Millisecond,
	AddrChangeBlock:     5 * time.Minute,
}

type srv struct {
	ctx    context.Context
	config Config

	server      io.Server
	conn        *net.UDPConn
	packetCh    chan encoding.Packet
	connections sync.Map
	firewall    Firewall
}

// Start starts the UDP frontend.
func Start(ctx context.Context, server io.Server, conn *net.UDPConn, config Config) {
	ctx = log.NewContextWithField(ctx, "namespace", "gatewayserver/io/udp")
	var firewall Firewall
	if config.AddrChangeBlock > 0 {
		firewall = NewMemoryFirewall(ctx, config.AddrChangeBlock)
	}
	s := &srv{
		ctx:      ctx,
		config:   config,
		server:   server,
		conn:     conn,
		packetCh: make(chan encoding.Packet, config.PacketBuffer),
		firewall: firewall,
	}
	go s.read()
	go s.gc()
	go func() {
		<-ctx.Done()
		s.conn.Close()
	}()
	for i := 0; i < config.PacketHandlers; i++ {
		go s.handlePackets()
	}
}

func (s *srv) read() {
	var buf [65507]byte
	for {
		n, addr, err := s.conn.ReadFromUDP(buf[:])
		if err != nil {
			log.FromContext(s.ctx).WithError(err).Warn("Read failed")
			return
		}

		ctx := log.NewContextWithField(s.ctx, "remote_addr", addr.String())
		packet := encoding.Packet{GatewayAddr: addr}
		if err = packet.UnmarshalBinary(buf[0:n]); err != nil {
			log.FromContext(ctx).WithError(err).Debug("Failed to unmarshal packet")
			continue
		}
		switch packet.PacketType {
		case encoding.PullData, encoding.PushData, encoding.TxAck:
		default:
			log.FromContext(ctx).WithField("packet_type", packet.PacketType).Debug("Invalid packet type for uplink")
			continue
		}
		if packet.GatewayEUI == nil {
			log.FromContext(ctx).Debug("No gateway EUI in uplink message")
			continue
		}

		select {
		case s.packetCh <- packet:
		default:
			log.FromContext(ctx).Warn("Packet handlers busy, dropping packet")
		}
	}
}

func (s *srv) handlePackets() {
	for {
		select {
		case <-s.ctx.Done():
			return

		case packet := <-s.packetCh:
			eui := *packet.GatewayEUI
			ctx := log.NewContextWithField(s.ctx, "gateway_eui", eui)
			logger := log.FromContext(ctx)

			switch packet.PacketType {
			case encoding.PullData, encoding.PushData:
				if err := s.writeAckFor(packet); err != nil {
					logger.WithError(err).Warn("Failed to write acknowledgement")
				}
			}

			if s.firewall != nil && !s.firewall.Filter(packet) {
				logger.Warn("Packet filtered")
				break
			}

			cs, err := s.connect(ctx, eui)
			if err != nil {
				logger.WithError(err).Warn("Failed to connect")
				break
			}

			s.handleUp(cs.io.Context(), cs, packet)
		}
	}
}

func (s *srv) connect(ctx context.Context, eui types.EUI64) (*state, error) {
	cs := &state{
		ioWait:          make(chan struct{}),
		startHandleDown: &sync.Once{},
		lastSeenPull:    time.Now().UnixNano(),
		lastSeenPush:    time.Now().UnixNano(),
	}
	val, loaded := s.connections.LoadOrStore(eui, cs)
	cs = val.(*state)
	if !loaded {
		var io *io.Connection
		var err error
		defer func() {
			if err != nil {
				s.connections.Delete(eui)
			}
			cs.io, cs.ioErr = io, err
			close(cs.ioWait)
		}()
		id := ttnpb.GatewayIdentifiers{EUI: &eui}
		ctx, id, err = s.server.FillGatewayContext(ctx, id)
		if err != nil {
			return nil, err
		}
		uid := unique.ID(ctx, id)
		ctx = log.NewContextWithField(ctx, "gateway_uid", uid)
		ctx = rights.NewContext(ctx, rights.Rights{
			GatewayRights: map[string][]ttnpb.Right{
				uid: {ttnpb.RIGHT_GATEWAY_LINK},
			},
		})
		io, err = s.server.Connect(ctx, "udp", id)
		if err != nil {
			return nil, err
		}
	} else {
		<-cs.ioWait
		if cs.ioErr != nil {
			return nil, cs.ioErr
		}
	}
	return cs, nil
}

func (s *srv) handleUp(ctx context.Context, state *state, packet encoding.Packet) error {
	logger := log.FromContext(ctx)

	switch packet.PacketType {
	case encoding.PullData:
		atomic.StoreInt64(&state.lastSeenPull, time.Now().UnixNano())
		logger.WithField("remote_addr", packet.GatewayAddr.String()).Debug("Storing downlink path")
		state.lastDownlinkPath.Store(downlinkPath{
			addr:    *packet.GatewayAddr,
			version: packet.ProtocolVersion,
		})
		state.startHandleDown.Do(func() {
			go s.handleDown(ctx, state)
		})

	case encoding.PushData:
		atomic.StoreInt64(&state.lastSeenPush, time.Now().UnixNano())
		if len(packet.Data.RxPacket) > 0 {
			var timestamp uint32
			for _, rxMetadata := range packet.Data.RxPacket {
				if rxMetadata.Tmst > timestamp {
					timestamp = rxMetadata.Tmst
				}
			}
			state.syncClock(timestamp)
		}
		msg, err := encoding.TranslateUpstream(*packet.Data, encoding.UpstreamMetadata{
			ID: state.io.Gateway().GatewayIdentifiers,
			IP: packet.GatewayAddr.IP.String(),
		})
		if err != nil {
			logger.WithError(err).Warn("Failed to unmarshal packet")
			return err
		}
		for _, up := range msg.UplinkMessages {
			if err := state.io.HandleUp(up); err != nil {
				logger.WithError(err).Warn("Failed to handle uplink message")
			}
		}
		if msg.GatewayStatus != nil {
			if err := state.io.HandleStatus(msg.GatewayStatus); err != nil {
				logger.WithError(err).Warn("Failed to handle status message")
			}
		}

	case encoding.TxAck:
		atomic.StoreInt64(&state.lastSeenPull, time.Now().UnixNano())
		if atomic.CompareAndSwapUint32(&state.receivedTxAck, 0, 1) {
			logger.Info("Received TX acknowledgement, JIT queue supported")
		}

		// TODO: Send event to NS (https://github.com/TheThingsIndustries/lorawan-stack/issues/1017)
	}

	return nil
}

var (
	errClaimDownlinkFailed = errors.DefineUnavailable("downlink_claim", "failed to claim downlink")
	errDownlinkPathExpired = errors.DefineAborted("downlink_path_expired", "downlink path expired")
)

func (s *srv) handleDown(ctx context.Context, state *state) error {
	logger := log.FromContext(ctx)
	if err := s.server.ClaimDownlink(ctx, state.io.Gateway().GatewayIdentifiers); err != nil {
		logger.WithError(err).Warn("Failed to claim downlink")
		return errClaimDownlinkFailed.WithCause(err)
	}

	healthCheck := time.NewTicker(s.config.DownlinkPathExpires / 2)
	defer healthCheck.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-state.io.Context().Done():
			return state.io.Context().Err()
		case down := <-state.io.Down():
			tx, err := encoding.TranslateDownstream(down)
			if err != nil {
				logger.WithError(err).Warn("Failed to marshal downlink message")
				// TODO: Report to Network Server: https://github.com/TheThingsIndustries/lorawan-stack/issues/1017
				break
			}
			downlinkPath := state.lastDownlinkPath.Load().(downlinkPath)
			logger = logger.WithField("remote_addr", downlinkPath.addr.String())
			packet := encoding.Packet{
				GatewayAddr:     &downlinkPath.addr,
				ProtocolVersion: downlinkPath.version,
				PacketType:      encoding.PullResp,
				Token:           state.nextToken(),
				Data: &encoding.Data{
					TxPacket: &tx,
				},
			}
			write := func() {
				logger.Info("Writing downlink message")
				if err := s.write(packet); err != nil {
					logger.WithError(err).Warn("Failed to write downlink message")
					// TODO: Report to Network Server: https://github.com/TheThingsIndustries/lorawan-stack/issues/1017
				}
			}
			canImmediate := atomic.LoadUint32(&state.receivedTxAck) == 1
			preferLate := state.io.Gateway().ScheduleDownlinkLate
			if canImmediate || !preferLate {
				write()
				break
			}
			gatewayTime, err := state.clock(tx.Tmst)
			if err != nil {
				logger.Warn("Schedule late preferred but no gateway clock available")
				write()
				break
			}
			go func() {
				d := time.Until(gatewayTime.Add(-s.config.ScheduleLateTime))
				logger.WithField("duration", d).Debug("Waiting to schedule downlink message late")
				<-time.After(d)
				write()
			}()
		case <-healthCheck.C:
			lastSeenPull := time.Unix(0, atomic.LoadInt64(&state.lastSeenPull))
			if time.Since(lastSeenPull) > s.config.DownlinkPathExpires {
				logger.Warn("Downlink path expired")
				s.server.UnclaimDownlink(ctx, state.io.Gateway().GatewayIdentifiers)
				state.lastDownlinkPath.Store(downlinkPath{})
				state.startHandleDown = &sync.Once{}
				return errDownlinkPathExpired
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
	ticker := time.NewTicker(s.config.ConnectionExpires / 2)
	for {
		select {
		case <-s.ctx.Done():
			ticker.Stop()
			return
		case <-ticker.C:
			s.connections.Range(func(k, v interface{}) bool {
				state := v.(*state)
				lastSeenPull := time.Unix(0, atomic.LoadInt64(&state.lastSeenPull))
				if time.Since(lastSeenPull) > s.config.ConnectionExpires {
					lastSeenPush := time.Unix(0, atomic.LoadInt64(&state.lastSeenPush))
					if time.Since(lastSeenPush) > s.config.ConnectionExpires {
						select {
						case <-state.ioWait:
							logger.WithField("gateway_eui", k.(types.EUI64)).Warn("Connection expired")
							s.connections.Delete(k)
							state.io.Disconnect(errConnectionExpired)
						default:
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
	// Align for sync/atomic, time are Unix ns
	timeOffset    int64
	lastSeenPull  int64
	lastSeenPush  int64
	receivedTxAck uint32
	pullRespToken uint32

	ioWait chan struct{}
	io     *io.Connection
	ioErr  error

	lastDownlinkPath atomic.Value // downlinkPath
	startHandleDown  *sync.Once
}

func (s *state) nextToken() [2]byte {
	val := atomic.AddUint32(&s.pullRespToken, 1)
	return [2]byte{byte(val >> 8 & 0xff), byte(val & 0xff)}
}

var errNoClock = errors.DefineUnavailable("no_clock_sync", "no clock sync")

// clock gets the synchronized time for a timestamp (in microseconds). The clock should be synchronized using
// syncClock, otherwise an error is returned.
func (s *state) clock(timestamp uint32) (t time.Time, err error) {
	offset := atomic.LoadInt64(&s.timeOffset)
	if offset == 0 {
		return time.Time{}, errNoClock
	}
	t = time.Unix(0, 0)
	t = t.Add(time.Duration(int64(timestamp)*1000 + offset))
	if t.Before(time.Now()) {
		t = t.Add(time.Duration(int64(1<<32) * 1000))
	}
	return
}

// syncClock synchronizes the clock with the timestamp (in microseconds) and the local system time.
func (s *state) syncClock(timestamp uint32) {
	t := time.Now().Add(-time.Duration(timestamp) * time.Microsecond)
	atomic.StoreInt64(&s.timeOffset, t.UnixNano())
	log.FromContext(s.io.Context()).WithField("time", t).Debug("Synchronized gateway time")
}
