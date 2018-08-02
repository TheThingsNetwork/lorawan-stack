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

import NavigationLink from '../link'
import Message from '../../message'
import Icon from '../../icon'
import PropTypes from '../../../lib/prop-types'

import style from './bar.styl'

const NavigationBar = function ({
  className,
  entries,
}) {
  return (
    <nav className={classnames(className, style.bar)}>
      {entries.map(function (entry, index) {
        const {
          path,
          title,
          icon = null,
          exact = true,
        } = entry

        return (
          <NavigationLink
            key={index}
            path={path}
            exact={exact}
            className={style.link}
            activeClassName={style.linkActive}
          >
            {icon && <Icon icon={icon} className={style.icon} />}
            <Message content={title} />
          </NavigationLink>
        )
      })}
    </nav>
  )
}

NavigationBar.propTypes = {
  /**
   * A list of navigation bar entries.
   * @param {title} The title to be displayed
   * @param {icon} The icon name to be displayed next to the title
   * @param {path} The path for a navigation tab
   * @param {exact} Boolean flag identifying whether the path should
   *  be matched exactly
   */
  entries: PropTypes.arrayOf(PropTypes.shape({
    path: PropTypes.string.isRequired,
    title: PropTypes.message.isRequired,
    icon: PropTypes.string,
    exact: PropTypes.bool,
  })),
}

NavigationBar.defaultProps = {
  entries: [],
}

export default NavigationBar
