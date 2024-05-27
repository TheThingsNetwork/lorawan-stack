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

import { createSelector } from 'reselect'

import {
  createPaginationIdsSelectorByEntity,
  createPaginationTotalCountSelectorByEntity,
} from '@ttn-lw/lib/store/selectors/pagination'
import { createFetchingSelector } from '@ttn-lw/lib/store/selectors/fetching'
import { createErrorSelector } from '@ttn-lw/lib/store/selectors/error'
import { getOrganizationId } from '@ttn-lw/lib/selectors/id'

import { GET_ORGS_LIST_BASE, GET_ORGS_RIGHTS_LIST_BASE } from '@console/store/actions/organizations'

import {
  createEventsSelector,
  createEventsErrorSelector,
  createEventsStatusSelector,
  createEventsInterruptedSelector,
  createEventsPausedSelector,
  createEventsTruncatedSelector,
  createEventsFilterSelector,
} from './events'
import { createRightsSelector, createPseudoRightsSelector } from './rights'

const ENTITY = 'organizations'

// Organization.
export const selectOrganizationStore = state => state.organizations
export const selectOrganizationEntitiesStore = state => selectOrganizationStore(state).entities
export const selectOrganizationById = (state, id) => selectOrganizationEntitiesStore(state)[id]
export const selectSelectedOrganizationId = state =>
  selectOrganizationStore(state).selectedOrganization
export const selectSelectedOrganization = state =>
  selectOrganizationById(state, selectSelectedOrganizationId(state))
export const selectOrganizationCollaboratorCounts = state =>
  selectOrganizationStore(state).collaboratorCounts
export const selectOrganizationCollaboratorCount = (state, id) =>
  selectOrganizationCollaboratorCounts(state)?.[id] || 0

// Organizations.
const selectOrgsIds = createPaginationIdsSelectorByEntity(ENTITY)
const selectOrgsTotalCount = createPaginationTotalCountSelectorByEntity(ENTITY)
const selectOrgsFetching = createFetchingSelector(GET_ORGS_LIST_BASE)
const selectOrgsError = createErrorSelector(GET_ORGS_LIST_BASE)

export const selectOrganizations = createSelector(
  [selectOrgsIds, selectOrganizationEntitiesStore],
  (ids, entities) => ids.map(id => entities[id]),
)
export const selectOrganizationsTotalCount = state => selectOrgsTotalCount(state)
export const selectOrganizationsFetching = state => selectOrgsFetching(state)
export const selectOrganizationsError = state => selectOrgsError(state)
export const selectOrganizationsWithCollaboratorCount = createSelector(
  [selectOrganizations, selectOrganizationCollaboratorCounts],
  (orgs, collaboratorCounts) =>
    orgs.map(org => ({
      ...org,
      _collaboratorCount: collaboratorCounts[getOrganizationId(org)] || 0,
    })),
)

// Rights.
export const selectOrganizationRights = createRightsSelector(ENTITY)
export const selectOrganizationPseudoRights = createPseudoRightsSelector(ENTITY)
export const selectOrganizationRightsError = createErrorSelector(GET_ORGS_RIGHTS_LIST_BASE)
export const selectOrganizationRightsFetching = createFetchingSelector(GET_ORGS_RIGHTS_LIST_BASE)

// Events.
export const selectOrganizationEvents = createEventsSelector(ENTITY)
export const selectOrganizationEventsError = createEventsErrorSelector(ENTITY)
export const selectOrganizationEventsStatus = createEventsStatusSelector(ENTITY)
export const selectOrganizationEventsInterrupted = createEventsInterruptedSelector(ENTITY)
export const selectOrganizationEventsPaused = createEventsPausedSelector(ENTITY)
export const selectOrganizationEventsTruncated = createEventsTruncatedSelector(ENTITY)
export const selectOrganizationEventsFilter = createEventsFilterSelector(ENTITY)
