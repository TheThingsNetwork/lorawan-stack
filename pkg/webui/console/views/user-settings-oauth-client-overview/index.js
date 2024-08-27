// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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
import { useSelector } from 'react-redux'
import { useIntl } from 'react-intl'

import applicationIcon from '@assets/misc/application.svg'

import { IconCollaborators } from '@ttn-lw/components/icon'
import DataSheet from '@ttn-lw/components/data-sheet'

import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'
import DateTime from '@ttn-lw/lib/components/date-time'

import EntityTitleSection from '@console/components/entity-title-section'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import capitalizeMessage from '@ttn-lw/lib/capitalize-message'
import { selectCollaboratorsTotalCount } from '@ttn-lw/lib/store/selectors/collaborators'

import { mayPerformAdminActions } from '@console/lib/feature-checks'

import {
  selectSelectedClient,
  selectClientFetching,
  selectClientRights,
} from '@console/store/selectors/clients'

import style from './user-settings-oauth-client-overview.styl'

const { Content } = EntityTitleSection

const OAuthClientOverview = () => {
  const oauthClient = useSelector(selectSelectedClient)
  const oauthClientId = oauthClient.ids.client_id || oauthClient.name
  const { created_at, updated_at, state, state_description, secret, name } = oauthClient
  const rights = useSelector(selectClientRights)
  const includeStateDescription = rights.includes('RIGHT_CLIENT_ALL')
  const collaboratorsTotalCount = useSelector(selectCollaboratorsTotalCount)
  const fetching = useSelector(selectClientFetching)

  const { formatMessage } = useIntl()

  const sheetData = [
    {
      header: sharedMessages.generalInformation,
      items: [
        { key: sharedMessages.oauthClientId, value: oauthClientId, type: 'code', sensitive: false },
        { key: sharedMessages.createdAt, value: <DateTime value={created_at} /> },
        { key: sharedMessages.updatedAt, value: <DateTime value={updated_at} /> },
        {
          key: sharedMessages.state,
          value: capitalizeMessage(formatMessage({ id: `enum:${state}` })),
        },
      ],
    },
  ]

  // Add secret, if it is available.
  if (secret) {
    sheetData[0].items.push({
      key: sharedMessages.secret,
      value: secret,
      type: 'byte',
      sensitive: true,
      enableUint32: true,
    })
  }

  // Include `state_description`.
  if (includeStateDescription) {
    sheetData[0].items.push({
      key: sharedMessages.stateDescription,
      value: state_description,
    })
  }

  const bottomBarLeft = (
    <>
      {mayPerformAdminActions && (
        <Content.EntityCount
          icon={IconCollaborators}
          value={collaboratorsTotalCount}
          keyMessage={sharedMessages.collaboratorCounted}
          errored={false}
          toAllUrl={`/oauth-clients/${oauthClientId}/collaborators`}
        />
      )}
    </>
  )

  return (
    <>
      <div className={style.titleSection}>
        <div className="container container--lg grid">
          <IntlHelmet title={sharedMessages.overview} />
          <div className="item-12">
            <EntityTitleSection
              id={oauthClientId}
              name={name}
              icon={applicationIcon}
              iconAlt={sharedMessages.overview}
            >
              <Content fetching={fetching} bottomBarLeft={bottomBarLeft} />
            </EntityTitleSection>
          </div>
        </div>
      </div>
      <div className="container container--lg grid">
        <div className="item-12 xl:item-6">
          <DataSheet data={sheetData} />
        </div>
      </div>
    </>
  )
}

export default OAuthClientOverview
