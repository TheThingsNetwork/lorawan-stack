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

import React from 'react'
import { useSelector } from 'react-redux'
import { useParams } from 'react-router-dom'

import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'

import GenericNotFound from '@ttn-lw/lib/components/full-view-error/not-found'
import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'
import WithRootClass from '@ttn-lw/lib/components/with-root-class'

import DeviceEvents from '@console/containers/device-events'

import appStyle from '@console/views/app/app.styl'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import { selectSelectedDevice, selectSelectedDeviceId } from '@console/store/selectors/devices'

import style from './device-data.styl'

const Data = () => {
  const { appId } = useParams()

  const device = useSelector(selectSelectedDevice)
  const devId = useSelector(selectSelectedDeviceId)

  useBreadcrumbs('device.single.data', [
    {
      path: `/applications/${appId}/devices/${devId}/data`,
      content: sharedMessages.liveData,
    },
  ])

  if (!device) {
    return <GenericNotFound />
  }

  const { ids } = device

  return (
    <WithRootClass className={appStyle.stageFlex} id="stage">
      <div className={style.overflowContainer}>
        <IntlHelmet hideHeading title={sharedMessages.liveData} />
        <DeviceEvents devIds={ids} />
      </div>
    </WithRootClass>
  )
}

export default Data
