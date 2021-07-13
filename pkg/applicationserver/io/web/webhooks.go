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
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"

	stdio "io"

	"github.com/gorilla/mux"
	"github.com/jtacoma/uritemplates"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	ttnweb "go.thethings.network/lorawan-stack/v3/pkg/web"
	"go.thethings.network/lorawan-stack/v3/pkg/webhandlers"
	"go.thethings.network/lorawan-stack/v3/pkg/webmiddleware"
)

const namespace = "applicationserver/io/web"

// Sink processes HTTP requests.
type Sink interface {
	Process(*http.Request) error
}

// ControllableSink is a controllable Sink.
type ControllableSink interface {
	Sink
	Run(context.Context) error
}

// HTTPClientSink contains an HTTP client to make outgoing requests.
type HTTPClientSink struct {
	*http.Client
}

var errRequest = errors.DefineUnavailable("request", "request failed with status `{code}`")

// Process uses the HTTP client to perform the request.
func (s *HTTPClientSink) Process(req *http.Request) error {
	res, err := s.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		stdio.Copy(ioutil.Discard, res.Body)
		res.Body.Close()
	}()
	if res.StatusCode >= 200 && res.StatusCode <= 299 {
		return nil
	}
	return errRequest.WithAttributes("code", res.StatusCode)
}

// QueuedSink is a ControllableSink with queue.
type QueuedSink struct {
	Target  Sink
	Queue   chan *http.Request
	Workers int
}

// Run starts concurrent workers to process messages from the queue.
// If Target is a ControllableSink, this method runs the target.
// This method blocks until the target (if controllable) and all workers are done.
func (s *QueuedSink) Run(ctx context.Context) error {
	if s.Workers < 1 {
		s.Workers = 1
	}
	wg := sync.WaitGroup{}
	if controllable, ok := s.Target.(ControllableSink); ok {
		wg.Add(1)
		go func() {
			if err := controllable.Run(ctx); err != nil && !errors.IsCanceled(err) {
				log.FromContext(ctx).WithError(err).Error("Target sink failed")
			}
			wg.Done()
		}()
	}
	for i := 0; i < s.Workers; i++ {
		wg.Add(1)
		go func() {
			for {
				select {
				case <-ctx.Done():
					wg.Done()
					return
				case req := <-s.Queue:
					registerWebhookDequeued()
					ctx := req.Context()
					if err := s.Target.Process(req); err != nil {
						registerWebhookFailed(err)
						log.FromContext(ctx).WithError(err).Warn("Failed to process message")
					} else {
						registerWebhookSent()
					}
				}
			}
		}()
	}
	<-ctx.Done()
	wg.Wait()
	return ctx.Err()
}

var errQueueFull = errors.DefineResourceExhausted("queue_full", "the queue is full")

// Process sends the request to the queue.
// This method returns immediately. An error is returned when the queue is full.
func (s *QueuedSink) Process(req *http.Request) error {
	select {
	case s.Queue <- req:
		registerWebhookQueued()
		return nil
	default:
		err := errQueueFull.New()
		registerWebhookFailed(err)
		return err
	}
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
	target    Sink
	downlinks DownlinksConfig
}

// NewWebhooks returns a new Webhooks.
func NewWebhooks(ctx context.Context, server io.Server, registry WebhookRegistry, target Sink, downlinks DownlinksConfig) (Webhooks, error) {
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
	server.StartTask(&component.TaskConfig{
		Context: ctx,
		ID:      "run_webhooks",
		Func: func(ctx context.Context) error {
			for {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-sub.Context().Done():
					return sub.Context().Err()
				case up := <-sub.Up():
					ctx := log.NewContextWithField(up.Context, "namespace", namespace)
					if err := w.handleUp(ctx, up.ApplicationUp); err != nil {
						log.FromContext(ctx).WithError(err).Warn("Failed to handle message")
					}
				}
			}
		},
		Restart: component.TaskRestartOnFailure,
		Backoff: component.DefaultTaskBackoffConfig,
	})
	return w, nil
}

func (w *webhooks) Registry() WebhookRegistry { return w.registry }

