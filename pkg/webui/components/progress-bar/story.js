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
import { storiesOf } from '@storybook/react'

import ProgressBar from '.'

export default class Helper extends React.Component {
  state = {
    percentage: 0,
  }

  increment() {
    const { percentage } = this.state
    const newVal = percentage > 100 ? 0 : percentage + Math.random() * 15

    this.setState({ percentage: newVal })
  }

  componentDidMount() {
    setInterval(this.increment.bind(this), 1000)
  }

  render() {
    const { percentage } = this.state

    return <ProgressBar percentage={percentage} showStatus />
  }
}

storiesOf('ProgressBar', module)
  .add('Default', () => <ProgressBar percentage={50} />)
  .add('With Status', () => <ProgressBar current={24} target={190} showStatus />)
  .add('With ETA Estimation', () => <Helper />)
