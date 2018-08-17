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
import classnames from 'classnames'

import PropTypes from '../../lib/prop-types'
import Message from '../../lib/components/message'
import Icon from '../icon'
import Tab from './tab'

import style from './tabs.styl'

const Tabs = function ({
  className,
  active,
  tabs,
  onTabChange,
}) {
  return (
    <ul className={classnames(className, style.tabs)}>
      {tabs.map(function (tab, index) {
        const {
          disabled = false,
          title,
          name,
          icon = null,
        } = tab

        return (
          <Tab
            key={index}
            tabIndex={index + 1}
            isActive={name === active}
            isDisabled={disabled}
            onClick={function () {
              if (!disabled) {
                onTabChange(name)
              }
            }}
          >
            {icon && <Icon icon={icon} className={style.icon} />}
            <Message content={title} />
          </Tab>
        )
      })}
    </ul>
  )
}

Tabs.propTypes = {
  /** The name of the active tab */
  active: PropTypes.string.isRequired,
  /** List of tabs */
  tabs: PropTypes.arrayOf(PropTypes.shape({
    title: PropTypes.message.isRequired,
    icon: PropTypes.string,
    disabled: PropTypes.bool,
  })).isRequired,
  /**
   * Function to be called when the selected tab changes. Passes
   * index of the tab as an argument.
   */
  onTabChange: PropTypes.func.isRequired,
}

export default Tabs
