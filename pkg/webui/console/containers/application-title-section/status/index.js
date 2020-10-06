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

import EntityTitleSection from '@console/components/entity-title-section'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import style from './status.styl'

const { Content } = EntityTitleSection

const ApplicationStatus = props => {
  const { linked, lastSeen } = props

  const linkStatus = linked ? 'good' : 'bad'
  const linkLabel = linked ? sharedMessages.linked : sharedMessages.notLinked

  if (linked && lastSeen) {
    return (
      <Status className={style.status} status={linkStatus} flipped>
        <Content.LastSeen lastSeen={lastSeen} />
      </Status>
    )
  }

  return <Status className={style.status} label={linkLabel} status={linkStatus} flipped />
}

ApplicationStatus.propTypes = {
  lastSeen: PropTypes.string,
  linked: PropTypes.bool.isRequired,
}

ApplicationStatus.defaultProps = {
  lastSeen: undefined,
}

export default ApplicationStatus
