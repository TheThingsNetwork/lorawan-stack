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

import {
  getApiKeysList,
  createGetApiKeysListActionType,
  getApiKeysListFailure,
  createGetApiKeysListFailureActionType,
  getApiKeysListSuccess,
  createGetApiKeysListSuccessActionType,
  getApiKey,
  createGetApiKeyActionType,
} from '../actions/api-keys'

export const SHARED_NAME = 'APPLICATION'

export const GET_APP = 'GET_APPLICATION'
export const GET_APP_SUCCESS = 'GET_APPLICATION_SUCCESS'
export const GET_APP_FAILURE = 'GET_APPLICATION_FAILURE'
export const GET_APP_API_KEYS_LIST = createGetApiKeysListActionType(SHARED_NAME)
export const GET_APP_API_KEYS_LIST_SUCCESS = createGetApiKeysListSuccessActionType(SHARED_NAME)
export const GET_APP_API_KEYS_LIST_FAILURE = createGetApiKeysListFailureActionType(SHARED_NAME)
export const GET_APP_API_KEY_PAGE_DATA = createGetApiKeyActionType(SHARED_NAME)

export const getApplication = id => (
  { type: GET_APP, id }
)

export const getApplicationSuccess = application => (
  { type: GET_APP_SUCCESS, application }
)

export const getApplicationFailure = error => (
  { type: GET_APP_FAILURE, error }
)

export const getApplicationApiKeysList = getApiKeysList(SHARED_NAME)

export const getApplicationApiKeysListSuccess = getApiKeysListSuccess(SHARED_NAME)

export const getApplicationApiKeysListFailure = getApiKeysListFailure(SHARED_NAME)

export const getApplicationApiKeyPageData = getApiKey(SHARED_NAME)
