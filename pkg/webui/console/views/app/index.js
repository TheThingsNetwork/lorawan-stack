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

import { ToastContainer } from '@ttn-lw/components/toast'
import Footer from '@ttn-lw/components/footer'

import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'
import { withEnv } from '@ttn-lw/lib/components/env'
import ErrorView from '@ttn-lw/lib/components/error-view'
import ScrollToTop from '@ttn-lw/lib/components/scroll-to-top'

import Header from '@console/containers/header'

import Landing from '@console/views/landing'
import Login from '@console/views/login'
import FullViewError from '@console/views/error'

import PropTypes from '@ttn-lw/lib/prop-types'
import dev from '@ttn-lw/lib/dev'

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
          <ScrollToTop />
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
                    {/* Routes for registration, privacy policy, other public pages */}
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
