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

import SideNavigation from '../side/side'

storiesOf('Navigation', module)
  .addDecorator((story, context) =>
    withInfo({
      inline: true,
      header: false,
      source: false,
      propTables: [SideNavigation],
    })(story)(context),
  )
  .add('SideNavigation', function() {
    const header = {
      title: 'test-application',
      icon: 'application',
    }

    const entries = [
      {
        title: 'Overview',
        icon: 'overview',
        path: '/overview',
      },
      {
        title: 'Devices',
        icon: 'devices',
        path: '/devices',
      },
      {
        title: 'Data',
        icon: 'data',
        path: '/data',
      },
      {
        title: 'Payload Formats',
        icon: 'code',
        path: '/payloadformats',
      },
      {
        title: 'Integrations',
        icon: 'link',
        path: '/integrations',
      },
      {
        title: 'API Keys',
        icon: 'lock',
        path: '/api-keys',
      },
      {
        title: 'General Settings',
        icon: 'settings',
        path: '/settings',
      },
    ]

    return (
      <div style={{ width: '300px', height: '700px' }}>
        <SideNavigation entries={entries} header={header} />
      </div>
    )
  })
  .add('SideNavigation (Nested)', function() {
    const header = {
      title: 'test-application',
      icon: 'application',
    }

    const entries = [
      {
        title: 'Overview',
        icon: 'overview',
        path: '/overview',
      },
      {
        title: 'Devices',
        icon: 'devices',
        path: '/devices',
      },
      {
        title: 'Data',
        icon: 'data',
        nested: true,
        items: [{ title: 'List', path: '/data/list' }, { title: 'Map', path: '/data/map' }],
      },
      {
        title: 'Payload Formats',
        icon: 'code',
        path: '/payloadformats',
      },
      {
        title: 'Integrations',
        icon: 'link',
        nested: true,
        items: [
          { title: 'Something', path: '/integrations/something' },
          { title: 'Somethingv2', path: '/integrations/somethingv2' },
        ],
      },
      {
        title: 'API Keys',
        icon: 'lock',
        path: '/api-keys',
      },
      {
        title: 'General Settings',
        icon: 'settings',
        path: '/settings',
      },
    ]

    return (
      <div style={{ width: '300px', height: '700px' }}>
        <SideNavigation entries={entries} header={header} />
      </div>
    )
  })
