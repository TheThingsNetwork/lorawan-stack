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

import Breadcrumb from './breadcrumb'
import { Breadcrumbs } from './breadcrumbs'

storiesOf('Breadcrumbs', module)
  .addDecorator((story, context) =>
    withInfo({
      inline: true,
      header: false,
      source: false,
      propTables: [Breadcrumbs, Breadcrumb],
    })(story)(context),
  )
  .add('Default', function() {
    const breadcrumbs = [
      <Breadcrumb key="1" path="/applications" content="Applications" />,
      <Breadcrumb key="2" path="/applications/test-app" content="test-app" />,
      <Breadcrumb key="3" path="/applications/test-app/devices" content="Devices" />,
    ]

    return <Breadcrumbs breadcrumbs={breadcrumbs} />
  })
  .add('With icons', function() {
    const breadcrumbs = [
      <Breadcrumb key="1" path="/gateways" icon="gateway" content="Gateways" />,
      <Breadcrumb
        key="2"
        path="/gateways/eui-0000024b0806021d"
        icon="devices"
        content="eui-0000024b0806021d"
      />,
      <Breadcrumb
        key="3"
        path="/gateways/eui-0000024b0806021d/traffic"
        icon="data"
        content="Traffic"
      />,
    ]
    return <Breadcrumbs breadcrumbs={breadcrumbs} />
  })
  .add('Mixed', function() {
    const breadcrumbs = [
      <Breadcrumb key="1" path="/applications" content="Applications" />,
      <Breadcrumb key="2" path="/applications/test-app" icon="application" content="test-app" />,
      <Breadcrumb key="3" path="/applications/test-app/traffix" content="Traffic" />,
    ]
    return <Breadcrumbs breadcrumbs={breadcrumbs} />
  })
