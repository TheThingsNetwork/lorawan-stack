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

import { GET_ORGS_LIST_BASE } from '../actions/organizations'
import {
  createPaginationIdsSelectorByEntity,
  createPaginationTotalCountSelectorByEntity,
} from './pagination'
import {
  createEventsSelector,
  createEventsErrorSelector,
  createEventsStatusSelector,
} from './events'
import { createFetchingSelector } from './fetching'
import { createErrorSelector } from './error'

const ENTITY = 'organizations'

// Organization
export const selectOrganizationStore = state => state.organizations
export const selectOrganizationEntitiesStore = state => selectOrganizationStore(state).entities
export const selectOrganizationById = (state, id) => selectOrganizationEntitiesStore(state)[id]

// Organizations
const selectOrgsIds = createPaginationIdsSelectorByEntity(ENTITY)
const selectOrgsTotalCount = createPaginationTotalCountSelectorByEntity(ENTITY)
const selectOrgsFetching = createFetchingSelector(GET_ORGS_LIST_BASE)
const selectOrgsError = createErrorSelector(GET_ORGS_LIST_BASE)

export const selectOrganizations = state =>
  selectOrgsIds(state).map(id => selectOrganizationById(state, id))
export const selectOrganizationsTotalCount = state => selectOrgsTotalCount(state)
export const selectOrganizationsFetching = state => selectOrgsFetching(state)
export const selectOrganizationsError = state => selectOrgsError(state)

// Events
export const selectOrganizationEvents = createEventsSelector(ENTITY)
export const selectOrganizationEventsError = createEventsErrorSelector(ENTITY)
export const selectOrganizationEventsStatus = createEventsStatusSelector(ENTITY)
