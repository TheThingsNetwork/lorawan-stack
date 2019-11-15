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
import { Col, Row } from 'react-grid-system'
import classnames from 'classnames'

import PropTypes from '../../lib/prop-types'
import IntlHelmet from '../../lib/components/intl-helmet'
import Message from '../../lib/components/message'

import style from './page-title.styl'

const PageTitle = ({ title, values, tall, className, hideHeading, children }) => {
  const containerClass = classnames(className, style.container, {
    [style.hideHeading]: hideHeading,
  })
  const titleClass = classnames(style.title, { [style.tall]: tall })
  return (
    <Row className={containerClass}>
      <Col>
        <IntlHelmet title={title} values={values} />
        {!hideHeading && (
          <Message component="h1" className={titleClass} content={title} values={values} />
        )}
        {children}
      </Col>
    </Row>
  )
}

PageTitle.propTypes = {
  children: PropTypes.node,
  className: PropTypes.string,
  hideHeading: PropTypes.bool,
  tall: PropTypes.bool,
  title: PropTypes.message.isRequired,
  values: PropTypes.shape({}),
}

PageTitle.defaultProps = {
  children: null,
  className: '',
  hideHeading: false,
  tall: false,
  values: undefined,
}

export default PageTitle
