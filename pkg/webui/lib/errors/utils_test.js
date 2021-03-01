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

import { getSentryErrorTitle, createFrontendError, ingestError } from './utils'
import errorMessages from './error-messages'

jest.mock('@sentry/browser')
jest.mock('@ttn-lw/lib/log')

const setTags = jest.fn()
const setExtras = jest.fn()
const setFingerprint = jest.fn()

captureException.mockImplementation(jest.fn)
withScope.mockImplementation(callback => callback({ setTags, setExtras, setFingerprint }))

const backendError = {
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
  request_details: { url: '/users/kschiffer/applications', method: 'post', stack_component: 'is' },
}

const frontendError = createFrontendError(
  errorMessages.unknownErrorTitle,
  undefined,
  undefined,
  500,
)
const plainFrontendError = createFrontendError(errorMessages.unknownErrorTitle)
const codeError = { code: 'ECONNABORTED' }
const statusCodeError = { statusCode: 404 }
const emptyError = {}
const undefinedError = undefined
const errorInstance = new Error('There was an unknown error')

describe('Get Sentry error title', () => {
  it('retrieves the right error title', () => {
    expect(getSentryErrorTitle(backendError)).toBe(backendError.message)
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
      ingestError(backendError)
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
        `Error: ${getSentryErrorTitle(backendError)}`,
      )
    })

    it('correctly discards sentry-unworthy errors (e.g. 409 conflict)', () => {
      ingestError(conflictBackendError)
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

  it('correctly discards errors irrelevant status code', () => {
    ingestError(statusCodeError)
    expect(withScope).not.toHaveBeenCalled()
    expect(captureException).not.toHaveBeenCalled()
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
      backendError,
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