// RegisterRoutes registers the webhooks to the web server to handle downlink requests.
func (w *webhooks) RegisterRoutes(server *ttnweb.Server) {
	router := server.Prefix(ttnpb.HTTPAPIPrefix + "/as/applications/{application_id}/webhooks/{webhook_id}/devices/{device_id}/down").Subrouter()
	router.Use(
		mux.MiddlewareFunc(webmiddleware.Namespace("applicationserver/io/web")),
		mux.MiddlewareFunc(webmiddleware.Metadata("Authorization")),
		w.validateAndFillIDs,
		w.requireApplicationRights(ttnpb.RIGHT_APPLICATION_TRAFFIC_DOWN_WRITE),
		w.requireRateLimits(),
	)

	router.Handle("/push", w.handleDown(io.Server.DownlinkQueuePush)).Methods(http.MethodPost)
	router.Handle("/replace", w.handleDown(io.Server.DownlinkQueueReplace)).Methods(http.MethodPost)
}

const (
	downlinkKeyHeader     = "X-Downlink-Apikey"
	downlinkPushHeader    = "X-Downlink-Push"
	downlinkReplaceHeader = "X-Downlink-Replace"

	downlinkOperationURLFormat = "%s/as/applications/%s/webhooks/%s/devices/%s/down/%s"

	domainHeader = "X-Tts-Domain"
)

func (w *webhooks) createDownlinkURL(ctx context.Context, webhookID ttnpb.ApplicationWebhookIdentifiers, devID ttnpb.EndDeviceIdentifiers, op string) string {
	downlinks := w.downlinks
	baseURL := downlinks.PublicTLSAddress
	if baseURL == "" {
		baseURL = downlinks.PublicAddress
	}
	return fmt.Sprintf(downlinkOperationURLFormat,
		baseURL,
		webhookID.ApplicationId,
		webhookID.WebhookId,
		devID.DeviceId,
		op,
	)
}

func (w *webhooks) createDomain(ctx context.Context) string {
	downlinks := w.downlinks
	baseURL := downlinks.PublicTLSAddress
	if baseURL == "" {
		baseURL = downlinks.PublicAddress
	}
	u, err := url.Parse(baseURL)
	if err != nil {
		return ""
	}
	return u.Host
}

func (w *webhooks) handleUp(ctx context.Context, msg *ttnpb.ApplicationUp) error {
	hooks, err := w.registry.List(ctx, msg.ApplicationIdentifiers,
		[]string{
			"base_url",
			"downlink_ack",
			"downlink_api_key",
			"downlink_failed",
			"downlink_nack",
			"downlink_queued",
			"downlink_queue_invalidated",
			"downlink_sent",
			"format",
			"headers",
			"join_accept",
			"location_solved",
			"service_data",
			"uplink_message",
		},
	)
	if err != nil {
		return err
	}
	wg := sync.WaitGroup{}
	for i := range hooks {
		hook := hooks[i]
		logger := log.FromContext(ctx).WithField("hook", hook.WebhookId)
		wg.Add(1)
		go func() {
			defer wg.Done()
			req, err := w.newRequest(ctx, msg, hook)
			if err != nil {
				logger.WithError(err).Warn("Failed to create request")
				return
			}
			if req == nil {
				return
			}
			logger.WithField("url", req.URL).Debug("Process message")
			if err := w.target.Process(req); err != nil {
				logger.WithError(err).Warn("Failed to process message")
			}
		}()
	}
	wg.Wait()
	return nil
}

