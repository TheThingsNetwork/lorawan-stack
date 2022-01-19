// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

package api

import (
	"encoding/json"
	"strings"

	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// AbstractLocationSolverResult provides a unified interface to the location solver results.
type AbstractLocationSolverResult interface {
	Location() ttnpb.Location
	Algorithm() string
}

// AbstractLocationSolverResponse provides a unified interface to the location solver responses.
type AbstractLocationSolverResponse interface {
	Result() AbstractLocationSolverResult
	Errors() []string
	Warnings() []string

	Raw() *json.RawMessage
}

type abstractLocationSolverResult struct {
	location  ttnpb.Location
	algorithm string
}

// Location implements AbstractLocationSolverResult.
func (r abstractLocationSolverResult) Location() ttnpb.Location { return r.location }

// Algorithm implements AbstractLocationSolverResult.
func (r abstractLocationSolverResult) Algorithm() string { return r.algorithm }

type abstractLocationSolverResponse struct {
	result   AbstractLocationSolverResult
	errors   []string
	warnings []string
	raw      *json.RawMessage
}

// Result implements AbstractLocationSolverResponse.
func (r abstractLocationSolverResponse) Result() AbstractLocationSolverResult { return r.result }

// Errors implements AbstractLocationSolverResponse.
func (r abstractLocationSolverResponse) Errors() []string { return r.errors[:] }

// Warnings implements AbstractLocationSolverResponse.
func (r abstractLocationSolverResponse) Warnings() []string { return r.warnings[:] }

// Raw implements AbstractLocationSolverResponse.
func (r abstractLocationSolverResponse) Raw() *json.RawMessage { return r.raw }

func (r *LocationSolverResult) abstractResult() AbstractLocationSolverResult {
	if r == nil {
		return nil
	}
	source := ttnpb.LocationSource_SOURCE_UNKNOWN
	switch r.Algorithm {
	case Algorithm_RSSI:
		source = ttnpb.LocationSource_SOURCE_LORA_RSSI_GEOLOCATION
	case Algorithm_TDOA, Algorithm_RSSITDOA:
		source = ttnpb.LocationSource_SOURCE_LORA_TDOA_GEOLOCATION
	}
	return abstractLocationSolverResult{
		algorithm: strings.ToLower(r.Algorithm),
		location: ttnpb.Location{
			Latitude:  r.Location.Latitude,
			Longitude: r.Location.Longitude,
			Accuracy:  int32(r.Location.Tolerance),
			Source:    source,
		},
	}
}

// AbstractResponse converts the location solver response to the abstract location response format.
func (r ExtendedLocationSolverResponse) AbstractResponse() AbstractLocationSolverResponse {
	return abstractLocationSolverResponse{
		result:   r.Result.abstractResult(),
		errors:   r.Errors[:],
		warnings: r.Warnings[:],
		raw:      r.Raw,
	}
}

func (r *GNSSLocationSolverResult) abstractResult() AbstractLocationSolverResult {
	if r == nil || len(r.LLH) != 3 {
		return nil
	}
	return abstractLocationSolverResult{
		algorithm: "gnss",
		location: ttnpb.Location{
			Latitude:  r.LLH[0],
			Longitude: r.LLH[1],
			Altitude:  int32(r.LLH[2]),
			Accuracy:  int32(r.Accuracy),
			Source:    ttnpb.LocationSource_SOURCE_GPS,
		},
	}
}

// AbstractResponse converts the GNSS location solver response to the abstract location response format.
func (r *ExtendedGNSSLocationSolverResponse) AbstractResponse() AbstractLocationSolverResponse {
	return abstractLocationSolverResponse{
		result:   r.Result.abstractResult(),
		errors:   r.Errors[:],
		warnings: r.Warnings[:],
		raw:      r.Raw,
	}
}

func (r *WiFiLocationSolverResult) abstractResult() AbstractLocationSolverResult {
	if r == nil {
		return nil
	}
	return abstractLocationSolverResult{
		algorithm: strings.ToLower(r.Algorithm),
		location: ttnpb.Location{
			Latitude:  r.Latitude,
			Longitude: r.Longitude,
			Altitude:  int32(r.Altitude),
			Accuracy:  int32(r.Accuracy),
			Source:    ttnpb.LocationSource_SOURCE_WIFI_RSSI_GEOLOCATION,
		},
	}
}

// AbstractResponse converts the WiFi location solver response to the abstract location response format.
func (r *ExtendedWiFiLocationSolverResponse) AbstractResponse() abstractLocationSolverResponse {
	return abstractLocationSolverResponse{
		result:   r.Result.abstractResult(),
		errors:   r.Errors[:],
		warnings: r.Warnings[:],
		raw:      r.Raw,
	}
}
