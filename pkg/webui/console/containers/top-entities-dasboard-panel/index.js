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

import Panel from '@ttn-lw/components/panel'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import style from './top-entities-dashboard-panel.styl'

const m = defineMessages({
  topEntities: 'Your top entities',
})

const TopEntitiesDashboardPanel = () => {
  const options = [
    { label: sharedMessages.all, value: 'all' },
    { label: sharedMessages.applications, value: 'apps' },
    { label: sharedMessages.gateways, value: 'gtws' },
    { label: sharedMessages.devices, value: 'devices' },
  ]

  const [active, setActive] = React.useState('apps')

  const handleChange = React.useCallback(
    (_, value) => {
      setActive(value)
    },
    [setActive],
  )

  return (
    <Panel
      title={m.topEntities}
      icon="star"
      toggleOptions={options}
      activeToggle={active}
      onToggleClick={handleChange}
      className={style.topEntitiesPanel}
    >
      {active === 'all' && <div style={{ height: '22rem' }}>all top entities</div>}
      {active === 'apps' && <div style={{ height: '22rem' }}>top applications</div>}
      {active === 'gtws' && <div style={{ height: '22rem' }}>top gateways</div>}
      {active === 'devices' && <div style={{ height: '22rem' }}>top devices</div>}
    </Panel>
  )
}

export default TopEntitiesDashboardPanel
