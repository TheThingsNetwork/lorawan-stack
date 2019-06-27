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

package provider

import (
	"context"

	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// Provider represents a PubSub service provider.
type Provider interface {
	// OpenConnection opens the Connection of a given ttnpb.ApplicationPubSub.
	OpenConnection(ctx context.Context, pb *ttnpb.ApplicationPubSub) (*Connection, error)
}

var (
	errNotImplemented    = errors.DefineUnimplemented("provider_not_implemented", "provider `{provider_id}` is not implemented")
	errAlreadyRegistered = errors.DefineAlreadyExists("already_registered", "provider `{provider_id}` already registered")

	providers = map[ttnpb.ApplicationPubSub_Provider]Provider{}
)

// RegisterProvider registers an implementation for a given PubSub provider.
func RegisterProvider(p ttnpb.ApplicationPubSub_Provider, implementation Provider) {
	if _, ok := providers[p]; ok {
		panic(errAlreadyRegistered.WithAttributes("provider_id", p))
	}
	providers[p] = implementation
}

// GetProvider returns an implementation for a given provider.
func GetProvider(p ttnpb.ApplicationPubSub_Provider) (Provider, error) {
	if implementation, ok := providers[p]; ok {
		return implementation, nil
	}
	return nil, errNotImplemented.WithAttributes("provider_id", p)
}
