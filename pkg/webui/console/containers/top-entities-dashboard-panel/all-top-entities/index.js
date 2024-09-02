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

import { APPLICATION, END_DEVICE, GATEWAY } from '@console/constants/entities'

import Icon, { entityIcons, IconApplication, IconGateway } from '@ttn-lw/components/icon'
import Status from '@ttn-lw/components/status'
import Button from '@ttn-lw/components/button'
import ButtonGroup from '@ttn-lw/components/button/group'

import Message from '@ttn-lw/lib/components/message'

import LastSeen from '@console/components/last-seen'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import { selectTopEntitiesAll } from '@console/store/selectors/top-entities'

import EntitiesList from '../list'

const m = defineMessages({
  noTopEntitiesDescription: 'Your most visited and bookmarked entities will be listed here',
  statusLastSeen: 'Status / Last seen',
})

const AllTopEntitiesList = () => {
  const items = useSelector(selectTopEntitiesAll)

  const headers = [
    {
      name: 'type',
      displayName: sharedMessages.type,
      width: '2.5rem',
      render: type => <Icon icon={entityIcons[type]} />,
    },
    {
      name: 'name',
      displayName: sharedMessages.name,
      align: 'left',
      getValue: entity => entity,
      render: ({ entity, id }) =>
        Boolean(entity?.name) ? (
          <>
            <span className="mt-0 mb-cs-xxs p-0 fw-bold d-block">{entity.name}</span>
            <span className="c-text-neutral-light d-block">{id}</span>
          </>
        ) : (
          <span className="mt-0 p-0 fw-bold d-block">{id}</span>
        ),
    },
    {
      name: 'status',
      displayName: m.statusLastSeen,
      width: '9rem',
      className: 'overflow-visible',
      getValue: entity => {
        if (entity.type === GATEWAY) {
          return entity?.entity?.status
        } else if (entity.type === END_DEVICE) {
          return entity?.entity?.last_seen_at
        } else if (entity.type === APPLICATION) {
          return entity?.entity?.lastSeen
        }
        return null
      },
      render: lastSeen => {
        if (!lastSeen) {
          return (
            <Status
              status="mediocre"
              label={sharedMessages.noRecentActivity}
              className="d-flex j-end al-center"
            />
          )
        }
        let indicator = 'unknown'
        let label = sharedMessages.unknown

        if (lastSeen === 'connected') {
          indicator = 'good'
          label = sharedMessages.connected
        } else if (lastSeen === 'disconnected') {
          indicator = 'bad'
          label = sharedMessages.disconnected
        } else if (lastSeen === 'other-cluster') {
          indicator = 'unknown'
          label = sharedMessages.otherCluster
        } else if (lastSeen === 'unknown') {
          indicator = 'mediocre'
          label = sharedMessages.unknown
        } else if (typeof lastSeen === 'string') {
          return <LastSeen lastSeen={lastSeen} short statusClassName="j-end" />
        }

        return <Status status={indicator} label={label} />
      },
    },
  ]

  return (
    <EntitiesList
      entities={items}
      headers={headers}
      renderWhenEmpty={
        <div className="d-flex direction-column flex-grow j-center gap-cs-l">
          <div>
            <Message
              content={sharedMessages.noTopEntities}
              className="d-block text-center fs-l fw-bold"
            />
            <Message
              content={m.noTopEntitiesDescription}
              className="d-block text-center c-text-neutral-light"
            />
          </div>
          <ButtonGroup align="center">
            <Button.Link
              to="/gateways/add"
              message={sharedMessages.addGateway}
              icon={IconGateway}
              primary
            />
            <Button.Link
              to="/applications/add"
              message={sharedMessages.addApplication}
              icon={IconApplication}
              primary
            />
          </ButtonGroup>
        </div>
      }
    />
  )
}

export default AllTopEntitiesList
