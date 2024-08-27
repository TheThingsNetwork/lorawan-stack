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

import createRequestActions from '@ttn-lw/lib/store/actions/create-request-actions'
import {
  createPaginationRequestActions,
  createPaginationBaseActionType,
  createPaginationDeleteBaseActionType,
  createPaginationDeleteActions,
  createPaginationRestoreBaseActionType,
  createPaginationRestoreActions,
} from '@ttn-lw/lib/store/actions/pagination'

import {
  startEventsStream,
  createStartEventsStreamActionType,
  startEventsStreamSuccess,
  createStartEventsStreamSuccessActionType,
  startEventsStreamFailure,
  createStartEventsStreamFailureActionType,
  pauseEventsStream,
  createPauseEventsStreamActionType,
  resumeEventsStream,
  createResumeEventsStreamActionType,
  stopEventsStream,
  createStopEventsStreamActionType,
  clearEvents,
  createClearEventsActionType,
  createSetEventsFilterActionType,
  setEventsFilter,
} from './events'
import createGetRightsListRequestActions, { createGetRightsListActionType } from './rights'

export const SHARED_NAME = 'ORGANIZATION'

export const GET_ORGS_LIST_BASE = createPaginationBaseActionType(SHARED_NAME)
export const [
  { request: GET_ORGS_LIST, success: GET_ORGS_LIST_SUCCESS, failure: GET_ORGS_LIST_FAILURE },
  {
    request: getOrganizationsList,
    success: getOrganizationsListSuccess,
    failure: getORganizationsListFailure,
  },
] = createPaginationRequestActions(SHARED_NAME)

export const GET_ORG_BASE = 'GET_ORGANIZATION'
export const [
  { request: GET_ORG, success: GET_ORG_SUCCESS, failure: GET_ORG_FAILURE },
  { request: getOrganization, success: getOrganizationSuccess, failure: getOrganizationFailure },
] = createRequestActions(
  GET_ORG_BASE,
  id => ({ id }),
  (id, selector, options = {}) => ({ selector, options }),
)

export const GET_ORG_COLLABORATOR_COUNT_BASE = 'GET_ORGANIZATION_COLLABORATOR_COUNT'
export const [
  {
    request: GET_ORG_COLLABORATOR_COUNT,
    success: GET_ORG_COLLABORATOR_COUNT_SUCCESS,
    failure: GET_ORG_COLLABORATOR_COUNT_FAILURE,
  },
  {
    request: getOrganizationCollaboratorCount,
    success: getOrganizationCollaboratorCountSuccess,
    failure: getOrganizationCollaboratorCountFailure,
  },
] = createRequestActions(GET_ORG_COLLABORATOR_COUNT_BASE, id => ({
  id,
}))

export const CREATE_ORG_BASE = 'CREATE_ORGANIZATION'
export const [
  { request: CREATE_ORG, success: CREATE_ORG_SUCCESS, failure: CREATE_ORG_FAILURE },
  {
    request: createOrganization,
    success: createOrganizationSuccess,
    failure: createOrganizationFailure,
  },
] = createRequestActions(CREATE_ORG_BASE)

export const UPDATE_ORG_BASE = 'UPDATE_ORGANIZATION'
export const [
  { request: UPDATE_ORG, success: UPDATE_ORG_SUCCESS, failure: UPDATE_ORG_FAILURE },
  {
    request: updateOrganization,
    success: updateOrganizationSuccess,
    failure: updateOrganizationFailure,
  },
] = createRequestActions(UPDATE_ORG_BASE, (id, patch) => ({ id, patch }))

export const GET_ORGS_RIGHTS_LIST_BASE = createGetRightsListActionType(SHARED_NAME)
export const [
  {
    request: GET_ORGS_RIGHTS_LIST,
    success: GET_ORGS_RIGHTS_LIST_SUCCESS,
    failure: GET_ORGS_RIGHTS_LIST_FAILURE,
  },
  {
    request: getOrganizationsRightsList,
    success: getOrganizationsRightsListSuccess,
    failure: getOrganizationsRightsListFailure,
  },
] = createGetRightsListRequestActions(SHARED_NAME)

export const DELETE_ORG_BASE = createPaginationDeleteBaseActionType(SHARED_NAME)
export const [
  { request: DELETE_ORG, success: DELETE_ORG_SUCCESS, failure: DELETE_ORG_FAILURE },
  {
    request: deleteOrganization,
    success: deleteOrganizationSuccess,
    failure: deleteORganizationFailure,
  },
] = createPaginationDeleteActions(SHARED_NAME, id => ({ id }))

export const RESTORE_ORG_BASE = createPaginationRestoreBaseActionType(SHARED_NAME)
export const [
  { request: RESTORE_ORG, success: RESTORE_ORG_SUCCESS, failure: RESTORE_ORG_FAILURE },
  {
    request: restoreOrganization,
    success: restoreOrganizationSuccess,
    failure: restoreOrganizationFailure,
  },
] = createPaginationRestoreActions(SHARED_NAME, id => ({ id }))

export const START_ORG_EVENT_STREAM = createStartEventsStreamActionType(SHARED_NAME)
export const START_ORG_EVENT_STREAM_SUCCESS = createStartEventsStreamSuccessActionType(SHARED_NAME)
export const START_ORG_EVENT_STREAM_FAILURE = createStartEventsStreamFailureActionType(SHARED_NAME)
export const PAUSE_ORG_EVENT_STREAM = createPauseEventsStreamActionType(SHARED_NAME)
export const RESUME_ORG_EVENT_STREAM = createResumeEventsStreamActionType(SHARED_NAME)
export const STOP_ORG_EVENT_STREAM = createStopEventsStreamActionType(SHARED_NAME)
export const CLEAR_ORG_EVENTS = createClearEventsActionType(SHARED_NAME)
export const SET_ORG_EVENTS_FILTER = createSetEventsFilterActionType(SHARED_NAME)

export const startOrganizationEventsStream = startEventsStream(SHARED_NAME)
export const startOrganizationEventsStreamSuccess = startEventsStreamSuccess(SHARED_NAME)
export const startOrganizationEventsStreamFailure = startEventsStreamFailure(SHARED_NAME)
export const pauseOrganizationEventsStream = pauseEventsStream(SHARED_NAME)
export const resumeOrganizationEventsStream = resumeEventsStream(SHARED_NAME)
export const stopOrganizationEventsStream = stopEventsStream(SHARED_NAME)
export const clearOrganizationEventsStream = clearEvents(SHARED_NAME)
export const setOrganizationEventsFilter = setEventsFilter(SHARED_NAME)
