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
import { isEmpty, isPlainObject } from 'lodash'
import { defineMessages } from 'react-intl'

import { END_DEVICE } from '@console/constants/entities'

import locationToMarkers from '@console/lib/location-to-markers'

import { selectDevices } from '@console/store/selectors/devices'

import MapPanel from '../map-panel'

const m = defineMessages({
  deviceLocations: 'Device locations',
})

const ApplicationMapPanel = () => {
  const devices = useSelector(selectDevices)

  const markers = []
  devices.forEach(device => {
    if (
      Boolean(device.locations) &&
      isPlainObject(device.locations) &&
      !isEmpty(device.locations)
    ) {
      markers.push(...locationToMarkers(device.locations, END_DEVICE))
    }
  })

  return (
    <MapPanel
      panelTitle={m.deviceLocations}
      markers={markers}
      entity={END_DEVICE}
      centerOnMarkers
    />
  )
}

export default ApplicationMapPanel
