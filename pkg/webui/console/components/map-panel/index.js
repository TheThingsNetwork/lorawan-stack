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
import classnames from 'classnames'

import { Map } from '@ttn-lw/components/map/widget'
import { IconMapPin } from '@ttn-lw/components/icon'
import Panel from '@ttn-lw/components/panel'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import style from './map-panel.styl'

const MapPanel = ({ panelTitle, markers, entity, locationLink, centerOnMarkers, className }) => (
  <Panel
    title={panelTitle}
    icon={IconMapPin}
    className={classnames(style.panel, className)}
    shortCutLinkTitle={locationLink ? sharedMessages.map : undefined}
    shortCutLinkPath={locationLink ?? undefined}
  >
    <div className={style.content}>
      <Map
        id={`${entity}-map-widget`}
        markers={markers}
        centerOnMarkers={centerOnMarkers}
        setupLocationLink={locationLink}
      />
    </div>
  </Panel>
)

MapPanel.propTypes = {
  centerOnMarkers: PropTypes.bool,
  className: PropTypes.string,
  entity: PropTypes.string.isRequired,
  locationLink: PropTypes.string,
  markers: PropTypes.array.isRequired,
  panelTitle: PropTypes.message.isRequired,
}

MapPanel.defaultProps = {
  centerOnMarkers: false,
  locationLink: undefined,
  className: undefined,
}

export default MapPanel
