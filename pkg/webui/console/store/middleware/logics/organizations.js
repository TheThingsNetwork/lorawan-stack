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

import api from '../../../api'
import * as organizations from '../../actions/organizations'
import { selectUserId } from '../../selectors/user'
import createRequestLogic from './lib'
import createEventsConnectLogics from './events'

const getOrganizationLogic = createRequestLogic({
  type: organizations.GET_ORG,
  async process({ action }, dispatch) {
    const {
      payload: { id },
      meta: { selector },
    } = action
    const org = await api.organization.get(id, selector)
    dispatch(organizations.startOrganizationEventsStream(id))
    return org
  },
})

const getOrganizationsLogic = createRequestLogic({
  type: organizations.GET_ORGS_LIST,
  latest: true,
  async process({ action }) {
    const {
      params: { page, limit },
    } = action.payload
    const { selectors } = action.meta

    const data = await api.organizations.list({ page, limit }, selectors)

    return {
      entities: data.organizations,
      totalCount: data.totalCount,
    }
  },
})

const createOrganizationLogic = createRequestLogic({
  type: organizations.CREATE_ORG,
  async process({ action, getState }) {
    const userId = selectUserId(getState())

    return api.organizations.create(userId, action.payload)
  },
})

const updateOrganizationLogic = createRequestLogic({
  type: organizations.UPDATE_ORG,
  async process({ action }) {
    const { id, patch } = action.payload

    const result = await api.organization.update(id, patch)

    return { ...patch, ...result }
  },
})

const deleteOrganizationLogic = createRequestLogic({
  type: organizations.DELETE_ORG,
  async process({ action }) {
    const { id } = action.payload

    await api.organization.delete(id)

    return { id }
  },
})

const getOrganizationsRightsLogic = createRequestLogic({
  type: organizations.GET_ORGS_RIGHTS_LIST,
  async process({ action }) {
    const { id } = action.payload
    const result = await api.rights.organizations(id)
    return result.rights.sort()
  },
})

const getOrganizationCollaboratorLogic = createRequestLogic({
  type: organizations.GET_ORG_COLLABORATOR,
  async process({ action }) {
    const { id: orgId, collaboratorId, isUser } = action.payload

    const collaborator = isUser
      ? await api.organization.collaborators.getUser(orgId, collaboratorId)
      : await api.organization.collaborators.getOrganization(orgId, collaboratorId)

    const { ids, ...rest } = collaborator

    return {
      id: collaboratorId,
      isUser,
      ...rest,
    }
  },
})

const getOrganizationCollaboratorsLogic = createRequestLogic({
  type: organizations.GET_ORG_COLLABORATORS_LIST,
  async process({ action }) {
    const { id, params } = action.payload
    const res = await api.organization.collaborators.list(id, params)
    const collaborators = res.collaborators.map(function(collaborator) {
      const { ids, ...rest } = collaborator
      const isUser = !!ids.user_ids
      const collaboratorId = isUser ? ids.user_ids.user_id : ids.organization_ids.organization_id

      return {
        id: collaboratorId,
        isUser,
        ...rest,
      }
    })
    return { id, collaborators, totalCount: res.totalCount }
  },
})

export default [
  getOrganizationLogic,
  getOrganizationsLogic,
  createOrganizationLogic,
  updateOrganizationLogic,
  deleteOrganizationLogic,
  getOrganizationsRightsLogic,
  getOrganizationCollaboratorLogic,
  getOrganizationCollaboratorsLogic,
  ...createEventsConnectLogics(
    organizations.SHARED_NAME,
    'organizations',
    api.organization.eventsSubscribe,
  ),
]
