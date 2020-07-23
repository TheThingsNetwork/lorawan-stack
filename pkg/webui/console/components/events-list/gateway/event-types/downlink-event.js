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

import Event from '@ttn-lw/components/events-list/event'

import PropTypes from '@ttn-lw/lib/prop-types'

import Entry from '../../shared/components/entries'

import style from './event-types.styl'

const { Overview, Details } = Event

const GatewayDownlinkEvent = props => {
  const { event, widget } = props
  const { name, time } = event

  const showData = 'data' in event && !widget

  return (
    <Event event={event} expandable={showData}>
      <Overview>
        <Entry.Icon className={style.icon} iconName="downlink" />
        <Entry.Time time={time} />
        <Entry.Type eventName={name} />
        <Entry.Data />
      </Overview>
      {showData && <Details />}
    </Event>
  )
}

GatewayDownlinkEvent.propTypes = {
  event: PropTypes.event.isRequired,
  widget: PropTypes.bool,
}

GatewayDownlinkEvent.defaultProps = {
  widget: false,
}

export default GatewayDownlinkEvent
