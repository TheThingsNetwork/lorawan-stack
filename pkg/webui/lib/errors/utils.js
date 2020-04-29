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

import errorMessages from './error-messages'
import grpcErrToHttpErr from './grpc-error-map'

/**
 * Tests whether the error is a backend error object.
 *
 * @param {object} error - The error to be tested.
 * @returns {boolean} `true` if `error` is a well known backend error object.
 */
export const isBackend = error =>
  Boolean(error) &&
  typeof error === 'object' &&
  !('id' in error) &&
  error.message &&
  error.details &&
  (error.code || error.grpc_code)

/**
 * Returns whether the error is a frontend defined error object.
 *
 * @param {object} error - The error to be tested.
 * @returns {boolean} `true` if `error` is a well known frontend error object.
 */
export const isFrontend = error => Boolean(error) && typeof error === 'object' && error.isFrontend

/**
 * Returns whether `details` is a backend error details object.
 *
 * @param {object} details - The object to be tested.
 * @returns {boolean} `true` if `details` is a well known backend error details object,
 * `false` otherwise.
 */
export const isBackendErrorDetails = details =>
  Boolean(details) &&
  Boolean(details.namespace) &&
  Boolean(details.name) &&
  Boolean(details.message_format) &&
  Boolean(details.code)

/**
 * Returns whether the error has a shape that is not well-known.
 *
 * @param {object} error - The error to be tested.
 * @returns {boolean} `true` if `error` is not of a well known shape.
 */
export const isUnknown = error => !isBackend(error) && !isFrontend(error)

/**
 * Returns a frontend error object, to be passed to error components.
 *
 * @param {object} errorTitle - The error message title (i18n message).
 * @param {object} errorMessage - The error message object (i18n message).
 * @param {string} errorCode - An optional error code to be used to identify
 * a specific error type easily. E.g. `user_status_unapproved`.
 * @param {number} statusCode - An optional status code corresponding to
 * the well known HTTP status codes. This can help categorizing the error if
 * necessary.
 * @returns {object} A frontend error object to be passed to error components.
 */
export const createFrontendError = (errorTitle, errorMessage, errorCode, statusCode) => ({
  errorTitle,
  errorMessage,
  errorCode,
  statusCode,
  isFrontend: true,
})

/**
 * Maps the error type to a HTTP Status code. Useful for quickly
 * determining the type of the error. Returns false if no status code can be
 * determined.
 *
 * @param {object} error - The error to be tested.
 * @returns {number} The (closest when grpc error) HTTP Status Code, otherwise
 * `undefined`.
 */
export const httpStatusCode = error => {
  if (!Boolean(error)) {
    return undefined
  }

  let statusCode = undefined
  if (isBackend(error)) {
    statusCode = error.http_code || grpcErrToHttpErr(error.code || error.grpc_code)
  } else if (isFrontend(error)) {
    statusCode = error.statusCode
  } else if (Boolean(error.statusCode)) {
    statusCode = error.statusCode
  }

  return Boolean(statusCode) ? parseInt(statusCode) : undefined
}
/**
 * Returns the GRPC Status code in case of a backend error.
 *
 * @param {object} error - The error to be tested.
 * @returns {number} The GRPC error code, or `false`.
 */
export const grpcStatusCode = error => isBackend(error) && (error.code || error.grpc_code)

/**
 * Tests whether the grpc error represents the not found erorr.
 *
 * @param {object} error - The error object to be tested.
 * @returns {boolean} `true` if `error` represents the not found error,
 * `false` otherwise.
 */
export const isNotFoundError = error => grpcStatusCode(error) === 5 || httpStatusCode(error) === 404

/**
 * Returns whether the grpc error represents an internal server error.
 *
 * @param {object} error - The error to be tested.
 * @returns {boolean} `true` if `error` represents an internal server error,
 * `false` otherwise.
 */
export const isInternalError = error => grpcStatusCode(error) === 13 // NOTE: HTTP 500 can also be UnknownError.

/**
 * Returns whether the grpc error represents an invalid argument or bad request
 * error.
 *
 * @param {object} error - The error to be tested.
 * @returns {boolean} `true` if `error` represents an invalid argument or bad
 * request error, `false` otherwise.
 */
