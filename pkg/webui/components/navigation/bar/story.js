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

import {
  IconApplication,
  IconGateway,
  IconOrganization,
  IconOverview,
} from '@ttn-lw/components/icon'
import 'focus-visible/dist/focus-visible'

import NavigationBar from '.'

export default {
  title: 'Navigation',
  component: NavigationBar,
}

export const _NavigationBar = () => (
  <NavigationBar>
    <NavigationBar.Item title="Overview" icon={IconOverview} path="/overview" />
    <NavigationBar.Item title="Applications" icon={IconApplication} path="/application" />
    <NavigationBar.Item title="Gateways" icon={IconGateway} path="/gateways" />
    <NavigationBar.Item title="Organizations" icon={IconOrganization} path="/organization" />
  </NavigationBar>
)

_NavigationBar.story = {
  name: 'NavigationBar',
}
