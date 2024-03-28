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

import {
  IconUser,
  IconLogout,
  IconAdminShield,
  IconCreditCard,
  IconChartBar,
  IconBook,
  IconRocket,
  IconApplication,
  IconDevice,
  IconGateway,
  IconSupport,
} from '@ttn-lw/components/icon'
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
    <Dropdown.Item title="Add new application" icon={IconApplication} path="/applications/add" />
    <Dropdown.Item title="Add new gateway" icon={IconGateway} path="/gateways/add" />
    <Dropdown.Item title="Add new organization" icon={IconSupport} path="/organizations/add" />
    <Dropdown.Item
      title="Register end device in application"
      icon={IconDevice}
      path="/devices/add"
    />
  </>
)

const starDropdownItems = (
  <>
    <Dropdown.Item title="Parking Lot Occupancy" icon={IconApplication} path="/parking1" />
    <Dropdown.Item title="Parking Lot Occupancy" icon={IconApplication} path="/parking2" />
    <Dropdown.Item title="Kerlink iZeptoCell-C 5.7.2" icon={IconGateway} path="/kerlink" />
    <Dropdown.Item title="Dragino LPS8 lgw-5.4.1689641188" icon={IconGateway} path="/dragino" />
    <Dropdown.Item title="Generic Node - AU915" icon={IconDevice} path="/au915" />
    <Dropdown.Item title="Generic Node - US915" icon={IconDevice} path="/us915" />
  </>
)

const profileDropdownItems = (
  <>
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
    <Dropdown.Item title="Logout" icon={IconLogout} path="/logout" />
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
