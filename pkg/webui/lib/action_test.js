// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

/* eslint-env jest */
/* eslint-disable arrow-body-style */

import PropTypes from "prop-types"
import actions from "./action"

test("should create the proper actions", () => {
  const ns = actions("ns", {
    foo: {
      types: {
        bar: PropTypes.string,
      },
    },
    bar: {
      baz: PropTypes.string,
      transform: PropTypes.number,
    },
    quu: {
      error: PropTypes.instanceOf(Error),
    },
    qux: {
      types: {
        bar: PropTypes.string,
      },
      transform (payload) {
        return {
          bar: (payload.bar || "").toString(),
        }
      },
    },
  })

  expect(ns).toHaveProperty("foo")
  expect(ns.foo).toHaveProperty("type", "ns/foo")
  expect(ns.foo.toString()).toEqual("ns/foo")

  expect(ns).toHaveProperty("bar")
  expect(ns.bar).toHaveProperty("type", "ns/bar")
  expect(ns.bar.toString()).toEqual("ns/bar")

  expect(() => ns.foo({ bar: "bar" })).not.toThrow()
  expect(() => ns.foo({ bar: 10 })).toThrow()

  expect(() => ns.quu({ error: new Error() })).not.toThrow()
  expect(() => ns.qux({ bar: 10 })).not.toThrow()
})
