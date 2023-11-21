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

import React, { Component } from 'react'
import bind from 'autobind-decorator'

import Tabs from '.'

class Example extends Component {
  constructor(props) {
    super(props)

    this.state = {
      activeTab: props.active,
    }
  }

  @bind
  onTabChange(activeTab) {
    this.setState({ activeTab })
  }

  render() {
    const { activeTab } = this.state

    return <Tabs {...this.props} active={activeTab} onTabChange={this.onTabChange} />
  }
}

export default {
  title: 'Tabs',
  component: Tabs,
}

export const Default = () => {
  const tabs = [
    { title: 'All', name: 'all' },
    { title: 'Starred', name: 'starred' },
  ]

  return <Example tabs={tabs} active={tabs[0].name} />
}

export const DefaultDisabled = () => {
  const tabs = [
    { title: 'All', name: 'all' },
    { title: 'Starred', name: 'starred', disabled: true },
  ]

  return <Example tabs={tabs} active={tabs[0].name} />
}

DefaultDisabled.story = {
  name: 'Default (disabled)',
}

export const DefaultNarrow = () => {
  const tabs = [
    { title: 'All', name: 'all' },
    { title: 'Starred', name: 'starred' },
  ]

  return <Example tabs={tabs} active={tabs[0].name} narrow />
}

DefaultNarrow.story = {
  name: 'Default (narrow)',
}

export const WithIcons = () => {
  const tabs = [
    { title: 'People', name: 'people', icon: 'organization' },
    { title: 'Data', name: 'data', icon: 'data' },
  ]

  return <Example tabs={tabs} active={tabs[0].name} narrow />
}

WithIcons.story = {
  name: 'With icons',
}

export const WithIconsDisabled = () => {
  const tabs = [
    { title: 'People', name: 'people', icon: 'organization' },
    { title: 'Data', name: 'data', icon: 'data', disabled: true },
  ]

  return <Example tabs={tabs} active={tabs[0].name} />
}

WithIconsDisabled.story = {
  name: 'With icons (disabled)',
}

export const Link = () => {
  const tabs = [
    { title: 'People', name: 'people', link: '/people' },
    { title: 'Data', name: 'data', link: '/data' },
  ]

  return <Example tabs={tabs} />
}

export const LinkDisabled = () => {
  const tabs = [
    { title: 'People', name: 'people', link: '/people' },
    { title: 'Data', name: 'data', link: '/data', disabled: true },
  ]

  return <Example tabs={tabs} />
}

LinkDisabled.story = {
  name: 'Link (disabled)',
}

export const LinkNarrow = () => {
  const tabs = [
    { title: 'People', name: 'people', link: '/people' },
    { title: 'Data', name: 'data', link: '/data' },
  ]

  return <Example tabs={tabs} narrow />
}

LinkNarrow.story = {
  name: 'Link (narrow)',
}
