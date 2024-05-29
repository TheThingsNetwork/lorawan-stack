// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

import React, { useCallback } from 'react'
import { useSelector } from 'react-redux'

import organizationIcon from '@assets/misc/organization.svg'

import { IconApiKeys, IconCollaborators } from '@ttn-lw/components/icon'

import RequireRequest from '@ttn-lw/lib/components/require-request'

import EntityTitleSection from '@console/components/entity-title-section'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import { selectCollaboratorsTotalCount } from '@ttn-lw/lib/store/selectors/collaborators'
import { getCollaboratorsList } from '@ttn-lw/lib/store/actions/collaborators'

import {
  checkFromState,
  mayViewOrEditOrganizationApiKeys,
  mayViewOrEditOrganizationCollaborators,
} from '@console/lib/feature-checks'

import { getApiKeysList } from '@console/store/actions/api-keys'

import { selectApiKeysTotalCount } from '@console/store/selectors/api-keys'
import {
  selectSelectedOrganization,
  selectSelectedOrganizationId,
} from '@console/store/selectors/organizations'

const { Content } = EntityTitleSection

const OrganizationTitleSection = () => {
  const organization = useSelector(selectSelectedOrganization)
  const orgId = useSelector(selectSelectedOrganizationId)
  const apiKeysTotalCount = useSelector(selectApiKeysTotalCount)
  const collaboratorsTotalCount = useSelector(state =>
    selectCollaboratorsTotalCount(state, { id: orgId }),
  )
  const mayViewCollaborators = useSelector(state =>
    checkFromState(mayViewOrEditOrganizationCollaborators, state),
  )
  const mayViewApiKeys = useSelector(state =>
    checkFromState(mayViewOrEditOrganizationApiKeys, state),
  )

  const loadData = useCallback(
    async dispatch => {
      if (mayViewCollaborators) {
        dispatch(getCollaboratorsList('organization', orgId))
      }

      if (mayViewApiKeys) {
        dispatch(getApiKeysList('organization', orgId))
      }
    },
    [mayViewApiKeys, mayViewCollaborators, orgId],
  )

  return (
    <RequireRequest requestAction={loadData}>
      <EntityTitleSection
        id={orgId}
        name={organization.name}
        icon={organizationIcon}
        iconAlt={sharedMessages.organization}
      >
        <Content creationDate={organization.created_at}>
          {mayViewCollaborators && (
            <Content.EntityCount
              icon={IconCollaborators}
              value={collaboratorsTotalCount}
              keyMessage={sharedMessages.collaboratorCounted}
              toAllUrl={`/organizations/${orgId}/collaborators`}
            />
          )}
          {mayViewApiKeys && (
            <Content.EntityCount
              icon={IconApiKeys}
              value={apiKeysTotalCount}
              keyMessage={sharedMessages.apiKeyCounted}
              toAllUrl={`/organizations/${orgId}/api-keys`}
            />
          )}
        </Content>
      </EntityTitleSection>
    </RequireRequest>
  )
}

export default OrganizationTitleSection
