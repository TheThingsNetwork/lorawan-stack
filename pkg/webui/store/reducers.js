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
