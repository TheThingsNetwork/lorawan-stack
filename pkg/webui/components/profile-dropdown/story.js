// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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
import { storiesOf } from '@storybook/react'
import { withInfo } from '@storybook/addon-info'

import ProfileDropdown, { DropdownItem } from '.'

const user = {
  name: 'kschiffer',
}

const handleLogout = function() {
  // eslint-disable-next-line no-console
  console.log('Click')
}

storiesOf('Profile Dropdown', module)
  .addDecorator((story, context) =>
    withInfo({
      inline: true,
      header: false,
      source: false,
      propTables: [ProfileDropdown],
    })(story)(context),
  )
  .add('Default', function() {
    return (
      <ProfileDropdown style={{ marginLeft: '60px' }} user={user}>
        <DropdownItem title="Profile Settings" icon="settings" link="/profile-settings" />
        <DropdownItem title="Logout" icon="power_settings_new" action={handleLogout} />
      </ProfileDropdown>
    )
  })
