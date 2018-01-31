// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

import { put } from "redux-saga/effects"
import createSagaMiddleware from "redux-saga"
import { Set } from "immutable"
import bind from "autobind-decorator"

import errors from "../actions/errors"

const flatten = function (s = []) {
  const res = []

  if (!Array.isArray(s)) {
    return [ s ]
  }

  for (const x of s) {
    const f = flatten(x)
    for (const y of f) {
      res.push(y)
    }
  }

  return res
}

// Sagas is a helper for managing sagas, injecting asynchronously loaded
// sagas on the fly.
@bind
export default class Sagas {
  // create a new Sagas object and inject the passed sagas.
  constructor (sagas = []) {
    this.middleware = createSagaMiddleware()
    this.sagas = Set(flatten(sagas))
  }

  // inject adds new sagas to the middleware, making sure the same
  // saga is never added twice.
  inject (sagas = []) {
    flatten(sagas).forEach(this.run)
  }

  run (saga) {
    if (!this.sagas.has(saga)) {
      this.exec(saga)
    }
  }

  exec (saga) {
    this.sagas = this.sagas.add(saga)
    this.middleware.run(wrap(saga))
  }

  bind (store) {
    this.sagas.forEach(this.exec)

    store.sagas = this
  }
}

const wrap = saga => function * () {
  try {
    yield * saga()
  } catch (error) {
    yield put(errors.uncaught({ error }))
  }
}
