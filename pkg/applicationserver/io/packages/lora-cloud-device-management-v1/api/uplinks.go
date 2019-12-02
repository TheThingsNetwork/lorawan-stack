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
	"bytes"
	"encoding/json"
	"net/http"

	"go.thethings.network/lorawan-stack/pkg/applicationserver/io/packages/lora-cloud-device-management-v1/api/objects"
)

// Uplinks is an API client for the Uplink API.
type Uplinks struct {
	*Client
}

const (
	uplinkEntity = "uplink"
)

// Send sends the given uplink to the Device Management service.
func (u *Uplinks) Send(uplinks objects.DeviceUplinks) (objects.DeviceUplinkResponses, error) {
	buffer := bytes.NewBuffer(nil)
	err := json.NewEncoder(buffer).Encode(uplinks)
	if err != nil {
		return nil, err
	}
	req, err := u.newRequest(http.MethodPost, uplinkEntity, "", sendOperation, buffer)
	if err != nil {
		return nil, err
	}
	resp, err := u.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	response := make(objects.DeviceUplinkResponses)
	err = parse(&response, resp.Body)
	if err != nil {
		return nil, err
	}
	return response, nil
}
