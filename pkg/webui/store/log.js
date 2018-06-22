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

/* global process */

window.log = (window.ENV || {}).debug || process.env.NODE_ENV === 'development'

const filter = function (action) {
  const {
    type = '',
  } = action

  return (
    !type.startsWith('@@redux-form/')
    && !type.startsWith('@@router/')
  )
}

export default store => next => function (action) {
  if (window.log && filter(action)) {
    console.groupCollapsed(action.type)

    // log meta
    const meta = action.meta || {}
    if (meta.id) {
      console.log(`%cid %c–%c ${action.meta.id}`, 'font-weight: bold', 'color: #aaa', action.meta.id)
    }

    if (meta.created) {
      console.log('%ccreated', 'font-weight: bold', action.meta.created)
    }

    // log payload
    console.group('payload')
    Object.keys(action.payload).forEach(function (key) {
      console.log(key, action.payload[key])
    })

    if (Object.keys(action.payload).length === 0) {
      console.log('%cno fields in payload', 'color: gray; font-style: italic')
    }
    console.groupEnd('payload')

    console.groupEnd(action.type)
  }

  return next(action)
}
