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

export const GET_WEBHOOKS_LIST = 'GET_WEBHOOKS_LIST_REQUEST'
export const GET_WEBHOOKS_LIST_SUCCESS = 'GET_WEBHOOKS_LIST_SUCCESS'
export const GET_WEBHOOKS_LIST_FAILURE = 'GET_WEBHOOKS_LIST_FAILURE'

export const getWebhooksList = appId => (
  { type: GET_WEBHOOKS_LIST, appId }
)

export const getWebhooksListSuccess = (webhooks, totalCount) => (
  { type: GET_WEBHOOKS_LIST_SUCCESS, webhooks, totalCount }
)

export const getWebhooksListFailure = error => (
  { type: GET_WEBHOOKS_LIST_FAILURE, error }
)
