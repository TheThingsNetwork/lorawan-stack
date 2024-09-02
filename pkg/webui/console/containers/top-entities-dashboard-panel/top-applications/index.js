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
import { FormattedNumber, defineMessages } from 'react-intl'
import { useSelector } from 'react-redux'
import { IconPlus } from '@tabler/icons-react'

import Spinner from '@ttn-lw/components/spinner'
import Status from '@ttn-lw/components/status'
import Button from '@ttn-lw/components/button'

import Message from '@ttn-lw/lib/components/message'

import LastSeen from '@console/components/last-seen'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import { selectApplicationTopEntities } from '@console/store/selectors/top-entities'

import EntitiesList from '../list'

const m = defineMessages({
  emptyMessage: 'No top application yet',
  emptyDescription: 'Your most visited, and bookmarked applications will be listed here',
})

const TopApplicationsList = () => {
  const items = useSelector(selectApplicationTopEntities)

  const headers = [
    {
      name: 'name',
      displayName: sharedMessages.name,
      getValue: entity => entity,
      render: ({ entity, id }) =>
        Boolean(entity?.name) ? (
          <>
            <span className="mt-0 mb-cs-xxs p-0 fw-bold d-block">{name}</span>
            <span className="c-text-neutral-light d-block">{id}</span>
          </>
        ) : (
          <span className="mt-0 p-0 fw-bold d-block">{id}</span>
        ),
    },
    {
      name: 'deviceCount',
      displayName: sharedMessages.devicesShort,
      width: '4rem',
      align: 'center',
      getValue: entity => entity?.entity,
      render: entity =>
        !Boolean(entity) ? (
          <span className="c-text-neutral-light">n/a</span>
        ) : typeof entity?.deviceCount !== 'number' ? (
          <Spinner micro center after={0} faded inline />
        ) : (
          <FormattedNumber value={entity?.deviceCount} />
        ),
    },
    {
      name: 'lastSeen',
      width: '9rem',
      getValue: entity => entity?.entity?.lastSeen,
      displayName: sharedMessages.lastSeen,
      render: lastSeen => {
        const showLastSeen = Boolean(lastSeen)
        return showLastSeen ? (
          <LastSeen lastSeen={lastSeen} short statusClassName="j-end" />
        ) : (
          <Status
            status="mediocre"
            label={sharedMessages.noRecentActivity}
            className="d-flex j-end al-center"
          />
        )
      },
    },
  ]

  return (
    <EntitiesList
      itemsCount={items.length}
      entities={items}
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
            <Button.Link
              to="/application/add"
              primary
              message={sharedMessages.addApplication}
              icon={IconPlus}
            />
          </div>
        </div>
      }
    />
  )
}

export default TopApplicationsList
