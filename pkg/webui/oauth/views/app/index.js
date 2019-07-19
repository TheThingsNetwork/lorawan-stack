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
import { Switch, Route } from 'react-router-dom'
import { ConnectedRouter } from 'connected-react-router'
import { createBrowserHistory } from 'history'
import { Provider } from 'react-redux'
import { Helmet } from 'react-helmet'

import WithLocale from '../../../lib/components/with-locale'
import withEnv, { EnvProvider } from '../../../lib/components/env'
import { pageDataSelector } from '../../../lib/selectors/env'
import ErrorView from '../../../lib/components/error-view'

import Landing from '../landing'
import Login from '../login'
import Authorize from '../authorize'
import CreateAccount from '../create-account'
import ResetPassword from '../reset-password'
import FullViewError from '../error'
import createStore from '../../store'
import Init from '../../../lib/components/init'
import Code from '../code'

const history = createBrowserHistory()
const store = createStore(history)
const env = {
  app_root: window.APP_ROOT,
  assets_root: window.ASSETS_ROOT,
  config: window.APP_CONFIG,
  page_data: window.PAGE_DATA,
  site_name: window.SITE_NAME,
  site_title: window.SITE_TITLE,
}

const GenericNotFound = () => <FullViewError error={{ statusCode: 404 }} />
@withEnv
export default class OAuthApp extends React.PureComponent {
  render () {
    const pageData = pageDataSelector(env)

    if (pageData && pageData.error) {
      return (
        <EnvProvider env={env}>
          <Provider store={store}>
            <WithLocale>
              <FullViewError error={pageData.error} />
            </WithLocale>
          </Provider>
        </EnvProvider>
      )
    }

    return (
      <EnvProvider env={env}>
        <Provider store={store}>
          <Init>
            <Helmet
              titleTemplate={`%s - ${env.site_title ? `${env.site_title} - ` : ''}${env.site_name}`}
              defaultTitle={`${env.site_title ? `${env.site_title} - ` : ''}${env.site_name}`}
            />
            <WithLocale>
              <ErrorView ErrorComponent={FullViewError}>
                <ConnectedRouter history={history}>
                  <Switch>
                    <Route path="/oauth" exact component={Landing} />
                    <Route path="/oauth/login" component={Login} />
                    <Route path="/oauth/authorize" component={Authorize} />
                    <Route path="/oauth/register" component={CreateAccount} />
                    <Route path="/oauth/reset-password" component={ResetPassword} />
                    <Route path="/oauth/code" component={Code} />
                    <Route component={GenericNotFound} />
                  </Switch>
                </ConnectedRouter>
              </ErrorView>
            </WithLocale>
          </Init>
        </Provider>
      </EnvProvider>
    )
  }
}
