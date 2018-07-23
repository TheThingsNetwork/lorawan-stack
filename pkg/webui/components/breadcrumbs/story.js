// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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
import 'focus-visible/dist/focus-visible'
import { withInfo } from '@storybook/addon-info'

import Breadcrumbs from '.'

storiesOf('Breacrumbs', module)
  .addDecorator((story, context) => withInfo({
    inline: true,
    header: false,
    source: false,
    propTables: [ Breadcrumbs ],
  })(story)(context))
  .add('Default', function () {
    const entries = [
      {
        title: 'Application',
        path: '/applications',
      },
      {
        title: 'test-app',
        path: '/test-app',
      },
      {
        title: 'Collaborators',
        path: '/collaborators',
      },
    ]

    return (
      <Breadcrumbs entries={entries} />
    )
  }).add('With icons', function () {
    const entries = [
      {
        title: 'Gateways',
        path: '/gateways',
        icon: 'gateway',
      },
      {
        title: 'eui-0000024b0806021d',
        path: '/eui-0000024b0806021d',
        icon: 'devices',
      },
      {
        title: 'Traffic',
        path: '/traffic',
        icon: 'data',
      },
    ]

    return (
      <Breadcrumbs entries={entries} />
    )
  })