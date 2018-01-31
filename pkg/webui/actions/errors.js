// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

import PropTypes from "prop-types"
import actions from "../lib/action"

export default actions("errors", {
  // fatal is a fatal error in the app
  fatal: {
    error: PropTypes.any,
  },

  // uncaught is an uncaught error in the app
  uncaught: {
    error: PropTypes.any,
  },
})
