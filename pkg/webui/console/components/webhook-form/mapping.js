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

const isBasicAuth = header =>
  isPlainObject(header) && header.key === 'Authorization' && header.value?.startsWith('Basic ')
const hasBasicAuth = headers => headers instanceof Array && headers.findIndex(isBasicAuth) !== -1

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

const decodeHeaders = headersType =>
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

const encodeHeaders = formValue =>
  (formValue &&
    formValue.reduce(
      (result, { key, value }) => ({
        ...result,
        [key]: value,
      }),
      {},
    )) ||
  null

export const encodeValues = formValues => {
  const {
    _basic_auth_enabled,
    _basic_auth_username,
    _basic_auth_password,
    _headers,
    ...newValues
  } = formValues

  newValues.headers = encodeHeaders(_headers)
  if (_basic_auth_enabled) {
    newValues.headers.Authorization = `Basic ${btoa(
      `${_basic_auth_username || ''}:${_basic_auth_password || ''}`,
    )}`
  }

  return newValues
}

export const decodeValues = backendValues => {
  const formValues = { ...backendValues }
  if (backendValues?.headers?.Authorization?.startsWith('Basic ')) {
    const encodedCredentials = backendValues.headers.Authorization.split('Basic ')[1]
    if (encodedCredentials) {
      const decodedCredentials = atob(encodedCredentials)
      formValues._basic_auth_enabled = true
      formValues._basic_auth_username = decodedCredentials.slice(0, decodedCredentials.indexOf(':'))
      formValues._basic_auth_password = decodedCredentials.slice(
        decodedCredentials.indexOf(':') + 1,
      )
    }
  } else {
    formValues._basic_auth_enabled = false
    formValues._basic_auth_username = ''
    formValues._basic_auth_password = ''
  }

  formValues._headers = decodeHeaders(backendValues?.headers)
  if (hasBasicAuth(formValues._headers)) {
    formValues._headers.find(isBasicAuth).readOnly = true
  }

  return formValues
}

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
  _basic_auth_enabled: false,
  _basic_auth_username: '',
  _basic_auth_password: '',
}
