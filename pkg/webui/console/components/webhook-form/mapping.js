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

import { isPlainObject } from 'lodash'

export const isBasicAuth = header =>
  isPlainObject(header) && header.key === 'Authorization' && header.value?.startsWith('Basic ')
export const hasBasicAuth = headers =>
  headers instanceof Array && headers.findIndex(isBasicAuth) !== -1
export const getBasicAuthValue = headers => headers.find(isBasicAuth).value.split('Basic ')[1]

export const decodeMessageType = messageType => {
  if (messageType && (messageType.enabled || messageType.path)) {
    return { enabled: true, value: messageType.path }
  }

  return { enabled: false, value: '' }
}

export const encodeMessageType = formValue => {
  if (formValue && formValue.enabled) {
    return { enabled: true, path: formValue.value }
  }
  return { enabled: false, path: '' }
}

export const decodeHeaders = headersType =>
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

export const encodeHeaders = formValue =>
  (formValue &&
    formValue.reduce(
      (result, { key, value }) => ({
        ...result,
        [key]: value,
      }),
      {},
    )) ||
  null

// Encode and decode basic auth header.

export const decodeBasicAuthRequest = value => hasBasicAuth(value)

export const encodeBasicAuthRequest = (value, fieldValue) => {
  if (value) {
    return [...(fieldValue || []), { key: 'Authorization', value: 'Basic ' }]
  }

  if (hasBasicAuth(fieldValue)) {
    return fieldValue.filter(h => !isBasicAuth(h))
  }

  return fieldValue
}

export const decodeBasicAuthHeaderUsername = value => {
  if (hasBasicAuth(value)) {
    const encodedCredentials = getBasicAuthValue(value)
    if (encodedCredentials) {
      const decodedCredentials = atob(encodedCredentials)
      const decodedUsername = decodedCredentials.slice(0, decodedCredentials.indexOf(':'))
      return decodedUsername
    }
  }

  return ''
}

export const mapCredentialsToAuthHeader = (forUsername, value, currentHeaders) => {
  const basicAuthIndex = currentHeaders.findIndex(isBasicAuth)
  if (basicAuthIndex !== -1) {
    const encodedCredentials = getBasicAuthValue(currentHeaders)
    if (encodedCredentials !== undefined) {
      const decodedCredentials = atob(encodedCredentials)
      const username = decodedCredentials.slice(0, decodedCredentials.indexOf(':'))
      const password = decodedCredentials.slice(
        decodedCredentials.indexOf(':') + 1,
        decodedCredentials.length,
      )
      const newHeaders = [...currentHeaders]
      if (forUsername && (value !== '' || password !== '')) {
        newHeaders[basicAuthIndex].value = `Basic ${btoa(`${value}:${password}`)}`
      } else if (!forUsername && (value !== '' || username !== '')) {
        newHeaders[basicAuthIndex].value = `Basic ${btoa(`${username}:${value}`)}`
      } else {
        newHeaders[basicAuthIndex].value = 'Basic '
      }
      return newHeaders
    }
  }
  return currentHeaders
}

export const decodeBasicAuthHeaderPassword = value => {
  if (hasBasicAuth(value)) {
    const encodedCredentials = getBasicAuthValue(value)
    if (encodedCredentials) {
      const decodedCredentials = atob(encodedCredentials)
      const decodedPassword = decodedCredentials.slice(
        decodedCredentials.indexOf(':') + 1,
        decodedCredentials.length,
      )
      return decodedPassword
    }
  }

  return ''
}

export const createBasicAuthEncoder = forPassword => (value, fieldValue) =>
  mapCredentialsToAuthHeader(forPassword, value, fieldValue)

export const encodeBasicAuthUsername = createBasicAuthEncoder(true)
export const encodeBasicAuthPassword = createBasicAuthEncoder(false)

export const blankValues = {
  ids: {
    webhook_id: undefined,
  },
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
