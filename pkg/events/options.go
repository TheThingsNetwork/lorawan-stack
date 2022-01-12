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

	pbtypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/v3/pkg/auth"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
)

// Option is an option that is used to build events.
type Option interface {
	applyTo(*event)
}

type optionFunc func(e *event)

func (f optionFunc) applyTo(e *event) { f(e) }

// WithIdentifiers returns an option that sets the identifiers of the event.
func WithIdentifiers(identifiers ...EntityIdentifiers) Option {
	return optionFunc(func(e *event) {
		for _, ids := range identifiers {
			e.innerEvent.Identifiers = append(e.innerEvent.Identifiers, ids.GetEntityIdentifiers())
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
				e.innerEvent.CorrelationIds = mergeStrings(e.innerEvent.CorrelationIds, cids)
			}
		}
		if data, ok := data.(interface{ GetCorrelationIds() []string }); ok {
			if cids := data.GetCorrelationIds(); len(cids) > 0 {
				cids = append(cids[:0:0], cids...)
				sort.Strings(cids)
				e.innerEvent.CorrelationIds = mergeStrings(e.innerEvent.CorrelationIds, cids)
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
		md := rpcmetadata.FromIncomingContext(e.ctx)
		authentication := &ttnpb.Event_Authentication{}
		if md.AuthType != "" {
			authentication.Type = md.AuthType
		}
		if md.AuthValue != "" {
			if tokenType, tokenID, _, err := auth.SplitToken(md.AuthValue); err == nil {
				authentication.TokenType = tokenType.String()
				authentication.TokenId = tokenID
			}
		}
		if authentication.TokenId != "" || authentication.TokenType != "" || authentication.Type != "" {
			e.innerEvent.Authentication = authentication
		}
	})
}

// WithClientInfoFromContext returns an option that extracts the UserAgent and the RemoteIP from the request context.
func WithClientInfoFromContext() Option {
	return optionFunc(func(e *event) {
		if p, ok := peer.FromContext(e.ctx); ok && p.Addr != nil && p.Addr.String() != "pipe" {
			if host, _, err := net.SplitHostPort(p.Addr.String()); err == nil {
				e.innerEvent.RemoteIp = host
			}
		}
		md := rpcmetadata.FromIncomingContext(e.ctx)
		if md.XForwardedFor != "" {
			xff := strings.Split(md.XForwardedFor, ",")
			e.innerEvent.RemoteIp = strings.Trim(xff[0], " ")
		}
		if md := rpcmetadata.FromIncomingContext(e.ctx); md.UserAgent != "" {
			e.innerEvent.UserAgent = md.UserAgent
		}
	})
}

// DefinitionOption is like Option, but applies to the definition instead.
type DefinitionOption interface {
	Option
	applyToDefinition(*definition)
}

type definitionOptionFunc func(d *definition)

func (definitionOptionFunc) applyTo(*event) {}

func (f definitionOptionFunc) applyToDefinition(d *definition) { f(d) }

// WithDataType returns an option that sets the data type of the event (for documentation).
func WithDataType(t interface{}) DefinitionOption {
	msg, err := marshalData(t)
	if err != nil {
		panic(err)
	}
	return definitionOptionFunc(func(d *definition) {
		d.dataType = msg
	})
}

var errorDataType, _ = marshalData(&ttnpb.ErrorDetails{
	Namespace:     "pkg/example",
	Name:          "example",
	MessageFormat: "example error for `{attr_name}`",
	Attributes: &pbtypes.Struct{
		Fields: map[string]*pbtypes.Value{
			"attr_name": {Kind: &pbtypes.Value_StringValue{
				StringValue: "attr_value",
			}},
		},
	},
	Code: uint32(codes.Unknown),
})

// WithErrorDataType is a convenience function that sets the data type of the event to an error.
func WithErrorDataType() DefinitionOption {
	return definitionOptionFunc(func(d *definition) {
		d.dataType = errorDataType
	})
}

var updatedFieldsDataType, _ = marshalData([]string{"list.of", "updated.fields"})

// WithUpdatedFieldsDataType is a convenience function that sets the data type of the event to a slice of updated fields.
func WithUpdatedFieldsDataType() DefinitionOption {
	return definitionOptionFunc(func(d *definition) {
		d.dataType = updatedFieldsDataType
	})
}

// WithPropagateToParent returns an option that propagate the event to its parent.
// Typically used to propagate end device events to applications.
func WithPropagateToParent() DefinitionOption {
	return definitionOptionFunc(func(d *definition) {
		d.propagateToParent = true
	})
}
