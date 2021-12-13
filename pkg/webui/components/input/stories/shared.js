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

/* eslint-disable react/prop-types, import/prefer-default-export */

import React from 'react'
import bind from 'autobind-decorator'

import Input from '..'

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
    const { type, component: Component, ...props } = this.props
    const { value } = this.state

    const InputComponent = Component ? Component : Input

    return <InputComponent {...props} type={type} onChange={this.onChange} value={value} />
  }
}

export { Example }
