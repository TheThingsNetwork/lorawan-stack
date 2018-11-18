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
	"bytes"
	"context"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"

	"github.com/labstack/echo"
	"go.thethings.network/lorawan-stack/pkg/applicationserver/io"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	ttnweb "go.thethings.network/lorawan-stack/pkg/web"
	"google.golang.org/grpc/metadata"
)

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
	defer res.Body.Close()
	if res.StatusCode >= 200 && res.StatusCode <= 299 {
		return nil
	}
	return errRequest.WithAttributes("code", res.StatusCode)
}

// BufferedSink is a ControllableSink with buffer.
type BufferedSink struct {
	Target  Sink
	Buffer  chan *http.Request
	Workers int
}

// Run starts concurrent workers to process messages from the buffer.
// If Target is a ControllableSink, this method runs the target.
// This method blocks until the target (if controllable) and all workers are done.
func (b *BufferedSink) Run(ctx context.Context) error {
	if b.Workers < 1 {
		b.Workers = 1
	}
	wg := sync.WaitGroup{}
	if controllable, ok := b.Target.(ControllableSink); ok {
		wg.Add(1)
		go func() {
			if err := controllable.Run(ctx); err != nil && !errors.IsCanceled(err) {
				log.FromContext(ctx).WithError(err).Error("Target sink failed")
			}
			wg.Done()
		}()
	}
	for i := 0; i < b.Workers; i++ {
		wg.Add(1)
		go func() {
			for {
				select {
				case <-ctx.Done():
					wg.Done()
					return
				case req := <-b.Buffer:
					if err := b.Target.Process(req); err != nil {
						log.FromContext(ctx).WithError(err).Warn("Failed to process message")
					}
				}
			}
		}()
	}
	<-ctx.Done()
	wg.Wait()
	return ctx.Err()
}

var errBufferFull = errors.DefineResourceExhausted("buffer_full", "the buffer is full")

// Process sends the request to the buffer.
// This method returns immediately. An error is returned when the buffer is full.
func (b *BufferedSink) Process(req *http.Request) error {
	select {
	case b.Buffer <- req:
		return nil
	default:
		return errBufferFull
	}
}

// Webhooks is an interface for registering incoming webhooks for downlink and creating a subscription to outgoing
// webhooks for upstream data.
type Webhooks interface {
	ttnweb.Registerer
	// NewSubscription returns a new webhooks integration subscription.
	NewSubscription() *io.Subscription
}

type webhooks struct {
	ctx      context.Context
	registry WebhookRegistry
	target   Sink
}

// NewWebhooks returns a new Webhooks.
func NewWebhooks(ctx context.Context, registry WebhookRegistry, target Sink) Webhooks {
	ctx = log.NewContextWithField(ctx, "namespace", "applicationserver/io/web")
	return &webhooks{
		ctx:      ctx,
		registry: registry,
		target:   target,
	}
}

// RegisterRoutes registers the webhooks to the web server to handle downlink requests.
func (w *webhooks) RegisterRoutes(server *ttnweb.Server) {
	middleware := []echo.MiddlewareFunc{
		w.handleError(),
		w.validateAndFillIDs(),
		w.requireApplicationRights(ttnpb.RIGHT_APPLICATION_TRAFFIC_DOWN_WRITE),
	}
	group := server.Group("/as/applications/:application_id/webhooks/:webhook_id/down/:device_id", middleware...)
	_ = group
}

var errHTTP = errors.Define("http", "HTTP error: {message}")

func (w *webhooks) handleError() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)
			if err == nil || c.Response().Committed {
				return err
			}
			log.FromContext(w.ctx).WithError(err).Debug("HTTP request failed")
			status := http.StatusInternalServerError
			if echoErr, ok := err.(*echo.HTTPError); ok {
				status = echoErr.Code
				if ttnErr, ok := errors.From(echoErr.Internal); ok {
					if status == http.StatusInternalServerError {
						status = errors.ToHTTPStatusCode(ttnErr)
					}
					err = ttnErr
				}
			} else if ttnErr, ok := errors.From(err); ok {
				status = errors.ToHTTPStatusCode(ttnErr)
				err = ttnErr
			} else {
				err = errHTTP.WithCause(err).WithAttributes("message", err.Error())
			}
			if strings.Contains(c.Request().Header.Get(echo.HeaderAccept), "application/json") {
				return c.JSON(status, err)
			}
			return c.String(status, err.Error())
		}
	}
}

