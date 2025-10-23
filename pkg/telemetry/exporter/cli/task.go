// Copyright © 2025 The Things Network Foundation, The Things Industries B.V.
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

// Package cli defines the telemetry task interface and its implementation.
package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/spf13/cobra"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	telemetry "go.thethings.network/lorawan-stack/v3/pkg/telemetry/exporter"
	"go.thethings.network/lorawan-stack/v3/pkg/telemetry/exporter/cli/internal/store"
	"go.thethings.network/lorawan-stack/v3/pkg/telemetry/exporter/models"
)

// defaultTimeout sends CLI data every 24 hours.
var defaultTimeout = 24 * time.Hour

func shouldSendTelemetry(st *store.StateData) bool {
	return st != nil && time.Now().After(st.LastSent.Add(defaultTimeout))
}

// Option is an option for the CLI telemetry.
type Option interface {
	apply(*cliTask)
}

type option func(*cliTask)

func (opt option) apply(ct *cliTask) { opt(ct) }

// WithCLITarget defines the URL to which the CLI data will be sent.
func WithCLITarget(s string) Option {
	return option(func(ct *cliTask) {
		ct.target = s
	})
}

// Task is a small task that sends telemetry data once a day.
type Task interface {
	SendData(ctx context.Context)
	SaveData(ctx context.Context, cmd *cobra.Command)
}

type cliTask struct{ target string }

// NewCLITelemetry returns a wrapper that contains the necessary methods to collect and send telemetry data regarding
// CLI usage.
func NewCLITelemetry(opts ...Option) Task {
	ct := &cliTask{}
	for _, opt := range opts {
		opt.apply(ct)
	}
	return ct
}

// SendData a small task that sends telemetry data.
func (ct *cliTask) SendData(ctx context.Context) {
	logger := log.FromContext(ctx)

	st, err := store.New()
	if err != nil {
		logger.WithError(err).Debug("Failed to open telemetry database")
		return
	}

	state, err := st.GetState()
	if err != nil {
		logger.WithError(err).Debug("Failed to retrieve telemetry state")
		return // Skip the telemetry procedure if an error is found creating or fetching the previous state.
	}

	if !shouldSendTelemetry(state) {
		return
	}

	cmdData, err := st.GetCmdData()
	if err != nil {
		logger.WithError(err).Debug("Failed to fetch commands usage data")
		return
	}

	data := &models.TelemetryMessage{
		UID: state.UID,
		CLI: &models.CLITelemetry{CmdData: *cmdData},
		OS:  telemetry.OSTelemetryData(),
	}

	b, err := json.Marshal(data)
	if err != nil {
		logger.WithError(err).Debug("Failed to marshal telemetry information")
		return
	}

	resp, err := http.DefaultClient.Post(ct.target, "application/json", bytes.NewReader(b))
	if err != nil {
		logger.WithError(err).Debug("Failed to send telemetry information")
		return
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body) // nolint:errcheck,gosec

	// Operations after sending the telemetry data
	if err := st.UpdateLastSent(); err != nil {
		logger.WithError(err).Debug("Failed to update last_sent field in telemetry state")
		return
	}

	if err := st.ResetCmdData(); err != nil {
		logger.WithError(err).Debug("Failed to reset command usage data")
		return
	}
}

// SaveData process and stores data related to the CLI usage.
func (*cliTask) SaveData(ctx context.Context, cmd *cobra.Command) {
	logger := log.FromContext(ctx)

	st, err := store.New()
	if err != nil {
		logger.WithError(err).Debug("Failed to open telemetry database")
		return
	}

	if err := st.StoreCommand(cmd); err != nil {
		logger.WithError(err).Debug("Failed to store command usage data")
		return
	}
}
