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

const selectWebhookStore = state => state.webhooks

// Webhook.
export const selectWebhookEntityStore = state => selectWebhookStore(state).entities
export const selectSelectedWebhookId = state => selectWebhookStore(state).selectedWebhook
export const selectSelectedWebhook = state =>
  selectWebhookEntityStore(state)[selectSelectedWebhookId(state)]

// Webhooks.
export const selectWebhooks = state => Object.values(selectWebhookEntityStore(state))
export const selectWebhooksTotalCount = state => selectWebhookStore(state).totalCount
