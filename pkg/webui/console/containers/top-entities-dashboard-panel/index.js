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
import { useSelector } from 'react-redux'
import { defineMessages } from 'react-intl'
import classNames from 'classnames'

import { IconStar } from '@ttn-lw/components/icon'
import Panel from '@ttn-lw/components/panel'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import { selectBookmarksList } from '@console/store/selectors/user-preferences'

import AllTopEntitiesList from './all-top-entities'
import TopApplicationsList from './top-applications'
import TopGatewaysList from './top-gateways'
import TopDevicesList from './top-devices'

import styles from './top-entities-panel.styl'

const m = defineMessages({
  title: 'Your top entities',
})

const TopEntitiesDashboardPanel = () => {
  const [active, setActive] = useState('all')
  const bookmarks = useSelector(state => selectBookmarksList(state))
  const hasEntities = bookmarks.length > 0

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
      {active === 'all' && <AllTopEntitiesList />}
      {active === 'applications' && <TopApplicationsList />}
      {active === 'gateways' && <TopGatewaysList />}
      {active === 'end-devices' && <TopDevicesList />}
    </Panel>
  )
}

export default TopEntitiesDashboardPanel
