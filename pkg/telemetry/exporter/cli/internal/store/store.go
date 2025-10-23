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

// Package store handles the interaction with the SQLite database associated with the CLI.
package store

import (
	"encoding/json"
	"io"
	"os"
	"time"

	"github.com/google/uuid"
	"go.thethings.network/lorawan-stack/v3/pkg/telemetry/exporter/models"
)

// Store defines the available operations to handle state management.
type Store interface {
	State
	UsageData
}

type data struct {
	State StateData      `json:"state"`
	Cmd   models.CmdData `json:"cmd"`
}

// dbStore represents the SQLite database connection.
type dbStore struct{ filepath string }

func (st *dbStore) Read() (*data, error) {
	f, err := os.Open(st.filepath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	d := &data{}
	if err := json.Unmarshal(b, d); err != nil {
		return nil, err
	}

	return d, nil
}

func (st *dbStore) Write(d *data) error {
	b, err := json.Marshal(d)
	if err != nil {
		return err
	}

	return os.WriteFile(st.filepath, b, 0o600)
}

// New opens a connection to the SQLite database and creates the necessary tables if they don't exist.
func New() (Store, error) {
	folderPath, err := ttnPath()
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(folderPath); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		// Creates the folder since it does not exist.
		if err := os.Mkdir(folderPath, 0o750); err != nil {
			return nil, err
		}
	}

	p, err := dbPath()
	if err != nil {
		return nil, err
	}

	// Check if the file exists
	_, err = os.Stat(p)
	if os.IsNotExist(err) {
		bytes, err := json.Marshal(
			&data{State: StateData{
				UID:      uuid.New().String(),
				LastSent: time.Time{},
			}},
		)
		if err != nil {
			return nil, err
		}

		file, err := os.Create(p)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		_, err = file.Write(bytes)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		// An unexpected error happened. Most likely related to perms to file creation.
		return nil, err
	}

	return &dbStore{p}, nil
}
