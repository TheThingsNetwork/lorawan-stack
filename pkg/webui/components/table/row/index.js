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

import React, { useCallback } from 'react'
import classnames from 'classnames'

import Link from '@ttn-lw/components/link'

import PropTypes from '@ttn-lw/lib/prop-types'

import style from './row.styl'

const Row = ({
  id,
  onClick,
  onMouseDown,
  body,
  clickable,
  className,
  children,
  head,
  footer,
  linkTo,
  linkState,
}) => {
  const handleClick = useCallback(
    evt => {
      onClick(id, evt)
    },
    [id, onClick],
  )

  const onKeyDown = useCallback(
    evt => {
      if (evt.key === 'Enter') {
        onClick(id, evt)
      }
    },
    [id, onClick],
  )

  const handleMouseDown = useCallback(
    evt => {
      onMouseDown(id, evt)
    },
    [id, onMouseDown],
  )

  const clickListener = body && clickable ? handleClick : undefined

  const tabIndex = body && clickable ? 0 : -1

  const rowClassNames = classnames(className, style.row, {
    [style.clickable]: body && clickable,
    [style.rowHead]: head,
    [style.rowBody]: body,
    [style.rowFooter]: footer,
  })

  const Row = linkTo && clickable ? Link : 'div'

  return (
    <Row
      className={rowClassNames}
      onKeyDown={onKeyDown}
      onClick={clickListener}
      onMouseDown={handleMouseDown}
      tabIndex={tabIndex.toString()}
      to={linkTo}
      role="row"
      state={linkState}
    >
      {children}
    </Row>
  )
}

Row.propTypes = {
  /** A flag indicating whether the row is wrapping the body of a table. */
  body: PropTypes.bool,
  children: PropTypes.node,
  className: PropTypes.string,
  /** A flag indicating whether the row is clickable. */
  clickable: PropTypes.bool,
  /** A flag indicating whether the row is wrapping the footer of a table. */
  footer: PropTypes.bool,
  /** A flag indicating whether the row is wrapping the head of a table. */
  head: PropTypes.bool,
  /** The identifier of the row. */
  id: PropTypes.number,
  /** The state to be passed as `state` prop to the `<Link />` component that wraps the row. */
  linkState: PropTypes.shape({}),
  /** The href to be passed as `to` prop to the `<Link />` component that wraps the row. */
  linkTo: PropTypes.string,
  /**
   * Function to be called when the row gets clicked. The identifier of the row
   * is passed as an argument.
   */
  onClick: PropTypes.func,
  onMouseDown: PropTypes.func,
}

Row.defaultProps = {
  children: undefined,
  className: undefined,
  clickable: true,
  head: false,
  body: false,
  footer: false,
  onClick: () => null,
  onMouseDown: () => null,
  id: undefined,
  linkTo: undefined,
  linkState: undefined,
}

export default Row
