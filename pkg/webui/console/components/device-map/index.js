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
import PropTypes from 'prop-types'
import MapWidget from '../../../components/map/widget/'

export default class DeviceMap extends React.Component {
  render() {
    const { device } = this.props

    const markers =
      device.locations && device.locations.user
        ? [
            {
              position: {
                latitude: device.locations.user.latitude || 0,
                longitude: device.locations.user.longitude || 0,
              },
            },
          ]
        : []

    return (
      <MapWidget
        id="device-map-widget"
        markers={markers}
        path={`/applications/${device.ids.device_id}/devices/${device.ids.application_ids.application_id}/location`}
      />
    )
  }
}

DeviceMap.propTypes = {
  device: PropTypes.shape({
    locations: PropTypes.shape({
      user: PropTypes.shape({
        latitude: PropTypes.number,
        longitude: PropTypes.number,
      }),
    }),
    ids: PropTypes.shape({
      device_id: PropTypes.string,
      application_ids: PropTypes.shape({
        application_id: PropTypes.string,
      }),
    }),
  }),
}

DeviceMap.defaultProps = {
  device: {},
}
