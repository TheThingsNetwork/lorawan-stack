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
import { action } from '@storybook/addon-actions'

import Button from '.'

@bind
class Example extends React.Component {
  state = {
    busy: false,
    error: false,
  }

  render() {
    const { busy, disabled, error } = this.state

    return (
      <div>
        <Button
          busy={busy}
          onClick={action('click')}
          message="A Test Button"
          disabled={disabled}
          error={error}
        />
        <br />
        <br />
        <button onClick={this.disable}>{disabled ? 'enable' : 'disable'}</button> &nbsp;
        <button onClick={this.toggle}>work</button> &nbsp;
        <button onClick={this.error}>error shake</button>
      </div>
    )
  }

  toggle() {
    this.setState(state => ({
      busy: !state.busy,
    }))
  }

  disable() {
    this.setState(state => ({
      disabled: !state.disabled,
    }))
  }

  error() {
    this.setState({
      error: true,
    })
    setTimeout(
      function() {
        this.setState({
          error: false,
        })
      }.bind(this),
      1200,
    )
  }
}

storiesOf('Button', module)
  .add('Default', () => (
    <div>
      <Button message="Default" />
      <br />
      <br />
      <Button message="Default" disabled />
      <br />
      <br />
      <Button message="Default" busy />
      <br />
      <br />
      <Button message="Default" large />
      <br />
      <br />
      <Button message="Default" large disabled />
      <br />
      <br />
      <Button message="Default" large busy />
      <br />
      <br />
      <Button.Link message="Router Link" to="/test" />
      <br />
      <br />
      <Button.AnchorLink message="Anchor Link" href="#" />
    </div>
  ))
  .add('Warning', () => (
    <div>
      <Button warning message="Warning" />
      <br />
      <br />
      <Button warning message="Warning" disabled />
      <br />
      <br />
      <Button warning message="Warning" busy />
    </div>
  ))
  .add('Danger', () => (
    <div>
      <Button danger message="Danger" />
      <br />
      <br />
      <Button danger message="Danger" disabled />
      <br />
      <br />
      <Button danger message="Danger" busy />
    </div>
  ))
  .add('Secondary', () => (
    <div>
      <Button secondary message="Secondary" />
      <br />
      <br />
      <Button secondary message="Secondary" disabled />
      <br />
      <br />
      <Button secondary message="Secondary" busy />
    </div>
  ))
  .add('Raw', () => (
    <div>
      <Button raw message="Raw" />
      <br />
      <br />
      <Button raw message="Raw" disabled />
      <br />
      <br />
      <Button raw message="Raw" busy />
    </div>
  ))
  .add('With Icon', () => (
    <div>
      <Button icon="check" message="With Icon" />
      <br />
      <br />
      <Button icon="check" message="With Icon" disabled />
      <br />
      <br />
      <Button icon="check" message="With Icon" busy />
      <br />
      <br />
      <Button icon="check" message="With Icon" large />
      <br />
      <br />
      <Button icon="check" message="With Icon" large disabled />
      <br />
      <br />
      <Button icon="check" message="With Icon" large busy />
    </div>
  ))
  .add('Naked', () => (
    <div>
      <Button naked message="Naked" />
      <br />
      <br />
      <Button naked message="Naked" disabled />
      <br />
      <br />
      <Button naked message="Naked" busy />
      <br />
      <br />
      <Button naked secondary message="Naked Secondary" />
      <br />
      <br />
      <Button naked secondary message="Naked Secondary" disabled />
      <br />
      <br />
      <Button naked secondary message="Naked Secondary" busy />
    </div>
  ))
  .add('Naked With Icon', () => (
    <div>
      <Button naked icon="favorite" message="Naked With Icon" />
      <br />
      <br />
      <Button naked icon="favorite" message="Naked With Icon" disabled />
      <br />
      <br />
      <Button naked icon="favorite" message="Naked With Icon" busy />
    </div>
  ))
  .add('Only Icon', () => (
    <div>
      <Button icon="check" />
      <br />
      <br />
      <Button icon="check" disabled />
      <br />
      <br />
      <Button icon="check" busy />
    </div>
  ))
  .add('Toggle', () => <Example />)
