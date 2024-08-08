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
import { useParams } from 'react-router-dom'

import PageTitle from '@ttn-lw/components/page-title'
import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'

import RequireRequest from '@ttn-lw/lib/components/require-request'

import DeviceOnboardingForm from '@console/containers/device-onboarding-form'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import { selectJsConfig } from '@ttn-lw/lib/selectors/env'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'

import { listBrands } from '@console/store/actions/device-repository'
import { getJoinEUIPrefixes } from '@console/store/actions/join-server'

const DeviceAdd = () => {
  const { appId } = useParams()
  const { enabled: jsEnabled } = selectJsConfig()
  const requestAction = useCallback(
    async dispatch => {
      if (jsEnabled) {
        await dispatch(attachPromise(getJoinEUIPrefixes()))
      }
      await dispatch(attachPromise(listBrands(appId, {}, ['name', 'lora_alliance_vendor_id'])))
    },
    [appId, jsEnabled],
  )

  useBreadcrumbs(
    'apps.single.devices.add',
    <Breadcrumb
      path={`/applications/${appId}/devices/add`}
      content={sharedMessages.registerGateway}
    />,
  )

  return (
    <RequireRequest requestAction={requestAction}>
      <div className="container container--xxl">
        <PageTitle title={sharedMessages.registerEndDevice} className="mb-cs-m" />
        <DeviceOnboardingForm />
      </div>
    </RequireRequest>
  )
}

export default DeviceAdd
