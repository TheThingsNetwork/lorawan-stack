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

import createSagaMiddleware from "redux-saga"
import { Set } from "immutable"
import bind from "autobind-decorator"
import errors from "../actions/errors"
import { put } from "redux-saga/effects"

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
