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
import PropTypes from 'prop-types'
import { withRouter, Redirect } from 'react-router'
import { connect } from 'react-redux'

/**
 * Auth is a component that wraps a tree that requires the user
 * to be authenticated.
 *
 * If no user is authenticated, it renders the Landing view.
 */
@withRouter
export class Auth extends React.PureComponent {
  static propTypes = {
    user: PropTypes.object,
    fetching: PropTypes.bool,
  }

  render () {
    const {
      user,
      fetching,
      children,
    } = this.props

    if (fetching) {
      return null
    }

    if (!user) {
      return <Redirect to={`/${window.ENV.console ? 'console' : 'oauth'}/login`} />
    }

    return children
  }
}

const mapStateToProps = function (state) {
  return {
    fetching: state.user.fetching,
    user: state.user.user,
  }
}

export default connect(mapStateToProps)(Auth)
