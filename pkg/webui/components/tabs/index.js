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
import classnames from 'classnames'

import Icon from '@ttn-lw/components/icon'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'

import Tab from './tab'

import style from './tabs.styl'

const Tabs = function ({ className, active, tabs, onTabChange, divider, narrow }) {
  return (
    <ul className={classnames(className, style.tabs, { [style.divider]: divider })}>
      {tabs.map(function (
        { name, disabled, narrow: nrw, link, exact, icon, title, hidden },
        index,
      ) {
        return (
          !Boolean(hidden) && (
            <Tab
              key={index}
              active={name === active}
              name={name}
              disabled={disabled}
              onClick={onTabChange}
              narrow={nrw || narrow}
              link={link}
              exact={exact}
            >
              {icon && <Icon icon={icon} className={style.icon} />}
              <Message content={title} />
            </Tab>
          )
        )
      })}
    </ul>
  )
}

Tabs.propTypes = {
  /** The name of the active tab. */
  active: PropTypes.string,
  className: PropTypes.string,
  /** Flag specifying whether the tab should render a bottom divider. */
  divider: PropTypes.bool,
  /**
   * A click handler to be called when the selected tab changes. Passes
   * the name of the new active tab as an argument.
   */
  narrow: PropTypes.bool,
  /** A list of tabs. */
  onTabChange: PropTypes.func,
  tabs: PropTypes.arrayOf(
    PropTypes.shape({
      title: PropTypes.message.isRequired,
      name: PropTypes.string.isRequired,
      icon: PropTypes.string,
      disabled: PropTypes.bool,
      hidden: PropTypes.bool,
    }),
  ).isRequired,
}

Tabs.defaultProps = {
  active: undefined,
  className: undefined,
  onTabChange: () => null,
  divider: false,
  narrow: false,
}

export default Tabs
