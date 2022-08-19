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
  if (isPlainObject(messageType)) {
    if ('path' in messageType) {
      return { enabled: true, value: messageType.path }
    }

    return { enabled: true, value: '' }
  }

  return { enabled: false, value: '' }
}

export const encodeMessageType = formValue => {
  if (formValue && formValue.enabled) {
    return { path: formValue.value }
  }

  return null
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
    webhook_id: '',
  },
  base_url: '',
  format: 'json',
  field_mask: {
    paths: [],
  },
  downlink_api_key: '',
  uplink_message: null,
  join_accept: null,
  downlink_ack: null,
  downlink_nack: null,
  downlink_sent: null,
  downlink_failed: null,
  downlink_queued: null,
  downlink_queue_invalidated: null,
  location_solved: null,
  service_data: null,
  _basic_auth_enabled: false,
  _basic_auth_username: '',
  _basic_auth_password: '',
}
