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

import Paginate from 'react-paginate'
import Icon from '../icon'

import style from './pagination.styl'

const Pagination = function (props) {
  const containerClassNames = classnames(style.pagination, props.className)
  const breakClassNames = classnames(style.item, style.itemBreak)
  const navigationNextClassNames = classnames(style.item, style.itemNavigationNext)
  const navigationPrevClassNames = classnames(style.item, style.itemNavigationPrev)

  return (
    <Paginate
      previousClassName={navigationPrevClassNames}
      previousLinkClassName={style.link}
      previousLabel={
        <Icon
          icon="navigate_before"
          small aria-label="Go to the previous page"
        />
      }
      nextClassName={navigationNextClassNames}
      nextLinkClassName={style.link}
      nextLabel={
        <Icon
          icon="navigate_next"
          small aria-label="Go to the next page"
        />
      }
      containerClassName={containerClassNames}
      pageClassName={style.item}
      breakClassName={breakClassNames}
      pageLinkClassName={style.link}
      disabledClassName={style.itemDisabled}
      activeClassName={style.itemActive}
      {...props}
    />
  )
}

Pagination.propTypes = {
  pageCount: PropTypes.number.isRequired,
  pageRangeDisplayed: PropTypes.number,
  marginPagesDisplayed: PropTypes.number,
  initialPage: PropTypes.number,
  onPageChange: PropTypes.func.isRequired,
}

Pagination.defaultProps = {
  onPageChange: () => null,
  pageRangeDisplayed: 1,
  marginPagesDisplayed: 1,
  initialPage: 0,
}

export default Pagination
