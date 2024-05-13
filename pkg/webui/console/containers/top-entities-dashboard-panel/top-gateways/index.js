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

import Status from '@ttn-lw/components/status'

import Message from '@ttn-lw/lib/components/message'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import {
  selectPerEntityBookmarks,
  selectPerEntityTotalCount,
} from '@console/store/selectors/user-preferences'

import EntitiesList from '../list'

const m = defineMessages({
  emptyMessage: 'No top gateway yet',
  emptyDescription: 'Your most visited, and bookmarked gateways will be listed here',
  emptyAction: 'Create gateway',
})

const TopGatewaysList = () => {
  const allBookmarks = useSelector(selectPerEntityBookmarks('gateway'))

  const headers = [
    {
      name: 'name',
      displayName: sharedMessages.name,
      render: (name, id) => (
        <>
          <Message content={name === '' ? id : name} component="p" className="mt-0 mb-cs-xs p-0" />
          {name && (
            <Message content={id} component="span" className="c-text-neutral-light fw-normal" />
          )}
        </>
      ),
    },
    {
      name: 'lastSeen',
      displayName: sharedMessages.status,
      render: lastSeen => {
        let indicator = 'unknown'
        let label = sharedMessages.unknown

        if (lastSeen.status === 'connected') {
          indicator = 'good'
          label = sharedMessages.connected
        } else if (lastSeen.status === 'disconnected') {
          indicator = 'bad'
          label = sharedMessages.disconnected
        } else if (lastSeen.status === 'other-cluster') {
          indicator = 'unknown'
          label = sharedMessages.otherCluster
        } else if (lastSeen.status === 'unknown') {
          indicator = 'mediocre'
          label = sharedMessages.unknown
        }

        return <Status status={indicator} label={label} />
      },
    },
  ]

  return (
    <EntitiesList
      allBookmarks={allBookmarks}
      itemsCountSelector={selectPerEntityTotalCount}
      headers={headers}
      emptyMessage={m.emptyMessage}
      emptyDescription={m.emptyDescription}
      emptyAction={m.emptyAction}
      emptyPath={'/gateways/add'}
      entity={'gateway'}
    />
  )
}

export default TopGatewaysList
