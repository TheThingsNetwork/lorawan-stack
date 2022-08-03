// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

import { withScope, captureException } from '@sentry/browser'

import { getSentryErrorTitle, createFrontendError, ingestError, toMessageProps } from './utils'
import errorMessages from './error-messages'
import { TokenError } from './custom-errors'

jest.mock('@sentry/browser')
jest.mock('@ttn-lw/lib/log')

const setTags = jest.fn()
const setExtras = jest.fn()
const setFingerprint = jest.fn()

captureException.mockImplementation(jest.fn)
withScope.mockImplementation(callback => callback({ setTags, setExtras, setFingerprint }))

const backendErrorWithDetails = {
  code: 2,
  message:
    'error:pkg/assets:http (HTTP error: `` is not a valid ID. Must be at least 2 and at most 36 characters long and may consist of only letters, numbers and dashes. It may not start or end with a dash)',
  details: [
    {
      '@type': 'type.googleapis.com/ttn.lorawan.v3.ErrorDetails',
      namespace: 'pkg/assets',
      name: 'http',
      message_format: 'HTTP error: {message}',
      attributes: {
        message:
          '`` is not a valid ID. Must be at least 2 and at most 36 characters long and may consist of only letters, numbers and dashes. It may not start or end with a dash',
      },
    },
  ],
}

const backendErrorWithDetailsAndCause = {
  code: 3,
  message: 'error:pkg/ttnpb:identifiers (invalid identifiers)',
  details: [
    {
      '@type': 'type.googleapis.com/ttn.lorawan.v3.ErrorDetails',
      cause: {
        attributes: {
          field: 'example_field',
          reason: 'Example validation error',
        },
        code: 3,
        correlation_id: 'ab15917421584dafb5c9abb50e91ae71',
        message_format: 'invalid `{field}`: {reason}',
        name: 'validation',
        namespace: 'pkg/errors',
      },
      code: 3,
      correlation_id: 'e319ea4850b84ef0a12b8d7080bf83d6',
      message_format: 'invalid identifiers',
      name: 'identifiers',
      namespace: 'pkg/ttnpb',
    },
  ],
}

const backendErrorDetails = {
  '@type': 'type.googleapis.com/ttn.lorawan.v3.ErrorDetails',
  namespace: 'pkg/networkserver',
  name: 'duplicate',
  message_format: 'uplink is a duplicate',
  correlation_id: 'c2b9568e95df4d369974b822bc3e1b48',
  code: 9,
}

const backendErrorDetailsWithPathErrors = {
  '@type': 'type.googleapis.com/ttn.lorawan.v3.ErrorDetails',
  namespace: 'pkg/gatewayserver',
  name: 'schedule',
  message_format: 'failed to schedule',
  correlation_id: '4827366aac5c437b9027904002f9f6ee',
  code: 10,
  details: [
    {
      '@type': 'type.googleapis.com/ttn.lorawan.v3.ScheduleDownlinkErrorDetails',
      path_errors: [
        {
          namespace: 'pkg/gatewayserver',
          name: 'schedule_path',
          message_format: 'failed to schedule on path `{gateway_uid}`',
          attributes: {
            gateway_uid: 'ttig-adrian-2@tti',
          },
          correlation_id: '563a0e2e63e64cb28322590c61b3dd73',
          cause: {
            namespace: 'pkg/gatewayserver/io',
            name: 'data_rate_rx_window',
            message_format: 'invalid data rate in Rx window `{window}`',
            attributes: {
              window: 1,
            },
            correlation_id: '0aa2e629add64961bd8276ca274fc17f',
            code: 3,
          },
          code: 3,
        },
      ],
    },
  ],
}

const backendErrorWithUnknownDetailStructure = {
  '@type': 'type.googleapis.com/ttn.lorawan.v3.ErrorDetails',
  namespace: 'pkg/applicationserver/io/packages/loradms/v1/api',
  name: 'request',
  message_format: 'LoRaCloud DMS request',
  correlation_id: 'b0de67a448334364a53df8cbd9f9b429',
  code: 14,
  details: [
    {
      '@type': 'type.googleapis.com/google.protobuf.Struct',
      value: {
        body: 'Unauthorized status',
        status_code: 401,
      },
    },
  ],
}

