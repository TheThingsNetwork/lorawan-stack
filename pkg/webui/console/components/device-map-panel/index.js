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

import { END_DEVICE } from '@console/constants/entities'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import locationToMarkers from '@console/lib/location-to-markers'

import { selectSelectedDevice } from '@console/store/selectors/devices'

import MapPanel from '../map-panel'

const DeviceMapPanel = ({ className }) => {
  const device = useSelector(selectSelectedDevice)
  const { device_id } = device.ids
  const { application_id } = device.ids.application_ids

  const markers =
    Boolean(device.locations) && isPlainObject(device.locations) && !isEmpty(device.locations)
      ? locationToMarkers(device.locations, END_DEVICE)
      : []

  const locationLink = `/applications/${application_id}/devices/${device_id}/location`

  return (
    <MapPanel
      panelTitle={sharedMessages.location}
      markers={markers}
      entity={END_DEVICE}
      locationLink={locationLink}
      className={className}
    />
  )
}

DeviceMapPanel.propTypes = {
  className: PropTypes.string,
}

DeviceMapPanel.defaultProps = {
  className: undefined,
}

export default DeviceMapPanel
