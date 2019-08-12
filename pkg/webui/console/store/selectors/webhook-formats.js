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

import { GET_WEBHOOK_FORMATS_BASE } from '../actions/webhook-formats'
import { createFetchingSelector } from './fetching'
import { createErrorSelector } from './error'

const selectWebhookFormatsStore = state => state.webhookFormats

export const selectWebhookFormats = function(state) {
  const store = selectWebhookFormatsStore(state)

  return store.formats || {}
}

export const selectWebhookFormatsError = createErrorSelector(GET_WEBHOOK_FORMATS_BASE)
export const selectWebhookFormatsFetching = createFetchingSelector(GET_WEBHOOK_FORMATS_BASE)
