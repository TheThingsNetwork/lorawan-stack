// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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
import classNames from 'classnames'

import Icon from '@ttn-lw/components/icon'

import PropTypes from '@ttn-lw/lib/prop-types'

import style from './badge.styl'

const Badge = ({ className, children, status, startIcon, endIcon }) => (
  <div
    className={classNames(style.badge, className, style[status], {
      [style.startIcon]: !!startIcon,
      [style.endIcon]: !!endIcon,
    })}
  >
    {startIcon && <Icon icon={startIcon} />}
    {children}
    {endIcon && <Icon icon={endIcon} />}
  </div>
)

Badge.propTypes = {
  children: PropTypes.node.isRequired,
  className: PropTypes.string,
  endIcon: PropTypes.icon,
  startIcon: PropTypes.icon,
  status: PropTypes.oneOf(['success', 'error', 'info', 'warning']).isRequired,
}

Badge.defaultProps = {
  startIcon: undefined,
  endIcon: undefined,
  className: undefined,
}
export default Badge
