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
import { ConnectedRouter } from 'connected-react-router'

import { Route, Switch } from 'react-router-dom'

import IntlHelmet from '../../../lib/components/intl-helmet'
import { withEnv } from '../../../lib/components/env'
import ErrorView from '../../../lib/components/error-view'
import dev from '../../../lib/dev'

import SideNavigation from '../../../components/navigation/side'
import Header from '../../containers/header'
import Footer from '../../../components/footer'
import Landing from '../landing'
import Login from '../login'
import FullViewError from '../error'

import style from './app.styl'

@withEnv
class ConsoleApp extends React.Component {
  render() {
    const {
      env: {
        siteTitle,
        pageData,
        siteName,
        config: { supportLink },
      },
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
          <div className={style.app}>
            <IntlHelmet
              titleTemplate={`%s - ${siteTitle ? `${siteTitle} - ` : ''}${siteName}`}
              defaultTitle={siteName}
            />
            <div id="modal-container" />
            <Header className={style.header} />
            <main className={style.main}>
              <div>
                <SideNavigation />
              </div>
              <div className={style.content}>
                <Switch>
                  {/* routes for registration, privacy policy, other public pages */}
                  <Route path="/login" component={Login} />
                  <Route path="/" component={Landing} />
                </Switch>
              </div>
            </main>
            <Footer className={style.footer} supportLink={supportLink} />
          </div>
        </ErrorView>
      </ConnectedRouter>
    )
  }
}

const ExportedApp = dev ? hot(ConsoleApp) : ConsoleApp

export default ExportedApp
