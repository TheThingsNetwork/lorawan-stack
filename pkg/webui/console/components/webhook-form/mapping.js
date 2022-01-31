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
let currentHeaders

export const decodeBasicAuthRequest = value => {
  currentHeaders = value
  const useBasicAuth = value?.Authorization?.startsWith('Basic')
  return useBasicAuth
}

export const encodeBasicAuthRequest = value => {
  if (value) {
    return { ...currentHeaders, Authorization: 'Basic' }
  }

  return currentHeaders
}

export const decodeBasicAuthHeaderUsername = value => {
  const basicAuth = value.Authorization

  if (basicAuth) {
    const encodedCredentials = basicAuth.split('Basic')[1]
    if (encodedCredentials) {
      const decodedCredentials = atob(encodedCredentials)
      const decodedUsername = decodedCredentials.slice(0, decodedCredentials.indexOf(':'))
      return decodedUsername
    }
  }

  return ''
}

export const mapCredentialsToAuthHeader = (forUsername, fieldValue) => {
  if (currentHeaders.Authorization && currentHeaders.Authorization.startsWith('Basic')) {
    const encodedCredentials = currentHeaders.Authorization.split('Basic')[1]

    if (encodedCredentials) {
      const decodedCredentials = atob(encodedCredentials)
      const username = decodedCredentials.slice(0, decodedCredentials.indexOf(':'))
      const password = decodedCredentials.slice(
        decodedCredentials.indexOf(':') + 1,
        decodedCredentials.length,
      )
      if (forUsername) {
        return {
          ...currentHeaders,
          Authorization: `Basic ${btoa(`${fieldValue}:${password}`)}`,
        }
      }
      return {
        ...currentHeaders,
        Authorization: `Basic ${btoa(`${username}:${fieldValue}`)}`,
      }
    }

    // If there is no password/username yet.
    if (forUsername) {
      return {
        ...currentHeaders,
        Authorization: `Basic ${btoa(`${fieldValue}:`)}`,
      }
    }

    return {
      ...currentHeaders,
      Authorization: `Basic ${btoa(`:${fieldValue}`)}`,
    }
  }
}

export const decodeBasicAuthHeaderPassword = value => {
  const basicAuth = value.Authorization

  if (basicAuth) {
    const encodedCredentials = basicAuth.split('Basic')[1]
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

export const createbasicAuthEncoder = forPassword => value =>
  mapCredentialsToAuthHeader(forPassword, value)

export const encodeBasicAuthUsername = createbasicAuthEncoder(true)
export const encodeBasicAuthPassword = createbasicAuthEncoder(false)

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
