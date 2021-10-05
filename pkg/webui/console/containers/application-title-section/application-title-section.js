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
import { defineMessages } from 'react-intl'

import applicationIcon from '@assets/misc/application.svg'

import Status from '@ttn-lw/components/status'
import DocTooltip from '@ttn-lw/components/tooltip/doc'
import Icon from '@ttn-lw/components/icon'

import Message from '@ttn-lw/lib/components/message'

import LastSeen from '@console/components/last-seen'
import EntityTitleSection from '@console/components/entity-title-section'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import style from './application-title-section.styl'

const m = defineMessages({
  lastSeenAvailableTooltip:
    'The elapsed time since the network registered activity (sent uplinks, confirmed downlinks or (re)join requests) of the end device(s) in this application.',
  noActivityTooltip:
    'The network has not recently registered any activity (sent uplinks, confirmed downlinks or (re)join requests) of the end device(s) in this application.',
})

const { Content } = EntityTitleSection

const ApplicationTitleSection = props => {
  const {
    appId,
    fetching,
    application,
    apiKeysTotalCount,
    apiKeysErrored,
    collaboratorsTotalCount,
    collaboratorsErrored,
    devicesTotalCount,
    devicesErrored,
    lastSeen,
    mayViewCollaborators,
    mayViewApiKeys,
    mayViewDevices,
  } = props

  const showLastSeen = Boolean(lastSeen)

  const bottomBarLeft = showLastSeen ? (
    <DocTooltip
      interactive
      docPath="/getting-started/console/troubleshooting"
      content={<Message content={m.lastSeenAvailableTooltip} />}
    >
      <LastSeen lastSeen={lastSeen} flipped>
        <Icon icon="help_outline" textPaddedLeft small nudgeUp className="tc-subtle-gray" />
      </LastSeen>
    </DocTooltip>
  ) : (
    <DocTooltip
      content={<Message content={m.noActivityTooltip} />}
      docPath="/getting-started/console/troubleshooting"
    >
      <Status
        status="mediocre"
        label={sharedMessages.noRecentActivity}
        className={style.lastSeen}
        flipped
      >
        <Icon icon="help_outline" textPaddedLeft small nudgeUp className="tc-subtle-gray" />
      </Status>
    </DocTooltip>
  )
  const bottomBarRight = (
    <>
      {mayViewDevices && (
        <Content.EntityCount
          icon="devices"
          value={devicesTotalCount}
          keyMessage={sharedMessages.deviceCounted}
          errored={devicesErrored}
          toAllUrl={`/applications/${appId}/devices`}
        />
      )}
      {mayViewCollaborators && (
        <Content.EntityCount
          icon="collaborators"
          value={collaboratorsTotalCount}
          keyMessage={sharedMessages.collaboratorCounted}
          errored={collaboratorsErrored}
          toAllUrl={`/applications/${appId}/collaborators`}
        />
      )}
      {mayViewApiKeys && (
        <Content.EntityCount
          icon="api_keys"
          value={apiKeysTotalCount}
          keyMessage={sharedMessages.apiKeyCounted}
          errored={apiKeysErrored}
          toAllUrl={`/applications/${appId}/api-keys`}
        />
      )}
    </>
  )

  return (
    <EntityTitleSection
      id={appId}
      name={application.name}
      icon={applicationIcon}
      iconAlt={sharedMessages.application}
    >
      <Content fetching={fetching} bottomBarLeft={bottomBarLeft} bottomBarRight={bottomBarRight} />
    </EntityTitleSection>
  )
}

ApplicationTitleSection.propTypes = {
  apiKeysErrored: PropTypes.bool.isRequired,
  apiKeysTotalCount: PropTypes.number,
  appId: PropTypes.string.isRequired,
  application: PropTypes.application.isRequired,
  collaboratorsErrored: PropTypes.bool.isRequired,
  collaboratorsTotalCount: PropTypes.number,
  devicesErrored: PropTypes.bool.isRequired,
  devicesTotalCount: PropTypes.number,
  fetching: PropTypes.bool.isRequired,
  lastSeen: PropTypes.string,
  mayViewApiKeys: PropTypes.bool.isRequired,
  mayViewCollaborators: PropTypes.bool.isRequired,
  mayViewDevices: PropTypes.bool.isRequired,
}

ApplicationTitleSection.defaultProps = {
  apiKeysTotalCount: undefined,
  collaboratorsTotalCount: undefined,
  devicesTotalCount: undefined,
  lastSeen: undefined,
}

export default ApplicationTitleSection
