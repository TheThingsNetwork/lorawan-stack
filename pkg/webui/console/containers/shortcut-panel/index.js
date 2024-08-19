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
import { useDispatch } from 'react-redux'

import { APPLICATION } from '@console/constants/entities'

import {
  IconUsersGroup,
  IconKey,
  IconBolt,
  IconApplication,
  IconDevice,
  IconGateway,
} from '@ttn-lw/components/icon'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import { setSearchOpen, setSearchScope } from '@console/store/actions/search'

import Panel from '../../../components/panel'

import ShortcutItem from './shortcut-item'

const m = defineMessages({
  shortcuts: 'Quick actions',
  addApplication: 'New application',
  addGateway: 'New gateway',
  addNewOrganization: 'New organization',
  addPersonalApiKey: 'New personal API key',
})

const ShortcutPanel = () => {
  const dispatch = useDispatch()
  const handleRegisterDeviceClick = React.useCallback(() => {
    dispatch(setSearchScope(APPLICATION))
    dispatch(setSearchOpen(true))
  }, [dispatch])

  return (
    <Panel title={m.shortcuts} icon={IconBolt} divider className="h-full">
      <div className="grid gap-cs-xs">
        <ShortcutItem
          icon={IconApplication}
          title={m.addApplication}
          link="/applications/add"
          className="item-6"
        />
        <ShortcutItem
          icon={IconGateway}
          title={m.addGateway}
          link="/gateways/add"
          className="item-6"
        />
        <ShortcutItem
          icon={IconUsersGroup}
          title={m.addNewOrganization}
          link="/organizations/add"
          className="item-4"
        />
        <ShortcutItem
          icon={IconKey}
          title={m.addPersonalApiKey}
          link="/user/api-keys/add"
          className="item-4"
        />
        <ShortcutItem
          icon={IconDevice}
          title={sharedMessages.registerDeviceInApplication}
          action={handleRegisterDeviceClick}
          className="item-4"
        />
      </div>
    </Panel>
  )
}

export default ShortcutPanel
