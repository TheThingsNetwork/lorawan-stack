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
import { defineMessages } from 'react-intl'
import { useSelector } from 'react-redux'
import { IconPlus } from '@tabler/icons-react'

import Status from '@ttn-lw/components/status'
import Button from '@ttn-lw/components/button'

import Message from '@ttn-lw/lib/components/message'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import { selectGatewayTopEntities } from '@console/store/selectors/top-entities'

import EntitiesList from '../list'

const m = defineMessages({
  emptyMessage: 'No top gateway yet',
  emptyDescription: 'Your most visited, and bookmarked gateways will be listed here',
  emptyAction: 'Create gateway',
})

const TopGatewaysList = () => {
  const items = useSelector(selectGatewayTopEntities)

  const headers = [
    {
      name: 'name',
      displayName: sharedMessages.name,
      getValue: entity => entity,
      render: ({ entity: { name }, id }) =>
        Boolean(name) ? (
          <>
            <span className="mt-0 mb-cs-xxs p-0 fw-bold d-block">{name}</span>
            <span className="c-text-neutral-light d-block">{id}</span>
          </>
        ) : (
          <span className="mt-0 p-0 fw-bold d-block">{id}</span>
        ),
    },
    {
      name: 'status',
      width: '9rem',
      displayName: sharedMessages.status,
      getValue: entity => entity?.entity?.status,
      render: status => {
        let indicator = 'unknown'
        let label = sharedMessages.unknown

        if (status === 'connected') {
          indicator = 'good'
          label = sharedMessages.connected
        } else if (status === 'disconnected') {
          indicator = 'bad'
          label = sharedMessages.disconnected
        } else if (status === 'other-cluster') {
          indicator = 'unknown'
          label = sharedMessages.otherCluster
        } else if (status === 'unknown') {
          indicator = 'mediocre'
          label = sharedMessages.unknown
        }

        return <Status status={indicator} label={label} />
      },
    },
  ]

  return (
    <EntitiesList
      entities={items}
      itemsCount={items.length}
      headers={headers}
      renderWhenEmpty={
        <div className="d-flex direction-column flex-grow j-center gap-cs-l">
          <div>
            <Message content={m.emptyMessage} className="d-block text-center fs-l fw-bold" />
            <Message
              content={m.emptyDescription}
              className="d-block text-center c-text-neutral-light"
            />
          </div>
          <div className="text-center">
            <Button.Link to="/gateways/add" primary message={m.emptyAction} icon={IconPlus} />
          </div>
        </div>
      }
    />
  )
}

export default TopGatewaysList
