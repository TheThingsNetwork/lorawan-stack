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

package web

import (
	"context"
	"strconv"

	pbtypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func setTotalHeader(ctx context.Context, total uint64) {
	grpc.SetHeader(ctx, metadata.Pairs("x-total-count", strconv.FormatUint(total, 10)))
}

// appendImplicitWebhookGetPaths appends implicit ttnpb.ApplicationWebhook get paths to paths.
func appendImplicitWebhookGetPaths(paths ...string) []string {
	return append(append(make([]string, 0, 2+len(paths)),
		"base_url",
		"format",
	), paths...)
}

type webhookRegistryRPC struct {
	webhooks  WebhookRegistry
	templates *TemplateStore
}

// NewWebhookRegistryRPC returns a new webhook registry gRPC server.
func NewWebhookRegistryRPC(webhooks WebhookRegistry, templates *TemplateStore) ttnpb.ApplicationWebhookRegistryServer {
	return &webhookRegistryRPC{
		webhooks:  webhooks,
		templates: templates,
	}
}

func (s webhookRegistryRPC) GetFormats(ctx context.Context, _ *pbtypes.Empty) (*ttnpb.ApplicationWebhookFormats, error) {
	fs := make(map[string]string, len(formats))
	for key, val := range formats {
		fs[key] = val.Name
	}
	return &ttnpb.ApplicationWebhookFormats{
		Formats: fs,
	}, nil
}

func (s webhookRegistryRPC) GetTemplate(ctx context.Context, req *ttnpb.GetApplicationWebhookTemplateRequest) (*ttnpb.ApplicationWebhookTemplate, error) {
	return s.templates.GetTemplate(ctx, req)
}

func (s webhookRegistryRPC) ListTemplates(ctx context.Context, req *ttnpb.ListApplicationWebhookTemplatesRequest) (*ttnpb.ApplicationWebhookTemplates, error) {
	return s.templates.ListTemplates(ctx, req)
}

func (s webhookRegistryRPC) Get(ctx context.Context, req *ttnpb.GetApplicationWebhookRequest) (*ttnpb.ApplicationWebhook, error) {
	if err := rights.RequireApplication(ctx, req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_TRAFFIC_READ); err != nil {
		return nil, err
	}
	return s.webhooks.Get(ctx, req.ApplicationWebhookIdentifiers, appendImplicitWebhookGetPaths(req.FieldMask.Paths...))
}

func (s webhookRegistryRPC) List(ctx context.Context, req *ttnpb.ListApplicationWebhooksRequest) (*ttnpb.ApplicationWebhooks, error) {
	if err := rights.RequireApplication(ctx, req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_TRAFFIC_READ); err != nil {
		return nil, err
	}
	webhooks, err := s.webhooks.List(ctx, req.ApplicationIdentifiers, appendImplicitWebhookGetPaths(req.FieldMask.Paths...))
	if err != nil {
		return nil, err
	}
	defer func() {
		if err == nil {
			setTotalHeader(ctx, uint64(len(webhooks)))
		}
	}()
	return &ttnpb.ApplicationWebhooks{
		Webhooks: webhooks,
	}, nil
}

func (s webhookRegistryRPC) Set(ctx context.Context, req *ttnpb.SetApplicationWebhookRequest) (*ttnpb.ApplicationWebhook, error) {
	if err := rights.RequireApplication(ctx, req.ApplicationIdentifiers,
		ttnpb.RIGHT_APPLICATION_SETTINGS_BASIC,
		ttnpb.RIGHT_APPLICATION_TRAFFIC_READ,
		ttnpb.RIGHT_APPLICATION_TRAFFIC_DOWN_WRITE,
	); err != nil {
		return nil, err
	}
	return s.webhooks.Set(ctx, req.ApplicationWebhookIdentifiers, appendImplicitWebhookGetPaths(req.FieldMask.Paths...),
		func(webhook *ttnpb.ApplicationWebhook) (*ttnpb.ApplicationWebhook, []string, error) {
			if webhook != nil {
				return &req.ApplicationWebhook, req.FieldMask.Paths, nil
			}
			return &req.ApplicationWebhook, append(req.FieldMask.Paths,
				"ids.application_ids",
				"ids.webhook_id",
			), nil
		},
	)
}

func (s webhookRegistryRPC) Delete(ctx context.Context, req *ttnpb.ApplicationWebhookIdentifiers) (*pbtypes.Empty, error) {
	if err := rights.RequireApplication(ctx, req.ApplicationIdentifiers,
		ttnpb.RIGHT_APPLICATION_SETTINGS_BASIC,
		ttnpb.RIGHT_APPLICATION_TRAFFIC_READ,
		ttnpb.RIGHT_APPLICATION_TRAFFIC_DOWN_WRITE,
	); err != nil {
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
