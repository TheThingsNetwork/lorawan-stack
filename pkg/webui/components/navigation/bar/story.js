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
import 'focus-visible/dist/focus-visible'
import { withInfo } from '@storybook/addon-info'

import NavigationBar, { NavigationBarItem } from '../bar'

storiesOf('Navigation', module)
  .addDecorator((story, context) =>
    withInfo({
      inline: true,
      header: false,
      source: false,
      propTables: [NavigationBar],
    })(story)(context),
  )
  .add('NavigationBar', () => (
    <NavigationBar>
      <NavigationBarItem title="Overview" icon="overview" path="/overview" />
      <NavigationBarItem title="Applications" icon="Application" path="/application" />
      <NavigationBarItem title="Gateways" icon="gateway" path="/gateways" />
      <NavigationBarItem title="Organizations" icon="organization" path="/organization" />
    </NavigationBar>
  ))
