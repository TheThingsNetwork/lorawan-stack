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

import Link from '@ttn-lw/components/link'
import LocationMap from '@ttn-lw/components/map'
import WidgetContainer from '@ttn-lw/components/widget-container'

import Message from '@ttn-lw/lib/components/message'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import style from './widget.styl'

const Map = ({ id, markers }) => {
  const leafletConfig = {
    zoomControl: false,
    zoom: 10,
    minZoom: 1,
  }
  const mapCenter =
    Boolean(markers) && markers.length !== 0
      ? [markers[0].position.latitude, markers[0].position.longitude]
      : undefined

  return (
    <div data-test-id="map-widget">
      {markers.length > 0 ? (
        <LocationMap
          id={id}
          mapCenter={mapCenter}
          markers={markers}
          leafletConfig={leafletConfig}
          widget
        />
      ) : (
        <div className={style.mapDisabled}>
          <Message component="span" content={sharedMessages.noLocation} />
        </div>
      )}
    </div>
  )
}

Map.propTypes = {
  id: PropTypes.string.isRequired,
  markers: PropTypes.arrayOf(
    PropTypes.shape({
      position: PropTypes.objectOf(PropTypes.number),
    }),
  ).isRequired,
}

const MapWidget = ({ id, markers, path }) => (
  <WidgetContainer
    title={sharedMessages.location}
    toAllUrl={path}
    linkMessage={sharedMessages.changeLocation}
  >
    <Link to={path} disabled={markers && markers.length > 0}>
      <Map id={id} markers={markers} />
    </Link>
  </WidgetContainer>
)

MapWidget.propTypes = {
  // Id is a string used to give the map a unique ID.
  id: PropTypes.string.isRequired,
  // Markers is an array of objects containing a specific properties.
  markers: PropTypes.arrayOf(
    // Position is a object containing two properties latitude and longitude
    // which are both numbers.
    PropTypes.shape({
      position: PropTypes.objectOf(PropTypes.number),
    }),
  ).isRequired,
  // Path is a string that is required to show the link at location form.
  path: PropTypes.string.isRequired,
}

export default MapWidget
