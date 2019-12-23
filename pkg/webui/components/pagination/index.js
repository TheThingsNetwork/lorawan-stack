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
import PropTypes from 'prop-types'
import classnames from 'classnames'
import bind from 'autobind-decorator'

import Paginate from 'react-paginate'
import Icon from '../icon'

import style from './pagination.styl'

@bind
class Pagination extends React.PureComponent {
  static propTypes = {
    className: PropTypes.string,
    /** Page to be displayed immediately */
    forcePage: PropTypes.number,
    /** A flag indicating whether the pagination should be hidden when there is
     * only one page.
     */
    hideIfOnlyOnePage: PropTypes.bool,
    /** The initial page number to be selected when the component
     * gets rendered for the first time. Is 1-based.
     */
    initialPage: PropTypes.number,
    /**
     * The number of pages to be displayed in the beginning/end of
     * the component. For example, marginPagesDisplayed = 2, then the
     * component will display at most two pages as margins:
     * [<][1][2]...[10]...[19][20][>]
     *
     */
    marginPagesDisplayed: PropTypes.number,
    /** An onClick handler that gets called with the new page number */
    onPageChange: PropTypes.func,
    /** The total number of pages */
    pageCount: PropTypes.number.isRequired,
    /**
     * The number of pages to be displayed. If is bigger than
     * pageCount, then all pages will be displayed without gaps.
     */
    pageRangeDisplayed: PropTypes.number,
  }

  static defaultProps = {
    className: undefined,
    forcePage: 1,
    hideIfOnlyOnePage: true,
    initialPage: 1,
    marginPagesDisplayed: 1,
    onPageChange: () => null,
    pageRangeDisplayed: 1,
  }

  onPageChange(page) {
    this.props.onPageChange(page.selected + 1)
  }

  render() {
    const {
      className,
      forcePage,
      initialPage,
      pageRangeDisplayed,
      marginPagesDisplayed,
      hideIfOnlyOnePage,
      onPageChange,
      pageCount,
      ...rest
    } = this.props

    // Don't show pagination if there is only one page
    if (hideIfOnlyOnePage && pageCount === 1) {
      return null
    }

    const containerClassNames = classnames(style.pagination, className)
    const breakClassNames = classnames(style.item, style.itemBreak)
    const navigationNextClassNames = classnames(style.item, style.itemNavigationNext)
    const navigationPrevClassNames = classnames(style.item, style.itemNavigationPrev)

    return (
      <Paginate
        previousClassName={navigationPrevClassNames}
        previousLinkClassName={style.link}
        previousLabel={<Icon icon="navigate_before" small aria-label="Go to the previous page" />}
        nextClassName={navigationNextClassNames}
        nextLinkClassName={style.link}
        nextLabel={<Icon icon="navigate_next" small aria-label="Go to the next page" />}
        containerClassName={containerClassNames}
        pageClassName={style.item}
        breakClassName={breakClassNames}
        pageLinkClassName={style.link}
        disabledClassName={style.itemDisabled}
        activeClassName={style.itemActive}
        forcePage={forcePage - 1}
        initialPage={initialPage - 1}
        pageRangeDisplayed={pageRangeDisplayed}
        marginPagesDisplayed={marginPagesDisplayed}
        onPageChange={this.onPageChange}
        pageCount={pageCount}
        {...rest}
      />
    )
  }
}

export default Pagination
