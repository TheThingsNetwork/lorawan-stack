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
import classnames from 'classnames'

import IntlHelmet from '../../../lib/components/intl-helmet'
import { withEnv } from '../../../lib/components/env'
import ErrorView from '../../../lib/components/error-view'
import dev from '../../../lib/dev'
import PropTypes from '../../../lib/prop-types'
import { ToastContainer } from '../../../components/toast'

import Header from '../../containers/header'
import Footer from '../../../components/footer'
import Landing from '../landing'
import Login from '../login'
import FullViewError from '../error'

import style from './app.styl'

@withEnv
@(Component => (dev ? hot(Component) : Component))
class ConsoleApp extends React.Component {
  static propTypes = {
    env: PropTypes.env.isRequired,
    history: PropTypes.shape({
      push: PropTypes.func,
      replace: PropTypes.func,
    }).isRequired,
  }

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
      <React.Fragment>
        <ToastContainer />
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
                <div className={classnames('breadcrumbs', style.mobileBreadcrumbs)} />
                <div className={style.sidebar} id="sidebar" />
                <div className={style.content}>
                  <div className={classnames('breadcrumbs', style.desktopBreadcrumbs)} />
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
      </React.Fragment>
    )
  }
}

export default ConsoleApp
