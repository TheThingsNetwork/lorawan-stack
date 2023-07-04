// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

import { useSelector } from 'react-redux'

import { APPLICATION, GATEWAY, ORGANIZATION, CLIENT } from '@console/constants/entities'

import { selectCollaboratorById } from '@ttn-lw/lib/store/selectors/collaborators'

import { selectUserId } from '@account/store/selectors/user'
import { selectUserById } from '@console/store/selectors/users'
import {
  selectGatewayRights,
  selectGatewayPseudoRights,
  selectGatewayRightsError,
} from '@console/store/selectors/gateways'
import {
  selectApplicationPseudoRights,
  selectApplicationRights,
  selectApplicationRightsError,
} from '@console/store/selectors/applications'
import {
  selectOrganizationPseudoRights,
  selectOrganizationRights,
  selectOrganizationRightsError,
} from '@console/store/selectors/organizations'
import { selectClientPseudoRights, selectClientRights } from '@account/store/selectors/clients'

const sdkServices = {
  [APPLICATION]: 'Applications',
  [GATEWAY]: 'Gateways',
  [ORGANIZATION]: 'Organizations',
  [CLIENT]: 'Clients',
}

const isCollaboratorUser = collaborator => collaborator.ids && 'user_ids' in collaborator.ids

const useCollaboratorData = (entity, entityId, collaboratorId, tts) => {
  const rightsSelector = {
    [GATEWAY]: selectGatewayRights,
    [APPLICATION]: selectApplicationRights,
    [ORGANIZATION]: selectOrganizationRights,
    [CLIENT]: selectClientRights,
  }
  const pseudoRightsSelector = {
    [GATEWAY]: selectGatewayPseudoRights,
    [APPLICATION]: selectApplicationPseudoRights,
    [ORGANIZATION]: selectOrganizationPseudoRights,
    [CLIENT]: selectClientPseudoRights,
  }
  const righsErrorSelector = {
    [GATEWAY]: selectGatewayRightsError,
    [APPLICATION]: selectApplicationRightsError,
    [ORGANIZATION]: selectOrganizationRightsError,
    [CLIENT]: selectOrganizationRightsError,
  }

  const collaborator = useSelector(state => selectCollaboratorById(state, collaboratorId))
  const rights = useSelector(rightsSelector[entity])
  const pseudoRights = useSelector(pseudoRightsSelector[entity])
  const error = useSelector(righsErrorSelector[entity])
  const update = Boolean(collaborator)
  const isUser = update && isCollaboratorUser(collaborator)
  const admin = useSelector(state => selectUserById(state, collaboratorId))?.admin
  const isAdmin = isUser && admin
  const currentUserId = useSelector(selectUserId)
  const isCurrentUser = isUser && currentUserId === collaboratorId
  const updateCollaborator = patch => tts[sdkServices[entity]].Collaborators.update(entityId, patch)
  const removeCollaborator = collaboratorIds =>
    tts[sdkServices[entity]].Collaborators.remove(entityId, collaboratorIds)

  return {
    collaborator,
    isCollaboratorUser: isUser,
    isCollaboratorAdmin: isAdmin,
    isCollaboratorCurrentUser: isCurrentUser,
    currentUserId,
    rights,
    pseudoRights,
    error,
    updateCollaborator,
    removeCollaborator,
  }
}

export default useCollaboratorData
