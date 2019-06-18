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

import React, { Component } from 'react'
import bind from 'autobind-decorator'
import { storiesOf } from '@storybook/react'
import { withInfo } from '@storybook/addon-info'

import Tabs from '.'

@bind
class Example extends Component {
  constructor (props) {
    super(props)

    this.state = {
      activeTab: props.active,
    }
  }

  onTabChange (activeTab) {
    this.setState({ activeTab })
  }

  render () {
    const { activeTab } = this.state

    return (
      <Tabs
        {...this.props}
        active={activeTab}
        onTabChange={this.onTabChange}
      />
    )
  }
}

storiesOf('Tabs', module)
  .addDecorator((story, context) => withInfo({
    inline: true,
    header: false,
    source: true,
    propTables: [ Tabs ],
    propTablesExclude: [ Example ],
  })(story)(context))
  .add('Default', function () {
    const tabs = [
      { title: 'All', name: 'all' },
      { title: 'Starred', name: 'starred' },
    ]

    return (
      <Example tabs={tabs} active={tabs[0].name} />
    )
  }).add('Default (disabled)', function () {

    const tabs = [
      { title: 'All', name: 'all' },
      { title: 'Starred', name: 'starred', disabled: true },
    ]

    return (
      <Example tabs={tabs} active={tabs[0].name} />
    )
  }).add('With icons', function () {
    const tabs = [
      { title: 'People', name: 'people', icon: 'organization' },
      { title: 'Data', name: 'data', icon: 'data' },
    ]

    return (
      <Example tabs={tabs} active={tabs[0].name} />
    )
  }).add('With icons (disabled)', function () {
    const tabs = [
      { title: 'People', name: 'people', icon: 'organization' },
      { title: 'Data', name: 'data', icon: 'data', disabled: true },
    ]

    return (
      <Example tabs={tabs} active={tabs[0].name} />
    )
  })
