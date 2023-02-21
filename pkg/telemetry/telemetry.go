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

// Package telemetry provides tools for working with tracing.
package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	otrace "go.opentelemetry.io/otel/trace"
	"go.thethings.network/lorawan-stack/v3/pkg/version"
)

func initResource(ctx context.Context) (*resource.Resource, error) {
	rsc, err := resource.New(ctx,
		resource.WithContainer(),
		resource.WithProcess(),
		resource.WithOS(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String("ttn-lw-stack"),
			semconv.ServiceVersionKey.String(version.String()),
		),
	)
	if err != nil {
		return nil, err
	}

	// Fill empty values with defaults
	return resource.Merge(resource.Default(), rsc)
}

// InitTelemetry initializes the telemetry package and returns the tracer provider.
// If telemetry is not enabled it returns a noop tracer provider instead.
func InitTelemetry(ctx context.Context, config *Config) (otrace.TracerProvider, func(context.Context) error, error) {
	if !config.Enable {
		return otrace.NewNoopTracerProvider(), func(_ context.Context) error { return nil }, nil
	}

	exp, err := exporterFromConfig(ctx, config)
	if err != nil {
		return nil, nil, err
	}
	bsp := sdktrace.NewBatchSpanProcessor(exp)

	rsc, err := initResource(ctx)
	if err != nil {
		return nil, nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(bsp),
		sdktrace.WithResource(rsc),
		sdktrace.WithSampler(sdktrace.ParentBased(
			sdktrace.TraceIDRatioBased(config.SampleProbability),
		)),
	)

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	return tp, tp.Shutdown, nil
}
