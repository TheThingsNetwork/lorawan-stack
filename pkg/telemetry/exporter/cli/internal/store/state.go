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

package store

import (
	"time"
)

// State defines the available operations to handle state management.
type State interface {
	GetState() (*StateData, error)
	UpdateLastSent() error
}

// StateData contains relevant data to the execution of the CLI telemetry.
type StateData struct {
	UID      string    `json:"uid"`
	LastSent time.Time `json:"last_sent"`
}

// GetState returns the telemetry state, if it doesn't exist it creates a new state.
func (st *dbStore) GetState() (*StateData, error) {
	data, err := st.Read()
	if err != nil {
		return nil, err
	}
	return &data.State, nil
}

// UpdateLastSent updates the last_sent timestamp to the current time.
func (st *dbStore) UpdateLastSent() error {
	data, err := st.Read()
	if err != nil {
		return err
	}

	data.State.LastSent = time.Now()

	return st.Write(data)
}
