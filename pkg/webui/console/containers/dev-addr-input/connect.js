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

import api from '../../api'

export default function (Component) {

  @bind
  class Connect extends React.PureComponent {
    state = {
      devAddr: '',
      error: undefined,
      fetching: false,
    }

    async generateDevAddr () {
      await this.setState({ error: undefined, fetching: true })

      try {
        const { dev_addr } = await api.ns.generateDevAddress()

        await this.setState({
          error: undefined,
          fetching: false,
          devAddr: dev_addr,
        })
      } catch (error) {
        await this.setState({
          error,
          fetching: false,
          devAddr: '',
        })
      }
    }

    render () {
      const { devAddr, fetching, error } = this.state

      return (
        <Component
          {...this.props}
          fetching={fetching}
          error={error}
          generatedDevAddr={devAddr}
          onDevAddrGenerate={this.generateDevAddr}
        />
      )
    }
  }

  return Connect
}
