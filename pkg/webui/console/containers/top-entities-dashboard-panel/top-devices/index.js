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

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'

import {
  selectEndDeviceBookmarks,
  selectPerEntityTotalCount,
} from '@console/store/selectors/user-preferences'

import EntitiesList from '../list'

const m = defineMessages({
  emptyMessage: 'No top Device yet',
  emptyDescription: 'Your most visited, and bookmarked Devices will be listed here.',
})

const TopDevicesList = ({ loadNextPage }) => {
  const headers = [
    {
      name: 'name',
      displayName: 'Name',
      render: (name, id) => (
        <>
          <Message content={name === '' ? id : name} component="p" className="mt-0 mb-cs-xs p-0" />
          {name && (
            <Message content={id} component="span" className="c-text-neutral-light fw-normal" />
          )}
        </>
      ),
    },
  ]

  return (
    <EntitiesList
      loadNextPage={loadNextPage}
      itemsCountSelector={selectPerEntityTotalCount}
      itemsSelector={selectEndDeviceBookmarks}
      headers={headers}
      emptyMessage={m.emptyMessage}
      emptyDescription={m.emptyDescription}
      entity={'device'}
    />
  )
}

TopDevicesList.propTypes = {
  loadNextPage: PropTypes.func.isRequired,
}

export default TopDevicesList