const (
	applicationIDKey = "application_id"
	deviceIDKey      = "device_id"
	webhookIDKey     = "webhook_id"
)

func (w *webhooks) validateAndFillIDs() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			appID := ttnpb.ApplicationIdentifiers{
				ApplicationID: c.Param(applicationIDKey),
			}
			if err := appID.Validate(); err != nil {
				return err
			}
			c.Set(applicationIDKey, appID)

			devID := ttnpb.EndDeviceIdentifiers{
				ApplicationIdentifiers: appID,
				DeviceID:               c.Param(deviceIDKey),
			}
			if err := devID.Validate(); err != nil {
				return err
			}
			c.Set(deviceIDKey, devID)

			hookID := ttnpb.ApplicationWebhookIdentifiers{
				ApplicationIdentifiers: appID,
				WebhookID:              c.Param(webhookIDKey),
			}
			if err := hookID.Validate(); err != nil {
				return err
			}
			c.Set(webhookIDKey, hookID)

			return next(c)
		}
	}
}

func (w *webhooks) requireApplicationRights(required ...ttnpb.Right) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := w.ctx
			appID := c.Get(applicationIDKey).(ttnpb.ApplicationIdentifiers)
			md := metadata.New(map[string]string{
				"id":            appID.ApplicationID,
				"authorization": c.Request().Header.Get(echo.HeaderAuthorization),
			})
			if ctxMd, ok := metadata.FromIncomingContext(ctx); ok {
				md = metadata.Join(ctxMd, md)
			}
			ctx = metadata.NewIncomingContext(ctx, md)
			if err := rights.RequireApplication(ctx, appID, required...); err != nil {
				return err
			}
			return next(c)
		}
	}
}

func (w *webhooks) NewSubscription() *io.Subscription {
	sub := io.NewSubscription(w.ctx, "webhook", nil)
	go func() {
		for {
			select {
			case <-w.ctx.Done():
				return
			case msg := <-sub.Up():
				if err := w.handleUp(w.ctx, msg); err != nil {
					log.FromContext(w.ctx).WithError(err).Warn("Failed to handle message")
				}
			}
		}
	}()
	return sub
}

func (w *webhooks) handleUp(ctx context.Context, msg *ttnpb.ApplicationUp) error {
	hooks, err := w.registry.List(ctx, msg.ApplicationIdentifiers,
		[]string{
			"base_url",
			"headers",
			"formatter",
			"uplink_message",
			"join_accept",
			"downlink_ack",
			"downlink_nack",
			"downlink_sent",
			"downlink_failed",
			"downlink_queued",
			"location_solved",
		},
	)
	if err != nil {
		return err
	}
	wg := sync.WaitGroup{}
	for i := range hooks {
		hook := hooks[i]
		logger := log.FromContext(ctx).WithField("hook", hook.WebhookID)
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
			logger.WithField("url", req.URL).Debug("Processing message")
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
	case *ttnpb.ApplicationUp_LocationSolved:
		cfg = hook.LocationSolved
	}
	if cfg == nil {
		return nil, nil
	}
	url, err := url.Parse(hook.BaseURL)
	if err != nil {
		return nil, err
	}
	url.Path = path.Join(url.Path, cfg.Path)
	formatter, ok := formatters[hook.Formatter]
	if !ok {
		return nil, errFormatterNotFound.WithAttributes("formatter", hook.Formatter)
	}
	buf, err := formatter.Encode(ctx, msg)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, url.String(), bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", formatter.ContentType())
	for key, value := range hook.Headers {
		req.Header.Set(key, value)
	}
	return req, nil
}
