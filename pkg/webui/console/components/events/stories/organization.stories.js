// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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
import { action } from '@storybook/addon-actions'

import Events from '..'

import { organizationEvents } from './event-data'

export default {
  title: 'Events/Organization',
}

export const Default = () => (
  <div style={{ display: 'flex', height: '100vh', width: '100%' }}>
    <Events
      entityId="test-app2"
      events={organizationEvents}
      onPause={action('Organization onPause')}
      onClear={action('Organization onClear')}
      paused={false}
      truncated={false}
      scoped
    />
  </div>
)

export const Widget = () => (
  <div style={{ width: '540px' }}>
    <Events.Widget events={organizationEvents} entityId="test-org-id" toAllUrl="/" scoped />
  </div>
)
