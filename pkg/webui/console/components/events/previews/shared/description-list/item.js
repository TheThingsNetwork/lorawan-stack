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
import classnames from 'classnames'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'

import style from './description-list.styl'

const DescriptionListItem = props => {
  const { className, children, data, title } = props
  const content = children || data

  if (!Boolean(content)) {
    return null
  }

  if (!Boolean(title)) {
    return (
      <div className={classnames(className, style.container)}>
        <div className={style.value}>{content}</div>
      </div>
    )
  }

  return (
    <div className={classnames(className, style.container)}>
      <dt className={style.term}>
        <Message content={title} />
      </dt>
      <dd className={style.value}>{content}</dd>
    </div>
  )
}

DescriptionListItem.propTypes = {
  children: PropTypes.oneOfType([PropTypes.arrayOf(PropTypes.node), PropTypes.node]),
  className: PropTypes.string,
  data: PropTypes.oneOfType([PropTypes.string, PropTypes.number]),
  title: PropTypes.message,
}

DescriptionListItem.defaultProps = {
  children: undefined,
  data: undefined,
  className: undefined,
  title: undefined,
}

export default DescriptionListItem
