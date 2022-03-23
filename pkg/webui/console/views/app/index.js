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

import React from 'react'
import { connect } from 'react-redux'
import { ConnectedRouter } from 'connected-react-router'
import { Route, Switch, Redirect } from 'react-router-dom'
import classnames from 'classnames'
import bind from 'autobind-decorator'

import { ToastContainer } from '@ttn-lw/components/toast'
import sidebarStyle from '@ttn-lw/components/navigation/side/side.styl'

import Footer from '@ttn-lw/containers/footer'
import LogBackInModal from '@ttn-lw/containers/log-back-in-modal'

import IntlHelmet from '@ttn-lw/lib/components/intl-helmet'
import { withEnv } from '@ttn-lw/lib/components/env'
import ErrorView from '@ttn-lw/lib/components/error-view'
import ScrollToTop from '@ttn-lw/lib/components/scroll-to-top'
import WithAuth from '@ttn-lw/lib/components/with-auth'
import FullViewError, { FullViewErrorInner } from '@ttn-lw/lib/components/full-view-error'

import Header from '@console/containers/header'

import Overview from '@console/views/overview'
import Applications from '@console/views/applications'
import Gateways from '@console/views/gateways'
import Organizations from '@console/views/organizations'
import Admin from '@console/views/admin'
import User from '@console/views/user'

import PropTypes from '@ttn-lw/lib/prop-types'
import { setStatusOnline } from '@ttn-lw/lib/store/actions/status'
import { selectIsLoggedIn, selectOnlineStatus } from '@ttn-lw/lib/store/selectors/status'

import {
  selectUser,
  selectUserFetching,
  selectUserError,
  selectUserRights,
  selectUserIsAdmin,
} from '@console/store/selectors/user'

import style from './app.styl'

const GenericNotFound = () => <FullViewErrorInner error={{ statusCode: 404 }} />
const errorRender = error => <FullViewError error={error} header={<Header />} />

@withEnv
@connect(
  state => ({
    user: selectUser(state),
    fetching: selectUserFetching(state),
    error: selectUserError(state),
    rights: selectUserRights(state),
    isAdmin: selectUserIsAdmin(state),
    isLoggedIn: selectIsLoggedIn(state),
    onlineStatus: selectOnlineStatus(state),
  }),
  {
    setStatusOnline,
  },
)
class ConsoleApp extends React.PureComponent {
  static propTypes = {
    env: PropTypes.env.isRequired,
    error: PropTypes.error,
    fetching: PropTypes.bool.isRequired,
    history: PropTypes.shape({
      push: PropTypes.func,
      replace: PropTypes.func,
      location: PropTypes.shape({
        pathname: PropTypes.string.isRequired,
      }).isRequired,
    }).isRequired,
    isAdmin: PropTypes.bool,
    isLoggedIn: PropTypes.bool.isRequired,
    rights: PropTypes.rights,
    setStatusOnline: PropTypes.func.isRequired,
    user: PropTypes.user,
  }
  static defaultProps = {
    user: undefined,
    error: undefined,
    isAdmin: undefined,
    rights: undefined,
  }

  @bind
  handleConnectionStatusChange({ type }) {
    const { setStatusOnline } = this.props

    setStatusOnline(type === 'online')
  }

  componentDidMount() {
    window.addEventListener('online', this.handleConnectionStatusChange)
    window.addEventListener('offline', this.handleConnectionStatusChange)
  }

  componentWillUnmount() {
    window.removeEventListener('online', this.handleConnectionStatusChange)
    window.removeEventListener('offline', this.handleConnectionStatusChange)
  }

  render() {
    const {
      user,
      fetching,
      error,
      rights,
      isAdmin,
      history,
      history: {
        location: { pathname },
      },
      env: { siteTitle, pageData, siteName },
      isLoggedIn,
    } = this.props

    if (pageData && pageData.error) {
      return (
        <ConnectedRouter history={history}>
          <FullViewError error={pageData.error} header={<Header />} />
        </ConnectedRouter>
      )
    }

    return (
      <React.Fragment>
        <ToastContainer />
        <ConnectedRouter history={history}>
          <ScrollToTop />
          <ErrorView errorRender={errorRender}>
            <div className={style.app}>
              <IntlHelmet
                titleTemplate={`%s - ${siteTitle ? `${siteTitle} - ` : ''}${siteName}`}
                defaultTitle={siteName}
              />
              <div id="modal-container" />
              <Header />
              <main className={style.main}>
                <WithAuth
                  user={user}
                  fetching={fetching}
                  error={error}
                  errorComponent={FullViewErrorInner}
                  rights={rights}
                  isAdmin={isAdmin}
                >
                  {!isLoggedIn && <LogBackInModal />}
                  <div className={classnames('breadcrumbs', style.mobileBreadcrumbs)} />
                  <div id="sidebar" className={sidebarStyle.container} />
                  <div className={style.content}>
                    <div className={classnames('breadcrumbs', style.desktopBreadcrumbs)} />
                    <div className={style.stage} id="stage">
                      <Switch>
                        <Redirect from="/:url*(/+)" to={pathname.slice(0, -1)} />
                        <Route exact path="/" component={Overview} />
                        <Route path="/applications" component={Applications} />
                        <Route path="/gateways" component={Gateways} />
                        <Route path="/organizations" component={Organizations} />
                        <Route path="/admin" component={Admin} />
                        <Route path="/user" component={User} />
                        <Route component={GenericNotFound} />
                      </Switch>
                    </div>
                  </div>
                </WithAuth>
              </main>
              <Footer className={style.footer} />
            </div>
          </ErrorView>
        </ConnectedRouter>
      </React.Fragment>
    )
  }
}

export default ConsoleApp
