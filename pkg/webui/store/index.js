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

import { applyMiddleware, createStore, compose } from 'redux'
import { routerMiddleware } from 'react-router-redux'

import initialSagas from '../sagas'
import initialReducers from '../reducers'

import Reducers from './reducers'
import Sagas from './sagas'

import log from './log'

/**
 * Create a redux store.
 *
 * @param {object} history - The history object to use in react-router-redux.
 * @param {object} reducers - Optional extra reducers.
 * @param {array} sagas - Optional extra sagas.
 *
 * @returns {object} - The redux store.
 */
export default function (history, reducers = {}, sagas = []) {
  const r = new Reducers({ ...initialReducers, ...reducers })
  const s = new Sagas([ ...initialSagas, ...sagas ])

  const store = compose(
    applyMiddleware(routerMiddleware(history)),
    applyMiddleware(s.middleware),
    applyMiddleware(routerMiddleware(history)),
    applyMiddleware(log),
  )(createStore)(r.reducer)

  r.bind(store)
  s.bind(store)

  return store
}
