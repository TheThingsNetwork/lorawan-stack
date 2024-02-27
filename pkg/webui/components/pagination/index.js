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

import React, { useCallback, useState } from 'react'
import PropTypes from 'prop-types'
import classnames from 'classnames'
import Paginate from 'react-paginate'
import { defineMessages } from 'react-intl'

import Icon from '@ttn-lw/components/icon'

import Message from '@ttn-lw/lib/components/message'

import Select from '../select'
import Input from '../input'
import Button from '../button'

import style from './pagination.styl'

const m = defineMessages({
  itemsPerPage: 'Items per page:',
  goToPage: 'Go to page:',
})

const Pagination = ({
  onPageChange,
  className,
  forcePage,
  pageRangeDisplayed,
  marginPagesDisplayed,
  hideIfOnlyOnePage,
  pageCount,
  pageSize,
  setPageSize,
  ...rest
}) => {
  const [selectedPage, setSelectedPage] = useState(forcePage)

  const handlePageChange = useCallback(
    page => {
      if (!page.selected) {
        return onPageChange(selectedPage)
      }

      setSelectedPage(page.selected + 1)
      onPageChange(page.selected + 1)
    },
    [onPageChange, selectedPage],
  )

  const handlePageInputChange = useCallback(
    val => {
      setSelectedPage(val)
    },
    [setSelectedPage],
  )

  const handlePageSizeChange = useCallback(
    val => {
      setPageSize(val)
    },
    [setPageSize],
  )

  // Don't show pagination if there is only one page.
  if (hideIfOnlyOnePage && pageCount === 1) {
    return null
  }

  const containerClassNames = classnames(style.pagination, className)
  const breakClassNames = classnames(style.item, style.itemBreak)
  const navigationNextClassNames = classnames(style.item, style.itemNavigationNext)
  const navigationPrevClassNames = classnames(style.item, style.itemNavigationPrev)

  return (
    <div className="d-flex al-center gap-cs-l w-full">
      <Paginate
        previousClassName={navigationPrevClassNames}
        previousLinkClassName={style.link}
        previousLabel={<Icon icon="chevron_left" nudgeUp aria-label="Go to the previous page" />}
        nextClassName={navigationNextClassNames}
        nextLinkClassName={style.link}
        nextLabel={<Icon icon="chevron_right" nudgeUp aria-label="Go to the next page" />}
        containerClassName={containerClassNames}
        pageClassName={style.item}
        breakClassName={breakClassNames}
        pageLinkClassName={style.link}
        disabledClassName={style.itemDisabled}
        activeClassName={style.itemActive}
        forcePage={selectedPage - 1}
        pageRangeDisplayed={pageRangeDisplayed}
        marginPagesDisplayed={marginPagesDisplayed}
        onPageChange={handlePageChange}
        pageCount={pageCount}
        {...rest}
      />
      <div className="d-flex al-center gap-cs-xs">
        <Message content={m.itemsPerPage} className={style.sizeMessage} />
        <Select
          options={[5, 10, 20, 30, 40, 50].map(value => ({
            value,
            label: `${value}`,
          }))}
          value={pageSize}
          onChange={handlePageSizeChange}
          inputWidth="xxs"
          className={style.selectSize}
        />
      </div>
      <div className="d-flex al-center gap-cs-xs">
        <Message content={m.goToPage} className="c-text-neutral-semilight" />
        <Input
          min={1}
          max={pageCount}
          onChange={handlePageInputChange}
          inputWidth="3xs"
          placeholder="1"
          value={selectedPage}
        />
        <Button message="Go" onClick={handlePageChange} secondary />
      </div>
    </div>
  )
}

Pagination.propTypes = {
  className: PropTypes.string,
  /** Page to be displayed immediately. */
  forcePage: PropTypes.number,
  /** A flag indicating whether the pagination should be hidden when there is
   * only one page.
   */
  hideIfOnlyOnePage: PropTypes.bool,
  /**
   * The number of pages to be displayed in the beginning/end of
   * the component. For example, marginPagesDisplayed = 2, then the
   * component will display at most two pages as margins:
   * [<][1][2]...[10]...[19][20][>].
   *
   */
  marginPagesDisplayed: PropTypes.number,
  /** An onClick handler that gets called with the new page number. */
  onPageChange: PropTypes.func,
  /** The total number of pages. */
  pageCount: PropTypes.number.isRequired,
  /**
   * The number of pages to be displayed. If is bigger than
   * pageCount, then all pages will be displayed without gaps.
   */
  pageRangeDisplayed: PropTypes.number,
  /** The number of items per page. */
  pageSize: PropTypes.number,
  /** A function to be called when the page size changes. */
  setPageSize: PropTypes.func,
}

Pagination.defaultProps = {
  className: undefined,
  forcePage: 1,
  hideIfOnlyOnePage: true,
  marginPagesDisplayed: 1,
  onPageChange: () => null,
  pageRangeDisplayed: 1,
  pageSize: 10,
  setPageSize: () => null,
}

export default Pagination
