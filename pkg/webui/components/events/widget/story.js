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
import bind from 'autobind-decorator'
import { storiesOf } from '@storybook/react'

import Events from '..'

const events = [
  {
    icon: 'add_circle',
    identifiers: [{ application_ids: { application_id: 'admin-app' }}],
    name: 'application.create',
    time: '2019-03-28T13:18:13.376022Z',
  },
  {
    icon: 'remove_circle',
    identifiers: [{ application_ids: { application_id: 'admin-app' }}],
    name: 'application.delete',
    time: '2019-03-28T13:18:24.376022Z',
  },
  {
    icon: 'edit',
    identifiers: [{ application_ids: { application_id: 'admin-app' }}],
    name: 'application.update',
    time: '2019-03-28T13:18:39.376022Z',
  },
  {
    icon: 'remove_circle',
    identifiers: [{ application_ids: { application_id: 'admin-app' }}],
    name: 'application.delete',
    time: '2019-03-28T13:18:48.376022Z',
  },
]

@bind
class WidgetExample extends React.Component {
  state = {
    events: [],
  }

  onEventPublish () {
    this.setState(prev => ({
      events: [
        {
          name: 'application.update',
          time: new Date().toISOString(),
          identifiers: [{ application_ids: { application_id: 'admin-app' }}],
        },
        ...prev.events,
      ],
    }))
  }

  render () {
    const { events } = this.state

    return (
      <div>
        <Events.Widget
          events={events}
          connectionStatus="good"
          emitterId="admin-app"
          toAllUrl="/"
        />
        <button style={{ marginTop: '20px' }} onClick={this.onEventPublish}>Publish event</button>
      </div>
    )
  }
}

storiesOf('Events/Widget', module)
  .add('Empty', () => (
    <Events.Widget
      events={[]}
      connectionStatus="good"
      toAllUrl="/"
      emitterId="admin-app"
    />
  ))
  .add('Default', () => (
    <Events.Widget
      events={events}
      connectionStatus="good"
      toAllUrl="/"
      emitterId="admin-app"
    />
  ))
  .add('Publish', () => (
    <WidgetExample />
  ))
