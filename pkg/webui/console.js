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
import DOM from 'react-dom'
import { ConnectedRouter } from 'connected-react-router'
import { createBrowserHistory } from 'history'
import { Provider } from 'react-redux'

import { EnvProvider } from './lib/components/env'
import { BreadcrumbsProvider } from './components/breadcrumbs/context'
import { SideNavigationProvider } from './components/navigation/side/context'
import Init from './lib/components/init'
import WithLocale from './lib/components/with-locale'

import createStore from './console/store'
import App from './console/views/app'

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

const Console = () => (
  <EnvProvider env={env}>
    <Provider store={store}>
      <Init>
        <WithLocale>
          <ConnectedRouter history={history}>
            <BreadcrumbsProvider>
              <SideNavigationProvider>
                <App />
              </SideNavigationProvider>
            </BreadcrumbsProvider>
          </ConnectedRouter>
        </WithLocale>
      </Init>
    </Provider>
  </EnvProvider>
)

DOM.render((<Console />), document.getElementById('app'))
