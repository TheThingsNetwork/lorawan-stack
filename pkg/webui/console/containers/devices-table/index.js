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
import bind from 'autobind-decorator'

import Button from '@ttn-lw/components/button'
import SafeInspector from '@ttn-lw/components/safe-inspector'
import Status from '@ttn-lw/components/status'

import FetchTable from '@ttn-lw/containers/fetch-table'

import Message from '@ttn-lw/lib/components/message'
import withRequest from '@ttn-lw/lib/components/with-request'

import LastSeen from '@console/components/last-seen'

import withFeatureRequirement from '@console/lib/components/with-feature-requirement'

import { selectNsConfig, selectJsConfig } from '@ttn-lw/lib/selectors/env'
import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import {
  checkFromState,
  mayCreateOrEditApplicationDevices,
  mayViewApplicationDevices,
} from '@console/lib/feature-checks'

import {
  getDeviceTemplateFormats,
  getDeviceTemplateFormatsError,
  getDeviceTemplateFormatsFetching,
} from '@console/store/actions/device-template-formats'
import { getDevicesList } from '@console/store/actions/devices'

import { selectSelectedApplicationId } from '@console/store/selectors/applications'
import { selectDeviceTemplateFormats } from '@console/store/selectors/device-template-formats'
import {
  selectDevices,
  selectDevicesTotalCount,
  selectDevicesFetching,
  selectDevicesError,
  selectDeviceDerivedLastSeen,
} from '@console/store/selectors/devices'

import style from './devices-table.styl'

const headers = [
  {
    name: 'ids.device_id',
    displayName: sharedMessages.id,
    sortable: true,
    sortKey: 'device_id',
  },
  {
    name: 'name',
    displayName: sharedMessages.name,
    sortable: true,
  },
  {
    name: 'ids.dev_eui',
    displayName: sharedMessages.devEUI,
    sortable: false,
    render: devEUI =>
      !Boolean(devEUI) ? (
        <Message className={style.none} content={sharedMessages.none} firstToLower />
      ) : (
        <SafeInspector data={devEUI} noTransform noCopyPopup small hideable={false} />
      ),
  },
  {
    name: 'ids.join_eui',
    displayName: sharedMessages.joinEUI,
    sortable: false,
    render: joinEUI =>
      !Boolean(joinEUI) ? (
        <Message className={style.none} content={sharedMessages.none} lowercase />
      ) : (
        <SafeInspector data={joinEUI} noTransform noCopyPopup small hideable={false} />
      ),
  },
  {
    name: '_derivedLastSeen',
    displayName: sharedMessages.lastSeen,
    width: 14,
    render: lastSeen =>
      lastSeen ? (
        <Status status="good">
          <LastSeen lastSeen={lastSeen} short />
        </Status>
      ) : (
        <Status status="mediocre" label={sharedMessages.unknown} />
      ),
  },
]

@connect(
  state => {
    const nsEnabled = selectNsConfig().enabled
    const jsEnabled = selectJsConfig().enabled
    const mayCreateDevices = checkFromState(mayCreateOrEditApplicationDevices, state)

    return {
      appId: selectSelectedApplicationId(state),
      deviceTemplateFormats: selectDeviceTemplateFormats(state),
      mayCreateDevices: mayCreateDevices && (nsEnabled || jsEnabled),
      mayImportDevices: mayCreateDevices,
      error: getDeviceTemplateFormatsError(state),
      fetching: getDeviceTemplateFormatsFetching(state),
    }
  },
  { getDeviceTemplateFormats },
)
@withFeatureRequirement(mayViewApplicationDevices)
@withRequest(({ getDeviceTemplateFormats }) => getDeviceTemplateFormats())
class DevicesTable extends React.Component {
  static propTypes = {
    appId: PropTypes.string.isRequired,
    devicePathPrefix: PropTypes.string,
    deviceTemplateFormats: PropTypes.shape({}).isRequired,
    error: PropTypes.error,
    fetching: PropTypes.bool,
    mayCreateDevices: PropTypes.bool.isRequired,
    mayImportDevices: PropTypes.bool.isRequired,
    totalCount: PropTypes.number,
  }

  static defaultProps = {
    devicePathPrefix: undefined,
    totalCount: 0,
    error: undefined,
    fetching: false,
  }

  constructor(props) {
    super(props)

    this.getDevicesList = filters =>
      getDevicesList(props.appId, filters, ['name'], { withLastSeen: true })
  }

  @bind
  baseDataSelector(state) {
    const { mayCreateDevices, appId } = this.props
    const devices = selectDevices(state)
    const decoratedDevices = []
    for (const device of devices) {
      decoratedDevices.push({
        ...device,
        _derivedLastSeen: selectDeviceDerivedLastSeen(state, appId, device.ids.device_id),
      })
    }
    return {
      devices: decoratedDevices,
      totalCount: selectDevicesTotalCount(state),
      fetching: selectDevicesFetching(state),
      error: selectDevicesError(state),
      mayAdd: mayCreateDevices,
    }
  }

  get importButton() {
    const { mayImportDevices, appId } = this.props

    return (
      mayImportDevices && (
        <Button.Link
          message={sharedMessages.importDevices}
          icon="import_devices"
          to={`/applications/${appId}/devices/import`}
          secondary
        />
      )
    )
  }

  render() {
    const { devicePathPrefix } = this.props
    return (
      <FetchTable
        entity="devices"
        headers={headers}
        addMessage={sharedMessages.addDevice}
        actionItems={this.importButton}
        tableTitle={<Message content={sharedMessages.devices} />}
        getItemsAction={this.getDevicesList}
        itemPathPrefix={devicePathPrefix}
        baseDataSelector={this.baseDataSelector}
        searchable
        {...this.props}
      />
    )
  }
}

export default DevicesTable
