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

/**
 * Tests whether the grpc error represents the not found erorr.
 * @param {Object} error - The error object to be tested.
 * @returns {boolean} `true` if `error` represents the not found error,
 * `false` otherwise.
 */
export const isNotFoundError = error => (
  error && error.code && error.code === 5
)

/**
 * Tests wether `error` is translated.
 * @param {Object} error - The error to be tested.
 * @returns {boolean} `true` if `error` is translated, `false` otherwise.
 */
export const isErrorTranslated = error => (
  typeof error === 'object' && error.id && error.defaultMessage
)
