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
import { defineMessages, useIntl } from 'react-intl'

import Form, { useFormContext } from '@ttn-lw/components/form'
import Icon from '@ttn-lw/components/icon'

import PropTypes from '@ttn-lw/lib/prop-types'

import style from './section.styl'

const m = defineMessages({
  expand: 'Expand section',
  collapse: 'Collapse section',
})

// `<FormCollapseSection />` aggregates a set of form fields under common title as well as adds
// functionality to hide or show the fields.
const FormCollapseSection = props => {
  const { className, id, title, onCollapse, initiallyCollapsed, children } = props
  const { formatMessage } = useIntl()
  const { disabled } = useFormContext()

  const [collapsed, setCollapsed] = React.useState(initiallyCollapsed)
  const onExpandedChange = React.useCallback(() => {
    setCollapsed(collapsed => !collapsed)
    onCollapse(!collapsed)
  }, [collapsed, onCollapse])

  return (
    <div className={className}>
      <button
        className={style.button}
        type="button"
        onClick={onExpandedChange}
        aria-label={collapsed ? formatMessage(m.expand) : formatMessage(m.collapse)}
        aria-expanded={!collapsed}
        aria-controls={id}
        disabled={disabled}
      >
        <Form.SubTitle className={style.title} title={title} />
        <Icon className={style.icon} icon={collapsed ? 'expand_down' : 'expand_up'} />
      </button>
      <div className={classnames(style.content, { [style.expanded]: !collapsed })} id={id}>
        {children}
      </div>
    </div>
  )
}

FormCollapseSection.propTypes = {
  children: PropTypes.oneOfType([PropTypes.arrayOf(PropTypes.node), PropTypes.node]).isRequired,
  className: PropTypes.string,
  id: PropTypes.string.isRequired,
  initiallyCollapsed: PropTypes.bool,
  onCollapse: PropTypes.func,
  title: PropTypes.message.isRequired,
}

FormCollapseSection.defaultProps = {
  className: undefined,
  onCollapse: () => null,
  initiallyCollapsed: true,
}

export default FormCollapseSection
