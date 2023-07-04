// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

import ServerIcon from '@assets/auxiliary-icons/server.svg'

import Status from '@ttn-lw/components/status'

import Message from '@ttn-lw/lib/components/message'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import style from './overview.styl'

const ComponentCard = ({ name, enabled, host }) => (
  <div className={style.componentCard}>
    <img src={ServerIcon} className={style.componentCardIcon} />
    <div className={style.componentCardDesc}>
      <div className={style.componentCardName}>
        <Status label={name} status={enabled ? 'good' : 'unknown'} flipped />
      </div>
      <span className={style.componentCardHost} title={host}>
        {enabled ? host : <Message content={sharedMessages.disabled} />}
      </span>
    </div>
  </div>
)

ComponentCard.propTypes = {
  enabled: PropTypes.bool.isRequired,
  host: PropTypes.string,
  name: PropTypes.message.isRequired,
}

ComponentCard.defaultProps = {
  host: undefined,
}

export default ComponentCard
