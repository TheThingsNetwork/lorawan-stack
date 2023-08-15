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

// Package cli contains telemetry functions regarding the collection of data in the CLI. Should be imported as
// cli_telemetry.
package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/google/uuid"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	telemetry "go.thethings.network/lorawan-stack/v3/pkg/telemetry/exporter"
	"go.thethings.network/lorawan-stack/v3/pkg/telemetry/exporter/models"
	"gopkg.in/yaml.v2"
)

// defaultTimeout sends CLI data every 24 hours.
var defaultTimeout = 24 * time.Hour

// cliStateTelemetry contains relevant data to the execution of the CLI telemetry.
//
// The CLI telemetry is essentially different from other telemetry collections of the Stack, this is due to the fact
// that we don't have to consider the possibility of having multiple servers for each component of the Stack. With this
// information we can simply just use a unique randomly generated number as an ID, writing it on the local machine.
type cliStateTelemetry struct {
	UID      string    `yaml:"uid"`
	LastSent time.Time `yaml:"last_sent"`
}

// Write telemetry state into the provided path.
func (cst *cliStateTelemetry) Write(p string) error {
	b, err := yaml.Marshal(cst)
	if err != nil {
		return err
	}

	// Creates the file since it does not exist.
	if err := os.WriteFile(p, b, 0o644); err != nil { //nolint:gas
		return err
	}
	return nil
}

// ttnPath returns the path to the folder in which the configuration of telemetry is stored.
func ttnPath() (string, error) {
	configPath, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return path.Join(configPath, "ttn-lw-cli"), nil
}

// cliStatePath returns the path to the file in which the CLI telemetry is stored.
func cliStatePath() (string, error) {
	p, err := ttnPath()
	if err != nil {
		return "", err
	}
	return path.Join(p, "telemetry.yml"), nil
}

// getCLIState returns the telemetry state, if it doesn't exist it creates a file with default values for the state.
func getCLIState() (*cliStateTelemetry, bool, error) {
	folderPath, err := ttnPath()
	if err != nil {
		return nil, false, err
	}
	if _, err := os.Stat(folderPath); err != nil {
		if !os.IsNotExist(err) {
			return nil, false, err
		}
		// Creates the folder since it does not exist.
		if err := os.Mkdir(folderPath, 0o755); err != nil {
			return nil, false, err
		}
	}

	statePath, err := cliStatePath()
	if err != nil {
		return nil, false, err
	}

	cliState := &cliStateTelemetry{}

	if _, err := os.Stat(statePath); err != nil {
		if !os.IsNotExist(err) {
			return nil, false, err
		}

		cliState.UID = uuid.NewString()
		cliState.LastSent = time.Now()
		return cliState, true, nil
	}

	b, err := os.ReadFile(statePath)
	if err != nil {
		return nil, false, err
	}
	if err := yaml.Unmarshal(b, cliState); err != nil {
		return nil, false, err
	}
	return cliState, false, nil
}

func shouldSendTelemetry(t time.Time) bool {
	return time.Now().After(t.Add(defaultTimeout))
}

type cliTask struct {
	target string
}

// Task is a small task that sends telemetry data once a day.
type Task interface {
	Run(context.Context)
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

// NewCLITelemetry returns a wrapper that contains the necessary methods to collect and send telemetry data regarding
// CLI usage.
func NewCLITelemetry(opts ...Option) Task {
	ct := &cliTask{}
	for _, opt := range opts {
		opt.apply(ct)
	}
	return ct
}

// Run a small task that send telemetry data once a day.
func (ct *cliTask) Run(ctx context.Context) {
	logger := log.FromContext(ctx)
	state, send, err := getCLIState()
	if err != nil {
		logger.WithError(err).Debug("Failed to retrieve telemetry state")
	}

	if send || shouldSendTelemetry(state.LastSent) {
		state.LastSent = time.Now()
		if statePath, err := cliStatePath(); err != nil {
			logger.WithError(err).Debug("Failed to retrieve telemetry state file path")
		} else {
			// If no error fetching the state file path, update the last sent timestamp.
			state.Write(statePath) //nolint:errcheck
		}

		b, err := json.Marshal(&models.TelemetryMessage{
			UID: state.UID,
			OS:  telemetry.OSTelemetryData(),
			CLI: &models.CLITelemetry{},
		})
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
		io.Copy(io.Discard, resp.Body) // nolint:errcheck
		logger.Info("Sent telemetry information")
	}
}
