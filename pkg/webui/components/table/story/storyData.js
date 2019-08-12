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
import Button from '../../button'

const headers = [
  {
    name: 'appId',
    displayName: 'Application ID',
  },
  {
    name: 'desc',
    displayName: 'Description',
  },
  {
    name: 'devices',
    displayName: 'Devices',
    centered: true,
  },
  {
    name: 'lastActivity',
    displayName: 'Last Activity',
  },
]

const rows = [
  {
    appId: 'my-app1',
    desc: 'Test Application',
    devices: '1',
    lastActivity: '10 sec. ago',
    tabs: ['all', 'starred'],
    clickable: false,
  },
  {
    appId: 'my-app2',
    desc: 'Test Application 2',
    devices: '2',
    lastActivity: '10 sec. ago',
    tabs: ['all'],
    clickable: false,
  },
  {
    appId: 'my-app3',
    desc: 'Test Application 3',
    devices: '3',
    lastActivity: '10 sec. ago',
    tabs: ['all', 'starred'],
    clickable: false,
  },
  {
    appId: 'my-app4',
    desc: 'Test Application 4',
    devices: '5',
    lastActivity: '10 sec. ago',
    tabs: ['all', 'starred'],
    clickable: false,
  },
  {
    appId: 'my-app5',
    desc: 'Test Application 5',
    devices: '4',
    lastActivity: '10 sec. ago',
    tabs: ['all', 'starred'],
    clickable: false,
  },
  {
    appId: 'my-app6',
    desc: 'Test Application 6',
    devices: '3',
    lastActivity: '10 sec. ago',
    tabs: ['all'],
    clickable: false,
  },
]

export default {
  defaultExample: {
    headers,
    rows,
  },
  loadingExample: {
    headers,
    rows,
  },
  paginatedExample: {
    headers,
    rows: rows
      .concat(rows)
      .concat(rows)
      .concat(rows)
      .concat(rows),
  },
  clickableRowsExample: {
    headers,
    rows: rows.slice(0, 3).map(row =>
      Object.assign({}, row, {
        clickable: true,
      }),
    ),
  },
  customCellExample: {
    headers: [
      ...headers,
      {
        name: 'options',
        displayName: 'Options',
        centered: true,
      },
    ],
    rows: rows.map(function(r) {
      return Object.assign({}, r, {
        options: (
          <div>
            <Button icon="settings" />
            <Button danger icon="delete" />
          </div>
        ),
      })
    }),
  },
  sortableExample: {
    headers: headers.map(function(header) {
      if (header.name === 'devices' || header.name === 'appId') {
        return Object.assign({}, header, {
          sortable: true,
        })
      }

      return header
    }),
    rows,
  },
  emptyExample: {
    headers,
    rows: [],
  },
  customWrapperExample: {
    headers,
    rows,
  },
}
