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
import PropTypes from '../../../lib/prop-types'
import MapWidget from '../../../components/map/widget/'

export default class GatewayMap extends React.Component {
  static propTypes = {
    // Gateway is an object.
    gateway: PropTypes.gateway.isRequired,
  }

  render() {
    const { gateway } = this.props
    const { gateway_id } = gateway.ids

    const markers =
      gateway.antennas && gateway.antennas.length > 0 && gateway.antennas[0].location
        ? gateway.antennas.map(location => ({
            position: {
              latitude: location.location.latitude || 0,
              longitude: location.location.longitude || 0,
            },
          }))
        : []

    return (
      <MapWidget
        id="gateway-map-widget"
        markers={markers}
        path={`/gateways/${gateway_id}/location`}
      />
    )
  }
}
