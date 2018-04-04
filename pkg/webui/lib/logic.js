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

// @flow

import delay from "./delay"
import { cancel, fork, take, cancelled, put } from "redux-saga/effects"

/**
 * Canceled is the error that gets passed to failure if a saga gets canceled.
 */
export const Canceled = new Error("Canceled")

declare interface typed {
  type : string,
}

type ActionCreator = Function & typed

/**
 * ActionType is anything that can represent an action.
 */
export type ActionType = string | ActionCreator

export type LogicDefinition = {
  type : ActionType | ActionType[],
  process : Function,
  throttle? : number,
  debounce? : number,
  canceledBy? : ActionType | ActionType[],
  latest? : bool,
  success? : ?ActionType,
  failure? : ?ActionType,
  canceled? : ?ActionType,
  init? : ?ActionType,
  failureOnCancel? : bool,
}

/**
 * Anything that can be an action type: a string or a function that has the type property.
 * @typedef {string|function} actiontype
 */

/**
 * redux-logic-like declarative saga builder.
 * useful for repetetive sagas (like data fetching with cancelation).
 *
 * @param {object} def - An object describing the logic.
 * @param {actiontype|actiontype[]} def.type - The type of actions that
 *   trigger this logic.
 * @param {function} def.process - The function used to process the logic.
 * @param {actiontype|actiontype[]} [def.cancelledBy] - The type of actions that cancel the logic.
 * @param {bool} [def.latest = false] - Only take the latest action (cancel running processes when a new one is started).
 * @param {function} [def.success] - The action creator for success, receives the result of process.
 * @param {function} [def.failure] - The action creator for failure, receives the error that occurred.
 * @param {function} [def.init] - The action creator for initialization (created every time right before process is started).
 * @param {function} [def.canceled] - The action creator for cancelation (created every time a process gets
 *   interupted due to cancelation or because of latest being true). This only gets used if failureOnCancel is false.
 * @param {bool} [def.failureOnCancel = true] - Wether or not to fail when process is canceled instead
 *   of creating a canceled action. By default this is true, unless canceled is set.
 * @returns {function} - A generator function that is the saga for this logic.
 *
 * @example
 * // a saga that will be triggered by the `foo` action.
 * // If process ends successfully, the `foo_ok` action creator will be invoked with the result and the original `foo` action.
 * // If process fails, the `foo_failure` action creator will be invoked with the error and the original `foo` action.
 * // The `foo_init` action will be invoked right before the body of process.
 * // The `foo_cancel` action will cancel any running process and prevent `foo_ok` or `foo_fail` to be called for them. The `foo_cancelled`
 * // action will be called if this happens.
 * //
 * // For instance:
 * //
 * //           foo --1----
 * //      foo_init ---*---
 * //        foo_ok ----5--
 * //      foo_fail -------
 * //    foo_cancel -------
 * // foo_cancelled -------
 * //
 * //           foo --0----
 * //      foo_init ---*---
 * //        foo_ok -------
 * //      foo_fail ----*--
 * //    foo_cancel -------
 * // foo_cancelled -------
 * //
 * //           foo --1----
 * //      foo_init ---*---
 * //        foo_ok -------
 * //      foo_fail -------
 * //    foo_cancel ----*--
 * // foo_cancelled -----*-
 * logic({
 *   type: foo,
 *   success: foo_ok,
 *   failure: foo_fail,
 *   init: foo_init,
 *   cancelledBy: foo_cancel,
 *   cancelled: foo_cancelled,
 *   * process (action) {
 *     if (action.payload === 0) {
 *       throw new Error("foo is 0")
 *     }
 *
 *     return 5*action.payload
 *   },
 * })
 *
 * @example
 * // a saga that will be triggered by the `foo` action.
 * // If process ends successfully, the `foo_ok` action creator will be invoked with the result and the original `foo` action.
 * // If process fails, the `foo_failure` action creator will be invoked with the error and the original `foo` action.
 * // The `foo_init` action will be invoked right before the body of process.
 * // The `foo_cancel` action will cancel any running process and prevent and cause `foo_fail` to be invoked.
 * //
 * // For instance:
 * //
 * //           foo --1----
 * //      foo_init ---*---
 * //        foo_ok -------
 * //      foo_fail -----*-
 * //    foo_cancel ----*--
 * //
 * logic({
 *   type: foo,
 *   success: foo_ok,
 *   failure: foo_fail,
 *   init: foo_init,
 *   cancelledBy: foo_cancel,
 *   * process (action) {
 *     if (action.payload === 0) {
 *       throw new Error("foo is 0")
 *     }
 *
 *     return 5*action.payload
 *   },
 * })
 */
