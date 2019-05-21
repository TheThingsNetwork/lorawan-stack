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

export const createGetApiKeyActionType = name => (
  `GET_${name}_API_KEY`
)

export const createGetApiKeySuccessActionType = name => (
  `GET_${name}_API_KEY_SUCCESS`
)

export const createGetApiKeyFailureActionType = name => (
  `GET_${name}_API_KEY_FAILURE`
)

export const getApiKey = name => (entityId, keyId) => (
  { type: createGetApiKeyActionType(name), entityId, keyId }
)

export const getApiKeySuccess = name => key => (
  { type: createGetApiKeySuccessActionType(name), key }
)

export const getApiKeyFailure = name => error => (
  { type: createGetApiKeyFailureActionType(name), error }
)
