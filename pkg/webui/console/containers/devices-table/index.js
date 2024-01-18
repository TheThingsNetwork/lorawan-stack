// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

import React, { useCallback } from 'react'
import { useSelector } from 'react-redux'
import { defineMessages } from 'react-intl'
import { createSelector } from 'reselect'

import Button from '@ttn-lw/components/button'
import SafeInspector from '@ttn-lw/components/safe-inspector'
import Status from '@ttn-lw/components/status'
import DocTooltip from '@ttn-lw/components/tooltip/doc'
import Icon from '@ttn-lw/components/icon'

import FetchTable from '@ttn-lw/containers/fetch-table'

import Message from '@ttn-lw/lib/components/message'
import GenericNotFound from '@ttn-lw/lib/components/full-view-error/not-found'
import RequireRequest from '@ttn-lw/lib/components/require-request'

import LastSeen from '@console/components/last-seen'

import Require from '@console/lib/components/require'

import { selectNsConfig, selectJsConfig } from '@ttn-lw/lib/selectors/env'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import {
  checkFromState,
  mayCreateOrEditApplicationDevices,
  mayViewApplicationDevices,
} from '@console/lib/feature-checks'

import { getDeviceTemplateFormats } from '@console/store/actions/device-template-formats'
import { getDevicesList } from '@console/store/actions/devices'

import { selectSelectedApplicationId } from '@console/store/selectors/applications'
import { selectDeviceTemplateFormats } from '@console/store/selectors/device-template-formats'
import {
  selectDevicesTotalCount,
  isOtherClusterDevice,
  selectDevicesWithLastSeen,
} from '@console/store/selectors/devices'

import style from './devices-table.styl'

const m = defineMessages({
  otherClusterTooltip:
    'This end device is registered on a different cluster (`{host}`). To access this device, use the Console of the cluster that this end device was registered on.',
})

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
    name: 'status',
    displayName: sharedMessages.lastSeen,
    sortable: true,
    sortKey: 'last_seen_at',
    width: 14,
    render: status => {
      if (status.otherCluster) {
        const host = status.host
        return (
          <DocTooltip
            docPath="/getting-started/cloud-hosted"
            content={<Message content={m.otherClusterTooltip} values={{ host }} convertBackticks />}
            placement="top-end"
          >
            <Status status="unknown" label={sharedMessages.otherCluster}>
              <Icon icon="help_outline" textPaddedLeft small nudgeUp className="c-text-neutral-light" />
            </Status>
          </DocTooltip>
        )
      } else if (status._lastSeen) {
        return <LastSeen lastSeen={status._lastSeen} short />
      }

      return <Status status="mediocre" label={sharedMessages.never} />
    },
  },
]

const DevicesTable = () => (
  <Require featureCheck={mayViewApplicationDevices}>
    <RequireRequest requestAction={getDeviceTemplateFormats()}>
      <DevicesTableInner />
    </RequireRequest>
  </Require>
)

const DevicesTableInner = () => {
  const nsEnabled = selectNsConfig().enabled
  const jsEnabled = selectJsConfig().enabled
  const mayCreate = useSelector(state => checkFromState(mayCreateOrEditApplicationDevices, state))
  const appId = useSelector(selectSelectedApplicationId)
  const deviceTemplateFormats = useSelector(selectDeviceTemplateFormats)
  const mayCreateDevices = mayCreate && (nsEnabled || jsEnabled)
  const mayImportDevices = mayCreateDevices

  const getItemsAction = useCallback(
    filters =>
      getDevicesList(appId, filters, [
        'name',
        'application_server_address',
        'network_server_address',
        'join_server_address',
        'last_seen_at',
      ]),
    [appId],
  )

  const selectDecoratedDevices = createSelector(selectDevicesWithLastSeen, devices =>
    devices.map(device => ({
      ...device,
      status: {
        otherCluster: isOtherClusterDevice(device),
        host:
          device.application_server_address ||
          device.network_server_address ||
          device.join_server_address,
        _lastSeen: device.last_seen_at,
      },
      _meta: {
        clickable: !isOtherClusterDevice(device),
      },
    })),
  )

  const baseDataSelector = createSelector(
    selectDecoratedDevices,
    selectDevicesTotalCount,
    (devices, totalCount) => ({
      devices,
      totalCount,
      mayAdd: mayCreateDevices,
    }),
  )

  const importButton = mayImportDevices && (
    <Button.Link
      message={sharedMessages.importDevices}
      icon="import_devices"
      to={`/applications/${appId}/devices/import`}
      secondary
    />
  )

  if (!deviceTemplateFormats) {
    return <GenericNotFound />
  }

  return (
    <FetchTable
      entity="devices"
      defaultOrder="-created_at"
      headers={headers}
      addMessage={sharedMessages.registerEndDevice}
      actionItems={importButton}
      tableTitle={<Message content={sharedMessages.devices} />}
      getItemsAction={getItemsAction}
      baseDataSelector={baseDataSelector}
      itemPathPrefix={`/applications/${appId}/devices/`}
      searchable
    />
  )
}

export default DevicesTable
