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

import Link from '@ttn-lw/components/link'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'

import style from './widget-container.styl'

const WidgetContainer = ({ children, title, toAllUrl, linkMessage, className }) => (
  <aside className={classnames(style.container, className)}>
    <div className={style.header}>
      {typeof title === 'object' && 'id' in title ? (
        <Message content={title} className={style.headerTitle} />
      ) : (
        title
      )}
      <Link className={style.seeAllLink} secondary to={toAllUrl}>
        <Message content={linkMessage} /> →
      </Link>
    </div>
    <div className={style.body}>
      <Link className={style.bodyLink} to={toAllUrl}>
        {children}
      </Link>
    </div>
  </aside>
)

WidgetContainer.propTypes = {
  children: PropTypes.node.isRequired,
  className: PropTypes.string,
  linkMessage: PropTypes.message.isRequired,
  title: PropTypes.oneOfType([PropTypes.message, PropTypes.node]).isRequired,
  toAllUrl: PropTypes.string.isRequired,
}

WidgetContainer.defaultProps = {
  className: undefined,
}

export default WidgetContainer
