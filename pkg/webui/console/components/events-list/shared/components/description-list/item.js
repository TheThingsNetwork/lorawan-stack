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

import style from './item.styl'

const DescriptionListItem = props => {
  const { className, children, title } = props

  if (!Boolean(title)) {
    return (
      <div className={classnames(className, style.container)}>
        <div className={style.description}>{children}</div>
      </div>
    )
  }

  return (
    <div className={classnames(className, style.container)}>
      <dt className={style.term}>
        <Message content={title} />
      </dt>
      <dd className={style.description}>{children}</dd>
    </div>
  )
}

DescriptionListItem.propTypes = {
  children: PropTypes.oneOfType([PropTypes.arrayOf(PropTypes.node), PropTypes.node]).isRequired,
  className: PropTypes.string,
  title: PropTypes.message,
}

DescriptionListItem.defaultProps = {
  className: undefined,
  title: undefined,
}

export default DescriptionListItem
