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

import Icon, { IconChevronDown, IconChevronUp } from '@ttn-lw/components/icon'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'

import { useFormContext } from '..'

import style from './section.styl'

const m = defineMessages({
  expand: 'Expand section',
  collapse: 'Collapse section',
})

// `<FormCollapseSection />` aggregates a set of form fields under common title as well as adds
// functionality to hide or show the fields.
const FormCollapseSection = props => {
  const { className, id, title, onCollapse, isCollapsed, initiallyCollapsed, children } = props
  const { formatMessage } = useIntl()
  const { disabled } = useFormContext()

  // Check if the component is 'controlled'. When the `isCollapsed` prop is passed the component
  // is considered as 'uncontrolled' and it's state must be controlled from the outside.
  const isControlled = typeof isCollapsed === 'undefined'

  const [collapsed, setCollapsed] = React.useState(initiallyCollapsed)
  const onExpandedChange = React.useCallback(() => {
    if (isControlled) {
      setCollapsed(collapsed => !collapsed)
    } else {
      onCollapse()
    }
  }, [isControlled, onCollapse])

  const isSectionClosed = isControlled ? collapsed : isCollapsed

  return (
    <div className={className}>
      <button
        className={style.button}
        type="button"
        onClick={onExpandedChange}
        aria-label={isSectionClosed ? formatMessage(m.expand) : formatMessage(m.collapse)}
        aria-expanded={!isSectionClosed}
        aria-controls={id}
        disabled={disabled}
      >
        <Message content={title} className={style.title} />
        <Icon className={style.icon} icon={isSectionClosed ? IconChevronDown : IconChevronUp} />
      </button>
      <div className={classnames(style.content, { [style.expanded]: !isSectionClosed })} id={id}>
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
  isCollapsed: PropTypes.bool,
  onCollapse: PropTypes.func,
  title: PropTypes.message.isRequired,
}

FormCollapseSection.defaultProps = {
  className: undefined,
  onCollapse: () => null,
  isCollapsed: undefined,
  initiallyCollapsed: true,
}

export default FormCollapseSection
