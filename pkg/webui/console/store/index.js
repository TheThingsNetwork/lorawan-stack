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

/* global process */

import { createStore, applyMiddleware, compose } from 'redux'
import { createLogicMiddleware } from 'redux-logic'
import { routerMiddleware } from 'connected-react-router'

import requestPromiseMiddleware from './middleware/request-promise-middleware'

import createRootReducer from './reducers'
import logics from './middleware/logics'

const composeEnhancers =
  (process.env.NODE_ENV === 'development' && window.__REDUX_DEVTOOLS_EXTENSION_COMPOSE__) || compose

export default function(history) {
  const middleware = applyMiddleware(
    requestPromiseMiddleware,
    routerMiddleware(history),
    createLogicMiddleware(logics),
  )

  return createStore(createRootReducer(history), composeEnhancers(middleware))
}
