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

import TtsLogo from '@assets/static/tts-logo.svg'

import Dropdown from '@ttn-lw/components/dropdown'
import ExampleLogo from '@ttn-lw/components/logo/story-logo-new.svg'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'

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

const plusDropdownItems = (
  <>
    <Dropdown.Item title="Add new application" icon="display_settings" path="/applications/add" />
    <Dropdown.Item title="Add new gateway" icon="router" path="/gateways/add" />
    <Dropdown.Item title="Add new organization" icon="support" path="/organizations/add" />
    <Dropdown.Item
      title="Register end device in application"
      icon="settings_remote"
      path="/devices/add"
    />
  </>
)

const starDropdownItems = (
  <>
    <Dropdown.Item title="Parking Lot Occupancy" icon="display_settings" path="/parking1" />
    <Dropdown.Item title="Parking Lot Occupancy" icon="display_settings" path="/parking2" />
    <Dropdown.Item title="Kerlink iZeptoCell-C 5.7.2" icon="router" path="/kerlink" />
    <Dropdown.Item title="Dragino LPS8 lgw-5.4.1689641188" icon="router" path="/dragino" />
    <Dropdown.Item title="Generic Node - AU915" icon="settings_remote" path="/au915" />
    <Dropdown.Item title="Generic Node - US915" icon="settings_remote" path="/us915" />
  </>
)

const profileDropdownItems = (
  <>
    <Dropdown.Item title="Profile settings" icon="person" path="/profile-settings" />
    <Dropdown.Item title="Manage cloud subscription" icon="credit_card" path="/manage-cloud-subs" />
    <Dropdown.Item title="Network Operations Center" icon="bar_chart" path="/network_ops" />
    <Dropdown.Item title="Admin panel" icon="admin_panel_settings" path="/admin-panel" />
    <hr />
    <Dropdown.Item title="Upgrade" icon="stars" path="/upgrade" />
    <Dropdown.Item title="Get support" icon="support" path="/support" />
    <Dropdown.Item title="Documentation" icon="menu_book" path="/documentation" />
    <hr />
    <Dropdown.Item title="Logout" icon="logout" path="/logout" />
  </>
)

export const Default = () => {
  const breadcrumbs = [
    <Breadcrumb key="1" path="/applications" content="Applications" />,
    <Breadcrumb key="2" path="/applications/test-app" content="test-app" />,
    <Breadcrumb key="3" path="/applications/test-app/devices" content="Devices" />,
  ]

  return (
    <div style={{ height: '25rem' }}>
      <Header
        user={user}
        breadcrumbs={breadcrumbs}
        profileDropdownItems={profileDropdownItems}
        addDropdownItems={plusDropdownItems}
        starDropdownItems={starDropdownItems}
        brandLogo={{ src: ExampleLogo, alt: 'Secondary Logo' }}
        logo={{ src: TtsLogo, alt: 'Logo' }}
      />
    </div>
  )
}