const conflictBackendError = {
  code: 6,
  message: 'error:pkg/identityserver/store:id_taken (ID already taken)',
  details: [
    {
      '@type': 'type.googleapis.com/ttn.lorawan.v3.ErrorDetails',
      namespace: 'pkg/identityserver/store',
      name: 'id_taken',
      message_format: 'ID already taken',
      correlation_id: '06d95b39becd435cbb67ab87a3c93312',
      code: 6,
    },
  ],
}

const loginFailedBackendError = {
  code: 3,
  message: 'error:pkg/account/session:no_user_id_password_match (incorrect password or user ID)',
  details: [
    {
      '@type': 'type.googleapis.com/ttn.lorawan.v3.ErrorDetails',
      namespace: 'pkg/account/session',
      name: 'no_user_id_password_match',
      message_format: 'incorrect password or user ID',
      correlation_id: '689aabd7985b4027871581d7b04d7b97',
      code: 3,
    },
  ],
}

const frontendError = createFrontendError(
  errorMessages.unknownErrorTitle,
  errorMessages.genericError,
  undefined,
  500,
)

const timeoutError = {
  code: 'ECONNABORTED',
  columnNumber: '[undefined]',
  config: {
    adapter: '[Function: <anonymous>]',
    baseURL: 'https://au1.cloud.thethings.network/api/v3',
    data: '[undefined]',
    headers: {
      Accept: 'application/json, text/plain, */*',
      Authorization: '[Filtered]',
    },
    maxBodyLength: -1,
    maxContentLength: -1,
    method: 'get',
    timeout: 10000,
    transformRequest: ['[Function: <anonymous>]'],
    transformResponse: ['[Function: <anonymous>]'],
    transitional: {
      clarifyTimeoutError: false,
      forcedJSONParsing: true,
      silentJSONParsing: true,
    },
    url: '/edtc/formats',
    validateStatus: '[Function: validateStatus]',
    xsrfCookieName: 'XSRF-TOKEN',
    xsrfHeaderName: 'X-XSRF-TOKEN',
  },
  description: '[undefined]',
  fileName: '[undefined]',
  lineNumber: '[undefined]',
  message: 'timeout of 10000ms exceeded',
  name: 'Error',
  number: '[undefined]',
  stack:
    'Error: timeout of 10000ms exceeded\n    at e.exports (https://assets.cloud.thethings.network/console.57b1ac5a73941623bdf2.js:1:1709004)',
  status: null,
}

const plainFrontendError = createFrontendError(errorMessages.unknownErrorTitle)
const codeError = { code: 'ECONNABORTED' }
const statusCodeError = { statusCode: 404 }
const emptyError = {}
const undefinedError = undefined
const errorInstance = new Error('There was an unknown error')
const networkError = new Error('Network Error')
const tokenTimeoutError = new TokenError('Could not fetch token', timeoutError)
const tokenNetworkError = new TokenError('Could not fetch token', networkError)
const tokenError = new TokenError('Token error', backendErrorWithDetails)

describe('Get Sentry error title', () => {
  it('retrieves the right error title', () => {
    expect(getSentryErrorTitle(backendErrorWithDetails)).toBe(backendErrorWithDetails.message)
    expect(getSentryErrorTitle(frontendError)).toBe(frontendError.errorTitle.defaultMessage)
    expect(getSentryErrorTitle(codeError)).toBe(codeError.code)
    expect(getSentryErrorTitle(statusCodeError)).toBe(`status code: ${statusCodeError.statusCode}`)
    expect(getSentryErrorTitle(emptyError)).toBe('untitled or empty error')
    expect(getSentryErrorTitle(undefinedError)).toBe(`invalid error type: undefined`)
  })
})

