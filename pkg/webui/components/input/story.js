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

/* eslint-disable react/prop-types */

import crypto from 'crypto'
import React from 'react'
import bind from 'autobind-decorator'
import { storiesOf } from '@storybook/react'

import Input from '.'

class Example extends React.Component {
  constructor(props) {
    super(props)

    this.state = {
      value: props.value || '',
    }
  }

  @bind
  onChange(value) {
    this.setState({ value })
  }

  render() {
    const { type, valid, component: Component, ...props } = this.props
    const { value } = this.state

    const InputComponent = Component ? Component : Input

    return (
      <InputComponent
        {...props}
        type={type}
        onChange={this.onChange}
        valid={valid || (Boolean(value) && value.length > 5)}
        value={value}
      />
    )
  }
}

const generateRandomValue = () => `random-value-${Math.floor(Math.random() * 100)}`
const generateRandom16Bytes = () => crypto.randomBytes(16).toString('hex').toUpperCase()

storiesOf('Input', module)
  .add('Default', () => (
    <div>
      <Example label="Username" />
      <Example label="Username" warning />
      <Example label="Username" error />
    </div>
  ))
  .add('With Placeholder', () => <Example placeholder="Placeholder..." />)
  .add('With icon', () => <Example icon="search" />)
  .add('Valid', () => <Example valid />)
  .add('Disabled', () => <Example value="1234" disabled />)
  .add('Readonly', () => <Example value="1234" readOnly />)
  .add('Password', () => <Example type="password" />)
  .add('Number', () => <Example type="number" />)
  .add('Byte', () => <Example type="byte" min={1} max={5} />)
  .add('Byte read-only', () => <Example type="byte" min={1} max={5} value="A0BF49A464" readOnly />)
  .add('Toggled', () => (
    <Example component={Input.Toggled} type="toggled" enabledMessage="Enabled" />
  ))
  .add('Textarea', () => <Example type="textarea" />)
  .add('With Spinner', () => <Example icon="search" loading />)
  .add('With Action', () => (
    <div>
      <Example action={{ icon: 'build', secondary: true }} />
      <Example action={{ icon: 'build', secondary: true }} warning />
      <Example action={{ icon: 'build', secondary: true }} error />
    </div>
  ))
  .add('Generate', () => (
    <>
      <Example component={Input.Generate} onGenerateValue={generateRandomValue} />
      <Example
        type="byte"
        component={Input.Generate}
        onGenerateValue={generateRandom16Bytes}
        min={16}
        max={16}
      />
    </>
  ))
