// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

import { applyMiddleware, createStore, compose } from "redux"
import { routerMiddleware } from "react-router-redux"

import initialSagas from "../sagas"
import initialReducers from "../reducers"

import Reducers from "./reducers"
import Sagas from "./sagas"

import log from "./log"

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
