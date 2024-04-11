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

package sink

import (
	"context"
	"net/http"

	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/workerpool"
)

// pooledSink is a Sink with worker pool.
type pooledSink struct {
	pool workerpool.WorkerPool[*http.Request]
}

func createPoolHandler(sink Sink) workerpool.Handler[*http.Request] {
	h := func(ctx context.Context, req *http.Request) {
		if err := sink.Process(req); err != nil {
			log.FromContext(ctx).WithError(err).Warn("Failed to process requests")
		}
	}
	return h
}

// NewPooledSink creates a Sink that queues requests and processes them in parallel workers.
func NewPooledSink(ctx context.Context, c workerpool.Component, sink Sink, workers int, queueSize int) Sink {
	wp := workerpool.NewWorkerPool(workerpool.Config[*http.Request]{
		Component:  c,
		Context:    ctx,
		Name:       "webhooks",
		Handler:    createPoolHandler(sink),
		MaxWorkers: workers,
		QueueSize:  queueSize,
	})
	return &pooledSink{
		pool: wp,
	}
}

// Process sends the request to the workers.
// This method returns immediately. An error is returned when the workers are too busy.
func (s *pooledSink) Process(req *http.Request) error {
	ctx := req.Context()
	if err := s.pool.Publish(ctx, req); err != nil {
		registerWebhookFailed(ctx, err, false)
		return err
	}
	// The underlying sink should register the success, or final failure.
	return nil
}
