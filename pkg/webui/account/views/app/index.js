// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

import { hot } from 'react-hot-loader/root'
import { connect } from 'react-redux'
import React from 'react'
import { Switch, Route } from 'react-router-dom'
import { ConnectedRouter } from 'connected-react-router'
import { Helmet } from 'react-helmet'

import ErrorView from '@ttn-lw/lib/components/error-view'
import { FullViewError } from '@ttn-lw/lib/components/full-view-error/error'

import Landing from '@account/views/landing'
import Authorize from '@account/views/authorize'

import PropTypes from '@ttn-lw/lib/prop-types'
import dev from '@ttn-lw/lib/dev'

import { selectUser } from '@account/store/selectors/user'

import Front from '../front'

const GenericNotFound = () => <FullViewError error={{ statusCode: 404 }} />

class AccountApp extends React.PureComponent {
  static propTypes = {
    env: PropTypes.env.isRequired,
    history: PropTypes.history.isRequired,
    user: PropTypes.user,
  }

  static defaultProps = {
    user: undefined,
  }

  render() {
    const {
      env: { siteTitle, pageData, siteName },
      history,
      user,
    } = this.props

    if (pageData && pageData.error) {
      return (
        <ConnectedRouter history={history}>
          <FullViewError error={pageData.error} />
        </ConnectedRouter>
      )
    }

    return (
      <ConnectedRouter history={history}>
        <ErrorView ErrorComponent={FullViewError}>
          <React.Fragment>
            <Helmet
              titleTemplate={`%s - ${siteTitle ? `${siteTitle} - ` : ''}${siteName}`}
              defaultTitle={`${siteTitle ? `${siteTitle} - ` : ''}${siteName}`}
            />
            {Boolean(user) ? (
              <Switch>
                <Route path="/" exact component={Landing} />
                <Route component={GenericNotFound} />
              </Switch>
            ) : (
              <Switch>
                <Route path="/authorize" component={Authorize} />
                <Route path="/" component={Front} />
              </Switch>
            )}
          </React.Fragment>
        </ErrorView>
      </ConnectedRouter>
    )
  }
}

const ExportedApp = dev ? hot(AccountApp) : AccountApp

export default connect(state => ({
  user: selectUser(state),
}))(ExportedApp)
