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

import React, { useEffect, useCallback } from 'react'
import { useDispatch, useSelector } from 'react-redux'
import { useParams } from 'react-router-dom'

import RequireRequest from '@ttn-lw/lib/components/require-request'

import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'
import { selectNsConfig } from '@ttn-lw/lib/selectors/env'
import { combineDeviceIds } from '@ttn-lw/lib/selectors/id'

import {
  mayReadApplicationDeviceKeys,
  mayViewApplicationLink,
  checkFromState,
} from '@console/lib/feature-checks'

import { getDevice, stopDeviceEventsStream } from '@console/store/actions/devices'
import { getApplicationLink } from '@console/store/actions/link'
import { getNsFrequencyPlans } from '@console/store/actions/configuration'
import { getInfoByJoinEUI } from '@console/store/actions/claim'

import { selectSelectedApplicationId } from '@console/store/selectors/applications'

import Device from './device'

const DeviceContainer = props => {
  const { devId } = useParams()
  const appId = useSelector(selectSelectedApplicationId)
  const mayReadKeys = useSelector(state => checkFromState(mayReadApplicationDeviceKeys, state))
  const mayViewLink = useSelector(state => checkFromState(mayViewApplicationLink, state))

  const dispatch = useDispatch()

  const linkSelector = ['skip_payload_crypto', 'default_formatters']
  const deviceSelector = [
    'name',
    'description',
    'version_ids',
    'frequency_plan_id',
    'mac_settings',
    'resets_join_nonces',
    'supports_class_b',
    'supports_class_c',
    'supports_join',
    'last_seen_at',
    'lorawan_version',
    'lorawan_phy_version',
    'network_server_address',
    'application_server_address',
    'join_server_address',
    'locations',
    'formatters',
    'multicast',
    'net_id',
    'application_server_id',
    'application_server_kek_label',
    'network_server_kek_label',
    'claim_authentication_code',
    'attributes',
    'skip_payload_crypto_override',
  ]

  if (mayReadKeys) {
    deviceSelector.push('session')
    deviceSelector.push('pending_session')
    deviceSelector.push('root_keys')
  }

  const loadDeviceData = useCallback(
    async dispatch => {
      const nsEnabled = selectNsConfig().enabled

      const device = await dispatch(
        attachPromise(getDevice(appId, devId, deviceSelector, { ignoreNotFound: true })),
      )

      if (nsEnabled) {
        dispatch(getNsFrequencyPlans())
      }
      if (mayViewLink) {
        dispatch(getApplicationLink(appId, linkSelector))
      }

      dispatch(getInfoByJoinEUI({ join_eui: device.ids.join_eui }))
    },
    [appId, devId, deviceSelector, linkSelector, mayViewLink],
  )

  useEffect(
    () => () => dispatch(stopDeviceEventsStream(combineDeviceIds(appId, devId))),
    [appId, devId, dispatch],
  )

  return (
    <RequireRequest requestAction={loadDeviceData}>
      <Device {...props} />
    </RequireRequest>
  )
}

export default DeviceContainer
