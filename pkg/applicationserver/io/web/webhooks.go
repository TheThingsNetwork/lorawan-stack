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
	stdio "io"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/web/internal"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/web/sink"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/task"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	ttnweb "go.thethings.network/lorawan-stack/v3/pkg/web"
	"go.thethings.network/lorawan-stack/v3/pkg/webhandlers"
	"go.thethings.network/lorawan-stack/v3/pkg/webmiddleware"
	"go.thethings.network/lorawan-stack/v3/pkg/workerpool"
)

const namespace = "applicationserver/io/web"

var webhookFanOutFieldMask = []string{
	"base_url",
	"downlink_ack",
	"downlink_api_key",
	"downlink_failed",
	"downlink_nack",
	"downlink_queue_invalidated",
	"downlink_queued",
	"downlink_sent",
	"field_mask",
	"format",
	"headers",
	"health_status",
	"join_accept",
	"location_solved",
	"service_data",
	"paused",
	"uplink_message",
	"uplink_normalized",
}

// Webhooks is an interface for registering incoming webhooks for downlink and creating a subscription to outgoing
// webhooks for upstream data.
type Webhooks interface {
	ttnweb.Registerer
	Registry() WebhookRegistry
}

type webhooks struct {
	ctx       context.Context
	server    io.Server
	registry  WebhookRegistry
	target    sink.Sink
	downlinks DownlinksConfig
}

// NewWebhooks returns a new Webhooks.
func NewWebhooks(
	ctx context.Context,
	server io.Server,
	registry WebhookRegistry,
	target sink.Sink,
	downlinks DownlinksConfig,
) (Webhooks, error) {
	ctx = log.NewContextWithField(ctx, "namespace", namespace)
	w := &webhooks{
		ctx:       ctx,
		server:    server,
		registry:  registry,
		target:    target,
		downlinks: downlinks,
	}
	sub, err := server.Subscribe(ctx, "webhooks", nil, false)
	if err != nil {
		return nil, err
	}
	wp := workerpool.NewWorkerPool(workerpool.Config[*ttnpb.ApplicationUp]{
		Component: server,
		Context:   ctx,
		Name:      "webhooks_fanout",
		Handler:   workerpool.HandlerFromUplinkHandler(w.handleUp),
	})
	sub.Pipe(ctx, server, "webhooks", wp.Publish)
	return w, nil
}

func (w *webhooks) Registry() WebhookRegistry { return w.registry }

// RegisterRoutes registers the webhooks to the web server to handle downlink requests.
func (w *webhooks) RegisterRoutes(server *ttnweb.Server) {
	router := server.Prefix(
		ttnpb.HTTPAPIPrefix + "/as/applications/{application_id}/webhooks/{webhook_id}/devices/{device_id}/down",
	).Subrouter()
	router.Use(
		mux.MiddlewareFunc(webmiddleware.Namespace("applicationserver/io/web")),
		mux.MiddlewareFunc(webmiddleware.Metadata("Authorization")),
		w.validateAndFillIDs,
		w.requireApplicationRights(ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_DOWN_WRITE),
		w.requireRateLimits(),
	)

	router.Handle("/push", w.handleDown(io.Server.DownlinkQueuePush)).Methods(http.MethodPost)
	router.Handle("/replace", w.handleDown(io.Server.DownlinkQueueReplace)).Methods(http.MethodPost)
}

func (w *webhooks) handleUp(ctx context.Context, msg *ttnpb.ApplicationUp) error {
	ctx = log.NewContextWithField(ctx, "namespace", namespace)
	hooks, err := w.registry.List(ctx, msg.EndDeviceIds.ApplicationIds, webhookFanOutFieldMask)
	if err != nil {
		return err
	}
	wg := sync.WaitGroup{}
	for i := range hooks {
		hook := hooks[i]
		ctx := internal.WithWebhookData(ctx, &internal.WebhookData{
			EndDeviceIDs: msg.EndDeviceIds,
			WebhookIDs:   hook.Ids,
			Health:       hook.HealthStatus,
		})
		ctx = log.NewContextWithField(ctx, "hook", hook.Ids.WebhookId)

		if hook.Paused {
			log.FromContext(ctx).Debug("Webhook is paused")
			continue
		}

		f := func(ctx context.Context) error {
			req, err := NewRequest(ctx, w.downlinks, msg, hook)
			if err != nil {
				log.FromContext(ctx).WithError(err).Warn("Failed to create request")
				return err
			}
			if req == nil {
				return nil
			}
			log.FromContext(ctx).WithField("url", req.URL).Debug("Process request")
			if err := w.target.Process(req); err != nil {
				log.FromContext(ctx).WithError(err).Warn("Failed to process request")
				return err
			}
			return nil
		}
		wg.Add(1)
		w.server.StartTask(&task.Config{
			Context: ctx,
			ID:      "execute_webhook",
			Func:    f,
			Done:    wg.Done,
			Restart: task.RestartNever,
			Backoff: task.DefaultBackoffConfig,
		})
	}
	wg.Wait()
	return nil
}

var (
	errWebhookNotFound = errors.DefineNotFound("webhook_not_found", "webhook not found")
	errReadBody        = errors.DefineCanceled("read_body", "read body")
	errDecodeBody      = errors.DefineInvalidArgument("decode_body", "decode body")
	errValidateBody    = errors.DefineInvalidArgument("validate_body", "validate body")
)

func (w *webhooks) handleDown(
	op func(io.Server, context.Context, *ttnpb.EndDeviceIdentifiers, []*ttnpb.ApplicationDownlink) error,
) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		devID := internal.DeviceIDFromContext(ctx)
		hookID := internal.WebhookIDFromContext(ctx)
		logger := log.FromContext(ctx).WithFields(log.Fields(
			"application_id", devID.ApplicationIds.ApplicationId,
			"device_id", devID.DeviceId,
			"webhook_id", hookID.WebhookId,
		))

		hook, err := w.registry.Get(ctx, hookID, []string{"format", "paused"})
		if err != nil {
			webhandlers.Error(res, req, err)
			return
		}
		if hook == nil {
			webhandlers.Error(res, req, errWebhookNotFound.New())
			return
		}
		if hook.Paused {
			logger.Debug("Webhook is paused")
			res.WriteHeader(http.StatusNotAcceptable)
			return
		}
		format, ok := formats[hook.Format]
		if !ok {
			webhandlers.Error(res, req, errFormatNotFound.WithAttributes("format", hook.Format))
			return
		}
		body, err := stdio.ReadAll(req.Body)
		if err != nil {
			webhandlers.Error(res, req, errReadBody.WithCause(err))
			return
		}
		items, err := format.ToDownlinks(body)
		if err != nil {
			webhandlers.Error(res, req, errDecodeBody.WithCause(err))
			return
		}
		if err := items.ValidateFields(); err != nil {
			webhandlers.Error(res, req, errValidateBody.WithCause(err))
			return
		}
		logger.Debug("Perform downlink queue operation")
		if err := op(w.server, ctx, devID, items.Downlinks); err != nil {
			webhandlers.Error(res, req, err)
			return
		}

		res.WriteHeader(http.StatusOK)
	})
}
