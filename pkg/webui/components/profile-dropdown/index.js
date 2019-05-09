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
import { Link } from 'react-router-dom'

import Icon from '../icon'
import Message from '../../lib/components/message'
import PropTypes from '../../lib/prop-types'

import styles from './profile-dropdown.styl'

@bind
export default class ProfileDropdown extends React.PureComponent {

  static propTypes = {
    /**
    * The id of the current user
    */
    userId: PropTypes.string.isRequired,
    /**
    * A list of items for the dropdown
    * @param {(string|Object)} title - The title to be displayed
    * @param {string} icon - The icon name to be displayed next to the title
    * @param {string} path - The path for a navigation tab
    * @param {function} action - Alternatively, the function to be called on click
    */
    dropdownItems: PropTypes.arrayOf(PropTypes.shape({
      title: PropTypes.message.isRequired,
      icon: PropTypes.string,
      path: PropTypes.string,
      action: PropTypes.func,
    })),
  }

  state = {
    expanded: false,
  }

  showDropdown () {
    document.addEventListener('mousedown', this.handleClickOutside)
    this.setState({ expanded: true })
  }

  hideDropdown () {
    document.removeEventListener('mousedown', this.handleClickOutside)
    this.setState({ expanded: false })
  }

  handleClickOutside (e) {
    if (!this.node.contains(e.target)) {
      this.hideDropdown()
    }
  }

  toggleDropdown () {
    let { expanded } = this.state
    expanded = !expanded
    if (expanded) {
      this.showDropdown()
    } else {
      this.hideDropdown()
    }
  }

  ref (node) {
    this.node = node
  }

  render () {
    const { userId, dropdownItems, ...rest } = this.props

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
        <div className={styles.profilePicture} />
        <span className={styles.id}>{userId}</span>
        <Icon icon="arrow_drop_down" />
        { this.state.expanded && <Dropdown items={dropdownItems} />}
      </div>
    )
  }
}

const Dropdown = ({ items }) => (
  <ul className={styles.dropdown}>
    { items.map( function (item) {
      const icon = item.icon && <Icon className={styles.icon} icon={item.icon} />
      return (
        <li className={styles.dropdownItem} key={item.title.id || item.title}>
          { item.action
            ? <button onClick={item.action} onKeyPress={item.action} role="tab" tabIndex="0">{icon}<Message content={item.title} /></button>
            : <Link to={item.path}>{icon}<Message content={item.title} /></Link>
          }
        </li>
      )
    }
    )}
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
  items: PropTypes.arrayOf(PropTypes.shape({
    title: PropTypes.message.isRequired,
    icon: PropTypes.string,
    path: PropTypes.string,
    action: PropTypes.func,
  })),
}

export { Dropdown }
