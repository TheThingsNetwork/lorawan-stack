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
import sharedMessages from '../../../lib/shared-messages'

import Message from '../../../lib/components/message'
import Link from '../../../components/link'
import Map from '../../map'

import style from './widget.styl'

export default class MapWidget extends React.Component {
  get Map() {
    const { id, markers } = this.props

    const leafletConfig = {
      zoomControl: false,
    }

    return markers.length > 0 ? (
      <Map id={id} markers={markers} leafletConfig={leafletConfig} widget />
    ) : (
      <div className={style.mapDisabled}>
        <Message component="span" content={sharedMessages.noLocation} />
      </div>
    )
  }

  render() {
    const { path } = this.props

    return (
      <aside className={style.wrapper}>
        <div className={style.header}>
          <Message className={style.titleMessage} content={sharedMessages.location} />
          <Link to={path}>
            <Message
              className={style.changeLocationMessage}
              content={sharedMessages.changeLocation}
            />
            →
          </Link>
        </div>
        {this.Map}
      </aside>
    )
  }
}

MapWidget.propTypes = {
  // Id is a string used to give the map a unique ID.
  id: PropTypes.string.isRequired,
  // Markers is an array of objects containing a specific properties
  markers: PropTypes.arrayOf(
    // Position is a object containing two properties latitude and longitude which are both numbers.
    PropTypes.shape({
      position: PropTypes.objectOf(PropTypes.number),
    }),
  ).isRequired,
  // Path is a string that is required to show the link at location form.
  path: PropTypes.string.isRequired,
}
