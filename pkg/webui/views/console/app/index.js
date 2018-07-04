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
import { Switch, Route, BrowserRouter } from 'react-router-dom'
import { Provider } from 'react-redux'

import Header from '../../../components/header'
import Login from '../login'
import Landing from '../landing'
import store from '../../../store'

export default class ConsoleApp extends React.PureComponent {
  render () {
    return (
      <Provider store={store}>
        <div>
          <Header />
          <BrowserRouter>
            <Switch>
              <Route exact path="/console" component={Landing} />
              <Route path="/console/login" component={Login} />
            </Switch>
          </BrowserRouter>
        </div>
      </Provider>
    )
  }
}
