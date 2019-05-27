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

export const GET_WEBHOOK_FORMATS = 'GET_WEBHOOK_FORMATS'
export const GET_WEBHOOK_FORMATS_SUCCESS = 'GET_WEBHOOK_FORMATS_SUCCESS'
export const GET_WEBHOOK_FORMATS_FAILURE = 'GET_WEBHOOK_FORMATS_FAILURE'

export const getWebhookFormats = () => (
  { type: GET_WEBHOOK_FORMATS }
)

export const getWebhookFormatsSuccess = formats => (
  { type: GET_WEBHOOK_FORMATS_SUCCESS, formats }
)

export const getWebhookFormatsFailure = error => (
  { type: GET_WEBHOOK_FORMATS_FAILURE, error }
)
