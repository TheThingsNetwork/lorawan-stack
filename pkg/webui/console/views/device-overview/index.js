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

import React, { useCallback } from 'react'
import { useSelector } from 'react-redux'
import { Col, Row, Container } from 'react-grid-system'
import { defineMessages } from 'react-intl'
import { isPlainObject } from 'lodash'

import tts from '@console/api/tts'

import DataSheet from '@ttn-lw/components/data-sheet'
import ModalButton from '@ttn-lw/components/button/modal-button'
import toast from '@ttn-lw/components/toast'

import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'
import Message from '@ttn-lw/lib/components/message'
import DateTime from '@ttn-lw/lib/components/date-time'

import DeviceMap from '@console/components/device-map'

import DeviceEvents from '@console/containers/device-events'

import Require from '@console/lib/components/require'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'
import { composeDataUri, downloadDataUriAsFile } from '@ttn-lw/lib/data-uri'
import { selectNsConfig, selectStackConfig } from '@ttn-lw/lib/selectors/env'
import getHostFromUrl from '@ttn-lw/lib/host-from-url'

import {
  parseLorawanMacVersion,
  LORAWAN_VERSIONS,
  LORAWAN_PHY_VERSIONS,
} from '@console/lib/device-utils'

import { selectNsFrequencyPlans } from '@console/store/selectors/configuration'
import { selectSelectedDevice, isOtherClusterDevice } from '@console/store/selectors/devices'
import { selectSelectedApplicationId } from '@console/store/selectors/applications'

import style from './device-overview.styl'

const m = defineMessages({
  activationInfo: 'Activation information',
  sessionInfo: 'Session information',
  pendingSessionInfo: 'Session information (pending)',
  latestData: 'Latest data',
  keysNotExposed: 'Keys are not exposed',
  failedAccessOtherHostDevice:
    'The end device you attempted to visit is registered on a different cluster and needs to be accessed using its host Console.',
  downloadMacData: 'Download MAC data',
  sensitiveDataWarning:
    'The MAC data can contain sensitive information such as session keys that can be used to decrypt messages. <b>Do not share this information publicly</b>.',
  noSessionWarning:
    'The end device is currently not connected to the network (no active session). The MAC data will hence only contain the current MAC settings.',
  macStateError: 'There was an error and MAC state could not be included in the MAC data.',
  sessionStartedAt: 'Session start',
  noSession: 'This device has not joined the network yet',
})

const nsHost = getHostFromUrl(selectNsConfig().base_url)
const nsEnabled = selectNsConfig().enabled

