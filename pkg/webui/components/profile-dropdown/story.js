// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

import {
  IconUser,
  IconLogout,
  IconAdminShield,
  IconCreditCard,
  IconChartBar,
  IconBook,
  IconRocket,
  IconSupport,
} from '@ttn-lw/components/icon'
import Dropdown from '@ttn-lw/components/dropdown'
import ExampleLogo from '@ttn-lw/components/logo/story-logo-new.svg'

import ProfileDropdown from '.'

const handleLogout = () => {
  // eslint-disable-next-line no-console
  console.log('Click')
}

export default {
  title: 'Profile Dropdown V2',
  component: ProfileDropdown,
}

export const Default = () => (
  <div style={{ height: '25rem', marginLeft: '15rem' }}>
    <ProfileDropdown brandLogo={{ src: ExampleLogo, alt: 'Secondary Logo' }}>
      <Dropdown.Item title="Profile settings" icon={IconUser} path="/profile-settings" />
      <Dropdown.Item
        title="Manage cloud subscription"
        icon={IconCreditCard}
        path="/manage-cloud-subs"
      />
      <Dropdown.Item title="Network Operations Center" icon={IconChartBar} path="/network_ops" />
      <Dropdown.Item title="Admin panel" icon={IconAdminShield} path="/admin-panel" />
      <hr />
      <Dropdown.Item title="Upgrade" icon={IconRocket} path="/upgrade" />
      <Dropdown.Item title="Get support" icon={IconSupport} path="/support" />
      <Dropdown.Item title="Documentation" icon={IconBook} path="/documentation" />
      <hr />
      <Dropdown.Item title="Logout" icon={IconLogout} action={handleLogout} />
    </ProfileDropdown>
  </div>
)
