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

import NavigationLink, { NavigationAnchorLink } from '../link'
import Message from '../../../lib/components/message'
import Icon from '../../icon'
import PropTypes from '../../../lib/prop-types'

import style from './bar.styl'

const NavigationBar = function({ className, entries, anchored }) {
  return (
    <nav className={classnames(className, style.bar)}>
      {entries.map(function(entry, index) {
        const { path, title, icon = null, exact = false, hidden } = entry

        if (hidden) return null

        const Component = anchored ? NavigationAnchorLink : NavigationLink

        return (
          <Component
            key={index}
            path={path}
            exact={exact}
            className={style.link}
            activeClassName={style.linkActive}
          >
            {icon && <Icon icon={icon} className={style.icon} />}
            <Message content={title} />
          </Component>
        )
      })}
    </nav>
  )
}

NavigationBar.propTypes = {
  /** Flag identifying whether links should be rendered as plain anchor link */
  anchored: PropTypes.bool,
  className: PropTypes.string,
  /**
   * A list of navigation bar entries.
   * @param {(string|Object)} title - The title to be displayed
   * @param {string} icon - The icon name to be displayed next to the title
   * @param {string} path -  The path for a navigation tab
   * @param {boolean} exact - Flag identifying whether the path should be matched exactly
   */
  entries: PropTypes.arrayOf(PropTypes.link),
}

NavigationBar.defaultProps = {
  className: undefined,
  entries: [],
  anchored: false,
}

export default NavigationBar
