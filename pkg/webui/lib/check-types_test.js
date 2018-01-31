// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

/* eslint-env jest */
/* eslint-disable arrow-body-style */

import PropTypes from "prop-types"
import check from "./check-types"

test("check-types should throw when a type is wrong", () => {
  const types = {
    foo: PropTypes.string,
  }

  const val = {
    foo: 10,
  }

  expect(() => check(types, val, "test", "lib")).toThrow(/string/)
})

test("check-types should not throw when a type is not wrong", () => {
  const types = {
    foo: PropTypes.string,
  }

  const val = {
    foo: "10",
  }

  expect(() => check(types, val, "test", "lib")).not.toThrow()
})

test("check-types should throw when a type definition is wrong", () => {
  const types = {
    foo: "invalid",
  }

  const val = {
    foo: "ok",
  }

  expect(() => check(types, val, "test", "lib")).toThrow(/type `foo` is invalid/)
})

test("check-types should throw when a type definition throws", () => {
  const types = {
    foo () {
      throw new Error("huh?")
    },
  }

  const val = {
    foo: "ok",
  }

  expect(() => check(types, val, "test", "lib")).toThrow(/Failed/)
})
