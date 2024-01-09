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

import React, { useCallback, useEffect, useRef, useState } from 'react'
import classnames from 'classnames'
import { NavLink } from 'react-router-dom'

import Icon from '@ttn-lw/components/icon'
import Link from '@ttn-lw/components/link'

import Message from '@ttn-lw/lib/components/message'

import PropTypes from '@ttn-lw/lib/prop-types'

import style from './dropdown.styl'

const Dropdown = ({ className, children, larger, onItemsClick, open }) => {
  const ref = useRef(null)
  const [isBelow, setIsBelow] = useState(false)
  const [isOnRight, setIsOnRight] = useState(false)

  const positionDropdown = useCallback(() => {
    if (ref.current) {
      const parentRect = ref.current.parentElement.getBoundingClientRect()
      const spaceBelow = window.innerHeight - parentRect.bottom
      const spaceAbove = parentRect.top
      const dropdownHeight = ref.current.clientHeight
      const dropdownWidth = ref.current.clientWidth
      const spaceOnLeft = parentRect.left

      setIsBelow(spaceBelow > dropdownHeight || spaceAbove < dropdownHeight)

      setIsOnRight(spaceOnLeft > dropdownWidth)
    }
  }, [])

  useEffect(() => {
    if (open) positionDropdown()
  }, [positionDropdown, open])

  useEffect(() => {
    window.addEventListener('scroll', positionDropdown)
    window.addEventListener('resize', positionDropdown)

    return () => {
      window.removeEventListener('resize', positionDropdown)
      window.removeEventListener('scroll', positionDropdown)
    }
  }, [positionDropdown])

  return (
    <ul
      onClick={onItemsClick}
      className={classnames(style.dropdown, className, {
        [style.larger]: larger,
        [style.below]: isBelow,
        [style.above]: !isBelow,
        [style.right]: isOnRight,
        [style.left]: !isOnRight,
        [style.open]: open,
      })}
      ref={ref}
    >
      {children}
    </ul>
  )
}

Dropdown.propTypes = {
  children: PropTypes.node.isRequired,
  className: PropTypes.string,
  larger: PropTypes.bool,
  onItemsClick: PropTypes.func,
  open: PropTypes.bool.isRequired,
}

Dropdown.defaultProps = {
  className: undefined,
  larger: false,
  onItemsClick: () => null,
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
  ...rest
}) => {
  const [expandedSubmenu, setExpandedSubmenu] = useState(false)
  const [leaveDelayTimer, setLeaveDelayTimer] = useState(null)
  const iconElement = icon && <Icon className={style.icon} icon={icon} nudgeUp />
  const ItemElement = action ? (
    <button
      onClick={action}
      onKeyPress={action}
      role="tab"
      tabIndex={tabIndex}
      className={style.button}
    >
      {iconElement}
      <Message content={title} />
    </button>
  ) : (
    <Link.Anchor href={path} external={external} tabIndex={tabIndex} className={style.button}>
      {Boolean(iconElement) ? iconElement : null}
      <Message content={title} />
    </Link.Anchor>
  )

  const handleMouseEnter = useCallback(() => {
    if (leaveDelayTimer) {
      clearTimeout(leaveDelayTimer)
      setLeaveDelayTimer(null)
    }
    setExpandedSubmenu(true)
  }, [leaveDelayTimer])

  const handleMouseLeave = useCallback(() => {
    // Set a timer when mouse leaves, to only close after a delay.
    // This prevents the menu from closing when the mouse moves over the submenu
    // and also makes the UI more forgiving when the mouse accidentally leaves the menu.
    setLeaveDelayTimer(
      setTimeout(() => {
        setExpandedSubmenu(false)
      }, 250),
    )
  }, [])

  useEffect(
    () => () => {
      if (leaveDelayTimer) {
        clearTimeout(leaveDelayTimer)
      }
    },
    [leaveDelayTimer],
  )

  const withSubmenu = (
    <button
      onMouseEnter={handleMouseEnter}
      onMouseLeave={handleMouseLeave}
      className={classnames(style.button, 'd-flex', 'j-between')}
    >
      <div>
        {iconElement}
        <Message content={title} />
      </div>
      <Icon icon="chevron_right" />
      <Dropdown open={expandedSubmenu} className={style.submenuDropdown}>
        {submenuItems}
      </Dropdown>
    </button>
  )

  return (
    <li className={style.dropdownItem} key={title.id || title} {...rest}>
      {Boolean(submenuItems) ? withSubmenu : ItemElement}
    </li>
  )
}

DropdownItem.propTypes = {
  action: PropTypes.func,
  active: PropTypes.bool,
  exact: PropTypes.bool,
  external: PropTypes.bool,
  icon: PropTypes.string,
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
}

const DropdownHeaderItem = ({ title }) => (
  <li className={style.dropdownHeaderItem}>
    <span>
      <Message content={title} />
    </span>
  </li>
)

DropdownHeaderItem.propTypes = {
  title: PropTypes.message.isRequired,
}

Dropdown.Item = DropdownItem
Dropdown.HeaderItem = DropdownHeaderItem

export default Dropdown
