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

import { hot } from 'react-hot-loader/root'
import React from 'react'
import { Switch, Route } from 'react-router-dom'
import { ConnectedRouter } from 'connected-react-router'
import { Helmet } from 'react-helmet'

import withEnv from '../../../lib/components/env'
import ErrorView from '../../../lib/components/error-view'
import dev from '../../../lib/dev'

import Landing from '../landing'
import Login from '../login'
import Authorize from '../authorize'
import CreateAccount from '../create-account'
import ForgotPassword from '../forgot-password'
import UpdatePassword from '../update-password'
import FullViewError from '../error'
import Code from '../code'
import Validate from '../validate'

const GenericNotFound = () => <FullViewError error={{ statusCode: 404 }} />

@withEnv
class OAuthApp extends React.PureComponent {
  render() {
    const {
      env: { siteTitle, pageData, siteName },
      history,
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
            <Switch>
              <Route path="/" exact component={Landing} />
              <Route path="/login" component={Login} />
              <Route path="/authorize" component={Authorize} />
              <Route path="/register" component={CreateAccount} />
              <Route path="/forgot-password" component={ForgotPassword} />
              <Route path="/code" component={Code} />
              <Route path="/update-password" component={UpdatePassword} />
              <Route path="/validate" component={Validate} />
              <Route component={GenericNotFound} />
            </Switch>
          </React.Fragment>
        </ErrorView>
      </ConnectedRouter>
    )
  }
}

const ExportedApp = dev ? hot(OAuthApp) : OAuthApp

export default ExportedApp