describe('Ingest error', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  describe('when passing backend errors', () => {
    it('correctly forwards sentry-worthy errors', () => {
      ingestError(backendErrorWithDetails)
      expect(withScope).toHaveBeenCalledTimes(1)
      expect(setTags).toHaveBeenCalledTimes(1)
      expect(setTags.mock.calls[0][0]).toHaveProperty('frontendOrigin', true)
      expect(setExtras).toHaveBeenCalledTimes(1)
      expect(setExtras.mock.calls[0][0]).toHaveProperty('code', 2)
      expect(setExtras.mock.calls[0][0]).toHaveProperty('details')
      expect(Object.keys(setExtras.mock.calls[0][0])).toHaveLength(3)
      expect(setFingerprint).toHaveBeenCalledTimes(1)
      expect(setFingerprint.mock.calls[0][0]).toBe('error:pkg/assets:http')
      expect(captureException).toHaveBeenCalledTimes(1)
      expect(captureException.mock.calls[0][0] instanceof Error).toBe(true)
      expect(captureException.mock.calls[0][0].toString()).toBe(
        `Error: ${getSentryErrorTitle(backendErrorWithDetails)}`,
      )
    })

    it('correctly discards sentry-unworthy errors', () => {
      ingestError(conflictBackendError)
      ingestError(loginFailedBackendError)
      expect(withScope).not.toHaveBeenCalled()
      expect(captureException).not.toHaveBeenCalled()
    })
  })

  describe('when passing frontend errors', () => {
    it('correctly forwards sentry-worthy errors', () => {
      ingestError(frontendError)
      expect(withScope).toHaveBeenCalledTimes(1)
      expect(setTags).toHaveBeenCalledTimes(1)
      expect(setTags.mock.calls[0][0]).toHaveProperty('frontendOrigin', true)
      expect(setExtras).toHaveBeenCalledTimes(1)
      expect(setExtras.mock.calls[0][0]).toHaveProperty('errorTitle', frontendError.errorTitle)
      expect(setExtras.mock.calls[0][0]).toHaveProperty('errorMessage')
      expect(setExtras.mock.calls[0][0]).toHaveProperty('isFrontend', true)
      expect(Object.keys(setExtras.mock.calls[0][0])).toHaveLength(5)
      expect(setFingerprint).toHaveBeenCalledTimes(1)
      expect(setFingerprint.mock.calls[0][0]).toBe(frontendError)
      expect(captureException).toHaveBeenCalledTimes(1)
      expect(captureException.mock.calls[0][0] instanceof Error).toBe(true)
      expect(captureException.mock.calls[0][0].toString()).toBe(
        `Error: ${frontendError.errorTitle.defaultMessage}`,
      )
    })

    it('correctly discards sentry-unworthy errors', () => {
      ingestError(plainFrontendError)
      expect(withScope).not.toHaveBeenCalled()
      expect(captureException).not.toHaveBeenCalled()
    })
  })

  it('correctly forwards error instances', () => {
    ingestError(errorInstance)
    expect(withScope).toHaveBeenCalledTimes(1)
    expect(setTags).toHaveBeenCalledTimes(1)
    expect(setTags.mock.calls[0][0]).toHaveProperty('frontendOrigin', true)
    expect(setExtras).toHaveBeenCalledTimes(1)
    expect(setExtras.mock.calls[0][0]).toHaveProperty('error', errorInstance)
    expect(Object.keys(setExtras.mock.calls[0][0])).toHaveLength(1)
    expect(setFingerprint).toHaveBeenCalledTimes(1)
    expect(setFingerprint.mock.calls[0][0]).toBe(errorInstance)
    expect(captureException).toHaveBeenCalledTimes(1)
    expect(captureException.mock.calls[0][0] instanceof Error).toBe(true)
    expect(captureException.mock.calls[0][0].toString()).toBe(errorInstance.toString())
  })

  it('correctly discards errors with irrelevant status code', () => {
    ingestError(statusCodeError)
    expect(withScope).not.toHaveBeenCalled()
    expect(captureException).not.toHaveBeenCalled()
  })

  it('correctly discards network and timeout errors', () => {
    ingestError(networkError)
    ingestError(timeoutError)
    ingestError(tokenNetworkError)
    ingestError(tokenTimeoutError)
    expect(withScope).not.toHaveBeenCalled()
    expect(captureException).not.toHaveBeenCalled()
  })

  it('correctly forwards token errors', () => {
    ingestError(tokenError)
    expect(withScope).toHaveBeenCalledTimes(1)
    expect(setTags).toHaveBeenCalledTimes(1)
    expect(setTags.mock.calls[0][0]).toHaveProperty('frontendOrigin', true)
    expect(setExtras).toHaveBeenCalledTimes(1)
    expect(Object.keys(setExtras.mock.calls[0][0])).toHaveLength(1)
    expect(setFingerprint).toHaveBeenCalledTimes(1)
    expect(setFingerprint.mock.calls[0][0]).toBe(tokenError)
    expect(captureException).toHaveBeenCalledTimes(1)
    expect(captureException.mock.calls[0][0] instanceof Error).toBe(true)
    expect(captureException.mock.calls[0][0].toString()).toBe('TokenError: Token error')
  })

  // Empty or otherwise malformed objects as errors should not occur, but if
  // they do, it's good to forward them to Sentry to be aware of the issue.
  it('correctly forwards empty object errors', () => {
    ingestError(emptyError)
    expect(withScope).toHaveBeenCalledTimes(1)
    expect(setTags).toHaveBeenCalledTimes(1)
    expect(setTags.mock.calls[0][0]).toHaveProperty('frontendOrigin', true)
    expect(setExtras).toHaveBeenCalledTimes(1)
    expect(Object.keys(setExtras.mock.calls[0][0])).toHaveLength(0)
    expect(setFingerprint).toHaveBeenCalledTimes(1)
    expect(setFingerprint.mock.calls[0][0]).toBe(emptyError)
    expect(captureException).toHaveBeenCalledTimes(1)
    expect(captureException.mock.calls[0][0] instanceof Error).toBe(true)
    expect(captureException.mock.calls[0][0].toString()).toBe('Error: untitled or empty error')
  })

  // Undefined errors should not occur, but if they do, it's good to forward
  // them to Sentry to be aware of the issue.
  it('correctly forwards undefined errors', () => {
    ingestError(undefinedError)
    expect(withScope).toHaveBeenCalledTimes(1)
    expect(setTags).toHaveBeenCalledTimes(1)
    expect(setTags.mock.calls[0][0]).toHaveProperty('frontendOrigin', true)
    expect(setExtras).toHaveBeenCalledTimes(1)
    expect(Object.keys(setExtras.mock.calls[0][0])).toHaveLength(1)
    expect(setFingerprint).toHaveBeenCalledTimes(1)
    expect(setFingerprint.mock.calls[0][0]).toBe(undefinedError)
    expect(captureException).toHaveBeenCalledTimes(1)
    expect(captureException.mock.calls[0][0] instanceof Error).toBe(true)
    expect(captureException.mock.calls[0][0].toString()).toBe(
      'Error: invalid error type: undefined',
    )
  })

  it('correctly decorates extras and tags', () => {
    ingestError(
      backendErrorWithDetails,
      { ingestedBy: 'ErrorNotification' },
      { requestAction: 'GET_APPLICATIONS_REQUEST' },
    )
    expect(withScope).toHaveBeenCalledTimes(1)
    expect(setTags).toHaveBeenCalledTimes(1)
    expect(setTags.mock.calls[0][0]).toHaveProperty('frontendOrigin', true)
    expect(setTags.mock.calls[0][0]).toHaveProperty('requestAction', 'GET_APPLICATIONS_REQUEST')
    expect(setExtras).toHaveBeenCalledTimes(1)
    expect(setExtras.mock.calls[0][0]).toHaveProperty('ingestedBy', 'ErrorNotification')
    expect(Object.keys(setExtras.mock.calls[0][0])).toHaveLength(4)
    expect(setFingerprint).toHaveBeenCalledTimes(1)
    expect(captureException).toHaveBeenCalledTimes(1)
  })
})

