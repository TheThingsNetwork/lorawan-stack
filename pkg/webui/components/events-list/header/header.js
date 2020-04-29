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
import classnames from 'classnames'

import Button from '@ttn-lw/components/button'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import eventsContext from '../context'

import style from './header.styl'

const EventsHeader = props => {
  const { className, children } = props
  const { onPause, onClear, widget, paused } = React.useContext(eventsContext)

  return (
    <div className={classnames(className, style.header)}>
      {children}
      {!widget && (
        <div className={style.actions}>
          <Button
            onClick={onPause}
            message={paused ? sharedMessages.resume : sharedMessages.pause}
            naked
            secondary
            icon={paused ? 'play_arrow' : 'pause'}
          />
          <Button onClick={onClear} message={sharedMessages.clear} naked secondary icon="delete" />
        </div>
      )}
    </div>
  )
}

EventsHeader.propTypes = {
  children: PropTypes.oneOfType([PropTypes.arrayOf(PropTypes.node), PropTypes.node]).isRequired,
  className: PropTypes.string,
}

EventsHeader.defaultProps = {
  className: undefined,
}

export default EventsHeader
