// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

import WidgetContainer from '.'

storiesOf('WidgetContainer', module).add('Default', () => (
  <div style={{ width: '500px' }}>
    <WidgetContainer title="Location" toAllUrl="#" linkMessage="Change location">
      <div
        style={{
          height: '300px',
          border: '1px solid gray',
          backgroundColor: '#eee',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          color: 'gray',
        }}
      >
        Map placeholder as example
      </div>
    </WidgetContainer>
  </div>
))
