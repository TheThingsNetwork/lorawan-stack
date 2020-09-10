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

import LocationMap from '@ttn-lw/components/map'
import WidgetContainer from '@ttn-lw/components/widget-container'

import Message from '@ttn-lw/lib/components/message'

import sharedMessages from '@ttn-lw/lib/shared-messages'

import style from './widget.styl'

export default class MapWidget extends React.Component {
  static propTypes = {
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

  get Map() {
    const { id, markers } = this.props

    const leafletConfig = {
      zoomControl: false,
      zoom: 10,
      minZoom: 1,
    }
    const mapCenter =
      Boolean(markers) && markers.length !== 0
        ? [markers[0].position.latitude, markers[0].position.longitude]
        : undefined

    return markers.length > 0 ? (
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
    )
  }

  render() {
    const { path } = this.props

    return (
      <WidgetContainer
        title={sharedMessages.location}
        toAllUrl={path}
        linkMessage={sharedMessages.changeLocation}
      >
        {this.Map}
      </WidgetContainer>
    )
  }
}
