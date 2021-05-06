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

import DateTime from '@ttn-lw/lib/components/date-time'
import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import style from './last-seen.styl'

const computeDeltaInSeconds = (from, to) => {
  // Avoid situations when server clock is ahead of the browser clock.
  if (from > to) {
    return 0
  }

  return Math.floor((from - to) / 1000)
}

const LastSeen = props => {
  const { className, lastSeen, short } = props

  return (
    <div className={classnames(className, style.container)}>
      {!short && <Message className={style.message} content={sharedMessages.lastSeen} />}
      <DateTime.Relative value={lastSeen} computeDelta={computeDeltaInSeconds} />
    </div>
  )
}

LastSeen.propTypes = {
  className: PropTypes.string,
  lastSeen: PropTypes.oneOfType([
    PropTypes.string,
    PropTypes.number, // Support timestamps.
    PropTypes.instanceOf(Date),
  ]).isRequired,
  short: PropTypes.bool,
}

LastSeen.defaultProps = {
  className: undefined,
  short: false,
}

export default LastSeen
