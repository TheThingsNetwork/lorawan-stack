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
import Header from '@ttn-lw/components/header'

import { ErrorView } from '@ttn-lw/lib/components/error-view'
import { FullViewError } from '@ttn-lw/lib/components/full-view-error'
import Init from '@ttn-lw/lib/components/init'
import WithLocale from '@ttn-lw/lib/components/with-locale'

import Logo from '@console/containers/logo'

import App from '@console/views/app'

import { selectApplicationRootPath, selectSentryDsnConfig } from '@ttn-lw/lib/selectors/env'

import createStore from './console/store'

const appRoot = selectApplicationRootPath()

// Initialize sentry before creating store.
if (selectSentryDsnConfig) {
  Sentry.init(sentryConfig)
}

const history = createBrowserHistory({ basename: `${appRoot}/` })
const store = createStore(history)
const rootElement = document.getElementById('app')

// Error renderer for the outermost error boundary.
// Do not use any components that depend on context
// e.g. Intl, Router, Redux store.
const errorRender = error => (
  <FullViewError error={error} header={<Header logo={<Logo safe />} />} safe />
)

DOM.render(
  <ErrorView errorRender={errorRender}>
    <Provider store={store}>
      <WithLocale>
        <Init>
          <BreadcrumbsProvider>
            <App history={history} />
          </BreadcrumbsProvider>
        </Init>
      </WithLocale>
    </Provider>
  </ErrorView>,
  rootElement,
)
