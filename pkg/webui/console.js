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
import { createBrowserHistory } from 'history'
import { Provider } from 'react-redux'
import * as Sentry from '@sentry/browser'

import sentryConfig from '@ttn-lw/constants/sentry'

import { BreadcrumbsProvider } from '@ttn-lw/components/breadcrumbs/context'

import { EnvProvider } from '@ttn-lw/lib/components/env'
import Init from '@ttn-lw/lib/components/init'
import WithLocale from '@ttn-lw/lib/components/with-locale'

import env from '@ttn-lw/lib/env'
import { selectApplicationRootPath } from '@ttn-lw/lib/selectors/env'
import '@ttn-lw/lib/yup-extensions'

import createStore from './console/store'

const appRoot = selectApplicationRootPath()
const history = createBrowserHistory({ basename: `${appRoot}/` })
// Initialize sentry before creating store.
if (env.sentryDsn) Sentry.init(sentryConfig)
const store = createStore(history)

const rootElement = document.getElementById('app')

const render = () => {
  const App = require('./console/views/app').default

  DOM.render(
    <EnvProvider env={env}>
      <Provider store={store}>
        <WithLocale>
          <Init>
            <BreadcrumbsProvider>
              <App history={history} />
            </BreadcrumbsProvider>
          </Init>
        </WithLocale>
      </Provider>
    </EnvProvider>,
    rootElement,
  )
}

if (module.hot) {
  module.hot.accept('./console/views/app', () => {
    setTimeout(render)
  })
}

render()
