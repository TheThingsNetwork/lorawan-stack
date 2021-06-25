// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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
	"context"
	"net/http"

	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/packages/loradms/v1/api/objects"
)

// Uplinks is an API client for the Uplink API.
type Uplinks struct {
	cl *Client
}

const uplinkEntity = "uplink"

// Send sends the given uplink to the Device Management service.
func (u *Uplinks) Send(ctx context.Context, uplinks objects.DeviceUplinks) (objects.DeviceUplinkResponses, error) {
	resp, err := u.cl.Do(ctx, http.MethodPost, uplinkEntity, "", sendOperation, uplinks)
	if err != nil {
		return nil, err
	}
	response := make(objects.DeviceUplinkResponses)
	err = parse(&response, resp)
	if err != nil {
		return nil, err
	}
	return response, nil
}
