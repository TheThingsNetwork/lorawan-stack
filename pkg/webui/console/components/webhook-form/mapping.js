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

const mapWebhookMessageTypeToFormValue = messageType =>
  (messageType && { enabled: true, value: messageType.path }) || { enabled: false, value: '' }

const mapWebhookHeadersTypeToFormValue = headersType =>
  (headersType &&
    Object.keys(headersType).reduce(
      (result, key) =>
        result.concat({
          key,
          value: headersType[key],
        }),
      [],
    )) ||
  []

// Encode and decode basic auth header.
let currentHeaders

export const mapBasicAuthHeaderToBoolean = value => {
  currentHeaders = value
  const useBasicAuth =
    value?.some(header => header.key === 'Authorization') &&
    value?.some(header => header.value.startsWith('Basic'))
  return useBasicAuth
}

export const mapBooleanToBasicAuthHeader = value => {
  if (value) {
    return [...currentHeaders, ...{ key: 'Authorization', value: 'Basic' }]
  }
  return currentHeaders.filter(
    header => header.key !== 'Authorizarion' && !header.value.startsWith('Basic'),
  )
}

export const mapBasicAuthHeaderToUsername = value => {
  const basicAuth = value.filter(
    header => header.key === 'Authorization' && header.value.startsWith('Basic'),
  )
  const encodedCredentials = basicAuth[0]?.value.split('Basic')[1]
  if (encodedCredentials) {
    const decodedCredentials = atob(encodedCredentials)
    const decodedUsername = decodedCredentials.slice(0, decodedCredentials.indexOf(':'))
    return decodedUsername
  }

  return ''
}

const createCredentialsToAuthHeaderMapper = forUsername => fieldValue => {
  const updatedHeaders = currentHeaders?.map(({ key, value }) => {
    if (key === 'Authorization' && value.startsWith('Basic')) {
      const encodedCredentials = value.split('Basic')[1]

      if (encodedCredentials) {
        const decodedCredentials = atob(encodedCredentials)
        const username = decodedCredentials.slice(0, decodedCredentials.indexOf(':'))
        const password = decodedCredentials.slice(
          decodedCredentials.indexOf(':') + 1,
          decodedCredentials.length,
        )
        if (forUsername) {
          return {
            key: 'Authorization',
            value: `Basic ${btoa(`${fieldValue}:${password}`)}`,
          }
        }
        return {
          key: 'Authorization',
          value: `Basic ${btoa(`${username}:${fieldValue}`)}`,
        }
      }
      if (forUsername) {
        return { key: 'Authorization', value: `Basic ${btoa(`${fieldValue}:`)}` }
      }

      return { key: 'Authorization', value: `Basic ${btoa(`:${fieldValue}`)}` }
    }

    return { key, value }
  })

  currentHeaders = updatedHeaders
  return currentHeaders
}

export const mapBasicAuthHeaderToPassword = value => {
  const basicAuth = value.filter(
    header => header.key === 'Authorization' && header.value.startsWith('Basic'),
  )
  const encodedCredentials = basicAuth[0]?.value.split('Basic')[1]
  if (encodedCredentials) {
    const decodedCredentials = atob(encodedCredentials)
    const decodedPassword = decodedCredentials.slice(
      decodedCredentials.indexOf(':') + 1,
      decodedCredentials.length,
    )
    return decodedPassword
  }

  return ''
}

export const mapCredentialsToPassword = createCredentialsToAuthHeaderMapper(true)
export const mapCredentialsToUsername = createCredentialsToAuthHeaderMapper(false)

export const mapWebhookToFormValues = webhook => ({
  webhook_id: webhook.ids.webhook_id,
  base_url: webhook.base_url,
  format: webhook.format,
  headers: mapWebhookHeadersTypeToFormValue(webhook.headers),
  downlink_api_key: webhook.downlink_api_key,
  uplink_message: mapWebhookMessageTypeToFormValue(webhook.uplink_message),
  join_accept: mapWebhookMessageTypeToFormValue(webhook.join_accept),
  downlink_ack: mapWebhookMessageTypeToFormValue(webhook.downlink_ack),
  downlink_nack: mapWebhookMessageTypeToFormValue(webhook.downlink_nack),
  downlink_sent: mapWebhookMessageTypeToFormValue(webhook.downlink_sent),
  downlink_failed: mapWebhookMessageTypeToFormValue(webhook.downlink_failed),
  downlink_queued: mapWebhookMessageTypeToFormValue(webhook.downlink_queued),
  downlink_queue_invalidated: mapWebhookMessageTypeToFormValue(webhook.downlink_queue_invalidated),
  location_solved: mapWebhookMessageTypeToFormValue(webhook.location_solved),
  service_data: mapWebhookMessageTypeToFormValue(webhook.service_data),
})

const mapMessageTypeFormValueToWebhookMessageType = formValue =>
  (formValue.enabled && { path: formValue.value }) || null

const mapHeadersTypeFormValueToWebhookHeadersType = formValue =>
  (formValue &&
    formValue.reduce(
      (result, { key, value }) => ({
        ...result,
        [key]: value,
      }),
      {},
    )) ||
  null

export const mapFormValuesToWebhook = (values, appId) => ({
  ids: {
    application_ids: {
      application_id: appId,
    },
    webhook_id: values.webhook_id,
  },
  base_url: values.base_url,
  format: values.format,
  headers: mapHeadersTypeFormValueToWebhookHeadersType(values.headers),
  downlink_api_key: values.downlink_api_key,
  uplink_message: mapMessageTypeFormValueToWebhookMessageType(values.uplink_message),
  join_accept: mapMessageTypeFormValueToWebhookMessageType(values.join_accept),
  downlink_ack: mapMessageTypeFormValueToWebhookMessageType(values.downlink_ack),
  downlink_nack: mapMessageTypeFormValueToWebhookMessageType(values.downlink_nack),
  downlink_sent: mapMessageTypeFormValueToWebhookMessageType(values.downlink_sent),
  downlink_failed: mapMessageTypeFormValueToWebhookMessageType(values.downlink_failed),
  downlink_queued: mapMessageTypeFormValueToWebhookMessageType(values.downlink_queued),
  downlink_queue_invalidated: mapMessageTypeFormValueToWebhookMessageType(
    values.downlink_queue_invalidated,
  ),
  location_solved: mapMessageTypeFormValueToWebhookMessageType(values.location_solved),
  service_data: mapMessageTypeFormValueToWebhookMessageType(values.service_data),
})

export const blankValues = {
  webhook_id: undefined,
  base_url: undefined,
  format: undefined,
  downlink_api_key: '',
  uplink_message: { enabled: false, value: '' },
  join_accept: { enabled: false, value: '' },
  downlink_ack: { enabled: false, value: '' },
  downlink_nack: { enabled: false, value: '' },
  downlink_sent: { enabled: false, value: '' },
  downlink_failed: { enabled: false, value: '' },
  downlink_queued: { enabled: false, value: '' },
  downlink_queue_invalidated: { enabled: false, value: '' },
  location_solved: { enabled: false, value: '' },
  service_data: { enabled: false, value: '' },
}
