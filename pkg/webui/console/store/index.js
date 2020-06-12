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

import * as Sentry from '@sentry/browser'
import { createStore, applyMiddleware, compose } from 'redux'
import { createLogicMiddleware } from 'redux-logic'
import { routerMiddleware } from 'connected-react-router'
import createSentryMiddleware from 'redux-sentry-middleware'

import sensitiveFields from '@ttn-lw/constants/sensitive-data'

import omitDeep from '@ttn-lw/lib/omit'
import env from '@ttn-lw/lib/env'
import dev from '@ttn-lw/lib/dev'

import createRootReducer from './reducers'
import requestPromiseMiddleware from './middleware/request-promise-middleware'
import logics from './middleware/logics'

const composeEnhancers = (dev && window.__REDUX_DEVTOOLS_EXTENSION_COMPOSE__) || compose
let middlewares = [requestPromiseMiddleware, createLogicMiddleware(logics)]

if (env.sentryDsn) {
  middlewares = [
    createSentryMiddleware(Sentry, {
      actionTransformer: action => omitDeep(action, sensitiveFields),
      stateTransformer: state => omitDeep(state, sensitiveFields),
      getUserContext: state => {
        return { user_id: state.user.user.ids.user_id }
      },
    }),
    ...middlewares,
  ]
}

export default function(history) {
  const middleware = applyMiddleware(...middlewares, routerMiddleware(history))

  const store = createStore(createRootReducer(history), composeEnhancers(middleware))
  if (dev && module.hot) {
    module.hot.accept('./reducers', () => {
      store.replaceReducer(createRootReducer(history))
    })
  }

  return store
}
