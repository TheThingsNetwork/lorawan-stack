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

import Dropdown from '@ttn-lw/components/dropdown-v2'
import ExampleLogo from '@ttn-lw/components/logo/story-logo.svg'

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
  <div style={{ height: '6rem', marginLeft: '120px' }}>
    <ProfileDropdown brandLogo={{ src: ExampleLogo, alt: 'Secondary Logo' }}>
      <Dropdown.Item title="Profile Settings" icon="settings" path="/profile-settings" />
      <Dropdown.Item title="Logout" icon="power_settings_new" action={handleLogout} />
    </ProfileDropdown>
  </div>
)
