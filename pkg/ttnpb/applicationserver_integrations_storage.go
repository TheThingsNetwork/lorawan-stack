// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

package ttnpb

import "context"

// ValidateContext wraps the generated validator with (optionally context-based) custom checks.
func (m *GetStoredApplicationUpRequest) ValidateContext(context.Context) error {
	return m.ValidateFields()
}

// StoredApplicationUpTypes is a list of available ApplicationUp message types.
var StoredApplicationUpTypes = map[string]struct{}{
	"":                           {},
	"uplink_message":             {},
	"join_accept":                {},
	"downlink_ack":               {},
	"downlink_nack":              {},
	"downlink_sent":              {},
	"downlink_failed":            {},
	"downlink_queued":            {},
	"downlink_queue_invalidated": {},
	"location_solved":            {},
	"service_data":               {},
}

// WithEndDeviceIDs returns the request with set EndDeviceIdentifiers
func (m *GetStoredApplicationUpRequest) WithEndDeviceIDs(ids *EndDeviceIdentifiers) *GetStoredApplicationUpRequest {
	m.EndDeviceIDs = ids
	return m
}

// WithApplicationIDs returns the request with set ApplicationIdentifiers
func (m *GetStoredApplicationUpRequest) WithApplicationIDs(ids *ApplicationIdentifiers) *GetStoredApplicationUpRequest {
	m.ApplicationIDs = ids
	return m
}