func (w *webhooks) newRequest(ctx context.Context, msg *ttnpb.ApplicationUp, hook *ttnpb.ApplicationWebhook) (*http.Request, error) {
	var cfg *ttnpb.ApplicationWebhook_Message
	switch msg.Up.(type) {
	case *ttnpb.ApplicationUp_UplinkMessage:
		cfg = hook.UplinkMessage
	case *ttnpb.ApplicationUp_JoinAccept:
		cfg = hook.JoinAccept
	case *ttnpb.ApplicationUp_DownlinkAck:
		cfg = hook.DownlinkAck
	case *ttnpb.ApplicationUp_DownlinkNack:
		cfg = hook.DownlinkNack
	case *ttnpb.ApplicationUp_DownlinkSent:
		cfg = hook.DownlinkSent
	case *ttnpb.ApplicationUp_DownlinkFailed:
		cfg = hook.DownlinkFailed
	case *ttnpb.ApplicationUp_DownlinkQueued:
		cfg = hook.DownlinkQueued
	case *ttnpb.ApplicationUp_DownlinkQueueInvalidated:
		cfg = hook.DownlinkQueueInvalidated
	case *ttnpb.ApplicationUp_LocationSolved:
		cfg = hook.LocationSolved
	case *ttnpb.ApplicationUp_ServiceData:
		cfg = hook.ServiceData
	}
	if cfg == nil {
		return nil, nil
	}
	baseURL, err := expandVariables(hook.BaseUrl, msg)
	if err != nil {
		return nil, err
	}
	pathURL, err := expandVariables(cfg.Path, msg)
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(pathURL.Path, "/") {
		// Trim the leading slash, in order to ensure that the path is not
		// interpreted as relative to the root of the URL.
		pathURL.Path = strings.TrimLeft(pathURL.Path, "/")
		// Add the "/" suffix here instead of the condition below in order
		// to treat the case in which the pathURL.Path is "/".
		if !strings.HasSuffix(baseURL.Path, "/") {
			baseURL.Path += "/"
		}
	}
	// If the path URL contains an actual path (i.e. is not only a query)
	// ensure that it does not override the top level path.
	if pathURL.Path != "" && !strings.HasSuffix(baseURL.Path, "/") {
		baseURL.Path += "/"
	}
	finalURL := baseURL.ResolveReference(pathURL)
	format, ok := formats[hook.Format]
	if !ok {
		return nil, errFormatNotFound.WithAttributes("format", hook.Format)
	}
	buf, err := format.FromUp(msg)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, finalURL.String(), bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}
	for key, value := range hook.Headers {
		req.Header.Set(key, value)
	}
	if hook.DownlinkApiKey != "" {
		req.Header.Set(downlinkKeyHeader, hook.DownlinkApiKey)
		req.Header.Set(downlinkPushHeader, w.createDownlinkURL(ctx, hook.ApplicationWebhookIdentifiers, msg.EndDeviceIdentifiers, "push"))
		req.Header.Set(downlinkReplaceHeader, w.createDownlinkURL(ctx, hook.ApplicationWebhookIdentifiers, msg.EndDeviceIdentifiers, "replace"))
	}
	if domain := w.createDomain(ctx); domain != "" {
		req.Header.Set(domainHeader, domain)
	}
	req.Header.Set("Content-Type", format.ContentType)
	return req, nil
}

var (
	errWebhookNotFound = errors.DefineNotFound("webhook_not_found", "webhook not found")
	errReadBody        = errors.DefineCanceled("read_body", "read body")
	errDecodeBody      = errors.DefineInvalidArgument("decode_body", "decode body")
)

func (w *webhooks) handleDown(op func(io.Server, context.Context, ttnpb.EndDeviceIdentifiers, []*ttnpb.ApplicationDownlink) error) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		devID := deviceIDFromContext(ctx)
		hookID := webhookIDFromContext(ctx)
		logger := log.FromContext(ctx).WithFields(log.Fields(
			"application_id", devID.ApplicationId,
			"device_id", devID.DeviceId,
			"webhook_id", hookID.WebhookId,
		))

		hook, err := w.registry.Get(ctx, hookID, []string{"format"})
		if err != nil {
			webhandlers.Error(res, req, err)
			return
		}
		if hook == nil {
			webhandlers.Error(res, req, errWebhookNotFound.New())
			return
		}
		format, ok := formats[hook.Format]
		if !ok {
			webhandlers.Error(res, req, errFormatNotFound.WithAttributes("format", hook.Format))
			return
		}
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			webhandlers.Error(res, req, errReadBody.WithCause(err))
			return
		}
		items, err := format.ToDownlinks(body)
		if err != nil {
			webhandlers.Error(res, req, errDecodeBody.WithCause(err))
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

func expandVariables(u string, up *ttnpb.ApplicationUp) (*url.URL, error) {
	var joinEUI, devEUI, devAddr string
	if up.JoinEui != nil {
		joinEUI = up.JoinEui.String()
	}
	if up.DevEui != nil {
		devEUI = up.DevEui.String()
	}
	if up.DevAddr != nil {
		devAddr = up.DevAddr.String()
	}
	tmpl, err := uritemplates.Parse(u)
	if err != nil {
		return nil, err
	}
	expanded, err := tmpl.Expand(map[string]interface{}{
		"appID":         up.ApplicationId,
		"applicationID": up.ApplicationId,
		"appEUI":        joinEUI,
		"joinEUI":       joinEUI,
		"devID":         up.DeviceId,
		"deviceID":      up.DeviceId,
		"devEUI":        devEUI,
		"devAddr":       devAddr,
	})
	if err != nil {
		return nil, err
	}
	return url.Parse(expanded)
}
