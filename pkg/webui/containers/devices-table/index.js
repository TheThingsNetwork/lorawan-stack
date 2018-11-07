// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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
import bind from 'autobind-decorator'
import { push } from 'connected-react-router'

import debounce from '../../lib/debounce'
import sharedMessages from '../../lib/shared-messages'
import Message from '../../lib/components/message'
import Tabular from '../../components/table'
import Input from '../../components/input'
import Button from '../../components/button'

import { getDevicesList, searchDevicesList } from '../../actions/devices'

import style from './devices-table.styl'

const m = defineMessages({
  deviceId: 'Device ID',
  connectedDevices: 'Connected Devices {deviceCount}',
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

@connect(function ({ devices, application }, props) {
  return {
    devices: devices.devices,
    totalCount: devices.totalCount,
    fetching: devices.fetching,
    fetchingSearch: devices.fetchingSearch,
    appId: application.application.application_id,
    pathname: location.pathname,
  }
})
@bind
export default class DevicesTable extends React.Component {

  constructor (props) {
    super(props)

    this.state = {
      query: '',
      page: 1,
      order: undefined,
      orderBy: undefined,
    }

    this.requestSearch = debounce(this.requestSearch, 350)
  }

  fetchDevices () {
    const { appId, dispatch, pageSize } = this.props
    const filters = { ...this.state, pageSize }

    if (filters.query) {
      dispatch(searchDevicesList(appId, filters))
    } else {
      dispatch(getDevicesList(appId, filters))
    }
  }

  onPageChange (page) {
    this.setState({ page }, () => this.fetchDevices())
  }

  requestSearch () {
    this.fetchDevices()
  }

  onQueryChange (query) {
    this.setState({ query }, () => this.requestSearch())
  }

  onOrderChange (order, orderBy) {
    this.setState({ order, orderBy }, () => this.fetchDevices())
  }

  onDeviceAdd () {
    const { dispatch, pathname } = this.props

    dispatch(push(`${pathname}/devices/add`))
  }

  onDeviceClick (id) {
    const { dispatch, pathname, devices } = this.props
    const { device_id } = devices[id]

    dispatch(push(`${pathname}/devices/${device_id}`))
  }

  componentDidMount () {
    this.fetchDevices()
  }

  render () {
    const {
      devices,
      totalCount,
      fetching,
      fetchingSearch,
      pageSize,
    } = this.props
    const { page, query } = this.state

    const deviceCount = `(${totalCount})`
    const devs = devices.map(device => ({ ...device, clickable: true }))

    return (
      <div>
        <div className={style.filters}>
          <div className={style.filtersLeft}>
            <Message
              component="h3"
              content={m.connectedDevices}
              values={{ deviceCount }}
            />
          </div>
          <div className={style.filtersRight}>
            <Input
              value={query}
              icon="search"
              loading={fetchingSearch}
              onChange={this.onQueryChange}
            />
            <Button
              onClick={this.onDeviceAdd}
              className={style.addButton}
              message={m.add}
              icon="add"
            />
          </div>
        </div>
        <Tabular
          paginated
          page={page}
          totalCount={totalCount}
          pageSize={pageSize}
          onRowClick={this.onDeviceClick}
          onPageChange={this.onPageChange}
          loading={fetching}
          headers={headers}
          data={devs}
          emptyMessage={sharedMessages.noMatch}
        />
      </div>
    )
  }
}