const DeviceInfo = ({ frequencyPlans, device, onExport }) => {
  const {
    ids,
    description,
    version_ids = {},
    root_keys = {},
    session: actualSession,
    pending_session,
    created_at,
    lorawan_version,
    supports_join,
    frequency_plan_id,
    lorawan_phy_version,
  } = device

  // Get session keys.
  const session = actualSession || pending_session
  const { keys: sessionKeys = {}, dev_addr } = session || {}

  const nsEnabled = selectStackConfig().ns.enabled
  let lorawanVersion, frequencyPlan, lorawanVersionName, phyVersionName
  if (nsEnabled) {
    lorawanVersion = parseLorawanMacVersion(lorawan_version)
    frequencyPlan = frequencyPlans.find(f => f.id === frequency_plan_id)?.name
    lorawanVersionName = LORAWAN_VERSIONS.find(v => v.value === lorawan_version)?.label
    phyVersionName = LORAWAN_PHY_VERSIONS.find(v => v.value === lorawan_phy_version)?.label
  }
  const {
    f_nwk_s_int_key = { key: undefined },
    s_nwk_s_int_key = { key: undefined },
    nwk_s_enc_key = { key: undefined },
    app_s_key = { key: undefined },
  } = sessionKeys

  const sheetData = [
    {
      header: sharedMessages.generalInformation,
      items: [
        { key: sharedMessages.devID, value: ids.device_id, type: 'code', sensitive: false },
        ...(description
          ? {
              key: sharedMessages.description,
              value: description || <Message content={sharedMessages.noDesc} />,
            }
          : undefined),
        {
          key: sharedMessages.frequencyPlan,
          value: frequencyPlan,
          type: 'code',
          sensitive: false,
        },
        {
          key: sharedMessages.macVersion,
          value: lorawanVersionName,
          type: 'code',
          sensitive: false,
        },
        {
          key: sharedMessages.phyVersion,
          value: phyVersionName,
          type: 'code',
          sensitive: false,
        },
        { key: sharedMessages.createdAt, value: <DateTime value={created_at} /> },
      ],
    },
  ]

  // Add version info, if it is available.
  if (Object.keys(version_ids).length > 0) {
    sheetData.push({
      header: sharedMessages.hardware,
      items: [
        { key: sharedMessages.brand, value: version_ids.brand_id },
        { key: sharedMessages.model, value: version_ids.model_id },
        { key: sharedMessages.hardwareVersion, value: version_ids.hardware_version },
        { key: sharedMessages.firmwareVersion, value: version_ids.firmware_version },
      ],
    })
  }

  // Add activation info, if available.
  const activationInfoData = {
    header: m.activationInfo,
    items: [],
  }

  if (ids.join_eui || ids.dev_eui) {
    const joinEUI =
      lorawanVersion < 100
        ? sharedMessages.appEUIJoinEUI
        : lorawanVersion >= 104
        ? sharedMessages.joinEUI
        : sharedMessages.appEUI

    activationInfoData.items.push(
      { key: joinEUI, value: ids.join_eui, type: 'byte', sensitive: false },
      { key: sharedMessages.devEUI, value: ids.dev_eui, type: 'byte', sensitive: false },
    )

    // Add root keys, if available.
    if (Object.keys(root_keys).length > 0) {
      if (root_keys.app_key) {
        activationInfoData.items.push({
          key: sharedMessages.appKey,
          value: root_keys.app_key.key,
          type: 'byte',
          sensitive: true,
        })
      }
      if (root_keys.nwk_key) {
        activationInfoData.items.push({
          key: sharedMessages.nwkKey,
          value: root_keys.nwk_key.key,
          type: 'byte',
          sensitive: true,
        })
      }
    } else if (supports_join) {
      activationInfoData.items.push({
        key: sharedMessages.rootKeys,
        value: <Message content={sharedMessages.provisionedOnExternalJoinServer} />,
      })
    }
  }

  sheetData.push(activationInfoData)

  // Add session info, if available.
  const sessionInfoData = {
    header: pending_session && !actualSession ? m.pendingSessionInfo : m.sessionInfo,
    items: [],
    emptyMessage: m.noSession,
  }

  if (isPlainObject(session)) {
    if (session.started_at) {
      sessionInfoData.items.push({
        key: m.sessionStartedAt,
        value: <DateTime value={session.started_at} />,
      })
    }

    sessionInfoData.items.push(
      {
        key: sharedMessages.devAddr,
        value: dev_addr,
        type: 'byte',
        sensitive: false,
        enableUint32: true,
      },
      {
        key: sharedMessages.nwkSKey,
        value: f_nwk_s_int_key.key,
        type: 'byte',
        sensitive: true,
      },
      {
        key: sharedMessages.sNwkSIKey,
        value: s_nwk_s_int_key.key,
        type: 'byte',
        sensitive: true,
      },
      {
        key: sharedMessages.nwkSEncKey,
        value: nwk_s_enc_key.key,
        type: 'byte',
        sensitive: true,
      },
      { key: sharedMessages.appSKey, value: app_s_key.key, type: 'byte', sensitive: true },
    )
  }

  sheetData.push(sessionInfoData)

  const macStateAndSettings = {
    header: 'MAC data',
    items: [
      {
        value: (
          <ModalButton
            modalData={{
              message: session
                ? {
                    values: { b: msg => <b>{msg}</b> },
                    ...m.sensitiveDataWarning,
                  }
                : {
                    ...m.noSessionWarning,
                  },
            }}
            onApprove={onExport}
            message={m.downloadMacData}
            type="button"
            icon="file_download"
          />
        ),
      },
    ],
  }

  sheetData.push(macStateAndSettings)

  return (
    <div className={style.overviewInfo}>
      <div>
        <DataSheet data={sheetData} />
      </div>
    </div>
  )
}

DeviceInfo.propTypes = {
  device: PropTypes.device.isRequired,
  frequencyPlans: PropTypes.arrayOf(PropTypes.shape({})).isRequired,
  onExport: PropTypes.func.isRequired,
}

const DeviceOverview = () => {
  const appId = useSelector(selectSelectedApplicationId)
  const device = useSelector(selectSelectedDevice)
  const shouldRedirect = useSelector(() => isOtherClusterDevice(device))
  const frequencyPlans = useSelector(selectNsFrequencyPlans)

  const onExport = useCallback(async () => {
    const { ids, mac_settings, session, network_server_address } = device

    let result
    if (session && nsEnabled && getHostFromUrl(network_server_address) === nsHost) {
      try {
        result = await tts.Applications.Devices.getById(appId, ids.device_id, ['mac_state'])

        if (!('mac_state' in result)) {
          toast({
            title: m.downloadMacData,
            message: m.macStateError,
            type: toast.types.ERROR,
          })
        }
      } catch {
        toast({
          title: m.downloadMacData,
          message: m.macStateError,
          type: toast.types.ERROR,
        })
      }
    }

    const toExport = { mac_state: result?.mac_state, mac_settings }
    const toExportData = composeDataUri(JSON.stringify(toExport, undefined, 2))
    downloadDataUriAsFile(toExportData, `${ids.device_id}_mac_data_${Date.now()}.json`)
  }, [appId, device])

  const devIds = device && device.ids
  const otherwise = {
    redirect: '/applications',
    message: m.failedAccessOtherHostDevice,
  }
  return (
    <Require condition={!shouldRedirect} otherwise={otherwise}>
      <Container>
        <IntlHelmet title={sharedMessages.overview} />
        <Row className={style.head}>
          <Col md={12} lg={6}>
            <DeviceInfo frequencyPlans={frequencyPlans} device={device} onExport={onExport} />
          </Col>
          <Col md={12} lg={6} className={style.latestEvents}>
            <DeviceEvents devIds={devIds} widget />
            <DeviceMap device={device} />
          </Col>
        </Row>
      </Container>
    </Require>
  )
}

export default DeviceOverview
