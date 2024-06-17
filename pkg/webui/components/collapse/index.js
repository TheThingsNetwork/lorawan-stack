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
import { defineMessages } from 'react-intl'

import Button from '@ttn-lw/components/button'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'
import sharedMessages from '@ttn-lw/lib/shared-messages'

import style from './collapse.styl'

const m = defineMessages({
  collapse: 'Collapse',
})

const Collapse = props => {
  const { className, id, title, description, initialCollapsed, children, disabled } = props

  const [collapsed, setCollapsed] = React.useState(initialCollapsed)
  const onCollapsedChange = React.useCallback(() => {
    setCollapsed(collapsed => !collapsed)
  }, [])

  const isOpen = !collapsed && !disabled

  const cls = classnames(className, style.section)
  return (
    <section className={cls} data-test-id={id}>
      <div className={style.header}>
        <div className={style.headerContent}>
          <Message className={style.title} component="h3" content={title} />
          <Message className={style.description} component="p" content={description} />
        </div>
        <div className={style.button}>
          <Button
            type="button"
            secondary
            className={style.expandButton}
            disabled={disabled}
            message={collapsed ? sharedMessages.expand : m.collapse}
            onClick={onCollapsedChange}
          />
        </div>
      </div>
      {isOpen && <div className={style.content}>{children}</div>}
    </section>
  )
}

Collapse.propTypes = {
  children: PropTypes.node.isRequired,
  className: PropTypes.string,
  description: PropTypes.message.isRequired,
  disabled: PropTypes.bool,
  id: PropTypes.string,
  initialCollapsed: PropTypes.bool,
  title: PropTypes.message.isRequired,
}

Collapse.defaultProps = {
  className: undefined,
  disabled: false,
  id: 'collapsible-section',
  initialCollapsed: true,
}

export default Collapse
