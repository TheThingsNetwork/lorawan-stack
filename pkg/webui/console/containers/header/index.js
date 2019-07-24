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

import { withEnv } from '../../../lib/components/env'
import { logout } from '../../store/actions/user'
import PropTypes from '../../../lib/prop-types'
import sharedMessages from '../../../lib/shared-messages'

import HeaderComponent from '../../../components/header'

@withRouter
@connect(state => ({
  user: state.user.user,
}))
@withEnv
@bind
class Header extends Component {

  handleLogout () {
    const { dispatch } = this.props
    dispatch(logout())
  }

  render () {
    const {
      user,
      anchored,
      handleSearchRequest,
      searchable,
      env: { appRoot },
    } = this.props

    const navigationEntries = [
      {
        title: sharedMessages.overview,
        icon: 'overview',
        path: appRoot,
        exact: true,
      },
      {
        title: sharedMessages.applications,
        icon: 'application',
        path: `${appRoot}/applications`,
      },
      {
        title: sharedMessages.gateways,
        icon: 'gateway',
        path: `${appRoot}/gateways`,
      },
    ]

    const dropdownItems = [
      {
        title: sharedMessages.logout,
        icon: 'power_settings_new',
        action: this.handleLogout,
      },
    ]

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
