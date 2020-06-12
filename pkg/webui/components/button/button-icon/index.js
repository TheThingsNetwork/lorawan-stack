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

import PropTypes from '@ttn-lw/lib/prop-types'

import style from './icon.styl'

const ButtonIcon = props => {
  const { className, icon, type, ...rest } = props

  const cls = classnames(className, style.icon, {
    [style.left]: type === 'left',
    [style.right]: type === 'right',
  })

  return <Icon className={cls} icon={icon} nudgeUp {...rest} />
}

ButtonIcon.propTypes = {
  className: PropTypes.string,
  icon: PropTypes.string.isRequired,
  type: PropTypes.oneOf(['left', 'right']),
}

ButtonIcon.defaultProps = {
  className: undefined,
  type: 'left',
}

export default ButtonIcon
