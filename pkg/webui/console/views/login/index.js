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
import Query from 'query-string'
import bind from 'autobind-decorator'
import { connect } from 'react-redux'
import { defineMessages } from 'react-intl'
import { Redirect } from 'react-router-dom'

import PropTypes from '../../../lib/prop-types'
import { selectApplicationSiteName, selectApplicationRootPath } from '../../../lib/selectors/env'
import { FullViewErrorInner } from '../../views/error'
import Spinner from '../../../components/spinner'
import Message from '../../../lib/components/message'

const m = defineMessages({
  cannotLogin: 'Login not possible',
  redirecting: 'Redirecting…',
})

@connect(state => ({
  error: state.user.error,
  user: state.user.user,
  siteName: selectApplicationSiteName(),
  appRoot: selectApplicationRootPath(),
}))
@bind
export default class Login extends React.PureComponent {
  static propTypes = {
    appRoot: PropTypes.string.isRequired,
    error: PropTypes.error,
    user: PropTypes.user,
  }

  static defaultProps = {
    error: undefined,
    user: undefined,
  }

  componentDidMount() {
    const { user, error, appRoot } = this.props
    const { next } = Query.parse(location.search)
    const redirectAppend = next ? `?next=${next}` : ''

    if (!user && !error) {
      window.location = `${appRoot}/login/ttn-stack${redirectAppend}`
    }
  }

  render() {
    const { error, user } = this.props

    // dont show the login page if the user is already logged in
    if (Boolean(user)) {
      return <Redirect to="/" />
    }

    if (Boolean(error)) {
      return <FullViewErrorInner title={m.cannotLogin} error={error} goBack />
    }

    return (
      <Spinner after={0} center>
        <Message content={m.redirecting} />
      </Spinner>
    )
  }
}
