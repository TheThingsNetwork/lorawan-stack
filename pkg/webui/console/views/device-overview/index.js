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
import { useSelector } from 'react-redux'
import { defineMessages } from 'react-intl'

import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'
import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'

import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'

import BlurryNetworkActivityPanel from '@console/components/blurry-network-activity-panel'
import DeviceMapPanel from '@console/components/device-map-panel'

import DeviceGeneralInformationPanel from '@console/containers/device-general-information-panel'
import DeviceInfoPanel from '@console/containers/device-info-panel'
import LatestDecodedPayloadPanel from '@console/containers/latest-decoded-payload-panel'

import Require from '@console/lib/components/require'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import { combineDeviceIds } from '@ttn-lw/lib/selectors/id'

import { selectNsFrequencyPlans } from '@console/store/selectors/configuration'
import {
  selectSelectedDevice,
  isOtherClusterDevice,
  selectDeviceEvents,
} from '@console/store/selectors/devices'
import { selectSelectedApplicationId } from '@console/store/selectors/applications'

const m = defineMessages({
  failedAccessOtherHostDevice:
    'The end device you attempted to visit is registered on a different cluster and needs to be accessed using its host Console.',
})

const DeviceOverview = () => {
  const appId = useSelector(selectSelectedApplicationId)
  const device = useSelector(selectSelectedDevice)
  const shouldRedirect = useSelector(() => isOtherClusterDevice(device))
  const frequencyPlans = useSelector(selectNsFrequencyPlans)
  const combinedId = combineDeviceIds(appId, device.ids.device_id)
  const events = useSelector(state => selectDeviceEvents(state, combinedId))

  useBreadcrumbs(
    'apps.single.devices.single.overview',
    <Breadcrumb
      path={`/applications/${appId}/devices/${device.ids.device_id}`}
      content={sharedMessages.endDeviceOverview}
    />,
  )

  const otherwise = {
    redirect: '/applications',
    message: m.failedAccessOtherHostDevice,
  }

  return (
    <Require condition={!shouldRedirect} otherwise={otherwise}>
      <IntlHelmet title={sharedMessages.overview} />
      <div className="container container--xl grid p-ls-s gap-ls-s">
        <div className="item-12 md:item-12 lg:item-6 sm:item-6">
          <DeviceInfoPanel events={events} />
        </div>
        <div className="item-12 md:item-12 lg:item-6 sm:item-6">
          <LatestDecodedPayloadPanel
            appId={appId}
            events={events}
            shortCutLinkPath={`/applications/${appId}/devices/${device.ids.device_id}/data`}
            className="h-full"
            isDevice
          />
        </div>
        <div className="item-12 md:item-12 lg:item-6 sm:item-6">
          <DeviceGeneralInformationPanel frequencyPlans={frequencyPlans} />
        </div>
        <div className="item-12 md:item-12 lg:item-6 sm:item-6">
          <BlurryNetworkActivityPanel />
        </div>
        <div className="item-12 md:item-12 lg:item-6 sm:item-6">
          <DeviceMapPanel />
        </div>
      </div>
    </Require>
  )
}

export default DeviceOverview
