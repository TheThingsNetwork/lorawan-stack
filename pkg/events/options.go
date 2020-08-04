// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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
	"net"
	"sort"
	"strings"

	"go.thethings.network/lorawan-stack/v3/pkg/auth"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/grpc/peer"
)

// Option is an option that is used to build events.
type Option interface {
	applyTo(*event)
}

type optionFunc func(e *event)

func (f optionFunc) applyTo(e *event) { f(e) }

// WithIdentifiers returns an option that sets the identifiers of the event.
func WithIdentifiers(identifiers ...ttnpb.Identifiers) Option {
	return optionFunc(func(e *event) {
		for _, ids := range identifiers {
			e.innerEvent.Identifiers = append(e.innerEvent.Identifiers, ids.EntityIdentifiers())
		}
	})
}

// WithData returns an option that sets the data of the event.
func WithData(data interface{}) Option {
	return optionFunc(func(e *event) {
		e.data = data
		if data, ok := data.(interface{ GetCorrelationIDs() []string }); ok {
			if cids := data.GetCorrelationIDs(); len(cids) > 0 {
				cids = append(cids[:0:0], cids...)
				sort.Strings(cids)
				e.innerEvent.CorrelationIDs = mergeStrings(e.innerEvent.CorrelationIDs, cids)
			}
		}
	})
}

// WithVisibility returns an option that sets the visibility of the event.
func WithVisibility(rights ...ttnpb.Right) Option {
	return optionFunc(func(e *event) {
		e.innerEvent.Visibility = ttnpb.RightsFrom(rights...)
	})
}

// WithAuthFromContext returns an option that extracts auth information from the context when the event is created.
func WithAuthFromContext() Option {
	return optionFunc(func(e *event) {
		authentication := &ttnpb.Event_Authentication{}
		if p, ok := peer.FromContext(e.ctx); ok && p.Addr != nil && p.Addr.String() != "pipe" {
			if host, _, err := net.SplitHostPort(p.Addr.String()); err == nil {
				authentication.RemoteIP = host
			}
		}
		md := rpcmetadata.FromIncomingContext(e.ctx)
		if md.AuthType != "" {
			authentication.Type = md.AuthType
		}
		if md.AuthValue != "" {
			if tokenType, tokenID, _, err := auth.SplitToken(md.AuthValue); err == nil {
				authentication.TokenType = tokenType.String()
				authentication.TokenID = tokenID
			}
		}
		if md.XForwardedFor != "" {
			xff := strings.Split(md.XForwardedFor, ",")
			authentication.RemoteIP = strings.Trim(xff[0], " ")
		}
		if authentication.RemoteIP != "" || authentication.TokenID != "" ||
			authentication.TokenType != "" || authentication.Type != "" {
			e.innerEvent.Authentication = authentication
		}
	})
}
