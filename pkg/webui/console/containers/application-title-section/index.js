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
import { defineMessages } from 'react-intl'
import { useSelector } from 'react-redux'

import applicationIcon from '@assets/misc/application.svg'

import Status from '@ttn-lw/components/status'
import DocTooltip from '@ttn-lw/components/tooltip/doc'
import Icon from '@ttn-lw/components/icon'

import Message from '@ttn-lw/lib/components/message'
import RequireRequest from '@ttn-lw/lib/components/require-request'

import LastSeen from '@console/components/last-seen'
import EntityTitleSection from '@console/components/entity-title-section'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'
import { selectCollaboratorsTotalCount } from '@ttn-lw/lib/store/selectors/collaborators'
import { getCollaboratorsList } from '@ttn-lw/lib/store/actions/collaborators'

import {
  checkFromState,
  mayViewOrEditApplicationApiKeys,
  mayViewOrEditApplicationCollaborators,
  mayViewApplicationDevices,
} from '@console/lib/feature-checks'

import { getApiKeysList } from '@console/store/actions/api-keys'
import { getApplicationDeviceCount } from '@console/store/actions/applications'

import {
  selectApplicationById,
  selectApplicationDeviceCount,
  selectApplicationDerivedLastSeen,
} from '@console/store/selectors/applications'
import { selectApiKeysTotalCount } from '@console/store/selectors/api-keys'

const m = defineMessages({
  lastSeenAvailableTooltip:
    'The elapsed time since the network registered activity (sent uplinks, confirmed downlinks or (re)join requests) of the end device(s) in this application.',
  noActivityTooltip:
    'The network has not recently registered any activity (sent uplinks, confirmed downlinks or (re)join requests) of the end device(s) in this application.',
})

const { Content } = EntityTitleSection

const ApplicationTitleSection = ({ appId }) => {
  const apiKeysTotalCount = useSelector(selectApiKeysTotalCount)
  const collaboratorsTotalCount = useSelector(state =>
    selectCollaboratorsTotalCount(state, { id: appId }),
  )
  const devicesTotalCount = useSelector(state => selectApplicationDeviceCount(state, appId))
  const application = useSelector(state => selectApplicationById(state, appId))
  const lastSeen = useSelector(state => selectApplicationDerivedLastSeen(state, appId))
  const mayViewCollaborators = useSelector(state =>
    checkFromState(mayViewOrEditApplicationCollaborators, state),
  )
  const mayViewApiKeys = useSelector(state =>
    checkFromState(mayViewOrEditApplicationApiKeys, state),
  )
  const mayViewDevices = useSelector(state => checkFromState(mayViewApplicationDevices, state))

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
      <Status status="mediocre" label={sharedMessages.noRecentActivity} className="mr-cs-l" flipped>
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
          toAllUrl={`/applications/${appId}/devices`}
        />
      )}
      {mayViewCollaborators && (
        <Content.EntityCount
          icon="collaborators"
          value={collaboratorsTotalCount}
          keyMessage={sharedMessages.collaboratorCounted}
          toAllUrl={`/applications/${appId}/collaborators`}
        />
      )}
      {mayViewApiKeys && (
        <Content.EntityCount
          icon="api_keys"
          value={apiKeysTotalCount}
          keyMessage={sharedMessages.apiKeyCounted}
          toAllUrl={`/applications/${appId}/api-keys`}
        />
      )}
    </>
  )

  const loadData = useCallback(
    async dispatch => {
      if (mayViewCollaborators) {
        dispatch(getCollaboratorsList('application', appId))
      }

      if (mayViewApiKeys) {
        dispatch(getApiKeysList('application', appId))
      }

      if (mayViewDevices) {
        dispatch(getApplicationDeviceCount(appId))
      }
    },
    [appId, mayViewApiKeys, mayViewCollaborators, mayViewDevices],
  )

  return (
    <RequireRequest requestAction={loadData}>
      <EntityTitleSection
        id={appId}
        name={application.name}
        icon={applicationIcon}
        iconAlt={sharedMessages.application}
      >
        <Content bottomBarLeft={bottomBarLeft} bottomBarRight={bottomBarRight} />
      </EntityTitleSection>
    </RequireRequest>
  )
}

ApplicationTitleSection.propTypes = {
  appId: PropTypes.string.isRequired,
}

export default ApplicationTitleSection
