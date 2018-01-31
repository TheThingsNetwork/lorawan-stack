// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

import { combineReducers } from "redux"
import { Map } from "immutable"
import bind from "autobind-decorator"

// Reducers is a helper for managing reducers, injecting asynchronously
// loaded reducers on the fly.
@bind
export default class Reducers {
  constructor (reducers = {}) {
    this.reducers = Map()
    this.inject(reducers)
  }

  // bind binds this reducer to a specific store, updating it when
  // new reducers get added.
  bind (store) {
    this.store = store
    this.replace()

    store.reducers = this
  }

  // inject injects new reducers to the bound store.
  inject (reducers) {
    if (!reducers) {
      return
    }

    this.reducers = this.reducers.merge(reducers)
    this.reducer = combineReducers(this.reducers.toJS())
    this.replace()
  }

  replace () {
    if (this.store) {
      this.store.replaceReducer(this.reducer)
    }
  }
}
