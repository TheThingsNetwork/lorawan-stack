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
import { withRouter, Redirect } from 'react-router-dom'
import { connect } from 'react-redux'
import PropTypes from '../prop-types'

import { withEnv } from '../../lib/components/env'

/**
 * Auth is a component that wraps a tree that requires the user
 * to be authenticated.
 *
 * If no user is authenticated, it renders the login view.
 */
@withEnv
@withRouter
@connect(state => ({
  fetching: state.user.fetching,
  user: state.user.user,
}))
class Auth extends React.PureComponent {

  render () {
    const {
      user,
      fetching,
      children,
      env: { appRoot },
    } = this.props

    if (fetching) {
      return null
    }

    if (!Boolean(user)) {
      const redirectPath = window.location.pathname.substring(appRoot.length)
      return (
        <Redirect
          to={{
            pathname: `/login`,
            search: redirectPath && `?next=${redirectPath}`,
            state: { from: this.props.location.pathname },
          }}
        />
      )
    }

    return children
  }
}

Auth.propTypes = {
  user: PropTypes.object,
  fetching: PropTypes.bool,
}

export default Auth