export default function (def : LogicDefinition) {
  if (!def) {
    throw TypeError("logic requires a definition")
  }

  const {
    type,
    canceledBy,
    latest = false,
    process,
    success,
    failure,
    canceled: canceledType,
    failureOnCancel = !canceledType,
    init,
    throttle = 0,
    debounce = 0,
  } = normalize(def)

  if (!type) {
    throw new TypeError("logic: requires a type")
  }

  if (!process) {
    throw new TypeError("logic: requires a process function")
  }

  if (debounce && throttle) {
    throw new TypeError("logic: cannot debounce and throttle at the same time")
  }

  if (debounce && latest) {
    throw new TypeError("logic: debounce and latest cannot be used together")
  }

  if (throttle && latest) {
    throw new TypeError("logic: throttle and latest cannot be used together")
  }

  // listen for actions
  const startTypes = type.map(getType)
  const cancelTypes = canceledBy.map(getType)

  return function * () : Generator<any, any, any> {
    let tasks = []

    const cancelAll = function * () {
      yield cancel(...tasks)
      tasks = []
    }

    try {
      // loop that listens for start actions
      yield fork(function * () {
        for (;;) {
          const action = yield take(startTypes)

          // cancel previous task
          if (tasks.length > 0 && (latest || debounce)) {
            yield cancelAll()
          }

          tasks.push(yield fork(function * () {
            if (debounce) {
              yield delay(debounce)
            }

            try {
              if (init) {
                yield put(init(action.payload))
              }

              const res = yield process(action, put)

              if (success) {
                yield put(success(res || {}, action))
              }
            } catch (err) {
              if (failure) {
                yield put(failure(err, action))
              }
            } finally {
              if (yield cancelled()) {
                if (!failureOnCancel && canceledType) {
                  yield put(canceledType(action.payload))
                }

                if (failureOnCancel && failure) {
                  yield put(failure(Canceled))
                }
              }
            }
          }))

          // throttle
          if (throttle) {
            yield delay(throttle)
          }
        }
      })

      // loop that listens for cancel actions
      yield fork(function * () {
        for (;;) {
          yield take(cancelTypes)
          yield delay(1)
          yield cancelAll()
        }
      })
    } catch (err) {
      console.log("logic error", err)
      throw err
    } finally {
      if (yield cancelled() && tasks.length > 0) {
        yield cancelAll()
      }
    }
  }
}

const getType = function (thing : ?ActionType) : string {
  if (!thing) {
    throw TypeError("passed invalid type to logic")
  }

  if (thing instanceof Function && thing.type) {
    return thing.type
  }

  return thing.toString()
}

const creator = function (type : ?ActionType) : ?Function {
  if (!type) {
    return
  }

  if (type instanceof Function) {
    return type
  }

  return payload => ({ type, payload })
}

const toArray = function (thing : any) : any[] {
  if (Array.isArray(thing)) {
    return thing
  }

  if (!thing) {
    return []
  }

  return [ thing ]
}

type NormalizedDefinition = {
  type : ActionType[],
  process : Function,
  throttle? : number,
  debounce? : number,
  canceledBy : ActionType[],
  latest? : bool,
  success? : ?Function,
  failure? : ?Function,
  canceled? : ?Function,
  init? : ?Function,
  failureOnCancel? : bool,
}

const normalize = function (def : LogicDefinition) : NormalizedDefinition {
  const {
    canceledBy,
    success,
    failure,
    canceled,
    init,
    type,
    ...rest
  } = def

  return {
    ...rest,
    type: toArray(type),
    canceledBy: toArray(canceledBy),
    success: creator(success),
    failure: creator(failure),
    init: creator(init),
    canceled: creator(canceled),
  }
}


