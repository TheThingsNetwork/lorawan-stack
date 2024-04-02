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

import React, { useCallback, useState } from 'react'
import { useDispatch, useSelector } from 'react-redux'
import { defineMessages } from 'react-intl'
import classNames from 'classnames'

import { IconStar } from '@ttn-lw/components/icon'
import Panel from '@ttn-lw/components/panel'

import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import { getBookmarksList } from '@console/store/actions/user-preferences'

import { selectBookmarksList } from '@console/store/selectors/user-preferences'

import AllTopEntitiesList from './all-top-entities'
import TopApplicationsList from './top-applications'
import TopGatewaysList from './top-gateways'
import TopDevicesList from './top-devices'

import styles from './top-entities-panel.styl'

const BATCH_SIZE = 20

const indicesToPage = (startIndex, stopIndex, limit) => {
  const startPage = Math.floor(startIndex / limit) + 1
  const stopPage = Math.floor(stopIndex / limit) + 1
  return [startPage, stopPage]
}

const m = defineMessages({
  title: 'Your top entities',
})

const TopEntitiesDashboardPanel = () => {
  const [active, setActive] = useState('all')
  const [fetching, setFetching] = useState(false)
  const bookmarks = useSelector(state => selectBookmarksList(state))
  const hasEntities = bookmarks.length > 0
  const dispatch = useDispatch()

  const handleChange = useCallback(
    (_, value) => {
      setActive(value)
    },
    [setActive],
  )

  const options = [
    { label: sharedMessages.all, value: 'all' },
    { label: sharedMessages.applications, value: 'applications' },
    { label: sharedMessages.gateways, value: 'gateways' },
    { label: sharedMessages.devices, value: 'end-devices' },
  ]

  const loadNextPage = useCallback(
    async (startIndex, stopIndex) => {
      if (fetching) return
      setFetching(true)

      // Calculate the number of items to fetch.
      const limit = Math.max(BATCH_SIZE, stopIndex - startIndex + 1)
      const [startPage, stopPage] = indicesToPage(startIndex, stopIndex, limit)

      // Fetch new notifications with a maximum of 1000 items.
      await dispatch(
        attachPromise(
          getBookmarksList({
            limit: Math.min((stopPage - startPage + 1) * BATCH_SIZE, 1000),
            page: startPage,
          }),
        ),
      )

      setFetching(false)
    },
    [fetching, dispatch],
  )

  return (
    <Panel
      title={m.title}
      icon={IconStar}
      toggleOptions={options}
      activeToggle={active}
      onToggleClick={handleChange}
      className={classNames(styles.topEntitiesPanel, {
        [styles.hasEntities]: hasEntities,
      })}
    >
      {active === 'all' && <AllTopEntitiesList loadNextPage={loadNextPage} />}
      {active === 'applications' && <TopApplicationsList loadNextPage={loadNextPage} />}
      {active === 'gateways' && <TopGatewaysList loadNextPage={loadNextPage} />}
      {active === 'end-devices' && <TopDevicesList loadNextPage={loadNextPage} />}
    </Panel>
  )
}

export default TopEntitiesDashboardPanel
