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
import { defineMessages } from 'react-intl'
import classNames from 'classnames'
import { useSelector } from 'react-redux'

import { IconDevice } from '@ttn-lw/components/icon'
import Panel from '@ttn-lw/components/panel'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import { selectSelectedApplicationId } from '@console/store/selectors/applications'

import TopDevicesList from '../top-entities-dashboard-panel/top-devices'

import RecentEndDevices from './recent-devices'

import style from './devices-panel.styl'

const m = defineMessages({
  recentDevices: 'Recently active',
})

const DevicesPanel = () => {
  const [active, setActive] = useState('top')
  const appId = useSelector(selectSelectedApplicationId)

  const handleChange = useCallback(
    (_, value) => {
      if (value !== 'all') {
        setActive(value)
      }
    },
    [setActive],
  )

  const options = [
    { label: sharedMessages.topDevices, value: 'top' },
    { label: m.recentDevices, value: 'recent' },
    { label: sharedMessages.all, value: 'all', link: `/applications/${appId}/devices` },
  ]

  return (
    <Panel
      title={sharedMessages.devices}
      icon={IconDevice}
      toggleOptions={options}
      activeToggle={active}
      onToggleClick={handleChange}
      className={classNames(style.devicesPanel)}
    >
      {active === 'top' && <TopDevicesList appId={appId} />}
      {active === 'recent' && <RecentEndDevices />}
    </Panel>
  )
}

export default DevicesPanel
