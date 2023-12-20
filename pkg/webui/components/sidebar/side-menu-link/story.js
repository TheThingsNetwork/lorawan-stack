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

import SidebarContext from '@ttn-lw/containers/side-bar/context'

import Link from '.'

export default {
  title: 'Side Menu Link',
  decorators: [
    storyFn => (
      <SidebarContext.Provider value={{ isMinimized: false }}>{storyFn()}</SidebarContext.Provider>
    ),
  ],
  parameters: {
    design: {
      type: 'figma',
      url: 'https://www.figma.com/file/7pBLWK4tsjoAbyJq2viMAQ/2023-console-redesign?type=design&node-id=1293%3A7435&mode=design&t=Hbk2Qngeg1xqg4V3-1',
    },
  },
}

export const Default = () => (
  <div style={{ maxWidth: '20rem' }}>
    <Link title="Example name" icon="group" path="example-path" exact />
  </div>
)
