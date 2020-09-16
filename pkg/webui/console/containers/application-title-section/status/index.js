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

import Status from '@ttn-lw/components/status'

import DateTime from '@ttn-lw/lib/components/date-time'
import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import style from './status.styl'

const ApplicationStatus = React.memo(props => {
  const { linked, lastSeen } = props

  const linkStatus = linked ? 'good' : 'bad'
  const linkLabel = linked ? sharedMessages.linked : sharedMessages.notLinked

  let linkElement
  if (linked && lastSeen) {
    linkElement = (
      <Status className={style.status} status={linkStatus} flipped>
        <Message content={sharedMessages.lastSeen} /> <DateTime.Relative value={lastSeen} />
      </Status>
    )
  }

  if (!linkElement) {
    linkElement = <Status className={style.status} label={linkLabel} status={linkStatus} flipped />
  }

  return linkElement
})

ApplicationStatus.propTypes = {
  lastSeen: PropTypes.string,
  linked: PropTypes.bool.isRequired,
}

ApplicationStatus.defaultProps = {
  lastSeen: undefined,
}

export default ApplicationStatus
