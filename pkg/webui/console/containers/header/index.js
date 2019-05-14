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

import React, { Component } from 'react'
import { connect } from 'react-redux'
import { withRouter } from 'react-router-dom'
import bind from 'autobind-decorator'

import { logout } from '../../store/actions/user'
import PropTypes from '../../../lib/prop-types'

import HeaderComponent from '../../../components/header'

@withRouter
@connect(state => ({
  user: state.user.user,
}))
@bind
class Header extends Component {

  handleLogout () {
    const { dispatch } = this.props
    dispatch(logout())
  }

  render () {
    const {
      user,
      dropdownItems,
      navigationEntries,
      anchored,
      handleSearchRequest,
      searchable,
    } = this.props

    return (
      <HeaderComponent
        user={user}
        handleLogout={this.handleLogout}
        dropdownItems={dropdownItems}
        navigationEntries={navigationEntries}
        anchored={anchored}
        searchable={searchable}
        handleSearchRequest={handleSearchRequest}
      />
    )
  }
}

Header.propTypes = {
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
    path: PropTypes.string.isRequired,
    action: PropTypes.func,
  })),
  /**
   * A list of navigation bar entries.
   * @param {(string|Object)} title - The title to be displayed
   * @param {string} icon - The icon name to be displayed next to the title
   * @param {string} path -  The path for a navigation tab
   * @param {boolean} exact - Flag identifying whether the path should be matched exactly
   */
  navigationEntries: PropTypes.arrayOf(PropTypes.shape({
    path: PropTypes.string.isRequired,
    title: PropTypes.message.isRequired,
    action: PropTypes.func,
    icon: PropTypes.string,
  })),
  /** Flag identifying whether links should be rendered as plain anchor link */
  anchored: PropTypes.bool,
  /**
  * A handler for when the user used the search input
  */
  handleSearchRequest: PropTypes.func,
  /** A flag identifying whether the header should display the search input */
  searchable: PropTypes.bool,
}

export default Header