export const isInvalidArgumentError = error =>
  grpcStatusCode(error) === 3 || httpStatusCode(error) === 400

/**
 * Returns whether the grpc error represents an already exists error.
 *
 * @param {object} error - The error to be tested.
 * @returns {boolean} `true` if `error` represents an already exists error,
 * `false` otherwise.
 */
export const isAlreadyExistsError = error => grpcStatusCode(error) === 6 // NOTE: HTTP 409 can also be AbortedError.

/**
 * Returns whether the grpc error represents a permission denied error.
 *
 * @param {object} error - The error to be tested.
 * @returns {boolean} `true` if `error` represents a permission denied error,
 * `false` otherwise.
 */
export const isPermissionDeniedError = error =>
  grpcStatusCode(error) === 7 || httpStatusCode(error) === 403

/**
 * Returns whether the grpc error represents an error due to not being
 * authenticated.
 *
 * @param {object} error - The error to be tested.
 * @returns {boolean} `true` if `error` represents an `Unauthenticated` error,
 * `false` otherwise.
 */
export const isUnauthenticatedError = error =>
  grpcStatusCode(error) === 16 || httpStatusCode(error) === 401

/**
 * Returns whether `error` has translation ids.
 *
 * @param {object} error - The error to be tested.
 * @returns {boolean} `true` if `error` has translation ids, `false` otherwise.
 */
export const isTranslated = error =>
  isBackend(error) || isFrontend(error) || (typeof error === 'object' && error.id)

/**
 * Returns the id of the error, used as message id.
 *
 * @param {object} error - The backend error object.
 * @returns {string} The ID.
 */
export const getBackendErrorId = error => error.message.split(' ')[0]

/**
 * Returns the id of the error details, used as message id.
 *
 * @param {object} details - The backend error details object.
 * @returns {string} The ID.
 */
export const getBackendErrorDetailsId = details => `error:${details.namespace}:${details.name}`

/**
 * Returns error details.
 *
 * @param {object} error - The backend error object.
 * @returns {object} - The details of `error`.
 */
export const getBackendErrorDetails = error => error.details[0]

/**
 * Returns the name of the error extracted from the details array.
 *
 * @param {object} error - The backend error object.
 * @returns {string} - The error name.
 */
export const getBackendErrorName = error =>
  error && error.details instanceof Array && error.details[0] && error.details[0].name
    ? error.details[0].name
    : undefined
/**
 * Returns the default message of the error, used as fallback translation.
 *
 * @param {object} error - The backend error object.
 * @returns {string} The default message.
 */
export const getBackendErrorDefaultMessage = error =>
  error.details[0].message_format || error.message.replace(/^.*\s/, '')

/**
 * Returns the root cause of the error.
 *
 * @param {object} error - The backend error object.
 * @returns {object} - The root cause of `error`.
 */
export const getBackendErrorRootCause = error => {
  const details = getBackendErrorDetails(error)

  let rootCause = details
  while ('cause' in rootCause) {
    rootCause = rootCause.cause
  }

  return rootCause
}

/**
 * Returns the attributes of the backend error message, if any.
 *
 * @param {object} error - The backend error object.
 * @returns {string} The attributes or undefined.
 */
export const getBackendErrorMessageAttributes = error => error.details[0].attributes

/**
 * Adapts the error object to props of message object, if possible.
 *
 * @param {object} error - The backend error object.
 * @returns {object} Message props of the error object, or generic error object.
 */
export const toMessageProps = function(error) {
  let props
  // Check if it is a error message and transform it to a intl message.
  if (isBackend(error)) {
    props = {
      content: {
        id: getBackendErrorId(error),
        defaultMessage: getBackendErrorDefaultMessage(error),
      },
      values: getBackendErrorMessageAttributes(error),
    }
  } else if (isBackendErrorDetails(error)) {
    props = {
      content: {
        id: getBackendErrorDetailsId(error),
        defaultMessage: error.message_format,
      },
      values: error.attributes,
    }
  } else if (isFrontend(error)) {
    props = { content: error.errorMessage }
  } else if (isTranslated(error)) {
    // Fall back to normal message.
    props = { content: error }
  } else {
    // Fall back to generic error message.
    props = { content: errorMessages.genericError }
  }

  return props
}
