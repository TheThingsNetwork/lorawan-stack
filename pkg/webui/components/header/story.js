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
import { action } from '@storybook/addon-actions'

import Dropdown from '../dropdown'
import NavigationBar from '../navigation/bar'
import Logo from '../logo'
import TtsLogo from '../../assets/static/logo.svg'
import ExampleLogo from '../logo/story-logo.svg'
import Header from '.'

const user = {
  name: 'kschiffer',
  ids: {
    user_id: 'ksc300',
  },
}

const singleLogo = <Logo logo={{ src: TtsLogo, alt: 'Logo' }} />
const doubleLogo = (
  <Logo
    logo={{ src: TtsLogo, alt: 'Logo' }}
    secondaryLogo={{ src: ExampleLogo, alt: 'Secondary Logo' }}
  />
)

const navigationEntries = (
  <React.Fragment>
    <NavigationBar.Item title="Overview" icon="overview" path="/overview" />
    <NavigationBar.Item title="Applications" icon="application" path="/application" />
    <NavigationBar.Item title="Gateways" icon="gateway" path="/gateways" />
    <NavigationBar.Item title="Organizations" icon="organization" path="/organization" />
  </React.Fragment>
)

const items = (
  <React.Fragment>
    <Dropdown.Item title="Profile Settings" icon="settings" path="/profile-settings" />
    <Dropdown.Item title="Logout" icon="power_settings_new" path="/logout" />
  </React.Fragment>
)

storiesOf('Header', module)
  .add('Single Logo', () => (
    <Header
      dropdownItems={items}
      handleSearchRequest={action('Search')}
      navigationEntries={navigationEntries}
      style={{ margin: '-1rem' }}
      user={user}
      logo={singleLogo}
    />
  ))
  .add('Double Logo', () => (
    <Header
      dropdownItems={items}
      handleSearchRequest={action('Search')}
      navigationEntries={navigationEntries}
      style={{ margin: '-1rem' }}
      user={user}
      logo={doubleLogo}
    />
  ))
