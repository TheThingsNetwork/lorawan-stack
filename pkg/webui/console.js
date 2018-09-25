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
import DOM from 'react-dom'
import { ConnectedRouter } from 'connected-react-router'
import { createBrowserHistory } from 'history'
import { Provider } from 'react-redux'
import { BreadcrumbsProvider } from './components/breadcrumbs/context'

import Init from './lib/components/init'
import WithLocale from './lib/components/with-locale'

import createStore from './store'
import App from './views/console/app'

const history = createBrowserHistory()
const store = createStore(history)

const Console = () => (
  <Provider store={store}>
    <Init>
      <WithLocale>
        <ConnectedRouter history={history}>
          <BreadcrumbsProvider>
            <App />
          </BreadcrumbsProvider>
        </ConnectedRouter>
      </WithLocale>
    </Init>
  </Provider>
)

DOM.render((<Console />), document.getElementById('app'))
