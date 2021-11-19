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

import gatewayIcon from '@assets/misc/gateway.svg'

import EntityTitleSection from '@console/components/entity-title-section'

import GatewayConnection from '@console/containers/gateway-connection'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

const { Content } = EntityTitleSection

const GatewayTitleSection = props => {
  const {
    fetching,
    gtwId,
    gateway,
    apiKeysTotalCount,
    apiKeysErrored,
    collaboratorsTotalCount,
    collaboratorsErrored,
    mayViewCollaborators,
    mayViewApiKeys,
  } = props

  const bottomBarLeft = <GatewayConnection gtwId={gtwId} />
  const bottomBarRight = (
    <>
      {mayViewCollaborators && (
        <Content.EntityCount
          icon="collaborators"
          value={collaboratorsTotalCount}
          keyMessage={sharedMessages.collaboratorCounted}
          errored={collaboratorsErrored}
          toAllUrl={`/gateways/${gtwId}/collaborators`}
        />
      )}
      {mayViewApiKeys && (
        <Content.EntityCount
          icon="api_keys"
          value={apiKeysTotalCount}
          keyMessage={sharedMessages.apiKeyCounted}
          errored={apiKeysErrored}
          toAllUrl={`/gateways/${gtwId}/api-keys`}
        />
      )}
    </>
  )

  return (
    <EntityTitleSection
      id={gtwId}
      name={gateway.name}
      icon={gatewayIcon}
      iconAlt={sharedMessages.gateway}
    >
      <Content
        creationDate={gateway.created_at}
        fetching={fetching}
        bottomBarLeft={bottomBarLeft}
        bottomBarRight={bottomBarRight}
      />
    </EntityTitleSection>
  )
}

GatewayTitleSection.propTypes = {
  apiKeysErrored: PropTypes.bool.isRequired,
  apiKeysTotalCount: PropTypes.number,
  collaboratorsErrored: PropTypes.bool.isRequired,
  collaboratorsTotalCount: PropTypes.number,
  fetching: PropTypes.bool.isRequired,
  gateway: PropTypes.gateway.isRequired,
  gtwId: PropTypes.string.isRequired,
  mayViewApiKeys: PropTypes.bool.isRequired,
  mayViewCollaborators: PropTypes.bool.isRequired,
}

GatewayTitleSection.defaultProps = {
  apiKeysTotalCount: undefined,
  collaboratorsTotalCount: undefined,
}

export default GatewayTitleSection
