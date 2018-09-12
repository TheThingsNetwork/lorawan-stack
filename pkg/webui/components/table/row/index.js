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
import PropTypes from '../../../lib/prop-types'

import style from './row.styl'

const Row = function ({
  className,
  children,
  head = false,
  body = true,
  footer = false,
  onClick,
}) {
  const clickable = !!onClick
  const rowClassNames = classnames(className, {
    [style.clickable]: body && clickable,
    [style.rowHead]: head,
    [style.rowBody]: body,
    [style.rowFooter]: footer,
  })

  return (
    <tr
      className={rowClassNames}
      onClick={body && clickable ? onClick : undefined}
      tabIndex={body && clickable ? 0 : -1}
    >
      {children}
    </tr>
  )
}

Row.propTypes = {
  /** A flag indicating whether the row is wrapping the head of a table */
  head: PropTypes.bool,
  /** A flag indicating whether the row is wrapping the body of a table */
  body: PropTypes.bool,
  /** A flag indicating whether the row is wrapping the footer of a table */
  footer: PropTypes.bool,
  /** Function to be called when the row gets clicked */
  onClick: PropTypes.func,
}

export default Row
