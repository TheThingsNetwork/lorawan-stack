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

import { Route, Switch, withRouter } from 'react-router-dom'
import { connect } from 'react-redux'
import bind from 'autobind-decorator'
import { Helmet } from 'react-helmet'

import { withEnv } from '../../../lib/components/env'
import SideNavigation from '../../../components/navigation/side'
import Header from '../../../components/header'
import Footer from '../../../components/footer'
import Landing from '../landing'
import Login from '../login'

import { logout } from '../../../actions/user'

import style from './app.styl'

@withRouter
@withEnv
@connect(state => ({
  user: state.user.user,
}))
@bind
export default class ConsoleApp extends React.Component {

  handleLogout () {
    const { dispatch } = this.props
    dispatch(logout())
  }

  render () {
    const {
      user,
      env,
    } = this.props

    return (
      <div className={style.app}>
        <Helmet
          titleTemplate="%s - Console - The Things Network"
          defaultTitle="The Things Network Console"
        />
        <div id="modal-container" />
        <Header className={style.header} user={user} handleLogout={this.handleLogout} />
        <main className={style.main}>
          <div>
            <SideNavigation />
          </div>
          <div className={style.content}>
            <Switch>
              {/* routes for registration, privacy policy, other public pages */}
              <Route path={`${env.app_root}/login`} component={Login} />
              <Route path={env.app_root} component={Landing} />
            </Switch>
          </div>
        </main>
        <Footer className={style.footer} />
      </div>
    )
  }
}
