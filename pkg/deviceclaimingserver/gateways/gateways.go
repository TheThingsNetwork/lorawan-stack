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

// Package gateways provides functions to claim gateways.
package gateways

import (
	"context"
	"strings"

	"go.thethings.network/lorawan-stack/v3/pkg/deviceclaimingserver/gateways/ttgc"
	dcstypes "go.thethings.network/lorawan-stack/v3/pkg/deviceclaimingserver/types"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

// Config is the configuration for the Gateway Claiming Server.
type Config struct {
	CreateOnNotFound            bool                `name:"create-on-not-found" description:"DEPRECATED"`                                      // nolint:lll
	DefaultGatewayServerAddress string              `name:"default-gateway-server-address" description:"The default Gateway Server Address"`   // nolint:lll
	Upstreams                   map[string][]string `name:"upstreams" description:"Map of upstream type and the supported Gateway EUI ranges"` // nolint:lll
	TTGC                        ttgc.Config         `name:"ttgc"`
}

var errInvalidUpstream = errors.DefineInvalidArgument("invalid_upstream", "upstream `{name}` is invalid")

// ParseGatewayEUIRanges parses the configured upstream map and returns map of ranges.
func ParseGatewayEUIRanges(conf map[string][]string) (map[string][]dcstypes.EUI64Range, error) {
	res := make(map[string][]dcstypes.EUI64Range, len(conf))
	for host, ranges := range conf {
		res[host] = make([]dcstypes.EUI64Range, 0, len(ranges))
		for _, val := range ranges {
			var r dcstypes.EUI64Range
			switch {
			case strings.Contains(val, "/"):
				var prefix types.EUI64Prefix
				if err := prefix.UnmarshalText([]byte(val)); err != nil {
					return nil, errInvalidUpstream.WithAttributes("name", host).WithCause(err)
				}
				r = dcstypes.RangeFromEUI64Prefix(prefix)
			case strings.Contains(val, "-"):
				parts := strings.Split(val, "-")
				if len(parts) != 2 {
					return nil, errInvalidUpstream.WithAttributes("name", host)
				}
				var start, end types.EUI64
				if err := start.UnmarshalText([]byte(parts[0])); err != nil {
					return nil, errInvalidUpstream.WithAttributes("name", host).WithCause(err)
				}
				if err := end.UnmarshalText([]byte(parts[1])); err != nil {
					return nil, errInvalidUpstream.WithAttributes("name", host).WithCause(err)
				}
				r = dcstypes.RangeFromEUI64Range(start, end)
			default:
				return nil, errInvalidUpstream.WithAttributes("name", host)
			}
			res[host] = append(res[host], r)
		}
	}
	return res, nil
}

// Claimer provides methods for claiming Gateways.
type Claimer interface {
	// Claim claims a gateway.
	Claim(ctx context.Context, eui types.EUI64, ownerToken string, clusterAddress string) error
	// Unclaim unclaims a gateway.
	Unclaim(context.Context, types.EUI64, string) error
}

// rangeClaimer supports claiming a range of EUIs.
type rangeClaimer struct {
	ranges []dcstypes.EUI64Range
	Claimer
}

// Upstream is a gateway claiming upstream.
type Upstream struct {
	claimers map[string]rangeClaimer
}

// NewUpstream returns a new upstream based on the provided configuration.
func NewUpstream(
	ctx context.Context,
	conf Config,
	opts ...Option,
) (*Upstream, error) {
	upstream := &Upstream{
		claimers: make(map[string]rangeClaimer),
	}
	for _, opt := range opts {
		opt(upstream)
	}

	hosts, err := ParseGatewayEUIRanges(conf.Upstreams)
	if err != nil {
		return nil, err
	}
	// Setup upstream table.
	for name, ranges := range hosts {
		if len(ranges) == 0 || name == "" {
			continue
		}
		var claimer Claimer
		switch name {
		case "ttgc":
			claimer, err = conf.TTGC.NewClient(ctx)
			if err != nil {
				return nil, err
			}
		default:
			return nil, errInvalidUpstream.WithAttributes("name", name)
		}
		upstream.claimers[name] = rangeClaimer{
			Claimer: claimer,
			ranges:  ranges,
		}
	}
	return upstream, nil
}

// Option configures Upstream.
type Option func(*Upstream)

// WithClaimer adds a claimer to Upstream.
func WithClaimer(name string, ranges []dcstypes.EUI64Range, claimer Claimer) Option {
	return func(upstream *Upstream) {
		upstream.claimers[name] = rangeClaimer{
			Claimer: claimer,
			ranges:  ranges,
		}
	}
}

// Claimer returns the Claimer for the given Gateway EUI.
func (upstream *Upstream) Claimer(gatewayEUI types.EUI64) Claimer {
	for _, claimer := range upstream.claimers {
		for _, r := range claimer.ranges {
			if r.Contains(gatewayEUI) {
				return claimer.Claimer
			}
		}
	}
	return nil
}
