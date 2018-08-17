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
import orders from '../orders'
import PropTypes from '../../../lib/prop-types'

import Message from '../../../lib/components/message'
import Icon from '../../icon'

import style from './header-cell.styl'

const HeaderCell = function ({
  content,
  name,
  order,
  active,
  centered = false,
  className,
  sortable = false,
  onSort = () => 0,
}) {
  const cellClassNames = classnames(className, style.cell, {
    [style.cellCentered]: centered,
    [style.cellSortable]: sortable,
  })

  if (sortable) {
    const sortButtonClassNames = classnames(style.sortButton, {
      [style.sortButtonActive]: active,
    })

    return (
      <th className={cellClassNames}>
        <button
          type="button"
          className={sortButtonClassNames}
          onClick={function () {
            onSort(name)
          }}
          data-hook="sort-button"
        >
          <Message content={content} />
          <SortIcon
            ascending={active && order === orders.ASCENDING}
          />
        </button>
      </th>
    )
  }

  return (
    <th className={cellClassNames}>
      <Message content={content} />
    </th>
  )
}

const SortIcon = function ({ ascending }) {
  const iconClassNames = classnames(style.sortIcon, {
    [style.sortIconAsc]: ascending,
  })

  return (
    <Icon className={iconClassNames} icon="arrow_drop_down" />
  )
}

HeaderCell.propTypes = {
  /** The name of the header */
  name: PropTypes.string.isRequired,
  /** The content to be displayed in the cell */
  content: PropTypes.message.isRequired,
  /** A flag specifying whether the column under the cell should be centered */
  centered: PropTypes.bool,
  /** A flag specifying whether the cell should have a button to sort the table */
  sortable: PropTypes.bool,
  /** Function to be called when the sortable cell is clicked */
  onSort: PropTypes.func,
}

export default HeaderCell
