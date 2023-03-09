// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

// Package tracer provides mechanisms to propagate tracer in context.
package tracer

import (
	"context"

	otrace "go.opentelemetry.io/otel/trace"
	"go.thethings.network/lorawan-stack/v3/pkg/telemetry/tracing"
)

type tracerKeyType struct{}

var tracerKey = &tracerKeyType{}

// NewContext returns a derived context with the tracer set.
func NewContext(ctx context.Context, t otrace.Tracer) context.Context {
	return context.WithValue(ctx, tracerKey, t)
}

// NewContextWithTracer returns a derived context with a new tracer set.
func NewContextWithTracer(ctx context.Context, name string, opts ...otrace.TracerOption) context.Context {
	t := tracing.FromContext(ctx).Tracer(name, opts...)
	return NewContext(ctx, t)
}

// FromContext returns the tracer that is attatched to the context
// or returns a new anonymous tracer if it does not exist.
func FromContext(ctx context.Context) otrace.Tracer {
	v := ctx.Value(tracerKey)
	if v == nil {
		return tracing.FromContext(ctx).Tracer("")
	}
	return v.(otrace.Tracer)
}

// StartFromContext returns a derived context with a new span started from the tracer in context.
func StartFromContext(ctx context.Context, name string, opts ...otrace.SpanStartOption) (context.Context, otrace.Span) {
	return FromContext(ctx).Start(ctx, name, opts...)
}
