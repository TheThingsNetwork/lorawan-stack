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

import Entry from '../entries'

import style from './error-event.styl'

const { Overview, Details } = Event

const ErrorEvent = props => {
  const { event, entityId, widget } = props
  const { data, name, time } = event

  return (
    <Event event={event} expandable={!widget}>
      <Overview>
        <Entry.Icon className={style.icon} icon="error" />
        <Entry.Time time={time} />
        {entityId && <Entry.ID entityId={entityId} />}
        <Entry.Type className={style.type} eventName={name} />
        {!widget && <Entry.ErrorDescription errorDetails={data} />}
      </Overview>
      <Details />
    </Event>
  )
}

ErrorEvent.propTypes = {
  entityId: PropTypes.string,
  event: PropTypes.event.isRequired,
  widget: PropTypes.bool,
}

ErrorEvent.defaultProps = {
  entityId: undefined,
  widget: false,
}

export default ErrorEvent
