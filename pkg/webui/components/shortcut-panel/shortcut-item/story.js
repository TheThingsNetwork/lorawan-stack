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

import ShortcutItem from '.'

export default {
  title: 'Panel/Shortcut Panel/Shortcut Item',
  component: ShortcutItem,
  parameters: {
    design: {
      type: 'figma',
      url: 'https://www.figma.com/file/7pBLWK4tsjoAbyJq2viMAQ/console-redesign?type=design&node-id=1661-5695&mode=design&t=2KlaQGRV9FQm7Nv3-4',
    },
  },
}

export const Default = () => (
  <div style={{ width: '192.5px' }}>
    <ShortcutItem icon="dashboard_customize" title="Add new item" link="/applications" />
  </div>
)
