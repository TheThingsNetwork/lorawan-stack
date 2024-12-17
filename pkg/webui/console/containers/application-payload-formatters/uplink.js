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
import { defineMessages } from 'react-intl'
import { useParams } from 'react-router-dom'

import PAYLOAD_FORMATTER_TYPES from '@console/constants/formatter-types'

import Notification from '@ttn-lw/components/notification'
import PageTitle from '@ttn-lw/components/page-title'
import toast from '@ttn-lw/components/toast'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import ErrorNotification from '@ttn-lw/components/error-notification'

import PayloadFormattersForm from '@console/components/payload-formatters-form'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import { isNotFoundError } from '@ttn-lw/lib/errors/utils'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'

import { checkFromState, mayViewApplicationLink } from '@console/lib/feature-checks'

import { updateApplicationLink, updateApplicationLinkSuccess } from '@console/store/actions/link'

import {
  selectApplicationLinkError,
  selectApplicationLinkFormatters,
} from '@console/store/selectors/applications'

const m = defineMessages({
  title: 'Default uplink payload formatter',
  infoText:
    'You can use the "Payload formatter" tab of individual end devices to test uplink payload formatters and to define individual payload formatter settings per end device.',
  uplinkResetWarning:
    'You do not have sufficient rights to view the current uplink payload formatter. Only overwriting is allowed.',
})

const ApplicationPayloadFormatters = () => {
  const { appId } = useParams()
  const formatters = useSelector(selectApplicationLinkFormatters) || {}
  const linkError = useSelector(selectApplicationLinkError)
  const mayViewLink = useSelector(state => checkFromState(mayViewApplicationLink, state))
  const [type, setType] = useState(formatters.down_formatter || PAYLOAD_FORMATTER_TYPES.NONE)
  const dispatch = useDispatch()

  useBreadcrumbs(
    'apps.single.payload-formatters.uplink',
    <Breadcrumb
      path={`/applications/${appId}/payload-formatters/uplink`}
      content={sharedMessages.uplink}
    />,
  )

  const onSubmit = useCallback(
    async values =>
      await dispatch(
        attachPromise(
          updateApplicationLink(appId, {
            default_formatters: {
              down_formatter: formatters.down_formatter || PAYLOAD_FORMATTER_TYPES.NONE,
              down_formatter_parameter: formatters.down_formatter_parameter || '',
              up_formatter: values.type,
              up_formatter_parameter: values.parameter,
            },
          }),
        ),
      ),
    [appId, formatters, dispatch],
  )

  const onSubmitSuccess = useCallback(
    link => {
      toast({
        title: appId,
        message: sharedMessages.payloadFormattersUpdateSuccess,
        type: toast.types.SUCCESS,
      })
      dispatch(updateApplicationLinkSuccess(link))
    },
    [appId, dispatch],
  )

  const onTypeChange = useCallback(type => {
    setType(type)
  }, [])

  const isNoneType = type === PAYLOAD_FORMATTER_TYPES.NONE
  const hasError = Boolean(linkError) && !isNotFoundError(linkError)

  return (
    <div className="container container--xl grid gap-ls-xxs box-border">
      <div className="item-12">
        <PageTitle title={m.title} />
        {hasError && <ErrorNotification content={linkError} small />}
        {!isNoneType && <Notification className="mb-ls-s" small info content={m.infoText} />}
        {!mayViewLink && <Notification content={m.uplinkResetWarning} info small />}
        <PayloadFormattersForm
          uplink
          onSubmit={onSubmit}
          onSubmitSuccess={onSubmitSuccess}
          initialType={formatters.up_formatter || PAYLOAD_FORMATTER_TYPES.NONE}
          initialParameter={formatters.up_formatter_parameter || ''}
          onTypeChange={onTypeChange}
        />
      </div>
    </div>
  )
}

export default ApplicationPayloadFormatters
