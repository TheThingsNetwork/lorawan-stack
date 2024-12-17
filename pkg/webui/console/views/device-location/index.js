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
import { defineMessages } from 'react-intl'
import { useSelector, useDispatch } from 'react-redux'

import Breadcrumb from '@ttn-lw/components/breadcrumbs/breadcrumb'
import { useBreadcrumbs } from '@ttn-lw/components/breadcrumbs/context'

import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'
import Message from '@ttn-lw/lib/components/message'

import LocationForm from '@console/components/location-form'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import attachPromise from '@ttn-lw/lib/store/actions/attach-promise'

import locationToMarkers from '@console/lib/location-to-markers'

import { updateDevice } from '@console/store/actions/devices'

import { selectSelectedApplicationId } from '@console/store/selectors/applications'
import { selectSelectedDevice, selectSelectedDeviceId } from '@console/store/selectors/devices'

const m = defineMessages({
  locationInfoTitle: 'Understanding end device location settings',
  locationInfo:
    'The Things Stack is capable of storing locations from multiple sources at the same time. Next to automatic location updates sourced from frame payloads of the end device and various other means of resolving location, it is also possible to set a location manually. You can use this form to update this location-type.',
  setDeviceLocation: 'Set end device location manually',
})

const DeviceGeneralSettings = () => {
  const dispatch = useDispatch()
  const appId = useSelector(selectSelectedApplicationId)
  const devId = useSelector(selectSelectedDeviceId)
  const device = useSelector(selectSelectedDevice)

  useBreadcrumbs(
    'device.single.location',
    <Breadcrumb
      path={`/applications/${appId}/devices/${devId}/location`}
      content={sharedMessages.location}
    />,
  )

  const handleSubmit = useCallback(
    async location => {
      const { locations } = device

      await dispatch(
        attachPromise(updateDevice(appId, devId, { locations: { ...locations, user: location } })),
      )
    },
    [appId, devId, device, dispatch],
  )

  const handleDelete = useCallback(
    async deleteAll => {
      const { locations } = device
      const { user, ...nonUserLocations } = locations || {}

      const newLocations = {
        ...(!deleteAll ? nonUserLocations : undefined),
      }

      return dispatch(attachPromise(updateDevice(appId, devId, { locations: newLocations })))
    },
    [appId, devId, device, dispatch],
  )

  const { user, ...nonUserLocations } = device.locations || {}

  return (
    <div className="container container--xl grid">
      <IntlHelmet title={sharedMessages.location} />
      <div className="item-12">
        <LocationForm
          entityId={devId}
          formTitle={m.setDeviceLocation}
          initialValues={user}
          additionalMarkers={locationToMarkers(nonUserLocations)}
          onSubmit={handleSubmit}
          onDelete={handleDelete}
          centerOnMarkers
        />
      </div>
      <div className="item-12 xl:item-4">
        <Message content={m.locationInfoTitle} component="h4" className="mb-0 mt-ls-xl" />
        <Message content={m.locationInfo} component="p" />
      </div>
    </div>
  )
}

export default DeviceGeneralSettings
