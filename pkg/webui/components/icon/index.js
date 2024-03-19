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

import React, { forwardRef } from 'react'
import classnames from 'classnames'
import PropTypes from 'prop-types'

import style from './icon.styl'

const Icon = forwardRef((props, ref) => {
  const {
    icon: ActualIcon,
    className,
    nudgeUp,
    nudgeDown,
    small,
    large,
    textPaddedLeft,
    textPaddedRight,
    size,
    ...rest
  } = props

  const classname = classnames(className, {
    [style.nudgeUp]: nudgeUp,
    [style.nudgeDown]: nudgeDown,
    [style.large]: large,
    [style.small]: small,
    [style.textPaddedLeft]: textPaddedLeft,
    [style.textPaddedRight]: textPaddedRight,
  })

  return <ActualIcon className={classname} ref={ref} size={small ? 16 : size} {...rest} />
})

Icon.propTypes = {
  className: PropTypes.string,
  /** Which icon to display, using tabler icon set. */
  icon: PropTypes.shape({}).isRequired,
  /** Renders a bigger icon. */
  large: PropTypes.bool,
  /** Nudges the icon down by one pixel using position: relative. */
  nudgeDown: PropTypes.bool,
  /** Nudges the icon up by one pixel using position: relative. */
  nudgeUp: PropTypes.bool,
  /** The size of the icon. */
  size: PropTypes.number,
  /** Renders a smaller icon. */
  small: PropTypes.bool,
  /** Whether icon should be padded for a text displayed left to it. */
  textPaddedLeft: PropTypes.bool,
  /** Whether icon should be padded for a text displayed right to it. */
  textPaddedRight: PropTypes.bool,
}

Icon.defaultProps = {
  className: undefined,
  large: false,
  nudgeDown: false,
  nudgeUp: false,
  size: 20,
  small: false,
  textPaddedLeft: false,
  textPaddedRight: false,
}

export default Icon
export * from '@tabler/icons-react'
export * from './common'
