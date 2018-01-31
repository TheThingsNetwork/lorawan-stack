// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

export default function (t = 0) {
  return new Promise(resolve => setTimeout(resolve, t))
}
