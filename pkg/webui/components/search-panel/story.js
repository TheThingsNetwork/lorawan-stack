// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

import SearchPanel from '.'

export default {
  title: 'Search Panel',
  component: SearchPanel,
}

const exampleItems = [
  {
    category: 'applications',
    items: [
      {
        name: 'My Application',
        id: 'my-application-id',
      },
      {
        name: 'My Second Application',
        id: 'my-second-application-id',
      },
    ],
  },
  {
    category: 'gateways',
    items: [
      {
        name: 'My Gateway',
        id: 'my-gateway-id',
      },
      {
        name: 'My Second Gateway',
        id: 'my-second-gateway-id',
      },
    ],
  },
  {
    category: 'organizations',
    items: [
      {
        name: 'My Organization',
        id: 'my-organization-id',
      },
      {
        name: 'My Second Organization',
        id: 'my-second-organization-id',
      },
    ],
  },
  {
    category: 'bookmarks',
    items: [
      {
        name: 'My Bookmark',
        id: 'my-bookmark-id',
      },
      {
        name: 'My Second Bookmark',
        id: 'my-second-bookmark-id',
      },
    ],
  },
]

export const Default = () => <SearchPanel items={exampleItems} />
