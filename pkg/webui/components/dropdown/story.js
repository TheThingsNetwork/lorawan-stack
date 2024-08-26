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

import { IconUser, IconLogout, IconAdminShield } from '@ttn-lw/components/icon'

import Dropdown from '.'

export default {
  title: 'Dropdown V2',
  component: Dropdown,
}

export const Default = () => (
  <div style={{ height: '8rem' }}>
    <Dropdown open>
      <Dropdown.HeaderItem title="dropdown items" />
      <Dropdown.Item title="Profile Settings" path="profile/path" icon={IconUser} />
      <Dropdown.Item title="Admin panel" path="admin/path" icon={IconAdminShield} />
      <hr />
      <Dropdown.Item title="Logout" path="logout/path" icon={IconLogout} />
    </Dropdown>
  </div>
)
