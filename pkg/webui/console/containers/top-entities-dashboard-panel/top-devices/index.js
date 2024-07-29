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

import React, { useMemo } from 'react'
import { useSelector } from 'react-redux'

import Status from '@ttn-lw/components/status'

import LastSeen from '@console/components/last-seen'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import { selectEndDeviceTopEntities } from '@console/store/selectors/top-entities'

import EntitiesList from '../list'

const TopDevicesList = ({ appId }) => {
  const topEntityFilter = useMemo(() => (appId ? e => e.id.startsWith(appId) : undefined), [appId])
  const items = useSelector(state => selectEndDeviceTopEntities(state, topEntityFilter))

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
      name: 'last_seen_at',
      width: '9rem',
      getValue: entity => entity?.entity?.last_seen_at,
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
      entities={items}
      itemsCount={items.length}
      headers={headers}
      emptyMessage={sharedMessages.noTopEndDevices}
      emptyDescription={sharedMessages.noTopEndDevicesDescription}
      emptyAction={appId ? sharedMessages.registerEndDevice : undefined}
      emptyPath={appId ? `/applications/${appId}/devices/add` : undefined}
    />
  )
}

TopDevicesList.propTypes = {
  appId: PropTypes.string,
}

TopDevicesList.defaultProps = {
  appId: undefined,
}

export default TopDevicesList
