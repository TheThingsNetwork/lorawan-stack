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
import { connect } from 'react-redux'
import bind from 'autobind-decorator'
import { push } from 'connected-react-router'

import Button from '@ttn-lw/components/button'

import sharedMessages from '@ttn-lw/lib/shared-messages'
import PropTypes from '@ttn-lw/lib/prop-types'

import { logout } from '@oauth/store/actions/user'

@connect(
  state => ({
    user: state.user.user,
  }),
  {
    redirectToLogin: () => push('/login'),
    logout,
  },
)
export default class OAuth extends React.PureComponent {
  static propTypes = {
    logout: PropTypes.func.isRequired,
    redirectToLogin: PropTypes.func.isRequired,
    user: PropTypes.user.isRequired,
  }

  @bind
  async handleLogout() {
    const { logout } = this.props

    logout()
  }

  render() {
    const { user, redirectToLogin } = this.props

    if (!Boolean(user)) {
      redirectToLogin()
      return null
    }

    return (
      <div>
        You are logged in as {user.ids.user_id}.{' '}
        <Button message={sharedMessages.logout} onClick={this.handleLogout} />
        <Button.Link message={sharedMessages.changePassword} to="/update-password" />
      </div>
    )
  }
}
