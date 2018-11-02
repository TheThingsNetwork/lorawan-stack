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

package web

import (
	"context"

	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// WebhookRegistry is a store for webhooks.
type WebhookRegistry interface {
	// Get returns the webhook by its identifiers.
	Get(ctx context.Context, ids ttnpb.ApplicationWebhookIdentifiers, paths []string) (*ttnpb.ApplicationWebhook, error)
	// List returns all webhooks of the application.
	List(ctx context.Context, ids ttnpb.ApplicationIdentifiers, paths []string) ([]*ttnpb.ApplicationWebhook, error)
	// Set creates, updates or deletes the webhook by its identifiers.
	Set(ctx context.Context, ids ttnpb.ApplicationWebhookIdentifiers, paths []string, f func(*ttnpb.ApplicationWebhook) (*ttnpb.ApplicationWebhook, []string, error)) (*ttnpb.ApplicationWebhook, error)
}
