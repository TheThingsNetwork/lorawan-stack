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

import React from 'react'
import PropTypes from 'prop-types'
import {
  MapContainer,
  Marker,
  CircleMarker,
  Circle,
  TileLayer,
  useMapEvent,
  useMap,
} from 'react-leaflet'
import classnames from 'classnames'
import Leaflet, { latLngBounds } from 'leaflet'
import shadowImg from 'leaflet/dist/images/marker-shadow.png'

import MarkerIcon from '@assets/auxiliary-icons/location_pin.svg'
import COLORS from '@ttn-lw/constants/colors'

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

const MarkerRenderer = ({ marker }) => {
  if (!marker) {
    return null
  }

  const hasAccuracy = typeof marker.accuracy === 'number'
  const children = (
    <>
      {typeof marker.accuracy === 'number' && (
        <Circle
          center={[marker.position.latitude, marker.position.longitude]}
          radius={marker.accuracy}
          weight={1}
          fillOpacity={0.1}
        />
      )}
      {marker.children}
    </>
  )
  return hasAccuracy ? (
    <CircleMarker
      key={`marker-${marker.position.latitude}-${marker.position.longitude}`}
      center={[marker.position.latitude, marker.position.longitude]}
      radius={8}
      children={children}
      color="#ffffff"
      fillColor={COLORS.C_ACTIVE_BLUE}
      fillOpacity={1}
    />
  ) : (
    <Marker
      key={`marker-${marker.position.latitude}-${marker.position.longitude}`}
      position={[marker.position.latitude, marker.position.longitude]}
      children={children}
    />
  )
}

const Controller = ({ onClick, centerOnMarkers, markers, bounds }) => {
  const map = useMap()

  useMapEvent('click', onClick)
  // Fix incomplete tile loading in some rare cases.
  map.invalidateSize()
  // Attach click handler.
  if (centerOnMarkers && markers.length > 1) {
    map.fitBounds(bounds, { padding: [50, 50], maxZoom: 14 })
  }
  return markers.map(marker => (
    <MarkerRenderer
      key={`${marker.position.latitude}-${marker.position.longitude}`}
      marker={marker}
    />
  ))
}

const LocationMap = props => {
  const {
    className,
    mapCenter,
    clickable,
    widget,
    markers,
    leafletConfig,
    centerOnMarkers,
    ...rest
  } = props

  const bounds = latLngBounds(
    markers.map(marker => [marker.position.latitude, marker.position.longitude]),
  )

  const hasValidCoordinates = mapCenter instanceof Array && mapCenter.length === 2

  let center = [0, 0]

  if (centerOnMarkers && markers.length >= 1) {
    center = bounds.getCenter()
  } else if (hasValidCoordinates) {
    center = mapCenter
  }

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
          center={center}
          {...leafletConfig}
        >
          <TileLayer
            url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
            attribution='&copy; <a href="http://osm.org/copyright">OpenStreetMap</a> contributors'
            noWrap
          />
          <Controller
            bounds={bounds}
            centerOnMarkers={centerOnMarkers}
            markers={markers}
            {...rest}
          />
        </MapContainer>
      )}
    </div>
  )
}

LocationMap.defaultProps = {
  centerOnMarkers: true,
  leafletConfig: {},
  className: undefined,
  widget: false,
  markers: [],
  onClick: () => null,
  mapCenter: undefined,
  clickable: false,
}

MarkerRenderer.propTypes = {
  marker: PropTypes.shape({
    position: PropTypes.shape({
      longitude: PropTypes.number,
      latitude: PropTypes.number,
    }),
    accuracy: PropTypes.number,
    children: PropTypes.node,
  }).isRequired,
}

LocationMap.propTypes = {
  // Whether the map should center on the provided markers (if exist), once loaded (regardless of `mapCenter`).
  centerOnMarkers: PropTypes.bool,
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
  markers: PropTypes.arrayOf(MarkerRenderer.propTypes.marker),
  onClick: PropTypes.func,
  // `widget` is a boolean used to add a class name to the map container div for styling.
  widget: PropTypes.bool,
}

export default LocationMap
