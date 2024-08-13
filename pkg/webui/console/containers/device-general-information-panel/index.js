// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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
import { useSelector } from 'react-redux'
import { defineMessages } from 'react-intl'
import { isPlainObject } from 'lodash'

import Panel from '@ttn-lw/components/panel'
import DataSheet from '@ttn-lw/components/data-sheet'

import Message from '@ttn-lw/lib/components/message'
import DateTime from '@ttn-lw/lib/components/date-time'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import { selectStackConfig } from '@ttn-lw/lib/selectors/env'
import PropTypes from '@ttn-lw/lib/prop-types'

import {
  LORAWAN_PHY_VERSIONS,
  LORAWAN_VERSIONS,
  parseLorawanMacVersion,
} from '@console/lib/device-utils'

import { selectSelectedDevice } from '@console/store/selectors/devices'

import style from './device-general-information-panel.styl'

const m = defineMessages({
  activationInfo: 'Activation information',
  sessionInfo: 'Session information',
  pendingSessionInfo: 'Session information (pending)',
  sessionStartedAt: 'Session start',
  noSession: 'This device has not joined the network yet',
})

const DeviceGeneralInformationPanel = ({ frequencyPlans }) => {
  const device = useSelector(selectSelectedDevice)
  const {
    ids,
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
        {
          key: sharedMessages.frequencyPlan,
          value: frequencyPlan,
        },
        {
          key: sharedMessages.macVersion,
          value: lorawanVersionName,
        },
        {
          key: sharedMessages.phyVersion,
          value: phyVersionName,
        },
        { key: sharedMessages.createdAt, value: <DateTime value={created_at} /> },
      ],
    },
  ]

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

    sessionInfoData.items.push({
      key: sharedMessages.devAddr,
      value: dev_addr,
      type: 'byte',
      sensitive: false,
      enableUint32: true,
    })

    if (lorawanVersion >= 100 && lorawanVersion < 110) {
      sessionInfoData.items.push(
        {
          key: sharedMessages.nwkSKey,
          value: f_nwk_s_int_key.key,
          type: 'byte',
          sensitive: true,
        },
        { key: sharedMessages.appSKey, value: app_s_key.key, type: 'byte', sensitive: true },
      )
    } else {
      sessionInfoData.items.push(
        {
          key: lorawanVersion >= 110 ? sharedMessages.fNwkSIntKey : sharedMessages.nwkSKey,
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
  }

  sheetData.push(sessionInfoData)

  return (
    <Panel className={style.deviceGeneralInfoPanel}>
      <DataSheet data={sheetData} />
    </Panel>
  )
}

DeviceGeneralInformationPanel.propTypes = {
  frequencyPlans: PropTypes.arrayOf(PropTypes.shape({})).isRequired,
}

export default DeviceGeneralInformationPanel
