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

	pbtypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/pkg/applicationserver/io/fmt"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

type webhookRegistryRPC struct {
	webhooks WebhookRegistry
}

// NewWebhookRegistryRPC returns a new webhook registry gRPC server.
func NewWebhookRegistryRPC(webhooks WebhookRegistry) ttnpb.ApplicationWebhookRegistryServer {
	return &webhookRegistryRPC{
		webhooks: webhooks,
	}
}

func (s webhookRegistryRPC) GetFormats(ctx context.Context, ids *ttnpb.ApplicationIdentifiers) (*ttnpb.ApplicationWebhookFormats, error) {
	if err := rights.RequireApplication(ctx, *ids, ttnpb.RIGHT_APPLICATION_TRAFFIC_READ); err != nil {
		return nil, err
	}
	formats := make(map[string]string, len(fmt.Formatters))
	for format, formatter := range fmt.Formatters {
		formats[format] = formatter.Name()
	}
	return &ttnpb.ApplicationWebhookFormats{
		Formats: formats,
	}, nil
}

func (s webhookRegistryRPC) Get(ctx context.Context, req *ttnpb.GetApplicationWebhookRequest) (*ttnpb.ApplicationWebhook, error) {
	if err := rights.RequireApplication(ctx, req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_TRAFFIC_READ); err != nil {
		return nil, err
	}
	return s.webhooks.Get(ctx, req.ApplicationWebhookIdentifiers, req.FieldMask.Paths)
}

func (s webhookRegistryRPC) List(ctx context.Context, req *ttnpb.ListApplicationWebhooksRequest) (*ttnpb.ApplicationWebhooks, error) {
	if err := rights.RequireApplication(ctx, req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_TRAFFIC_READ); err != nil {
		return nil, err
	}
	webhooks, err := s.webhooks.List(ctx, req.ApplicationIdentifiers, req.FieldMask.Paths)
	if err != nil {
		return nil, err
	}
	return &ttnpb.ApplicationWebhooks{
		Webhooks: webhooks,
	}, nil
}

func (s webhookRegistryRPC) Set(ctx context.Context, req *ttnpb.SetApplicationWebhookRequest) (*ttnpb.ApplicationWebhook, error) {
	if err := rights.RequireApplication(ctx, req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_TRAFFIC_READ); err != nil {
		return nil, err
	}
	return s.webhooks.Set(ctx, req.ApplicationWebhookIdentifiers, req.FieldMask.Paths,
		func(webhook *ttnpb.ApplicationWebhook) (*ttnpb.ApplicationWebhook, []string, error) {
			return &req.ApplicationWebhook, req.FieldMask.Paths, nil
		},
	)
}

func (s webhookRegistryRPC) Delete(ctx context.Context, req *ttnpb.ApplicationWebhookIdentifiers) (*pbtypes.Empty, error) {
	if err := rights.RequireApplication(ctx, req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_TRAFFIC_READ); err != nil {
		return nil, err
	}
	_, err := s.webhooks.Set(ctx, *req, nil,
		func(webhook *ttnpb.ApplicationWebhook) (*ttnpb.ApplicationWebhook, []string, error) {
			return nil, nil, nil
		},
	)
	if err != nil {
		return nil, err
	}
	return ttnpb.Empty, nil
}
