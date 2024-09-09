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
import classNames from 'classnames'

import { ALL, APPLICATION, END_DEVICE, GATEWAY } from '@console/constants/entities'

import { IconStar } from '@ttn-lw/components/icon'
import Panel from '@ttn-lw/components/panel'

import RequireRequest from '@ttn-lw/lib/components/require-request'
import Message from '@ttn-lw/lib/components/message'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import { getTopEntities } from '@console/store/actions/top-entities'

import AllTopEntitiesList from './all-top-entities'
import TopApplicationsList from './top-applications'
import TopGatewaysList from './top-gateways'
import TopDevicesList from './top-devices'

import styles from './top-entities-panel.styl'

const PanelError = () => (
  <Message
    className="text-center mt-cs-xxl c-text-neutral-light"
    content={sharedMessages.topEntitiesError}
  />
)

const TopEntitiesDashboardPanel = () => {
  const [active, setActive] = useState(ALL)

  const handleChange = useCallback(
    (_, value) => {
      setActive(value)
    },
    [setActive],
  )

  const options = [
    { label: sharedMessages.all, value: ALL },
    { label: sharedMessages.applications, value: APPLICATION },
    { label: sharedMessages.gateways, value: GATEWAY },
    { label: sharedMessages.devices, value: END_DEVICE },
  ]

  return (
    <Panel
      title={sharedMessages.topEntities}
      icon={IconStar}
      toggleOptions={options}
      activeToggle={active}
      onToggleClick={handleChange}
      className={classNames(styles.topEntitiesPanel)}
    >
      <RequireRequest requestAction={getTopEntities()} errorRenderFunction={PanelError}>
        {active === ALL && <AllTopEntitiesList />}
        {active === APPLICATION && <TopApplicationsList />}
        {active === GATEWAY && <TopGatewaysList />}
        {active === END_DEVICE && <TopDevicesList />}
      </RequireRequest>
    </Panel>
  )
}

export default TopEntitiesDashboardPanel
