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

package tracing

import "go.thethings.network/lorawan-stack/v3/pkg/config/tlsconfig"

// Config represents configuration for OpenTelemetry tracing.
type Config struct {
	Enable            bool            `name:"enable" description:"Enable telemetry"`
	Exporter          string          `name:"exporter" description:"Telemetry exporter (otlp, writer)"`
	CollectorConfig   CollectorConfig `name:"collector-config" description:"Trace collector exporter configuration"`
	WriterConfig      WriterConfig    `name:"writer-config" description:"Writer exporter configuration"`
	SampleProbability float64         `name:"sample-probability" description:"Sampling probability. Fractions >= 1 will always sample. Fractions < 0 are treated as zero"` //nolint:lll
}

// CollectorConfig represents configuration for the trace collector exporter.
type CollectorConfig struct {
	EndpointURL string           `name:"endpoint-url" description:"The URL of the collector endpoint"`
	Insecure    bool             `name:"insecure" description:"Use insecure connection"`
	TLS         tlsconfig.Client `name:"tls"`
}

// WriterConfig represents configuration for the stdout exporter.
type WriterConfig struct {
	Destination string `name:"destination" description:"Destination of telemetry writer (stdout, stderr)"`
	Timestamps  bool   `name:"timestamps" description:"Print timestamps"`
	Pretty      bool   `name:"pretty" description:"Human readable format"`
}
