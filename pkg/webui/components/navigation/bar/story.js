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

import NavigationBar from '../bar'

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
      <NavigationBar.Item title="Overview" icon="overview" path="/overview" />
      <NavigationBar.Item title="Applications" icon="application" path="/application" />
      <NavigationBar.Item title="Gateways" icon="gateway" path="/gateways" />
      <NavigationBar.Item title="Organizations" icon="organization" path="/organization" />
    </NavigationBar>
  ))
