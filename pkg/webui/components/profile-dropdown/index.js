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
import bind from 'autobind-decorator'

import Icon from '../icon'
import Message from '../../lib/components/message'
import Link from '../link'
import PropTypes from '../../lib/prop-types'

import styles from './profile-dropdown.styl'

const dropdownItemsPropTypes = PropTypes.arrayOf(
  PropTypes.oneOfType([
    PropTypes.shape({
      title: PropTypes.message.isRequired,
      icon: PropTypes.string,
      path: PropTypes.string,
      action: PropTypes.func.isRequired,
    }),
    PropTypes.shape({
      title: PropTypes.message.isRequired,
      icon: PropTypes.string,
      path: PropTypes.string.isRequired,
      action: PropTypes.func,
    }),
  ]),
).isRequired

@bind
export default class ProfileDropdown extends React.PureComponent {
  static propTypes = {
    /** The id of the current user */
    userId: PropTypes.string,
    /**
     * A list of items for the dropdown component
     * See `<Dropdown />`'s `items` proptypes for details
     */
    dropdownItems: dropdownItemsPropTypes,
  }

  state = {
    expanded: false,
  }

  showDropdown() {
    document.addEventListener('mousedown', this.handleClickOutside)
    this.setState({ expanded: true })
  }

  hideDropdown() {
    document.removeEventListener('mousedown', this.handleClickOutside)
    this.setState({ expanded: false })
  }

  handleClickOutside(e) {
    if (!this.node.contains(e.target)) {
      this.hideDropdown()
    }
  }

  toggleDropdown() {
    let { expanded } = this.state
    expanded = !expanded
    if (expanded) {
      this.showDropdown()
    } else {
      this.hideDropdown()
    }
  }

  ref(node) {
    this.node = node
  }

  render() {
    const { userId, dropdownItems, anchored, ...rest } = this.props

    return (
      <div
        className={styles.container}
        onClick={this.toggleDropdown}
        onKeyPress={this.toggleDropdown}
        ref={this.ref}
        tabIndex="0"
        role="button"
        {...rest}
      >
        <span className={styles.id}>{userId}</span>
        <Icon icon="arrow_drop_down" />
        {this.state.expanded && <Dropdown items={dropdownItems} anchored={anchored} />}
      </div>
    )
  }
}

const Dropdown = ({ items, anchored }) => (
  <ul className={styles.dropdown}>
    {items.map(function(item) {
      const icon = item.icon && <Icon className={styles.icon} icon={item.icon} />
      const ItemElement = item.action ? (
        <button onClick={item.action} onKeyPress={item.action} role="tab" tabIndex="0">
          {icon}
          <Message content={item.title} />
        </button>
      ) : anchored ? (
        <Link.BaseAnchor href={item.path}>
          {icon}
          <Message content={item.title} />
        </Link.BaseAnchor>
      ) : (
        <Link to={item.path}>
          {icon}
          <Message content={item.title} />
        </Link>
      )
      return (
        <li className={styles.dropdownItem} key={item.title.id || item.title}>
          {ItemElement}
        </li>
      )
    })}
  </ul>
)

Dropdown.propTypes = {
  /**
   * A list of items for the dropdown
   * @param {(string|Object)} title - The title to be displayed
   * @param {string} icon - The icon name to be displayed next to the title
   * @param {string} path - The path for a navigation tab
   * @param {function} action - Alternatively, the function to be called on click
   */
  items: dropdownItemsPropTypes,
  /** Flag identifying whether link should be rendered as plain anchor link */
  anchored: PropTypes.bool,
}

export { Dropdown }
