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

// Package provider implements pub/sub provider interfaces and registration.
package provider

import (
	"context"
	"fmt"
	"reflect"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// Topics provide a pub/sub base topic and optional, per-message sub-topics.
type Topics interface {
	GetBaseTopic() string
	GetUplinkMessage() *ttnpb.ApplicationPubSub_Message
	GetUplinkNormalized() *ttnpb.ApplicationPubSub_Message
	GetJoinAccept() *ttnpb.ApplicationPubSub_Message
	GetDownlinkAck() *ttnpb.ApplicationPubSub_Message
	GetDownlinkNack() *ttnpb.ApplicationPubSub_Message
	GetDownlinkSent() *ttnpb.ApplicationPubSub_Message
	GetDownlinkFailed() *ttnpb.ApplicationPubSub_Message
	GetDownlinkQueued() *ttnpb.ApplicationPubSub_Message
	GetDownlinkQueueInvalidated() *ttnpb.ApplicationPubSub_Message
	GetLocationSolved() *ttnpb.ApplicationPubSub_Message
	GetServiceData() *ttnpb.ApplicationPubSub_Message
	GetDownlinkPush() *ttnpb.ApplicationPubSub_Message
	GetDownlinkReplace() *ttnpb.ApplicationPubSub_Message
}

// Target represents settings for a pub/sub provider to connect.
type Target interface {
	Topics
	GetProvider() ttnpb.ApplicationPubSub_Provider
}

// Enabler enables providers.
type Enabler interface {
	Enabled(ctx context.Context, provider ttnpb.ApplicationPubSub_Provider) error
}

// Provider represents a pub/sub service provider.
type Provider interface {
	OpenConnection(ctx context.Context, target Target, enabler Enabler) (*Connection, error)
}

var (
	errNotImplemented = errors.DefineUnimplemented(
		"provider_not_implemented",
		"provider `{provider_id}` is not implemented",
	)
	errAlreadyRegistered = errors.DefineAlreadyExists(
		"provider_already_registered",
		"provider `{provider_id}` already registered",
	)

	providers = map[reflect.Type]Provider{}
)

// RegisterProvider registers an implementation for a given pub/sub provider.
func RegisterProvider(p ttnpb.ApplicationPubSub_Provider, implementation Provider) {
	t := reflect.TypeOf(p)
	if _, ok := providers[t]; ok {
		panic(errAlreadyRegistered.WithAttributes("provider_id", fmt.Sprintf("%T", p)))
	}
	providers[t] = implementation
}

// GetProvider returns an implementation for a given target.
func GetProvider(target Target) (Provider, error) {
	t := reflect.TypeOf(target.GetProvider())
	if implementation, ok := providers[t]; ok {
		return implementation, nil
	}
	return nil, errNotImplemented.WithAttributes("provider_id", fmt.Sprintf("%T", target.GetProvider()))
}
