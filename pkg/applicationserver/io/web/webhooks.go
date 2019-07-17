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
	stdio "io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"

	echo "github.com/labstack/echo/v4"
	"go.thethings.network/lorawan-stack/pkg/applicationserver/io"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/errors"
	web_errors "go.thethings.network/lorawan-stack/pkg/errors/web"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/version"
	ttnweb "go.thethings.network/lorawan-stack/pkg/web"
	"google.golang.org/api/googleapi"
	"google.golang.org/grpc/metadata"
)

var userAgent = "ttn-lw-application-server/" + version.TTN

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
					if err := s.Target.Process(req); err != nil {
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

var errQueueFull = errors.DefineResourceExhausted("queue_full", "the queue is full")

// Process sends the request to the queue.
// This method returns immediately. An error is returned when the queue is full.
func (s *QueuedSink) Process(req *http.Request) error {
	select {
	case s.Queue <- req:
		return nil
	default:
		return errQueueFull
	}
}

// Webhooks is an interface for registering incoming webhooks for downlink and creating a subscription to outgoing
// webhooks for upstream data.
type Webhooks interface {
	ttnweb.Registerer
	Registry() WebhookRegistry
	// NewSubscription returns a new webhooks integration subscription.
	NewSubscription() *io.Subscription
}

type webhooks struct {
	ctx      context.Context
	server   io.Server
	registry WebhookRegistry
	target   Sink
}

// NewWebhooks returns a new Webhooks.
func NewWebhooks(ctx context.Context, server io.Server, registry WebhookRegistry, target Sink) Webhooks {
	ctx = log.NewContextWithField(ctx, "namespace", "applicationserver/io/web")
	return &webhooks{
		ctx:      ctx,
		server:   server,
		registry: registry,
		target:   target,
	}
}

func (w *webhooks) Registry() WebhookRegistry { return w.registry }

// RegisterRoutes registers the webhooks to the web server to handle downlink requests.
func (w *webhooks) RegisterRoutes(server *ttnweb.Server) {
	middleware := []echo.MiddlewareFunc{
		w.handleError(),
		w.validateAndFillIDs(),
		w.requireApplicationRights(ttnpb.RIGHT_APPLICATION_TRAFFIC_DOWN_WRITE),
	}
	group := server.Group(ttnpb.HTTPAPIPrefix+"/as/applications/:application_id/webhooks/:webhook_id/devices/:device_id/down", middleware...)
	group.POST("/push", func(c echo.Context) error {
		return w.handleDown(c, io.Server.DownlinkQueuePush)
	})
	group.POST("/replace", func(c echo.Context) error {
		return w.handleDown(c, io.Server.DownlinkQueueReplace)
	})
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
			statusCode, err := web_errors.ProcessError(err)
			if strings.Contains(c.Request().Header.Get(echo.HeaderAccept), "application/json") {
				return c.JSON(statusCode, err)
			}
			return c.String(statusCode, err.Error())
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
			if err := appID.ValidateContext(w.ctx); err != nil {
				return err
			}
			c.Set(applicationIDKey, appID)

			devID := ttnpb.EndDeviceIdentifiers{
				ApplicationIdentifiers: appID,
				DeviceID:               c.Param(deviceIDKey),
			}
			if err := devID.ValidateContext(w.ctx); err != nil {
				return err
			}
			c.Set(deviceIDKey, devID)

			hookID := ttnpb.ApplicationWebhookIdentifiers{
				ApplicationIdentifiers: appID,
				WebhookID:              c.Param(webhookIDKey),
			}
			if err := hookID.ValidateFields(); err != nil {
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
			ctx := w.server.FillContext(c.Request().Context())
			ctx = log.NewContextWithField(ctx, "namespace", "applicationserver/io/web")

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
				if err := w.handleUp(msg.Context, msg.ApplicationUp); err != nil {
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
			"downlink_ack",
			"downlink_failed",
			"downlink_nack",
			"downlink_queued",
			"downlink_sent",
			"format",
			"headers",
			"join_accept",
			"location_solved",
			"uplink_message",
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
	expandVariables(url, msg)
	if err != nil {
		return nil, err
	}
	format, ok := formats[hook.Format]
	if !ok {
		return nil, errFormatNotFound.WithAttributes("format", hook.Format)
	}
	buf, err := format.FromUp(msg)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, url.String(), bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}
	for key, value := range hook.Headers {
		req.Header.Set(key, value)
	}
	req.Header.Set("Content-Type", format.ContentType)
	req.Header.Set("User-Agent", userAgent)
	return req, nil
}

var errWebhookNotFound = errors.DefineNotFound("webhook_not_found", "webhook not found")

func (w *webhooks) handleDown(c echo.Context, op func(io.Server, context.Context, ttnpb.EndDeviceIdentifiers, []*ttnpb.ApplicationDownlink) error) error {
	ctx := w.server.FillContext(c.Request().Context())
	devID := c.Get(deviceIDKey).(ttnpb.EndDeviceIdentifiers)
	hookID := c.Get(webhookIDKey).(ttnpb.ApplicationWebhookIdentifiers)
	logger := log.FromContext(ctx).WithFields(log.Fields(
		"application_id", devID.ApplicationID,
		"device_id", devID.DeviceID,
		"webhook_id", hookID.WebhookID,
	))
	hook, err := w.registry.Get(ctx, hookID, []string{"format"})
	if err != nil {
		return err
	}
	if hook == nil {
		return errWebhookNotFound
	}
	format, ok := formats[hook.Format]
	if !ok {
		return errFormatNotFound.WithAttributes("format", hook.Format)
	}
	body, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return err
	}
	items, err := format.ToDownlinks(body)
	if err != nil {
		return err
	}
	logger.Debug("Perform downlink queue operation")
	if err := op(w.server, ctx, devID, items.Downlinks); err != nil {
		return err
	}
	return nil
}

func expandVariables(url *url.URL, up *ttnpb.ApplicationUp) {
	var joinEUI, devEUI, devAddr string
	if up.JoinEUI != nil {
		joinEUI = up.JoinEUI.String()
	}
	if up.DevEUI != nil {
		devEUI = up.DevEUI.String()
	}
	if up.DevAddr != nil {
		devAddr = up.DevAddr.String()
	}
	googleapi.Expand(url, map[string]string{
		"appID":   up.ApplicationID,
		"appEUI":  joinEUI,
		"joinEUI": joinEUI,
		"devID":   up.DeviceID,
		"devEUI":  devEUI,
		"devAddr": devAddr,
	})
}
