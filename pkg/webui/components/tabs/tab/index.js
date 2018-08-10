// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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
import PropTypes from 'prop-types'
import classnames from 'classnames'

import style from './tab.styl'

const Tab = function ({
  className,
  onClick,
  isActive = false,
  isDisabled = false,
  children,
  ...rest
}) {
  return (
    <li
      {...rest}
      role="button"
      onClick={onClick}
      className={classnames(className, style.tab, {
        [style.tabActive]: !isDisabled && isActive,
        [style.tabDefault]: !isDisabled && !isActive,
        [style.tabDisabled]: isDisabled,
      })}
    >
      {children}
    </li>
  )
}

Tab.propTypes = {
  /** Function to be called when the tab gets clicked */
  onClick: PropTypes.func.isRequired,
  /** Boolean flag identifying whether the tab is active */
  isActive: PropTypes.bool,
  /** Boolean flag identifying whether the tab is disabled */
  isDisabled: PropTypes.bool,
}

export default Tab
