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
      <Dropdown.Item title="Profile settings" icon="user" path="/profile-settings" />
      <Dropdown.Item
        title="Manage cloud subscription"
        icon="credit-card"
        path="/manage-cloud-subs"
      />
      <Dropdown.Item title="Network Operations Center" icon="chart-bar" path="/network_ops" />
      <Dropdown.Item title="Admin panel" icon="admin-shield" path="/admin-panel" />
      <hr />
      <Dropdown.Item title="Upgrade" icon="rocket" path="/upgrade" />
      <Dropdown.Item title="Get support" icon="support" path="/support" />
      <Dropdown.Item title="Documentation" icon="book" path="/documentation" />
      <hr />
      <Dropdown.Item title="Logout" icon="logout" action={handleLogout} />
    </ProfileDropdown>
  </div>
)
