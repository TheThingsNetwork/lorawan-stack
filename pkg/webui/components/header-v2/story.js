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

import TtsLogo from '@assets/static/logo.svg'

import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import Dropdown from '@ttn-lw/components/dropdown-v2'
import ExampleLogo from '@ttn-lw/components/logo/story-logo.svg'

import Header from '.'

const user = {
  name: 'johndoe',
  ids: {
    user_id: 'jdoe300',
  },
}

export default {
  title: 'Header V2',
}

const dropdownItems = (
  <>
    <Dropdown.Item title="Profile Settings" icon="settings" path="/profile-settings" />
    <Dropdown.Item title="Logout" icon="power_settings_new" path="/logout" />
  </>
)

export const Default = () => {
  const breadcrumbs = [
    <Breadcrumb key="1" path="/applications" content="Applications" />,
    <Breadcrumb key="2" path="/applications/test-app" content="test-app" />,
    <Breadcrumb key="3" path="/applications/test-app/devices" content="Devices" />,
  ]

  return (
    <div style={{ height: '6rem' }}>
      <Header
        user={user}
        breadcrumbs={breadcrumbs}
        profileDropdownItems={dropdownItems}
        addDropdownItems={dropdownItems}
        starDropdownItems={dropdownItems}
        brandLogo={{ src: ExampleLogo, alt: 'Secondary Logo' }}
        logo={{ src: TtsLogo, alt: 'Logo' }}
      />
    </div>
  )
}
