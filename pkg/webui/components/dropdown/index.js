// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

import React, { useEffect, useRef } from 'react'
import classnames from 'classnames'

import Icon, { IconChevronRight } from '@ttn-lw/components/icon'
import Link from '@ttn-lw/components/link'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'

import AttachedDropdown from './attached'

import style from './dropdown.styl'

const Dropdown = ({
  className,
  children,
  larger,
  onItemsClick,
  onOutsideClick,
  open,
  position,
  hover,
}) => {
  const ref = useRef(null)
  // Attach event listeners to the document to close the dropdown when clicking outside of it.
  useEffect(() => {
    const handleClickOutside = e => {
      if (ref.current && !ref.current.contains(e.target)) {
        onOutsideClick(e)
      }
    }

    document.addEventListener('mousedown', handleClickOutside)

    return () => {
      document.removeEventListener('mousedown', handleClickOutside)
    }
  }, [onOutsideClick])

  if (!open) {
    return null
  }

  const pos = position.split(' ')
  // Based on the order of the position string, we can determine the vertical and horizontal
  // placement of the dropdown, which will be different for eg. 'below right' and 'right below'.
  // The first part of the position string determines the primary axis of the dropdown, which
  // also decides how the offset is applied (set via --dropdown-offset CSS variable).
  const verticalPlacement = pos[0] === 'above' || pos[0] === 'below'
  const manual = pos.includes('manual')
  const cls = classnames(
    style.dropdown,
    className,
    {
      [style.larger]: larger,
      [style.hover]: hover,
      [style.vertical]: verticalPlacement,
    },
    !manual
      ? {
          [style.below]: pos.includes('below'),
          [style.above]: pos.includes('above'),
          [style.right]: pos.includes('right'),
          [style.left]: pos.includes('left'),
        }
      : {},
  )

  return (
    <ul onClick={onItemsClick} className={cls} ref={ref}>
      {children}
    </ul>
  )
}

Dropdown.propTypes = {
  children: PropTypes.node.isRequired,
  className: PropTypes.string,
  hover: PropTypes.bool,
  larger: PropTypes.bool,
  onItemsClick: PropTypes.func,
  onOutsideClick: PropTypes.func,
  open: PropTypes.bool.isRequired,
  position: PropTypes.string,
}

Dropdown.defaultProps = {
  className: undefined,
  hover: false,
  larger: false,
  onItemsClick: () => null,
  onOutsideClick: () => null,
  position: 'below left',
}

const DropdownItem = ({
  active,
  icon,
  title,
  path,
  action,
  exact,
  showActive,
  tabIndex,
  submenuItems,
  external,
  messageClassName,
  ...rest
}) => {
  const ref = useRef()
  const iconElement = icon && <Icon className={style.icon} icon={icon} nudgeUp />
  const itemElement = action ? (
    <button
      onClick={action}
      onKeyPress={action}
      role="tab"
      tabIndex={tabIndex}
      className={classnames(style.button, { [style.buttonActive]: active })}
    >
      {iconElement}
      <Message content={title} component="span" className={messageClassName} />
    </button>
  ) : external ? (
    <Link.Anchor href={path} external tabIndex={tabIndex} className={style.button}>
      {iconElement}
      <Message content={title} component="span" className={messageClassName} />
    </Link.Anchor>
  ) : (
    <Link to={path} tabIndex={tabIndex} className={style.button}>
      {iconElement}
      <Message content={title} component="span" className={messageClassName} />
    </Link>
  )

  const submenu = Boolean(submenuItems) && (
    <>
      <button className={classnames(style.button, 'd-flex', 'j-between')}>
        <div className="d-flex al-center">
          {iconElement}
          <Message content={title} component="span" className={messageClassName} />
        </div>
        <Icon className={style.submenuDropdownIcon} icon={IconChevronRight} />
      </button>
      <Dropdown.Attached attachedRef={ref} className={style.submenuDropdown} position="left" hover>
        {submenuItems}
      </Dropdown.Attached>
    </>
  )

  return (
    <li className={style.dropdownItem} key={title.id || title} {...rest} ref={ref}>
      {submenu || itemElement}
    </li>
  )
}

DropdownItem.propTypes = {
  action: PropTypes.func,
  active: PropTypes.bool,
  exact: PropTypes.bool,
  external: PropTypes.bool,
  icon: PropTypes.icon,
  messageClassName: PropTypes.string,
  path: PropTypes.string,
  showActive: PropTypes.bool,
  submenuItems: PropTypes.arrayOf(PropTypes.node),
  tabIndex: PropTypes.string,
  title: PropTypes.message.isRequired,
}

DropdownItem.defaultProps = {
  active: false,
  action: undefined,
  exact: false,
  external: false,
  icon: undefined,
  path: undefined,
  showActive: true,
  tabIndex: '0',
  submenuItems: undefined,
  messageClassName: undefined,
}

const DropdownHeaderItem = ({ title, className }) => (
  <li className={classnames(style.dropdownHeaderItem, className)}>
    <span>
      <Message content={title} />
    </span>
  </li>
)

DropdownHeaderItem.propTypes = {
  className: PropTypes.string,
  title: PropTypes.message.isRequired,
}

DropdownHeaderItem.defaultProps = {
  className: undefined,
}

Dropdown.Item = DropdownItem
Dropdown.HeaderItem = DropdownHeaderItem
Dropdown.Attached = AttachedDropdown

export default Dropdown
