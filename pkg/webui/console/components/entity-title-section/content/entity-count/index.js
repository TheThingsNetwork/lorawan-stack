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

import Icon from '@ttn-lw/components/icon'
import Link from '@ttn-lw/components/link'

import Message from '@ttn-lw/lib/components/message'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import style from './entity-count.styl'

const EntityCount = props => {
  const { icon, keyMessage, value, errored, toAllUrl } = props

  return (
    <Link
      className={classnames(style.container, {
        [style.error]: errored,
      })}
      to={toAllUrl}
    >
      <Icon className={style.icon} icon={icon} nudgeUp />
      <span className={style.value}>{value} </span>
      <Message
        className={style.message}
        content={errored ? sharedMessages.notAvailable : keyMessage}
        values={{
          count: errored ? undefined : value,
        }}
      />
    </Link>
  )
}

EntityCount.propTypes = {
  errored: PropTypes.bool,
  icon: PropTypes.string.isRequired,
  keyMessage: PropTypes.message.isRequired,
  toAllUrl: PropTypes.string.isRequired,
  value: PropTypes.node,
}

EntityCount.defaultProps = {
  errored: false,
  value: undefined,
}

export default EntityCount
