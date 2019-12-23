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
import classnames from 'classnames'
import Leaflet from 'leaflet'

import MarkerIcon from '../../assets/auxiliary-icons/location_pin.svg'

import style from './map.styl'

// Reset default marker icon
delete Leaflet.Icon.Default.prototype._getIconUrl
Leaflet.Icon.Default.mergeOptions({
  iconRetinaUrl: MarkerIcon,
  iconUrl: MarkerIcon,
  iconSize: [26, 36],
  shadowSize: [26, 36],
  iconAnchor: [13, 36],
  shadowAnchor: [8, 37],
  popupAnchor: [0, -40],
  shadowUrl: require('leaflet/dist/images/marker-shadow.png'),
})

export default class Map extends React.Component {
  static propTypes = {
    // Id is a string used to give the map a unique ID.
    id: PropTypes.string.isRequired,
    // LeafletConfig is an object which can contain any number of properties defined by the leaflet plugin and is used to overwrite the default configuration of leaflet.
    leafletConfig: PropTypes.shape({}),
    // Markers is an array of objects containing a specific properties
    markers: PropTypes.arrayOf(
      // Position is a object containing two properties latitude and longitude which are both numbers.
      PropTypes.shape({
        position: PropTypes.objectOf(PropTypes.number),
      }),
    ).isRequired,
    // Widget is a boolean used to add a class name to the map container div for styling.
    widget: PropTypes.bool,
  }

  static defaultProps = {
    leafletConfig: {},
    widget: false,
  }

  getMapCenter(markers) {
    // This will calculate zoom and map center long/lang based on all markers provided.
    // Currently it just returns the first marker.
    // TODO: action (https://github.com/TheThingsNetwork/lorawan-stack/issues/1241)
    return markers[0]
  }

  createMap(config, id) {
    this.map = Leaflet.map(id, {
      ...config,
    })
  }

  createMarkers(markers) {
    markers.map(marker =>
      Leaflet.marker([marker.position.latitude, marker.position.longitude]).addTo(this.map),
    )
  }

  componentDidMount() {
    const { id, markers } = this.props

    const { position } = markers.length >= 1 ? this.getMapCenter(markers) : markers[0]

    const config = {
      layers: [
        Leaflet.tileLayer('http://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
          attribution: '&copy; <a href="http://osm.org/copyright">OpenStreetMap</a> contributors',
        }),
      ],
      center: [position.latitude, position.longitude],
      zoom: 11,
      minZoom: 1,
      ...this.props.leafletConfig,
    }

    this.createMap(config, id)
    this.createMarkers(markers, id)
  }

  render() {
    const { id, widget } = this.props

    return (
      <div className={style.container}>
        <div className={classnames(style.map, { [style.widget]: widget })} id={id} />
      </div>
    )
  }
}
