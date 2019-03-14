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

export const GET_DEV = 'GET_DEVICE'
export const GET_DEV_SUCCESS = 'GET_DEVICE_SUCCESS'
export const GET_DEV_FAILURE = 'GET_DEVICE_FAILURE'

export const getDevice = (appId, deviceId, selector, options) => (
  { type: GET_DEV, appId, deviceId, selector, options }
)

export const getDeviceSuccess = device => (
  { type: GET_DEV_SUCCESS, device }
)

export const getDeviceFailure = error => (
  { type: GET_DEV_FAILURE, error }
)
