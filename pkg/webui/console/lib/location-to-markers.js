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

import React from 'react'
import { isArray, isEmpty, isPlainObject } from 'lodash'
import { Tooltip } from 'react-leaflet'

import Message from '@ttn-lw/lib/components/message'

import sharedMessages from '@ttn-lw/lib/shared-messages'

const sourceMessages = {
  SOURCE_GPS: sharedMessages.locationSourceGps,
  SOURCE_REGISTRY: sharedMessages.locationSourceRegistry,
  SOURCE_UNKNOWN: sharedMessages.locationSourceUnknown,
  SOURCE_IP_GEOLOCATION: sharedMessages.locationSourceIpGeolocation,
  SOURCE_WIFI_RSSI_GEOLOCATION: sharedMessages.locationSourceWifiRssi,
  SOURCE_BT_RSSI_GEOLOCATION: sharedMessages.locationSourceBtRssi,
  SOURCE_LORA_RSSI_GEOLOCATION: sharedMessages.locationSourceLoraRssi,
  SOURCE_LORA_TDOA_GEOLOCATION: sharedMessages.locationSourceLoraTdoa,
  SOURCE_COMBINED_GEOLOCATION: sharedMessages.locationSourceCombined,
}

const createLocationObject = (location, key) => ({
  position: {
    latitude: location.latitude || 0,
    longitude: location.longitude || 0,
  },
  accuracy: location.accuracy,
  children: (
    <Tooltip direction="top" offset={[-15, -10]} opacity={1}>
      <Message
        component="strong"
        content={location?.source ? sourceMessages[location.source] : sourceMessages.SOURCE_UNKNOWN}
      />
      <br />
      <Message
        content={
          key === 'user' || location?.source === 'SOURCE_REGISTRY'
            ? sharedMessages.locationMarkerDescriptionUser
            : location.trusted === false
              ? sharedMessages.locationMarkerDescriptionUntrusted
              : sharedMessages.locationMarkerDescriptionNonUser
        }
      />
      <br />
      Long: {location.longitude} / Lat: {location.latitude}
    </Tooltip>
  ),
})

export default locations => {
  if (isPlainObject(locations)) {
    return Object.keys(locations).map(key => createLocationObject(locations[key], key))
  }

  if (isArray(locations)) {
    return locations
      .filter(l => Boolean(l) && isPlainObject(l) && !isEmpty(l))
      .map(location => createLocationObject(location))
  }

  return []
}
