// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

import Button from '../../../components/button'
import WithAuth from '../../../lib/components/with-auth'
import api from '../../../api'

@connect((state, props) => ({
  user: state.user.user,
})
)
export default class OAuth extends React.PureComponent {


  async handleLogout () {
    await api.oauth.logout()
    window.location = '/oauth/login'
  }

  render () {
    const { user = {}} = this.props

    return (
      <WithAuth>
        <div>You are logged in as {user.user_id}. <Button message="Logout" onClick={this.handleLogout} /></div>
      </WithAuth>
    )
  }
}
