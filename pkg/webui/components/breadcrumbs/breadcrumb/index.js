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

import { Link } from 'react-router-dom'
import Message from '../../../lib/components/message'
import PropTypes from '../../../lib/prop-types'

import style from './breadcrumb.styl'

const Breadcrumb = function({ className, path, content, isLast }) {
  const isRawText = typeof content === 'string' || typeof content === 'number'
  let Component
  let componentProps
  if (!isLast) {
    Component = Link
    componentProps = { className: classnames(className, style.link), to: path }
  } else {
    Component = 'span'
    componentProps = { className: classnames(className, style.last) }
  }

  return (
    <Component {...componentProps}>
      {isRawText ? <span>{content}</span> : <Message content={content} />}
    </Component>
  )
}

Breadcrumb.propTypes = {
  className: PropTypes.string,
  /** The content of the breadcrumb */
  content: PropTypes.message.isRequired,
  /** The flag for rendering last breadcrumb as plain text */
  isLast: PropTypes.bool,
  /** The path for a breadcrumb */
  path: PropTypes.string.isRequired,
}

Breadcrumb.defaultProps = {
  className: undefined,
  isLast: false,
}

export default Breadcrumb
