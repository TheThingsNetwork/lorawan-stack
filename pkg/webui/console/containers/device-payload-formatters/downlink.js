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

import RequireRequest from '@ttn-lw/lib/components/require-request'
import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'

import PayloadFormattersForm from '@console/components/payload-formatters-form'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'

import { checkFromState } from '@account/lib/feature-checks'
import { mayViewApplicationLink } from '@console/lib/feature-checks'

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
  const encodeDownlink = tts.As.encodeDownlink
  const repositoryPayloadFormatters = useSelector(selectDeviceRepoPayloadFromatters)
  const [type, setType] = useState(
    Boolean(formatters)
      ? formatters.down_formatter || PAYLOAD_FORMATTER_TYPES.NONE
      : PAYLOAD_FORMATTER_TYPES.DEFAULT,
  )
  const dispatch = useDispatch()

  useBreadcrumbs('device.single.payload-formatters.downlink', [
    {
      path: `/applications/${appId}/devices/${devId}/payload-formatters/downlink`,
      content: sharedMessages.downlink,
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
              down_formatter: values.type,
              down_formatter_parameter: values.parameter,
              up_formatter: formatters?.up_formatter || PAYLOAD_FORMATTER_TYPES.NONE,
              up_formatter_parameter: formatters?.up_formatter_parameter,
            },
          }),
        ),
      )
    },
    [dispatch, appId, devId, formatters],
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

      const { downlink } = await encodeDownlink(appId, devId, {
        downlink: {
          f_port,
          decoded_payload: JSON.parse(payload),
        },
        version_ids: Object.keys(version_ids).length > 0 ? version_ids : undefined,
        formatter,
        parameter,
      })

      return downlink
    },
    [encodeDownlink, appId, devId, device],
  )

  const onTypeChange = useCallback(type => {
    setType(type)
  }, [])

  const defaultFormatters = link?.default_formatters || {}
  const formatterType = Boolean(formatters)
    ? formatters.down_formatter || PAYLOAD_FORMATTER_TYPES.NONE
    : PAYLOAD_FORMATTER_TYPES.DEFAULT
  const formatterParameter = Boolean(formatters) ? formatters.down_formatter_parameter : undefined
  const appFormatterType = Boolean(defaultFormatters.down_formatter)
    ? defaultFormatters.down_formatter
    : PAYLOAD_FORMATTER_TYPES.NONE
  const appFormatterParameter = Boolean(defaultFormatters.down_formatter_parameter)
    ? defaultFormatters.down_formatter_parameter
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
      <IntlHelmet title={sharedMessages.payloadFormattersDownlink} />
      {!mayViewLink && <Notification content={m.mayNotViewLink} small warning />}
      <PayloadFormattersForm
        uplink={false}
        allowReset
        allowTest
        onSubmit={onSubmit}
        onSubmitSuccess={onSubmitSuccess}
        onTestSubmit={onTestSubmit}
        title={sharedMessages.payloadFormattersDownlink}
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
