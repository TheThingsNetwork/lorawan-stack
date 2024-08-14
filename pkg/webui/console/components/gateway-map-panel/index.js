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

import { GATEWAY } from '@console/constants/entities'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import locationToMarkers from '@console/lib/location-to-markers'

import MapPanel from '../map-panel'

const GatewayMapPanel = ({ gateway, className }) => {
  const { gateway_id } = gateway.ids

  const markers = gateway.antennas
    ? locationToMarkers(
        gateway.antennas.map(antenna => antenna.location),
        GATEWAY,
      )
    : []

  const locationLink = `/gateways/${gateway_id}/location`

  return (
    <MapPanel
      panelTitle={sharedMessages.location}
      markers={markers}
      entity={GATEWAY}
      locationLink={locationLink}
      className={className}
    />
  )
}

GatewayMapPanel.propTypes = {
  className: PropTypes.string,
  gateway: PropTypes.gateway.isRequired,
}

GatewayMapPanel.defaultProps = {
  className: undefined,
}

export default GatewayMapPanel
