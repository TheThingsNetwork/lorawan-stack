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

package internal

import (
	"context"

	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

type webhookDataKeyType struct{}

var webhookDataKey webhookDataKeyType

// WebhookData contains the data that is passed through the context.
type WebhookData struct {
	EndDeviceIDs *ttnpb.EndDeviceIdentifiers
	WebhookIDs   *ttnpb.ApplicationWebhookIdentifiers
	Health       interface {
		// Health should always be either nil or *ttnpb.ApplicationWebhookHealth.
		// As Go does not support sum types, this interface acts as a workaround.
		GetHealthy() *ttnpb.ApplicationWebhookHealth_WebhookHealthStatusHealthy
	}
}

// WithWebhookData returns a new context with the given WebhookData.
func WithWebhookData(ctx context.Context, data *WebhookData) context.Context {
	return context.WithValue(ctx, webhookDataKey, data)
}

func webhookDataFromContext(ctx context.Context) *WebhookData {
	data, ok := ctx.Value(webhookDataKey).(*WebhookData)
	if !ok {
		panic("no webhook data found in context")
	}
	return data
}

// DeviceIDFromContext returns the EndDeviceIdentifiers from the context.
func DeviceIDFromContext(ctx context.Context) *ttnpb.EndDeviceIdentifiers {
	data := webhookDataFromContext(ctx)
	if data.EndDeviceIDs.IsZero() {
		panic("no end device identifiers found in context")
	}
	return data.EndDeviceIDs
}

// WebhookIDFromContext returns the ApplicationWebhookIdentifiers from the context.
func WebhookIDFromContext(ctx context.Context) *ttnpb.ApplicationWebhookIdentifiers {
	data := webhookDataFromContext(ctx)
	if data.WebhookIDs.IsZero() {
		panic("no webhook identifiers found in context")
	}
	return data.WebhookIDs
}

// WebhookHealthFromContext returns the ApplicationWebhookHealth from the context.
func WebhookHealthFromContext(ctx context.Context) (*ttnpb.ApplicationWebhookHealth, bool) {
	data := webhookDataFromContext(ctx)
	health, ok := data.Health.(*ttnpb.ApplicationWebhookHealth)
	return health, ok
}
