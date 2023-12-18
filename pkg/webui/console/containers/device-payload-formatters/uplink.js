// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

import React, { useCallback, useState } from 'react'
import { useDispatch, useSelector } from 'react-redux'
import { isEmpty } from 'lodash'
import { useParams } from 'react-router-dom'

import PAYLOAD_FORMATTER_TYPES from '@console/constants/formatter-types'
import tts from '@console/api/tts'

import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import toast from '@ttn-lw/components/toast'
import Notification from '@ttn-lw/components/notification'

import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'
import RequireRequest from '@ttn-lw/lib/components/require-request'

import PayloadFormattersForm from '@console/components/payload-formatters-form'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'

import { hexToBase64 } from '@console/lib/bytes'
import { mayViewApplicationLink } from '@console/lib/feature-checks'
import { checkFromState } from '@account/lib/feature-checks'

import { updateDevice } from '@console/store/actions/devices'
import { getRepositoryPayloadFormatters } from '@console/store/actions/device-repository'

import { selectApplicationLink } from '@console/store/selectors/applications'
import {
  selectSelectedDeviceFormatters,
  selectSelectedDevice,
} from '@console/store/selectors/devices'
import { selectDeviceRepoPayloadFromatters } from '@console/store/selectors/device-repository'

import m from './messages'

const DevicePayloadFormatters = () => {
  const { appId, devId } = useParams()
  const device = useSelector(selectSelectedDevice)
  const mayViewLink = useSelector(state => checkFromState(mayViewApplicationLink, state))
  const link = useSelector(selectApplicationLink)
  const formatters = useSelector(selectSelectedDeviceFormatters)
  const decodeUplink = tts.As.decodeUplink
  const repositoryPayloadFormatters = useSelector(selectDeviceRepoPayloadFromatters)
  const [type, setType] = useState(
    Boolean(formatters)
      ? formatters.down_formatter || PAYLOAD_FORMATTER_TYPES.NONE
      : PAYLOAD_FORMATTER_TYPES.DEFAULT,
  )
  const dispatch = useDispatch()

  useBreadcrumbs('device.single.payload-formatters.uplink', [
    {
      path: `/applications/${appId}/devices/${devId}/payload-formatters/uplink`,
      content: sharedMessages.uplink,
    },
  ])

  const onSubmit = useCallback(
    async values => {
      if (values.type === PAYLOAD_FORMATTER_TYPES.DEFAULT) {
        return dispatch(
          attachPromise(
            updateDevice(appId, devId, {
              formatters: null,
            }),
          ),
        )
      }

      return dispatch(
        attachPromise(
          updateDevice(appId, devId, {
            formatters: {
              down_formatter: formatters?.down_formatter || PAYLOAD_FORMATTER_TYPES.NONE,
              down_formatter_parameter: formatters?.down_formatter_parameter,
              up_formatter: values.type,
              up_formatter_parameter: values.parameter,
            },
          }),
        ),
      )
    },
    [appId, devId, dispatch, formatters],
  )

  const onSubmitSuccess = useCallback(() => {
    toast({
      title: devId,
      message: sharedMessages.payloadFormattersUpdateSuccess,
      type: toast.types.SUCCESS,
    })
  }, [devId])

  const onTestSubmit = useCallback(
    async data => {
      const { f_port, payload, formatter, parameter } = data
      const { version_ids } = device

      const { uplink } = await decodeUplink(appId, devId, {
        uplink: {
          f_port,
          frm_payload: hexToBase64(payload),
          // `rx_metadata` and `settings` fields are required by the validation middleware in AS.
          // These fields won't affect the result of decoding an uplink message.
          rx_metadata: [
            { gateway_ids: { gateway_id: 'test' }, rssi: 42, channel_rssi: 42, snr: 4.2 },
          ],
          settings: {
            data_rate: { lora: { bandwidth: 125000, spreading_factor: 7 } },
            frequency: 868000000,
          },
        },
        version_ids: Object.keys(version_ids).length > 0 ? version_ids : undefined,
        formatter,
        parameter,
      })

      return uplink
    },
    [appId, devId, device, decodeUplink],
  )

  const onTypeChange = useCallback(type => {
    setType(type)
  }, [])

  const defaultFormatters = link?.default_formatters || {}
  const formatterType = Boolean(formatters)
    ? formatters.up_formatter || PAYLOAD_FORMATTER_TYPES.NONE
    : PAYLOAD_FORMATTER_TYPES.DEFAULT
  const formatterParameter = Boolean(formatters) ? formatters.up_formatter_parameter : undefined
  const appFormatterType = Boolean(defaultFormatters.up_formatter)
    ? defaultFormatters.up_formatter
    : PAYLOAD_FORMATTER_TYPES.NONE
  const appFormatterParameter = Boolean(defaultFormatters.up_formatter_parameter)
    ? defaultFormatters.up_formatter_parameter
    : undefined

  const isDefaultType = type === PAYLOAD_FORMATTER_TYPES.DEFAULT

  return (
    <RequireRequest
      requestAction={
        !isEmpty(device.version_ids)
          ? getRepositoryPayloadFormatters(appId, device.version_ids)
          : []
      }
    >
      <IntlHelmet title={sharedMessages.payloadFormattersUplink} />
      {!mayViewLink && <Notification content={m.mayNotViewLink} small warning />}
      <PayloadFormattersForm
        uplink
        allowReset
        allowTest
        onSubmit={onSubmit}
        onSubmitSuccess={onSubmitSuccess}
        onTestSubmit={onTestSubmit}
        title={sharedMessages.payloadFormattersUplink}
        initialType={formatterType}
        initialParameter={formatterParameter}
        defaultType={appFormatterType}
        defaultParameter={appFormatterParameter}
        onTypeChange={onTypeChange}
        isDefaultType={isDefaultType}
        repoFormatters={repositoryPayloadFormatters}
      />
    </RequireRequest>
  )
}

export default DevicePayloadFormatters
