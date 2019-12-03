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

import errorMessages from '../errors/error-messages'
import grpcErrToHttpErr from './grpc-error-map'

/**
 * Tests wether the error is a backend error object
 * @param {Object} error - The error to be tested.
 * @returns {boolean} `true` if `error` is translated, `false` otherwise.
 */
export const isBackend = error =>
  Boolean(error) &&
  typeof error === 'object' &&
  !('id' in error) &&
  error.message &&
  error.details &&
  (error.code || error.grpc_code)

/**
 * Returns wether the error is a frontend defined error object
 * @param {Object} error - The error to be tested.
 * @returns {boolean} `true` if `error` is translated, `false` otherwise.
 */
export const isFrontend = error =>
  // TODO: Define proper object shape, once we need it, for now translated
  // messages are enough
  Boolean(error) && typeof error === 'object' && error.id && error.defaultMessage

/**
 * Returns wether the error has a shape that is not well-known
 * @param {Object} error - The error to be tested.
 * @returns {boolean} `true` if `error` is translated, `false` otherwise.
 */
export const isUnknown = error => !isBackend(error) && !isFrontend(error)

/**
 * Maps the error type to a HTTP Status code. Useful for quickly
 * determining the type of the error. Returns false if no status code can be
 * determined.
 * @param {Object} error - The error to be tested.
 * @returns {number} The (clostest when grpc error) HTTP Status Code
 */
export const httpStatusCode = error =>
  isBackend(error)
    ? error.http_code || grpcErrToHttpErr(error.code || error.grpc_code)
    : Boolean(error) && error.statusCode

/**
 * Returns the GRPC Status code in case of a backend error.
 * @param {Object} error - The error to be tested.
 * @returns {number} The GRPC error code, or `false`
 */
export const grpcStatusCode = error => isBackend(error) && (error.code || error.grpc_code)

/**
 * Tests whether the grpc error represents the not found erorr.
 * @param {Object} error - The error object to be tested.
 * @returns {boolean} `true` if `error` represents the not found error,
 * `false` otherwise.
 */
export const isNotFoundError = error => grpcStatusCode(error) === 5 || httpStatusCode(error) === 404

/**
 * Returns whether the grpc error represents an internal server error.
 * @param {Object} error - The error to be tested.
 * @returns {boolean} `true` if `error` represents an internal server error,
 * `false` otherwise.
 */
export const isInternalError = error => grpcStatusCode(error) === 13 // NOTE: HTTP 500 can also be UnknownError.

/**
 * Returns whether the grpc error represents an already exists error.
 * @param {Object} error - The error to be tested.
 * @returns {boolean} `true` if `error` represents an already exists error,
 * `false` otherwise.
 */
export const isAlreadyExistsError = error => grpcStatusCode(error) === 6 // NOTE: HTTP 409 can also be AbortedError.

/**
 * Returns whether the grpc error represents a permission denied error.
 * @param {Object} error - The error to be tested.
 * @returns {boolean} `true` if `error` represents a permission denied error,
 * `false` otherwise.
 */
export const isPermissionDeniedError = error =>
  grpcStatusCode(error) === 7 || httpStatusCode(error) === 403

/**
 * Returns whether the grpc error represents an error due to not being authenticated.
 * @param {Object} error - The error to be tested.
 * @returns {boolean} `true` if `error` represents an `Unauthenticated` error,
 * `false` otherwise.
 */
export const isUnauthenticatedError = error =>
  grpcStatusCode(error) === 16 || httpStatusCode(error) === 401

/**
 * Returns wether `error` has translation ids.
 * @param {Object} error - The error to be tested.
 * @returns {boolean} `true` if `error` has translation ids, `false` otherwise.
 */
export const isTranslated = error => isBackend() || (typeof error === 'object' && error.id)

/**
 * Returns the id of the error, used as message id.
 * @param {Object} error - The backend error object.
 * @returns {string} The ID.
 */
export const getBackendErrorId = error => error.message.split(' ')[0]

/**
 * Returns boolean which determines if error details should be displayed.
 * @param {Object} error - The backend error object.
 * @returns {Object} - Display error details or not.
 */

export const getBackendErrorDetails = error => (error.details[0].cause ? error : undefined)

/**
 * Returns the name of the error extracted from the details array.
 * @param {Object} error - The backend error object.
 * @returns {string} - The error name.
 */
export const getBackendErrorName = error =>
  error && error.details instanceof Array && error.details[0] && error.details[0].name
    ? error.details[0].name
    : undefined
/**
 * Returns the default message of the error, used as fallback translation.
 * @param {Object} error - The backend error object.
 * @returns {string} The default message.
 */
export const getBackendErrorDefaultMessage = error =>
  error.details[0].message_format || error.message.replace(/^.*\s/, '')

/**
 * Returns the attributes of the backend error message, if any.
 * @param {Object} error - The backend error object.
 * @returns {string} The attributes or undefined.
 */
export const getBackendErrorMessageAttributes = error => error.details[0].attributes

/**
 * Adapts the error object to props of message object, if possible.
 * @param {Object} error - The backend error object.
 * @returns {Object} Message props of the error object, or generic error object
 */
export const toMessageProps = function(error) {
  let props
  // Check if it is a error message and transform it to a intl message
  if (isBackend(error)) {
    props = {
      content: {
        id: getBackendErrorId(error),
        defaultMessage: getBackendErrorDefaultMessage(error),
      },
      values: getBackendErrorMessageAttributes(error),
    }
  } else if (isTranslated(error)) {
    // Fall back to normal message
    props = { content: error }
  } else {
    // Fall back to generic error message
    props = { content: errorMessages.genericError }
  }

  return props
}
