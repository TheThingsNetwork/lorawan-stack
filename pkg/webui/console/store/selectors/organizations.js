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
  GET_ORGS_LIST_BASE,
  GET_ORG_BASE,
  GET_ORG_COLLABORATORS_LIST_BASE,
  GET_ORGS_RIGHTS_LIST_BASE,
  GET_ORG_COLLABORATOR_BASE,
} from '../actions/organizations'
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
import { createRightsSelector, createPseudoRightsSelector } from './rights'
import {
  createCollaboratorsSelector,
  createTotalCountSelector as createCollaboratorsTotalCountSelector,
  createUserCollaboratorSelector,
  createOrganizationCollaboratorSelector,
} from './collaborators'

const ENTITY = 'organizations'
const ENTITY_SINGLE = 'organization'

// Organization
export const selectOrganizationStore = state => state.organizations
export const selectOrganizationEntitiesStore = state => selectOrganizationStore(state).entities
export const selectOrganizationById = (state, id) => selectOrganizationEntitiesStore(state)[id]
export const selectSelectedOrganizationId = state =>
  selectOrganizationStore(state).selectedOrganization
export const selectSelectedOrganization = state =>
  selectOrganizationById(state, selectSelectedOrganizationId(state))
export const selectOrganizationFetching = createFetchingSelector(GET_ORG_BASE)
export const selectOrganizationError = createErrorSelector(GET_ORG_BASE)

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

// Rights
export const selectOrganizationRights = createRightsSelector(ENTITY)
export const selectOrganizationPseudoRights = createPseudoRightsSelector(ENTITY)
export const selectOrganizationRightsError = createErrorSelector(GET_ORGS_RIGHTS_LIST_BASE)
export const selectOrganizationRightsFetching = createFetchingSelector(GET_ORGS_RIGHTS_LIST_BASE)

// Events
export const selectOrganizationEvents = createEventsSelector(ENTITY)
export const selectOrganizationEventsError = createEventsErrorSelector(ENTITY)
export const selectOrganizationEventsStatus = createEventsStatusSelector(ENTITY)

// Collaborators
export const selectOrganizationCollaborators = createCollaboratorsSelector(ENTITY)
export const selectOrganizationCollaboratorsTotalCount = createCollaboratorsTotalCountSelector(
  ENTITY,
)
export const selectOrganizationCollaboratorsFetching = createFetchingSelector(
  GET_ORG_COLLABORATORS_LIST_BASE,
)
export const selectOrganizationUserCollaborator = createUserCollaboratorSelector(ENTITY_SINGLE)
export const selectOrganizationOrganizationCollaborator = createOrganizationCollaboratorSelector(
  ENTITY_SINGLE,
)
export const selectOrganizationCollaboratorFetching = createFetchingSelector(
  GET_ORG_COLLABORATOR_BASE,
)
export const selectOrganizationCollaboratorError = createErrorSelector(GET_ORG_COLLABORATOR_BASE)
