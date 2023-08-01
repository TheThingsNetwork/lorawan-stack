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

import MapWidget from '@ttn-lw/components/map/widget'

import PropTypes from '@ttn-lw/lib/prop-types'

import locationToMarkers from '@console/lib/location-to-markers'

const DeviceMap = ({ device }) => {
  const { device_id } = device.ids
  const { application_id } = device.ids.application_ids

  const markers = locationToMarkers(device.locations)

  return (
    <MapWidget
      id="device-map-widget"
      markers={markers}
      path={`/applications/${application_id}/devices/${device_id}/location`}
    />
  )
}

DeviceMap.propTypes = {
  device: PropTypes.device.isRequired,
}

export default DeviceMap
