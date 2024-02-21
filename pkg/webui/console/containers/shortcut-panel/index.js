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

import Panel from '../../../components/panel'

import ShortcutItem from './shortcut-item'

const m = defineMessages({
  shortcuts: 'Shortcuts',
  edit: 'Edit shortcuts',
  addApplication: 'Add a new Application',
  addGateway: 'Add a new Gateway',
  addOrganization: 'Add a new Organization',
  addPersonalApiKey: 'Add a new personal API key',
  registerDevice: 'Register a device',
})

const ShortcutPanel = () => (
  <Panel
    title={m.shortcuts}
    path="/edit-shortcuts"
    icon="bolt"
    buttonTitle={m.edit}
    filledIcon
    divider
  >
    <div className="grid">
      <ShortcutItem
        icon="display_settings"
        title={m.addApplication}
        link="/applications/add"
        className="item-6"
      />
      <ShortcutItem icon="router" title={m.addGateway} link="/gateways/add" className="item-6" />
      <ShortcutItem
        icon="group"
        title={m.addOrganization}
        link="/organizations/add"
        className="item-4"
      />
      <ShortcutItem
        icon="key"
        title={m.addPersonalApiKey}
        link="/user/api-keys/add"
        className="item-4"
      />
      <ShortcutItem
        icon="settings_remote"
        title={m.registerDevice}
        link="/applications"
        className="item-4"
      />
    </div>
  </Panel>
)

export default ShortcutPanel