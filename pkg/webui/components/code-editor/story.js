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

import CodeEditor from '.'

const containerStyles = { height: '500px' }

@bind
class Example extends React.Component {
  constructor(props) {
    super(props)

    this.state = {
      value: props.placeholder,
    }
  }

  onChange(value) {
    this.setState({ value })
  }

  render() {
    const { value } = this.state
    return (
      <div style={containerStyles}>
        <CodeEditor {...this.props} onChange={this.onChange} value={value} />
      </div>
    )
  }
}

const code = `
// Decode raw data
function Decoder(bytes, port) {
  var colors = ["red", "green", "blue", "yellow", "cyan", "magenta", "white", "black"];
  var decoded = {
    light: (bytes[0] << 8) | bytes[1],
    temperature: ((bytes[2] << 8) | bytes[3]) / 100,
    state: {
      color: colors[bytes[4]]
    }
  };

  return decoded;
}
`

storiesOf('CodeEditor', module)
  .add('Default', () => (
    <Example language="javascript" name="storybook-code-editor" placeholder={code} />
  ))
  .add('Readonly', () => (
    <Example language="javascript" name="storybook-code-editor" placeholder={code} readOnly />
  ))
  .add('With warning', () => (
    <Example language="javascript" name="storybook-code-editor" placeholder={`${code}}`} />
  ))
