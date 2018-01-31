// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

/* global process */

window.log = (window.ENV || {}).debug || process.env.NODE_ENV === "development"

const filter = function (action) {
  const {
    type = "",
  } = action

  return (
    !type.startsWith("@@redux-form/")
    && !type.startsWith("@@router/")
  )
}

export default store => next => function (action) {
  if (window.log && filter(action)) {
    console.groupCollapsed(action.type)

    // log meta
    const meta = action.meta || {}
    if (meta.id) {
      console.log(`%cid %c–%c ${action.meta.id}`, "font-weight: bold", "color: #aaa", action.meta.id)
    }

    if (meta.created) {
      console.log("%ccreated", "font-weight: bold", action.meta.created)
    }

    // log payload
    console.group("payload")
    Object.keys(action.payload).forEach(function (key) {
      console.log(key, action.payload[key])
    })

    if (Object.keys(action.payload).length === 0) {
      console.log("%cno fields in payload", "color: gray; font-style: italic")
    }
    console.groupEnd("payload")

    console.groupEnd(action.type)
  }

  return next(action)
}
