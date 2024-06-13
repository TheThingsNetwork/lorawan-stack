// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

import * as Sentry from '@sentry/react'
import { configureStore } from '@reduxjs/toolkit'
import { createLogicMiddleware } from 'redux-logic'

import sensitiveFields from '@ttn-lw/constants/sensitive-data'

import omitDeep from '@ttn-lw/lib/omit'
import env from '@ttn-lw/lib/env'
import dev from '@ttn-lw/lib/dev'
import requestPromiseMiddleware from '@ttn-lw/lib/store/middleware/request-promise-middleware'
import { trimEvents } from '@ttn-lw/lib/store/util'

import { selectUserId } from '@console/store/selectors/logout'

import rootReducer from './reducers'
import logics from './middleware/logics'

const logicMiddleware = createLogicMiddleware(logics)

const middlewares = [requestPromiseMiddleware, logicMiddleware]

const sentryEnhancer = Sentry.createReduxEnhancer({
  stateTransformer: state => omitDeep(trimEvents(state), sensitiveFields),
  actionTransformer: action => omitDeep(action, sensitiveFields),
  configureScopeWithState: (scope, state) => scope.setUser({ id: selectUserId(state) }),
})

const enhancers = env.sentryDsn ? [sentryEnhancer] : []

const store = configureStore({
  reducer: rootReducer,
  middleware: getDefaultMiddleware =>
    getDefaultMiddleware({
      serializableCheck: {
        ignoredActionPaths: ['meta._resolve', 'meta._reject'],
      },
    }).concat(middlewares),
  enhancer: getDefaultEnhancers => getDefaultEnhancers().concat(enhancers),
  devTools: dev && window.__REDUX_DEVTOOLS_EXTENSION_COMPOSE__,
})

if (dev && module.hot) {
  module.hot.accept('./reducers', () => {
    store.replaceReducer(rootReducer)
  })
}

export default store
