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
import { action } from '@storybook/addon-actions'

import Events from '..'

import { events, gatewayEvents, organizationEvents, deviceEvents } from './event-data'

storiesOf('Events/Application', module).add('Default', () => (
  <div style={{ display: 'flex', height: '100vh', width: '100%' }}>
    <Events
      entityId="test-app2"
      events={events}
      onClear={action('Application onClear')}
      onPause={action('Application onPause')}
    />
  </div>
))
storiesOf('Events/Application', module).add('Widget', () => (
  <div style={{ width: '540px' }}>
    <Events.Widget events={events} entityId="test-app2" toAllUrl="/" />
  </div>
))
storiesOf('Events/Gateway', module).add('Default', () => (
  <div style={{ display: 'flex', height: '100vh', width: '100%' }}>
    <Events
      entityId="test-app2"
      events={gatewayEvents}
      onPause={action('Gateway onPause')}
      onClear={action('Gateway onClear')}
      scoped
    />
  </div>
))
storiesOf('Events/Gateway', module).add('Widget', () => (
  <div style={{ width: '540px' }}>
    <Events.Widget events={gatewayEvents} entityId="test-gtw-id" toAllUrl="/" scoped />
  </div>
))
storiesOf('Events/Organization', module).add('Default', () => (
  <div style={{ display: 'flex', height: '100vh', width: '100%' }}>
    <Events
      entityId="test-app2"
      events={organizationEvents}
      onPause={action('Organization onPause')}
      onClear={action('Organization onClear')}
      scoped
    />
  </div>
))
storiesOf('Events/Organization', module).add('Widget', () => (
  <div style={{ width: '540px' }}>
    <Events.Widget events={organizationEvents} entityId="test-org-id" toAllUrl="/" scoped />
  </div>
))
storiesOf('Events/Device', module).add('Default', () => (
  <div style={{ display: 'flex', height: '100vh', width: '100%' }}>
    <Events
      entityId="test-app2"
      events={deviceEvents}
      onClear={action('Device onClear')}
      onPause={action('Device onPause')}
      scoped
    />
  </div>
))
storiesOf('Events/Device', module).add('Widget', () => (
  <div style={{ width: '540px' }}>
    <Events.Widget events={deviceEvents} entityId="test-dev-c" toAllUrl="/" scoped />
  </div>
))
