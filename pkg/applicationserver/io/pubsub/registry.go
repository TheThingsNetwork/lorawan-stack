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

package pubsub

import (
	"context"

	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// Registry is a registry for PubSub integrations.
type Registry interface {
	// Get returns the PubSub integration by its identifiers.
	Get(ctx context.Context, ids ttnpb.ApplicationPubSubIdentifiers, paths []string) (*ttnpb.ApplicationPubSub, error)
	// List returns all PubSub integrations of the application.
	List(ctx context.Context, ids ttnpb.ApplicationIdentifiers, paths []string) ([]*ttnpb.ApplicationPubSub, error)
	// Set creates, updates or deletes the PubSub integration by its identifiers.
	Set(ctx context.Context, ids ttnpb.ApplicationPubSubIdentifiers, paths []string, f func(*ttnpb.ApplicationPubSub) (*ttnpb.ApplicationPubSub, []string, error)) (*ttnpb.ApplicationPubSub, error)
}
