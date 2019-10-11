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

import Status from '.'

const containerStyle = {
  width: '100px',
  display: 'flex',
  alignItems: 'center',
  justifyContent: 'space-between',
}

@bind
class Toggle extends React.Component {
  state = {
    status: 'unknown',
  }

  toggleStatus() {
    const { status } = this.state
    let nextStatus
    switch (status) {
      case 'unknown':
        nextStatus = 'bad'
        break
      case 'bad':
        nextStatus = 'mediocre'
        break
      case 'mediocre':
        nextStatus = 'good'
        break
      case 'good':
        nextStatus = 'unknown'
        break
    }

    this.setState({ status: nextStatus })
  }

  render() {
    const { status } = this.state
    return (
      <div>
        <Status status={status} />
        <br />
        <button onClick={this.toggleStatus}>Toggle</button>
      </div>
    )
  }
}

storiesOf('Status', module)
  .add('All types', () => (
    <div>
      <div style={containerStyle}>
        <span>Good:</span>
        <Status status="good" />
      </div>
      <div style={containerStyle}>
        <span>Bad:</span>
        <Status status="bad" />
      </div>
      <div style={containerStyle}>
        <span>Mediocre:</span>
        <Status status="mediocre" />
      </div>
      <div style={containerStyle}>
        <span>Unknown:</span>
        <Status status="unknown" />
      </div>
    </div>
  ))
  .add('With label', () => (
    <div>
      <Status label="Network Status" status="good" />
      <Status label="Network Status" status="bad" />
      <Status label="Network Status" status="mediocre" />
      <Status label="Network Status" status="unknown" />
    </div>
  ))
  .add('Without Pulse', () => (
    <div>
      <Status label="No Pulse" status="good" pulse={false} />
      <Status label="No Pulse" status="bad" pulse={false} />
      <Status label="No Pulse" status="mediocre" pulse={false} />
      <Status label="No Pulse" status="unknown" pulse={false} />
    </div>
  ))
  .add('Toggle', () => <Toggle />)
