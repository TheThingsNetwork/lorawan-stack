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

import React, { Fragment } from 'react'
import classnames from 'classnames'

import PropTypes from '../../lib/prop-types'
import Message from '../message'
import Icon from '../icon'
import Breacrumb from './breadcrumb'

import style from './breadcrumbs.styl'

const Breadcrumbs = function ({
  className,
  entries,
}) {
  return (
    <nav className={classnames(className, style.breadcrumbs)}>
      {entries.map(function (entry, index, arr) {
        const {
          title,
          icon = null,
          path,
        } = entry

        return (
          <Fragment key={index}>
            <Breacrumb
              path={path}
            >
              {icon && <Icon icon={icon} className={style.icon} />}
              <Message content={title} />
            </Breacrumb>
          </Fragment>
        )
      })}
    </nav>
  )
}

Breadcrumbs.propTypes = {
  /**
   * A list of breadcrumb entries.
   * @param {(string|Object)} title - The title to be displayed
   * @param {string} title - The icon name to be displayed next to the title
   * @param {string} path - The path for a breadcrumb
   */
  entries: PropTypes.arrayOf(PropTypes.shape({
    title: PropTypes.message.isRequired,
    path: PropTypes.string.isRequired,
    icon: PropTypes.string,
  })),
}

Breadcrumbs.defaultProps = {
  entries: [],
}

export default Breadcrumbs
