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
import { connect } from 'react-redux'
import { defineMessages } from 'react-intl'

import sharedMessages from '../../lib/shared-messages'
import Message from '../../lib/components/message'
import FetchTable from '../fetch-table'

import { getDevicesList, searchDevicesList } from '../../actions/devices'

const m = defineMessages({
  deviceId: 'Device ID',
  connectedDevices: 'Connected Devices ({deviceCount})',
  add: 'Add Device',
})

const headers = [
  {
    name: 'device_id',
    displayName: m.deviceId,
  },
  {
    name: 'description',
    displayName: sharedMessages.description,
  },
]

@connect(function ({ application, devices }, props) {
  return {
    appId: application.application.application_id,
    totalCount: devices.totalCount,
  }
})
export default class DevicesTable extends React.Component {
  constructor (props) {
    super(props)

    this.searchDevicesList = filters => searchDevicesList(props.appId, filters)
    this.getDevicesList = filters => getDevicesList(props.appId, filters)
  }

  render () {
    const { totalCount } = this.props
    return (
      <FetchTable
        entity="devices"
        headers={headers}
        addMessage={m.add}
        tableTitle={<Message content={m.connectedDevices} values={{ deviceCount: totalCount }} />}
        getItemsAction={this.getDevicesList}
        searchItemsAction={this.searchDevicesList}
        itemPathPrefix="/devices"
        {...this.props}
      />
    )
  }
}
