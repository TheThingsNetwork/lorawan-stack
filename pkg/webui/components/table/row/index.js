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
import bind from 'autobind-decorator'
import PropTypes from '../../../lib/prop-types'

import style from './row.styl'

@bind
class Row extends React.Component {
  onClick() {
    const { id, onClick } = this.props

    onClick(id)
  }

  onKeyDown(evt) {
    const { id, onClick } = this.props
    if (evt.key === 'Enter') {
      onClick(id)
    }
  }

  get clickListener() {
    const { body, clickable } = this.props

    if (body && clickable) {
      return this.onClick
    }
  }

  get tabIndex() {
    const { body, clickable } = this.props

    return body && clickable ? 0 : -1
  }

  render() {
    const { className, children, clickable, head, body, footer } = this.props

    const rowClassNames = classnames(className, {
      [style.clickable]: body && clickable,
      [style.rowHead]: head,
      [style.rowBody]: body,
      [style.rowFooter]: footer,
    })

    return (
      <tr
        className={rowClassNames}
        onKeyDown={this.onKeyDown}
        onClick={this.clickListener}
        tabIndex={this.tabIndex}
      >
        {children}
      </tr>
    )
  }
}

Row.propTypes = {
  /** A flag indicating whether the row is wrapping the head of a table */
  head: PropTypes.bool,
  /** A flag indicating whether the row is wrapping the body of a table */
  body: PropTypes.bool,
  /** A flag indicating whether the row is wrapping the footer of a table */
  footer: PropTypes.bool,
  /** A flag indicating whether the row is clickable */
  clickable: PropTypes.bool,
  /** The idenntifier of the row */
  id: PropTypes.number,
  /**
   * Function to be called when the row gets clicked. The identifier of the row
   * is passed as an argument.
   */
  onClick: PropTypes.func,
}

Row.defaultProps = {
  clickable: true,
  head: false,
  body: true,
  footer: false,
}

export default Row