describe('Converting errors to message props', () => {
  it('correctly extracts from error details', () => {
    const messageProps = toMessageProps(backendErrorDetails)
    expect(messageProps).toMatchObject({
      content: {
        id: 'error:pkg/networkserver:duplicate',
        defaultMessage: 'uplink is a duplicate',
      },
      values: undefined,
    })
  })

  it('correctly extracts from errors with details and no causes', () => {
    const messageProps = toMessageProps(backendErrorWithDetails)
    expect(messageProps).toMatchObject({
      content: {
        id: 'error:pkg/assets:http',
        defaultMessage: 'HTTP error: {message}',
      },
      values: {
        message:
          '`` is not a valid ID. Must be at least 2 and at most 36 characters long and may consist of only letters, numbers and dashes. It may not start or end with a dash',
      },
    })
  })

  it('correctly extracts from errors with details and cause', () => {
    const messageProps = toMessageProps(backendErrorWithDetailsAndCause)
    expect(messageProps).toMatchObject({
      content: {
        id: 'error:pkg/errors:validation',
        defaultMessage: 'invalid `{field}`: {reason}',
      },
      values: {
        field: 'example_field',
        reason: 'Example validation error',
      },
    })
  })

  it('correctly extracts all message props when using the `each` option (details with cause)', () => {
    const messageProps = toMessageProps(backendErrorWithDetailsAndCause, true)
    expect(messageProps).toBeInstanceOf(Array)
    expect(messageProps).toHaveLength(2)
    expect(messageProps[0]).toMatchObject({
      content: {
        id: 'error:pkg/errors:validation',
        defaultMessage: 'invalid `{field}`: {reason}',
      },
      values: {
        field: 'example_field',
        reason: 'Example validation error',
      },
    })
    expect(messageProps[1]).toMatchObject({
      content: {
        id: 'error:pkg/ttnpb:identifiers',
        defaultMessage: 'invalid identifiers',
      },
    })
  })

  it('correctly extracts all message props when using the `each` option (only details)', () => {
    const messageProps = toMessageProps(backendErrorWithDetails, true)
    expect(messageProps).toBeInstanceOf(Array)
    expect(messageProps).toHaveLength(1)
    expect(messageProps[0]).toMatchObject({
      content: {
        id: 'error:pkg/assets:http',
        defaultMessage: 'HTTP error: {message}',
      },
      values: {
        message:
          '`` is not a valid ID. Must be at least 2 and at most 36 characters long and may consist of only letters, numbers and dashes. It may not start or end with a dash',
      },
    })
  })

  it('correctly extracts from error details with unknown detail structure', () => {
    const messageProps = toMessageProps(backendErrorWithUnknownDetailStructure, true)
    expect(messageProps).toBeInstanceOf(Array)
    expect(messageProps).toHaveLength(1)
    expect(messageProps[0]).toMatchObject({
      content: {
        id: 'error:pkg/applicationserver/io/packages/loradms/v1/api:request',
        defaultMessage: 'LoRaCloud DMS request',
      },
    })
  })

  it('correctly extracts from error details with path errors', () => {
    const messageProps = toMessageProps(backendErrorDetailsWithPathErrors)
    expect(messageProps).toMatchObject({
      content: {
        id: 'error:pkg/gatewayserver/io:data_rate_rx_window',
        defaultMessage: 'invalid data rate in Rx window `{window}`',
      },
      values: {
        window: 1,
      },
    })
  })

  it('correctly extracts from frontend errors', () => {
    const messageProps = toMessageProps(frontendError)
    expect(messageProps).toMatchObject({
      title: errorMessages.unknownErrorTitle,
      content: errorMessages.genericError,
    })
  })

  it('correctly extracts from unknown errors', () => {
    const messageProps = toMessageProps(null)
    expect(messageProps).toMatchObject({
      content: errorMessages.genericError,
    })
  })
})
