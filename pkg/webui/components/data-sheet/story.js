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

import DataSheet from '.'

const testData = [
  {
    header: 'Hardware',
    items: [
      { key: 'Brand', value: 'SemTech' },
      { key: 'Model', value: 'Some Model' },
      { key: 'Hardware Version', value: 'v1.2.5' },
      { key: 'Firmware Version', value: 'v1.1.1' },
    ],
  },
  {
    header: 'Activation Info',
    items: [
      { key: 'Device EUI', value: '1212121212', type: 'byte', sensitive: false },
      {
        key: 'Device EUI with Uint32_t',
        value: '1212121212',
        type: 'byte',
        sensitive: false,
        enableUint32: true,
      },
      { key: 'Join EUI', value: '1212121212', type: 'byte', sensitive: false },
      {
        key: 'Value with Nesting',
        value: 'ae9d78fe9aed8fe',
        type: 'code',
        sensitive: false,
        subItems: [
          { key: 'Application Key', value: 'ae9d78fe9aed8fe', type: 'code', sensitive: true },
          { key: 'Network Key', value: 'ae9d78fe9aed8fe', type: 'code', sensitive: true },
        ],
      },
    ],
  },
]

const containerStyles = {
  maxWidth: '600px',
}

export default {
  title: 'Data Sheet',
  component: DataSheet,
}

export const Default = () => (
  <div style={containerStyles}>
    <DataSheet data={testData} />
  </div>
)
