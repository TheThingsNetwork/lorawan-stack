// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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
import classNames from 'classnames'

import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'
import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'

import style from './page-title.styl'

const PageTitle = ({ title, values, tall, hideHeading, children, className, colProps, noGrid }) => {
  const titleClass = classNames(style.title, className, { [style.tall]: tall })
  const pageTitle = <IntlHelmet title={title} values={values} />

  if (noGrid) {
    return (
      <>
        {pageTitle}
        {!hideHeading && (
          <Message component="h1" className={titleClass} content={title} values={values} />
        )}
        {children}
      </>
    )
  }

  return hideHeading ? (
    pageTitle
  ) : (
    <div {...colProps} className={classNames(colProps?.className, 'item-12')}>
      {pageTitle}
      {!hideHeading && (
        <Message component="h1" className={titleClass} content={title} values={values} />
      )}
      {children}
    </div>
  )
}

PageTitle.propTypes = {
  children: PropTypes.node,
  className: PropTypes.string,
  colProps: PropTypes.shape({
    className: PropTypes.string,
  }),
  hideHeading: PropTypes.bool,
  noGrid: PropTypes.bool,
  rowProps: PropTypes.shape({}),
  tall: PropTypes.bool,
  title: PropTypes.message.isRequired,
  values: PropTypes.shape({}),
}

PageTitle.defaultProps = {
  children: null,
  className: undefined,
  colProps: {},
  hideHeading: false,
  noGrid: false,
  rowProps: {},
  tall: false,
  values: undefined,
}

export default PageTitle
