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

import React, { useEffect } from 'react'
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

import DefaultMarkerIcon from '@assets/auxiliary-icons/default-map-pin.svg'
import GatewayMarkerIcon from '@assets/auxiliary-icons/gateway-map-pin.svg'
import DeviceMarkerIcon from '@assets/auxiliary-icons/device-map-pin.svg'
import COLORS from '@ttn-lw/constants/colors'
import { END_DEVICE, GATEWAY } from '@console/constants/entities'

import PropTypes from '@ttn-lw/lib/prop-types'

import style from './map.styl'

const defaultMinZoom = 7

const MarkerRenderer = ({ marker }) => {
  if (!marker) {
    return null
  }

  if (!marker.mapPinType) {
    marker.mapPinType = 'DEFAULT'
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
  const markerImage =
    marker.mapPinType === GATEWAY
      ? GatewayMarkerIcon
      : marker.mapPinType === END_DEVICE
        ? DeviceMarkerIcon
        : DefaultMarkerIcon

  const customIcon = Leaflet.icon({
    iconRetinaUrl: markerImage,
    iconUrl: markerImage,
    ...(marker.mapPinType === 'DEFAULT'
      ? {
          iconSize: [26, 36],
          shadowSize: [26, 36],
          iconAnchor: [13, 36],
          shadowAnchor: [8, 37],
          shadowUrl: shadowImg,
        }
      : {
          iconSize: [36, 36],
          iconAnchor: [18, 18],
          shadowUrl: null,
        }),
  })

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
      icon={customIcon}
    />
  )
}

const Controller = ({ onClick, centerOnMarkers, markers, bounds }) => {
  const map = useMap()

  useEffect(() => {
    const handleWheel = e => {
      if (e.ctrlKey || e.metaKey) {
        e.preventDefault()

        const delta = e.deltaY > 0 ? 1 : -1 // Determine scroll direction
        const zoomLevel = map.getZoom() - delta // Calculate the new zoom level

        map.setZoom(zoomLevel)
      }
    }

    map.getContainer().addEventListener('wheel', handleWheel)

    return () => {
      map.getContainer().removeEventListener('wheel', handleWheel)
    }
  }, [map])

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
    panel,
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
      className={classnames(style.container, className, {
        [style.widget]: widget,
        [style.panel]: panel,
      })}
      data-test-id="location-map"
    >
      {hasValidCoordinates && (
        <MapContainer
          className={classnames(style.map, {
            [style.click]: clickable,
          })}
          minZoom={defaultMinZoom}
          center={center}
          maxBounds={[
            [-90, -180],
            [90, 180],
          ]}
          scrollWheelZoom={false}
          maxBoundsViscosity={1.0}
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
  panel: false,
}

MarkerRenderer.propTypes = {
  marker: PropTypes.shape({
    position: PropTypes.shape({
      longitude: PropTypes.number,
      latitude: PropTypes.number,
    }),
    mapPinType: PropTypes.oneOf(['DEFAULT', GATEWAY, END_DEVICE]),
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
  // `panel` is a boolean used to add a class name to the
  // map container div for styling.
  panel: PropTypes.bool,
  // `widget` is a boolean used to add a class name to the map container div for styling.
  widget: PropTypes.bool,
}

export default LocationMap
