// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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
import PropTypes from 'prop-types'
import { MapContainer, Marker, TileLayer } from 'react-leaflet'
import classnames from 'classnames'
import Leaflet from 'leaflet'
import shadowImg from 'leaflet/dist/images/marker-shadow.png'

import MarkerIcon from '@assets/auxiliary-icons/location_pin.svg'

import style from './map.styl'

// Reset default marker icon.

delete Leaflet.Icon.Default.prototype._getIconUrl
Leaflet.Icon.Default.mergeOptions({
  iconRetinaUrl: MarkerIcon,
  iconUrl: MarkerIcon,
  iconSize: [26, 36],
  shadowSize: [26, 36],
  iconAnchor: [13, 36],
  shadowAnchor: [8, 37],
  popupAnchor: [0, -40],
  // eslint-disable-next-line import/no-commonjs
  shadowUrl: shadowImg,
})

const defaultMinZoom = 7

const LocationMap = props => {
  const { className, mapCenter, clickable, widget, markers, leafletConfig, onClick } = props

  const hasValidCoordinates = mapCenter instanceof Array && mapCenter.length === 2

  const handleCreated = useCallback(
    map => {
      // Fix incomplete tile loading in some rare cases.
      map.invalidateSize()
      // Attach click handler.
      map.on('click', onClick)
    },
    [onClick],
  )

  return (
    <div
      className={classnames(style.container, className, { [style.widget]: widget })}
      data-test-id="location-map"
    >
      {hasValidCoordinates && (
        <MapContainer
          className={classnames(style.map, {
            [style.click]: clickable,
          })}
          minZoom={defaultMinZoom}
          whenCreated={handleCreated}
          center={mapCenter}
          {...leafletConfig}
        >
          <TileLayer
            url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
            attribution='&copy; <a href="http://osm.org/copyright">OpenStreetMap</a> contributors'
            noWrap
          />
          {markers.map(
            marker =>
              marker && (
                <Marker
                  key={`marker-${marker.position.latitude}-${marker.position.altitude}`}
                  position={[marker.position.latitude, marker.position.longitude]}
                />
              ),
          )}
        </MapContainer>
      )}
    </div>
  )
}

LocationMap.propTypes = {
  className: PropTypes.string,
  clickable: PropTypes.bool,
  // `LeafletConfig` is an object which can contain any number of properties
  // defined by the leaflet plugin and is used to overwrite the default
  // configuration of leaflet.
  leafletConfig: PropTypes.shape({
    zoom: PropTypes.number,
  }),
  mapCenter: PropTypes.arrayOf(PropTypes.number),
  // `markers` is an array of objects containing a specific properties.
  markers: PropTypes.arrayOf(
    // `position` is a object containing two properties latitude and longitude which are both numbers.
    PropTypes.shape({
      position: PropTypes.shape({
        longitude: PropTypes.number,
        latitude: PropTypes.number,
      }),
    }),
  ),
  onClick: PropTypes.func,
  // `widget` is a boolean used to add a class name to the map container div for styling.
  widget: PropTypes.bool,
}

LocationMap.defaultProps = {
  leafletConfig: {},
  className: undefined,
  widget: false,
  markers: [],
  onClick: () => null,
  mapCenter: undefined,
  clickable: false,
}

export default LocationMap
