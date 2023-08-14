// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

import { get } from 'lodash'

const selectAsStore = state => state.as

export const selectAsConfiguration = state => selectAsStore(state).configuration

export const selectPubSubProviders = state =>
  get(selectAsConfiguration(state), 'pubsub.providers') || {}
export const selectNatsProviderDisabled = state =>
  selectPubSubProviders(state).nats === 'DISABLED' || false
export const selectMqttProviderDisabled = state =>
  selectPubSubProviders(state).mqtt === 'DISABLED' || false

export const selectWebhooksConfiguration = state => selectAsConfiguration(state).webhooks
export const selectWebhooksHealthStatusEnabled = state => {
  const webhooksConfig = selectWebhooksConfiguration(state)

  if (!webhooksConfig) {
    return false
  }

  return (
    webhooksConfig.unhealthy_retry_interval !== '0s' ||
    'unhealthy_attempts_threshold' in webhooksConfig
  )
}

export const selectWebhookHasUnhealthyConfig = state => {
  const webhooksConfig = selectWebhooksConfiguration(state)

  if (!webhooksConfig) {
    return false
  }

  return (
    webhooksConfig.unhealthy_retry_interval !== '0s' &&
    'unhealthy_attempts_threshold' in webhooksConfig
  )
}

export const selectWebhookRetryInterval = state => {
  const webhooksConfig = selectWebhooksConfiguration(state)

  if (!webhooksConfig) {
    return false
  }

  return 'unhealthy_attempts_threshold' in webhooksConfig
    ? selectWebhooksConfiguration(state).unhealthy_retry_interval
    : null
}
