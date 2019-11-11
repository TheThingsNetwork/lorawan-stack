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

import Icon from '../../../components/icon'
import Message from '../../../lib/components/message'

import PropTypes from '../../../lib/prop-types'

import style from './key-value-tag.styl'

const KeyValueTag = ({ icon, keyMessage, value }) => {
  return (
    <div className={style.container}>
      <Icon className={style.icon} icon={icon} nudgeUp />
      <React.Fragment>
        <span className={style.value}>{value} </span>
        <Message
          content={keyMessage}
          values={{ count: typeof value === 'number' ? value : undefined }}
        />
      </React.Fragment>
    </div>
  )
}

KeyValueTag.propTypes = {
  icon: PropTypes.string.isRequired,
  keyMessage: PropTypes.message.isRequired,
  value: PropTypes.node.isRequired,
}

export default KeyValueTag
