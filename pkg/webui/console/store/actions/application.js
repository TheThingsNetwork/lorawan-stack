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
} from '../actions/api-keys'

import {
  getCollaboratorsList,
  createGetCollaboratorsListActionType,
  getCollaboratorsListFailure,
  createGetCollaboratorsListFailureActionType,
  getCollaboratorsListSuccess,
  createGetCollaboratorsListSuccessActionType,
  createGetCollaboratorActionType,
  getCollaborator,
} from '../actions/collaborators'

import {
  startEventsStream,
  createStartEventsStreamActionType,
  startEventsStreamSuccess,
  createStartEventsStreamSuccessActionType,
  startEventsStreamFailure,
  createStartEventsStreamFailureActionType,
  stopEventsStream,
  createStopEventsStreamActionType,
  clearEvents,
  createClearEventsActionType,
} from '../actions/events'

import {
  getApiKey,
  createGetApiKeyActionType,
  getApiKeySuccess,
  createGetApiKeySuccessActionType,
  getApiKeyFailure,
  createGetApiKeyFailureActionType,
} from './api-key'

export const SHARED_NAME = 'APPLICATION'

export const GET_APP = 'GET_APPLICATION'
export const GET_APP_SUCCESS = 'GET_APPLICATION_SUCCESS'
export const GET_APP_FAILURE = 'GET_APPLICATION_FAILURE'
export const GET_APP_API_KEYS_LIST = createGetApiKeysListActionType(SHARED_NAME)
export const GET_APP_API_KEYS_LIST_SUCCESS = createGetApiKeysListSuccessActionType(SHARED_NAME)
export const GET_APP_API_KEYS_LIST_FAILURE = createGetApiKeysListFailureActionType(SHARED_NAME)
export const GET_APP_API_KEY = createGetApiKeyActionType(SHARED_NAME)
export const GET_APP_API_KEY_SUCCESS = createGetApiKeySuccessActionType(SHARED_NAME)
export const GET_APP_API_KEY_FAILURE = createGetApiKeyFailureActionType(SHARED_NAME)
export const GET_APP_COLLABORATOR_PAGE_DATA = createGetCollaboratorActionType(SHARED_NAME)
export const GET_APP_COLLABORATORS_LIST = createGetCollaboratorsListActionType(SHARED_NAME)
export const GET_APP_COLLABORATORS_LIST_SUCCESS = createGetCollaboratorsListSuccessActionType(SHARED_NAME)
export const GET_APP_COLLABORATORS_LIST_FAILURE = createGetCollaboratorsListFailureActionType(SHARED_NAME)
export const START_APP_EVENT_STREAM = createStartEventsStreamActionType(SHARED_NAME)
export const START_APP_EVENT_STREAM_SUCCESS = createStartEventsStreamSuccessActionType(SHARED_NAME)
export const START_APP_EVENT_STREAM_FAILURE = createStartEventsStreamFailureActionType(SHARED_NAME)
export const STOP_APP_EVENT_STREAM = createStopEventsStreamActionType(SHARED_NAME)
export const CLEAR_APP_EVENTS = createClearEventsActionType(SHARED_NAME)

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

export const getApplicationApiKey = getApiKey(SHARED_NAME)

export const getApplicationApiKeySuccess = getApiKeySuccess(SHARED_NAME)

export const getApplicationApiKeyFailure = getApiKeyFailure(SHARED_NAME)

export const getApplicationCollaboratorsList = getCollaboratorsList(SHARED_NAME)

export const getApplicationCollaboratorsListSuccess = getCollaboratorsListSuccess(SHARED_NAME)

export const getApplicationCollaboratorsListFailure = getCollaboratorsListFailure(SHARED_NAME)

export const getApplicationCollaboratorPageData = getCollaborator(SHARED_NAME)

export const startApplicationEventsStream = startEventsStream(SHARED_NAME)

export const startApplicationEventsStreamSuccess = startEventsStreamSuccess(SHARED_NAME)

export const startApplicationEventsStreamFailure = startEventsStreamFailure(SHARED_NAME)

export const stopApplicationEventsStream = stopEventsStream(SHARED_NAME)

export const clearApplicationEventsStream = clearEvents(SHARED_NAME)
