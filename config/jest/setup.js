// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

import 'raf/polyfill'
import { configure, shallow, render, mount } from 'enzyme'
import Adapter from 'enzyme-adapter-react-16'
import snapshotDiff from 'snapshot-diff'

configure({ adapter: new Adapter() })

global.shallow = shallow
global.mount = mount
global.render = render
global.snapshotDiff = snapshotDiff

/* eslint-disable no-console */
const originalConsoleError = console.error
console.error = function (message, ...args) {
  console.log(message)
  if (/(Invalid prop|Failed prop type|Failed context type)/gi.test(message)) {
    throw new Error(message)
  }

  originalConsoleError.apply(console, [ message, ...args ])
}
/* eslint-enable no-console */
