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

import React from 'react'

import organizationIcon from '@assets/misc/organization.svg'

import Spinner from '@ttn-lw/components/spinner'

import Message from '@ttn-lw/lib/components/message'

import EntityTitleSection from '@console/components/entity-title-section'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

const { Content } = EntityTitleSection

const OrganizationTitleSection = props => {
  const {
    fetching,
    orgId,
    organization,
    apiKeysTotalCount,
    apiKeysErrored,
    collaboratorsTotalCount,
    collaboratorsErrored,
    mayViewCollaborators,
    mayViewApiKeys,
  } = props

  return (
    <EntityTitleSection
      id={orgId}
      name={organization.name}
      icon={organizationIcon}
      iconAlt={sharedMessages.organization}
    >
      <Content creationDate={organization.created_at}>
        {fetching ? (
          <Spinner after={0} faded micro inline>
            <Message content={sharedMessages.fetching} />
          </Spinner>
        ) : (
          <>
            {mayViewCollaborators && (
              <Content.EntityCount
                icon="collaborators"
                value={collaboratorsTotalCount}
                keyMessage={sharedMessages.collaboratorCounted}
                errored={collaboratorsErrored}
                toAllUrl={`/organizations/${orgId}/collaborators`}
              />
            )}
            {mayViewApiKeys && (
              <Content.EntityCount
                icon="api_keys"
                value={apiKeysTotalCount}
                keyMessage={sharedMessages.apiKeyCounted}
                errored={apiKeysErrored}
                toAllUrl={`/organizations/${orgId}/api-keys`}
              />
            )}
          </>
        )}
      </Content>
    </EntityTitleSection>
  )
}

OrganizationTitleSection.propTypes = {
  apiKeysErrored: PropTypes.bool.isRequired,
  apiKeysTotalCount: PropTypes.number,
  collaboratorsErrored: PropTypes.bool.isRequired,
  collaboratorsTotalCount: PropTypes.number,
  fetching: PropTypes.bool.isRequired,
  mayViewApiKeys: PropTypes.bool.isRequired,
  mayViewCollaborators: PropTypes.bool.isRequired,
  orgId: PropTypes.string.isRequired,
  organization: PropTypes.organization.isRequired,
}

OrganizationTitleSection.defaultProps = {
  apiKeysTotalCount: undefined,
  collaboratorsTotalCount: undefined,
}

export default OrganizationTitleSection
