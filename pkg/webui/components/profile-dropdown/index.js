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
import classnames from 'classnames'

import Icon from '../icon'
import Dropdown from '../dropdown'
import PropTypes from '../../lib/prop-types'

import styles from './profile-dropdown.styl'

export default class ProfileDropdown extends React.PureComponent {
  state = {
    expanded: false,
  }

  @bind
  showDropdown() {
    document.addEventListener('mousedown', this.handleClickOutside)
    this.setState({ expanded: true })
  }

  @bind
  hideDropdown() {
    document.removeEventListener('mousedown', this.handleClickOutside)
    this.setState({ expanded: false })
  }

  @bind
  handleClickOutside(e) {
    if (!this.node.contains(e.target)) {
      this.hideDropdown()
    }
  }

  @bind
  toggleDropdown() {
    let { expanded } = this.state
    expanded = !expanded
    if (expanded) {
      this.showDropdown()
    } else {
      this.hideDropdown()
    }
  }

  @bind
  ref(node) {
    this.node = node
  }

  render() {
    const { userId, className, children, ...rest } = this.props

    return (
      <div
        className={classnames(styles.container, className)}
        onClick={this.toggleDropdown}
        onKeyPress={this.toggleDropdown}
        ref={this.ref}
        tabIndex="0"
        role="button"
        {...rest}
      >
        <span className={styles.id}>{userId}</span>
        <Icon icon="arrow_drop_down" />
        {this.state.expanded && <Dropdown className={styles.dropdown}>{children}</Dropdown>}
      </div>
    )
  }
}

ProfileDropdown.propTypes = {
  /**
   * A list of items for the dropdown component
   * See `<Dropdown />`'s `items` proptypes for details
   */
  children: PropTypes.node.isRequired,
  className: PropTypes.string,
  /** The id of the current user */
  userId: PropTypes.string.isRequired,
}

ProfileDropdown.defaultProps = {
  className: undefined,
}
