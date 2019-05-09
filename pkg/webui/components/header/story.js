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
import { action } from '@storybook/addon-actions'

import Header from '.'

const user = {
  name: 'kschiffer',
  ids: {
    user_id: 'ksc300',
  },
}

const items = [
  {
    title: 'Profile Settings',
    icon: 'settings',
    link: '/profile-settings',
  },
  {
    title: 'Logout',
    icon: 'power_settings_new',
    link: '/logout',
  },
]

storiesOf('Header', module)
  .addDecorator((story, context) => withInfo({
    inline: true,
    header: false,
    propTables: [ Header ],
  })(story)(context))
  .add('Default', () => (
    <Header
      dropdownItems={items}
      user={user}
      style={{ margin: '-1rem' }}
      handleSearchRequest={action('Search')}
    />
  ))
